package middleware

import (
	"log/slog"
	"net/http"

	"IDIG4110/shared/httperror"
)

func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					slog.Error("panic recovered", "panic", rec)
					httperror.HandleError(w, http.StatusInternalServerError, nil, httperror.ErrInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
