package bplus

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
)

// Bptree ???
type Bptree struct {
	Root   *Node
	Fanout int
}

// Node ???
//terminology of the aggrigate paper
// ??? Hashes, Values and entries has size fanout, and each of their entry corresponds to the 'leaf' according to the
type Node struct {
	Labels  []string
	Leaf    bool
	Ks      []int
	Ps      []*Node
	Hashes  [][]byte
	Values  []int
	Sib     *Node
	Entries []int
	Agg     func(aggs ...int) int
	AggLeaf func(val int) int
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
func NewTree(elems []int, fanout int, agg func(aggs ...int) int, aggLeaf func(val int) int) (*Bptree, error) {
	sort.Ints(elems)

	roots := []*Node{}

	lastChunkLen := len(elems) - (fanout - 1)

	for i := 0; i < lastChunkLen; i += fanout / 2 {
		n := createLeaf(elems, i, fanout/2, roots, aggLeaf)
		roots = append(roots, n)
	}

	n := createLeaf(elems, max(lastChunkLen, 0), fanout-1, roots, aggLeaf)
	roots = append(roots, n)

	for len(roots) != 1 {
		temp := []*Node{}

		lastChunkLen := 0

		if len(roots) > fanout {
			lastChunkLen = len(roots) - fanout
			lastChunkLen = lastChunkLen + lastChunkLen%((fanout+1)/2)
		}

		for i := 0; i < lastChunkLen; i += (fanout + 1) / 2 {
			n := createInternal(roots, i, (fanout+1)/2, agg)
			temp = append(temp, n)
		}

		n := createInternal(roots, lastChunkLen, len(roots)-lastChunkLen, agg)
		temp = append(temp, n)

		roots = temp
	}

	t := new(Bptree)
	t.Fanout = fanout
	t.Root = roots[0]

	return t, nil
}

func createInternal(roots []*Node, i int, amount int, agg func(aggs ...int) int) *Node {
	n := new(Node)
	n.Ks = []int{}
	n.Ps = []*Node{}
	n.Agg = agg
	n.Values = []int{}
	n.Hashes = [][]byte{}

	for j := 0; j < amount-1; j++ {
		n.Ps = append(n.Ps, roots[i+j])
		n.Ks = append(n.Ks, roots[i+j+1].Ks[0])
	}

	n.Ps = append(n.Ps, roots[i+amount-1])

	for k := range n.Ps {

		n.Values = append(n.Values, 0)
		n.Values[k] = n.Agg(n.Ps[k].Values...)

		hashVal := []byte{}

		for j := range n.Ps[k].Hashes {
			hashVal = append(hashVal, n.Ps[k].Hashes[j]...)
			hashVal = append(hashVal, []byte(strconv.Itoa(n.Ps[k].Values[j]))...)
		}

		hash := sha256.Sum256(hashVal)

		n.Hashes = append(n.Hashes, nil)
		n.Hashes[k] = hash[:] //causes panic
	}

	return n
}

func createLeaf(elems []int, i int, amount int, roots []*Node, aggLeaf func(val int) int) *Node {
	n := new(Node)
	n.Entries = []int{}
	n.Ks = []int{}
	n.Leaf = true
	n.Values = []int{}
	n.Hashes = [][]byte{}
	n.AggLeaf = aggLeaf

	for j := 0; j < amount; j++ {
		n.Ks = append(n.Ks, elems[i+j])
		n.Entries = append(n.Entries, elems[i+j])
	}

	for i := range n.Ks {
		n.Values = append(n.Values, n.AggLeaf(n.Entries[i]))

		hashVal := []byte(strconv.Itoa(n.Ks[i]) + strconv.Itoa(n.Values[i]))

		hash := sha256.Sum256(hashVal)

		n.Hashes = append(n.Hashes, hash[:])
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

//not working
func insertLabels(t *Bptree) *Bptree {
	labels := []string{}
	root := t.Root
	root.Labels = labels
	fanout := t.Fanout

	for i := 1; i < fanout; i++ {
		labels = append(labels, fmt.Sprint(i))
	}

	root.Labels = append(root.Labels, labels...)

	for i, n := range root.Ps {
		addLabel(n, labels, i+1)
	}

	t.Root = root
	return t
}

//not working:
func addLabel(n *Node, labels []string, i int) {
	nlbs := []string{} //newlabels
	nlbs = append(nlbs, labels...)

	for j := range nlbs {
		nlbs[j] = Reverse(nlbs[j])
		nlbs[j] += fmt.Sprint(i)
		nlbs[j] = Reverse(nlbs[j])
	}

	n.Labels = append(n.Labels, nlbs...)

	if n.Leaf {
		return
	}

	for i, c := range n.Ps {
		addLabel(c, nlbs, i+1)
	}

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
		//ID = ID + fmt.Sprintln(":", n.Ks)
		ID = ID + fmt.Sprintln(":", n.Labels)
		return ID
	}

	Ps := n.Ps
	//Ks := n.Ks
	Ks := n.Labels

	ID = ID + fmt.Sprintln(Ks)

	for _, c := range Ps {
		ID = Iterate(c, lvl+1, ID)
	}
	return ID
}

//Reverse is used for labelling tree
func Reverse(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}
	return
}
