package router

import (
	"errors"
	"fmt"
	"log"
	"strconv"
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
	pathString := string(path)
	methodString := string(method)
	routerKey := fmt.Sprintf("%s-%s", pathString, methodString)
	if _, ok := router.mappings[routerKey]; !ok {
		return nil, errors.New("No path-method mapping present in router")
	}
	request := Request{
		path:    pathString,
		method:  methodString,
		headers: headers,
		body:    body,
	}
	routeHandler := router.mappings[routerKey]
	log.Println(router.mappings)
	response := routeHandler(&request)
	res, err = formatResponse(&response)
	if err != nil {
		return nil, errors.New("Something went wrong while formatting the request")
	}
	log.Println("response sent?")
	return res, nil
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
