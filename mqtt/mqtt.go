package mqtt

import (
	"errors"
	"log/slog"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	logger *slog.Logger
	client paho.Client
	opts   *paho.ClientOptions
}

func New(opts *paho.ClientOptions) *Client {
	logger := slog.Default().
		With(slog.String("service", "mqtt-client")).
		With(slog.Any("servers", opts.Servers))

	return &Client{
		logger: logger,
		opts:   opts,
		client: paho.NewClient(opts),
	}
}

func (c *Client) Connect() error {
	if c.client == nil {
		return errors.New("client is nil")
	}

	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	c.logger.Info("connected")

	return nil
}

func (c *Client) Subscribe(topic string, callback paho.MessageHandler) error {
	if token := c.client.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	c.logger.Info("subscribed", slog.String("topic", topic))

	return nil
}

func (c *Client) Disconnect() {
	c.client.Disconnect(250)
}
