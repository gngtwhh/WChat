package sensitive

import "strings"

type TrieNode struct {
	children map[rune]*TrieNode
	fail     *TrieNode
	isEnd    bool
	length   int
}

type ACFilter struct {
	root *TrieNode
}

func NewACFilter() *ACFilter {
	return &ACFilter{
		root: &TrieNode{children: make(map[rune]*TrieNode)},
	}
}

func (ac *ACFilter) Build(words []string) {
	for _, word := range words {
		cur := ac.root
		runes := []rune(word)
		for _, r := range runes {
			if _, ok := cur.children[r]; !ok {
				cur.children[r] = &TrieNode{children: make(map[rune]*TrieNode)}
			}
			cur = cur.children[r]
		}
		cur.isEnd = true
		cur.length = len(runes)
	}
	ac.buildFailPointer()
}

func (ac *ACFilter) buildFailPointer() {
	queue := []*TrieNode{}

	for _, child := range ac.root.children {
		child.fail = ac.root
		queue = append(queue, child)
	}

	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		for r, child := range parent.children {
			failNode := parent.fail
			for failNode != nil {
				if next, ok := failNode.children[r]; ok {
					child.fail = next
					break
				}
				failNode = failNode.fail
			}
			if child.fail == nil {
				child.fail = ac.root
			}
			queue = append(queue, child)
		}
	}
}

func (ac *ACFilter) Filter(text string) string {
	runes := []rune(text)
	replaceMask := make([]bool, len(runes))

	cur := ac.root
	for i, r := range runes {
		// locate r
		for cur.children[r] == nil && cur != ac.root {
			cur = cur.fail
		}
		if next, ok := cur.children[r]; ok {
			cur = next
		}

		if cur.isEnd {
			for j := 0; j < cur.length; j++ {
				replaceMask[i-j] = true
			}
		}

		// temp := cur
		// for temp != ac.root {
		// 	if temp.isEnd {
		// 		for j := 0; j < temp.length; j++ {
		// 			replaceMask[i-j] = true
		// 		}
		// 	}
		// 	temp = temp.fail
		// }
	}

	var sb strings.Builder
	for i, r := range runes {
		if replaceMask[i] {
			sb.WriteRune('*')
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
