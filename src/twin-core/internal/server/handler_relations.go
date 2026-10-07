package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// handleRelations routes everything under /api/v1/relations: POST
// registers an edge, PATCH {id} updates one, DELETE {id} removes one.
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

// createRelation registers one edge of a home's graph.
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

// updateRelation changes an edge's metadata.
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

// deleteRelation removes one edge.
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
