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

func authCenterpoint() *VOCenter {
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

	pruneVOs := []*VOPrune{}

	for true {
		vo, newRt, newPs, pruning := prune(ps, rt)

		if !pruning {
			break
		}

		pruneVOs = append(pruneVOs, vo)
		rt = newRt
		ps = newPs
	}

	centerVO := new(VOCenter)
	centerVO.Final = rt.AuthCountPoints(ps)
	centerVO.Prunes = pruneVOs

	return centerVO
}

func verifyCenterpoint(digest []byte, initSize int, vo *VOCenter, f int) ([][2]float64, bool) {
	size := initSize

	for _, pruneVO := range vo.Prunes {
		lContains := verifyHalfSpace(size, pruneVO.L, pruneVO.LCount, digest, 0, f)
		uContains := verifyHalfSpace(size, pruneVO.U, pruneVO.UCount, digest, 1, f)
		dContains := verifyHalfSpace(size, pruneVO.D, pruneVO.DCount, digest, 2, f)
		rContains := verifyHalfSpace(size, pruneVO.R, pruneVO.RCount, digest, 3, f)

		if !lContains || !uContains || !dContains || !rContains {
			return nil, false
		}

		LU := 0
		LD := 0
		RU := 0
		RD := 0

		for _, countVO := range pruneVO.Prune {
			if len(countVO.Mcs) != 1 {
				return nil, false
			}

			containsLU := cornerContains(pruneVO.L, pruneVO.U, 0, 1, countVO.Mcs[0].MBR)
			containsLD := cornerContains(pruneVO.L, pruneVO.D, 0, 2, countVO.Mcs[0].MBR)
			containsRU := cornerContains(pruneVO.R, pruneVO.U, 3, 1, countVO.Mcs[0].MBR)
			containsRD := cornerContains(pruneVO.R, pruneVO.D, 3, 2, countVO.Mcs[0].MBR)

			switch true {
			case containsLU:
				LU++
			case containsLD:
				LD++
			case containsRU:
				RU++
			case containsRD:
				RD++
			default:
				return nil, false
			}

			count, valid := AuthCountVerify(countVO, digest, f)

			if !valid || count != 1 {
				return nil, false
			}

			ls := divideByLabel(countVO.Sib)
			roots := verifyLayers(ls, f)

			if len(roots) != 1 {
				panic("Roots should always be len 1")
			}

			digest = roots[0].Hash

		}

		if LU != LD || LD != RU || RU != RD {
			return nil, false
		}

	}

	finalPs := [][2]float64{}

	for _, countVO := range vo.Final {
		if len(countVO.Mcs) != 1 {
			return nil, false
		}

		count, valid := AuthCountVerify(countVO, digest, f)

		if !valid || count != 1 {
			return nil, false
		}

		mbr := countVO.Mcs[0].MBR

		p := [2]float64{
			mbr[0],
			mbr[1],
		}

		finalPs = append(finalPs, p)
	}

	return finalPs, true
}

func cornerContains(l1, l2 *line, dir1, dir2 int, mbr [4]float64) bool {
	sign1 := halfSpaceSign(l1, dir1)
	sign2 := halfSpaceSign(l2, dir2)

	contains1 := containsHalfSpace(l1, mbr, sign1)
	contains2 := containsHalfSpace(l2, mbr, sign2)

	if !contains1 || !contains2 {
		return false
	}
	return true
}

func verifyHalfSpace(size int, l *line, vo *VOCount, digest []byte, dir int, f int) bool {
	for _, n := range vo.Mcs {
		sign := halfSpaceSign(l, dir)

		if !containsHalfSpace(l, n.MBR, sign) {
			return false
		}
	}

	count, valid := AuthCountVerify(vo, digest, f)

	if !valid {
		return false
	}

	if (count+2)/3-1 <= size {
		return false
	}

	return true
}

func prune(ps [][2]float64, rt *Rtree) (*VOPrune, *Rtree, [][2]float64, bool) {
	center := centerpoint(ps)

	delPs, _ := diff(ps, center.PS)

	if len(delPs) == 0 {
		return nil, nil, ps, false
	}

	lSign := halfSpaceSign(center.L, 0)
	uSign := halfSpaceSign(center.U, 1)
	dSign := halfSpaceSign(center.D, 2)
	rSign := halfSpaceSign(center.R, 3)

	vo := new(VOPrune)
	vo.L = center.L
	vo.U = center.U
	vo.D = center.D
	vo.R = center.R

	vo.LCount = rt.AuthCountHalfSpace(center.L, lSign)
	vo.UCount = rt.AuthCountHalfSpace(center.U, uSign)
	vo.DCount = rt.AuthCountHalfSpace(center.D, dSign)
	vo.RCount = rt.AuthCountHalfSpace(center.R, rSign)

	vo.Prune = rt.AuthCountPoints(delPs)

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
