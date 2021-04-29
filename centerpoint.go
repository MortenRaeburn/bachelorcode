package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
)

var centerpoint_url string = "http://127.0.0.1:5000/centerpoint"

type center_res struct {
	L *line
	U *line
	D *line
	R *line
}

func (cr *center_res) addDirAndSign() {
	cr.L = NewLine(cr.L.M, cr.L.B, 0)
	cr.U = NewLine(cr.U.M, cr.U.B, 1)
	cr.D = NewLine(cr.D.M, cr.D.B, 2)
	cr.R = NewLine(cr.R.M, cr.R.B, 3)
}

func centerpoint(ps [][2]float64) *center_res {
	json_data, err := json.Marshal(ps)

	if err != nil {
		panic(err)
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

	res := new(center_res)

	err = json.Unmarshal(bodyBytes, res)

	if err != nil {
		panic(err)
	}

	res.addDirAndSign()

	return res
}

func main() {}

func AuthCenterpoint(ps [][2]float64, rt *Rtree) *VOCenter {

	pruneVOs := []*VOPrune{}

	for {
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

func VerifyCenterpoint(digest []byte, initSize int, vo *VOCenter, f int) ([][2]float64, bool) {
	size := initSize

	for _, pruneVO := range vo.Prunes {
		lContains := verifyHalfSpace(size, pruneVO.L, pruneVO.LCount, digest, f)
		uContains := verifyHalfSpace(size, pruneVO.U, pruneVO.UCount, digest, f)
		dContains := verifyHalfSpace(size, pruneVO.D, pruneVO.DCount, digest, f)
		rContains := verifyHalfSpace(size, pruneVO.R, pruneVO.RCount, digest, f)

		if !lContains || !uContains || !dContains || !rContains {
			return nil, false
		}

		for _, countVOs := range pruneVO.Prune {

			var LU, LD, RU, RD *VOCount

			sibs := []*Node{}

			for _, countVO := range countVOs {
				if len(countVO.Mcs) != 1 {
					return nil, false
				}

				containsLU := cornerContains(pruneVO.L, pruneVO.U, countVO.Mcs[0].MBR)
				containsLD := cornerContains(pruneVO.L, pruneVO.D, countVO.Mcs[0].MBR)
				containsRU := cornerContains(pruneVO.R, pruneVO.U, countVO.Mcs[0].MBR)
				containsRD := cornerContains(pruneVO.R, pruneVO.D, countVO.Mcs[0].MBR)

				switch true {
				case containsLU:
					LU = countVO
				case containsLD:
					LD = countVO
				case containsRU:
					RU = countVO
				case containsRD:
					RD = countVO
				default:
					return nil, false
				}

				count, valid := AuthCountVerify(countVO, digest, f)

				if !valid || count != 1 {
					return nil, false
				}

				sibs = append(sibs, countVO.Sib...)
			}

			if LU == nil || LD == nil || RU == nil || RD == nil {
				return nil, false
			}

			ls := divideByLabel(sibs)
			roots := verifyLayers(ls, f)

			if len(roots) != 1 {
				panic("Roots should always be len 1")
			}

			digest = roots[0].Hash

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

func cornerContains(l1, l2 *line, mbr [4]float64) bool {
	contains1 := containsHalfSpace(l1, mbr)
	contains2 := containsHalfSpace(l2, mbr)

	if !contains1 || !contains2 {
		return false
	}
	return true
}

func verifyHalfSpace(size int, l *line, vo *VOCount, digest []byte, f int) bool {
	for _, n := range vo.Mcs {
		if !containsHalfSpace(l, n.MBR) {
			return false
		}
	}

	count, valid := AuthCountVerify(vo, digest, f)

	if !valid {
		return false
	}

	if (size+2)/3-1 >= count {
		return false
	}

	return true
}

func prune(ps [][2]float64, rt *Rtree) (*VOPrune, *Rtree, [][2]float64, bool) {
	center := centerpoint(ps)

	vo := new(VOPrune)
	vo.L = center.L
	vo.U = center.U
	vo.D = center.D
	vo.R = center.R

	vo.LCount = rt.AuthCountHalfSpace(center.L)
	vo.UCount = rt.AuthCountHalfSpace(center.U)
	vo.DCount = rt.AuthCountHalfSpace(center.D)
	vo.RCount = rt.AuthCountHalfSpace(center.R)

	LU := [][2]float64{}
	LD := [][2]float64{}
	RU := [][2]float64{}
	RD := [][2]float64{}

	_ps := ps

	for i, p := range ps {
		mbr := [4]float64{
			p[0],
			p[1],
			p[0],
			p[1],
		}

		found := false

		switch true {
		case cornerContains(center.L, center.U, mbr):
			found = true
			LU = append(LU, p)
		case cornerContains(center.L, center.D, mbr):
			found = true
			LD = append(LD, p)
		case cornerContains(center.R, center.U, mbr):
			found = true
			RU = append(RU, p)
		case cornerContains(center.R, center.D, mbr):
			found = true
			RD = append(RD, p)
		}

		if !found {
			continue
		}

		_ps = append(_ps[:i], _ps[i+1:]...)
	}

	iterated := false

	for {
		if len(LU) == 0 || len(LD) == 0 || len(RU) == 0 || len(RD) == 0 {
			break
		}

		ps = _ps
		iterated = true

		var lu, ld, ru, rd [2]float64
		lu, LU = LU[0], LU[1:]
		ld, LD = LD[0], LD[1:]
		ru, RU = RU[0], RU[1:]
		rd, RD = RD[0], RD[1:]

		prune := [4]*VOCount{
			rt.AuthCountPoint(lu),
			rt.AuthCountPoint(ld),
			rt.AuthCountPoint(ru),
			rt.AuthCountPoint(rd),
		}

		vo.Prune = append(vo.Prune, prune)

		radon := calcRadon([2][2][2]float64{
			{lu, rd},
			{ld, ru},
		})

		ps = append(ps, radon)
	}

	rt, err := NewRTree(ps, rt.Fanout, rt.Root.Agg, rt.Root.AggLeaf)

	if err != nil {
		panic(err)
	}

	return vo, rt, ps, iterated

}

func calcRadon(pss [2][2][2]float64) [2]float64 {
	aux := func(ps [2][2]float64) *line {
		if ps[0][0] > ps[1][0] {
			ps[0], ps[1] = ps[1], ps[0]
		}

		l := new(line)
		l.M = (ps[1][1] - ps[0][1]) / (ps[1][0] - ps[0][0])
		l.B = ps[0][1] - l.M*ps[0][0]

		return l
	}

	l1 := aux(pss[0])
	l2 := aux(pss[1])

	x := (l2.B - l1.B) / (l1.M - l2.M)

	return linePoint(l1, x)
}

func linePoint(l *line, x float64) [2]float64 {
	return [2]float64{x, l.M*x + l.B}
}

func filter(l *line, ps [][2]float64) [][2]float64 {
	filterPs := [][2]float64{}

	for _, p := range ps {
		lp := linePoint(l, p[0])

		if lp[1] < p[1] != l.Sign {
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

// func diff(ps1, ps2 [][2]float64) ([][2]float64, [][2]float64) {
// 	newPs1 := ps1 // TODO May need actual deep cloning depending on slice behavior
// 	newPs2 := ps2

// 	for i, p := range ps1 {
// 		j, found := pointSearch(newPs2, p)

// 		if !found {
// 			continue
// 		}

// 		newPs1 = append(newPs1[:i], newPs1[i+1:]...)
// 		newPs2 = append(newPs2[:j], newPs2[j+1:]...)
// 	}

// 	return newPs1, newPs2
// }
