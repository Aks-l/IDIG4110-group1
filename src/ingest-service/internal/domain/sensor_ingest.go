package domain

import (
	"context"
	"time"

	"IDIG4110/shared/dto"
)

// should be a domain sturct for sensor ingest here
// but use shared/dto/reading for now

type SensorIngestRepo interface {
	InsertReading(ctx context.Context, reading dto.Reading) error
	InsertRawMessage(ctx context.Context, raw dto.RawMessage) error
	FindByTimeRange(ctx context.Context, entityId string, from, to *time.Time) ([]dto.Reading, error)
	FindSensors(ctx context.Context, sensorID *string) ([]Sensor, error)
}

type SensorIngestSvc interface {
	Create(ctx context.Context, raw dto.RawMessage) error
	CreateRaw(ctx context.Context, raw dto.RawMessage) error
	GetSensorData(ctx context.Context, entityId string, from, to *time.Time) ([]dto.Reading, error)
	GetSensors(ctx context.Context, sensorID *string) ([]Sensor, error)
}

