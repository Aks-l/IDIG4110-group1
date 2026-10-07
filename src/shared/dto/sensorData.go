package dto

import "time"

// SensorStateEvent mirrors the Home Assistant state_changed event a hub
// such as HA Green emits when an entity updates. It is the raw gateway
// payload; ingest translates it into NormalizedReading
// (docs/architecture/mqtt-envelope.md).

type SensorStateEvent struct {
	EventType string       `json:"event_type"` // "state_changed" for these events
	TimeFired time.Time    `json:"time_fired"` // when the hub fired the event
	Data      EventData    `json:"data"`
	Context   EventContext `json:"context"` // event-level context; ID is the dedup key
}

// EventData carries the event's subject and its new state object.
type EventData struct {
	EntityID string      `json:"entity_id"` // the authoritative id of the entity that changed
	NewState StateObject `json:"new_state"`
}

// StateObject is HA's snapshot of one entity. HA always carries the state
// as a string, and attributes can hold any JSON value.
type StateObject struct {
	EntityID    string         `json:"entity_id"` // snapshot field HA repeats inside the state
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes"`
	LastChanged time.Time      `json:"last_changed"`
	LastUpdated time.Time      `json:"last_updated"`
	Context     EventContext   `json:"context"`
}

// EventContext links an event to the command or automation behind it.
type EventContext struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id"`
	UserID   *string `json:"user_id"`
}
