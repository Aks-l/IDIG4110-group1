package repository

import (
	"IDIG4110/ingest-service/internal/db"
	"context"
)

type SensorIngestRepoImpl struct {
	db *db.Database
}

func NewSensorIngestRepoImpl(db *db.Database) *SensorIngestRepoImpl{
	return &SensorIngestRepoImpl{
		db: db,
	}
}

func (r *SensorIngestRepoImpl) Insert(ctx context.Context) error {
	return nil
}
