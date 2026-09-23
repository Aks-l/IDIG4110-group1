package jsonutils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Encode[T any](w http.ResponseWriter, code int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	if code != http.StatusOK {
		w.WriteHeader(code)
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("error: encoding json: %w", err)
	}

	return nil
}

func Decode[T any](r *http.Request) error {
	var data T
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return fmt.Errorf("error: decoding json: %w", err)
	}

	return nil
}
