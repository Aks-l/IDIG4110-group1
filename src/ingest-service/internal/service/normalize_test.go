package service

import (
	"encoding/json"
	"testing"
	"time"

	"IDIG4110/shared/dto"
)

var fired = time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

// envelope builds a RawMessage whose envelope time deliberately differs from
// the event time, so tests can tell the two apart.
func envelope(t *testing.T, event dto.StateEvent) dto.RawMessage {
	t.Helper()
	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return dto.RawMessage{
		Time:      fired.Add(time.Hour),
		GatewayID: "00000000-0000-0000-0000-000000000001",
		Topic:     "sensors/some-entity/state",
		Payload:   payload,
	}
}

func TestNormalizeTakesIdentityFromPayload(t *testing.T) {
	r, err := Normalize(envelope(t, dto.StateEvent{
		EventType: "state_changed",
		TimeFired: fired,
		EntityID:  "sensor.living_room_temperature",
		NewState:  dto.NewState{State: "21.5"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if r.ExternalEntityID != "sensor.living_room_temperature" {
		t.Fatalf("external_entity_id = %q, want the payload entity id, not the topic", r.ExternalEntityID)
	}
	if r.GatewayID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("gateway_id = %q", r.GatewayID)
	}
	if !r.Timestamp.Equal(fired) {
		t.Fatalf("timestamp = %v, want the event's time_fired %v", r.Timestamp, fired)
	}
	if r.EventID != nil {
		t.Fatalf("event_id = %v, want nil unless the source stamps event ids", *r.EventID)
	}
}

func TestNormalizeFallsBackToEnvelopeTime(t *testing.T) {
	r, err := Normalize(envelope(t, dto.StateEvent{
		EntityID: "sensor.living_room_temperature",
		NewState: dto.NewState{State: "21.5"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Timestamp.Equal(fired.Add(time.Hour)) {
		t.Fatalf("timestamp = %v, want the envelope time as fallback", r.Timestamp)
	}
}

func TestNormalizeRejectsMissingEntityID(t *testing.T) {
	if _, err := Normalize(envelope(t, dto.StateEvent{NewState: dto.NewState{State: "21.5"}})); err == nil {
		t.Fatal("want an error when the payload has no entity_id")
	}
}

func TestNormalizeNumericState(t *testing.T) {
	r, err := Normalize(envelope(t, dto.StateEvent{
		TimeFired: fired,
		EntityID:  "sensor.living_room_temperature",
		NewState:  dto.NewState{State: " 21.5 "},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if r.ValueNum == nil || *r.ValueNum != 21.5 || r.ValueText != nil {
		t.Fatalf("numeric state not mapped to value_num: %+v", r)
	}
}

func TestNormalizeTextAndSentinelStates(t *testing.T) {
	// Sentinels carry meaning (device offline, no data yet) and must stay
	// in value_text, never parsed to a number or dropped.
	for _, state := range []string{"on", "unavailable", "unknown"} {
		r, err := Normalize(envelope(t, dto.StateEvent{
			TimeFired: fired,
			EntityID:  "binary_sensor.hall",
			NewState:  dto.NewState{State: state},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if r.ValueText == nil || *r.ValueText != state || r.ValueNum != nil {
			t.Fatalf("state %q not kept verbatim in value_text: %+v", state, r)
		}
	}
}

func TestNormalizeLiftsClassAndUnitAndKeepsAttributes(t *testing.T) {
	r, err := Normalize(envelope(t, dto.StateEvent{
		TimeFired: fired,
		EntityID:  "sensor.living_room_temperature",
		NewState: dto.NewState{
			State: "21.5",
			Attributes: map[string]any{
				"device_class":        "temperature",
				"unit_of_measurement": "°C",
				"friendly_name":       "Living room temperature",
				"restored":            true,
			},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if r.DeviceClass == nil || *r.DeviceClass != "temperature" {
		t.Fatalf("device_class not lifted: %+v", r)
	}
	if r.Unit == nil || *r.Unit != "°C" {
		t.Fatalf("unit not lifted: %+v", r)
	}
	if r.Attributes["device_class"] != nil || r.Attributes["unit_of_measurement"] != nil {
		t.Fatalf("lifted attributes must not repeat in attributes: %v", r.Attributes)
	}
	if r.Attributes["friendly_name"] != "Living room temperature" {
		t.Fatalf("attributes not preserved verbatim: %v", r.Attributes)
	}
	if r.Attributes["restored"] != true {
		t.Fatalf("non-string attributes must survive: %v", r.Attributes)
	}
}

func TestNormalizeNonStringClassIsIgnored(t *testing.T) {
	r, err := Normalize(envelope(t, dto.StateEvent{
		TimeFired: fired,
		EntityID:  "binary_sensor.hall",
		NewState: dto.NewState{
			State:      "on",
			Attributes: map[string]any{"device_class": 42},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if r.DeviceClass != nil {
		t.Fatalf("non-string device_class must not be lifted: %+v", r)
	}
}
