package topics

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Model struct {
	topics *Tree
	cursor *Node
}

func New(hostname string) Model {
	s := defaultStyles()
	topicTree := NewTree(hostname, s)
	return Model{
		topics: topicTree,
		cursor: topicTree.Root,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) moveUp() {
	if m.cursor == m.topics.Root {
		return
	}
	parent := m.findParent(m.topics.Root, m.cursor)
	if parent != nil {
		for i, child := range parent.Children {
			if child == m.cursor && i > 0 {
				m.cursor = parent.Children[i-1]
				return
			}
		}
	}
}

func (m *Model) moveDown() {
	if len(m.cursor.Children) > 0 && m.cursor.Open {
		m.cursor = m.cursor.Children[0]
		return
	}
	parent := m.findParent(m.topics.Root, m.cursor)
	if parent != nil {
		for i, child := range parent.Children {
			if child == m.cursor && i < len(parent.Children)-1 {
				m.cursor = parent.Children[i+1]
				return
			}
		}
	}
}

func (m *Model) findParent(root, node *Node) *Node {
	if root == nil {
		return nil
	}
	for _, child := range root.Children {
		if child == node {
			return root
		}
		if parent := m.findParent(child, node); parent != nil {
			return parent
		}
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case paho.Message:
		m.topics.Add(msg.Topic())
	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			slog.Debug("h key pressed")
			m.cursor.Open = false
		case "l":
			slog.Debug("l key pressed")
			m.cursor.Open = true
		case "j":
			slog.Debug("j key pressed")
			m.moveDown()
		case "k":
			slog.Debug("k key pressed")
			m.moveUp()
		}
	}
	return m, nil
}

func (m Model) View() string {
	return m.topics.Render().String()
}
