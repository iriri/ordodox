package lru

// i think this is everything that can be shared as of go 1, unfortunately,
// since we can't do unboxed upcasts or c-style structual subtyping
//
//update: i actually ended up doing some c-style structural substyping anyways
//sigh. go 2 can't come soon enough
type Node struct {
	Prev *Node
	Next *Node
	Key  string
}

type List = Node

func (l *List) Init() {
	l.Prev = l
	l.Next = l
}

func (n *Node) Remove() {
	n.Prev.Next = n.Next
	n.Next.Prev = n.Prev
}

func (l *List) Push(n *Node) {
	n.Next = l
	n.Prev = l.Prev
	l.Prev.Next = n
	l.Prev = n
}

func (l *List) Shift() string {
	n := l.Next
	n.Remove()
	n.Next = nil // gc
	n.Prev = nil
	return n.Key
}
