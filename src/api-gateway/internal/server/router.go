package server

import (
	"net/http"

	"IDIG4110/api-gateway/internal/handlers"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+HEALTH, handlers.GetHealth())

	return mux
}
