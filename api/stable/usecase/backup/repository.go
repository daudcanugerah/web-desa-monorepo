package backup

import (
	"context"

	"webdesa/api/domain/backup"
)

// Repository defines the persistence interface for backup metadata management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/backup_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
//
// Note: This repository only handles backup metadata (records in the database).
// The actual backup file operations (pg_dump, pg_restore) are handled by the backup service.
type Repository interface {
	// Create creates a new backup metadata record in the database
	Create(ctx context.Context, b *backup.Backup) error

	// FindByID retrieves a backup metadata record by ID
	FindByID(ctx context.Context, id string) (*backup.Backup, error)

	// List retrieves paginated backup metadata records
	// Returns backup slice, total count, and error
	List(ctx context.Context, offset, limit int) ([]*backup.Backup, int, error)
}
