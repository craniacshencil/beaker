package webserver

import (
	"log"
	"net"
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
		_, _, err = parseFirstLineAndHeader(data)
		if err != nil {
			log.Println("While parsing first line and headers: ", err)
		}
	}
}
