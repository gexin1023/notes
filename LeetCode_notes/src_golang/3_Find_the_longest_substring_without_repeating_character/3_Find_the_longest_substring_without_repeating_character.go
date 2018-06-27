package main

import (
	"fmt"
)

func main() {

	s := string("dvdf")

	fmt.Println(lengthOfLongestSubstring(s))
}

func lengthOfLongestSubstring(s string) int {

	/*
	 * ret：最终结果
	 * index_0: 当前字符串的起始位置
	 */
	var ret, index_0 int
	m := make(map[rune]int)

	for i, v := range s {
		j, ok := m[v]
		if ok && (j >= index_0) {
			// 存在重复字符
			if (i - index_0) > ret {
				ret = i - index_0
			}
			index_0 = j + 1
			delete(m, v)
		}

		m[v] = i
	}

	if len(s)-index_0 > ret {
		ret = len(s) - index_0
	}

	return ret
}
