## 2. Add Two Numbers

用两个非空链表分别表示两个非负整数，链表的节点表示数字的位，链表头表示数字的低位，链表尾表示数字高位。求两个链表所表示数字的和。

比如:

	Input: (2 -> 4 -> 3) + (5 -> 6 -> 4)
	Output: 7 -> 0 -> 8
	Explanation: 342 + 465 = 807.


```go
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
		if head == nil {
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
```