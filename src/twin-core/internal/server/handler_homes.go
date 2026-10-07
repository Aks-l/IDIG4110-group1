package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// Routes everything under /api/v1/homes
// POST creates, GET lists, PATCH {id} updates, DELETE {id} removes
// {id}/state: dashboard view, {id}/relations: relation graph
//
// # Inputs:
//
//   - structureSvc [domain.StructureSvc] structure service
//   - querySvc [domain.StateQuerySvc] state query service
//
// # Returns:
//
//   - Collection handler
func handleHomes(structureSvc domain.StructureSvc, querySvc domain.StateQuerySvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest := belowCollection(r)

		switch {
		case len(rest) == 0:
			switch r.Method {
			case http.MethodPost:
				createHome(w, r, structureSvc)
			case http.MethodGet:
				listHomes(w, r, querySvc)
			default:
				respondMethodNotAllowed(w, r, http.MethodPost, http.MethodGet)
			}
		case len(rest) == 1:
			switch r.Method {
			case http.MethodPatch:
				updateHome(w, r, structureSvc, rest[0])
			case http.MethodDelete:
				deleteHome(w, r, structureSvc, rest[0])
			default:
				respondMethodNotAllowed(w, r, http.MethodPatch, http.MethodDelete)
			}
		case len(rest) == 2:
			switch rest[1] {
			case "state":
				if r.Method != http.MethodGet {
					respondMethodNotAllowed(w, r, http.MethodGet)
					return
				}
				getHomeState(w, r, querySvc, rest[0])
			case "relations":
				if r.Method != http.MethodGet {
					respondMethodNotAllowed(w, r, http.MethodGet)
					return
				}
				listHomeRelations(w, r, structureSvc, rest[0])
			default:
				respondNotFound(w)
			}
		default:
			respondNotFound(w)
		}
	}
}

// Registers a home
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
func createHome(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc) {
	req, ok := decodeBody[domain.CreateHomeRequest](w, r)
	if !ok {
		return
	}
	home, err := svc.CreateHome(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondCreated(w, home)
}

// Serves homes list
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StateQuerySvc] state query service
func listHomes(w http.ResponseWriter, r *http.Request, svc domain.StateQuerySvc) {
	homes, err := svc.ListHomes(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, homes)
}

// Updates one home
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
//   - id [string] home id path segment
func updateHome(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	homeID, ok := validID(w, id, "home")
	if !ok {
		return
	}
	req, ok := decodeBody[domain.UpdateHomeRequest](w, r)
	if !ok {
		return
	}
	home, err := svc.UpdateHome(r.Context(), homeID, req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, home)
}

// Removes one home
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
//   - id [string] home id path segment
func deleteHome(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	homeID, ok := validID(w, id, "home")
	if !ok {
		return
	}
	if err := svc.DeleteHome(r.Context(), homeID); err != nil {
		respondError(w, err)
		return
	}
	respondNoContent(w)
}

// Serves one home's dashboard view
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StateQuerySvc] state query service
//   - id [string] home id path segment
func getHomeState(w http.ResponseWriter, r *http.Request, svc domain.StateQuerySvc, id string) {
	homeID, ok := validID(w, id, "home")
	if !ok {
		return
	}
	state, err := svc.GetHomeState(r.Context(), homeID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, state)
}

// Serves one home's relation edges
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
//   - id [string] home id path segment
func listHomeRelations(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	homeID, ok := validID(w, id, "home")
	if !ok {
		return
	}
	relations, err := svc.ListRelationsByHome(r.Context(), homeID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, relations)
}
