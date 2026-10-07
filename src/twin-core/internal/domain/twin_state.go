package domain

import (
	"context"
	"errors"

	"IDIG4110/shared/dto"
)

// DefaultHomeID is the sentinel home that placeholder devices are registered
// under when a reading arrives for an entity never seen before. The fixed id
// lets every worker resolve to the same row; gateway sync moves devices to
// their real homes later.
const DefaultHomeID = "00000000-0000-0000-0000-000000000001"

// ErrNotFound signals a read for an id that does not exist; handlers map it
// to 404.
var ErrNotFound = errors.New("resource not found")

// ErrInvalidReading wraps a contract violation in a received reading;
// handlers map it to 400.
var ErrInvalidReading = errors.New("invalid reading")

// ErrBadRequest wraps a request validation failure; handlers map it to 400
// with the reason.
var ErrBadRequest = errors.New("invalid request")

// ErrConflict signals a uniqueness violation, e.g. a device that already
// exists on the gateway; handlers map it to 409.
var ErrConflict = errors.New("conflict")

// TwinStateRepo is the write path: apply one normalized reading to
// twin_state.
type TwinStateRepo interface {
	ApplyReading(ctx context.Context, reading dto.NormalizedReading) error
}

// TwinStateSvc validates normalized readings and applies them.
type TwinStateSvc interface {
	ApplyReading(ctx context.Context, reading dto.NormalizedReading) error
}

// StateQueryRepo is the read path behind the frontend API.
type StateQueryRepo interface {
	ListHomes(ctx context.Context) ([]HomeSummary, error)
	GetHome(ctx context.Context, homeID string) (HomeSummary, error)
	ListAreasByHome(ctx context.Context, homeID string) ([]Area, error)
	ListDevicesByHome(ctx context.Context, homeID string) ([]Device, error)
	ListEntityStatesByHome(ctx context.Context, homeID string) ([]EntityState, error)
	ListEntityStates(ctx context.Context, filter EntityFilter) (EntityStatePage, error)
	GetEntityState(ctx context.Context, entityID string) (EntityState, error)
}

// StateQuerySvc serves the current state of the twin to the frontend.
type StateQuerySvc interface {
	ListHomes(ctx context.Context) ([]HomeSummary, error)
	GetHomeState(ctx context.Context, homeID string) (HomeState, error)
	ListEntityStates(ctx context.Context, filter EntityFilter) (EntityStatePage, error)
	GetEntityState(ctx context.Context, entityID string) (EntityState, error)
}

// StructureRepo is the write path for the structural model: homes, areas,
// devices, entities, and the relations that connect them.
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

// StructureSvc manages the structural model; the frontend reaches it
// through the gateway API.
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
