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
		return m, nil
	}

	// All other navigation keys disable auto-scroll
	m.autoScroll = false

	switch msg.String() {
	case "j":
		m.viewport.LineDown(1)
	case "k":
		m.viewport.LineUp(1)
	case "u":
		m.viewport.ViewUp()
	case "d":
		m.viewport.ViewDown()
	case "gg":
		m.viewport.GotoTop()
	case "G":
		m.viewport.GotoBottom()
	}

	return m, nil
}

func (m *Model) filter() {
	var s strings.Builder
	for _, msg := range m.messages {
		if m.topic != "#" && !strings.HasPrefix(msg.Topic, m.topic) {
			continue
		}

		if s.Len() > 0 {
			s.WriteString("\n\n")
		}

		s.WriteString("Topic: " + msg.Topic + "\n")
		s.WriteString(msg.Content)
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
	case config.TopicFilterMsg:
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
