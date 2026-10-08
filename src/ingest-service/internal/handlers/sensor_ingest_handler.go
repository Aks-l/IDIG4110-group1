package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/httperror"
	jsonutils "IDIG4110/shared/json-utils"
)

// sensorDataDefaultWindow bounds a sensor-data request when the caller sends
// no from time; without it one request scans the entity's whole hypertable
// history.
const sensorDataDefaultWindow = 24 * time.Hour

// GetSensors lists sensors seen recently, optional entity_id filter
func GetSensors(svc domain.SensorIngestSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var entityID *string
		if id := q.Get("entity_id"); id != "" {
			entityID = &id
		}

		sensors, err := svc.GetSensors(r.Context(), entityID)
		if err != nil {
			httperror.HandleError(w, http.StatusInternalServerError, err, "failed to fetch sensors")
			return
		}
		if err := jsonutils.Encode(w, http.StatusOK, sensors); err != nil {
			// The 200 headers are already written, a second response is
			// impossible; log instead.
			slog.Error("encoding sensors response", "error", err)
		}
	}
}

// GetSensorDataByTimeRange returns one entity's readings, from inclusive
// (default: the last 24 hours), to inclusive and optional
func GetSensorDataByTimeRange(svc domain.SensorIngestSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		entityID := q.Get("entity_id")
		if entityID == "" {
			httperror.HandleError(w, http.StatusBadRequest, nil, "entity_id is required")
			return
		}

		from, err := parseOptionalTime(q.Get("from"))
		if err != nil {
			httperror.HandleError(w, http.StatusBadRequest, err, "from must be a valid RFC3339 timestamp")
			return
		}
		to, err := parseOptionalTime(q.Get("to"))
		if err != nil {
			httperror.HandleError(w, http.StatusBadRequest, err, "to must be a valid RFC3339 timestamp")
			return
		}
		if from != nil && to != nil && from.After(*to) {
			httperror.HandleError(w, http.StatusBadRequest, nil, "from must not be after to")
			return
		}
		if from == nil {
			defaultFrom := time.Now().Add(-sensorDataDefaultWindow)
			from = &defaultFrom
		}

		sensorData, err := svc.GetSensorData(r.Context(), entityID, from, to)
		if err != nil {
			httperror.HandleError(w, http.StatusInternalServerError, err, "failed to fetch measurements")
			return
		}
		if err := jsonutils.Encode(w, http.StatusOK, sensorData); err != nil {
			// The 200 headers are already written, a second response is
			// impossible; log instead.
			slog.Error("encoding sensor data response", "error", err)
		}
	}
}

// parseOptionalTime parses an RFC3339 timestamp, nil for an absent param
func parseOptionalTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
