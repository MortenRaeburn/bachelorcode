package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testStartLabel(t *testing.T) {
	fmt.Println("---" + t.Name() + "---")
	fmt.Println()
}

func testEndLabel() {
	fmt.Println()
	fmt.Println()
}

func TestHalfspaceCount(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	assert := assert.New(t)

	data := [][2]float64{
		{-3, -1},
		{1, 2},
		{3, 4},
		{5, 6},
	}

	tree, _ := NewRTree(data, 3, sumOfSlice, one)

	l := new(line)

	l.B = 0
	l.M = 0

	VO := tree.AuthCountHalfSpace(l, 1)

	valid := verifyHalfSpace(len(data), l, VO, tree.Digest, 1, tree.Fanout)

	assert.True(valid, "Should be true")

}

func TestAuthCountPoint(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	assert := assert.New(t)

	data := [][2]float64{
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

	tree, _ := NewRTree(data, 3, sumOfSlice, one)

	VO := tree.AuthCountPoint(data[3])

	res, valid := AuthCountVerify(VO, tree.Digest, 3)

	assert.Equal(1, res, "Wrong number of points")
	assert.True(valid, "Should be true")

}

func TestPositive(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	data := [][2]float64{
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
	tree, _ := NewRTree(data, 3, sumOfSlice, one)

	VO := tree.AuthCountArea([4]float64{15, 1, 20, 20})
	_ = VO
	fmt.Println(tree)

	// assert := assert.New(t)

}

func TestCount(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	data := [][2]float64{
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

	tree, _ := NewRTree(data, 3, sumOfSlice, one)
	_ = tree
	if tree.Root.Value != 12 {
		t.Errorf("TestCount failed. Expected 12, but got %d", tree.Root.Value)
	}
}

func TestCenterPointQueryPositive(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	ps := GeneratePoints()

	rt, err := NewRTree(ps, 3, sumOfSlice, one)

	if err != nil {
		panic(err)
	}

	VO := AuthCenterpoint(ps, rt)
	digest := rt.Digest
	dataSize := len(ps)

	_, valid := VerifyCenterpoint(digest, dataSize, VO, rt.Fanout)

	if !valid {
		t.Error("TestCenterPointQueryPositive failed. Expected true, but got false")
	}

}
