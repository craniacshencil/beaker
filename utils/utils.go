package utils

import (
	"fmt"
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
Takes an arr and subarr and returns indcies for
all occurences of the first element of subarr.
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
