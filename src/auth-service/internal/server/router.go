package server

import "net/http"

func NewRouter(handlers *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc(AuthPath+"/register", handlers.User.Register)

	return mux
}
