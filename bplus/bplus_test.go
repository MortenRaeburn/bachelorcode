package bplus

import (
	"fmt"
	"testing"
)

func testStartLabel(t *testing.T) {
	fmt.Println("---" + t.Name() + "---")
	fmt.Println()
}

func testEndLabel() {
	fmt.Println()
	fmt.Println()
}

func TestPositive(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	//data := []int{1, 2, 3, 4, 5, 6}

	//tree, _ := NewTree(data, 3)

	// ???
	//_ = tree
	fmt.Println("???")

	// tree.Print()

	// assert := assert.New(t)

}

func TestPrinting(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 100, 34, 29, 2} //fanout = 3

	tree, _ := NewTree(data, 3)

	fmt.Println(tree)
}
