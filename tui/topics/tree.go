package tree

type Node interface {
	Value() string
	Children() Children
	Hidden() bool
}

type Tree struct {
}

type Node struct {
}

// Leaf is a node without children.
type Leaf struct {
	value  string
	hidden bool
}

func 
