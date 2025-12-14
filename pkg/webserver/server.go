package webserver

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/craniacshencil/beaker/pkg/router"
	"github.com/craniacshencil/beaker/utils"
)

const (
	// 8 Kb
	MAX_HEADER_SIZE            = 8192
	MAX_HEADER_SIZE_EXCEED_ERR = "Headers exceed 8kb"
)

type Job struct {
	conn net.Conn
}

type HttpServer struct {
	server    *net.TCPListener
	address   net.TCPAddr
	Webrouter *router.Router
	JobQueue  chan Job
	BufPool   *sync.Pool
}

func (httpServer *HttpServer) Listen() {
	log.Printf(
		"Server up and running on %s:%d\n",
		httpServer.address.IP,
		httpServer.address.Port,
	)
	for {
		conn, err := httpServer.server.Accept()
		if err != nil {
			log.Println("While listening: ", err)
			continue
		}
		httpServer.JobQueue <- Job{conn: conn}
	}
}

func CreateServer(host string, port int, workers int) *HttpServer {
	address := net.TCPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}
	server, err := net.ListenTCP("tcp", &address)
	if err != nil {
		log.Println("While creating server: ", err)
	}

	webrouter := router.CreateRouter()
	httpServer := HttpServer{
		server:    server,
		address:   address,
		Webrouter: webrouter,
		JobQueue:  make(chan Job, 100),
		BufPool: &sync.Pool{
			New: func() any {
				return make([]byte, MAX_HEADER_SIZE)
			},
		},
	}

	for i := 0; i < workers; i++ {
		go httpServer.worker(i)
	}
	return &httpServer
}

func (httpServer *HttpServer) worker(id int) {
	for job := range httpServer.JobQueue {
		conn := job.conn
		httpServer.handleConnection(conn, id)
	}
}

func (httpServer *HttpServer) handleConnection(conn net.Conn, workerId int) {
	log.Printf("New connection: %v handled by %d", conn.RemoteAddr(), workerId)

	headersBytes := httpServer.BufPool.Get().([]byte)
	defer httpServer.BufPool.Put(headersBytes)
	defer conn.Close()

	headersBytes, bodyReader, err := readHeaders(conn)
	if err != nil && err.Error() == MAX_HEADER_SIZE_EXCEED_ERR {
		handleError(conn, 413, "ENTITY TOO LARGE", MAX_HEADER_SIZE_EXCEED_ERR)
		return
	}

	if err != nil {
		log.Println("While reading headers: ", err)
		handleError(conn, 400, "BAD REQUEST", "Malformed Headers")
		return
	}

	path, method, headers, err := parseHeaders(headersBytes)
	if err != nil {
		log.Println("While parsing headers: ", err)
		handleError(conn, 400, "BAD REQUEST", "Malformed Headers")
		return
	}

	res, err := httpServer.Webrouter.ServiceRequest(path, method, headers, bodyReader)
	if err != nil {
		log.Println("While servicing request: ", err)
		handleError(conn, 400, "BAD REQUEST", "Malformed Body")
		return
	}

	conn.Write(res)
	return
}

func readHeaders(conn net.Conn) (headersBytes []byte, bodyReader io.Reader, err error) {
	br := bufio.NewReader(conn)
	var headersBuf bytes.Buffer

	for {
		line, readErr := br.ReadString('\n')
		if len(line) > 0 {
			_, _ = headersBuf.WriteString(line)
		}
		if readErr != nil && readErr != io.EOF {
			return nil, nil, readErr
		}

		if line == "\r\n" || line == "\n" {
			break
		}

		if readErr == io.EOF {
			break
		}
	}

	headersBytes = headersBuf.Bytes()
	if len(headersBytes) > MAX_HEADER_SIZE {
		return nil, nil, errors.New(MAX_HEADER_SIZE_EXCEED_ERR)
	}
	bodyReader = br
	return headersBytes, bodyReader, nil
}

func parseHeaders(
	headersBytes []byte,
) (path, method string, headers map[string]string, err error) {
	var headersOnly []byte
	CRLF_BYTES := []byte("\r\n")
	headersStartIndex := utils.ArrIndex(headersBytes, CRLF_BYTES)
	if headersStartIndex == -1 {
		return "", "", nil, errors.New("No CRLF present in headers")
	}
	requestLine := string(headersBytes[:headersStartIndex])
	path, method, err = parseRequestLine(requestLine)
	if err != nil {
		return "", "", nil, err
	}

	headersOnly = headersBytes[headersStartIndex+2:]
	// headersOnly = headersBytes[headersStartIndex+2 : headersEndIndex]
	headers, err = parseAndValidateHeaders(headersOnly)
	if err != nil {
		return "", "", nil, err
	}

	return path, method, headers, nil
}

func parseRequestLine(requestLine string) (path, method string, err error) {
	// request-line format: http-method path HTTP/version_no
	requestLineSplit := strings.Split(requestLine, " ")

	if len(requestLineSplit) != 3 {
		return "", "", errors.New("Request line not formed correctly")
	}

	method = requestLineSplit[0]
	path = requestLineSplit[1]
	httpVersion := requestLineSplit[2]

	err = validateHttpVersion([]byte(httpVersion))
	if err != nil {
		return "", "", err
	}

	err = validatePath([]byte(path))
	if err != nil {
		return "", "", err
	}

	err = validateMethod([]byte(method))
	if err != nil {
		return "", "", err
	}

	return path, method, nil
}

func parseAndValidateHeaders(headerBytes []byte) (headers map[string]string, err error) {
	headers = make(map[string]string)
	CRLF_occurences := utils.ArrAllIndex(headerBytes, []byte("\r\n"))
	startIndex := 0
	for _, endIndex := range CRLF_occurences {
		if startIndex == endIndex {
			break
		}
		currentLine := headerBytes[startIndex:endIndex]
		err = utils.Mapify(headers, currentLine, []byte(": "))
		if err != nil {
			return nil, err
		}
		startIndex = endIndex + 2
	}
	return headers, nil
}

func handleError(conn net.Conn, statusCode int, statusText, body string) {
	errRes := router.Response{
		StatusCode: statusCode,
		StatusText: statusText,
		Headers:    map[string]string{"Content-Type": "text/html"},
		Body:       []byte(fmt.Sprintf("%d: %s. %s.", statusCode, statusText, body)),
	}
	conn.Write(errRes.Serialize())
	return
}
