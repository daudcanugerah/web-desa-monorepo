package banner

import (
	"context"
	"errors"
	"fmt"
	"io"

	"webdesa/api/domain/banner"
	"webdesa/api/domain/bannercategory"
	galleryUsecase "webdesa/api/usecase/gallery"
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
// It accepts interfaces (Repository, FileStore) and returns concrete structs (Banner).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo       Repository
	fileStore  galleryUsecase.FileStore
	categories CategoryLookup
	clock      clock.Clock
}

// NewService creates a new banner management service.
// Dependencies are injected via constructor following Clean Architecture principles.
//
// categories may be nil to skip FK validation (legacy mode).
func NewService(repo Repository, fileStore galleryUsecase.FileStore, categories CategoryLookup, clk clock.Clock) *Service {
	return &Service{
		repo:       repo,
		fileStore:  fileStore,
		categories: categories,
		clock:      clk,
	}
}

const featureSlug = galleryUsecase.FeatureBanner

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
	// ImageMediaID references an image already uploaded via
	// POST /banners/upload-media. Mutually exclusive with ImageFile.
	ImageMediaID *string
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
	// ImageMediaID references an image already uploaded via
	// POST /banners/upload-media. Mutually exclusive with ImageFile.
	ImageMediaID *string
}

// ListBannersInput represents the input for listing banners
type ListBannersInput struct {
	Status   *string // Optional filter: "active" or "inactive"
	Query    *string // Optional search across title and description
	Category *string // Optional UUID FK to banner_categories
	Page     int
	Limit    int
}

// resolveImage returns the media id for a pre-uploaded media ref or a newly
// saved inline file. The second return tells the caller whether the media
// was created here (only that case may delete it on failure).
func (s *Service) resolveImage(ctx context.Context, mediaRef *string, file io.Reader, saveFile func() (galleryUsecase.SavedFile, error)) (string, bool, error) {
	if mediaRef != nil && *mediaRef != "" {
		if file != nil {
			return "", false, errors.New("provide either a media id or a file, not both")
		}
		if _, err := s.fileStore.Open(ctx, *mediaRef); err != nil {
			return "", false, fmt.Errorf("invalid media id: %w", err)
		}
		return *mediaRef, false, nil
	}
	if file != nil {
		saved, err := saveFile()
		if err != nil {
			return "", false, err
		}
		return saved.MediaID, true, nil
	}
	return "", false, nil
}

// Create creates a new banner with an image file.
// Returns concrete Banner struct.
//
// Validates: Requirements 7.3
func (s *Service) Create(ctx context.Context, input CreateBannerInput) (*banner.Banner, error) {
	// Resolve the image: pre-uploaded media ref or inline file. Only media
	// created here may be deleted on failure.
	mediaID, created, err := s.resolveImage(ctx, input.ImageMediaID, input.ImageFile, func() (galleryUsecase.SavedFile, error) {
		return s.fileStore.SaveImage(ctx, featureSlug, galleryUsecase.FileInput{
			OriginalName: input.ImageName,
			Content:      input.ImageFile,
			Size:         input.ImageSize,
			ContentType:  input.ContentType,
		})
	})
	if err != nil {
		return nil, err
	}
	if mediaID == "" {
		return nil, fmt.Errorf("banner image or media id is required")
	}
	cleanup := func() {
		if created {
			_ = s.fileStore.Delete(ctx, mediaID)
		}
	}

	if input.Category != nil && s.categories != nil {
		if _, err := s.categories.FindByID(ctx, *input.Category); err != nil {
			cleanup()
			return nil, fmt.Errorf("invalid category: %w", err)
		}
	}

	now := s.clock.Now()
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	b := &banner.Banner{
		ID:           uuid.NewString(),
		Title:        input.Title,
		Description:  input.Description,
		Link:         input.Link,
		ImageMediaID: &mediaID,
		Status:       banner.StatusInactive,
		Category:     input.Category,
		Metadata:     metadata,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := b.Validate(); err != nil {
		cleanup()
		return nil, fmt.Errorf("invalid banner data: %w", err)
	}

	if err := s.repo.Create(ctx, b); err != nil {
		cleanup()
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
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	banners, total, err := s.repo.List(ctx, input.Status, input.Query, input.Category, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list banners: %w", err)
	}

	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return banners, paginationResult, nil
}

// Update updates an existing banner.
// If ImageFile is provided, updates the image; otherwise keeps existing image.
// Returns concrete Banner struct.
//
// Validates: Requirements 7.4
func (s *Service) Update(ctx context.Context, id string, input UpdateBannerInput) (*banner.Banner, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("banner not found: %w", err)
	}

	oldMediaID := ""
	if b.ImageMediaID != nil {
		oldMediaID = *b.ImageMediaID
	}

	var newMediaID string
	var newMediaCreated bool
	if input.ImageFile != nil || (input.ImageMediaID != nil && *input.ImageMediaID != "") {
		var err error
		newMediaID, newMediaCreated, err = s.resolveImage(ctx, input.ImageMediaID, input.ImageFile, func() (galleryUsecase.SavedFile, error) {
			return s.fileStore.SaveImage(ctx, featureSlug, galleryUsecase.FileInput{
				OriginalName: input.ImageName,
				Content:      input.ImageFile,
				Size:         input.ImageSize,
				ContentType:  input.ContentType,
			})
		})
		if err != nil {
			return nil, err
		}
	}
	cleanup := func() {
		if newMediaCreated {
			_ = s.fileStore.Delete(ctx, newMediaID)
		}
	}

	if input.Category != nil && s.categories != nil {
		if _, err := s.categories.FindByID(ctx, *input.Category); err != nil {
			cleanup()
			return nil, fmt.Errorf("invalid category: %w", err)
		}
		b.Category = input.Category
	}

	b.Title = input.Title
	b.Description = input.Description
	b.Link = input.Link
	if input.Metadata != nil {
		b.Metadata = input.Metadata
	}
	if newMediaID != "" {
		id := newMediaID
		b.ImageMediaID = &id
	}
	b.UpdatedAt = s.clock.Now()

	if err := b.Validate(); err != nil {
		cleanup()
		return nil, fmt.Errorf("invalid banner data: %w", err)
	}

	if err := s.repo.Update(ctx, b); err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to update banner: %w", err)
	}

	if newMediaID != "" && oldMediaID != "" {
		_ = s.fileStore.Delete(ctx, oldMediaID)
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
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("banner not found: %w", err)
	}

	b.Status = status
	b.UpdatedAt = s.clock.Now()

	if err := b.Validate(); err != nil {
		return nil, fmt.Errorf("invalid banner data: %w", err)
	}

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("failed to update banner status: %w", err)
	}

	return b, nil
}

// Delete removes a banner and its associated image file.
//
// Validates: Requirements 7.6
func (s *Service) Delete(ctx context.Context, id string) error {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("banner not found: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete banner: %w", err)
	}

	if b.ImageMediaID != nil && *b.ImageMediaID != "" {
		_ = s.fileStore.Delete(ctx, *b.ImageMediaID)
	}

	return nil
}
