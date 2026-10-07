package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// handleAreas routes everything under /api/v1/areas: POST registers an
// area, PATCH {id} updates one, DELETE {id} removes one.
func handleAreas(svc domain.StructureSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest := belowCollection(r)

		switch {
		case len(rest) == 0:
			if r.Method != http.MethodPost {
				respondMethodNotAllowed(w, r, http.MethodPost)
				return
			}
			createArea(w, r, svc)
		case len(rest) == 1:
			switch r.Method {
			case http.MethodPatch:
				updateArea(w, r, svc, rest[0])
			case http.MethodDelete:
				deleteArea(w, r, svc, rest[0])
			default:
				respondMethodNotAllowed(w, r, http.MethodPatch, http.MethodDelete)
			}
		default:
			respondNotFound(w)
		}
	}
}

// createArea registers an area.
func createArea(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc) {
	req, ok := decodeBody[domain.CreateAreaRequest](w, r)
	if !ok {
		return
	}
	area, err := svc.CreateArea(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondCreated(w, area)
}

// updateArea modifies one area.
func updateArea(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	areaID, ok := validID(w, id, "area")
	if !ok {
		return
	}
	req, ok := decodeBody[domain.UpdateAreaRequest](w, r)
	if !ok {
		return
	}
	area, err := svc.UpdateArea(r.Context(), areaID, req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, area)
}

// deleteArea removes one area.
func deleteArea(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	areaID, ok := validID(w, id, "area")
	if !ok {
		return
	}
	if err := svc.DeleteArea(r.Context(), areaID); err != nil {
		respondError(w, err)
		return
	}
	respondNoContent(w)
}
