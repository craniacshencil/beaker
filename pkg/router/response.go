package router

import (
	"fmt"
	"strconv"
)

type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

// Write a util function to convert map[string]string to []byte (also include \r\n, everytime too)
func (response Response) marshalHeaders() (headerBytes []byte) {
	for key, val := range response.Headers {
		line := []byte(key + ": " + val + "\r\n")
		headerBytes = append(headerBytes, line...)
	}
	// Need to add an extra \r\n for line separating headers-body but that can be done in FormatResponse
	return headerBytes
}

func (response Response) Serialize() (res []byte) {
	response.Headers["Content-Length"] = strconv.Itoa(len(response.Body))
	// This isn't adding \r\n after every value
	headers := response.marshalHeaders()
	responseLine := []byte(
		fmt.Sprintf("HTTP/1.1 %d %s\r\n", response.StatusCode, response.StatusText),
	)
	responseLineAndHeaders := append(responseLine, headers...)
	responseLineAndHeaders = append(responseLineAndHeaders, []byte("\r\n")...)
	return append(responseLineAndHeaders, response.Body...)
}
