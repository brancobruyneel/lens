package mqtt

import (
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Client interface {
	Subscribe(topic string, callback paho.MessageHandler) error
	Disconnect()
}

type client struct {
	client paho.Client
}

func New(broker, clientID string) (Client, error) {
	opts := paho.NewClientOptions().AddBroker(broker).SetClientID(clientID)
	c := paho.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &client{client: c}, nil
}

func (c *client) Subscribe(topic string, callback paho.MessageHandler) error {
	if token := c.client.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *client) Disconnect() {
	c.client.Disconnect(250)
}
