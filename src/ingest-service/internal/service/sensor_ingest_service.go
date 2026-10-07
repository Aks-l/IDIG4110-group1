package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/dto"
)

type SensorIngestSvcImpl struct {
	repo domain.SensorIngestRepo
}

func NewImplSensorIngestSvc(repo domain.SensorIngestRepo) *SensorIngestSvcImpl {
	return &SensorIngestSvcImpl{
		repo: repo,
	}
}

func (s *SensorIngestSvcImpl) Create(ctx context.Context, reading dto.Reading) error {
	if (reading.ValueNum == nil) == (reading.ValueText == nil) {
		return fmt.Errorf("reading must have exactly one of value_num or value_text")
	}
	if err := s.repo.InsertReading(ctx, reading); err != nil {
		return err
	}
	slog.Info("successfully added reading")
	return nil
}

func (s *SensorIngestSvcImpl) CreateRaw(ctx context.Context, at time.Time, gatewayID, topic string, payload []byte) error {
	return s.repo.InsertRawMessage(ctx, at, gatewayID, topic, payload)
}

func (s *SensorIngestSvcImpl) GetSensorData(ctx context.Context, entityId string, from, to *time.Time) ([]dto.Reading, error) {
	data, err := s.repo.FindByTimeRange(ctx, entityId, from, to)
	if err != nil {
		return nil, err
	}

	return data, nil
}
