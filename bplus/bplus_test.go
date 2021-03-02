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

	data := []int{1, 3, 5, 7, 9, 11, 13, 15}

	tree, _ := NewTree(data, 2)

	tree.Print()

	// assert := assert.New(t)

}
