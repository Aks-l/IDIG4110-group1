// Package commands forwards device commands from the Kafka topic
// twin.commands to the gateway over MQTT (docs/architecture/gateway-api.md).
package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"IDIG4110/shared/dto"
)

// Sink publishes a payload on an MQTT topic.
type Sink interface {
	Publish(topic string, qos byte, payload []byte) error
}

type Forwarder struct {
	sink Sink
	now  func() time.Time
}

func NewForwarder(sink Sink) *Forwarder {
	return &Forwarder{sink: sink, now: time.Now}
}

// Topic is the MQTT topic a gateway's adapter receives commands on.
func Topic(gatewayID string) string {
	return "twin/" + gatewayID + "/commands"
}

// Handle forwards one command record. Expired commands are dropped, never
// executed late. The message is passed on unchanged with QoS 1; the adapter
// uses its id to ignore redeliveries.
func (f *Forwarder) Handle(_ context.Context, _, value []byte) error {
	var cmd dto.Command
	if err := json.Unmarshal(value, &cmd); err != nil {
		return fmt.Errorf("decoding command: %w", err)
	}
	if cmd.ID == "" || cmd.GatewayID == "" || cmd.ExternalEntityID == "" || cmd.Command == "" {
		return errors.New("command is missing id, gateway_id, external_entity_id or command")
	}
	if cmd.ExpiresAt != nil && f.now().After(*cmd.ExpiresAt) {
		slog.Warn("Dropping expired command", "id", cmd.ID, "entity", cmd.ExternalEntityID, "expired_at", cmd.ExpiresAt)
		return nil
	}
	if err := f.sink.Publish(Topic(cmd.GatewayID), 1, value); err != nil {
		return fmt.Errorf("forwarding command %s: %w", cmd.ID, err)
	}
	slog.Info("Forwarded command", "id", cmd.ID, "entity", cmd.ExternalEntityID, "command", cmd.Command, "issued_by", cmd.IssuedBy)
	return nil
}
