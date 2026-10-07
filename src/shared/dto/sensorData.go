package dto

import "time"

// Home Assistant state_changed event, raw gateway payload
// ingest translates it into NormalizedReading
type SensorStateEvent struct {
	EventType string       `json:"event_type"` // "state_changed" for these events
	TimeFired time.Time    `json:"time_fired"` // when the hub fired the event
	Data      EventData    `json:"data"`
	Context   EventContext `json:"context"` // event-level context; ID is the dedup key
}

// Carries event subject and its new state
type EventData struct {
	EntityID string      `json:"entity_id"` // the authoritative id of the entity that changed
	NewState StateObject `json:"new_state"`
}

// HA snapshot of one entity, state always a string
type StateObject struct {
	EntityID    string         `json:"entity_id"` // snapshot field HA repeats inside the state
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes"`
	LastChanged time.Time      `json:"last_changed"`
	LastUpdated time.Time      `json:"last_updated"`
	Context     EventContext   `json:"context"`
}

// Links event to the command or automation behind it
type EventContext struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id"`
	UserID   *string `json:"user_id"`
}
