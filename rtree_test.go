package main

import (
	"fmt"
	"math/rand"
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

// func TestHalfspaceCount(t *testing.T) {
// 	testStartLabel(t)
// 	defer testEndLabel()

// 	assert := assert.New(t)

// 	data := [][2]float64{
// 		{-3, -1},
// 		{1, 2},
// 		{3, 4},
// 		{5, 6},
// 	}

// 	tree, _ := NewRTree(data, 3, sumOfSlice, one)

// 	l := NewLine(0, 0, 1)

// 	VO := tree.AuthCountHalfSpace(l)

// 	valid := verifyHalfSpace(len(data), l, VO, tree.Digest, tree.Fanout)

// 	assert.True(valid, "Should be true")
// }

func TestAuthCenterPoint(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()
	assert := assert.New(t)

	rand.Seed(69)

	ps := GeneratePoints(900, 100)

	tree, _ := NewRTree(ps, 10, sumOfSlice, one)

	digest := tree.Digest

	VO := AuthCenterpoint(ps, tree)

	fmt.Println(len(VO.Final))

	_, valid := VerifyCenterpoint(digest, len(ps), VO, tree.Fanout)

	assert.True(valid)

}

// func TestCalcRadonPoint(t *testing.T) {
// 	testStartLabel(t)
// 	defer testEndLabel()

// 	assert := assert.New(t)

// 	data := [4][2]float64{
// 		{-3, 2},
// 		{2.16, -1.53},
// 		{3.04, 2.27},
// 		{-1.08, -1.61},
// 	}

// 	l := drawLine(data[1], data[2])

// 	f := filter(l, data[:])
// 	_ = f

// 	radon := calcRadon(data[0], data[1], data[2], data[3])

// 	print(radon[0], radon[1])

// 	assert.True(math.Abs(radon[0]-0.33) < eps && math.Abs(radon[1]-(-0.28)) < eps)
// }

// func TestCalcRadonPointSimple(t *testing.T) {
// 	testStartLabel(t)
// 	defer testEndLabel()

// 	assert := assert.New(t)

// 	data := [4][2]float64{
// 		{0, 0},
// 		{2, 2},
// 		{0, 2},
// 		{2, 0},
// 	}

// 	radon := calcRadon(data[0], data[1], data[2], data[3])

// 	assert.Equal([2]float64{1, 1}, radon)
// }

// func TestHalfspaceCountTwoNegative(t *testing.T) {
// 	testStartLabel(t)
// 	defer testEndLabel()

// 	assert := assert.New(t)

// 	data := [][2]float64{
// 		{-3, 1},
// 		{3, -4},
// 		{5, -6},
// 		{5, -5},
// 		{5, -8},
// 		{3, -6},
// 		{5, -6},
// 		{1, -2},
// 		{5, -622},
// 	}

// 	tree, _ := NewRTree(data, 3, sumOfSlice, one)

// 	l := NewLine(1, 0, 0)

// 	VO := tree.AuthCountHalfSpace(l)

// 	valid := verifyHalfSpace(len(data), l, VO, tree.Digest, tree.Fanout)

// 	assert.False(valid, "Should be false")

// }

func TestHalfspaceCountTwo(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	assert := assert.New(t)

	data := [][2]float64{
		{-3, 1},
		{-1, 2},
		{3, -4},
		{5, -6},
	}

	tree, _ := NewRTree(data, 3, sumOfSlice, one)

	l := NewLine(1, 0, 1)

	VO := tree.AuthCountHalfSpace(l)

	valid := verifyHalfSpace(len(data), l, VO, tree.Digest, tree.Fanout)

	assert.True(valid, "Should be true")

}

func TestCountArea(t *testing.T) {
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

	area := [4]float64{5, 12, 11, 6}

	VO := tree.AuthCountArea(area)

	res, valid := AuthCountVerify(VO, tree.Digest, tree.Fanout)

	assert.Equal(4, res, "Should be 4")
	assert.True(valid, "Should be true")

}

// func TestHalfspaceCountNegative(t *testing.T) {
// 	testStartLabel(t)
// 	defer testEndLabel()

// 	assert := assert.New(t)

// 	data := [][2]float64{
// 		{1, 2},
// 		{-3, -4},
// 		{-5, -6},
// 		{-4, -2},
// 		{-5, -2},
// 		{-35, -2},
// 		{-5, -23},
// 	}

// 	tree, _ := NewRTree(data, 3, sumOfSlice, one)

// 	l := NewLine(0, 0, 1)

// 	VO := tree.AuthCountHalfSpace(l)

// 	valid := verifyHalfSpace(len(data), l, VO, tree.Digest, tree.Fanout)

// 	assert.False(valid, "Should be false")

// }

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