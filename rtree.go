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
}

// Node ???
type Node struct {
	Leaf    bool
	Ks      [][4]float64
	Ps      []*Node
	Hash    []byte
	Value   int
	Label   string
	Agg     func(aggs ...int) int
	AggLeaf func(val int) int
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

	for i := 0; i < len(elems); i += fanout {
		n := createLeaf(elems, i, min(fanout, len(elems)-i), roots, aggLeaf, agg)
		roots = append(roots, n)
	}

	for len(roots) != 1 {
		temp := []*Node{}

		for i := 0; i < len(roots); i += fanout {
			n := createInternal(roots, i, min(fanout, len(roots)-i), agg)
			temp = append(temp, n)
		}

		roots = temp

	}

	roots[0].labelMaker()

	t := new(Rtree)
	t.Fanout = fanout
	t.Root = roots[0]

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

func (r *Rtree) AuthCountPoints(ps [][2]float64) ([][]*Node, []map[string][]byte) {
	pMcss := [][]*Node{}
	pSibs := []map[string][]byte{}

	for _, p := range ps {
		mcs, sib := r.AuthCountPoint(p)

		pMcss = append(pMcss, mcs)
		pSibs = append(pSibs, sib)
	}

	return pMcss, pSibs
}

// Add labels recursively
func (n *Node) labelMaker() {
	for i, c := range n.Ps {
		label := n.Label + strconv.Itoa(i)
		c.Label = label

		c.labelMaker()
	}
}

func createInternal(roots []*Node, i int, amount int, agg func(aggs ...int) int) *Node {
	n := new(Node)
	n.Ks = [][4]float64{}
	n.Ps = []*Node{}
	n.Agg = agg

	for j := 0; j < amount; j++ {
		p := roots[i+j].Ks[0]

		for _, k := range roots[i+j].Ks {
			p[0] = math.Min(p[0], k[0])
			p[1] = math.Max(p[1], k[1])
			p[2] = math.Max(p[2], k[2])
			p[3] = math.Min(p[3], k[3])
		}

		n.Ks = append(n.Ks, p)
		n.Ps = append(n.Ps, roots[i+j])
	}

	n.CalcAgg()
	n.CalcHash()

	return n
}

func (n *Node) CalcAgg() {
	childVals := []int{}

	for i := range n.Ps {
		childVals = append(childVals, n.Ps[i].Value)
	}

	n.Value = n.Agg(childVals...)
}

func (n *Node) CalcHash() {
	childVals := []int{}
	childHashes := [][]byte{}

	for i := range n.Ps {
		childVals = append(childVals, n.Ps[i].Value)
		childHashes = append(childHashes, n.Ps[i].Hash)
	}

	hashVal := []byte{}

	for i := range childHashes {
		hashVal = append(hashVal, childHashes[i]...)
		hashVal = append(hashVal, []byte(strconv.Itoa(childVals[i]))...)
	}

	hash := sha256.Sum256(hashVal)
	n.Hash = hash[:]

	n.Value = n.Agg(childVals...)
}

func createLeaf(elems [][2]float64, i int, amount int, roots []*Node, aggLeaf func(val int) int, agg func(vals ...int) int) *Node {
	n := new(Node)
	n.Ks = [][4]float64{}
	n.Ps = []*Node{}
	n.Agg = agg

	for j := 0; j < amount; j++ {
		p := [4]float64{}
		p[0] = elems[i+j][0]
		p[1] = elems[i+j][1]
		p[2] = elems[i+j][0]
		p[3] = elems[i+j][1]
		n.Ks = append(n.Ks, p)

		c := new(Node)
		n.Ps = append(n.Ps, c)
		c.Leaf = true
		c.AggLeaf = aggLeaf
		c.Value = aggLeaf(-69) // TODO Allow for other aggregate values than COUNT

		hashVal := []byte{}

		for i := range p {
			var buf []byte
			binary.BigEndian.PutUint64(buf, math.Float64bits(p[i]))

			hashVal = append(hashVal, buf...)
		}

		hashVal = append(hashVal, []byte(strconv.Itoa(c.Value))...)

		hash := sha256.Sum256(hashVal)
		c.Hash = hash[:]
	}

	n.CalcAgg()
	n.CalcHash()

	return n
}

// Search ???
func (t *Rtree) Search(area [4]float64) []*Node {
	return t.Root.searchAux(area)
}

func (n *Node) searchAux(area [4]float64) []*Node {
	acc := []*Node{}

	for i, k := range n.Ks {
		if !intersectsArea(area, k) {
			continue
		}

		if n.Leaf {
			return []*Node{n}
		}

		acc = append(acc, n.Ps[i].searchAux(area)...)
	}

	return acc
}

func (t *Rtree) AuthCountPoint(p [2]float64) ([]*Node, map[string][]byte) {
	return t.AuthCountArea([4]float64{p[0], p[1], p[0], p[1]})
}

// AuthCountArea ???
func (t *Rtree) AuthCountArea(area [4]float64) ([]*Node, map[string][]byte) {
	return t.Root.authCountAreaAux(area)
}

func (n *Node) authCountAreaAux(area [4]float64) ([]*Node, map[string][]byte) {
	mcs := []*Node{}
	sib := map[string][]byte{}

	for i, k := range n.Ks {
		if !intersectsArea(area, k) {
			sib[n.Ps[i].Label] = n.Ps[i].Hash
			continue
		}

		if containsArea(area, k) {
			mcs = append(mcs, n.Ps[i])
			continue
		}

		cMcs, cSib := n.Ps[i].authCountAreaAux(area)

		mcs = append(mcs, cMcs...)

		for k, v := range cSib {
			sib[k] = v
		}
	}

	return mcs, sib
}

// AuthCountHalfSpace ???
func (t *Rtree) AuthCountHalfSpace(l *line, sign bool) ([]*Node, map[string][]byte) {
	return t.Root.authCountHalfSpaceAux(l, sign)
}

func (n *Node) authCountHalfSpaceAux(l *line, sign bool) ([]*Node, map[string][]byte) {
	mcs := []*Node{}
	sib := map[string][]byte{}

	for i, k := range n.Ks {
		if !intersectsHalfSpace(l, k, sign) {
			sib[n.Ps[i].Label] = n.Ps[i].Hash
			continue
		}

		if containsHalfSpace(l, k, sign) {
			mcs = append(mcs, n.Ps[i])
			continue
		}

		cMcs, cSib := n.Ps[i].authCountHalfSpaceAux(l, sign)

		mcs = append(mcs, cMcs...)

		for k, v := range cSib {
			sib[k] = v
		}
	}

	return mcs, sib
}

// AuthCountVerify ???
func AuthCountVerify(mcs []*Node, sib map[string][]byte, digest []byte) (int, bool) {
	panic("todo")
}

func intersectsArea(x, y [4]float64) bool {
	return x[0] < y[2] && x[2] > y[0] && x[3] < y[1] && x[1] > y[3] // Proof by contradiction, any of these cases mean that x and y cannot intersect; so if none exist, then they intersect: https://stackoverflow.com/a/306332
}

func containsArea(outer, inner [4]float64) bool {
	return outer[0] <= inner[0] && outer[1] >= inner[1] && outer[2] >= inner[2] && outer[3] <= inner[3]
}

func intersectsHalfSpace(l *line, r [4]float64, sign bool) bool {
	amount := intersectsHalfSpaceAux(r, l, sign)

	return amount > 0

}

func intersectsHalfSpaceAux(r [4]float64, l *line, sign bool) int {
	// TODO correct int to float64 and remove conversion
	ps := [][2]float64{
		{float64(r[0]), float64(r[1])},
		{float64(r[0]), float64(r[3])},
		{float64(r[2]), float64(r[1])},
		{float64(r[2]), float64(r[3])},
	}

	f := filter(l, ps, sign)

	return len(f)
}

func containsHalfSpace(l *line, r [4]float64, sign bool) bool {
	amount := intersectsHalfSpaceAux(r, l, sign)

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
	Ks := n.Ks

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
