package banner

import (
	"context"
	"fmt"
	"io"

	"webdesa/api/domain/banner"
	"webdesa/api/domain/bannercategory"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// CategoryLookup is the subset of the banner-category service the
// banner.Service needs. Defined as an interface here (consumer side) so the
// usecase layer doesn't import the category implementation.
//
// Implemented by *bannercategory.Service.
type CategoryLookup interface {
	FindByID(ctx context.Context, id string) (*bannercategory.Category, error)
}

// Service implements banner management business logic.
// It accepts interfaces (Repository, FileHandler) and returns concrete structs (Banner).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo        Repository
	fileHandler FileHandler
	categories  CategoryLookup
	clock       clock.Clock
}

// NewService creates a new banner management service.
// Dependencies are injected via constructor following Clean Architecture principles.
//
// categories may be nil to skip FK validation (legacy mode).
func NewService(repo Repository, fileHandler FileHandler, categories CategoryLookup, clk clock.Clock) *Service {
	return &Service{
		repo:        repo,
		fileHandler: fileHandler,
		categories:  categories,
		clock:       clk,
	}
}

// CreateBannerInput represents the input for creating a banner
type CreateBannerInput struct {
	Title       string
	Description string
	Link        string
	ImageFile   io.Reader
	ImageName   string
	ImageSize   int64
	ContentType string
	Category    *string // Optional UUID FK to banner_categories
	Metadata    map[string]interface{} // Optional metadata (e.g., HTML content)
}

// UpdateBannerInput represents the input for updating a banner
type UpdateBannerInput struct {
	Title       string
	Description string
	Link        string
	ImageFile   io.Reader // Optional - only if updating image
	ImageName   string
	ImageSize   int64
	ContentType string
	Category    *string // Optional UUID FK to banner_categories (nil = leave unchanged)
	Metadata    map[string]interface{} // Optional metadata (e.g., HTML content)
}

// ListBannersInput represents the input for listing banners
type ListBannersInput struct {
	Status   *string // Optional filter: "active" or "inactive"
	Query    *string // Optional search across title and description
	Category *string // Optional UUID FK to banner_categories
	Page     int
	Limit    int
}

// Create creates a new banner with an image file.
// Returns concrete Banner struct.
//
// Validates: Requirements 7.3
func (s *Service) Create(ctx context.Context, input CreateBannerInput) (*banner.Banner, error) {
	// Save image file
	imageURL, err := s.fileHandler.SaveImage(ctx, input.ImageName, input.ImageFile, input.ImageSize, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	// Validate category FK before doing other work — DB-level FK violation
	// would surface as a 23503 error after image upload, so we pre-check.
	if input.Category != nil && s.categories != nil {
		if _, err := s.categories.FindByID(ctx, *input.Category); err != nil {
			_ = s.fileHandler.Delete(ctx, imageURL)
			return nil, fmt.Errorf("invalid category: %w", err)
		}
	}

	// Create banner entity
	now := s.clock.Now()
	// Initialize metadata as empty map if not provided
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	b := &banner.Banner{
		ID:          uuid.New().String(),
		Title:       input.Title,
		Description: input.Description,
		Link:        input.Link,
		ImageURL:    imageURL,
		Status:      banner.StatusInactive, // Default to inactive
		Category:    input.Category,
		Metadata:    metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Validate domain invariants
	if err := b.Validate(); err != nil {
		// Clean up uploaded file on validation failure
		_ = s.fileHandler.Delete(ctx, imageURL)
		return nil, fmt.Errorf("invalid banner data: %w", err)
	}

	// Persist banner
	if err := s.repo.Create(ctx, b); err != nil {
		// Clean up uploaded file on persistence failure
		_ = s.fileHandler.Delete(ctx, imageURL)
		return nil, fmt.Errorf("failed to create banner: %w", err)
	}

	return b, nil
}

// GetByID retrieves a banner by ID.
// Returns concrete Banner struct.
//
// Validates: Requirements 7.2
func (s *Service) GetByID(ctx context.Context, id string) (*banner.Banner, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("banner not found: %w", err)
	}

	return b, nil
}

// List retrieves paginated banners with optional status filtering.
// Returns concrete Banner structs and pagination metadata.
//
// Validates: Requirements 7.1, 7.7, 17.2, 17.3
func (s *Service) List(ctx context.Context, input ListBannersInput) ([]*banner.Banner, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Retrieve banners from repository
	banners, total, err := s.repo.List(ctx, input.Status, input.Query, input.Category, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list banners: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return banners, paginationResult, nil
}

// Update updates an existing banner.
// If ImageFile is provided, updates the image; otherwise keeps existing image.
// Returns concrete Banner struct.
//
// Validates: Requirements 7.4
func (s *Service) Update(ctx context.Context, id string, input UpdateBannerInput) (*banner.Banner, error) {
	// Retrieve existing banner
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("banner not found: %w", err)
	}

	oldImageURL := b.ImageURL

	// Update image if provided
	if input.ImageFile != nil {
		newImageURL, err := s.fileHandler.SaveImage(ctx, input.ImageName, input.ImageFile, input.ImageSize, input.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to save new image: %w", err)
		}
		b.ImageURL = newImageURL
	}

	// Validate category FK if explicitly provided. nil means "leave unchanged";
	// non-nil means the caller explicitly set the category and we must verify.
	if input.Category != nil && s.categories != nil {
		if _, err := s.categories.FindByID(ctx, *input.Category); err != nil {
			if input.ImageFile != nil {
				_ = s.fileHandler.Delete(ctx, b.ImageURL)
			}
			return nil, fmt.Errorf("invalid category: %w", err)
		}
		b.Category = input.Category
	}

	// Update fields
	b.Title = input.Title
	b.Description = input.Description
	b.Link = input.Link
	if input.Metadata != nil {
		b.Metadata = input.Metadata
	}
	// Note: Status is NOT updated here - use UpdateStatus for that
	b.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := b.Validate(); err != nil {
		// Clean up new image if validation fails
		if input.ImageFile != nil {
			_ = s.fileHandler.Delete(ctx, b.ImageURL)
		}
		return nil, fmt.Errorf("invalid banner data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, b); err != nil {
		// Clean up new image if persistence fails
		if input.ImageFile != nil {
			_ = s.fileHandler.Delete(ctx, b.ImageURL)
		}
		return nil, fmt.Errorf("failed to update banner: %w", err)
	}

	// Delete old image if a new one was uploaded successfully
	if input.ImageFile != nil && oldImageURL != "" {
		_ = s.fileHandler.Delete(ctx, oldImageURL)
	}

	return b, nil
}

// UpdateStatus updates a banner's status.
// The active-banner count limit (max 20) is enforced atomically by the
// database trigger `enforce_active_banner_limit` (migration 00038), which
// eliminates the TOCTOU race between SELECT count and UPDATE that existed
// when the check was performed in application code.
// Returns concrete Banner struct.
//
// Validates: Requirements 7.5
func (s *Service) UpdateStatus(ctx context.Context, id string, status string) (*banner.Banner, error) {
	// Retrieve existing banner
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("banner not found: %w", err)
	}

	// Update status — DB trigger enforces the 20-active-banner limit atomically.
	b.Status = status
	b.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := b.Validate(); err != nil {
		return nil, fmt.Errorf("invalid banner data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("failed to update banner status: %w", err)
	}

	return b, nil
}

// Delete removes a banner and its associated image file.
//
// Validates: Requirements 7.6
func (s *Service) Delete(ctx context.Context, id string) error {
	// Retrieve banner to get image URL
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("banner not found: %w", err)
	}

	// Delete banner from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete banner: %w", err)
	}

	// Delete associated image file
	if b.ImageURL != "" {
		if err := s.fileHandler.Delete(ctx, b.ImageURL); err != nil {
			// Log error but don't fail the operation since banner is already deleted
			// In production, this should be logged properly
			_ = err
		}
	}

	return nil
}
