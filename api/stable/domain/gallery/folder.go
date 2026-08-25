// Package gallery defines the Folder and Media domain entities.
package gallery

import (
	"fmt"
	"strings"
	"time"
)

type Folder struct {
	ID                string    `db:"id"`
	Name              string    `db:"name"`
	Description       string    `db:"description"`
	IsPublic          bool      `db:"is_public"`
	CoverMediaID      *string   `db:"cover_media_id"`
	CoverThumbnailURL *string   `db:"-"`
	MediaCount        int       `db:"media_count"`
	CreatedBy         string    `db:"created_by"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

const (
	folderNameMaxLen        = 255
	folderDescriptionMaxLen = 1000
)

func (f *Folder) Validate() error {
	if f.ID == "" {
		return fmt.Errorf("folder ID is required")
	}
	if err := validateFolderName(f.Name); err != nil {
		return err
	}
	if err := validateFolderDescription(f.Description); err != nil {
		return err
	}
	if f.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}
	if f.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}
	if f.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at timestamp is required")
	}
	return nil
}

func validateFolderName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("folder name is required")
	}
	if len(name) > folderNameMaxLen {
		return fmt.Errorf("folder name must not exceed %d characters", folderNameMaxLen)
	}
	return nil
}

func validateFolderDescription(description string) error {
	if len(description) > folderDescriptionMaxLen {
		return fmt.Errorf("folder description must not exceed %d characters", folderDescriptionMaxLen)
	}
	return nil
}
