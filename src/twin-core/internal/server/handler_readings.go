package server

import (
	"net/http"

	"IDIG4110/shared/dto"
	"IDIG4110/twin-core/internal/domain"
)

// handleReadings routes /api/v1/readings, where ingest-service delivers
// normalized readings one at a time.
func handleReadings(svc domain.TwinStateSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if belowCollection(r) != nil {
			respondNotFound(w)
			return
		}
		if r.Method != http.MethodPost {
			respondMethodNotAllowed(w, r, http.MethodPost)
			return
		}
		applyReading(w, r, svc)
	}
}

// applyReading decodes, validates, and applies one normalized reading.
func applyReading(w http.ResponseWriter, r *http.Request, svc domain.TwinStateSvc) {
	reading, ok := decodeBody[dto.NormalizedReading](w, r)
	if !ok {
		return
	}

	if err := svc.ApplyReading(r.Context(), reading); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
