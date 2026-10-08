package service

import (
	"context"
	"testing"
	"time"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/dto"
)

type fakeRepo struct {
	readings []dto.Reading
	raws     []dto.RawMessage
}

func (f *fakeRepo) InsertReading(_ context.Context, reading dto.Reading) error {
	f.readings = append(f.readings, reading)
	return nil
}

func (f *fakeRepo) InsertRawMessage(_ context.Context, raw dto.RawMessage) error {
	f.raws = append(f.raws, raw)
	return nil
}

func (f *fakeRepo) FindByTimeRange(context.Context, string, *time.Time, *time.Time) ([]dto.Reading, error) {
	return nil, nil
}

func (f *fakeRepo) FindSensors(context.Context, *string) ([]domain.Sensor, error) {
	return nil, nil
}

type fakePublisher struct{ readings []dto.Reading }

func (f *fakePublisher) PublishReading(_ context.Context, reading dto.Reading) error {
	f.readings = append(f.readings, reading)
	return nil
}

func rawReading(t *testing.T) dto.RawMessage {
	t.Helper()
	return envelope(t, dto.StateEvent{
		EventType: "state_changed",
		TimeFired: fired,
		EntityID:  "sensor.living_room_temperature",
		NewState:  dto.NewState{State: "21.5"},
	})
}

func TestCreateStoresAndPublishes(t *testing.T) {
	repo, pub := &fakeRepo{}, &fakePublisher{}
	if err := NewImplSensorIngestSvc(repo, pub).Create(context.Background(), rawReading(t)); err != nil {
		t.Fatal(err)
	}
	if len(repo.readings) != 1 || len(pub.readings) != 1 {
		t.Fatalf("stored=%d published=%d, want 1 and 1", len(repo.readings), len(pub.readings))
	}
	if pub.readings[0].ExternalEntityID != "sensor.living_room_temperature" {
		t.Fatalf("published wrong reading: %+v", pub.readings[0])
	}
}

func TestCreateWithoutPublisher(t *testing.T) {
	repo := &fakeRepo{}
	if err := NewImplSensorIngestSvc(repo, nil).Create(context.Background(), rawReading(t)); err != nil {
		t.Fatal(err)
	}
	if len(repo.readings) != 1 {
		t.Fatal("reading not stored")
	}
}

func TestCreateRawStoresMessage(t *testing.T) {
	repo := &fakeRepo{}
	raw := rawReading(t)
	if err := NewImplSensorIngestSvc(repo, nil).CreateRaw(context.Background(), raw); err != nil {
		t.Fatal(err)
	}
	if len(repo.raws) != 1 {
		t.Fatal("raw message not stored")
	}
}
