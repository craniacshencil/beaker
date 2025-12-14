package main

import (
	"log"
	"os"
	"strconv"

	"github.com/craniacshencil/beaker/pkg/router"
	"github.com/craniacshencil/beaker/pkg/webserver"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env didn't load, continuing with defaults")
	}
	host := getenv("HOST", "localhost")
	portString := getenv("PORT", "4200")
	port, _ := strconv.Atoi(portString)

	myServer := webserver.CreateServer(host, port, 3)
	myServer.Webrouter.Register("GET", "/", helloWorld)
	myServer.Webrouter.Register("GET", "/another", helloAnotherPath)
	myServer.Webrouter.Register("POST", "/", helloPost)
	myServer.Listen()
}

func getenv(key, fallback string) (value string) {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
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
