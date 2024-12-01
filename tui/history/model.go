package history

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Model struct {
	messages []string
}

func New() Model {
	return Model{
		messages: make([]string, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case paho.Message:
		m.messages = append(m.messages, string(msg.Payload()))
	}
	return m, nil
}

func (m Model) View() string {
	return strings.Join(m.messages, "\n")
}
