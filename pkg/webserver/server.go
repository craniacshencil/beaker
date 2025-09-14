package webserver

import (
	"errors"
	"log"
	"net"
	"slices"
	"strconv"
	"sync"

	"github.com/craniacshencil/beaker/pkg/router"
	"github.com/craniacshencil/beaker/utils"
)

const (
	MAX_REQUEST_SIZE = 10240
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
				return make([]byte, MAX_REQUEST_SIZE)
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

	buf := httpServer.BufPool.Get().([]byte)
	defer httpServer.BufPool.Put(buf)
	defer conn.Close()

	n, err := conn.Read(buf)
	if err != nil {
		log.Println("While reading: ", err)
	}

	data := buf[:n]
	path, method, headers, body, err := parseRequest(data)
	if err != nil {
		log.Println("While parsing first line and headers: ", err)
	}

	res, err := httpServer.Webrouter.ServiceRequest(path, method, headers, body)
	if err != nil {
		log.Println("While servicing request: ", err)
	}
	conn.Write(res)
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

func parseRequest(
	requestStream []byte,
) (path []byte, method []byte, headers []byte, body []byte, err error) {
	// start := time.Now()
	CRLF_BYTES := []byte("\r\n")
	headersStartIndex := utils.ArrIndex(requestStream, CRLF_BYTES)
	if headersStartIndex == -1 {
		return nil, nil, nil, nil, errors.New("CRLF not present for header start")
	}
	request := []byte(requestStream)[:headersStartIndex]
	path, method, err = parseRequestLine(request)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// log.Println("Time taken to parse request: ", time.Since(start))

	DOUBLE_CRLF_BYTES := []byte("\r\n\r\n")
	if slices.Equal(method, []byte("GET")) || slices.Equal(method, []byte("DELETE")) {
		headers = requestStream[headersStartIndex+2:]
		body = nil
	} else {
		headersEndIndex := utils.ArrIndex(requestStream, DOUBLE_CRLF_BYTES)
		if headersEndIndex == -1 {
			return nil, nil, nil, nil, errors.New("CRLF not present for headers end")
		}
		headers = requestStream[headersStartIndex+2 : headersEndIndex]
		body = requestStream[headersEndIndex+4:]
	}
	headersMap, err := parseAndValidateHeaders(headers)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	contentLengthString, ok := headersMap["Content-Length"]
	if body != nil && ok {
		// Incase of empty body
		contentLength, err := strconv.Atoi(contentLengthString)
		if err != nil {
			return nil, nil, nil, nil, errors.New("Invalid content-length")
		}
		if contentLength > MAX_REQUEST_SIZE {
			return nil, nil, nil, nil, errors.New("Request body size exceeded max limit")
		}
		err = validateBody([]byte(headersMap["Content-Type"]), body)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}
	// log.Printf("For path: %s, time taken to parse headers: %v", path, time.Since(start))
	return path, method, headers, body, nil
}

func parseRequestLine(requestLine []byte) (path, method []byte, err error) {
	// request-line format: http-method path HTTP/version_no
	WHITESPACE_BYTE := []byte(" ")
	firstWhitespace := utils.ArrIndex(requestLine, WHITESPACE_BYTE)
	secondWhitespace := utils.ArrLastIndex(requestLine, WHITESPACE_BYTE)
	if firstWhitespace == secondWhitespace {
		return nil, nil, errors.New("Request line not formed correctly")
	}

	method = requestLine[:firstWhitespace]
	path = requestLine[firstWhitespace+1 : secondWhitespace]
	httpVersion := requestLine[secondWhitespace+1:]

	err = validateHttpVersion(httpVersion)
	if err != nil {
		return nil, nil, err
	}

	err = validatePath(path)
	if err != nil {
		return nil, nil, err
	}

	err = validateMethod(method)
	if err != nil {
		return nil, nil, err
	}

	return path, method, nil
}

func parseAndValidateHeaders(headerBytes []byte) (headersMap map[string]string, err error) {
	headersMap = make(map[string]string)
	CRLF_occurences := utils.ArrAllIndex(headerBytes, []byte("\r\n"))
	startIndex := 0
	for _, endIndex := range CRLF_occurences {
		if startIndex == endIndex {
			break
		}
		currentLine := headerBytes[startIndex:endIndex]
		err = utils.Mapify(headersMap, currentLine, []byte(": "))
		if err != nil {
			return nil, err
		}
		startIndex = endIndex + 2
	}
	return headersMap, nil
}
