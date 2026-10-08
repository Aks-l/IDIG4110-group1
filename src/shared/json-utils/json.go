package jsonutils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Writes v as JSON response with status code
//
// # Inputs:
//
//   - w [http.ResponseWriter] response writer
//   - code [int] http status code
//   - v [T] value to encode
//
// # Returns:
//
//   - Encoding error
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

// Reads request body as JSON into T
//
// # Inputs:
//
//   - r [*http.Request] request carrying the body
//
// # Returns:
//
//   - Decoded value
//   - Decode error
func Decode[T any](r *http.Request) (T, error) {
	var data T
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("error: decoding json: %w", err)
	}

	return data, nil
}
