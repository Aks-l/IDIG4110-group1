package domain

import (
	"context"
	"errors"

	"IDIG4110/shared/dto"
)

// Sentinel home for auto-provisioned placeholder devices
// fixed id keeps every worker on the same row
const DefaultHomeID = "00000000-0000-0000-0000-000000000001"

// Unknown id, handlers map to 404
var ErrNotFound = errors.New("resource not found")

// Reading contract violation, handlers map to 400
var ErrInvalidReading = errors.New("invalid reading")

// Request validation failure, handlers map to 400
var ErrBadRequest = errors.New("invalid request")

// Uniqueness violation, handlers map to 409
var ErrConflict = errors.New("conflict")

// Write path, applies one normalized reading to twin_state
type TwinStateRepo interface {
	ApplyReading(ctx context.Context, reading dto.Reading) error
}

// Validates readings and applies them
type TwinStateSvc interface {
	ApplyReading(ctx context.Context, reading dto.Reading) error
}

// Read path behind the frontend API
type StateQueryRepo interface {
	ListHomes(ctx context.Context) ([]HomeSummary, error)
	GetHome(ctx context.Context, homeID string) (HomeSummary, error)
	ListAreasByHome(ctx context.Context, homeID string) ([]Area, error)
	ListDevicesByHome(ctx context.Context, homeID string) ([]Device, error)
	ListEntityStatesByHome(ctx context.Context, homeID string) ([]EntityState, error)
	ListEntityStates(ctx context.Context, filter EntityFilter) (EntityStatePage, error)
	GetEntityState(ctx context.Context, entityID string) (EntityState, error)
}

// Serves current twin state to the frontend
type StateQuerySvc interface {
	ListHomes(ctx context.Context) ([]HomeSummary, error)
	GetHomeState(ctx context.Context, homeID string) (HomeState, error)
	ListEntityStates(ctx context.Context, filter EntityFilter) (EntityStatePage, error)
	GetEntityState(ctx context.Context, entityID string) (EntityState, error)
}

// Write path for the structural model
type StructureRepo interface {
	CreateHome(ctx context.Context, req CreateHomeRequest) (HomeSummary, error)
	UpdateHome(ctx context.Context, homeID string, req UpdateHomeRequest) (HomeSummary, error)
	DeleteHome(ctx context.Context, homeID string) error
	CreateArea(ctx context.Context, req CreateAreaRequest) (Area, error)
	UpdateArea(ctx context.Context, areaID string, req UpdateAreaRequest) (Area, error)
	DeleteArea(ctx context.Context, areaID string) error
	CreateDevice(ctx context.Context, req CreateDeviceRequest) (Device, error)
	UpdateDevice(ctx context.Context, deviceID string, req UpdateDeviceRequest) (Device, error)
	DeleteDevice(ctx context.Context, deviceID string) error
	CreateEntity(ctx context.Context, req CreateEntityRequest) (EntityState, error)
	UpdateEntity(ctx context.Context, entityID string, req UpdateEntityRequest) (EntityState, error)
	DeleteEntity(ctx context.Context, entityID string) error
	ResolveNodeHome(ctx context.Context, kind, nodeID string) (string, error)
	CreateRelation(ctx context.Context, homeID string, req CreateRelationRequest) (Relation, error)
	UpdateRelation(ctx context.Context, relationID string, req UpdateRelationRequest) (Relation, error)
	DeleteRelation(ctx context.Context, relationID string) error
	ListRelationsByHome(ctx context.Context, homeID string) ([]Relation, error)
}

// Manages the structural model for the frontend
type StructureSvc interface {
	CreateHome(ctx context.Context, req CreateHomeRequest) (HomeSummary, error)
	UpdateHome(ctx context.Context, homeID string, req UpdateHomeRequest) (HomeSummary, error)
	DeleteHome(ctx context.Context, homeID string) error
	CreateArea(ctx context.Context, req CreateAreaRequest) (Area, error)
	UpdateArea(ctx context.Context, areaID string, req UpdateAreaRequest) (Area, error)
	DeleteArea(ctx context.Context, areaID string) error
	CreateDevice(ctx context.Context, req CreateDeviceRequest) (Device, error)
	UpdateDevice(ctx context.Context, deviceID string, req UpdateDeviceRequest) (Device, error)
	DeleteDevice(ctx context.Context, deviceID string) error
	CreateEntity(ctx context.Context, req CreateEntityRequest) (EntityState, error)
	UpdateEntity(ctx context.Context, entityID string, req UpdateEntityRequest) (EntityState, error)
	DeleteEntity(ctx context.Context, entityID string) error
	CreateRelation(ctx context.Context, req CreateRelationRequest) (Relation, error)
	UpdateRelation(ctx context.Context, relationID string, req UpdateRelationRequest) (Relation, error)
	DeleteRelation(ctx context.Context, relationID string) error
	ListRelationsByHome(ctx context.Context, homeID string) ([]Relation, error)
}
