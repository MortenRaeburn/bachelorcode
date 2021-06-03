package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strconv"
)

// Rtree ???
type Rtree struct {
	Root   *Node
	Fanout int
	Digest []byte
}

// Node ???
type Node struct {
	Leaf    bool
	Ps      []*Node
	Hash    []byte
	Value   int
	Label   string
	Agg     func(aggs ...int) int
	AggLeaf func(val int) int
	MBR     [4]float64
}

func (n *Node) Clone() Node {
	newN := *n
	copy(newN.Ps, n.Ps)
	copy(newN.Hash, n.Hash)

	return newN
}

// NewRTree ???
// ??? Needs a certain amount of elements to work - around fanout
// ??? elems should be entries
//TODO : compute Hash and Aggregate value of each node
func NewRTree(elems [][2]float64, fanout int, agg func(aggs ...int) int, aggLeaf func(val int) int) (*Rtree, error) {
	sort.Slice(elems, func(i, j int) bool {
		return elems[i][0] < elems[j][0]
	})

	roots := []*Node{}

	for _, e := range elems {
		n := createLeaf(e, aggLeaf, agg)
		roots = append(roots, n)
	}

	roots = createInternals(roots, fanout, agg)

	roots[0].labelMaker()

	t := new(Rtree)
	t.Fanout = fanout
	t.Root = roots[0]
	t.Digest = t.Root.Hash

	return t, nil
}

func createInternals(roots []*Node, fanout int, agg func(aggs ...int) int) []*Node {
	for len(roots) != 1 {
		temp := []*Node{}

		for i := 0; i < len(roots); i += fanout {
			j := min(i+fanout, len(roots))

			n := createInternal(roots[i:j], agg)
			temp = append(temp, n)
		}

		roots = temp
	}
	return roots
}

func (n *Node) listAux() []*Node {
	if n.Leaf {
		return []*Node{n}
	}

	list := []*Node{}

	for _, c := range n.Ps {
		list = append(list, c.listAux()...)
	}
	return list
}

func (r *Rtree) List() []*Node {
	n := r.Root
	return n.listAux()
}

func (r *Rtree) AuthCountPoints(ps [][2]float64) []*VOCount {
	vos := []*VOCount{}

	for _, p := range ps {
		vos = append(vos, r.AuthCountPoint(p))
	}

	return vos
}

// Add labels recursively
func (n *Node) labelMaker() {
	for i, c := range n.Ps {
		label := n.Label + strconv.Itoa(i)
		c.Label = label

		c.labelMaker()
	}
}

func (n *Node) remove(t *Node) bool {
	for i, child := range n.Ps {
		if child.Label == t.Label {
			n.Ps = append(n.Ps[:i], n.Ps[i+1:]...)

			n.CalcAgg()
			n.CalcMBR()
			n.CalcHash()
			return true
		}

		done := child.remove(t)

		if !done {
			continue
		}

		if len(child.Ps) == 0 {
			n.Ps = append(n.Ps[:i], n.Ps[i+1:]...)
		}

		n.CalcAgg()
		n.CalcMBR()
		n.CalcHash()
		return true
	}

	return false
}

func (n *Node) maskRemoval() {
	for i, c := range n.Ps {
		c.maskRemoval()

		if len(c.Ps) == 1 {
			label := c.Label
			n.Ps[i] = c.Ps[0]
			n.Ps[i].Label = label

			n.CalcAgg()
			n.CalcMBR()
			n.CalcHash()
		}
	}
}

func (n *Node) replace(t, s *Node) bool {
	for i, child := range n.Ps {
		if child.Label == t.Label {
			n.Ps[i] = s
			n.CalcMBR()
			n.CalcHash()
			return true
		}

		done := child.replace(t, s)

		if !done {
			continue
		}

		if len(child.Ps) == 0 {
			panic("Should never happen!")
		}

		n.CalcMBR()
		n.CalcHash()
		return true
	}

	return false
}

func createInternal(ns []*Node, agg func(aggs ...int) int) *Node {
	internal := new(Node)
	internal.Ps = []*Node{}
	internal.Agg = agg
	internal.MBR = ns[0].MBR

	internal.Ps = append(internal.Ps, ns...)

	internal.CalcAgg()
	internal.CalcMBR()
	internal.CalcHash()

	return internal
}

func (n *Node) CalcMBR() {
	if len(n.Ps) == 0 {
		return
	}

	mbr := n.Ps[0].MBR

	for _, p := range n.Ps {
		mbr[0] = math.Min(mbr[0], p.MBR[0])
		mbr[1] = math.Max(mbr[1], p.MBR[1])
		mbr[2] = math.Max(mbr[2], p.MBR[2])
		mbr[3] = math.Min(mbr[3], p.MBR[3])

		n.MBR = mbr
	}
}

func (n *Node) CalcAgg() {
	childVals := []int{}

	for i := range n.Ps {
		childVals = append(childVals, n.Ps[i].Value)
	}

	n.Value = n.Agg(childVals...)
}

func (n *Node) CalcHash() {
	hashVal := []byte{}

	for _, c := range n.Ps {
		agg := []byte(strconv.Itoa(c.Value))

		var mbr []byte
		for _, corner := range c.MBR {
			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], math.Float64bits(corner))
			mbr = append(mbr, buf[:]...)
		}

		hashVal = append(hashVal, agg...)
		hashVal = append(hashVal, mbr...)
		hashVal = append(hashVal, c.Hash...)
	}

	hash := sha256.Sum256(hashVal)
	n.Hash = hash[:]
}

func createLeaf(p [2]float64, aggLeaf func(val int) int, agg func(vals ...int) int) *Node {
	n := new(Node)
	n.MBR = [4]float64{}
	n.Leaf = true
	n.AggLeaf = aggLeaf
	n.Agg = agg
	n.Value = one(69)

	n.MBR[0] = p[0]
	n.MBR[1] = p[1]
	n.MBR[2] = p[0]
	n.MBR[3] = p[1]

	n.CalcHash()

	return n
}

// Search ???
func (t *Rtree) Search(area [4]float64) []*Node {
	return t.Root.searchAux(area)
}

func (n *Node) searchAux(area [4]float64) []*Node {
	acc := []*Node{}

	for _, c := range n.Ps {
		if !intersectsArea(area, c.MBR) {
			continue
		}

		if c.Leaf {
			acc = append(acc, c)
			continue
		}

		acc = append(acc, c.searchAux(area)...)
	}

	return acc
}

func (t *Rtree) AuthCountPoint(p [2]float64) *VOCount {
	return t.AuthCountArea([4]float64{p[0], p[1], p[0], p[1]})
}

// AuthCountArea ???
func (t *Rtree) AuthCountArea(area [4]float64) *VOCount {
	vo := t.Root.authCountAreaAux(area)

	if len(vo.Mcs) == 0 {
		panic("MCS should never be 0")
	}

	return vo
}

func (n *Node) authCountAreaAux(area [4]float64) *VOCount {
	SPY.countAreaAux()

	vo := new(VOCount)
	vo.Mcs = []*Node{}
	vo.Sib = []*Node{}

	for i, c := range n.Ps {
		_ = i

		if !intersectsArea(area, c.MBR) {
			cc := c.Clone()
			vo.Sib = append(vo.Sib, &cc)
			continue
		}

		if containsArea(area, c.MBR) {
			cc := c.Clone()
			vo.Mcs = append(vo.Mcs, &cc)
			continue
		}

		voChild := c.authCountAreaAux(area)

		vo.Mcs = append(vo.Mcs, voChild.Mcs...)
		vo.Sib = append(vo.Sib, voChild.Sib...)

	}

	return vo
}

// AuthCountHalfSpace ???
func (t *Rtree) AuthCountHalfSpace(l *line) *VOCount {
	return t.Root.authCountHalfSpaceAux(l)
}

func (n *Node) authCountHalfSpaceAux(l *line) *VOCount {
	SPY.halfSpaceAux()

	vo := new(VOCount)
	vo.Mcs = []*Node{}
	vo.Sib = []*Node{}

	for i, c := range n.Ps {
		_ = i

		if !intersectsHalfSpace(l, c.MBR, false) {
			cc := c.Clone()
			vo.Sib = append(vo.Sib, &cc)
			continue
		}

		if containsHalfSpace(l, c.MBR, true) {
			cc := c.Clone()
			vo.Mcs = append(vo.Mcs, &cc)
			continue
		}

		voChild := c.authCountHalfSpaceAux(l)

		vo.Mcs = append(vo.Mcs, voChild.Mcs...)
		vo.Sib = append(vo.Sib, voChild.Sib...)
	}

	return vo
}

// AuthCountHalfSpace ???
func (t *Rtree) AuthCountHalfSpaces(ls [][2]*line) *VOCount {
	return t.Root.authCountHalfSpacesAux(ls)
}

func (n *Node) authCountHalfSpacesAux(ls [][2]*line) *VOCount {
	SPY.halfSpaceAux()

	vo := new(VOCount)
	vo.Mcs = []*Node{}
	vo.Sib = []*Node{}

	for i, c := range n.Ps {
		_ = i

		if containsHalfSpaces(ls, c.MBR, false) {
			cc := c.Clone()
			vo.Sib = append(vo.Sib, &cc)
			continue
		}

		cornerContained := false

		for _, l := range ls {
			if !cornerContains(l[0], l[1], c.MBR) {
				continue
			}

			cornerContained = true

			cs := c.listAux()

			for _, n := range cs {
				_n := n.Clone()
				vo.Mcs = append(vo.Mcs, &_n)
			}

			break
		}

		if cornerContained {
			continue
		}

		voChild := c.authCountHalfSpacesAux(ls)

		vo.Mcs = append(vo.Mcs, voChild.Mcs...)
		vo.Sib = append(vo.Sib, voChild.Sib...)
	}

	return vo
}

func intersectsHalfSpaces(ls [][2]*line, r [4]float64, incl bool) bool {
	amounts := intersectsHalfSpacesAux(r, ls, incl)

	for _, amount := range amounts {
		if amount < 1 {
			continue
		}

		return true
	}

	return false

}

func intersectsHalfSpacesAux(r [4]float64, ls [][2]*line, incl bool) []int {
	// TODO correct int to float64 and remove conversion
	ps := [][2]float64{
		{float64(r[0]), float64(r[1])},
		{float64(r[0]), float64(r[3])},
		{float64(r[2]), float64(r[1])},
		{float64(r[2]), float64(r[3])},
	}

	res := []int{}

	for _, l := range ls {
		f0 := filter(l[0], ps, incl)
		f1 := filter(l[1], f0, incl)

		res = append(res, len(f1))
	}

	return res
}

func containsHalfSpaces(ls [][2]*line, r [4]float64, incl bool) bool {
	amounts := intersectsHalfSpacesAux(r, ls, incl)

	for _, amount := range amounts {
		if amount < 4 {
			continue
		}

		return true
	}

	return false
}

// AuthCountVerify ???
func AuthCountVerify(vo *VOCount, digest []byte, f int) (int, bool) {
	mcs := make([]*Node, len(vo.Mcs))
	sib := make([]*Node, len(vo.Sib))
	copy(mcs, vo.Mcs)
	copy(sib, vo.Sib)
	
	ns := append(mcs, sib...)
	ls := divideByLabel(ns)
	roots := verifyLayers(ls, f)

	root := roots[0]

	if len(digest) != len(root.Hash) {
		return -1, false
	}
	for i := range digest {
		if root.Hash[i] != digest[i] {
			return -1, false
		}
	}

	count := 0

	for _, mcs := range vo.Mcs {
		count += mcs.Value
	}

	return count, true
}

func verifyLayers(ls map[int][]*Node, f int) []*Node {
	ks := []int{}
	for k := range ls {
		ks = append(ks, k)
	}
	sort.Ints(ks)

	calc := []*Node{}

	last := len(ks) - 1
	for i := ks[last]; i > 0; i-- {
		l := []*Node{}
		if ls[i] != nil {
			l = append(l, ls[i]...)
		}

		l = append(l, calc...)

		calc = calcNext(l, f, i)
	}

	return calc
}

func calcNext(ns []*Node, f, h int) []*Node {
	res := []*Node{}

	nsAmount := int(math.Pow(float64(f), float64(h)))

	nsSorted := make([]*Node, nsAmount)

	for _, n := range ns {
		index := labelToString(n.Label, h, f)

		nsSorted[index] = n
	}

	last := 0

	for {
		SPY.calcNext()

		var n *Node

		for i := last; i < len(nsSorted); i++ {
			if nsSorted[i] == nil {
				continue
			}

			last = i
			n = nsSorted[i]
			break
		}

		if n == nil {
			break
		}

		parLabel := n.Label[:len(n.Label)-1]

		nextNs := nsSorted
		ss := []*Node{}
		for i := 0; i < f; i++ {
			iStr := strconv.Itoa(i)

			sLabel := parLabel + iStr
			j := labelToString(sLabel, h, f)
			sNode := nsSorted[j]

			if sNode == nil {
				continue
			}

			nextNs[j] = nil
			ss = append(ss, sNode)
		}

		nsSorted = nextNs
		internal := createInternal(ss, sumOfSlice)
		internal.Label = parLabel
		res = append(res, internal)
	}

	return res
}

func labelToString(label string, h int, f int) int {
	index := 0

	for i, d := range label {
		d := int(d) - 48
		j := h - i - 1

		index += int(math.Pow(float64(f), float64(j))) * d
	}

	return index
}

func divideByLabel(ns []*Node) map[int][]*Node {
	_ns := ns

	less := func(i, j int) bool {
		return len(_ns[i].Label) < len(_ns[j].Label)
	}

	sort.Slice(_ns, less)

	res := map[int][]*Node{}

	l := len(_ns[0].Label)
	i := 0

	for j, n := range _ns {
		if len(n.Label) == l {
			continue
		}

		res[l] = _ns[i:j]
		l = len(n.Label)
		i = j
	}

	res[l] = _ns[i:]

	return res
}

func intersectsArea(x, y [4]float64) bool {
	return x[0] <= y[2] && x[2] >= y[0] && x[3] <= y[1] && x[1] >= y[3] // Proof by contradiction, any of these cases mean that x and y cannot intersect; so if none exist, then they intersect: https://stackoverflow.com/a/306332
}

func containsArea(outer, inner [4]float64) bool {
	return outer[0] <= inner[0] && outer[1] >= inner[1] && outer[2] >= inner[2] && outer[3] <= inner[3]
}

func intersectsHalfSpace(l *line, r [4]float64, incl bool) bool {
	amount := intersectsHalfSpaceAux(r, l, incl)

	return amount > 0

}

func intersectsHalfSpaceAux(r [4]float64, l *line, incl bool) int {
	// TODO correct int to float64 and remove conversion
	ps := [][2]float64{
		{float64(r[0]), float64(r[1])},
		{float64(r[0]), float64(r[3])},
		{float64(r[2]), float64(r[1])},
		{float64(r[2]), float64(r[3])},
	}

	f := filter(l, ps, incl)

	return len(f)
}

func containsHalfSpace(l *line, r [4]float64, incl bool) bool {
	amount := intersectsHalfSpaceAux(r, l, incl)

	return amount == 4
}

//Print ???
func (t *Rtree) String() string {
	fmt.Println("Printing Tree...")
	return iterate(t.Root, 0, "")
}

//for printing of tree
const (
	spaces = "     "
	branch = "├──"
)

//iterate ???
func iterate(n *Node, lvl int, ID string) string {

	Ps := n.Ps
	Ks := [][4]float64{}

	for _, p := range Ps {
		Ks = append(Ks, p.MBR)
	}

	addBranch := true

	if n.Leaf {
		for _, k := range Ks {
			ID = format(lvl, ID, addBranch)
			ID += fmt.Sprintln(k[:2])
			addBranch = false
		}
		return ID
	}

	for _, k := range Ks {
		ID = format(lvl, ID, addBranch)
		ID += fmt.Sprint("(")
		ID += fmt.Sprint(k[:2])
		ID += fmt.Sprint(",")
		ID += fmt.Sprint(k[2:])
		ID += fmt.Sprintln(")")
		addBranch = false
	}

	for _, c := range Ps {
		ID = iterate(c, lvl+1, ID)
	}

	return ID
}

func format(lvl int, ID string, addBranch bool) string {
	for i := 0; i < lvl; i++ {
		ID += fmt.Sprint(spaces)
	}

	if addBranch {
		ID += fmt.Sprint(branch)
	} else {
		ID += fmt.Sprint("   ")
	}

	return ID
}
