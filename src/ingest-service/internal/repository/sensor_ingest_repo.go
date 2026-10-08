package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"IDIG4110/ingest-service/internal/db"
	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/dto"
)

const (
	insertReadingQuery = `
		INSERT INTO ingest.readings (
			time, gateway_id, external_entity_id, event_id, device_class,
			value_num, value_text, unit, attributes
		)
		VALUES ($1, $2::uuid, $3, $4, $5, $6, $7, $8, $9)
	`
	insertRawMessageQuery = `
		INSERT INTO ingest.raw_messages (time, gateway_id, topic, payload)
		VALUES ($1, $2::uuid, $3, $4)
	`
	findReadingsByTimeRange = `
		SELECT time, gateway_id, external_entity_id, event_id, device_class, value_num, value_text, unit
		FROM ingest.readings
		WHERE external_entity_id = $1
			AND ($2::timestamptz IS NULL OR time >= $2::timestamptz)
			AND ($3::timestamptz IS NULL OR time <= $3::timestamptz)
		ORDER BY time ASC
	`
	findSensorsQuery = `
		SELECT event_id, external_entity_id, device_class, unit,
			MAX(time) AS last_seen,
			COUNT(*) AS reading_count
		FROM ingest.readings
		WHERE event_id IS NOT NULL
			AND ($1::text IS NULL OR event_id::text = $1)
		GROUP BY event_id, external_entity_id, device_class, unit
		ORDER BY event_id
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

func (r *SensorIngestRepoImpl) InsertReading(ctx context.Context, reading dto.Reading) error {
	attributes, err := marshalAttributes(reading.Attributes)
	if err != nil {
		return err
	}

	tag, err := r.db.Conn.Exec(
		ctx,
		insertReadingQuery,
		reading.Timestamp,
		reading.GatewayID,
		reading.ExternalEntityID,
		reading.EventID,
		reading.DeviceClass,
		reading.ValueNum,
		reading.ValueText,
		reading.Unit,
		attributes,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("error inserting reading")
	}
	return nil
}

func (r *SensorIngestRepoImpl) InsertRawMessage(ctx context.Context, raw dto.RawMessage) error {
	tag, err := r.db.Conn.Exec(
		ctx,
		insertRawMessageQuery,
		raw.Time,
		raw.GatewayID,
		raw.Topic,
		raw.Payload,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("error inserting raw message")
	}
	return nil
}

func (r *SensorIngestRepoImpl) FindByTimeRange(ctx context.Context, entityId string, from, to *time.Time) ([]dto.Reading, error) {
	rows, err := r.db.Conn.Query(ctx, findReadingsByTimeRange, entityId, from, to)
	if err != nil {
		return nil, fmt.Errorf("find readings by time range: %w", err)
	}
	defer rows.Close()

	readings := []dto.Reading{}

	for rows.Next() {
		var reading dto.Reading

		if err := rows.Scan(
			&reading.Timestamp,
			&reading.GatewayID,
			&reading.ExternalEntityID,
			&reading.EventID,
			&reading.DeviceClass,
			&reading.ValueNum,
			&reading.ValueText,
			&reading.Unit,
		); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}

		readings = append(readings, reading)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return readings, nil
}

func (r *SensorIngestRepoImpl) FindSensors(ctx context.Context, sensorID *string) ([]domain.Sensor, error) {
	rows, err := r.db.Conn.Query(ctx, findSensorsQuery, sensorID)
	if err != nil {
		return nil, fmt.Errorf("find sensors: %w", err)
	}
	defer rows.Close()

	sensors := []domain.Sensor{}

	for rows.Next() {
		var sensor domain.Sensor

		if err := rows.Scan(
			&sensor.SensorID,
			&sensor.ExternalEntityID,
			&sensor.DeviceClass,
			&sensor.Unit,
			&sensor.LastSeen,
			&sensor.ReadingCount,
		); err != nil {
			return nil, fmt.Errorf("scan sensor: %w", err)
		}

		sensors = append(sensors, sensor)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return sensors, nil
}

func marshalAttributes(attributes map[string]any) ([]byte, error) {
	if attributes == nil {
		return nil, nil
	}
	data, err := json.Marshal(attributes)
	if err != nil {
		return nil, fmt.Errorf("marshal attributes: %w", err)
	}
	return data, nil
}
