package webserver

import (
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
	_, err = utils.PercentDecode(path)
	if err != nil {
		return err
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
