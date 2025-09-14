package webserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/craniacshencil/beaker/utils"
)

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

	// Hex encoder util make using this snippet
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

func validateBody(contentType, body []byte) (err error) {
	switch {
	case slices.Equal(contentType, []byte("application/json")):
		var temp struct{}
		if err = json.Unmarshal(body, &temp); err != nil {
			return err
		}
	case slices.Equal(contentType, []byte("application/x-www-form-urlencoded")):
		if err = validateURLEncodedForm(body); err != nil {
			return err
		}
		_, err = parseURLEncodedForm(body)
		return err
	// logic
	case slices.Equal(contentType, []byte("multipart/form-data")):
	// logic
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
