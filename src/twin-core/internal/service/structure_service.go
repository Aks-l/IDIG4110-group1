package service

import (
	"context"
	"fmt"
	"strings"

	jsonutils "IDIG4110/shared/json-utils"
	"IDIG4110/twin-core/internal/domain"
)

// Manages the structural model behind the structure API
// applies defaults and cross-field rules validate tags cannot express
type StructureSvcImpl struct {
	repo domain.StructureRepo
}

func NewImplStructureSvc(repo domain.StructureRepo) *StructureSvcImpl {
	return &StructureSvcImpl{
		repo: repo,
	}
}

func (s *StructureSvcImpl) CreateHome(ctx context.Context, req domain.CreateHomeRequest) (domain.HomeSummary, error) {
	if req.Timezone == nil {
		utc := "UTC"
		req.Timezone = &utc
	}
	return s.repo.CreateHome(ctx, req)
}

func (s *StructureSvcImpl) UpdateHome(ctx context.Context, homeID string, req domain.UpdateHomeRequest) (domain.HomeSummary, error) {
	if err := requireNotNull("name", req.Name); err != nil {
		return domain.HomeSummary{}, err
	}
	return s.repo.UpdateHome(ctx, homeID, req)
}

func (s *StructureSvcImpl) CreateArea(ctx context.Context, req domain.CreateAreaRequest) (domain.Area, error) {
	return s.repo.CreateArea(ctx, req)
}

func (s *StructureSvcImpl) UpdateArea(ctx context.Context, areaID string, req domain.UpdateAreaRequest) (domain.Area, error) {
	if err := requireNotNull("name", req.Name); err != nil {
		return domain.Area{}, err
	}
	return s.repo.UpdateArea(ctx, areaID, req)
}

func (s *StructureSvcImpl) CreateDevice(ctx context.Context, req domain.CreateDeviceRequest) (domain.Device, error) {
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		name := req.ExternalID
		req.Name = &name
	}
	return s.repo.CreateDevice(ctx, req)
}

func (s *StructureSvcImpl) UpdateDevice(ctx context.Context, deviceID string, req domain.UpdateDeviceRequest) (domain.Device, error) {
	if err := requireNotNull("name", req.Name); err != nil {
		return domain.Device{}, err
	}
	if req.HomeID.Set {
		// Row level security pins every device to one home
		return domain.Device{}, fmt.Errorf("%w: moving a device to another home is not supported", domain.ErrBadRequest)
	}
	return s.repo.UpdateDevice(ctx, deviceID, req)
}

func (s *StructureSvcImpl) CreateEntity(ctx context.Context, req domain.CreateEntityRequest) (domain.EntityState, error) {
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		name := req.ExternalEntityID
		req.Name = &name
	}
	if req.Domain == nil || strings.TrimSpace(*req.Domain) == "" {
		derived := domain.EntityDomain(req.ExternalEntityID)
		req.Domain = &derived
	}
	if req.Controllable == nil {
		controllable := false
		req.Controllable = &controllable
	}
	return s.repo.CreateEntity(ctx, req)
}

func (s *StructureSvcImpl) UpdateEntity(ctx context.Context, entityID string, req domain.UpdateEntityRequest) (domain.EntityState, error) {
	if err := requireNotNull("name", req.Name); err != nil {
		return domain.EntityState{}, err
	}
	return s.repo.UpdateEntity(ctx, entityID, req)
}

// Rejects update clearing a required field with null
//
// # Inputs:
//
//   - field [string] field name for the error
//   - v [jsonutils.Optional[T]] the field value
//
// # Returns:
//
//   - ErrBadRequest when field is set to null
func requireNotNull[T any](field string, v jsonutils.Optional[T]) error {
	if v.Set && v.Value == nil {
		return fmt.Errorf("%w: %s cannot be null", domain.ErrBadRequest, field)
	}
	return nil
}

// Deletes home, refuses the auto-provisioning home
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//
// # Returns:
//
//   - ErrConflict when home is the default home
func (s *StructureSvcImpl) DeleteHome(ctx context.Context, homeID string) error {
	if homeID == domain.DefaultHomeID {
		return fmt.Errorf("%w: cannot delete the auto-provisioning home", domain.ErrConflict)
	}
	return s.repo.DeleteHome(ctx, homeID)
}

func (s *StructureSvcImpl) DeleteArea(ctx context.Context, areaID string) error {
	return s.repo.DeleteArea(ctx, areaID)
}

func (s *StructureSvcImpl) DeleteDevice(ctx context.Context, deviceID string) error {
	return s.repo.DeleteDevice(ctx, deviceID)
}

func (s *StructureSvcImpl) DeleteEntity(ctx context.Context, entityID string) error {
	return s.repo.DeleteEntity(ctx, entityID)
}

// Registers one edge of a home's graph
// enforces no self-edges, endpoints in same home, matching home_id
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - req [domain.CreateRelationRequest] relation payload
//
// # Returns:
//
//   - Created relation
//   - Error on rule violation or repo failure
func (s *StructureSvcImpl) CreateRelation(ctx context.Context, req domain.CreateRelationRequest) (domain.Relation, error) {
	if req.FromKind == req.ToKind && req.FromID == req.ToID {
		return domain.Relation{}, fmt.Errorf("%w: a relation cannot connect a node to itself", domain.ErrBadRequest)
	}

	fromHome, err := s.repo.ResolveNodeHome(ctx, req.FromKind, req.FromID)
	if err != nil {
		return domain.Relation{}, err
	}
	toHome, err := s.repo.ResolveNodeHome(ctx, req.ToKind, req.ToID)
	if err != nil {
		return domain.Relation{}, err
	}
	if fromHome != toHome {
		return domain.Relation{}, fmt.Errorf("%w: relation endpoints must belong to the same home", domain.ErrBadRequest)
	}
	if req.HomeID != nil && *req.HomeID != fromHome {
		return domain.Relation{}, fmt.Errorf("%w: home_id does not match the endpoints' home", domain.ErrBadRequest)
	}

	if req.Bidirectional == nil {
		bidirectional := true
		req.Bidirectional = &bidirectional
	}
	return s.repo.CreateRelation(ctx, fromHome, req)
}

func (s *StructureSvcImpl) UpdateRelation(ctx context.Context, relationID string, req domain.UpdateRelationRequest) (domain.Relation, error) {
	if err := requireNotNull("relation_type", req.RelationType); err != nil {
		return domain.Relation{}, err
	}
	if err := requireNotNull("bidirectional", req.Bidirectional); err != nil {
		return domain.Relation{}, err
	}
	return s.repo.UpdateRelation(ctx, relationID, req)
}

func (s *StructureSvcImpl) DeleteRelation(ctx context.Context, relationID string) error {
	return s.repo.DeleteRelation(ctx, relationID)
}

func (s *StructureSvcImpl) ListRelationsByHome(ctx context.Context, homeID string) ([]domain.Relation, error) {
	return s.repo.ListRelationsByHome(ctx, homeID)
}
