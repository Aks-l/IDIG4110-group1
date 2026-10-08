package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// Routes everything under /api/v1/devices
// POST creates, PATCH {id} updates, DELETE {id} removes
//
// # Inputs:
//
//   - svc [domain.StructureSvc] structure service
//
// # Returns:
//
//   - Collection handler
func handleDevices(svc domain.StructureSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest := belowCollection(r)

		switch {
		case len(rest) == 0:
			if r.Method != http.MethodPost {
				respondMethodNotAllowed(w, r, http.MethodPost)
				return
			}
			createDevice(w, r, svc)
		case len(rest) == 1:
			switch r.Method {
			case http.MethodPatch:
				updateDevice(w, r, svc, rest[0])
			case http.MethodDelete:
				deleteDevice(w, r, svc, rest[0])
			default:
				respondMethodNotAllowed(w, r, http.MethodPatch, http.MethodDelete)
			}
		default:
			respondNotFound(w)
		}
	}
}

// Registers a device
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
func createDevice(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc) {
	req, ok := decodeBody[domain.CreateDeviceRequest](w, r)
	if !ok {
		return
	}
	device, err := svc.CreateDevice(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondCreated(w, device)
}

// Updates one device
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
//   - id [string] device id path segment
func updateDevice(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	deviceID, ok := validID(w, id, "device")
	if !ok {
		return
	}
	req, ok := decodeBody[domain.UpdateDeviceRequest](w, r)
	if !ok {
		return
	}
	device, err := svc.UpdateDevice(r.Context(), deviceID, req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, device)
}

// Removes one device
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - svc [domain.StructureSvc] structure service
//   - id [string] device id path segment
func deleteDevice(w http.ResponseWriter, r *http.Request, svc domain.StructureSvc, id string) {
	deviceID, ok := validID(w, id, "device")
	if !ok {
		return
	}
	if err := svc.DeleteDevice(r.Context(), deviceID); err != nil {
		respondError(w, err)
		return
	}
	respondNoContent(w)
}
