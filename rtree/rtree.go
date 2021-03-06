package rtree

import (
	"fmt"
	"sort"
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
	Value   int //aggregate value
	Entries []int
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
func NewTree(elems [][2]int, fanout int) (*Rtree, error) {
	sort.Slice(elems, func(i, j int) bool {
		return elems[i][0] < elems[j][0]
	})

	roots := make([]*Node, 0)

	for i := 0; i < len(elems); i += fanout {
		n := createLeaf(elems, i, min(fanout, len(elems)-i), roots)
		roots = append(roots, n)
	}

	for len(roots) != 1 {
		temp := make([]*Node, 0)

		for i := 0; i < len(roots); i += fanout {
			n := createInternal(roots, i, min(fanout, len(roots)-i))
			temp = append(temp, n)
		}

		roots = temp

	}

	t := new(Rtree)
	t.Fanout = fanout
	t.Root = roots[0]

	return t, nil
}

func createInternal(roots []*Node, i int, amount int) *Node {
	n := new(Node)
	n.Ks = make([][4]int, 0)
	n.Ps = make([]*Node, 0)

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

	return n
}

func createLeaf(elems [][2]int, i int, amount int, roots []*Node) *Node {
	n := new(Node)
	n.Entries = make([]int, 0)
	n.Ks = make([][4]int, 0)
	n.Leaf = true

	for j := 0; j < amount; j++ {
		p := [4]int{}
		p[0] = elems[i+j][0]
		p[1] = elems[i+j][1]
		p[2] = elems[i+j][0]
		p[3] = elems[i+j][1]

		n.Ks = append(n.Ks, p)
		n.Entries = append(n.Entries, elems[i+j][0])
	}

	return n
}

//Print ???
func (t *Rtree) String() string {
	fmt.Println("Printing Tree...")
	return Iterate(t.Root, 0, "")
}

//for printing of tree
const (
	spaces = "     "
	branch = "├──"
)

//Iterate ???
func Iterate(n *Node, lvl int, ID string) string {

	for i := 0; i < lvl; i++ {
		ID += fmt.Sprint(spaces)
	}

	ID += fmt.Sprint(branch)

	if n.Leaf {
		ID += fmt.Sprintln(":", n.Entries)
		return ID
	}

	Ps := n.Ps
	Ks := n.Ks

	for k := range Ks {
		ID += fmt.Sprintln(k)
	}

	for _, c := range Ps {
		ID = Iterate(c, lvl+1, ID)
	}

	return ID
}
