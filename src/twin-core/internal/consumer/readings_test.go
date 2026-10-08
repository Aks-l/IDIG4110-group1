package consumer

import (
	"context"
	"errors"
	"testing"

	"IDIG4110/shared/dto"
	"IDIG4110/twin-core/internal/domain"
)

type fakeSvc struct {
	applied []dto.Reading
	err     error
}

func (f *fakeSvc) ApplyReading(_ context.Context, reading dto.Reading) error {
	if f.err != nil {
		return f.err
	}
	f.applied = append(f.applied, reading)
	return nil
}

func TestHandleAppliesReading(t *testing.T) {
	svc := &fakeSvc{}
	payload := `{"gateway_id":"00000000-0000-0000-0000-000000000001","external_entity_id":"sensor.living_room_temperature","timestamp":"2026-10-08T17:31:17.799647814Z","value_num":21.5}`
	if err := NewReadings(svc).Handle(context.Background(), []byte("gateway/sensor.living_room_temperature"), []byte(payload)); err != nil {
		t.Fatal(err)
	}
	if len(svc.applied) != 1 || svc.applied[0].ExternalEntityID != "sensor.living_room_temperature" || *svc.applied[0].ValueNum != 21.5 {
		t.Fatalf("applied = %+v", svc.applied)
	}
}

func TestHandleRejectsMalformedRecord(t *testing.T) {
	svc := &fakeSvc{}
	if err := NewReadings(svc).Handle(context.Background(), nil, []byte("{")); err == nil {
		t.Fatal("want error for malformed record")
	}
	if len(svc.applied) != 0 {
		t.Fatal("malformed record must not be applied")
	}
}

func TestHandleReturnsApplyError(t *testing.T) {
	want := errors.New("boom")
	payload := `{"external_entity_id":"e","value_num":1}`
	if err := NewReadings(&fakeSvc{err: want}).Handle(context.Background(), nil, []byte(payload)); err == nil {
		t.Fatal("want apply error to surface for logging and skipping")
	}
}

var _ domain.TwinStateSvc = (*fakeSvc)(nil)