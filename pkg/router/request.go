package router

import "io"

type Request struct {
	path    string
	method  string
	headers map[string]string
	body    io.Reader
}

// body parsing, especially for application/json POST request and stuff
func (request *Request) parseBody() {
	// if body != nil && ok {
	// 	// Incase of empty body
	// 	contentLength, err := strconv.Atoi(contentLengthString)
	// 	if err != nil {
	// 		return nil, nil, nil, nil, errors.New("Invalid content-length")
	// 	}
	// 	if contentLength > MAX_REQUEST_SIZE {
	// 		return nil, nil, nil, nil, errors.New("Request body size exceeded max limit")
	// 	}
	// 	err = validateBody([]byte(headersMap["Content-Type"]), body)
	// 	if err != nil {
	// 		return nil, nil, nil, nil, err
	// 	}
	// }
}

// application/json and application/x-www-form-urlencoded: Takes in strictly typed format and gives validated output or error
