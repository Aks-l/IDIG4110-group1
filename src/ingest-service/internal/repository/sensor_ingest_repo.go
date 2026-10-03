package repository

import (
	"context"

	"IDIG4110/ingest-service/internal/db"
	"IDIG4110/shared/dto"
)

type SensorIngestRepoImpl struct {
	db *db.Database
}

func NewSensorIngestRepoImpl(db *db.Database) *SensorIngestRepoImpl {
	return &SensorIngestRepoImpl{
		db: db,
	}
}

func (r *SensorIngestRepoImpl) Insert(ctx context.Context, sensorData dto.SensorStateEvent) error {
	return nil
}
