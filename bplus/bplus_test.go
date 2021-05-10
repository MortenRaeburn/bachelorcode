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

	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9} //fanout = 3

	tree, _ := NewTree(data, 4, sumOfSlice, Identity) //Has 4 children

	tree = insertLabels(tree)

	fmt.Println(tree)
}

func TestSearch(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 100, 34, 29} //fanout = 3

	tree, _ := NewTree(data, 3, sumOfSlice, Identity)

	fmt.Println(tree)
}

func TestCount(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	tree, _ := NewTree(data, 3, sumOfSlice, One)

	vals := tree.Root.Values
	sum := sumOfSlice(vals...)
	if sum != 9 {
		t.Errorf("TestCount failed, expected 9 but got %d", sum)
	}
}

func TestSum(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	tree, _ := NewTree(data, 3, sumOfSlice, Identity)

	vals := tree.Root.Values
	sum := sumOfSlice(vals...)
	if sum != 45 {
		t.Errorf("TestCount failed, expected 45 but got %d", sum)
	}
}

//Auxilary Functions:

func sumOfSlice(i ...int) int {
	res := 0
	for _, x := range i {
		res += x
	}
	return res
}

func One(i int) int {
	return 1
}

func Identity(i int) int {
	return i
}

func minOfSlize(i ...int) int {
	min := int(^uint(0) >> 1) //maxvalue
	for _, s := range i {
		if s < min {
			min = s
		}
	}
	return min
}
