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
		[2]int{3, 4},
		[2]int{5, 6},
		[2]int{7, 8},
		[2]int{9, 10},
		[2]int{11, 12},
		[2]int{13, 14},
	}

	fmt.Println("???")
	tree, _ := NewTree(data, 3)

	fmt.Println(tree)

	// assert := assert.New(t)

}
