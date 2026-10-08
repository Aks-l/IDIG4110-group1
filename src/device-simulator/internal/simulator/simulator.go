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
		eventId := fmt.Sprintf("b%015d", i+1)
		go s.run(eventId)
	}
	select {}
}

func (s *Simulator) run(id string) {
	sensor := NewTemperatureSensor(id)

	for {
		event := sensor.Generate()
		//mock := GenerateMockData(id)

		slog.Info(
			"generated temperature",
			"deviceId", id,
			"temperature", event.NewState.State,
			"unit", event.NewState.Attributes["unit_of_measurement"],
		)

		payload, err := json.Marshal(event)
		if err != nil {
			slog.Error("JSON marshal failed",
				"devicedID", id,
				"error", err,
			)
			continue
		}

		topic := fmt.Sprintf("%s/%s/%s", s.topic, id, "state")

		if err := s.broker.Publish(topic, s.qos, payload); err != nil {
			slog.Error(
				"failed to publish message",
				"deviceId", id,
				"topic", topic,
				"error", err,
			)
		}

		time.Sleep(time.Duration(s.interval) * time.Second)
	}
}
