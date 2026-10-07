package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// handleEntities routes everything under /api/v1/entities:
//
//   - the collection: POST registers an entity
//   - one entity by id: PATCH updates it, DELETE removes it
//   - sub-resource: GET {id}/state is the entity with its current state
func handleEntities(structureSvc domain.StructureSvc, querySvc domain.StateQuerySvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest := belowCollection(r)

		switch {
		case len(rest) == 0:
			if r.Method != http.MethodPost {
				respondMethodNotAllowed(w, r, http.MethodPost)
				return
			}
			createEntity(w, r, structureSvc)
		case len(rest) == 1:
			switch r.Method {
			case http.MethodPatch:
				updateEntity(w, r, structureSvc, rest[0])
			case http.MethodDelete:
				deleteEntity(w, r, structureSvc, rest[0])
			default:
				respondMethodNotAllowed(w, r, http.MethodPatch, http.MethodDelete)
			}
		case len(rest) == 2 && rest[1] == "state":
			if r.Method != http.MethodGet {
				respondMethodNotAllowed(w, r, http.MethodGet)
				return
			}
			getEntityState(w, r, querySvc, rest[0])
		default:
			respondNotFound(w)
		}
	}
}

// createEntity registers an entity.
func createEntity(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc) {
	req, ok := decodeBody[domain.CreateEntityRequest](w, r)
	if !ok {
		return
	}
	entity, err := svc.CreateEntity(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondCreated(w, entity)
}

// updateEntity modifies one entity.
func updateEntity(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	entityID, ok := validID(w, id, "entity")
	if !ok {
		return
	}
	req, ok := decodeBody[domain.UpdateEntityRequest](w, r)
	if !ok {
		return
	}
	entity, err := svc.UpdateEntity(r.Context(), entityID, req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, entity)
}

// deleteEntity removes one entity.
func deleteEntity(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	entityID, ok := validID(w, id, "entity")
	if !ok {
		return
	}
	if err := svc.DeleteEntity(r.Context(), entityID); err != nil {
		respondError(w, err)
		return
	}
	respondNoContent(w)
}

// getEntityState serves one entity with its current state.
func getEntityState(w http.ResponseWriter, r *http.Request, svc domain.StateQuerySvc, id string) {
	entityID, ok := validID(w, id, "entity")
	if !ok {
		return
	}
	state, err := svc.GetEntityState(r.Context(), entityID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, state)
}
