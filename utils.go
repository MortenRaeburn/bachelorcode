package main

var eps float64 = 0.00000001

type line struct {
	B float64
	M float64
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
func halfSpaceSign(m float64, dir int) bool {
	signLookup := map[bool]map[int]bool{
		true: {
			0: true,
			1: true,
			2: false,
			3: false,
		},
		false: {
			0: false,
			1: true,
			2: false,
			3: true,
		},
	}

	mPositive := true
	if m < 0 {
		mPositive = false
	}

	return signLookup[mPositive][dir]
}

type VOCenter struct {
	LMcs, UMcs, DMcs, RMcs []*Node
	LSib, USib, DSib, RSib map[string][]byte

	PMcss [][]*Node
	PSibs []map[string][]byte
}

type VOFinal struct {
	PMcss [][]*Node
	PSibs []map[string][]byte
}