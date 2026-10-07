package service

import (
	"context"
	"fmt"

	"IDIG4110/shared/dto"
	"IDIG4110/twin-core/internal/domain"
)

type TwinStateSvcImpl struct {
	repo domain.TwinStateRepo
}

func NewImplTwinStateSvc(repo domain.TwinStateRepo) *TwinStateSvcImpl {
	return &TwinStateSvcImpl{
		repo: repo,
	}
}

// Applies reading after checking value_num/value_text xor contract
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - reading [dto.NormalizedReading] the reading to apply
//
// # Returns:
//
//   - ErrInvalidReading when both or neither value set
func (s *TwinStateSvcImpl) ApplyReading(ctx context.Context, reading dto.NormalizedReading) error {
	if (reading.ValueNum == nil) == (reading.ValueText == nil) {
		return fmt.Errorf("%w: exactly one of value_num and value_text must be set", domain.ErrInvalidReading)
	}
	return s.repo.ApplyReading(ctx, reading)
}
