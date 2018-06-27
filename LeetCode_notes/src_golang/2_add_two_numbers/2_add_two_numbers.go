package main

import (
	"fmt"
)

type ListNode struct {
	Val  int
	Next *ListNode
}

func main() {

	var num1 = []int{1}
	var num2 = []int{9, 9}

	var l1, tmp1, l2, tmp2 *ListNode

	for _, v := range num1 {
		if tmp1 == nil {
			tmp1 = new(ListNode)
			l1 = tmp1
		} else {
			tmp1.Next = new(ListNode)
			tmp1 = tmp1.Next
		}
		tmp1.Val = v
	}
	for _, v := range num2 {
		if tmp2 == nil {
			tmp2 = new(ListNode)
			l2 = tmp2
		} else {
			tmp2.Next = new(ListNode)
			tmp2 = tmp2.Next
		}
		tmp2.Val = v
	}

	result := addTwoNumbers(l1, l2)

	var ret = make([]int, 3)
	var i = 0
	for result != nil {
		ret[i] = result.Val
		i++
		result = result.Next
	}

	fmt.Println(ret)
}

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {

	var over = 0
	var head, tmp *ListNode
	for l1 != nil && l2 != nil {
		if tmp == nil {
			tmp = new(ListNode)
			head = tmp
		} else {
			tmp.Next = new(ListNode)
			tmp = tmp.Next
		}
		tmp.Val = l1.Val + l2.Val + over
		over = tmp.Val / 10
		tmp.Val -= over * 10
		l1 = l1.Next
		l2 = l2.Next
	}

	for l1 != nil {
		if head == nil {
			tmp = new(ListNode)
			head = tmp
		} else {
			tmp.Next = new(ListNode)
			tmp = tmp.Next
		}

		tmp.Val = l1.Val + over
		over = tmp.Val / 10
		tmp.Val -= over * 10

		l1 = l1.Next
	}

	for l2 != nil {

		if head == nil {
			tmp = new(ListNode)
			head = tmp
		} else {
			tmp.Next = new(ListNode)
			tmp = tmp.Next
		}

		tmp.Val = l2.Val + over
		over = tmp.Val / 10
		tmp.Val -= over * 10

		l2 = l2.Next
	}

	if over != 0 {
		if tmp == nil {
			tmp = new(ListNode)
			head = tmp
		} else {
			tmp.Next = new(ListNode)
			tmp = tmp.Next
		}
		tmp.Val = over
	}

	return head
}
