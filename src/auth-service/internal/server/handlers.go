package server

import "IDIG4110/auth-service/internal/user"

type Handlers struct {
	User *user.UserHandler
}

func NewHandlers(
	userHandler *user.UserHandler,
) *Handlers {
	return &Handlers{
		User: userHandler,
	}
}
