package service

import (
	"context"
	"testing"
	"time"

	"IDIG4110/shared/dto"
)

const gw = "11111111-1111-1111-1111-111111111111"

func event(state string) dto.SensorStateEvent {
	return dto.SensorStateEvent{
		EventType: "state_changed",
		TimeFired: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC),
		EntityID:  "sensor.living_room_temperature",
		NewState: dto.NewState{State: state, Attributes: map[string]string{
			"device_class":        "temperature",
			"unit_of_measurement": "°C",
			"friendly_name":       "Living room temperature",
		}},
	}
}

func TestToReadingNumericState(t *testing.T) {
	r := ToReading(gw, event("21.5"))
	if r.ValueNum == nil || *r.ValueNum != 21.5 || r.ValueText != nil {
		t.Fatalf("numeric state not mapped to value_num: %+v", r)
	}
	if r.GatewayID != gw || r.ExternalEntityID != "sensor.living_room_temperature" || r.Timestamp.IsZero() {
		t.Fatalf("core fields wrong: %+v", r)
	}
	if r.DeviceClass == nil || *r.DeviceClass != "temperature" || r.Unit == nil || *r.Unit != "°C" {
		t.Fatalf("device_class/unit not lifted out of attributes: %+v", r)
	}
	if r.Attributes["friendly_name"] != "Living room temperature" || r.Attributes["device_class"] != nil {
		t.Fatalf("attributes = %v", r.Attributes)
	}
}

func TestToReadingTextState(t *testing.T) {
	r := ToReading(gw, event("on"))
	if r.ValueText == nil || *r.ValueText != "on" || r.ValueNum != nil {
		t.Fatalf("text state not mapped to value_text: %+v", r)
	}
}

type fakeRepo struct{ inserted int }

func (f *fakeRepo) Insert(context.Context, dto.SensorStateEvent) error {
	f.inserted++
	return nil
}

type fakePublisher struct{ readings []dto.Reading }

func (f *fakePublisher) PublishReading(_ context.Context, r dto.Reading) error {
	f.readings = append(f.readings, r)
	return nil
}

func TestCreateStoresAndPublishes(t *testing.T) {
	repo, pub := &fakeRepo{}, &fakePublisher{}
	if err := NewImplSensorIngestSvc(repo, pub, gw).Create(context.Background(), event("22")); err != nil {
		t.Fatal(err)
	}
	if repo.inserted != 1 || len(pub.readings) != 1 {
		t.Fatalf("inserted=%d published=%d, want 1 and 1", repo.inserted, len(pub.readings))
	}
}

func TestCreateWithoutPublisher(t *testing.T) {
	repo := &fakeRepo{}
	if err := NewImplSensorIngestSvc(repo, nil, gw).Create(context.Background(), event("22")); err != nil {
		t.Fatal(err)
	}
	if repo.inserted != 1 {
		t.Fatal("event not stored")
	}
}
