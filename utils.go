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