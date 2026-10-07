package user

import (
	"IDIG4110/shared/dto"
	"IDIG4110/shared/httperror"
	jsonutils "IDIG4110/shared/json-utils"
	"net/http"
)

type UserHandler struct {
	userService *UserService
}

func NewUserHandler(uS *UserService) *UserHandler {
	return &UserHandler{
		userService: uS,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user dto.CreateUserDTO
	if err := jsonutils.Decode[dto.CreateUserDTO](r, &user); err != nil {
		httperror.HandleError(w, http.StatusBadRequest, err, httperror.ErrInternalServerError)
		return
	}

	if err := h.userService.RegisterNewAccount(r.Context(), user); err != nil {
		return
	}

	w.WriteHeader(http.StatusCreated)
}
