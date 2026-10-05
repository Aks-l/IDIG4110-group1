package service

import (
	"context"

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
	return nil
}
