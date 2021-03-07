package bplus

import (
	"fmt"
	"sort"
)

// Bptree ???
type Bptree struct {
	Root   *Node
	Fanout int
}

// Node ???
type Node struct {
	Leaf    bool
	Ks      []int
	Ps      []*Node
	Hash    []byte
	Value   int //aggregate value
	Sib     *Node
	Entries []int
}

func max(x int, y int) int {
	if x > y {
		return x
	}

	return y
}

// NewTree ???
// ??? Needs a certain amount of elements to work - around fanout
// ??? elems should be entries
//TODO : compute Hash and Aggregate value of each node
func NewTree(elems []int, fanout int) (*Bptree, error) {
	sort.Ints(elems)

	roots := make([]*Node, 0)

	lastChunkLen := len(elems) - (fanout - 1)

	for i := 0; i < lastChunkLen; i += fanout / 2 {
		n := createLeaf(elems, i, fanout/2, roots)
		roots = append(roots, n)
	}

	n := createLeaf(elems, max(lastChunkLen, 0), fanout-1, roots)
	roots = append(roots, n)

	for len(roots) != 1 {
		temp := make([]*Node, 0)

		lastChunkLen := 0

		if len(roots) > fanout {
			lastChunkLen = len(roots) - fanout
			lastChunkLen = lastChunkLen + lastChunkLen%((fanout+1)/2)
		}

		for i := 0; i < lastChunkLen; i += (fanout + 1) / 2 {
			n := createInternal(roots, i, (fanout+1)/2)
			temp = append(temp, n)
		}

		n := createInternal(roots, lastChunkLen, len(roots)-lastChunkLen)
		temp = append(temp, n)

		roots = temp

	}

	t := new(Bptree)
	t.Fanout = fanout
	t.Root = roots[0]

	return t, nil
}

func createInternal(roots []*Node, i int, amount int) *Node {
	n := new(Node)
	n.Ks = make([]int, 0)
	n.Ps = make([]*Node, 0)

	for j := 0; j < amount-1; j++ {
		n.Ps = append(n.Ps, roots[i+j])
		n.Ks = append(n.Ks, roots[i+j+1].Ks[0])
	}

	n.Ps = append(n.Ps, roots[i+amount-1])

	return n
}

func createLeaf(elems []int, i int, amount int, roots []*Node) *Node {
	n := new(Node)
	n.Entries = make([]int, 0)
	n.Ks = make([]int, 0)
	n.Leaf = true

	for j := 0; j < amount; j++ {
		n.Ks = append(n.Ks, elems[i+j])
		n.Entries = append(n.Entries, elems[i+j])
	}

	if len(roots) != 0 {
		roots[len(roots)-1].Sib = n
	}
	return n
}

//Print ???
func (t *Bptree) String() string {
	fmt.Println("Printing Tree...")
	return Iterate(t.Root, 0, "")
}

// Search ???
func (t *Bptree) Search(k int) *Node {
	return t.Root.searchAux(k)
}

func (n *Node) searchAux(k int) *Node {
	if n.Leaf {
		for _, kv := range n.Ks {
			if kv != k {
				continue
			}
			return n
		}

		return nil
	}

	for i := range n.Ks {
		j := len(n.Ks) - 1 - i
		kv := n.Ks[j]

		if k >= kv {
			continue
		}

		return n.Ps[j].searchAux(k)
	}

	return n.Ps[len(n.Ps)-1]
}

//for printing of tree
const (
	spaces = "     "
	branch = "├──"
)

//Iterate ???
func Iterate(n *Node, lvl int, ID string) string {

	for i := 0; i < lvl; i++ {
		ID = ID + fmt.Sprint(spaces)
	}

	ID = ID + fmt.Sprint(branch)

	if n.Leaf {
		ID = ID + fmt.Sprintln(":", n.Entries)
		return ID
	}

	Ps := n.Ps
	Ks := n.Ks

	ID = ID + fmt.Sprintln(Ks)

	for _, c := range Ps {
		ID = Iterate(c, lvl+1, ID)
	}
	return ID
}
