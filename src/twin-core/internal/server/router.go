package server

import (
	"net/http"
	"strings"

	"IDIG4110/twin-core/internal/domain"
)

// Wires each endpoint collection to one handler
// mux resolves collection only, handler resolves method and target
//
// # Inputs:
//
//   - querySvc [domain.StateQuerySvc] state query service
//   - stateSvc [domain.TwinStateSvc] twin state service
//   - structureSvc [domain.StructureSvc] structure service
//
// # Returns:
//
//   - Router handling all endpoints
func NewRouter(querySvc domain.StateQuerySvc, stateSvc domain.TwinStateSvc, structureSvc domain.StructureSvc) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealth)

	// Internal: ingest-service pushes normalized readings here.
	registerCollection(mux, "readings", handleReadings(stateSvc))

	// Structure API: the frontend reaches these through the gateway.
	registerCollection(mux, "homes", handleHomes(structureSvc, querySvc))
	registerCollection(mux, "areas", handleAreas(structureSvc))
	registerCollection(mux, "devices", handleDevices(structureSvc))
	registerCollection(mux, "entities", handleEntities(structureSvc, querySvc))
	registerCollection(mux, "relations", handleRelations(structureSvc))

	// Read model.
	registerCollection(mux, "state", handleState(querySvc))

	return mux
}

// Points collection path and its subtree at one handler
//
// # Inputs:
//
//   - mux [*http.ServeMux] router to register on
//   - collection [string] collection name
//   - handler [http.Handler] collection handler
func registerCollection(mux *http.ServeMux, collection string, handler http.Handler) {
	mux.Handle(API_ROUTE+"/"+collection, handler)
	mux.Handle(API_ROUTE+"/"+collection+"/", handler)
}

// Splits path into segments below the api prefix
//
// # Inputs:
//
//   - r [*http.Request] incoming request
//
// # Returns:
//
//   - Segments, /api/v1/homes/<id>/state -> [homes <id> state]
func apiSegments(r *http.Request) []string {
	rest := strings.TrimPrefix(r.URL.Path, API_ROUTE)
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return nil
	}
	return strings.Split(rest, "/")
}

// Returns segments below the collection segment
//
// # Inputs:
//
//   - r [*http.Request] incoming request
//
// # Returns:
//
//   - Segments after collection, empty for bare collection path
func belowCollection(r *http.Request) []string {
	segments := apiSegments(r)
	if len(segments) < 2 {
		return nil
	}
	return segments[1:]
}
