package server

import "net/http"

// Serves liveness probe
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
func handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, map[string]string{"status": "ok"})
}
