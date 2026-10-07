package dto

import "time"

type Reading struct {
	GatewayID        string         `json:"gateway_id"`
	ExternalEntityID string         `json:"external_entity_id"`
	Time             time.Time      `json:"time"`
	DeviceClass      string         `json:"device_class"`
	ValueNum         *float64       `json:"value_num,omitempty"`
	ValueText        *string        `json:"value_text,omitempty"`
	Unit             string         `json:"unit"`
	Attributes       map[string]any `json:"attributes,omitempty"`
}
