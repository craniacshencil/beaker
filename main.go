package main

import (
	"github.com/craniacshencil/beaker/pkg/router"
	"github.com/craniacshencil/beaker/pkg/webserver"
)

func main() {
	myServer := webserver.CreateServer("127.0.0.1", 4200, 3)
	myServer.Webrouter.Register("GET", "/", helloWorld)
	myServer.Webrouter.Register("GET", "/another", helloAnotherPath)
	myServer.Webrouter.Register("POST", "/", helloPost)
	myServer.Listen()
}

func helloWorld(request *router.Request) (response router.Response) {
	response.Headers = make(map[string]string)
	response.Headers["Content-Type"] = "text/html"
	response.StatusCode = 200
	response.StatusText = "OK"
	response.Body = []byte("hello world")
	return response
}

func helloAnotherPath(request *router.Request) (response router.Response) {
	response.Headers = make(map[string]string)
	response.Headers["Content-Type"] = "text/html"
	response.StatusCode = 200
	response.StatusText = "OK"
	response.Body = []byte("hello another path")
	return response
}

func helloPost(request *router.Request) (response router.Response) {
	response.Headers = make(map[string]string)
	response.Headers["Content-type"] = "text/html"
	response.StatusCode = 200
	response.StatusText = "OK"
	response.Body = []byte("POST received")
	return response
}
