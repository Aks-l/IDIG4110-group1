package service

import (
	"context"
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

func (s *SensorIngestSvcImpl) Create(ctx context.Context, payload dto.SensorStateEvent) error {
	if err := s.repo.Insert(ctx, payload); err != nil {
		return err
	}
	slog.Info("successfully added sensor data")
	return nil
}

func (s *SensorIngestSvcImpl) GetSensorData(ctx context.Context, entityId string, from, to *time.Time) ([]dto.SensorStateEvent, error) {
	data, err := s.repo.FindByTimeRange(ctx, entityId, from, to)
	if err != nil {
		return nil, err
	}

	return data, nil
}
