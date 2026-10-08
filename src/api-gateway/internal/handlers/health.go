package handlers

import (
	"net/http"

	"IDIG4110/shared/httperror"
	jsonutils "IDIG4110/shared/json-utils"
)

func GetHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := jsonutils.Encode(w, http.StatusOK, map[string]string{"status": "ok"}); err != nil {
			httperror.HandleError(w, http.StatusInternalServerError, err, httperror.ErrInternalServerError)
		}
	}
}
