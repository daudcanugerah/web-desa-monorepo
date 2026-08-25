package berita

import (
	"fmt"
	"strings"
	"time"
)

// Berita represents a news article entity in the domain layer.
// News articles are published to inform the public about village events and updates.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/berita for the Repository interface definition.
type Berita struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	Category  string    `db:"category"`  // FK to berita_categories.id (UUID)
	CategoryName *string `db:"category_name"`
	ImageURL  *string   `db:"image_url"` // Optional image for the news article
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Validate checks if the Berita entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (b *Berita) Validate() error {
	if b.ID == "" {
		return fmt.Errorf("berita ID is required")
	}

	if err := validateTitle(b.Title); err != nil {
		return err
	}

	if err := validateContent(b.Content); err != nil {
		return err
	}

	if strings.TrimSpace(b.Category) == "" {
		return fmt.Errorf("category is required")
	}

	if b.ImageURL != nil {
		if err := validateImageURL(*b.ImageURL); err != nil {
			return err
		}
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

// validateContent checks if the content meets domain requirements
func validateContent(content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("content is required")
	}

	// Content is TEXT type in database, no strict length limit but should not be empty
	return nil
}

// validateImageURL checks if the image URL meets domain requirements
func validateImageURL(imageURL string) error {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return fmt.Errorf("image URL cannot be empty string")
	}

	if len(imageURL) > 500 {
		return fmt.Errorf("image URL must not exceed 500 characters")
	}

	return nil
}
