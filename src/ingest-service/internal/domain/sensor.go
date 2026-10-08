package domain

import "time"

// Sensor describes one available sensor, derived from ingest.readings and
// keyed by the same external_entity_id the sensor-data endpoint uses.
// DeviceClass and Unit are nullable, readings without them exist.
type Sensor struct {
	ExternalEntityID string    `json:"external_entity_id"`
	DeviceClass      *string   `json:"device_class"`
	Unit             *string   `json:"unit"`
	LastSeen         time.Time `json:"last_seen"`
	ReadingCount     int64     `json:"reading_count"`
}
