package main

import (
	"errors"
	"log"
	"net"

	"github.com/craniacshencil/beaker/utils"
)

func createServer() *net.TCPListener {
	address := net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 4200,
	}
	server, err := net.ListenTCP("tcp", &address)
	if err != nil {
		log.Println("While creating server: ", err)
	}
	return server
}

func main() {
	data := make([]byte, 1024)
	server := createServer()
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("While listening: ", err)
		}
		_, err = conn.Read(data)
		if err != nil {
			log.Println("While reading: ", err)
		}
		headers, request, err := parseFirstLineAndHeader(data)
		if err != nil {
			log.Println("While parsing first line and headers: ", err)
		}
		log.Println(headers, request)
	}
}

// ADD Error Handling in here
func parseFirstLineAndHeader(requestStream []byte) (headers, request []byte, err error) {
	CRLF_BYTES := []byte("\r\n")
	headersIndex := utils.ArrLastIndex(requestStream, CRLF_BYTES)
	firstLineIndex := utils.ArrIndex(requestStream, CRLF_BYTES)
	if headersIndex != -1 && firstLineIndex != -1 {
		headers = []byte(requestStream)[firstLineIndex+2 : headersIndex]
		request = []byte(requestStream)[:firstLineIndex]
	}
	if headersIndex == -1 {
		return nil, nil, errors.New("CRLF not present for header")
	}
	if firstLineIndex == -1 {
		return nil, nil, errors.New("CRLF not present for first line")
	}
	return headers, request, nil
}
