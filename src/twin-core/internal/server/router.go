package server

import (
	"net/http"
	"strings"

	"IDIG4110/twin-core/internal/domain"
)

// NewRouter wires each endpoint collection to a single handler. The mux
// resolves only which collection a request belongs to; the collection
// handler resolves the method and whether the request targets the
// collection, a specific item, or a sub-resource.
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

// registerCollection points the bare collection path and its subtree at
// one handler, so /api/v1/homes and every path under /api/v1/homes/
// resolve in the same place.
func registerCollection(mux *http.ServeMux, collection string, handler http.Handler) {
	mux.Handle(API_ROUTE+"/"+collection, handler)
	mux.Handle(API_ROUTE+"/"+collection+"/", handler)
}

// apiSegments splits the request path into the segments below the API
// prefix: /api/v1/homes/<id>/state becomes [homes <id> state].
func apiSegments(r *http.Request) []string {
	rest := strings.TrimPrefix(r.URL.Path, API_ROUTE)
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return nil
	}
	return strings.Split(rest, "/")
}

// belowCollection returns the segments below the collection segment: for
// /api/v1/homes/<id>/state it is [<id> state], and for the bare
// collection path it is empty.
func belowCollection(r *http.Request) []string {
	segments := apiSegments(r)
	if len(segments) < 2 {
		return nil
	}
	return segments[1:]
}
