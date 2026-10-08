package simulator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"IDIG4110/device-simulator/internal/broker"
	"IDIG4110/device-simulator/internal/config"
	"IDIG4110/shared/dto"
)

type Simulator struct {
	broker      *broker.Broker
	interval    int
	topic       string
	deviceCount int
	gatewayID   string
	qos         byte
}

func Init(b *broker.Broker, cfg config.MqttConfig) *Simulator {
	return &Simulator{
		broker:      b,
		interval:    cfg.Interval,
		topic:       cfg.Topic,
		deviceCount: cfg.DeviceCount,
		gatewayID:   cfg.GatewayID,
		qos:         byte(0),
	}
}

func (s *Simulator) Start() {
	for i := 0; i < s.deviceCount; i++ {
		entityID := fmt.Sprintf("00000000-0000-0000-0000-%012d", i+1)
		go s.run(entityID)
	}
	select {}
}

func (s *Simulator) run(id string) {
	temperature := 20.0
	for {
		temperature = NextTemperature(temperature)

		slog.Info(
			"generated temperature",
			"deviceID", id,
			"temperature", temperature,
			"unit", "°C",
		)

		event := GenerateMockData(id, temperature)

		payload, _ := json.Marshal(event)

		raw := dto.RawMessage{
			Time:      event.TimeFired,
			GatewayID: s.gatewayID,
			Topic:     fmt.Sprintf("%s/%s/%s", s.topic, id, "state"),
			Payload:   payload,
		}

		envelope, _ := json.Marshal(raw)

		s.broker.Publish(raw.Topic, s.qos, envelope)

		time.Sleep(time.Duration(s.interval) * time.Second)
	}
}
