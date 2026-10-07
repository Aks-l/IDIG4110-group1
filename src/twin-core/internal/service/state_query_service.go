package service

import (
	"context"

	"IDIG4110/twin-core/internal/domain"
)

type StateQuerySvcImpl struct {
	repo domain.StateQueryRepo
}

func NewImplStateQuerySvc(repo domain.StateQueryRepo) *StateQuerySvcImpl {
	return &StateQuerySvcImpl{
		repo: repo,
	}
}

func (s *StateQuerySvcImpl) ListHomes(ctx context.Context) ([]domain.HomeSummary, error) {
	return s.repo.ListHomes(ctx)
}

// Assembles dashboard view of one home
// entity with area appears under both the area and its device
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//
// # Returns:
//
//   - Home state
//   - ErrNotFound when home missing
func (s *StateQuerySvcImpl) GetHomeState(ctx context.Context, homeID string) (domain.HomeState, error) {
	home, err := s.repo.GetHome(ctx, homeID)
	if err != nil {
		return domain.HomeState{}, err
	}

	areas, err := s.repo.ListAreasByHome(ctx, homeID)
	if err != nil {
		return domain.HomeState{}, err
	}

	devices, err := s.repo.ListDevicesByHome(ctx, homeID)
	if err != nil {
		return domain.HomeState{}, err
	}

	entities, err := s.repo.ListEntityStatesByHome(ctx, homeID)
	if err != nil {
		return domain.HomeState{}, err
	}

	areaStates := make([]domain.AreaState, len(areas))
	areasByID := make(map[string]int, len(areas))
	for i, a := range areas {
		areaStates[i] = domain.AreaState{Area: a, Entities: []domain.EntityState{}}
		areasByID[a.ID] = i
	}

	deviceStates := make([]domain.DeviceState, len(devices))
	devicesByID := make(map[string]int, len(devices))
	for i, d := range devices {
		deviceStates[i] = domain.DeviceState{Device: d, Entities: []domain.EntityState{}}
		devicesByID[d.ID] = i
	}

	for _, e := range entities {
		if e.AreaID != nil {
			if i, ok := areasByID[*e.AreaID]; ok {
				areaStates[i].Entities = append(areaStates[i].Entities, e)
			}
		}
		if i, ok := devicesByID[e.DeviceID]; ok {
			deviceStates[i].Entities = append(deviceStates[i].Entities, e)
		}
	}

	return domain.HomeState{
		Home:    home,
		Areas:   areaStates,
		Devices: deviceStates,
	}, nil
}

func (s *StateQuerySvcImpl) ListEntityStates(ctx context.Context, filter domain.EntityFilter) (domain.EntityStatePage, error) {
	return s.repo.ListEntityStates(ctx, filter)
}

func (s *StateQuerySvcImpl) GetEntityState(ctx context.Context, entityID string) (domain.EntityState, error) {
	return s.repo.GetEntityState(ctx, entityID)
}
