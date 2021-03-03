package bplus

import (
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
	Value   int
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

// Print ???
// func (t *Bptree) Print() {
// 	fmt.Println("Printing Tree...")
// 	Iterate(t.Root, 0)
// }

// //Iterate ???
// // TODO: need find a way to sort kids by key, since it appears in randomized order
// func Iterate(n *Node, lvl int) {
// 	var keys []int

// 	for i := 0; i < lvl; i++ {
// 		fmt.Print("     ")
// 	}

// 	fmt.Print("├──")

// 	kids := n.Childs //needs to be sorted by key somehow

// 	if kids == nil {
// 		for _, e := range n.Entries {
// 			fmt.Println(":", e)
// 		}
// 		return
// 	}

// 	for i := range kids {
// 		keys = append(keys, i)
// 	}

// 	sort.Ints(keys) //maybe unecessary

// 	fmt.Println(keys[1:])

// 	for c := range kids {
// 		Iterate(kids[c], lvl+1)
// 	}
// }
