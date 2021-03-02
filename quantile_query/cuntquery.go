package main

import (
	"strconv"

	"github.com/collinglass/bptree"
	"github.com/golang-collections/collections/stack"
)

var data []int

func newTree(data []int) *bptree.Tree {
	bpt := bptree.NewTree()
	for _, v := range data {
		bpt.Insert(v, []byte(strconv.Itoa(v)))
	}
	return bpt
}

func query() (int, *stack.Stack) {
	// t := bptree.NewTree()
	return 0, nil
}

func verify() bool {
	return true
}

func main() {
	data = []int{23, 24, 10, 84, 21, 48, 12}

	tree := newTree(data)
	tree.PrintTree()

}
