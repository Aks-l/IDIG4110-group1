package dto

import "time"

// NormalizedReading is the normalized reading model from
// docs/architecture/mqtt-envelope.md: adapters translate gateway-native
// payloads into this shape, it flows over the readings topic to
// ingest-service, and consumers such as twin-core receive it. A message
// missing a core field is not a reading; extended fields are nullable.
type NormalizedReading struct {
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
