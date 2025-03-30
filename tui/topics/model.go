package topics

import (
	"github.com/brancobruyneel/lens/tui/config"
	tea "github.com/charmbracelet/bubbletea"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Model struct {
	topics        *Tree
	cursor        *Node
	selected      bool
	filteredTopic string
}

func New(broker string) Model {
	s := defaultStyles()
	topicTree := NewTree(broker, s)
	return Model{
		topics:        topicTree,
		cursor:        topicTree.Root,
		selected:      false,
		filteredTopic: "#",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) moveCursor(node *Node) {
	if m.cursor != nil {
		m.cursor.Selected = false
	}
	m.cursor = node
	m.cursor.Selected = true
}

func (m *Model) moveUp() {
	if m.cursor == m.topics.Root {
		return
	}

	parent := m.cursor.Parent
	if parent == nil {
		return
	}

	// Find current node's index among its siblings
	cursorIdx := m.findNodeIndex(m.cursor, parent.Children)
	if cursorIdx == -1 {
		return
	}

	// If this is the first child, move to parent
	if cursorIdx == 0 {
		m.moveCursor(parent)
		return
	}

	// Move to previous sibling
	prevSibling := parent.Children[cursorIdx-1]

	// If previous sibling has children and is open, navigate to its deepest last open child
	if len(prevSibling.Children) > 0 && prevSibling.Open {
		m.moveCursor(m.topics.FindDeepestLastChild(prevSibling))
	} else {
		m.moveCursor(prevSibling)
	}
}

// Helper function to find a node's index in a slice
func (m *Model) findNodeIndex(node *Node, siblings []*Node) int {
	for i, sibling := range siblings {
		if sibling == node {
			return i
		}
	}
	return -1
}

func (m *Model) moveDown() {
	// If current node has children and is open, move to first child
	if len(m.cursor.Children) > 0 && m.cursor.Open {
		m.moveCursor(m.cursor.Children[0])
		return
	}

	parent := m.cursor.Parent
	if parent == nil {
		return
	}

	// Find current node's index among siblings
	cursorIdx := m.findNodeIndex(m.cursor, parent.Children)
	if cursorIdx == -1 {
		return
	}

	// If there's a next sibling, move to it
	if cursorIdx < len(parent.Children)-1 {
		m.moveCursor(parent.Children[cursorIdx+1])
		return
	}

	// Try to find next sibling of an ancestor
	oldCursor := m.cursor
	m.cursor.Selected = false

	if m.moveToNextAncestorSibling(parent) {
		m.cursor.Selected = true
	} else {
		m.moveCursor(oldCursor)
	}
}

func (m *Model) moveToNextAncestorSibling(startNode *Node) bool {
	current := startNode
	ancestor := current.Parent

	for ancestor != nil {
		currentIdx := m.findNodeIndex(current, ancestor.Children)

		if currentIdx != -1 && currentIdx < len(ancestor.Children)-1 {
			m.cursor = ancestor.Children[currentIdx+1]
			return true
		}

		current = ancestor
		ancestor = current.Parent
	}

	return false
}

func (m *Model) topicFilterCmd() tea.Cmd {
	topicPath := m.topics.Path(m.cursor)

	// Toggle: if current topic is already filtered, reset to all topics
	if m.filteredTopic == topicPath {
		topicPath = "#"
	}

	m.filteredTopic = topicPath
	m.topics.FilterTopic(m.filteredTopic)

	return func() tea.Msg {
		return config.TopicFilterMsg(topicPath)
	}
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "backspace":
		parent := m.cursor.Parent
		if parent != nil && len(m.cursor.Children) == 0 {
			parent.Open = false
			m.moveCursor(parent)
			return m, nil
		}
		m.cursor.Open = false
	case "enter":
		m.cursor.Open = true
	case "j":
		m.moveDown()
	case "k":
		m.moveUp()
	}

	if msg.Type == tea.KeySpace {
		return m, m.topicFilterCmd()
	}

	return m, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case paho.Message:
		m.topics.Add(msg.Topic())
	case config.TopicFilterMsg:
		return m, nil
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) View() string {
	return m.topics.Render().String()
}
