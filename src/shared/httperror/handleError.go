package httperror

import (
	jsonutils "IDIG4110/shared/json-utils"
	"log/slog"
	"net/http"
)

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func HandleError(w http.ResponseWriter, code int, err error, msg string) {
	if err != nil {
		slog.Warn("HandleError called with no error", "message", msg)
		return
	}

	slog.Error("request error","code", code, "err", err, "message", msg)

	resp := ErrorMessage {
		Code: code,
		Message: msg,
	}

	if errEncode := jsonutils.Encode(w, code, resp); err != nil {
		slog.Error("failed to write error response", "error", errEncode)
	}
}
