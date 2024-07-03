package tui

import (
	"sort"
	"strings"

	"github.com/brancobruyneel/lens/mqtt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type TopicNode struct {
	Name     string
	Children map[string]*TopicNode
	Messages []string
}

type Model struct {
	root         *TopicNode
	client       mqtt.Client
	sub          chan paho.Message
	selectedPath []string
}

func NewModel(client mqtt.Client, serverURL string) *Model {
	return &Model{
		root: &TopicNode{
			Name:     serverURL,
			Children: make(map[string]*TopicNode),
			Messages: []string{},
		},
		client:       client,
		sub:          make(chan paho.Message),
		selectedPath: []string{"dx-climate-control", "reported"}, // Hardcoded selection
	}
}

func listenForActivity(c mqtt.Client, sub chan paho.Message) tea.Cmd {
	return func() tea.Msg {
		topic := "#"
		_ = c.Subscribe(topic, func(_ paho.Client, msg paho.Message) {
			sub <- msg
		})
		return nil
	}
}

func waitForMessage(sub chan paho.Message) tea.Cmd {
	return func() tea.Msg {
		msg := <-sub
		return msg
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		listenForActivity(m.client, m.sub),
		waitForMessage(m.sub),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case paho.Message:
		m.addMessage(msg.Topic(), string(msg.Payload()))
		return m, waitForMessage(m.sub)
	}
	return m, nil
}

func (m *Model) addMessage(topic, message string) {
	parts := strings.Split(topic, "/")
	currentNode := m.root
	for _, part := range parts {
		if _, exists := currentNode.Children[part]; !exists {
			currentNode.Children[part] = &TopicNode{
				Name:     part,
				Children: make(map[string]*TopicNode),
				Messages: []string{},
			}
		}
		currentNode = currentNode.Children[part]
	}
	currentNode.Messages = append(currentNode.Messages, message)
}

func (m Model) View() string {
	var leftColumn strings.Builder
	leftColumn.WriteString("MQTT Topics Tree:\n\n")
	m.viewNode(&leftColumn, m.root, 0, m.selectedPath)

	var rightColumn strings.Builder
	rightColumn.WriteString("Messages:\n\n")
	if selectedNode := m.getSelectedNode(); selectedNode != nil {
		for _, msg := range selectedNode.Messages {
			rightColumn.WriteString(msg + "\n")
		}
	}

	left := lipgloss.NewStyle().Width(50).Render(leftColumn.String())
	right := lipgloss.NewStyle().Width(50).Render(rightColumn.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) viewNode(builder *strings.Builder, node *TopicNode, depth int, path []string) {
	prefix := strings.Repeat("  ", depth)
	if len(path) > 0 && node.Name == path[0] {
		builder.WriteString(lipgloss.NewStyle().Bold(true).Render(prefix+node.Name) + "\n")
		path = path[1:]
	} else {
		builder.WriteString(prefix + node.Name + "\n")
	}

	// Collect and sort the children keys
	childKeys := make([]string, 0, len(node.Children))
	for key := range node.Children {
		childKeys = append(childKeys, key)
	}
	sort.Strings(childKeys)

	// Render the sorted children nodes
	for _, key := range childKeys {
		m.viewNode(builder, node.Children[key], depth+1, path)
	}
}

func (m Model) getSelectedNode() *TopicNode {
	currentNode := m.root
	for _, part := range m.selectedPath {
		if nextNode, exists := currentNode.Children[part]; exists {
			currentNode = nextNode
		} else {
			return nil
		}
	}
	return currentNode
}
