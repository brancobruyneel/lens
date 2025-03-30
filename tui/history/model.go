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
	messages   []Message
	topic      string
	ready      bool
	viewport   viewport.Model
	autoScroll bool
}

func New() Model {
	return Model{
		messages:   make([]Message, 0),
		topic:      "#",
		autoScroll: true,
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

func (m *Model) filter() {
	matched := make([]int, 0)
	for i, msg := range m.messages {
		if m.topic == "#" || strings.HasPrefix(msg.Topic, m.topic) {
			matched = append(matched, i)
		}
	}

	var s strings.Builder
	for i, idx := range matched {
		if i > 0 {
			s.WriteString("\n\n") // Add extra newline for better separation
		}
		// Add topic as a header for each message
		s.WriteString("Topic: " + m.messages[idx].Topic + "\n")
		s.WriteString(m.messages[idx].Content)
	}

	m.viewport.SetContent(s.String())

	if m.autoScroll {
		m.viewport.GotoBottom()
	}
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

	m.filter()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case paho.Message:
		m.handleMQTTMessage(msg)
		return m, nil
	case config.TopicSelectMsg:
		m.topic = string(msg)
		m.filter()
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
