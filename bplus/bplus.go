package bplus

import (
	"fmt"
	"sort"

	"github.com/golang-collections/collections/tree/master/stack"
)

// Btree ???
type Bptree struct {
	Root   *Node
	Fanout int
	Depth  int //for printing
}

// Node ???
type Node struct {
	Key        int
	Hash       []byte
	Value      int
	Childs     map[int]*Node
	Sib        *Node
	Entries    map[int]int
	Discovered bool //for printing
	Depth      int  //for printing
}

// ??? elems should be entries
// NewTree ???
func NewTree(elems []int, fanout int) (*Bptree, error) {
	sort.Ints(elems)

	roots := make([]*Node, 0)

	for i := 0; i < len(elems); i += fanout - 1 {
		n := new(Node)
		n.Key = elems[i]
		n.Entries = make(map[int]int)

		for j := 0; j < fanout-1; j++ {
			n.Entries[elems[i+j]] = elems[i+j]
		}

		if len(roots) != 0 {
			roots[i/fanout].Sib = n
		}

		roots = append(roots, n)
	}

	for len(roots) != 1 {
		temp := make([]*Node, 0)

		for i := 0; i < len(roots); i += fanout {
			n := new(Node)
			n.Key = roots[i+1].Key
			n.Childs = make(map[int]*Node)

			for j := 0; j < fanout; j++ {
				if i+j >= len(roots) {
					break
				}

				n.Childs[roots[i+j].Key] = roots[i+j]
			}

			temp = append(temp, n)

		}

		roots = temp

	}

	t := new(Bptree)
	t.Fanout = fanout
	t.Root = roots[0]

	return t, nil
}

//Print ???
func (t *Bptree) Print() {
	fmt.Println("Printing Tree...")
	Iterate(t.Root, 0)
}

//Iterate ???
func Iterate(n *Node, depth int) {
	S := stack.New()
	S.Push(n)
	for S.Len() > 0 {
		v := S.Pop()

		for i := 0; i < v.Depth-depth; i++ {
			print("\t") //tabs?
		}

		allKeys := ""

		//TODO

		print("└──" + allKeys)

		if !v.Discovered {
			v.Discovered = true
			for c := range v.Childs {
				S.push(c)
			}
		}
	}
}
