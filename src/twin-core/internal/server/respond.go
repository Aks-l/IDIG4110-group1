package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"IDIG4110/shared/dto"
	"IDIG4110/shared/httperror"
	jsonutils "IDIG4110/shared/json-utils"
	"IDIG4110/shared/validate"
	"IDIG4110/twin-core/internal/domain"
)

// Decodes and validates request body into T, writes 400 on failure
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//
// # Returns:
//
//   - Decoded value
//   - False when response already written
func decodeBody[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	req, err := jsonutils.Decode[T](r)
	if err != nil {
		httperror.HandleError(w, http.StatusBadRequest, err, "invalid json body")
		return req, false
	}
	if err := validate.Struct(&req); err != nil {
		respondBadRequest(w, err)
		return req, false
	}
	return req, true
}

// Checks id path segment is a uuid, writes 400 when not
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - id [string] id path segment
//   - kind [string] resource kind for the error
//
// # Returns:
//
//   - The id
//   - False when response already written
func validID(w http.ResponseWriter, id, kind string) (string, bool) {
	if !dto.IsValidUUID(id) {
		respondBadRequest(w, fmt.Errorf("invalid %s id: %q", kind, id))
		return "", false
	}
	return id, true
}

// Writes v as 200 JSON body
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - v [any] value to encode
func respondJSON(w http.ResponseWriter, v any) {
	if err := jsonutils.Encode(w, http.StatusOK, v); err != nil {
		httperror.HandleError(w, http.StatusInternalServerError, err, httperror.ErrInternalServerError)
	}
}

// Writes v as 201 JSON body
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - v [any] value to encode
func respondCreated(w http.ResponseWriter, v any) {
	if err := jsonutils.Encode(w, http.StatusCreated, v); err != nil {
		httperror.HandleError(w, http.StatusInternalServerError, err, httperror.ErrInternalServerError)
	}
}

// Writes empty 204 for delete endpoints
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
func respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Writes 400 with err reason as message
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - err [error] error carrying the reason
func respondBadRequest(w http.ResponseWriter, err error) {
	httperror.HandleError(w, http.StatusBadRequest, err, err.Error())
}

// Writes 405 with allowed methods in Allow header
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - r [*http.Request] incoming request
//   - allowed [...string] allowed methods
func respondMethodNotAllowed(w http.ResponseWriter, r *http.Request, allowed ...string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	httperror.HandleError(w, http.StatusMethodNotAllowed, fmt.Errorf("%s on %s", r.Method, r.URL.Path), httperror.ErrMethodNotAllowed)
}

// Marks 404 from undefined path, not unknown id
var errUnknownPath = errors.New("unknown path")

// Writes 404 for undefined path
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
func respondNotFound(w http.ResponseWriter) {
	httperror.HandleError(w, http.StatusNotFound, errUnknownPath, httperror.ErrNotFound)
}

// Maps service error onto http status via domain sentinels
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - err [error] service error
func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		httperror.HandleError(w, http.StatusNotFound, err, httperror.ErrNotFound)
	case errors.Is(err, domain.ErrConflict):
		httperror.HandleError(w, http.StatusConflict, err, httperror.ErrConflict)
	case errors.Is(err, domain.ErrBadRequest), errors.Is(err, domain.ErrInvalidReading):
		httperror.HandleError(w, http.StatusBadRequest, err, err.Error())
	default:
		httperror.HandleError(w, http.StatusInternalServerError, err, httperror.ErrInternalServerError)
	}
}
