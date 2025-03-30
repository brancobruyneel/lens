package topics

import (
	"strings"

	"github.com/brancobruyneel/lens/tui/config"
	tea "github.com/charmbracelet/bubbletea"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Model struct {
	topics   *Tree
	cursor   *Node
	selected bool
}

func New(broker string) Model {
	s := defaultStyles()
	topicTree := NewTree(broker, s)
	return Model{
		topics:   topicTree,
		cursor:   topicTree.Root,
		selected: false,
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
	if parent == nil {
		return
	}

	for i, child := range parent.Children {
		if child == m.cursor {
			m.cursor.Selected = false
			if i > 0 {
				m.cursor = parent.Children[i-1]
			} else {
				m.cursor = parent
			}
			m.cursor.Selected = true
			return
		}
	}
}

func (m *Model) moveDown() {
	if len(m.cursor.Children) > 0 && m.cursor.Open {
		m.cursor.Selected = false // Deselect the current cursor node
		m.cursor = m.cursor.Children[0]
		m.cursor.Selected = true // Select the new cursor node
		return
	}

	parent := m.findParent(m.topics.Root, m.cursor)
	if parent == nil {
		return
	}

	for i, child := range parent.Children {
		if child == m.cursor && i < len(parent.Children)-1 {
			m.cursor.Selected = false // Deselect the current cursor node
			m.cursor = parent.Children[i+1]
			m.cursor.Selected = true // Select the new cursor node
			return
		}
	}
}

func (m Model) findParent(root, node *Node) *Node {
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

func (m Model) TopicPath() string {
	if m.cursor == m.topics.Root {
		return "#"
	}

	var path []string
	current := m.cursor
	for current != m.topics.Root {
		path = append([]string{current.Name}, path...)
		parent := m.findParent(m.topics.Root, current)
		if parent == nil {
			break
		}
		current = parent
	}

	return "/" + strings.Join(path, "/")
}

func (m Model) topicSelectCmd() tea.Cmd {
	return func() tea.Msg {
		return config.TopicSelectMsg(m.TopicPath())
	}
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "backspace":
		parent := m.findParent(m.topics.Root, m.cursor)
		if parent != nil && len(m.cursor.Children) == 0 {
			parent.Open = false
			m.cursor.Selected = false
			m.cursor = parent
			m.cursor.Selected = true
			return m, m.topicSelectCmd()
		} else {
			m.cursor.Open = false
		}
	case "enter":
		m.cursor.Open = true
		return m, nil
	case "j":
		m.moveDown()
		return m, m.topicSelectCmd()
	case "k":
		m.moveUp()
		return m, m.topicSelectCmd()
	}

	return m, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case paho.Message:
		m.topics.Add(msg.Topic())
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) View() string {
	return m.topics.Render().String()
}
