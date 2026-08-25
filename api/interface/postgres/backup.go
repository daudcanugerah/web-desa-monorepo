package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/backup"
	backupUsecase "webdesa/api/usecase/backup"
)

// BackupRepository implements usecase/backup.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/backup where it's USED, not here where it's implemented.
//
// Note: This repository only handles backup metadata (records in the database).
// The actual backup file operations (pg_dump, pg_restore) are handled by the backup service.
type BackupRepository struct {
	db *sqlx.DB
}

// NewBackupRepository creates a new BackupRepository instance
func NewBackupRepository(db *sqlx.DB) backupUsecase.Repository {
	return &BackupRepository{db: db}
}

// Create creates a new backup metadata record in the database
// Generates UUID, sets timestamp, and inserts the backup record
// Validates: Requirements 14.3, 20.2, 20.3
func (r *BackupRepository) Create(ctx context.Context, b *backup.Backup) error {
	// Generate UUID if not provided
	if b.ID == "" {
		b.ID = uuid.New().String()
	}

	query := `
		INSERT INTO backups (id, filename, size, kind, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`

	if b.Kind == "" {
		b.Kind = "db"
	}

	_, err := r.db.ExecContext(ctx, query,
		b.ID,
		b.Filename,
		b.Size,
		b.Kind,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create backup record: %w", err))
	}

	return nil
}

// FindByID retrieves a backup metadata record by ID
// Returns error if backup not found
// Validates: Requirements 14.4
func (r *BackupRepository) FindByID(ctx context.Context, id string) (*backup.Backup, error) {
	var b backup.Backup

	query := `
		SELECT id, filename, size, kind, created_at
		FROM backups
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &b, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("backup not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find backup by ID: %w", err))
	}

	return &b, nil
}

// List retrieves paginated backup metadata records
// Returns backup slice, total count, and error
// Uses LIMIT/OFFSET for pagination
// Validates: Requirements 14.4, 14.5, 17.2, 20.5
func (r *BackupRepository) List(ctx context.Context, offset, limit int) ([]*backup.Backup, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM backups`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count backups: %w", err))
	}

	// Get paginated backups
	var backups []*backup.Backup
	query := `
		SELECT id, filename, size, kind, created_at
		FROM backups
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	err = r.db.SelectContext(ctx, &backups, query, limit, offset)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list backups: %w", err))
	}

	return backups, total, nil
}
