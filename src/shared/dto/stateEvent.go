package dto

import "time"

// StateEvent is the device-native event carried in a RawMessage payload.
type StateEvent struct {
	EventType string    `json:"event_type"`
	TimeFired time.Time `json:"time_fired"`
	EntityID  string    `json:"entity_id"`
	NewState  NewState  `json:"new_state"`
}

// NewState carries the state value and its attributes. Attributes are
// preserved verbatim, so values can be any JSON type, not only strings.
type NewState struct {
	State      string         `json:"state"`
	Attributes map[string]any `json:"attributes"`
}
