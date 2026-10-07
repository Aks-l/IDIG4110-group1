package server

import (
	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/ingest-service/internal/handlers"
	"net/http"
)

func NewRouter(
	svc domain.SensorIngestSvc,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET " +SENSOR_DATA, handlers.GetSensorDataByTimeRange(svc))
	mux.HandleFunc("GET " +SENSORS, handlers.GetSensors(svc))

	return mux
}
