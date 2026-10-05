package repository

import (
	"context"
	"fmt"

	"IDIG4110/ingest-service/internal/db"
	"IDIG4110/shared/dto"
)

const (
	insertSensorDataQuery = `
		INSERT INTO ingest.sensor_data (entityID, event_type, time_fired)
		VALUES ($1, $2, $3)
	`
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
	tag, err := r.db.Conn.Exec(
		ctx,
		insertSensorDataQuery,
		sensorData.EntityID,
		sensorData.EventType,
		sensorData.TimeFired,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("error inserting sensor data")
	}
	return nil
}
