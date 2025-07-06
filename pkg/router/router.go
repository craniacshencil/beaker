package router

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"

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
	method,
	body,
	headers []byte,
) (res []byte, err error) {
	var response Response
	if len(path) > 6 && (slices.Equal(path[len(path)-5:], []byte(".jpeg")) ||
		slices.Equal(path[len(path)-4:], []byte(".jpg")) ||
		slices.Equal(path[len(path)-4:], []byte(".png")) ||
		slices.Equal(path[len(path)-4:], []byte(".gif"))) {
		response = serveImage(path)
	} else {
		pathString := string(path)
		methodString := string(method)
		routerKey := fmt.Sprintf("%s-%s", pathString, methodString)
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
				path:    pathString,
				method:  methodString,
				headers: headers,
				body:    body,
			}
			routeHandler := router.mappings[routerKey]
			response = routeHandler(&request)
		}
	}

	res, err = formatResponse(&response)
	if err != nil {
		return nil, errors.New("Something went wrong while formatting the request")
	}
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

func formatResponse(response *Response) (res []byte, err error) {
	response.Headers["Content-Length"] = strconv.Itoa(len(response.Body))
	// This isn't adding \r\n after every value
	headers := response.marshalHeaders()
	responseLine := []byte(
		fmt.Sprintf("HTTP/1.1 %d %s\r\n", response.StatusCode, response.StatusText),
	)
	responseLineAndHeaders := append(responseLine, headers...)
	responseLineAndHeaders = append(responseLineAndHeaders, []byte("\r\n")...)
	return append(responseLineAndHeaders, response.Body...), nil
}
