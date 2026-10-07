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
	InsertRawMessage(ctx context.Context, at time.Time, gatewayID, topic string, payload []byte) error
	FindByTimeRange(ctx context.Context, entityId string, from, to *time.Time) ([]dto.Reading, error)
}

type SensorIngestSvc interface {
	Create(ctx context.Context, reading dto.Reading) error
	CreateRaw(ctx context.Context, at time.Time, gatewayID, topic string, payload []byte) error
	GetSensorData(ctx context.Context, entityId string, from, to *time.Time) ([]dto.Reading, error)
}

