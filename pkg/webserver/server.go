package webserver

import (
	"errors"
	"log"
	"net"

	"github.com/craniacshencil/beaker/pkg/router"
	"github.com/craniacshencil/beaker/utils"
)

type HttpServer struct {
	server  *net.TCPListener
	address net.TCPAddr
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
	httpServer := HttpServer{
		server:  server,
		address: address,
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
		path, method, headers, err := parseFirstLineAndHeader(data)
		if err != nil {
			log.Println("While parsing first line and headers: ", err)
		}
		res, err := router.ServiceRequest(path, method, headers)
		if err != nil {
			log.Println("While servicing request: ", err)
		}
		conn.Write(res)
	}
}

func parseFirstLineAndHeader(
	requestStream []byte,
) (path []byte, method []byte, headers map[string]string, err error) {
	CRLF_BYTES := []byte("\r\n")
	headersIndex := utils.ArrLastIndex(requestStream, CRLF_BYTES)
	firstLineIndex := utils.ArrIndex(requestStream, CRLF_BYTES)
	if headersIndex == -1 {
		return nil, nil, nil, errors.New("CRLF not present for header")
	}
	if firstLineIndex == -1 {
		return nil, nil, nil, errors.New("CRLF not present for first line")
	}
	headerBytes := []byte(requestStream)[firstLineIndex+2 : headersIndex]
	request := []byte(requestStream)[:firstLineIndex]
	path, method, err = parseRequestLine(request)
	if err != nil {
		return nil, nil, nil, err
	}
	headers, err = parseHeaders(headerBytes)
	if err != nil {
		return nil, nil, nil, err
	}
	return path, method, headers, nil
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
