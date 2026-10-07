package server

import (
	"net/http"

	"IDIG4110/twin-core/internal/domain"
)

// handleDevices routes everything under /api/v1/devices: POST registers a
// device, PATCH {id} updates one, DELETE {id} removes one.
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

// createDevice registers a device.
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

// updateDevice modifies one device.
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

// deleteDevice removes one device.
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
