package main

import (
	"log"

	"github.com/brancobruyneel/lens/mqtt"
	"github.com/brancobruyneel/lens/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	serverURL := "mqtt://127.0.0.1:1883"
	mqttClient, err := mqtt.New(serverURL, "lens")
	if err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()

	model := tui.NewModel(mqttClient, serverURL)

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		log.Fatalf("Failed to start TUI: %v", err)
	}
}
