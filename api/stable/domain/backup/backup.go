package backup

import (
	"fmt"
	"strings"
	"time"
)

// Backup represents a database backup entity in the domain layer.
// This entity contains metadata about database backup files including
// filename, file size, and creation timestamp.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/backup for the Repository interface definition.
type Backup struct {
	ID        string    `db:"id"`         // UUID identifier for the backup record
	Filename  string    `db:"filename"`   // Name of the backup file (e.g., "backup_20240115_143022.sql")
	Size      int64     `db:"size"`       // Size of the backup file in bytes
	CreatedAt time.Time `db:"created_at"` // Timestamp when the backup was created
}

// Validate checks if the Backup entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (b *Backup) Validate() error {
	if b.ID == "" {
		return fmt.Errorf("backup ID is required")
	}

	if err := validateFilename(b.Filename); err != nil {
		return err
	}

	if err := validateSize(b.Size); err != nil {
		return err
	}

	if b.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}

	return nil
}

// validateFilename checks if the filename meets domain requirements
func validateFilename(filename string) error {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	if len(filename) > 255 {
		return fmt.Errorf("filename must not exceed 255 characters")
	}

	return nil
}

// validateSize checks if the size meets domain requirements
func validateSize(size int64) error {
	if size < 0 {
		return fmt.Errorf("size must be non-negative")
	}

	if size == 0 {
		return fmt.Errorf("size must be greater than zero")
	}

	return nil
}
