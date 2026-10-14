package server

import "net/http"

func NewRouter(handlers *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc(http.MethodPost+AuthPath+"/register", handlers.Auth.Register)
	mux.HandleFunc(http.MethodPost+AuthPath+"/login", handlers.Auth.Login)
	mux.HandleFunc(http.MethodPost+AuthPath+"/refresh", handlers.Auth.Refresh)
	mux.HandleFunc(http.MethodPost+AuthPath+"/logout", handlers.Auth.Logout)
	return mux
}
