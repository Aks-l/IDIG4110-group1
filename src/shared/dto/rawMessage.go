package dto

import (
	"encoding/json"
	"time"
)

// RawMessage is the wire format published by gateways, mirroring the
// ingest.raw_messages table (raw MQTT data for audit and analytics).
type RawMessage struct {
	Time      time.Time       `json:"time"`
	GatewayID string          `json:"gateway_id"`
	Topic     string          `json:"topic"`
	Payload   json.RawMessage `json:"payload"`
}
