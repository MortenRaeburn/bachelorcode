package main

import "math"

var eps float64 = 0.00001

type line struct {
	B    float64
	M    float64
	Dir  int
	Sign bool
}

func NewLine(m, b float64, dir int) *line {
	l := new(line)
	l.B = b
	l.M = m
	l.Dir = dir

	l.Sign = halfSpaceSign(l)

	return l
}

func max(x, y int) int {
	if x > y {
		return x
	}

	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}

	return y
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

// dir is in the order L, U, D, and R
func halfSpaceSign(l *line) bool {
	signLookup := map[bool]map[int]bool{
		true: {
			0: false,
			1: false,
			2: true,
			3: true,
		},
		false: {
			0: true,
			1: false,
			2: true,
			3: false,
		},
	}

	mPositive := true
	if l.M < 0 {
		mPositive = false
	}

	return signLookup[mPositive][l.Dir]
}

type VOCenter struct {
	Prunes []*VOPrune
	Final  []*VOCount
}

type VOPrune struct {
	L *line
	U *line
	D *line
	R *line

	LCount *VOCount
	UCount *VOCount
	DCount *VOCount
	RCount *VOCount

	Prune [][4]*VOCount
}

type VOCount struct {
	Mcs []*Node
	Sib []*Node
}

func labelSearch(ns []*Node, l string) (*Node, int) {
	for i, n := range ns {
		if n.Label != l {
			continue
		}

		return n, i
	}

	return nil, -1
}

func roundFloat(x, prec float64) float64 {
	recPrec := 1 / prec
	return math.Floor(x*recPrec) / recPrec
}
