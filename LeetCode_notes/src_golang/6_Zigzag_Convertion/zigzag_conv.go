package main

import (
	"fmt"
)

func main() {

	s := "PAYPALISHIRING"
	fmt.Println(zigzag_convert(s, 4))
}

func zigzag_convert(s string, numRows int) string {
	if numRows == 1 {
		return s
	}
	var str string
	tmp := make([]string, numRows)
	for i, c := range s {
		j := i % (numRows*2 - 2)
		if j < numRows {
			tmp[j] += string(c)
		} else {
			tmp[numRows*2-2-j] += string(c)
		}
	}

	for _, str1 := range tmp {
		str += str1
	}
	return str
}
