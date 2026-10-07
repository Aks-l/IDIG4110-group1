package handlers

import (
	"fmt"
	"net/http"
	"time"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/httperror"
	jsonutils "IDIG4110/shared/json-utils"
)

func GetSensorDataByTimeRange(svc domain.SensorIngestSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		entityId := q.Get("entity_id")

		fromStr := q.Get("from")
		if fromStr == "" {
			httperror.HandleError(w, http.StatusInternalServerError, fmt.Errorf(httperror.ErrBadRequest), httperror.ErrBadRequest)
			return
		}

		toStr := q.Get("to")
		if toStr == "" {
			httperror.HandleError(w, http.StatusInternalServerError, fmt.Errorf(httperror.ErrBadRequest), httperror.ErrBadRequest)
			return
		}
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			httperror.HandleError(w, http.StatusBadRequest, err, "from must be a valid RFC3339 timestamp")
			return
		}

		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			httperror.HandleError(w, http.StatusBadRequest, err, "to must be a valid RFC3339 timestamp")
			return
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
