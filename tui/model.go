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
	width  int
	height int

	topicTree tea.Model
	history   tea.Model

	client *mqtt.Client
	msgs   chan paho.Message
}

func NewModel(c *mqtt.Client, mqttHost string) *Model {
	return &Model{
		topicTree: topics.New(mqttHost),
		history:   history.New(),
		client:    c,
		msgs:      make(chan paho.Message),
	}
}

func (m Model) subscribe() tea.Cmd {
	return func() tea.Msg {
		topic := "#"
		_ = m.client.Subscribe(topic, func(_ paho.Client, msg paho.Message) {
			m.msgs <- msg
		})
		return nil
	}
}

func (m Model) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.msgs
		return msg
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.subscribe(),
		m.waitForMsg(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case paho.Message:
		return m.propagate(msg), m.waitForMsg()
	}

	return m.propagate(msg), nil
}

// Propagate message to child models.
func (m Model) propagate(msg tea.Msg) tea.Model {
	m.topicTree, _ = m.topicTree.Update(msg)
	m.history, _ = m.history.Update(msg)

	return m
}

func (m Model) View() string {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("225"))

	topicsStyle := lipgloss.NewStyle().
		Height(m.height - 2).
		Width(70).
		Padding(1).
		Inherit(border)

	historyStyle := lipgloss.NewStyle().
		Height(m.height - 2).
		Padding(1).
		Inherit(border)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		topicsStyle.Render(m.topicTree.View()),
		historyStyle.Render(m.history.View()),
	)
}
