package banner

import (
	"fmt"
	"strings"
	"time"
)

// Banner represents a promotional banner entity in the domain layer.
// Banners are displayed on the public interface to show announcements.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/banner for the Repository interface definition.
type Banner struct {
	ID            string                 `db:"id"`
	Title         string                 `db:"title"`
	Description   string                 `db:"description"`
	ImageMediaID  *string                `db:"image_media_id"`
	Link          string                 `db:"link"`
	Status        string                 `db:"status"`
	Category      *string                `db:"category"`
	CategoryName  *string                `db:"category_name"`
	Metadata      map[string]interface{} `db:"metadata"`
	CreatedAt     time.Time              `db:"created_at"`
	UpdatedAt     time.Time              `db:"updated_at"`
}

// Status constants for banner states
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// Validate checks if the Banner entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
//
// At least one of ImageURL (legacy string URL) or ImageMediaID (gallery
// media UUID) must be set on a banner. New rows persist only the
// media_id; legacy rows keep the image_url so existing front-ends
// continue to render.
func (b *Banner) Validate() error {
	if b.ID == "" {
		return fmt.Errorf("banner ID is required")
	}

	if err := validateTitle(b.Title); err != nil {
		return err
	}

	if b.ImageMediaID == nil || *b.ImageMediaID == "" {
		return fmt.Errorf("banner image or media id is required")
	}

	if err := validateStatus(b.Status); err != nil {
		return err
	}

	if b.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}

	if b.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at timestamp is required")
	}

	return nil
}

// validateTitle checks if the title meets domain requirements
func validateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("title is required")
	}

	if len(title) > 255 {
		return fmt.Errorf("title must not exceed 255 characters")
	}

	return nil
}

// validateImageURL checks if the image URL meets domain requirements
func validateImageURL(imageURL string) error {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return fmt.Errorf("image URL is required")
	}

	if len(imageURL) > 500 {
		return fmt.Errorf("image URL must not exceed 500 characters")
	}

	return nil
}

// validateStatus checks if the status is valid
func validateStatus(status string) error {
	if status != StatusActive && status != StatusInactive {
		return fmt.Errorf("status must be either 'active' or 'inactive'")
	}

	return nil
}
