package main

import (
	"crypto/sha256"
	"fmt"
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
	Ks      [][4]int
	Ps      []*Node
	Hash    []byte
	Value   int
	Label   string
	Agg     func(aggs ...int) int
	AggLeaf func(val int) int
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

// NewTree ???
// ??? Needs a certain amount of elements to work - around fanout
// ??? elems should be entries
//TODO : compute Hash and Aggregate value of each node
func NewTree(elems [][2]int, fanout int, agg func(aggs ...int) int, aggLeaf func(val int) int) (*Rtree, error) {
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
	n.Ks = [][4]int{}
	n.Ps = []*Node{}
	n.Agg = agg

	for j := 0; j < amount; j++ {
		p := roots[i+j].Ks[0]

		for _, k := range roots[i+j].Ks {
			p[0] = min(p[0], k[0])
			p[1] = max(p[1], k[1])
			p[2] = max(p[2], k[2])
			p[3] = min(p[3], k[3])
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

func createLeaf(elems [][2]int, i int, amount int, roots []*Node, aggLeaf func(val int) int, agg func(vals ...int) int) *Node {
	n := new(Node)
	n.Ks = [][4]int{}
	n.Ps = []*Node{}
	n.Agg = agg

	for j := 0; j < amount; j++ {
		p := [4]int{}
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
			hashVal = append(hashVal, []byte(strconv.Itoa(p[i]))...)
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
func (t *Rtree) Search(area [4]int) []*Node {
	return t.Root.searchAux(area)
}

func (n *Node) searchAux(area [4]int) []*Node {
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

// AuthCountArea ???
func (t *Rtree) AuthCountArea(area [4]int) ([]*Node, map[string][]byte) {
	return t.Root.authCountAreaAux(area)
}

func (n *Node) authCountAreaAux(area [4]int) ([]*Node, map[string][]byte) {
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

// AuthCountLine ???
func (t *Rtree) AuthCountLine(l *line) ([]*Node, map[string][]byte) {
	return t.Root.authCountAreaAux(area)
}

func (n *Node) authCountLineAux(l *line) ([]*Node, map[string][]byte) {
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

// AuthCountVerify ???
func AuthCountVerify(mcs []*Node, sib map[string][]byte, digest []byte) {

}

func intersectsArea(x, y [4]int) bool {
	return x[0] < y[2] && x[2] > y[0] && x[3] < y[1] && x[1] > y[3] // Proof by contradiction, any of these cases mean that x and y cannot intersect; so if none exist, then they intersect: https://stackoverflow.com/a/306332
}

func containsArea(outer, inner [4]int) bool {
	return outer[0] <= inner[0] && outer[1] >= inner[1] && outer[2] >= inner[2] && outer[3] <= inner[3]
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
