package rtree

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

	data := [][2]int{
		[2]int{1, 2},
		[2]int{1, 2},
	}

	fmt.Println("???")
	tree, _ := NewTree(data, 3)

	fmt.Println(tree)

	// assert := assert.New(t)

}
