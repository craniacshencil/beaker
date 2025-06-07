package main

import (
	"github.com/craniacshencil/beaker/pkg/webserver"
)

func main() {
	myServer := webserver.CreateServer("127.0.0.1", 4200)
	myServer.Listen()
}
