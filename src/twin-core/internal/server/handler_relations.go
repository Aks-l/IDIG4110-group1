package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// Routes everything under /api/v1/relations
// POST creates, PATCH {id} updates, DELETE {id} removes
//
// # Inputs:
//
//   - svc [domain.StructureSvc] structure service
//
// # Returns:
//
//   - Collection handler
func handleRelations(svc domain.StructureSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest := belowCollection(r)

		switch {
		case len(rest) == 0:
			if r.Method != http.MethodPost {
				respondMethodNotAllowed(w, r, http.MethodPost)
				return
			}
			createRelation(w, r, svc)
		case len(rest) == 1:
			switch r.Method {
			case http.MethodPatch:
				updateRelation(w, r, svc, rest[0])
			case http.MethodDelete:
				deleteRelation(w, r, svc, rest[0])
			default:
				respondMethodNotAllowed(w, r, http.MethodPatch, http.MethodDelete)
			}
		default:
			respondNotFound(w)
		}
	}
}

// Registers one edge of a home's graph
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
func createRelation(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc) {
	req, ok := decodeBody[domain.CreateRelationRequest](w, r)
	if !ok {
		return
	}
	relation, err := svc.CreateRelation(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondCreated(w, relation)
}

// Updates one edge's metadata
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
//   - id [string] relation id path segment
func updateRelation(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	relationID, ok := validID(w, id, "relation")
	if !ok {
		return
	}
	req, ok := decodeBody[domain.UpdateRelationRequest](w, r)
	if !ok {
		return
	}
	relation, err := svc.UpdateRelation(r.Context(), relationID, req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, relation)
}

// Removes one edge
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
//   - id [string] relation id path segment
func deleteRelation(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	relationID, ok := validID(w, id, "relation")
	if !ok {
		return
	}
	if err := svc.DeleteRelation(r.Context(), relationID); err != nil {
		respondError(w, err)
		return
	}
	respondNoContent(w)
}
