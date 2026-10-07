package domain

import "time"

// EntityState is one entity joined with its latest known state. State is
// nil when the entity has not been read yet; Previous is nil until a second
// accepted reading arrives.
type EntityState struct {
	EntityID         string      `json:"entity_id"`
	ExternalEntityID string      `json:"external_entity_id"`
	Name             string      `json:"name"`
	Domain           string      `json:"domain"`
	Controllable     bool        `json:"controllable"`
	DeviceClass      string      `json:"device_class,omitempty"`
	Unit             string      `json:"unit,omitempty"`
	AreaID           *string     `json:"area_id"`
	DeviceID         string      `json:"device_id"`
	HomeID           string      `json:"home_id"`
	GatewayID        string      `json:"gateway_id"`
	State            *StateValue `json:"state"`
	Previous         *StateValue `json:"previous"`
}

// StateValue is the latest known value of an entity. Exactly one of
// ValueNum and ValueText is set, mirroring the readings contract.
type StateValue struct {
	ValueNum   *float64       `json:"value_num"`
	ValueText  *string        `json:"value_text"`
	Attributes map[string]any `json:"attributes"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// HomeSummary is the homes list item.
type HomeSummary struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Address  *string `json:"address,omitempty"`
	Timezone string  `json:"timezone"`
}

// Area is a room or zone within a home. Floor, area type, and geometry are
// all optional.
type Area struct {
	ID       string  `json:"id"`
	HomeID   string  `json:"home_id"`
	Name     string  `json:"name"`
	Floor    *int    `json:"floor"`
	AreaType *string `json:"area_type,omitempty"`
	Geometry any     `json:"geometry,omitempty"`
}

// Device is the normalized device model. A device is not required to be in
// an area, and the descriptive fields are optional.
type Device struct {
	ID           string  `json:"id"`
	HomeID       string  `json:"home_id"`
	AreaID       *string `json:"area_id"`
	GatewayID    string  `json:"gateway_id"`
	ExternalID   string  `json:"external_id"`
	Name         string  `json:"name"`
	Manufacturer *string `json:"manufacturer,omitempty"`
	Model        *string `json:"model,omitempty"`
	SwVersion    *string `json:"sw_version,omitempty"`
	DeviceType   *string `json:"device_type,omitempty"`
}

// AreaState is an area with the entities registered to it.
type AreaState struct {
	Area
	Entities []EntityState `json:"entities"`
}

// DeviceState is a device with its entities.
type DeviceState struct {
	Device
	Entities []EntityState `json:"entities"`
}

// HomeState is the dashboard view of one home: its areas and devices, each
// carrying their entities with current state.
type HomeState struct {
	Home    HomeSummary   `json:"home"`
	Areas   []AreaState   `json:"areas"`
	Devices []DeviceState `json:"devices"`
}

// EntityStatePage is one page of the filterable entity list. Total counts
// every matching entity before paging so the frontend can build pagers.
type EntityStatePage struct {
	Items  []EntityState `json:"items"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// Relation is one edge of a home's structural graph: how two nodes (an
// area, a device, or an entity) are connected. Edges are stored
// directional, but bidirectional ones are semantically undirected.
type Relation struct {
	ID            string     `json:"id"`
	HomeID        string     `json:"home_id"`
	FromKind      string     `json:"from_kind"`
	FromID        string     `json:"from_id"`
	ToKind        string     `json:"to_kind"`
	ToID          string     `json:"to_id"`
	RelationType  string     `json:"relation_type"`
	Bidirectional bool       `json:"bidirectional"`
	Label         *string    `json:"label,omitempty"`
	Properties    any        `json:"properties,omitempty"`
	ValidFrom     *time.Time `json:"valid_from,omitempty"`
	ValidTo       *time.Time `json:"valid_to,omitempty"`
}
