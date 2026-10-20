package auth

import (
	"IDIG4110/shared/dto"
	"IDIG4110/shared/httperror"
	jsonutils "IDIG4110/shared/json-utils"
	"net/http"
)

type AuthHandler struct {
	userService *AuthService
}

func NewAuthHandler(uS *AuthService) *AuthHandler {
	return &AuthHandler{
		userService: uS,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	user, err := jsonutils.Decode[dto.UserDTO](r)
	if err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, httperror.ErrUnprocessableEntity)
		return
	}

	if err := jsonutils.Validate[dto.UserDTO](user); err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, httperror.ErrUnprocessableEntity)
		return
	}

	if err := h.userService.RegisterNewAccount(r.Context(), user); err != nil {
		httperror.HandleAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	user, err := jsonutils.Decode[dto.UserDTO](r)
	if err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, httperror.ErrUnprocessableEntity)
		return
	}

	if err := jsonutils.Validate[dto.UserDTO](user); err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, httperror.ErrUnprocessableEntity)
		return
	}

	token, err := h.userService.Login(r.Context(), user)
	if err != nil {
		httperror.HandleAppError(w, err)
		return
	}

	jsonutils.Encode[dto.TokenResponseDTO](w, http.StatusOK, *token)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refresh, err := jsonutils.Decode[dto.RefreshTokenDTO](r)
	if err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, httperror.ErrUnprocessableEntity)
		return
	}

	if err := jsonutils.Validate[dto.RefreshTokenDTO](refresh); err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, httperror.ErrUnprocessableEntity)
		return
	}

	token, err := h.userService.RefreshToken(r.Context(), refresh)
	if err != nil {
		httperror.HandleAppError(w, err)
		return
	}

	jsonutils.Encode[dto.TokenResponseDTO](w, http.StatusOK, *token)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refresh, err := jsonutils.Decode[dto.RefreshTokenDTO](r)
	if err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, httperror.ErrUnprocessableEntity)
		return
	}

	if err := jsonutils.Validate[dto.RefreshTokenDTO](refresh); err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, httperror.ErrUnprocessableEntity)
		return
	}

	if err := h.userService.Logout(r.Context(), refresh); err != nil {
		httperror.HandleAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
