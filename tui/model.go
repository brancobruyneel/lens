package tui

import (
	"github.com/brancobruyneel/lens/mqtt"
	"github.com/brancobruyneel/lens/tui/history"
	"github.com/brancobruyneel/lens/tui/topics"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Model struct {
	topicTree    tea.Model
	history      tea.Model
	client       *mqtt.Client
	messages     chan paho.Message
	selectedPath []string
}

func NewModel(c *mqtt.Client, mqttHost string) *Model {
	return &Model{
		topicTree:    topics.New(mqttHost),
		history:      history.New(),
		client:       c,
		messages:     make(chan paho.Message),
		selectedPath: make([]string, 0),
	}
}

func subscribe(c *mqtt.Client, messages chan paho.Message) tea.Cmd {
	return func() tea.Msg {
		topic := "#"
		_ = c.Subscribe(topic, func(_ paho.Client, msg paho.Message) {
			messages <- msg
		})
		return nil
	}
}

func waitForMsg(messages chan paho.Message) tea.Cmd {
	return func() tea.Msg {
		msg := <-messages
		return msg
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		subscribe(m.client, m.messages),
		waitForMsg(m.messages),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case paho.Message:
		m.topicTree, _ = m.topicTree.Update(msg)
		m.history, _ = m.history.Update(msg)
		return m, waitForMsg(m.messages)
	}

	m.topicTree, _ = m.topicTree.Update(msg)
	m.history, _ = m.history.Update(msg)

	return m, nil
}

// Propagate message to child models.
func (m Model) propagate(msg tea.Msg) tea.Model {
	m.topicTree, _ = m.topicTree.Update(msg)
	m.history, _ = m.history.Update(msg)

	return m
}

func (m Model) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.topicTree.View(),
		m.history.View(),
	)
}
