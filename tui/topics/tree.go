package topics

import (
	"fmt"
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
		Bold(true).
		Foreground(lipgloss.Color("212"))
	s.filtered = s.base.
		Bold(true).
		Foreground(lipgloss.Color("86"))
	return s
}

type styles struct {
	base       lipgloss.Style
	block      lipgloss.Style
	enumerator lipgloss.Style
	node       lipgloss.Style
	toggle     lipgloss.Style
	leaf       lipgloss.Style
	selected   lipgloss.Style
	filtered   lipgloss.Style
}

type Node struct {
	Root     bool
	Name     string
	Children []*Node
	Parent   *Node
	Styles   styles
	Open     bool
	Selected bool
	Filtered bool
	MsgCount int
	childMap map[string]*Node // For quick lookups
}

func (t *Node) String() string {
	var nodeStyle lipgloss.Style

	if t.Selected {
		nodeStyle = t.Styles.selected
	} else if t.Filtered {
		nodeStyle = t.Styles.filtered
	} else {
		nodeStyle = t.Styles.node
	}

	countStyle := t.Styles.base.Foreground(lipgloss.Color("240"))

	msgCount := ""
	if t.MsgCount > 0 {
		msgCount = countStyle.Render(fmt.Sprintf(" (%d messages)", t.MsgCount))
	}

	topicCount := ""
	if len(t.Children) > 0 {
		topicCount = countStyle.Render(fmt.Sprintf(" (%d topics)", len(t.Children)))
	}

	// Root or leaf nodes don't show toggle indicators
	if t.Root || len(t.Children) == 0 {
		return nodeStyle.Render(t.Name) + msgCount
	}

	toggle := "▶"
	if t.Open {
		toggle = "▼"
	}

	return t.Styles.toggle.Render(toggle) + nodeStyle.Render(t.Name) + topicCount + msgCount
}

func (t *Node) Toggle() {
	t.Open = !t.Open
}

type Tree struct {
	Root *Node
}

func NewTree(rootName string, styles styles) *Tree {
	return &Tree{
		Root: &Node{
			Name:     rootName,
			Styles:   styles,
			Open:     true,
			Root:     true,
			MsgCount: 0,
			childMap: make(map[string]*Node),
		},
	}
}

func (t *Tree) Add(topic string) {
	levels := strings.Split(topic, "/")
	current := t.Root

	for _, level := range levels {
		if level == "" {
			continue
		}

		// Fast lookup using map
		if child, exists := current.childMap[level]; exists {
			current = child
			current.MsgCount++
		} else {
			newNode := &Node{
				Name:     level,
				Styles:   current.Styles,
				Open:     false,
				Parent:   current,
				childMap: make(map[string]*Node),
				MsgCount: 1, // Initialize with 1 since this is a new message
			}

			// Add to parent's children and map
			current.Children = append(current.Children, newNode)
			current.childMap[level] = newNode
			current.Open = true

			// Move to the new node
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

func (t *Tree) FindDeepestLastChild(node *Node) *Node {
	current := node
	for len(current.Children) > 0 && current.Open {
		current = current.Children[len(current.Children)-1]
	}
	return current
}

func (t *Tree) Path(node *Node) string {
	if node == t.Root {
		return "#"
	}

	var path []string
	current := node
	for current != t.Root && current != nil {
		path = append([]string{current.Name}, path...)
		current = current.Parent
	}

	return "/" + strings.Join(path, "/")
}

func (t *Tree) FilterTopic(topic string) {
	t.clearFilters(t.Root)

	if topic == "#" {
		return
	}

	t.markFilteredNode(topic)
}

func (t *Tree) clearFilters(node *Node) {
	node.Filtered = false
	for _, child := range node.Children {
		t.clearFilters(child)
	}
}

func (t *Tree) markFilteredNode(topic string) {
	tp := strings.TrimPrefix(topic, "/")
	levels := strings.Split(tp, "/")

	current := t.Root

	for _, level := range levels {
		if level == "" {
			continue
		}

		// Look for the child with this name
		found := false
		for _, child := range current.Children {
			if child.Name == level {
				// Only mark the final node as filtered
				if level == levels[len(levels)-1] {
					child.Filtered = true
				}
				child.Open = true // Ensure the node is open
				current = child
				found = true
				break
			}
		}

		if !found {
			return
		}
	}
}
