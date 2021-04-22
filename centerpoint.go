package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"time"
)

var centerpoint_url string = "http://127.0.0.1:5000/centerpoint"
var eps float64 = 0.00000001

type center_res struct {
	L  line
	U  line
	R  line
	D  line
	PS [][2]float64
}

type line struct {
	B float64
	M float64
}

func centerpoint(ps [][2]float64) *center_res {
	json_data, err := json.Marshal(ps)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(centerpoint_url, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	res := &center_res{}

	err = json.Unmarshal(bodyBytes, res)

	if err != nil {
		panic(err)
	}

	return res
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	ps := [][2]float64{}

	for i := 0; i < 100; i++ {
		x := rand.Float64()*200 - 100
		y := rand.Float64()*200 - 100

		ps = append(ps, [2]float64{x, y})
	}

	center := centerpoint(ps)

	diff(ps, center.PS)
}

func pointsSort(ps [][2]float64) [][2]float64 {
	less := func(i, j int) bool {
		return ps[i][0] < ps[j][0] && ps[i][1] < ps[j][1]
	}

	sort.Slice(ps, less)

	return ps
}

func pointEqual(p1, p2 [2]float64) bool {
	return math.Abs(p1[0]-p2[0]) < eps && math.Abs(p1[1]-p2[1]) < eps
}

func pointSearch(ps [][2]float64, x [2]float64) (int, bool) {
	for i, p := range ps {
		xEq := math.Abs(p[0]-x[0]) < eps
		yEq := math.Abs(p[1]-x[1]) < eps

		if !xEq || !yEq {
			continue
		}

		return i, true
	}

	return -1, false
}

func diff(ps1, ps2 [][2]float64) ([][2]float64, [][2]float64) {
	for i, p := range ps1 {
		j, found := pointSearch(ps2, p)

		if !found {
			continue
		}

		ps1 = append(ps1[:i], ps1[i+1:]...)
		ps2 = append(ps2[:j], ps2[j+1:]...)
	}

	return ps1, ps2
}
