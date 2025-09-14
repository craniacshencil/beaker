package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"slices"
)

/*
Takes an arr and subarr and returns the index of
first occurence of the first element of subarr.
- arr[:index] will not include the subarr
*/
func ArrIndex(arr []byte, subarr []byte) int {
	subarrLen := len(subarr)
	for index := range arr {
		end := index + subarrLen
		if end > len(arr) {
			break
		}
		if slices.Equal(arr[index:end], subarr) {
			return index
		}
	}
	return -1
}

/*
Takes an arr and subarr and returns the index of
last occurence of the first element of subarr.
- arr[:index] will not include the subarr
*/
func ArrLastIndex(arr []byte, subarr []byte) int {
	var arrLen, subarrLen int
	arrLen = len(arr)
	subarrLen = len(subarr)
	for index := arrLen - subarrLen; index >= 0; index-- {
		if slices.Equal(arr[index:index+subarrLen], subarr) {
			return index
		}
	}
	return -1
}

/*
Takes an arr and subarr and returns indcies for all occurences of the first element of subarr.
  - special case ignored: if subarr is {10, 10} arr is {10, 10, 10, 10}
    indices returned would be {0, 2}
  - as illustrated above it doesn't return interleaved
    subarr, index jumps to the end of the subarr
*/
func ArrAllIndex(arr []byte, subarr []byte) []int {
	startIndex := 0
	lastIndex := len(arr)
	appearances := []int{}
	for startIndex < lastIndex {
		foundAt := ArrIndex(arr[startIndex:], subarr)
		if foundAt != -1 {
			// Gives absolute indexing instead of relative indexing
			absoluteIndex := foundAt + startIndex
			appearances = append(appearances, absoluteIndex)
			startIndex = absoluteIndex + len(subarr)
		} else {
			break
		}
	}
	return appearances
}

func tests() {
	fmt.Println(
		ArrIndex(
			[]byte{12, 14, 15, 13, 10, 12, 1, 5, 13, 10, 11, 14, 13},
			[]byte{47, 114, 47, 110},
		),
	)
	fmt.Println(
		ArrLastIndex(
			[]byte{12, 14, 15, 13, 10, 12, 1, 5, 13, 10, 11, 14, 13},
			[]byte{47, 114, 47, 110},
		),
	)
}

/*
Takes an arr and converts each entry into a string

For testing mostly
*/
func StringifyByteArray(arr [][]byte) {
	var res []string
	for _, val := range arr {
		res = append(res, string(val))
	}
	log.Println(res)
}

/*
Takes a pair, separated by a separator. Separates into key and value and adds to resMap
gap is a variable to account for length of separator
*/
func Mapify(
	resMap map[string]string,
	pair, separator []byte,
) (err error) {
	gap := len(separator)
	separatorIdx := ArrIndex(pair, separator)
	if separatorIdx == -1 {
		return errors.New("Invalid header key-value pair, no separator found")
	}
	// +gap to get rid of separator
	key, value := pair[:separatorIdx], pair[separatorIdx+gap:]
	resMap[string(key)] = string(value)
	return nil
}

func PercentDecode(input []byte) ([]byte, error) {
	var decoded []byte
	for i := 0; i < len(input); {
		if input[i] == '%' {
			if i+2 >= len(input) {
				return nil, errors.New("invalid percent-encoding, beyond length")
			}
			hexDigits := input[i+1 : i+3]

			b := make([]byte, 1)
			_, err := hex.Decode(b, hexDigits)
			if err != nil {
				return nil, fmt.Errorf("invalid percent-encoding: %s", hexDigits)
			}

			decoded = append(decoded, b[0])
			i += 3 // skip "%XX"
		} else {
			if input[i] == byte('+') {
				input[i] = byte(' ')
			}
			decoded = append(decoded, input[i])
			i++
		}
	}

	return decoded, nil
}
