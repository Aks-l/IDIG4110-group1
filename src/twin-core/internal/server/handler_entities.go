package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// Routes everything under /api/v1/entities
// POST creates, PATCH {id} updates, DELETE {id} removes
// {id}/state: entity with current state
//
// # Inputs:
//
//   - structureSvc [domain.StructureSvc] structure service
//   - querySvc [domain.StateQuerySvc] state query service
//
// # Returns:
//
//   - Collection handler
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

// Registers an entity
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
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

// Updates one entity
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
//   - id [string] entity id path segment
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

// Removes one entity
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
//   - id [string] entity id path segment
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

// Serves one entity with its current state
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StateQuerySvc] state query service
//   - id [string] entity id path segment
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
