// Package consumer keeps twin_state in sync with the readings stream on
// the event bus (docs/decisions/0001-kafka-event-bus.md).
package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"IDIG4110/shared/dto"
	"IDIG4110/twin-core/internal/domain"
)

// Readings applies twin.readings records to the twin state, the same
// service path POST /api/v1/readings uses.
type Readings struct {
	svc domain.TwinStateSvc
}

func NewReadings(svc domain.TwinStateSvc) *Readings {
	return &Readings{svc: svc}
}

// Handle applies one record. A decode or apply failure is returned so the
// shared consumer logs it and skips the record: liveness over completeness
// (docs/architecture/rules-engine.md, Delivery semantics).
func (r *Readings) Handle(ctx context.Context, key, value []byte) error {
	var reading dto.Reading
	if err := json.Unmarshal(value, &reading); err != nil {
		return fmt.Errorf("twin.readings decode: %w", err)
	}
	if err := r.svc.ApplyReading(ctx, reading); err != nil {
		return fmt.Errorf("twin.readings apply %s: %w", reading.ExternalEntityID, err)
	}
	return nil
}