package jsonutils

import "encoding/json"

// Optional distinguishes an absent JSON field (no change on update) from a
// field explicitly set to null (clear the value). Set reports whether the
// field was present and Value the decoded value; Value is nil for null.
type Optional[T any] struct {
	Set   bool
	Value *T
}

// UnmarshalJSON runs only for fields present in the request body, so an
// absent field keeps its zero value.
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
