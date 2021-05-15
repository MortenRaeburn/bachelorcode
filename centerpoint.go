package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

var centerpoint_url string = "http://127.0.0.1:5000/centerpoint"
var SPY *spy = &spy{}

type spy struct {
	CalcNext     int
	HalfSpaceAux int
	CountAreaAux int
	CenterTime   int64
}

func (s *spy) calcNext() {
	s.CalcNext += 1
}

func (s *spy) halfSpaceAux() {
	s.HalfSpaceAux += 1
}

func (s *spy) countAreaAux() {
	s.CountAreaAux += 1
}

func (s *spy) reset() {
	s.CalcNext = 0
	s.HalfSpaceAux = 0
	s.CountAreaAux = 0
}

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
		return nil
	}

	time.Now().UnixNano()

	res.addDirAndSign()

	return res
}

func main() {
	rand.Seed(time.Now().UnixNano())

	n := 900
	f := 3

	fs := []string{
		"1.csv",
		"2.csv",
		"3.csv",
	}
	csvs := [][][]string{
		{}, {}, {},
	}

	readCsvs(fs, &csvs)

	for {
		SPY.reset()

		ps := GeneratePoints(n, 100)

		tree, _ := NewRTree(ps, f, sumOfSlice, one)

		digest := tree.Digest

		servStart := time.Now()
		VO := AuthCenterpoint(ps, tree)
		servTime := time.Since(servStart).Milliseconds()

		n := tree.Root.Value

		finalAmount := len(VO.Final)

		for _, pruneVO := range VO.Prunes {
			lMcs := pruneVO.LCount.Mcs
			lSib := pruneVO.LCount.Sib

			uMcs := pruneVO.UCount.Mcs
			uSib := pruneVO.UCount.Sib

			dMcs := pruneVO.DCount.Mcs
			dSib := pruneVO.DCount.Sib

			rMcs := pruneVO.RCount.Mcs
			rSib := pruneVO.RCount.Sib

			res2 := []string{
				strconv.Itoa(n),
				strconv.Itoa(f),
				strconv.Itoa(len(lMcs)),
				strconv.Itoa(len(lSib)),
				strconv.Itoa(len(uMcs)),
				strconv.Itoa(len(uSib)),
				strconv.Itoa(len(dMcs)),
				strconv.Itoa(len(dSib)),
				strconv.Itoa(len(rMcs)),
				strconv.Itoa(len(rSib)),
				strconv.Itoa(len(pruneVO.Prune)),
			}

			csvs[1] = append(csvs[1], res2)

			for _, countVOs := range pruneVO.Prune {
				for _, countVO := range countVOs {
					mcs := countVO.Mcs
					sib := pruneVO.LCount.Sib

					res3 := []string{
						strconv.Itoa(n),
						strconv.Itoa(f),
						strconv.Itoa(len(mcs)),
						strconv.Itoa(len(sib)),
					}

					csvs[2] = append(csvs[2], res3)
				}
			}
		}

		clientStart := time.Now()
		VerifyCenterpoint(digest, len(ps), VO, tree.Fanout)
		clientTime := time.Since(clientStart).Milliseconds()

		res1 := []string{
			strconv.Itoa(n),
			strconv.Itoa(f),
			strconv.Itoa(SPY.CalcNext),
			strconv.Itoa(SPY.CountAreaAux),
			strconv.Itoa(SPY.HalfSpaceAux),
			strconv.FormatInt(SPY.CenterTime, 10),
			strconv.FormatInt(servTime, 10),
			strconv.FormatInt(clientTime, 10),
			strconv.Itoa(finalAmount),
		}

		csvs[0] = append(csvs[0], res1)

		writeCsvs(fs, csvs)
	}
}

func writeCsvs(fs []string, csvs [][][]string) {
	for i := range fs {
		f, err := os.Create(fs[i])

		if err != nil {
			panic("Failed to write to: " + fs[i])
		}

		w := csv.NewWriter(f)

		w.WriteAll(csvs[i])
	}
}

func readCsvs(fs []string, csvs *[][][]string) {
	for i := range fs {
		f, err1 := os.Open(fs[i])

		if err1 != nil {
			continue
		}

		r1 := csv.NewReader(f)

		var err error
		(*csvs)[i], err = r1.ReadAll()

		if err != nil {
			panic("Failed to read: " + fs[i])
		}
	}
}

func AuthCenterpoint(ps [][2]float64, rt *Rtree) *VOCenter {
	pruneVOs := []*VOPrune{}

	for {
		vo, newRt, newPs, pruning := prune(ps, *rt)

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
			LU := countVOs[0]
			LD := countVOs[1]
			RU := countVOs[2]
			RD := countVOs[3]

			minNs := []*Node{}

			for _, countVO := range countVOs {
				if len(countVO.Mcs) != 1 {
					return nil, false
				}

				count, valid := AuthCountVerify(countVO, digest, f)

				if !valid || count != 1 {
					return nil, false
				}

				minNs = append(minNs, countVO.Mcs...)
				minNs = append(minNs, countVO.Sib...)
			}

			containsLU := cornerContains(pruneVO.L, pruneVO.U, LU.Mcs[0].MBR)
			containsLD := cornerContains(pruneVO.L, pruneVO.D, LD.Mcs[0].MBR)
			containsRU := cornerContains(pruneVO.R, pruneVO.U, RU.Mcs[0].MBR)
			containsRD := cornerContains(pruneVO.R, pruneVO.D, RD.Mcs[0].MBR)

			if !containsLU || !containsLD || !containsRU || !containsRD {
				return nil, false
			}

			minNs = dedupNodes(minNs)
			ls := divideByLabel(minNs)
			roots := verifyLayers(ls, f)

			root := roots[0]

			if len(root.Hash) != len(digest) {
				return nil, false
			}

			for i := range digest {
				if digest[i] != root.Hash[i] {
					return nil, false
				}
			}

			var lu, ld, ru, rd [2]float64
			copy(lu[:], LU.Mcs[0].MBR[:])
			copy(ld[:], LD.Mcs[0].MBR[:])
			copy(ru[:], RU.Mcs[0].MBR[:])
			copy(rd[:], RD.Mcs[0].MBR[:])

			radon := calcRadon(lu, ld, ru, rd)

			radonN := createLeaf(radon, one, sumOfSlice)
			radonN.Label = LU.Mcs[0].Label

			r1 := root.replace(LU.Mcs[0], radonN)
			r2 := root.remove(LD.Mcs[0])
			r3 := root.remove(RU.Mcs[0])
			r4 := root.remove(RD.Mcs[0])

			if !r1 || !r2 || !r3 || !r4 {
				panic("Removal/Replacement process failed!")
			}

			digest = root.Hash
			size = root.Value
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

func dedupNodes(ns []*Node) []*Node {
	res := []*Node{}

	nsMap := map[string]*Node{}

	for _, n := range ns {
		i := nsMap[n.Label]
		_ = i
		nsMap[n.Label] = n
	}

	for _, n := range nsMap {
		res = append(res, n)
	}

	return res
}

func cornerContains(l1, l2 *line, mbr [4]float64) bool {
	contains1 := intersectsHalfSpace(l1, mbr)
	contains2 := intersectsHalfSpace(l2, mbr)

	if contains1 || contains2 {
		return false
	}
	return true
}

func verifyHalfSpace(size int, l *line, vo *VOCount, digest []byte, f int) bool {
	for i, n := range vo.Mcs {
		_ = i

		if !containsHalfSpace(l, n.MBR) {
			return false
		}
	}

	count, valid := AuthCountVerify(vo, digest, f)

	if !valid {
		return false
	}

	_ = count
	// if (size+2)/3 - 2 > count {
	// 	return false
	// }

	return true
}

func prune(ps [][2]float64, rt Rtree) (*VOPrune, *Rtree, [][2]float64, bool) {
	start := time.Now()

	center := centerpoint(ps)

	SPY.CenterTime = time.Since(start).Milliseconds()

	if center == nil {
		return nil, &rt, ps, false
	}

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

	for _, p := range ps {
		mbr := [4]float64{
			p[0],
			p[1],
			p[0],
			p[1],
		}

		var dest *[][2]float64

		if cornerContains(center.L, center.U, mbr) {
			dest = &LU
		}

		if cornerContains(center.L, center.D, mbr) {
			if dest == nil || len(*dest) > len(LD) {
				dest = &LD
			}
		}

		if cornerContains(center.R, center.U, mbr) {
			if dest == nil || len(*dest) > len(RU) {
				dest = &RU
			}
		}

		if cornerContains(center.R, center.D, mbr) {
			if dest == nil || len(*dest) > len(RD) {
				dest = &RD
			}
		}

		if dest == nil {
			continue
		}

		*dest = append(*dest, p)

		i, found := pointSearch(_ps, p)

		if !found {
			panic("Something went very wrong")
		}

		_ps[i] = _ps[len(_ps)-1]
		_ps = _ps[:len(_ps)-1]
	}

	done := func(LU, LD, RU, RD [][2]float64) bool {
		return len(LU) == 0 || len(LD) == 0 || len(RU) == 0 || len(RD) == 0
	}

	if done(LU, LD, RU, RD) {
		return nil, &rt, ps, false
	}

	ps = _ps

	for {
		if done(LU, LD, RU, RD) {
			break
		}

		var lu, ld, ru, rd [2]float64
		lu, LU = LU[0], LU[1:]
		ld, LD = LD[0], LD[1:]
		ru, RU = RU[0], RU[1:]
		rd, RD = RD[0], RD[1:]

		luN := rt.Search([4]float64{
			lu[0],
			lu[1],
			lu[0],
			lu[1],
		})[0]

		ldN := rt.Search([4]float64{
			ld[0],
			ld[1],
			ld[0],
			ld[1],
		})[0]

		ruN := rt.Search([4]float64{
			ru[0],
			ru[1],
			ru[0],
			ru[1],
		})[0]

		rdN := rt.Search([4]float64{
			rd[0],
			rd[1],
			rd[0],
			rd[1],
		})[0]

		prune := [4]*VOCount{
			rt.AuthCountPoint(lu),
			rt.AuthCountPoint(ld),
			rt.AuthCountPoint(ru),
			rt.AuthCountPoint(rd),
		}

		vo.Prune = append(vo.Prune, prune)

		radon := calcRadon(lu, ld, ru, rd)
		ps = append(ps, radon)
		radonN := createLeaf(radon, one, sumOfSlice)
		radonN.Label = luN.Label

		rt.Root.replace(luN, radonN)
		rt.Root.remove(ldN)
		rt.Root.remove(ruN)
		rt.Root.remove(rdN)

		rt.Digest = rt.Root.Hash
	}

	return vo, &rt, ps, true

}

func calcRadon(lu, ld, ru, rd [2]float64) [2]float64 {
	ps := [][2]float64{lu, ld, ru, rd}
	// hull := openConvexHull(ps)
	hull := [][2]float64{}
	//open convex hull please

	//case 2 (see paper)
	if len(hull) == 1 {
		return hull[0]
	}

	//case 3
	if len(hull) == 0 {
		radon := [2]float64{}
		D := (lu[0]-ld[0])*(ru[1]-rd[1]) - (lu[1]-ld[1])*(ru[0]-rd[0])
		radon[0] = ((lu[0]*ld[1]-lu[1]*ld[0])*(ru[0]-rd[0]) - (lu[0]-ld[0])*(ru[0]*rd[1]-ru[1]*rd[0])) / D //works
		radon[1] = ((lu[0]*ld[1]-lu[1]*ld[0])*(ru[1]-rd[1]) - (lu[1]-ld[1])*(ru[0]*rd[1]-ru[1]*rd[0])) / D

		radon[0] = roundFloat(radon[0], eps)
		radon[1] = roundFloat(radon[1], eps)

		return radon
	}

	//case 1
	ps = pointsSort(ps)

	return ps[1]

}

func drawLine(p1, p2 [2]float64) *line {
	l := new(line)

	l.M = (p1[1] - p2[1]) / (p1[0] - p2[0])

	l.B = p1[1] - l.M*p1[0]
	l.Dir = 0
	l.Sign = halfSpaceSign(l)

	return l
}

func openConvexHull(ps [][2]float64) [][2]float64 {
	//TODO
	return [][2]float64{}

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

		filterPs = append(filterPs, p)
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
