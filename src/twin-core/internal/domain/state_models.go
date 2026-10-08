package domain

import "time"

// Entity joined with its latest known state
// state nil when unread, previous nil until second reading
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

// Latest known value of an entity, exactly one of value_num/value_text set
type StateValue struct {
	ValueNum   *float64       `json:"value_num"`
	ValueText  *string        `json:"value_text"`
	Attributes map[string]any `json:"attributes"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// Homes list item
type HomeSummary struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Address  *string `json:"address,omitempty"`
	Timezone string  `json:"timezone"`
}

// Room or zone within a home
type Area struct {
	ID       string  `json:"id"`
	HomeID   string  `json:"home_id"`
	Name     string  `json:"name"`
	Floor    *int    `json:"floor"`
	AreaType *string `json:"area_type,omitempty"`
	Geometry any     `json:"geometry,omitempty"`
}

// Normalized device model
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

// Area with its registered entities
type AreaState struct {
	Area
	Entities []EntityState `json:"entities"`
}

// Device with its entities
type DeviceState struct {
	Device
	Entities []EntityState `json:"entities"`
}

// Dashboard view of one home, areas and devices with entity states
type HomeState struct {
	Home    HomeSummary   `json:"home"`
	Areas   []AreaState   `json:"areas"`
	Devices []DeviceState `json:"devices"`
}

// One page of the filterable entity list, total counts all matches
type EntityStatePage struct {
	Items  []EntityState `json:"items"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// One edge of a home's graph between two nodes (area, device or entity)
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
