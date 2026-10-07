package repository

import (
	"context"
	"fmt"
	"time"

	"IDIG4110/ingest-service/internal/db"
	"IDIG4110/shared/dto"
)

const (
	insertSensorDataQuery = `
		INSERT INTO ingest.sensor_data (entityID, event_type, time_fired)
		VALUES ($1, $2, $3)
	`
	findSensorDataByID = `
		SELECT entityID, event_type, time_fired
		FROM ingest.sensor_data
		WHERE entityID = $1
			AND ($2::timestamptz IS NULL OR time_fired >= $2::timestamptz)
			AND ($3::timestamptz IS NULL OR time_fired <= $3::timestamptz)
		ORDER BY time_fired ASC
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

func (r *SensorIngestRepoImpl) FindByTimeRange(ctx context.Context, entityId string, from, to *time.Time) ([]dto.SensorStateEvent, error) {
	rows, err := r.db.Conn.Query(ctx, findSensorDataByID, entityId, from, to)
	if err != nil {
		return nil, fmt.Errorf("find time by range: %w", err)
	}
	defer rows.Close()

	sensorStates := []dto.SensorStateEvent{}

	for rows.Next() {
		var data dto.SensorStateEvent

		if err := rows.Scan(&data.EntityID, &data.EventType, &data.TimeFired); err != nil {
			return nil, fmt.Errorf("scan sensor data: %w", err)
		}

		sensorStates = append(sensorStates, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error %w", err)
	}

	return sensorStates, nil
}
