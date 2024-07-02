package tui

import (
	"github.com/brancobruyneel/lens/mqtt"
	tea "github.com/charmbracelet/bubbletea"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Model struct {
	messages []string
	client   mqtt.Client
	sub      chan paho.Message
}

func NewModel(client mqtt.Client) *Model {
	return &Model{
		messages: make([]string, 0),
		client:   client,
		sub:      make(chan paho.Message), // Initialize the channel
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case paho.Message:
		m.messages = append(m.messages, string(msg.Payload()))
		return m, waitForMessage(m.sub) // wait for next message
	}

	return m, nil
}

func (m Model) View() string {
	s := "MQTT Messages:\n\n"

	for _, message := range m.messages {
		s += message + "\n"
	}
	return s
}
