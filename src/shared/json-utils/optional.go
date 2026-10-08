package jsonutils

import "encoding/json"

// Optional PATCH field, absent means no change, null clears the value
type Optional[T any] struct {
	Set   bool
	Value *T
}

// Decodes present fields only, absent fields keep zero value
//
// # Inputs:
//
//   - data []byte raw json value of the field
//
// # Returns:
//
//   - Decode error
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}

	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}
