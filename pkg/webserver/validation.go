package webserver

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"slices"

	"github.com/craniacshencil/beaker/utils"
)

func parseFirstLineAndHeader(
	requestStream []byte,
) (headers map[string]string, request []byte, err error) {
	CRLF_BYTES := []byte("\r\n")
	headersIndex := utils.ArrLastIndex(requestStream, CRLF_BYTES)
	firstLineIndex := utils.ArrIndex(requestStream, CRLF_BYTES)
	if headersIndex == -1 {
		return nil, nil, errors.New("CRLF not present for header")
	}
	if firstLineIndex == -1 {
		return nil, nil, errors.New("CRLF not present for first line")
	}
	headerBytes := []byte(requestStream)[firstLineIndex+2 : headersIndex]
	request = []byte(requestStream)[:firstLineIndex]
	path, method, err := parseRequestLine(request)
	if err != nil {
		return nil, nil, err
	}
	headers, err = parseHeaders(headerBytes)
	if err != nil {
		return nil, nil, err
	}
	log.Println("path:", string(path))
	log.Println("method: ", string(method))
	return headers, request, nil
}

func parseRequestLine(requestLine []byte) (path, method []byte, err error) {
	// request-line format: http-method path HTTP/version_no
	WHITESPACE_BYTE := []byte(" ")
	firstWhitespace := utils.ArrIndex(requestLine, WHITESPACE_BYTE)
	secondWhitespace := utils.ArrLastIndex(requestLine, WHITESPACE_BYTE)
	method = requestLine[:firstWhitespace]
	path = requestLine[firstWhitespace+1 : secondWhitespace]
	httpVersion := requestLine[secondWhitespace+1:]
	err = validateHttpVersion(httpVersion)
	if err != nil {
		return nil, nil, err
	}
	err = validatePath(path)
	if err != nil {
		return nil, nil, err
	}
	err = validateMethod(method)
	if err != nil {
		return nil, nil, err
	}
	return path, method, nil
}

func validateHttpVersion(httpVersion []byte) (err error) {
	SLASH_BYTE := []byte("/")
	slashIndex := utils.ArrIndex(httpVersion, SLASH_BYTE)
	if slashIndex == -1 {
		// HTTP/0.9 case, this is not mentioned the last field is kept empty
		// For HTTP/0.9:
		// request-line format: http-method path
		return nil
	}
	protocolName := httpVersion[:slashIndex]
	if !slices.Equal(protocolName, []byte("HTTP")) {
		return errors.New(fmt.Sprintf("Protocol is not HTTP: %s\n", protocolName))
	}
	version := httpVersion[slashIndex+1:]
	if !slices.Equal(version, []byte("1.0")) && !slices.Equal(version, []byte("1.1")) &&
		!slices.Equal(version, []byte("2.0")) {
		return errors.New(fmt.Sprintf("Http version: %s is invalid\n", version))
	}
	return nil
}

func validatePath(path []byte) (err error) {
	if len(path) == 0 {
		return errors.New("No path provided")
	}

	if len(path) > 2048 {
		return errors.New("Path too long")
	}

	isFirstCharSlash := utils.ArrIndex(path, []byte("/"))
	if isFirstCharSlash != 0 {
		return errors.New("Path doesn't begin with a slash")
	}

	if slices.Min(path) < 32 {
		return errors.New("Illegal control character present in path")
	}

	percentageIndices := utils.ArrAllIndex(path, []byte("%"))
	temp := make([]byte, 1)
	for i := 0; i < len(percentageIndices); i++ {
		startIndex := percentageIndices[i]
		if startIndex+3 > len(path) {
			return errors.New("Invalid percentage encoded characters")
		}
		_, err = hex.Decode(temp, path[startIndex+1:startIndex+3])
		if err != nil {
			return errors.New(
				fmt.Sprintf(
					"Invalid percentage encoded characters: %s",
					path[startIndex+1:startIndex+3],
				),
			)
		}
	}

	if utils.ArrIndex(path, []byte("..")) != -1 {
		return errors.New(".. is not allowed in the path")
	}

	if utils.ArrIndex(path, []byte("//")) != -1 {
		return errors.New("// are not allowed in the path")
	}

	// Case for not allowing backslashes
	// Double backslashes because single '\' is an escape sequence
	if utils.ArrIndex(path, []byte("\\")) != -1 {
		return errors.New("\\ are not allowed in the path")
	}
	return nil
}

func validateMethod(method []byte) (err error) {
	OPTIONS := [][]byte{[]byte("GET"), []byte("POST"), []byte("DELETE"), []byte("PUT")}
	for _, val := range OPTIONS {
		if slices.Equal(method, val) {
			return nil
		}
	}
	return errors.New("Invalid HTTP method")
}

func parseHeaders(headerBytes []byte) (headers map[string]string, err error) {
	headers = make(map[string]string)
	CRLF_occurences := utils.ArrAllIndex(headerBytes, []byte("\r\n"))
	double_CRLF_occurences := utils.ArrAllIndex(headerBytes, []byte("\r\n\r\n"))
	log.Println(double_CRLF_occurences)
	startIndex := 0
	for _, endIndex := range CRLF_occurences {
		currentLine := headerBytes[startIndex:endIndex]
		keyValSeparator := utils.ArrIndex(currentLine, []byte(":"))
		if keyValSeparator == -1 {
			return nil, errors.New("Invalid header key-value pair, no colon found")
		}
		key := currentLine[:keyValSeparator]
		// +2 to get rid of ": ", colon and whitespace
		value := currentLine[keyValSeparator+2:]
		headers[string(key)] = string(value)
		// To skip the \r\n
		startIndex = endIndex + 2
	}
	return headers, nil
}
