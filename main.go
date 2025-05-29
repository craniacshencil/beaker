package main

import (
	"log"
	"net"
)

func createServer() *net.TCPListener {
	address := net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 42069,
	}
	server, err := net.ListenTCP("tcp", &address)
	if err != nil {
		log.Println("While creating server: ", err)
	}
	return server
}

func main() {
	var data []byte
	server := createServer()
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("While listening: ", err)
		}
		num, err := conn.Read(data)
		if err != nil {
			log.Println("While reading: ", err)
		}
		log.Println("what num is this: ", num)
	}
}
