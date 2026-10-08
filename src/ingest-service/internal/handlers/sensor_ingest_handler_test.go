package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/dto"
)

type fakeSvc struct {
	sensorEntityID *string
	dataEntityID   string
	from, to       *time.Time
	sensors        []domain.Sensor
	data           []dto.Reading
}

func (f *fakeSvc) Create(context.Context, dto.RawMessage) error    { return nil }
func (f *fakeSvc) CreateRaw(context.Context, dto.RawMessage) error { return nil }

func (f *fakeSvc) GetSensorData(_ context.Context, entityID string, from, to *time.Time) ([]dto.Reading, error) {
	f.dataEntityID, f.from, f.to = entityID, from, to
	return f.data, nil
}

func (f *fakeSvc) GetSensors(_ context.Context, entityID *string) ([]domain.Sensor, error) {
	f.sensorEntityID = entityID
	return f.sensors, nil
}

func do(handler http.HandlerFunc, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func errorBody(t *testing.T, rec *httptest.ResponseRecorder) (int, string) {
	t.Helper()
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response not the error contract: %v", err)
	}
	return body.Code, body.Message
}

func TestGetSensorDataRequiresEntityID(t *testing.T) {
	rec := do(GetSensorDataByTimeRange(&fakeSvc{}), "/api/v1/ingest/sensor-data")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if code, msg := errorBody(t, rec); code != http.StatusBadRequest || msg != "entity_id is required" {
		t.Fatalf("body = %d %q", code, msg)
	}
}

func TestGetSensorDataRejectsBadFrom(t *testing.T) {
	rec := do(GetSensorDataByTimeRange(&fakeSvc{}), "/api/v1/ingest/sensor-data?entity_id=e&from=garbage")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetSensorDataRejectsFromAfterTo(t *testing.T) {
	rec := do(GetSensorDataByTimeRange(&fakeSvc{}), "/api/v1/ingest/sensor-data?entity_id=e&from=2026-10-08T12:00:00Z&to=2026-10-07T12:00:00Z")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if code, msg := errorBody(t, rec); code != http.StatusBadRequest || msg != "from must not be after to" {
		t.Fatalf("body = %d %q", code, msg)
	}
}

func TestGetSensorDataDefaultsWindow(t *testing.T) {
	svc := &fakeSvc{}
	before := time.Now()
	rec := do(GetSensorDataByTimeRange(svc), "/api/v1/ingest/sensor-data?entity_id=e")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.dataEntityID != "e" {
		t.Fatalf("entity_id = %q, want e", svc.dataEntityID)
	}
	if svc.from == nil {
		t.Fatal("from must default, got nil")
	}
	earliest := before.Add(-24*time.Hour - time.Minute)
	latest := time.Now().Add(-24*time.Hour + time.Minute)
	if svc.from.Before(earliest) || svc.from.After(latest) {
		t.Fatalf("default from = %v, want about %v", svc.from, before.Add(-24*time.Hour))
	}
	if svc.to != nil {
		t.Fatalf("to = %v, want nil when absent", svc.to)
	}
}

func TestGetSensorDataReturnsReadings(t *testing.T) {
	value := 21.5
	unit := "°C"
	svc := &fakeSvc{data: []dto.Reading{{
		GatewayID:        "00000000-0000-0000-0000-000000000001",
		ExternalEntityID: "e",
		Timestamp:        time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC),
		ValueNum:         &value,
		Unit:             &unit,
	}}}
	rec := do(GetSensorDataByTimeRange(svc), "/api/v1/ingest/sensor-data?entity_id=e&from=2026-10-08T00:00:00Z")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var readings []dto.Reading
	if err := json.Unmarshal(rec.Body.Bytes(), &readings); err != nil {
		t.Fatal(err)
	}
	if len(readings) != 1 || readings[0].ExternalEntityID != "e" || *readings[0].ValueNum != 21.5 {
		t.Fatalf("readings = %+v", readings)
	}
}

func TestGetSensorsPassesEntityIDFilter(t *testing.T) {
	svc := &fakeSvc{}
	rec := do(GetSensors(svc), "/api/v1/ingest/sensors?entity_id=sensor.living_room_temperature")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.sensorEntityID == nil || *svc.sensorEntityID != "sensor.living_room_temperature" {
		t.Fatalf("entity_id filter = %v", svc.sensorEntityID)
	}
}

func TestGetSensorsReturnsList(t *testing.T) {
	class := "temperature"
	svc := &fakeSvc{sensors: []domain.Sensor{{
		ExternalEntityID: "e",
		DeviceClass:      &class,
		LastSeen:         time.Now(),
		ReadingCount:     3,
	}}}
	rec := do(GetSensors(svc), "/api/v1/ingest/sensors")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var sensors []domain.Sensor
	if err := json.Unmarshal(rec.Body.Bytes(), &sensors); err != nil {
		t.Fatal(err)
	}
	if len(sensors) != 1 || sensors[0].ExternalEntityID != "e" || *sensors[0].DeviceClass != "temperature" {
		t.Fatalf("sensors = %+v", sensors)
	}
}
