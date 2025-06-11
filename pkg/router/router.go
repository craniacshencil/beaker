package router

import (
	"errors"
	"fmt"
	"log"
	"slices"
)

func ServiceRequest(path []byte, method []byte, headers map[string]string) (res []byte, err error) {
	if slices.Equal(path, []byte("/")) && slices.Equal(method, []byte("GET")) {
		log.Println("path match!")
		body := []byte("Hello server!")
		response_headers := fmt.Sprintf(
			"HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: %d\r\n\r\n",
			len(body),
		)
		return append([]byte(response_headers), body...), nil
	}
	return nil, errors.New("Something went wrong while servicing the request")
}
