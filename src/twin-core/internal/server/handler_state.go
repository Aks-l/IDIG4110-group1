package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// handleState routes /api/v1/state, the filterable, paginated list of
// every entity with its current state.
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

// listEntityStates parses the filter params and serves one page of
// entities.
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
