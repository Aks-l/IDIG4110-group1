package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"IDIG4110/shared/dto"
)

const (
	gw      = "11111111-1111-1111-1111-111111111111"
	eventID = "326ef27d19415c60c492fe330945f954"
)

func event(state string) dto.SensorStateEvent {
	return dto.SensorStateEvent{
		EventType: "state_changed",
		TimeFired: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC),
		Context:   dto.EventContext{ID: eventID},
		Data: dto.EventData{
			EntityID: "sensor.living_room_temperature",
			NewState: dto.StateObject{State: state, Attributes: map[string]any{
				"device_class":        "temperature",
				"unit_of_measurement": "°C",
				"friendly_name":       "Living room temperature",
			}},
		},
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
	if r.EventID != eventID {
		t.Fatalf("event_id not carried from context.id: %+v", r)
	}
	if r.DeviceClass != "temperature" || r.Unit != "°C" {
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

type failingPublisher struct{}

func (failingPublisher) PublishReading(_ context.Context, _ dto.Reading) error {
	return errors.New("kafka down")
}

func TestCreateSurvivesPublishFailure(t *testing.T) {
	repo := &fakeRepo{}
	if err := NewImplSensorIngestSvc(repo, failingPublisher{}, gw).Create(context.Background(), event("22")); err != nil {
		t.Fatalf("a stored event must survive a publish failure: %v", err)
	}
	if repo.inserted != 1 {
		t.Fatal("event not stored")
	}
}
