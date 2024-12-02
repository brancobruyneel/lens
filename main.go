package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/brancobruyneel/lens/mqtt"
	"github.com/brancobruyneel/lens/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	out, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer out.Close()

	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

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
