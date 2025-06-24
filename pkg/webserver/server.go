package webserver

import (
	"errors"
	"log"
	"net"
	"slices"

	"github.com/craniacshencil/beaker/pkg/router"
	"github.com/craniacshencil/beaker/utils"
)

type HttpServer struct {
	server    *net.TCPListener
	address   net.TCPAddr
	Webrouter *router.Router
}

func CreateServer(host string, port int) *HttpServer {
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
	}
	return &httpServer
}

func (httpServer *HttpServer) Listen() {
	data := make([]byte, 1024)
	log.Printf(
		"Server up and running on %s:%d\n",
		httpServer.address.IP,
		httpServer.address.Port,
	)
	for {
		conn, err := httpServer.server.Accept()
		if err != nil {
			log.Println("While listening: ", err)
		}
		_, err = conn.Read(data)
		if err != nil {
			log.Println("While reading: ", err)
		}
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
}

func parseRequest(
	requestStream []byte,
) (path []byte, method []byte, headers []byte, body []byte, err error) {
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
	_, err = parseAndValidateHeaders(headers)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return path, method, headers, body, nil
}

func parseRequestLine(requestLine []byte) (path, method []byte, err error) {
	// request-line format: http-method path HTTP/version_no
	WHITESPACE_BYTE := []byte(" ")
	firstWhitespace := utils.ArrIndex(requestLine, WHITESPACE_BYTE)
	secondWhitespace := utils.ArrLastIndex(requestLine, WHITESPACE_BYTE)
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
		currentLine := headerBytes[startIndex:endIndex]
		keyValSeparator := utils.ArrIndex(currentLine, []byte(":"))
		if keyValSeparator == -1 {
			return nil, errors.New("Invalid header key-value pair, no colon found")
		}
		key := currentLine[:keyValSeparator]
		// +2 to get rid of ": ", colon and whitespace
		value := currentLine[keyValSeparator+2:]
		headersMap[string(key)] = string(value)
		// To skip the \r\n
		startIndex = endIndex + 2
	}
	return headersMap, nil
}
