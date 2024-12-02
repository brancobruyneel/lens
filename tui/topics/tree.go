package topics

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
)

func defaultStyles() styles {
	var s styles
	s.base = lipgloss.NewStyle().
		Foreground(lipgloss.Color("225"))
	s.block = s.base.
		Padding(1, 3).
		Margin(1, 3).
		Width(40)
	s.enumerator = s.base.
		Foreground(lipgloss.Color("212")).
		PaddingRight(1)
	s.node = s.base.
		Inline(true)
	s.toggle = s.base.
		Foreground(lipgloss.Color("207")).
		PaddingRight(1)
	s.leaf = s.base
	s.selected = s.base.
		Foreground(lipgloss.Color("82")) // Define the selected style with a different color
	return s
}

type styles struct {
	base       lipgloss.Style
	block      lipgloss.Style
	enumerator lipgloss.Style
	node       lipgloss.Style
	toggle     lipgloss.Style
	leaf       lipgloss.Style
	selected   lipgloss.Style // Add the selected style
}

type Node struct {
	Root     bool
	Name     string
	Children []*Node
	Styles   styles
	Open     bool
	Selected bool
}

func (t *Node) String() string {
	var nodeStyle lipgloss.Style
	if t.Selected {
		nodeStyle = t.Styles.selected
	} else {
		nodeStyle = t.Styles.node
	}

	if t.Root || len(t.Children) == 0 {
		return nodeStyle.Render(t.Name)
	}

	if t.Open {
		return t.Styles.toggle.Render("▼") + nodeStyle.Render(t.Name)
	}

	return t.Styles.toggle.Render("▶") + nodeStyle.Render(t.Name)
}

func (t *Node) Toggle() {
	t.Open = !t.Open
}

type Tree struct {
	Root *Node
}

func NewTree(rootName string, styles styles) *Tree {
	return &Tree{
		Root: &Node{Name: rootName, Styles: styles, Open: true, Root: true},
	}
}

func (t *Tree) Add(topic string) {
	levels := strings.Split(topic, "/")
	current := t.Root

	for _, level := range levels {
		if level == "" {
			continue
		}
		found := false
		for _, child := range current.Children {
			if child.Name == level {
				current = child
				found = true
				break
			}
		}
		if !found {
			newNode := &Node{Name: level, Styles: current.Styles, Open: false}
			current.Children = append(current.Children, newNode)
			current.Open = true
			current = newNode
		}
	}
}

func (t *Tree) Render() *tree.Tree {
	var convert func(*Node) *tree.Tree
	convert = func(n *Node) *tree.Tree {
		tr := tree.Root(n)
		for _, child := range n.Children {
			childTree := convert(child)
			if !n.Open {
				childTree.Hide(true)
			}
			tr = tr.Child(childTree)
		}
		return tr
	}
	return convert(t.Root)
}
