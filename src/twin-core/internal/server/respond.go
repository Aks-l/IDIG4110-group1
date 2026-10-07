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

// decodeBody decodes and validates the request body into T, writing a 400
// on a malformed body or a tag violation.
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

// validID checks that an id path segment is a uuid, writing a 400 naming
// the resource kind when it is not.
func validID(w http.ResponseWriter, id, kind string) (string, bool) {
	if !dto.IsValidUUID(id) {
		respondBadRequest(w, fmt.Errorf("invalid %s id: %q", kind, id))
		return "", false
	}
	return id, true
}

// respondJSON writes v as a 200 JSON body.
func respondJSON(w http.ResponseWriter, v any) {
	if err := jsonutils.Encode(w, http.StatusOK, v); err != nil {
		httperror.HandleError(w, http.StatusInternalServerError, err, httperror.ErrInternalServerError)
	}
}

// respondCreated writes v as a 201 JSON body.
func respondCreated(w http.ResponseWriter, v any) {
	if err := jsonutils.Encode(w, http.StatusCreated, v); err != nil {
		httperror.HandleError(w, http.StatusInternalServerError, err, httperror.ErrInternalServerError)
	}
}

// respondNoContent writes the empty 204 used by the delete endpoints.
func respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// respondBadRequest writes a 400 with err's reason as the message.
func respondBadRequest(w http.ResponseWriter, err error) {
	httperror.HandleError(w, http.StatusBadRequest, err, err.Error())
}

// respondMethodNotAllowed writes a 405 listing the methods the path
// supports, in the Allow header and the error body.
func respondMethodNotAllowed(w http.ResponseWriter, r *http.Request, allowed ...string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	httperror.HandleError(w, http.StatusMethodNotAllowed, fmt.Errorf("%s on %s", r.Method, r.URL.Path), httperror.ErrMethodNotAllowed)
}

// errUnknownPath marks a 404 caused by an undefined path rather than an
// unknown id.
var errUnknownPath = errors.New("unknown path")

// respondNotFound writes a 404 for a path the API does not define.
func respondNotFound(w http.ResponseWriter) {
	httperror.HandleError(w, http.StatusNotFound, errUnknownPath, httperror.ErrNotFound)
}

// respondError maps a service error onto its HTTP status using the domain
// sentinels.
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
