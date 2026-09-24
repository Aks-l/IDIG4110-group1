package mqttclient

import (
	"time"

	"IDIG4110/ingest-service/internal/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const disconnectTime = 250 // in millis

type MqttClient struct {
	client mqtt.Client
}

func Init(cfg config.MqttConfig) (*MqttClient, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Url).
		SetClientID(cfg.ClientId).
		SetAutoReconnect(cfg.AutoConnect).
		SetConnectRetry(cfg.ConnectRetry).
		SetConnectRetryInterval(time.Duration(cfg.ConnectRetryInterval) * time.Second).
		SetKeepAlive(time.Duration(cfg.KeepAlive) * time.Second).
		SetPingTimeout(time.Duration(cfg.PingTimeout) * time.Second).
		SetCleanSession(cfg.CleanSession)

	c := mqtt.NewClient(opts)

	return &MqttClient{
		client: c,
	}, nil
}

func (c *MqttClient) Subscribe(topic string) error {
	token := c.client.Subscribe(topic, 0, nil)
	token.Wait()
	return token.Error()
}

func (c *MqttClient) Close() {
	c.client.Disconnect(disconnectTime)
}
