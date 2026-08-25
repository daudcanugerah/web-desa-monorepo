package desa

import (
	"context"

	"webdesa/api/domain/desa"
)

// Repository defines the persistence interface for village profile (desa) management.
// Defined here (where it's used) following Go best practices.
// Implementation lives in interface/repository/desa_postgres.go.
type Repository interface {
	// Get retrieves the village profile stored as JSON in the settings table.
	// Returns a zero-value Desa (empty profile) if the key does not exist yet.
	Get(ctx context.Context) (*desa.Desa, error)

	// Upsert inserts or updates the village profile JSON in the settings table.
	Upsert(ctx context.Context, d *desa.Desa) error
}
