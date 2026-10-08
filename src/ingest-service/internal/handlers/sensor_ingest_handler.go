package handlers

import (
	"net/http"
	"time"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/httperror"
	jsonutils "IDIG4110/shared/json-utils"
)

func GetSensors(svc domain.SensorIngestSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var sensorID *string
		if id := q.Get("sensor_id"); id != "" {
			sensorID = &id
		}

		sensors, err := svc.GetSensors(r.Context(), sensorID)
		if err != nil {
			httperror.HandleError(w, http.StatusInternalServerError, err, "failed to fetch sensors")
			return
		}
		if err := jsonutils.Encode(w, http.StatusOK, sensors); err != nil {
			httperror.HandleError(w, http.StatusBadRequest, err, httperror.ErrInternalServerError)
		}
	}
}

func GetSensorDataByTimeRange(svc domain.SensorIngestSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		entityId := q.Get("entity_id")

		fromStr := q.Get("from")
		var from *time.Time = nil
		var parsedFrom time.Time
		if fromStr != "" {
			var err error
			parsedFrom, err = time.Parse(time.RFC3339, fromStr)
			if err != nil {
				httperror.HandleError(w, http.StatusBadRequest, err, "from must be a valid RFC3339 timestamp")
				return
			}
			from = &parsedFrom
		}

		toStr := q.Get("to")
		var to *time.Time = nil
		var parsedTo time.Time
		if toStr != "" {
			var err error
			parsedTo, err = time.Parse(time.RFC3339, toStr)
			if err != nil {
				httperror.HandleError(w, http.StatusBadRequest, err, "to must be a valid RFC3339 timestamp")
				return
			}
			to = &parsedTo
		}

		sensorData, err := svc.GetSensorData(r.Context(), entityId, from, to)
		if err != nil {
			httperror.HandleError(w, http.StatusInternalServerError, err, "failed to fetch measurements")
			return
		}
		if err := jsonutils.Encode(w, http.StatusOK, sensorData); err != nil {
			httperror.HandleError(w, http.StatusBadRequest, err, httperror.ErrInternalServerError)
		}
	}
}
