package tui

import (
	"log/slog"

	"github.com/brancobruyneel/lens/mqtt"
	"github.com/brancobruyneel/lens/tui/config"
	"github.com/brancobruyneel/lens/tui/history"
	"github.com/brancobruyneel/lens/tui/topics"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Model struct {
	width  int
	height int

	topics  tea.Model
	history tea.Model

	active config.View

	client *mqtt.Client
	msgs   chan paho.Message
}

func NewModel(c *mqtt.Client, mqttHost string) *Model {
	return &Model{
		topics:  topics.New(mqttHost),
		history: history.New(),
		client:  c,
		active:  config.HistoryView,
		msgs:    make(chan paho.Message),
	}
}

func (m *Model) switchView(view config.View) tea.Cmd {
	m.active = view
	return func() tea.Msg {
		return m.active
	}
}

func (m Model) connect() tea.Cmd {
	return func() tea.Msg {
		err := m.client.Connect()

		topic := "#"
		_ = m.client.Subscribe(topic, func(_ paho.Client, msg paho.Message) {
			m.msgs <- msg
		})

		return config.ErrorMsg(err)
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
		m.connect(),
		m.waitForMsg(),
		m.switchView(m.active),
	)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "ctrl+h":
		return m, m.switchView(config.TopicsView)
	case "ctrl+l":
		return m, m.switchView(config.HistoryView)
	}

	var cmd tea.Cmd
	switch m.active {
	case config.TopicsView:
		m.topics, cmd = m.topics.Update(msg)
		return m, cmd
	case config.HistoryView:
		m.history, cmd = m.history.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		historyCmd tea.Cmd
		topicsCmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.propagate(msg)
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case config.TopicFilterMsg:
		return m.propagate(msg)
	case paho.Message:
		m.history, historyCmd = m.history.Update(msg)
		m.topics, topicsCmd = m.topics.Update(msg)
		return m, tea.Batch(historyCmd, topicsCmd, m.waitForMsg())
	case config.ErrorMsg:
		slog.Error("received unexpected error", slog.Any("error", msg))
	}

	return m.propagate(m)
}

// Propagate tea.Msg to all child models.
func (m Model) propagate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		historyCmd tea.Cmd
		topicsCmd  tea.Cmd
	)

	m.history, historyCmd = m.history.Update(msg)
	m.topics, topicsCmd = m.topics.Update(msg)

	return m, tea.Batch(historyCmd, topicsCmd)
}

func (m Model) View() string {
	defaultColor := lipgloss.Color("225")
	selectColor := lipgloss.Color("86")

	topicsStyle := lipgloss.NewStyle().
		Height(m.height - 2).
		Width(70).
		Padding(1).
		Border(lipgloss.RoundedBorder())

	historyStyle := lipgloss.NewStyle().
		Height(m.height - 2).
		Padding(1).
		Border(lipgloss.RoundedBorder())

	if m.active == config.TopicsView {
		topicsStyle = topicsStyle.BorderForeground(selectColor)
		historyStyle = historyStyle.BorderForeground(defaultColor)
	} else {
		topicsStyle = topicsStyle.BorderForeground(defaultColor)
		historyStyle = historyStyle.BorderForeground(selectColor)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		topicsStyle.Render(m.topics.View()),
		historyStyle.Render(m.history.View()),
	)
}
