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

func (n *Node) Clone() *Node {
	newN := *n
	copy(newN.Ps, n.Ps)
	copy(newN.Hash, n.Hash)

	return &newN
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

	for len(roots) != 1 {
		temp := []*Node{}

		for i := 0; i < len(roots); i += fanout {
			j := min(i+fanout, len(roots)) // TODO maybe off by 1

			n := createInternal(roots[i:j], agg)
			temp = append(temp, n)
		}

		roots = temp
	}

	roots[0].labelMaker()

	t := new(Rtree)
	t.Fanout = fanout
	t.Root = roots[0]
	t.Digest = t.Root.Hash

	return t, nil
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

		if done {
			n.CalcAgg()
			n.CalcMBR()
			n.CalcHash()
			return true
		}
	}

	return false
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

		if done {
			n.CalcMBR()
			n.CalcHash()
			return true
		}
	}

	return false
}

func createInternal(ns []*Node, agg func(aggs ...int) int) *Node {
	internal := new(Node)
	internal.Ps = []*Node{}
	internal.Agg = agg
	internal.MBR = ns[0].MBR

	internal.Ps = append(internal.Ps, ns...)

	internal.CalcMBR()
	internal.CalcAgg()
	internal.CalcHash()

	return internal
}

func (n *Node) CalcMBR() {
	for _, p := range n.Ps {
		mbr := n.MBR

		mbr[0] = roundFloat(math.Min(mbr[0], p.MBR[0]), eps)
		mbr[1] = roundFloat(math.Max(mbr[1], p.MBR[1]), eps)
		mbr[2] = roundFloat(math.Max(mbr[2], p.MBR[2]), eps)
		mbr[3] = roundFloat(math.Min(mbr[3], p.MBR[3]), eps)

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
	return t.Root.authCountAreaAux(area)
}

func (n *Node) authCountAreaAux(area [4]float64) *VOCount {
	vo := new(VOCount)
	vo.Mcs = []*Node{}
	vo.Sib = []*Node{}

	for i, c := range n.Ps {
		_ = i

		if !intersectsArea(area, c.MBR) {
			cc := c.Clone()
			vo.Sib = append(vo.Sib, cc)
			continue
		}

		if containsArea(area, c.MBR) {
			cc := c.Clone()
			vo.Mcs = append(vo.Mcs, cc)
			continue
		}

		voChild := c.authCountAreaAux(area)

		vo.Mcs = append(vo.Mcs, voChild.Mcs...)

		if len(voChild.Mcs) != 0 {
			vo.Sib = append(vo.Sib, voChild.Sib...)
			continue
		}

		vo.Sib = append(vo.Sib, c)

	}

	return vo
}

// AuthCountHalfSpace ???
func (t *Rtree) AuthCountHalfSpace(l *line) *VOCount {
	return t.Root.authCountHalfSpaceAux(l)
}

func (n *Node) authCountHalfSpaceAux(l *line) *VOCount {
	vo := new(VOCount)
	vo.Mcs = []*Node{}
	vo.Sib = []*Node{}

	for i, c := range n.Ps {
		_ = i

		if !intersectsHalfSpace(l, c.MBR) {
			cc := c.Clone()
			vo.Sib = append(vo.Sib, cc)
			continue
		}

		if containsHalfSpace(l, c.MBR) {
			cc := c.Clone()
			vo.Mcs = append(vo.Mcs, cc)
			continue
		}

		voChild := c.authCountHalfSpaceAux(l)

		vo.Mcs = append(vo.Mcs, voChild.Mcs...)

		if len(voChild.Mcs) != 0 {
			vo.Sib = append(vo.Sib, voChild.Sib...)
			continue
		}

		vo.Sib = append(vo.Sib, c)
	}

	return vo
}

// AuthCountVerify ???
func AuthCountVerify(vo *VOCount, digest []byte, f int) (int, bool) {
	ns := append(vo.Mcs, vo.Sib...)
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

	calc := map[string]*Node{}

	last := len(ks) - 1
	for i := ks[last]; i > 0; i-- {
		l := map[string]*Node{}

		for _, n := range ls[i] {
			l[n.Label] = n
		}

		for lab, n := range calc {
			l[lab] = n
		}

		calc = calcNext(l, f)
	}

	res := []*Node{}

	for _, n := range calc {
		res = append(res, n)
	}

	return res
}

func calcNext(ns map[string]*Node, f int) map[string]*Node {
	res := map[string]*Node{}

	for len(ns) != 0 {
		var n *Node

		for _, node := range ns {
			n = node
			break
		}

		parLabel := n.Label[:len(n.Label)-1]

		ss := map[string]*Node{}
		for i := 0; i < f; i++ {
			iStr := strconv.Itoa(i)

			sLabel := parLabel + iStr
			sNode := ns[sLabel]

			if sNode == nil {
				continue
			}

			delete(ns, sLabel)
			ss[sLabel] = sNode
		}

		internalNs := []*Node{}

		for _, n := range ss {
			internalNs = append(internalNs, n)
		}

		internal := createInternal(internalNs, sumOfSlice)
		internal.Label = parLabel
		res[internal.Label] = internal
	}

	return res
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
	x[0] -= eps
	x[1] += eps
	x[2] += eps
	x[3] -= eps

	y[0] -= eps
	y[1] += eps
	y[2] += eps
	y[3] -= eps

	return x[0] < y[2] && x[2] > y[0] && x[3] < y[1] && x[1] > y[3] // Proof by contradiction, any of these cases mean that x and y cannot intersect; so if none exist, then they intersect: https://stackoverflow.com/a/306332
}

func containsArea(outer, inner [4]float64) bool {
	return outer[0] <= inner[0] && outer[1] >= inner[1] && outer[2] >= inner[2] && outer[3] <= inner[3]
}

func intersectsHalfSpace(l *line, r [4]float64) bool {
	amount := intersectsHalfSpaceAux(r, l)

	return amount > 0

}

func intersectsHalfSpaceAux(r [4]float64, l *line) int {
	// TODO correct int to float64 and remove conversion
	ps := [][2]float64{
		{float64(r[0]), float64(r[1])},
		{float64(r[0]), float64(r[3])},
		{float64(r[2]), float64(r[1])},
		{float64(r[2]), float64(r[3])},
	}

	f := filter(l, ps)

	return len(f)
}

func containsHalfSpace(l *line, r [4]float64) bool {
	amount := intersectsHalfSpaceAux(r, l)

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
