package broker

import (
	"IDIG4110/device-simulator/internal/config"
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	disconnectTime = 250 // milliseconds
	publishTimeout = 10 * time.Second
)

type Broker struct {
	client mqtt.Client
}

func Init(cfg config.MqttConfig) (*Broker, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Url).
		SetClientID(cfg.ClientId).
		SetAutoReconnect(cfg.AutoConnect).
		SetConnectRetry(cfg.ConnectRetry).
		SetConnectRetryInterval(time.Duration(cfg.ConnectRetryInterval) * time.Second).
		SetKeepAlive(time.Duration(cfg.KeepAlive) * time.Second).
		SetPingTimeout(time.Duration(cfg.PingTimeout) * time.Second).
		SetCleanSession(cfg.CleanSession)

	opts.OnConnect = func(c mqtt.Client) {
		slog.Info("MQTT Connected")
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		slog.Info("MQTT Connection lost", "error", err)
	}

	c := mqtt.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &Broker{
		client: c,
	}, nil
}

func (b *Broker) Publish(topic string, qos byte, payload []byte) error {
	token := b.client.Publish(topic, qos, false, payload)
	if !token.WaitTimeout(publishTimeout) {
		return fmt.Errorf("publishing to %s timed out", topic)
	}
	if err := token.Error(); err != nil {
		return err
	}
	slog.Info("published MQTT message", "topic", topic, "qos", qos)
	return nil
}

func (b *Broker) Subscribe(topic string, handler mqtt.MessageHandler) error {
	token := b.client.Subscribe(topic, 1, handler)
	token.Wait()
	return token.Error()
}

func (b *Broker) Close() {
	b.client.Disconnect(disconnectTime)
}
