package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/craniacshencil/beaker/utils"
)

type Request struct {
	path    string
	method  string
	headers map[string]string
	body    io.Reader
}

// body parsing, especially for application/json POST request and stuff
func (request *Request) parseBody() (err error) {
	contentLength, _ := strconv.Atoi(request.headers["Content-Length"])
	err = validateBody(contentLength, request.headers["Content-Type"], request.body)
	if err != nil {
		return err
	}
	return nil
}

func validateBody(contentLength int, contentType string, body io.Reader) (err error) {
	bodyBytes := make([]byte, contentLength)

	if contentType == "application/json" || contentType == "application/x-www-form-urlencoded" ||
		contentType == "text/plain" {
		_, err := body.Read(bodyBytes)
		if err != nil {
			return fmt.Errorf("While converting body to []byte: %s\n", err.Error())
		}
	}

	switch {
	case contentType == "application/json":
		var temp struct{}
		if err = json.Unmarshal(bodyBytes, &temp); err != nil {
			return err
		}

	case contentType == "application/x-www-form-urlencoded":
		if err = validateURLEncodedForm(bodyBytes); err != nil {
			return err
		}
		_, err = parseURLEncodedForm(bodyBytes)
		return err

	case contentType == "multipart/form-data":
	// logic

	case contentType == "text/plain":

	default:
		return errors.New("Toy server: unsupported body type")
	}
	return nil
}

func validateURLEncodedForm(body []byte) (err error) {
	ampersands := utils.ArrAllIndex(body, []byte("&"))
	// To get the last key-val pair in
	ampersands = append(ampersands, len(body))
	startIndex := 0
	for _, endIndex := range ampersands {
		_, err := utils.PercentDecode(body[startIndex:endIndex])
		if err != nil {
			return err
		}
		startIndex = endIndex + 1
	}
	return nil
}

func parseURLEncodedForm(body []byte) (jsonBody []byte, err error) {
	ampersands := utils.ArrAllIndex(body, []byte("&"))
	// To get the last key-val pair in
	ampersands = append(ampersands, len(body))
	var ampersandSeparatedPairs [][]byte
	startIndex := 0
	for _, endIndex := range ampersands {
		keyValue, _ := utils.PercentDecode(body[startIndex:endIndex])
		ampersandSeparatedPairs = append(ampersandSeparatedPairs, keyValue)
		startIndex = endIndex + 1
	}

	keyValuePairs := make(map[string]string)
	for _, pair := range ampersandSeparatedPairs {
		utils.Mapify(keyValuePairs, pair, []byte("="))
	}

	// Adding keys like tags[0]="a", tags[1]="b" to tags= {"a", "b"}
	listKeys := make(map[string][]string)
	for key, val := range keyValuePairs {
		if idx := utils.ArrIndex([]byte(key), []byte("[")); idx != -1 {
			listKey := key[:idx]
			delete(keyValuePairs, key)
			if arr, ok := listKeys[listKey]; !ok {
				listKeys[listKey] = []string{val}
			} else {
				listKeys[listKey] = append(arr, val)
			}
		}
	}

	combinedMap := make(map[string]interface{})
	for key, val := range keyValuePairs {
		combinedMap[key] = val
	}
	for key, val := range listKeys {
		combinedMap[key] = val
	}

	// Convert to JSON
	jsonForm, err := json.Marshal(combinedMap)
	if err != nil {
		return nil, fmt.Errorf("while marshaling: %s", err.Error())
	}
	return jsonForm, nil
}
