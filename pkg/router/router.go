package router

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/craniacshencil/beaker/utils"
)

type (
	HandlerFunc func(request *Request) (response Response)
	Router      struct {
		mappings map[string]HandlerFunc
	}
)

func CreateRouter() (router *Router) {
	return &Router{
		mappings: make(map[string]HandlerFunc),
	}
}

func (router *Router) Register(method, path string, handler HandlerFunc) {
	routerKey := fmt.Sprintf("%s-%s", path, method)
	if _, ok := router.mappings[routerKey]; ok {
		fmt.Println(ok)
		panic("Duplicate path in router")
	}
	router.mappings[routerKey] = handler
}

func (router *Router) ServiceRequest(
	path,
	method string,
	headers map[string]string,
	body io.Reader,
) (res []byte, err error) {
	var response Response

	if strings.Contains(path, ".jpeg") || strings.Contains(path, ".jpg") ||
		strings.Contains(path, ".png") ||
		strings.Contains(path, ".gif") {
		response = serveImage([]byte(path))
	} else {
		routerKey := fmt.Sprintf("%s-%s", path, method)
		if _, ok := router.mappings[routerKey]; !ok {
			response = Response{
				StatusCode: 404,
				StatusText: "NOT FOUND",
				Headers:    make(map[string]string),
				Body:       []byte("Path not found"),
			}
			response.Headers["Content-type"] = "text/html"
		} else {
			request := Request{
				path:    path,
				method:  method,
				headers: headers,
				body:    body,
			}
			routeHandler := router.mappings[routerKey]
			response = routeHandler(&request)
		}
	}

	res = response.Serialize()
	return res, nil
}

func serveImage(path []byte) (response Response) {
	image, err := os.ReadFile("public/" + string(path))
	if err != nil {
		response = Response{
			StatusCode: 404,
			StatusText: "NOT FOUND",
			Headers:    make(map[string]string),
			Body:       []byte("Image not found"),
		}
		response.Headers["Content-type"] = "text/html"
		return response
	}
	response = Response{
		StatusCode: 200,
		StatusText: "OK",
		Headers:    make(map[string]string),
		Body:       image,
	}
	if utils.ArrIndex(path, []byte(".jpeg")) != -1 || utils.ArrIndex(path, []byte(".jpg")) != -1 {
		response.Headers["Content-type"] = "image/jpeg"
	} else if utils.ArrIndex(path, []byte(".png")) != -1 {
		response.Headers["Content-type"] = "image/png"
	} else {
		response.Headers["Content-type"] = "image/gif"
	}
	return response
}
