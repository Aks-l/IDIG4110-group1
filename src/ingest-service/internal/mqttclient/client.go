package mqttclient

import (
	"fmt"
	"log/slog"
	"time"

	"IDIG4110/ingest-service/internal/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const disconnectTime = 250 // in millis

type MqttClient struct {
	client mqtt.Client
}

func Init(cfg config.MqttConfig, msg mqtt.MessageHandler) (*MqttClient, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Url).
		SetClientID(cfg.ClientId).
		SetAutoReconnect(cfg.AutoConnect).
		SetConnectRetry(cfg.ConnectRetry).
		SetConnectRetryInterval(time.Duration(cfg.ConnectRetryInterval) * time.Second).
		SetKeepAlive(time.Duration(cfg.KeepAlive) * time.Second).
		SetPingTimeout(time.Duration(cfg.PingTimeout) * time.Second).
		SetCleanSession(cfg.CleanSession).
		SetDefaultPublishHandler(msg)

	opts.OnConnect = func(c mqtt.Client) {
		slog.Info("MQTT Connected")
	}
	
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		slog.Error("MQTT Connection lost")
	}

	c := mqtt.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &MqttClient{
		client: c,
	}, nil
}

func (c *MqttClient) Subscribe(topic string) error {
	token := c.client.Subscribe(topic, 0, nil)
	token.Wait()
	return token.Error()
}

const publishTimeout = 10 * time.Second

func (c *MqttClient) Publish(topic string, qos byte, payload []byte) error {
	token := c.client.Publish(topic, qos, false, payload)
	if !token.WaitTimeout(publishTimeout) {
		return fmt.Errorf("publish to %s timed out", topic)
	}
	return token.Error()
}

func (c *MqttClient) Close() {
	c.client.Disconnect(disconnectTime)
}
