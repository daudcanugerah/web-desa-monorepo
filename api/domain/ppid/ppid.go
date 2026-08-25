package ppid

import (
	"fmt"
	"strings"
	"time"
)

// PPID represents a public information disclosure document entity in the domain layer.
// PPID (Pejabat Pengelola Informasi dan Dokumentasi) documents provide government
// transparency information that citizens can access and request.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/ppid for the Repository interface definition.
//
// Files are stored either via legacy *_url string paths or via the gallery
// *_media_id FK. Legacy rows keep their URL for one release; new rows persist
// only the media_id and the *_url columns remain NULL.
type PPID struct {
	ID                string     `db:"id"`
	Title             string     `db:"title"`
	Category          *string    `db:"category"`           // Optional document category (e.g., "budget", "regulation", "report")
	CategoryName      *string    `db:"category_name"`
	DocumentMediaID   *string    `db:"document_media_id"`  // Gallery media UUID for the document
	ThumbnailMediaID  *string    `db:"thumbnail_media_id"` // Gallery media UUID for the thumbnail
	Description       *string    `db:"description"`        // Optional description of the PPID document
	PublicationAt     *time.Time `db:"publication_at"`     // Optional publication date for the document
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

// PPIDRequest represents a public request for a PPID document in the domain layer.
// Citizens can submit requests to access specific PPID documents, and the system
// records their information for transparency and accountability.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/ppid for the Repository interface definition.
type PPIDRequest struct {
	ID             string     `db:"id"`
	PPIDId         string     `db:"ppid_id"` // Foreign key reference to PPID document
	RequesterName  string     `db:"requester_name"`
	RequesterEmail string     `db:"requester_email"`
	Purpose        *string    `db:"notes"`       // Optional purpose/notes for the request
	Status         string     `db:"status"`      // pending, approved, revoked
	ApprovedAt     *time.Time `db:"approved_at"` // Timestamp when request was approved
	ApprovedBy     *string    `db:"approved_by"` // User ID who approved the request
	RevokedAt      *time.Time `db:"revoked_at"`  // Timestamp when request was revoked
	RevokedBy      *string    `db:"revoked_by"`  // User ID who revoked the request
	CreatedAt      time.Time  `db:"created_at"`
}

// Validate checks if the PPID entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
//
// A PPID document must have a DocumentMediaID (gallery). The thumbnail
// is optional and is now also expressed exclusively through its
// gallery media id.
func (p *PPID) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("ppid ID is required")
	}

	if err := validateTitle(p.Title); err != nil {
		return err
	}

	if p.Category != nil {
		if err := validateCategory(*p.Category); err != nil {
			return err
		}
	}

	if p.DocumentMediaID == nil || *p.DocumentMediaID == "" {
		return fmt.Errorf("ppid document media id is required")
	}

	if p.Description != nil {
		if err := validateDescription(*p.Description); err != nil {
			return err
		}
	}

	if p.PublicationAt != nil {
		if err := validatePublicationAt(*p.PublicationAt); err != nil {
			return err
		}
	}

	if p.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}

	if p.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at timestamp is required")
	}

	return nil
}

// Validate checks if the PPIDRequest entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (r *PPIDRequest) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("ppid request ID is required")
	}

	if r.PPIDId == "" {
		return fmt.Errorf("ppid_id is required")
	}

	if err := validateRequesterName(r.RequesterName); err != nil {
		return err
	}

	if err := validateRequesterEmail(r.RequesterEmail); err != nil {
		return err
	}

	if r.Purpose != nil {
		if err := validatePurpose(*r.Purpose); err != nil {
			return err
		}
	}

	if err := validateRequestStatus(r.Status); err != nil {
		return err
	}

	// If approved, both ApprovedAt and ApprovedBy must be set
	if r.Status == "approved" {
		if r.ApprovedAt == nil || r.ApprovedBy == nil {
			return fmt.Errorf("approved request must have approved_at and approved_by set")
		}
	}

	// If revoked, both RevokedAt and RevokedBy must be set
	if r.Status == "revoked" {
		if r.RevokedAt == nil || r.RevokedBy == nil {
			return fmt.Errorf("revoked request must have revoked_at and revoked_by set")
		}
	}

	if r.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
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

// validateCategory checks if the category meets domain requirements
func validateCategory(category string) error {
	category = strings.TrimSpace(category)
	if category == "" {
		return fmt.Errorf("category cannot be empty string")
	}

	if len(category) > 100 {
		return fmt.Errorf("category must not exceed 100 characters")
	}

	return nil
}

// validateDocumentURL checks if the document URL meets domain requirements
func validateDocumentURL(url string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("document URL cannot be empty string")
	}

	if len(url) > 500 {
		return fmt.Errorf("document URL must not exceed 500 characters")
	}

	return nil
}

// validateThumbnailURL checks if the thumbnail URL meets domain requirements
func validateThumbnailURL(url string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("thumbnail URL cannot be empty string")
	}

	if len(url) > 500 {
		return fmt.Errorf("thumbnail URL must not exceed 500 characters")
	}

	// Validate that thumbnail is an image file
	validImageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	lowerURL := strings.ToLower(url)
	hasValidExtension := false
	for ext := range validImageExtensions {
		if strings.HasSuffix(lowerURL, ext) {
			hasValidExtension = true
			break
		}
	}

	if !hasValidExtension {
		return fmt.Errorf("thumbnail must be an image file (jpg, jpeg, png, gif, or webp)")
	}

	return nil
}

// validateDescription checks if the description meets domain requirements
func validateDescription(description string) error {
	description = strings.TrimSpace(description)
	if description == "" {
		return fmt.Errorf("description cannot be empty string")
	}

	// Description is TEXT type in database, no strict length limit
	return nil
}

// validateRequesterName checks if the requester name meets domain requirements
func validateRequesterName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("requester name is required")
	}

	if len(name) > 255 {
		return fmt.Errorf("requester name must not exceed 255 characters")
	}

	return nil
}

// validateRequesterEmail checks if the requester email meets domain requirements
func validateRequesterEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("requester email is required")
	}

	if len(email) > 255 {
		return fmt.Errorf("requester email must not exceed 255 characters")
	}

	// Basic email format validation (domain layer keeps it simple)
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("requester email must be a valid format")
	}

	return nil
}

// validatePurpose checks if the purpose meets domain requirements
func validatePurpose(purpose string) error {
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		return fmt.Errorf("purpose cannot be empty string")
	}

	// Purpose is TEXT type in database, no strict length limit
	return nil
}

// validatePublicationAt checks if the publication_at timestamp meets domain requirements
func validatePublicationAt(publicationAt time.Time) error {
	if publicationAt.IsZero() {
		return fmt.Errorf("publication_at cannot be zero value")
	}

	return nil
}

// validateRequestStatus checks if the status meets domain requirements
func validateRequestStatus(status string) error {
	status = strings.TrimSpace(status)
	if status == "" {
		return fmt.Errorf("status is required")
	}

	validStatuses := map[string]bool{
		"pending":  true,
		"approved": true,
		"revoked":  true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("status must be one of: pending, approved, revoked")
	}

	return nil
}

// IsApproved returns true if the request status is approved
func (r *PPIDRequest) IsApproved() bool {
	return r.Status == "approved"
}

// IsRevoked returns true if the request status is revoked
func (r *PPIDRequest) IsRevoked() bool {
	return r.Status == "revoked"
}

// IsPending returns true if the request status is pending
func (r *PPIDRequest) IsPending() bool {
	return r.Status == "pending"
}
