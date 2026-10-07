package server

import "net/http"

// handleHealth serves the liveness probe.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, map[string]string{"status": "ok"})
}
