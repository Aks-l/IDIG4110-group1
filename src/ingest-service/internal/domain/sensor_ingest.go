package domain

import (
	"context"

	"IDIG4110/shared/dto"
)

// should be a domain sturct for sensor ingest here
// but use shared/dto/sensorData for now

type SensorIngestRepo interface {
	Insert(ctx context.Context, sensorData dto.SensorStateEvent) error
}

type SensorIngestSvc interface {
	Create(ctx context.Context, payload dto.SensorStateEvent) error
}

