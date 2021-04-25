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

type center_res struct {
	L  *line
	U  *line
	R  *line
	D  *line
	PS [][2]float64
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
	authCenterpoint()
}

func authCenterpoint() ([]*VOCenter, *VOFinal) {
	rand.Seed(time.Now().UTC().UnixNano())

	ps := [][2]float64{}

	for i := 0; i < 100; i++ {
		x := rand.Float64()*200 - 100
		y := rand.Float64()*200 - 100

		ps = append(ps, [2]float64{x, y})
	}

	rt, err := NewRTree(ps, 3, sumOfSlice, one)

	if err != nil {
		panic(err)
	}

	centerVOs := []*VOCenter{}

	for true {
		vo, newRt, newPs, pruning := prune(ps, rt)

		if !pruning {
			break
		}

		centerVOs = append(centerVOs, vo)
		rt = newRt
		ps = newPs
	}

	finalVO := new(VOFinal)

	finalVO.PMcss, finalVO.PSibs = rt.AuthCountPoints(ps)

	return centerVOs, finalVO

}

func prune(ps [][2]float64, rt *Rtree) (*VOCenter, *Rtree, [][2]float64, bool) {
	center := centerpoint(ps)

	delPs, _ := diff(ps, center.PS)

	if len(delPs) == 0 {
		return nil, nil, ps, false
	}

	lSign := halfSpaceSign(center.L.M, 0)
	uSign := halfSpaceSign(center.U.M, 1)
	dSign := halfSpaceSign(center.D.M, 2)
	rSign := halfSpaceSign(center.R.M, 3)

	vo := new(VOCenter)
	vo.LMcs, vo.LSib = rt.AuthCountHalfSpace(center.L, lSign)
	vo.UMcs, vo.USib = rt.AuthCountHalfSpace(center.U, uSign)
	vo.DMcs, vo.DSib = rt.AuthCountHalfSpace(center.D, dSign)
	vo.RMcs, vo.RSib = rt.AuthCountHalfSpace(center.R, rSign)

	vo.PMcss, vo.PSibs = rt.AuthCountPoints(delPs)

	newRt, err := NewRTree(center.PS, rt.Fanout, rt.Root.Agg, rt.Root.AggLeaf)

	if err != nil {
		panic(err)
	}

	return vo, newRt, center.PS, true

}

func linePoint(l *line, x float64) [2]float64 {
	return [2]float64{x, l.M*x + l.B}
}

func filter(l *line, ps [][2]float64, sign bool) [][2]float64 {
	filterPs := [][2]float64{}

	for _, p := range ps {
		lp := linePoint(l, p[0])

		if lp[1] < p[1] != sign {
			continue
		}

		filterPs = append(filterPs, lp)
	}

	return filterPs
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
	newPs1 := ps1 // TODO May need actual deep cloning depending on slice behavior
	newPs2 := ps2

	for i, p := range newPs1 {
		j, found := pointSearch(newPs2, p)

		if !found {
			continue
		}

		newPs1 = append(newPs1[:i], newPs1[i+1:]...)
		newPs2 = append(newPs2[:j], newPs2[j+1:]...)
	}

	return newPs1, newPs2
}
