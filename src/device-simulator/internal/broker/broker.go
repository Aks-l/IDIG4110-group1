package broker

import (
	"IDIG4110/device-simulator/internal/config"
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

func (b *Broker) Publish(topic string, qos byte, payload []byte) {
	token := b.client.Publish(topic, qos, false, payload)
	if !token.WaitTimeout(publishTimeout) {
		slog.Error("Publish timed out", "topic", topic)
		return
	}
	if err := token.Error(); err != nil {
		slog.Error("Publish failed", "topic", topic, "error", err)
	}
}

func (b *Broker) Close() {
	b.client.Disconnect(disconnectTime)
}
