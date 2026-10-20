package server

import "net/http"

func NewRouter(handlers *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc(AuthRegisterPath, handlers.Auth.Register)
	mux.HandleFunc(AuthLoginPath, handlers.Auth.Login)
	mux.HandleFunc(AuthRefreshPath, handlers.Auth.Refresh)
	mux.HandleFunc(AuthLogoutPath, handlers.Auth.Logout)
	return mux
}
