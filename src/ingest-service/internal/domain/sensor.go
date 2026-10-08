package domain

import "time"

// Sensor describes an available sensor, derived from ingest.readings,
// and is used as the client-facing response type.
type Sensor struct {
	SensorID         string    `json:"sensor_id"`
	ExternalEntityID string    `json:"external_entity_id"`
	DeviceClass      string    `json:"device_class"`
	Unit             string    `json:"unit"`
	LastSeen         time.Time `json:"last_seen"`
	ReadingCount     int64     `json:"reading_count"`
}
