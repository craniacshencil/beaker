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
