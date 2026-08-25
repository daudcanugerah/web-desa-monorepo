package berita

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"webdesa/api/domain/berita"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// Service implements berita (news) management business logic.
// It accepts interfaces (Repository, FileHandler, CategoryLookup) and
// returns concrete structs (Berita).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo         Repository
	fileHandler  FileHandler
	clock        clock.Clock
	uploadsDir   string
	categoryRepo CategoryLookup
}

// NewService creates a new berita management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, fileHandler FileHandler, clk clock.Clock, uploadsDir string, categoryRepo CategoryLookup) *Service {
	return &Service{
		repo:         repo,
		fileHandler:  fileHandler,
		clock:        clk,
		uploadsDir:   uploadsDir,
		categoryRepo: categoryRepo,
	}
}

// processDeltaImages processes Quill delta JSON to move images from tmp to permanent location
// and updates the delta with new URLs
func (s *Service) processDeltaImages(ctx context.Context, content string, uploadsDir string) (string, error) {
	// Try to parse content as JSON (it might be Quill delta)
	var delta map[string]interface{}
	if err := json.Unmarshal([]byte(content), &delta); err != nil {
		// Not JSON, return as-is
		return content, nil
	}

	// Check if it has ops (Quill delta structure)
	ops, ok := delta["ops"].([]interface{})
	if !ok {
		// Not a Quill delta, return as-is
		return content, nil
	}

	// Process each operation to find and move images
	for _, op := range ops {
		opMap, ok := op.(map[string]interface{})
		if !ok {
			continue
		}

		insert, ok := opMap["insert"].(map[string]interface{})
		if !ok {
			continue
		}

		// Check for image URL
		if imgURL, ok := insert["image"].(string); ok && imgURL != "" {
			if strings.Contains(imgURL, "/tmp/") {
				// Move image from tmp to permanent location
				newURL, err := s.moveImageFromTmp(uploadsDir, imgURL)
				if err == nil {
					insert["image"] = newURL
				}
			}
		}

		// Check for video URL
		if videoURL, ok := insert["video"].(string); ok && videoURL != "" {
			if strings.Contains(videoURL, "/tmp/") {
				// Move video from tmp to permanent location
				newURL, err := s.moveImageFromTmp(uploadsDir, videoURL)
				if err == nil {
					insert["video"] = newURL
				}
			}
		}
	}

	// Marshal back to JSON
	updatedDelta, err := json.Marshal(delta)
	if err != nil {
		return content, nil // Return original if marshaling fails
	}

	return string(updatedDelta), nil
}

// moveImageFromTmp moves a file from /uploads/tmp to /uploads
func (s *Service) moveImageFromTmp(uploadsDir string, tmpURL string) (string, error) {
	// Extract filename from URL
	parts := strings.Split(tmpURL, "/")
	filename := parts[len(parts)-1]

	tmpPath := filepath.Join(uploadsDir, "tmp", filename)
	permanentPath := filepath.Join(uploadsDir, filename)

	// Check if tmp file exists
	if _, err := os.Stat(tmpPath); err != nil {
		return "", err
	}

	// Move file
	if err := os.Rename(tmpPath, permanentPath); err != nil {
		return "", err
	}

	// Return new URL
	return fmt.Sprintf("/uploads/%s", filename), nil
}

// CreateBeritaInput represents the input for creating a berita
type CreateBeritaInput struct {
	Title       string
	Content     string
	Category    string
	ImageFile   io.Reader // Optional
	ImageName   string
	ImageSize   int64
	ContentType string
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

	var imageURL *string

	// Save image file if provided
	if input.ImageFile != nil {
		savedPath, err := s.fileHandler.SaveImage(ctx, input.ImageName, input.ImageFile, input.ImageSize, input.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
		imageURL = &savedPath
	}

	// Process delta to move images from tmp to permanent location
	content := input.Content
	if s.uploadsDir != "" {
		processedContent, err := s.processDeltaImages(ctx, content, s.uploadsDir)
		if err == nil {
			content = processedContent
		}
	}

	// Create berita entity
	now := s.clock.Now()
	b := &berita.Berita{
		ID:        uuid.New().String(),
		Title:     input.Title,
		Content:   content,
		Category:  input.Category,
		ImageURL:  imageURL,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate domain invariants
	if err := b.Validate(); err != nil {
		// Clean up uploaded file on validation failure
		if imageURL != nil {
			_ = s.fileHandler.Delete(ctx, *imageURL)
		}
		return nil, fmt.Errorf("invalid berita data: %w", err)
	}

	// Persist berita
	if err := s.repo.Create(ctx, b); err != nil {
		// Clean up uploaded file on persistence failure
		if imageURL != nil {
			_ = s.fileHandler.Delete(ctx, *imageURL)
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

	var oldImageURL *string
	if b.ImageURL != nil {
		oldImageURL = new(string)
		*oldImageURL = *b.ImageURL
	}

	// Update image if provided
	if input.ImageFile != nil {
		newImageURL, err := s.fileHandler.SaveImage(ctx, input.ImageName, input.ImageFile, input.ImageSize, input.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to save new image: %w", err)
		}
		b.ImageURL = &newImageURL
	}

	// Update fields
	b.Title = input.Title
	b.Category = input.Category

	// Process delta to move images from tmp to permanent location
	content := input.Content
	if s.uploadsDir != "" {
		processedContent, err := s.processDeltaImages(ctx, content, s.uploadsDir)
		if err == nil {
			content = processedContent
		}
	}
	b.Content = content
	b.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := b.Validate(); err != nil {
		// Clean up new image if validation fails
		if input.ImageFile != nil && b.ImageURL != nil {
			_ = s.fileHandler.Delete(ctx, *b.ImageURL)
		}
		return nil, fmt.Errorf("invalid berita data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, b); err != nil {
		// Clean up new image if persistence fails
		if input.ImageFile != nil && b.ImageURL != nil {
			_ = s.fileHandler.Delete(ctx, *b.ImageURL)
		}
		return nil, fmt.Errorf("failed to update berita: %w", err)
	}

	// Delete old image if a new one was uploaded successfully
	if input.ImageFile != nil && oldImageURL != nil && *oldImageURL != "" {
		_ = s.fileHandler.Delete(ctx, *oldImageURL)
	}

	return b, nil
}

// Delete removes a berita and its associated image file.
//
// Validates: Requirements 9.6, 10.8
func (s *Service) Delete(ctx context.Context, id string) error {
	// Retrieve berita to get image URL
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("berita not found: %w", err)
	}

	// Delete berita from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete berita: %w", err)
	}

	// Delete associated image file if it exists
	if b.ImageURL != nil && *b.ImageURL != "" {
		if err := s.fileHandler.Delete(ctx, *b.ImageURL); err != nil {
			// Log error but don't fail the operation since berita is already deleted
			// In production, this should be logged properly
			_ = err
		}
	}

	return nil
}
