package simulator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"IDIG4110/device-simulator/internal/broker"
	"IDIG4110/device-simulator/internal/config"
	"IDIG4110/shared/dto"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Simulator struct {
	broker   *broker.Broker
	gateway  *Gateway
	topic    string
	interval time.Duration
	qos      byte
}

func Init(b *broker.Broker, mqttConfig config.MqttConfig, house config.House) (*Simulator, error) {
	gateway, err := NewGateway(house, NewMemoryStorage())
	if err != nil {
		return nil, err
	}
	if gateway.ID() != mqttConfig.GatewayID {
		return nil, fmt.Errorf("mqtt gatewayId %q does not match house gateway_id %q", mqttConfig.GatewayID, gateway.ID())
	}
	if mqttConfig.Interval <= 0 {
		return nil, fmt.Errorf("mqtt interval must be greater than zero")
	}
	return &Simulator{
		broker: b, gateway: gateway, topic: mqttConfig.Topic,
		interval: time.Duration(mqttConfig.Interval) * time.Second, qos: 1,
	}, nil
}

func (s *Simulator) Start() error {
	if err := s.broker.Subscribe("twin/"+s.gateway.ID()+"/commands", s.handleCommand); err != nil {
		return fmt.Errorf("subscribe to gateway commands: %w", err)
	}
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		readings, err := s.gateway.Simulate(time.Now())
		if err != nil {
			return err
		}
		for _, reading := range readings {
			if err := s.publish(reading); err != nil {
				return err
			}
		}
		<-ticker.C
	}
}

func (s *Simulator) handleCommand(_ mqtt.Client, message mqtt.Message) {
	var command dto.Command
	if err := json.Unmarshal(message.Payload(), &command); err != nil {
		slog.Error("decode device command", "error", err)
		return
	}
	readings, err := s.gateway.ApplyCommand(command, time.Now())
	if err != nil {
		slog.Error("apply device command", "error", err, "device", command.ExternalEntityID)
		return
	}
	for _, reading := range readings {
		if err := s.publish(reading); err != nil {
			slog.Error("publish command state", "error", err)
		}
	}
}

func (s *Simulator) publish(reading Reading) error {
	// Device attribute values are formatted strings; widen them to the
	// envelope's map[string]any attribute type.
	attributes := make(map[string]any, len(reading.Attributes))
	for key, value := range reading.Attributes {
		attributes[key] = value
	}

	event := dto.StateEvent{
		EventType: "state_changed",
		TimeFired: reading.Time,
		EntityID:  reading.DeviceID + "." + reading.Property,
		NewState:  dto.NewState{State: reading.State, Attributes: attributes},
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode device event: %w", err)
	}
	raw := dto.RawMessage{
		Time: reading.Time, GatewayID: s.gateway.ID(),
		Topic:   fmt.Sprintf("%s/%s/%s/state", s.topic, reading.DeviceID, reading.Property),
		Payload: payload,
	}
	envelope, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("encode MQTT envelope: %w", err)
	}
	if err := s.broker.Publish(raw.Topic, s.qos, envelope); err != nil {
		return fmt.Errorf("publish %s: %w", raw.Topic, err)
	}
	slog.Info("generated device reading", "deviceID", reading.DeviceID, "property", reading.Property, "value", reading.State)
	return nil
}
