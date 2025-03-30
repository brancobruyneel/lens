package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/brancobruyneel/lens/mqtt"
	"github.com/brancobruyneel/lens/tui"
	tea "github.com/charmbracelet/bubbletea"
	paho "github.com/eclipse/paho.mqtt.golang"
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

	brokerURI := "mqtt://127.0.0.1:1883"
	opts := paho.NewClientOptions().
		SetClientID("mqtt-lens").
		AddBroker(brokerURI)

	mqttClient := mqtt.New(opts)
	defer mqttClient.Disconnect()

	model := tui.NewModel(mqttClient, brokerURI)

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		log.Fatalf("Failed to start TUI: %v", err)
	}
}
