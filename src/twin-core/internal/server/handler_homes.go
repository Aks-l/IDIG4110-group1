package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// handleHomes routes everything under /api/v1/homes:
//
//   - the collection: POST registers a home, GET lists them
//   - one home by id: PATCH updates it, DELETE removes it
//   - sub-resources: GET {id}/state is the dashboard view, GET
//     {id}/relations is the home's relation graph
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

// createHome registers a home.
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

// listHomes serves the homes list.
func listHomes(w http.ResponseWriter, r *http.Request, svc domain.StateQuerySvc) {
	homes, err := svc.ListHomes(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, homes)
}

// updateHome modifies one home.
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

// deleteHome removes one home.
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

// getHomeState serves one home's dashboard view: areas and devices with
// their entities and current state.
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

// listHomeRelations serves the edges of one home's graph view.
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
