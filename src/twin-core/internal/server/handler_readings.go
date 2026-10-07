package server

import (
	"net/http"

	"IDIG4110/shared/dto"
	"IDIG4110/twin-core/internal/domain"
)

// Routes /api/v1/readings, ingest pushes one normalized reading at a time
//
// # Inputs:
//
//   - svc [domain.TwinStateSvc] twin state service
//
// # Returns:
//
//   - Collection handler
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

// Decodes, validates and applies one normalized reading
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.TwinStateSvc] twin state service
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
