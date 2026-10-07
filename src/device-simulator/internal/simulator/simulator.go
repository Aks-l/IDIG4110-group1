package simulator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"IDIG4110/device-simulator/internal/broker"
	"IDIG4110/device-simulator/internal/config"
)

type Simulator struct {
	broker      *broker.Broker
	interval    int
	topic       string
	deviceCount int
	qos         byte
}

func Init(b *broker.Broker, cfg config.MqttConfig) *Simulator {
	return &Simulator{
		broker:      b,
		interval:    cfg.Interval,
		topic:       cfg.Topic,
		deviceCount: cfg.DeviceCount,
		qos:         byte(0),
	}
}

func (s *Simulator) Start() {
	for i := 0; i < s.deviceCount; i++ {
		entityID := fmt.Sprintf("sensor.mock_temperature_%02d", i+1)
		go s.run(entityID)
	}
	select {}
}

func (s *Simulator) run(id string) {
	for {
		mock := GenerateMockData(id)
		payload, err := json.Marshal(mock)
		if err != nil {
			slog.Error("JSON marshal failed", "error", err)
		}

		topic := fmt.Sprintf("%s/%s/%s", s.topic, id, "state")
		s.broker.Publish(topic, s.qos, payload)

		time.Sleep(time.Duration(s.interval) * time.Second)
	}
}
