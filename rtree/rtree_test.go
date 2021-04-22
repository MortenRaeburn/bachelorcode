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
		{1, 2},
		{3, 4},
		{5, 6},
		{7, 8},
		{9, 10},
		{11, 12},
		{13, 14},
		{15, 16},
		{17, 18},
		{19, 20},
		{21, 22},
		{23, 24},
	}

	fmt.Println("???")
	tree, _ := NewTree(data, 3, sumOfSlice, one)

	mcs, sib := tree.AuthCount([4]int{15, 1, 20, 20})
	_ = mcs
	_ = sib

	fmt.Println(tree)

	// assert := assert.New(t)

}

func TestCount(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	data := [][2]int{
		{1, 2},
		{3, 4},
		{5, 6},
		{7, 8},
		{9, 10},
		{11, 12},
		{13, 14},
		{15, 16},
		{17, 18},
		{19, 20},
		{21, 22},
		{23, 24},
	}

	tree, _ := NewTree(data, 3, sumOfSlice, one)
	_ = tree
	if tree.Root.Value != 12 {
		t.Errorf("TestCount failed. Expected 12, but got %d", tree.Root.Value)
	}
}

func identity(i int) int {
	return i
}

func one(i int) int {
	return 1
}

func sumOfSlice(i ...int) int {
	res := 0
	for _, s := range i {
		res += s
	}
	return res
}
