package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// Routes /api/v1/state, filterable paginated entity state list
//
// # Inputs:
//
//   - svc [domain.StateQuerySvc] state query service
//
// # Returns:
//
//   - Collection handler
func handleState(svc domain.StateQuerySvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if belowCollection(r) != nil {
			respondNotFound(w)
			return
		}
		if r.Method != http.MethodGet {
			respondMethodNotAllowed(w, r, http.MethodGet)
			return
		}
		listEntityStates(w, r, svc)
	}
}

// Parses filter params and serves one entity page
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StateQuerySvc] state query service
func listEntityStates(w http.ResponseWriter, r *http.Request, svc domain.StateQuerySvc) {
	filter, err := parseEntityFilter(r.URL.Query())
	if err != nil {
		respondError(w, err)
		return
	}

	page, err := svc.ListEntityStates(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, page)
}
