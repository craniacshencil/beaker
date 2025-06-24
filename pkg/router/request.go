package router

import (
	"errors"

	"github.com/craniacshencil/beaker/utils"
)

type Request struct {
	path    string
	method  string
	headers []byte
	body    []byte
}

// Add function to get headers as a map
func (request *Request) GetHeaders() (headersMap map[string]string) {
	headersMap = make(map[string]string)
	CRLF_occurences := utils.ArrAllIndex(request.headers, []byte("\r\n"))
	startIndex := 0
	for _, endIndex := range CRLF_occurences {
		currentLine := request.headers[startIndex:endIndex]
		keyValSeparator := utils.ArrIndex(currentLine, []byte(":"))
		key := currentLine[:keyValSeparator]
		// +2 to get rid of ": ", colon and whitespace
		value := currentLine[keyValSeparator+2:]
		headersMap[string(key)] = string(value)
		// To skip the \r\n
		startIndex = endIndex + 2
	}
	return headersMap
}

// Add function to get headers using a key
func (request *Request) GetHeaderValue(key string) (value string, err error) {
	headers := request.GetHeaders()
	val, ok := headers[key]
	if !ok {
		return "", errors.New("Invalid header key")
	}
	return val, nil
}

// body parsing, especially for application/json POST request and stuff
