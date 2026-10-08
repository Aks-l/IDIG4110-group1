package dto

import "time"

// Reading is the normalized read model from docs/architecture/mqtt-envelope.md.
// ingest-service publishes it on the Kafka topic twin.readings; downstream
// services consume it. Exactly one of ValueNum and ValueText is set.
type Reading struct {
	// Core.
	GatewayID        string    `json:"gateway_id" validate:"required,uuid"`            // uuid of the source gateway
	ExternalEntityID string    `json:"external_entity_id" validate:"required,max=255"` // identifier for the thing being read, verbatim from the source
	Timestamp        time.Time `json:"timestamp" validate:"required"`                  // RFC 3339, when the event happened

	// Extended.
	EventID     string         `json:"event_id,omitempty"`     // the source's unique id for this event; dedup key
	ValueNum    *float64       `json:"value_num"`              // set when the reading is numeric
	ValueText   *string        `json:"value_text"`             // set when the reading is a state or any non-numeric value
	DeviceClass string         `json:"device_class,omitempty"` // what kind of quantity or state this is
	Unit        string         `json:"unit,omitempty"`         // unit of measurement for numeric readings
	Attributes  map[string]any `json:"attributes,omitempty"`   // extra source-specific detail, preserved verbatim
}

// Key identifies the entity a reading belongs to. Entity ids are only unique
// within one gateway, so the gateway is part of the key. It is also the Kafka
// message key, which keeps all readings for one entity in order.
func (r Reading) Key() string {
	return EntityKey(r.GatewayID, r.ExternalEntityID)
}

func EntityKey(gatewayID, externalEntityID string) string {
	return gatewayID + "/" + externalEntityID
}

// Command is a request to change one entity on one gateway, as defined in
// docs/architecture/gateway-api.md. Services publish it on twin.commands;
// ingest-service forwards it to the gateway over MQTT.
type Command struct {
	ID               string         `json:"id"`
	GatewayID        string         `json:"gateway_id"`
	ExternalEntityID string         `json:"external_entity_id"`
	Command          string         `json:"command"`
	Parameters       map[string]any `json:"parameters,omitempty"`
	IssuedAt         time.Time      `json:"issued_at"`
	ExpiresAt        *time.Time     `json:"expires_at,omitempty"`
	IssuedBy         string         `json:"issued_by,omitempty"`
}

// IncidentEvent announces a new incident on twin.incidents, for the
// notification service and the UI.
type IncidentEvent struct {
	ID          string         `json:"id"`
	RuleID      string         `json:"rule_id"`
	HomeID      string         `json:"home_id"`
	Severity    string         `json:"severity"`
	Message     string         `json:"message"`
	Channels    []string       `json:"channels"`
	Context     map[string]any `json:"context"`
	TriggeredAt time.Time      `json:"triggered_at"`
}
