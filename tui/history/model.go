package history

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/brancobruyneel/lens/tui/config"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Message struct {
	Topic   string
	Content string
}

type Model struct {
	messages      []Message
	filteredIndex []int
	topic         string
	ready         bool
	viewport      viewport.Model
	autoScroll    bool
}

func New() Model {
	return Model{
		messages:      make([]Message, 0),
		filteredIndex: make([]int, 0),
		topic:         "#", // Default to all topics
		autoScroll:    true,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		m.autoScroll = !m.autoScroll
		if m.autoScroll {
			m.viewport.GotoBottom()
		}
	case "j":
		m.autoScroll = false
		m.viewport.LineDown(1)
	case "k":
		m.autoScroll = false
		m.viewport.LineUp(1)
	case "u":
		m.autoScroll = false
		m.viewport.ViewUp()
	case "d":
		m.autoScroll = false
		m.viewport.ViewDown()
	case "gg":
		m.autoScroll = false
		m.viewport.GotoTop()
	case "G":
		m.autoScroll = false
		m.viewport.GotoBottom()
	}

	return m, nil
}

func (m *Model) filterMessages() {
	// Create a new filtered index
	filteredIndex := make([]int, 0)

	for i, msg := range m.messages {
		// If selected topic is "#" (all topics) or the message topic starts with the selected topic
		if m.topic == "#" || strings.HasPrefix(msg.Topic, m.topic) {
			filteredIndex = append(filteredIndex, i)
		}
	}

	// Update viewport content with filtered messages
	var content strings.Builder
	for i, idx := range filteredIndex {
		if i > 0 {
			content.WriteString("\n\n") // Add extra newline for better separation
		}
		// Add topic as a header for each message
		content.WriteString("Topic: " + m.messages[idx].Topic + "\n")
		content.WriteString(m.messages[idx].Content)
	}

	// Create a copy of the viewport and set its content
	viewport := m.viewport
	viewport.SetContent(content.String())
	viewport.GotoBottom()

	// Return updated model
	m.filteredIndex = filteredIndex
	m.viewport = viewport
}

func (m *Model) handleMQTTMessage(msg paho.Message) {
	var pjson bytes.Buffer
	_ = json.Indent(&pjson, msg.Payload(), "", "\t")

	out := strings.Builder{}
	_ = quick.Highlight(&out, pjson.String(), "JSON", "terminal", "")

	// Store both topic and formatted content
	m.messages = append(m.messages, Message{
		Topic:   msg.Topic(),
		Content: out.String(),
	})

	m.filterMessages()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case paho.Message:
		m.handleMQTTMessage(msg)
		return m, nil
	case config.TopicSelectMsg:
		m.topic = string(msg)
		m.filterMessages()
		return m, nil
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-76, msg.Height-4)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 76
			m.viewport.Height = msg.Height - 4
		}

		return m, nil
	}

	// m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return m.viewport.View()
}
