package service

import (
	"context"
	"fmt"
	"time"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/dto"
)

const publishTimeout = 5 * time.Second

type SensorIngestSvcImpl struct {
	repo      domain.SensorIngestRepo
	publisher domain.ReadingPublisher
}

// NewImplSensorIngestSvc stores events through repo and, when publisher is
// not nil, publishes them as normalized readings. The reading carries its own
// gateway_id from the message; the service does not override it.
func NewImplSensorIngestSvc(repo domain.SensorIngestRepo, publisher domain.ReadingPublisher) *SensorIngestSvcImpl {
	return &SensorIngestSvcImpl{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *SensorIngestSvcImpl) Create(ctx context.Context, raw dto.RawMessage) error {
	reading, err := Normalize(raw)
	if err != nil {
		return err
	}
	if (reading.ValueNum == nil) == (reading.ValueText == nil) {
		return fmt.Errorf("reading must have exactly one of value_num or value_text")
	}
	if err := s.repo.InsertReading(ctx, reading); err != nil {
		return err
	}

	if s.publisher == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, publishTimeout)

	defer cancel()
	if err := s.publisher.PublishReading(ctx, reading); err != nil {
		return fmt.Errorf("publishing reading: %w", err)
	}
	return nil
}

func (s *SensorIngestSvcImpl) CreateRaw(ctx context.Context, raw dto.RawMessage) error {
	return s.repo.InsertRawMessage(ctx, raw)
}

func (s *SensorIngestSvcImpl) GetSensorData(ctx context.Context, entityID string, from, to *time.Time) ([]dto.Reading, error) {
	data, err := s.repo.FindByTimeRange(ctx, entityID, from, to)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *SensorIngestSvcImpl) GetSensors(ctx context.Context, entityID *string) ([]domain.Sensor, error) {
	sensors, err := s.repo.FindSensors(ctx, entityID)
	if err != nil {
		return nil, err
	}

	return sensors, nil
}
