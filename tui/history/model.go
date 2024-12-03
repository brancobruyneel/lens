package history

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Model struct {
	messages []string
	ready    bool
	viewport viewport.Model
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
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case paho.Message:
		var pjson bytes.Buffer
		_ = json.Indent(&pjson, msg.Payload(), "", "\t")
		out := strings.Builder{}
		_ = quick.Highlight(&out, pjson.String(), "JSON", "terminal", "")
		m.messages = append(m.messages, out.String())
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-76, msg.Height-4)
			m.viewport.SetContent("test")
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 76
			m.viewport.Height = msg.Height - 4
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, cmd
}

func (m Model) View() string {
	return m.viewport.View()
}
