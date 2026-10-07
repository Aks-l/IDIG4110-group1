package domain

import (
	"time"

	jsonutils "IDIG4110/shared/json-utils"
)

// Create and update payloads for the structure API
// field rules run as validate tags, Optional marks PATCH semantics

type CreateHomeRequest struct {
	Name     string  `json:"name" validate:"nonblank"`
	Address  *string `json:"address"`
	Timezone *string `json:"timezone"`
}

type UpdateHomeRequest struct {
	Name     jsonutils.Optional[string] `json:"name" validate:"omitempty,nonblank"`
	Address  jsonutils.Optional[string] `json:"address"`
	Timezone jsonutils.Optional[string] `json:"timezone"`
}

type CreateAreaRequest struct {
	HomeID   string  `json:"home_id" validate:"required,uuid"`
	Name     string  `json:"name" validate:"nonblank"`
	Floor    *int    `json:"floor"`
	AreaType *string `json:"area_type"`
	Geometry any     `json:"geometry"`
}

type UpdateAreaRequest struct {
	Name     jsonutils.Optional[string] `json:"name" validate:"omitempty,nonblank"`
	Floor    jsonutils.Optional[int]    `json:"floor"`
	AreaType jsonutils.Optional[string] `json:"area_type"`
	Geometry jsonutils.Optional[any]    `json:"geometry"`
}

type CreateDeviceRequest struct {
	HomeID       string  `json:"home_id" validate:"required,uuid"`
	GatewayID    string  `json:"gateway_id" validate:"required,uuid"`
	ExternalID   string  `json:"external_id" validate:"nonblank"`
	Name         *string `json:"name"`
	AreaID       *string `json:"area_id" validate:"omitempty,uuid"`
	Manufacturer *string `json:"manufacturer"`
	Model        *string `json:"model"`
	SwVersion    *string `json:"sw_version"`
	DeviceType   *string `json:"device_type"`
}

type UpdateDeviceRequest struct {
	Name         jsonutils.Optional[string] `json:"name" validate:"omitempty,nonblank"`
	HomeID       jsonutils.Optional[string] `json:"home_id" validate:"omitempty,uuid"`
	AreaID       jsonutils.Optional[string] `json:"area_id" validate:"omitempty,uuid"`
	Manufacturer jsonutils.Optional[string] `json:"manufacturer"`
	Model        jsonutils.Optional[string] `json:"model"`
	SwVersion    jsonutils.Optional[string] `json:"sw_version"`
	DeviceType   jsonutils.Optional[string] `json:"device_type"`
}

type CreateEntityRequest struct {
	DeviceID         string         `json:"device_id" validate:"required,uuid"`
	ExternalEntityID string         `json:"external_entity_id" validate:"nonblank,max=255"`
	Name             *string        `json:"name"`
	AreaID           *string        `json:"area_id" validate:"omitempty,uuid"`
	Domain           *string        `json:"domain"`
	DeviceClass      *string        `json:"device_class"`
	Unit             *string        `json:"unit"`
	Controllable     *bool          `json:"controllable"`
	CommandMap       map[string]any `json:"command_map"`
	StateTTLSeconds  *int           `json:"state_ttl_seconds"`
}

type UpdateEntityRequest struct {
	Name            jsonutils.Optional[string]         `json:"name" validate:"omitempty,nonblank"`
	AreaID          jsonutils.Optional[string]         `json:"area_id" validate:"omitempty,uuid"`
	DeviceClass     jsonutils.Optional[string]         `json:"device_class"`
	Unit            jsonutils.Optional[string]         `json:"unit"`
	Controllable    jsonutils.Optional[bool]           `json:"controllable"`
	CommandMap      jsonutils.Optional[map[string]any] `json:"command_map"`
	StateTTLSeconds jsonutils.Optional[int]            `json:"state_ttl_seconds"`
}

// Query params for the /state list endpoint, filters combine with AND
type EntityFilter struct {
	HomeID        string
	AreaID        string
	DeviceID      string
	Domains       []string
	DeviceClasses []string
	DeviceTypes   []string
	Controllable  *bool
	Floor         *int
	Limit         int
	Offset        int
}

// Relation payloads, endpoints are (kind, id) pairs

type CreateRelationRequest struct {
	HomeID        *string        `json:"home_id" validate:"omitempty,uuid"`
	FromKind      string         `json:"from_kind" validate:"required,oneof=area device entity"`
	FromID        string         `json:"from_id" validate:"required,uuid"`
	ToKind        string         `json:"to_kind" validate:"required,oneof=area device entity"`
	ToID          string         `json:"to_id" validate:"required,uuid"`
	RelationType  string         `json:"relation_type" validate:"required,oneof=connects_to contains monitors controls same_physical_device depends_on"`
	Bidirectional *bool          `json:"bidirectional"`
	Label         *string        `json:"label"`
	Properties    map[string]any `json:"properties"`
	ValidFrom     *time.Time     `json:"valid_from"`
	ValidTo       *time.Time     `json:"valid_to"`
}

type UpdateRelationRequest struct {
	RelationType  jsonutils.Optional[string]         `json:"relation_type" validate:"omitempty,oneof=connects_to contains monitors controls same_physical_device depends_on"`
	Bidirectional jsonutils.Optional[bool]           `json:"bidirectional"`
	Label         jsonutils.Optional[string]         `json:"label"`
	Properties    jsonutils.Optional[map[string]any] `json:"properties"`
	ValidFrom     jsonutils.Optional[time.Time]      `json:"valid_from"`
	ValidTo       jsonutils.Optional[time.Time]      `json:"valid_to"`
}
