package httperror

import (
	jsonutils "IDIG4110/shared/json-utils"
	"log/slog"
	"net/http"
)

// Shared error response body {code, message}
type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Logs error and writes JSON error response
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - code [int] http status code
//   - err [error] error to log, returns early when nil
//   - msg [string] message written to response
func HandleError(w http.ResponseWriter, code int, err error, msg string) {
	if err == nil {
		slog.Warn("HandleError called with no error", "message", msg)
		return
	}

	slog.Error("request error", "code", code, "err", err, "message", msg)

	resp := ErrorMessage{
		Code: code,
		Message: msg,
	}

	if errEncode := jsonutils.Encode(w, code, resp); errEncode != nil {
		slog.Error("failed to write error response", "error", errEncode)
	}
}
