package calculation

import (
	"container/heap"
	"fmt"
	"strings"
	"sync"
)

var cache sync.Map

type HuffmanNode struct {
	Char  rune
	Freq  int
	Left  *HuffmanNode
	Right *HuffmanNode
}

type HuffmanHeep []*HuffmanNode

func (h HuffmanHeep) Len() int           { return len(h) }
func (h HuffmanHeep) Less(i, j int) bool { return h[i].Freq < h[j].Freq }
func (h HuffmanHeep) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *HuffmanHeep) Push(x interface{}) {
	*h = append(*h, x.(*HuffmanNode))
}
func (h *HuffmanHeep) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func BuildHuffmanTree(content string) *HuffmanNode {
	if content == "" {
		return nil
	}
	freqMap := make(map[rune]int)
	for _, char := range content {
		freqMap[char]++
	}

	h := &HuffmanHeep{}
	heap.Init(h)

	for char, freq := range freqMap {
		heap.Push(h, &HuffmanNode{Char: char, Freq: freq})
	}
	if h.Len() == 1 {
		node := heap.Pop(h).(*HuffmanNode)
		return &HuffmanNode{Char: node.Char, Freq: node.Freq}
	}
	for h.Len() > 1 {
		left := heap.Pop(h).(*HuffmanNode)
		right := heap.Pop(h).(*HuffmanNode)
		heap.Push(h, &HuffmanNode{
			Freq:  left.Freq + right.Freq,
			Left:  left,
			Right: right,
		})
	}
	return heap.Pop(h).(*HuffmanNode)
}

func GenerateHuffmanCodes(root *HuffmanNode) map[rune]string {
	codes := make(map[rune]string)
	if root == nil {
		return codes
	}

	if root.Left == nil && root.Right == nil {
		codes[root.Char] = "0"
		return codes
	}

	generate := func(node *HuffmanNode, code string, codes map[rune]string) {}
	generate = func(node *HuffmanNode, code string, codes map[rune]string) {
		if node.Left == nil && node.Right == nil {
			codes[node.Char] = code
			return
		}
		if node.Left != nil {
			generate(node.Left, code+"0", codes)
		}
		if node.Right != nil {
			generate(node.Right, code+"1", codes)
		}
	}

	generate(root, "", codes)
	return codes
}

func Encode(content string) (string, error) {
	if cached, ok := cache.Load(content); ok {
		return cached.(string), nil
	}

	tree := BuildHuffmanTree(content)
	codes := GenerateHuffmanCodes(tree)

	var sb strings.Builder
	for _, r := range content {
		code, ok := codes[r]
		if !ok {
			return "", fmt.Errorf("character %q not found in Huffman codes", r)
		}
		sb.WriteString(code)
	}

	encoded := sb.String()
	cache.Store(content, encoded)
	return encoded, nil
}
