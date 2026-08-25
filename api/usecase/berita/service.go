package berita

import (
	"context"
	"fmt"
	"io"
	"time"

	"webdesa/api/domain/berita"
	galleryUsecase "webdesa/api/usecase/gallery"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// Service implements berita (news) management business logic.
// It accepts interfaces (Repository, FileStore, CategoryLookup) and
// returns concrete structs (Berita).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo         Repository
	fileStore    galleryUsecase.FileStore
	clock        clock.Clock
	uploadsDir   string
	categoryRepo CategoryLookup
}

// NewService creates a new berita management service.
// Dependencies are injected via constructor following Clean Architecture principles.
//
// uploadsDir is retained for signature compatibility with the existing wiring
// in cmd/serve.go and integration/setup_test.go. The new FileStore port is
// used for image persistence, so uploadsDir is no longer read.
func NewService(repo Repository, fileStore galleryUsecase.FileStore, clk clock.Clock, uploadsDir string, categoryRepo CategoryLookup) *Service {
	return &Service{
		repo:         repo,
		fileStore:    fileStore,
		clock:        clk,
		uploadsDir:   uploadsDir,
		categoryRepo: categoryRepo,
	}
}

const featureSlug = galleryUsecase.FeatureBerita

// CreateBeritaInput represents the input for creating a berita
type CreateBeritaInput struct {
	Title       string
	Content     string
	Category    string
	ImageFile   io.Reader // Optional
	ImageName   string
	ImageSize   int64
	ContentType string
	// ImageMediaID references an image already uploaded via
	// POST /berita/upload-media. Mutually exclusive with ImageFile.
	ImageMediaID *string
}

// UpdateBeritaInput represents the input for updating a berita
type UpdateBeritaInput struct {
	Title       string
	Content     string
	Category    string
	ImageFile   io.Reader // Optional - only if updating image
	ImageName   string
	ImageSize   int64
	ContentType string
	// ImageMediaID references an image already uploaded via
	// POST /berita/upload-media. Mutually exclusive with ImageFile.
	ImageMediaID *string
}

// ListBeritaInput represents the input for listing berita
type ListBeritaInput struct {
	Query    *string    // Optional search query (title or content)
	Category *string    // Optional filter by category
	Since    *time.Time // Optional filter: articles created after this date
	Until    *time.Time // Optional filter: articles created before this date
	Sort     string     // Column to sort by: "created_at" or "title"; defaults to "created_at"
	Order    string     // "asc" or "desc"; defaults to "desc"
	Page     int
	Limit    int
}

// Create creates a new berita with an optional image file.
// Returns concrete Berita struct.
//
// Validates: Requirements 9.3, 10.6
func (s *Service) Create(ctx context.Context, input CreateBeritaInput) (*berita.Berita, error) {
	// Validate the category exists. The FK is enforced at the DB level,
	// but checking here provides a friendlier 400 response.
	if _, err := s.categoryRepo.FindByID(ctx, input.Category); err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}

	var imageMediaID *string

	// Save image file if provided (media ref path skips the file store)
	var imageCreated bool
	if input.ImageMediaID != nil && *input.ImageMediaID != "" {
		if input.ImageFile != nil {
			return nil, fmt.Errorf("provide either an image file or an image media id, not both")
		}
		if _, err := s.fileStore.Open(ctx, *input.ImageMediaID); err != nil {
			return nil, fmt.Errorf("invalid media id: %w", err)
		}
		id := *input.ImageMediaID
		imageMediaID = &id
	} else if input.ImageFile != nil {
		saved, err := s.fileStore.SaveImage(ctx, featureSlug, galleryUsecase.FileInput{
			OriginalName: input.ImageName,
			Content:      input.ImageFile,
			Size:         input.ImageSize,
			ContentType:  input.ContentType,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
		id := saved.MediaID
		imageMediaID = &id
		imageCreated = true
	}

	// Create berita entity. Content is stored as supplied; the legacy
	// processDeltaImages helper has been removed because raw /tmp/ images
	// are gone — Quill deltas must reference media_ids, not URLs.
	now := s.clock.Now()
	b := &berita.Berita{
		ID:           uuid.New().String(),
		Title:        input.Title,
		Content:      input.Content,
		Category:     input.Category,
		ImageMediaID: imageMediaID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Validate domain invariants
	if err := b.Validate(); err != nil {
		// Clean up uploaded media on validation failure (only if created here)
		if imageCreated {
			_ = s.fileStore.Delete(ctx, *imageMediaID)
		}
		return nil, fmt.Errorf("invalid berita data: %w", err)
	}

	// Persist berita
	if err := s.repo.Create(ctx, b); err != nil {
		// Clean up uploaded media on persistence failure (only if created here)
		if imageCreated {
			_ = s.fileStore.Delete(ctx, *imageMediaID)
		}
		return nil, fmt.Errorf("failed to create berita: %w", err)
	}

	return b, nil
}

// GetByID retrieves a berita by ID.
// Returns concrete Berita struct.
//
// Validates: Requirements 9.4, 10.5
func (s *Service) GetByID(ctx context.Context, id string) (*berita.Berita, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("berita not found: %w", err)
	}

	return b, nil
}

// List retrieves paginated berita with optional filtering.
// Returns concrete Berita structs and pagination metadata.
//
// Validates: Requirements 9.1, 10.1, 10.2, 10.3, 10.4, 17.2, 17.3
func (s *Service) List(ctx context.Context, input ListBeritaInput) ([]*berita.Berita, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Retrieve berita from repository with filters
	beritaList, total, err := s.repo.List(ctx, input.Query, input.Category, input.Since, input.Until, input.Sort, input.Order, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list berita: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return beritaList, paginationResult, nil
}

// Update updates an existing berita.
// If ImageFile is provided, updates the image; otherwise keeps existing image.
// Returns concrete Berita struct.
//
// Validates: Requirements 9.5, 10.7
func (s *Service) Update(ctx context.Context, id string, input UpdateBeritaInput) (*berita.Berita, error) {
	// Validate the new category exists before mutating the berita.
	if _, err := s.categoryRepo.FindByID(ctx, input.Category); err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}

	// Retrieve existing berita
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("berita not found: %w", err)
	}

	var oldMediaID string
	if b.ImageMediaID != nil {
		oldMediaID = *b.ImageMediaID
	}

	// Update image if provided (media ref path skips the file store)
	var newMediaID string
	var newMediaCreated bool
	if input.ImageMediaID != nil && *input.ImageMediaID != "" {
		if input.ImageFile != nil {
			return nil, fmt.Errorf("provide either an image file or an image media id, not both")
		}
		if _, err := s.fileStore.Open(ctx, *input.ImageMediaID); err != nil {
			return nil, fmt.Errorf("invalid media id: %w", err)
		}
		newMediaID = *input.ImageMediaID
	} else if input.ImageFile != nil {
		saved, err := s.fileStore.SaveImage(ctx, featureSlug, galleryUsecase.FileInput{
			OriginalName: input.ImageName,
			Content:      input.ImageFile,
			Size:         input.ImageSize,
			ContentType:  input.ContentType,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to save new image: %w", err)
		}
		newMediaID = saved.MediaID
		newMediaCreated = true
	}

	// Update fields
	b.Title = input.Title
	b.Category = input.Category
	b.Content = input.Content
	if newMediaID != "" {
		id := newMediaID
		b.ImageMediaID = &id
	}
	b.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := b.Validate(); err != nil {
		if newMediaCreated {
			_ = s.fileStore.Delete(ctx, newMediaID)
		}
		return nil, fmt.Errorf("invalid berita data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, b); err != nil {
		if newMediaCreated {
			_ = s.fileStore.Delete(ctx, newMediaID)
		}
		return nil, fmt.Errorf("failed to update berita: %w", err)
	}

	// Delete old image if a new one was attached successfully
	if newMediaID != "" && oldMediaID != "" && oldMediaID != newMediaID {
		_ = s.fileStore.Delete(ctx, oldMediaID)
	}

	return b, nil
}

// Delete removes a berita and its associated gallery media.
//
// Validates: Requirements 9.6, 10.8
func (s *Service) Delete(ctx context.Context, id string) error {
	// Retrieve berita to get the media id
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("berita not found: %w", err)
	}

	// Delete berita from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete berita: %w", err)
	}

	// Delete associated gallery media if it exists
	if b.ImageMediaID != nil && *b.ImageMediaID != "" {
		if err := s.fileStore.Delete(ctx, *b.ImageMediaID); err != nil {
			// Log error but don't fail the operation since berita is already deleted
			_ = err
		}
	}

	return nil
}
