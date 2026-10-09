package server

import "IDIG4110/auth-service/internal/auth"

type Handlers struct {
	Auth *auth.AuthHandler
}

func NewHandlers(
	authHandler *auth.AuthHandler,
) *Handlers {
	return &Handlers{
		Auth: authHandler,
	}
}
