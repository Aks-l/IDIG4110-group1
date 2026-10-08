package repository

import (
	"IDIG4110/twin-core/internal/db"
)

// Persists readings and serves home structure and twin state queries,
// split by collection: registry, readings, homes, areas, devices, entities.
// Row level security isolates homes at the database layer, WithHome
// (internal/db) opens every home transaction and sets app.home_id, the
// policies on the home tables hide rows from every other home.
type TwinStateRepoImpl struct {
	db *db.Database
}

func NewTwinStateRepoImpl(db *db.Database) *TwinStateRepoImpl {
	return &TwinStateRepoImpl{
		db: db,
	}
}
