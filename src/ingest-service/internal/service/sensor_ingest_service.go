package service

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/dto"
)

const publishTimeout = 5 * time.Second

type SensorIngestSvcImpl struct {
	repo      domain.SensorIngestRepo
	publisher domain.ReadingPublisher
	gatewayID string
}

// NewImplSensorIngestSvc stores events through repo and, when publisher is
// not nil, publishes them as normalized readings from gatewayID.
func NewImplSensorIngestSvc(repo domain.SensorIngestRepo, publisher domain.ReadingPublisher, gatewayID string) *SensorIngestSvcImpl {
	return &SensorIngestSvcImpl{
		repo:      repo,
		publisher: publisher,
		gatewayID: gatewayID,
	}
}

func (s *SensorIngestSvcImpl) Create(ctx context.Context, payload dto.SensorStateEvent) error {
	if err := s.repo.Insert(ctx, payload); err != nil {
		return err
	}
	slog.Info("successfully added sensor data")

	if s.publisher == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()
	if err := s.publisher.PublishReading(ctx, ToReading(s.gatewayID, payload)); err != nil {
		slog.Error("Publishing reading to kafka", "entity", payload.Data.EntityID, "error", err)
	}
	return nil
}

// ToReading converts a Home Assistant style state event into the normalized
// reading model (docs/architecture/mqtt-envelope.md). Numeric states become
// value_num, everything else value_text.
func ToReading(gatewayID string, ev dto.SensorStateEvent) dto.Reading {
	r := dto.Reading{
		GatewayID:        gatewayID,
		ExternalEntityID: ev.Data.EntityID,
		Timestamp:        ev.TimeFired,
		EventID:          ev.Context.ID,
	}

	state := strings.TrimSpace(ev.Data.NewState.State)
	if f, err := strconv.ParseFloat(state, 64); err == nil {
		r.ValueNum = &f
	} else {
		r.ValueText = &state
	}

	attributes := map[string]any{}
	for k, v := range ev.Data.NewState.Attributes {
		switch k {
		case "device_class":
			if s, ok := v.(string); ok {
				r.DeviceClass = s
			}
		case "unit_of_measurement":
			if s, ok := v.(string); ok {
				r.Unit = s
			}
		default:
			attributes[k] = v
		}
	}
	if len(attributes) > 0 {
		r.Attributes = attributes
	}
	return r
}
