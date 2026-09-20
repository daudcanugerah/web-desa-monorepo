package umkm

import (
	"context"
	"fmt"
	"io"

	galleryUsecase "webdesa/api/usecase/gallery"

	"webdesa/api/domain/umkm"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// BulkLimits exposes the gallery bulk-upload caps (max files per record,
// max total bytes) enforced on multi-image create/update flows. The
// config.GalleryConfig type satisfies this interface.
type BulkLimits interface {
	GetBulkUploadMaxFiles() int
	GetBulkUploadMaxTotalMB() int
}

// Service implements UMKM (small business) management business logic.
// It accepts interfaces (Repository, FileStore, CategoryLookup) and
// returns concrete structs (UMKM).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo         Repository
	fileStore    galleryUsecase.FileStore
	clock        clock.Clock
	categoryRepo CategoryLookup
	limits       BulkLimits
}

// NewService creates a new UMKM management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, fileStore galleryUsecase.FileStore, clk clock.Clock, categoryRepo CategoryLookup, limits BulkLimits) *Service {
	return &Service{
		repo:         repo,
		fileStore:    fileStore,
		clock:        clk,
		categoryRepo: categoryRepo,
		limits:       limits,
	}
}

// enforceBulkLimits rejects a multi-image batch that exceeds the gallery
// bulk caps (20 files / 250MB total by default). Refs and inline files are
// counted together because together they form the record's image set.
func (s *Service) enforceBulkLimits(refs []string, files []ImageInput) error {
	maxFiles := s.limits.GetBulkUploadMaxFiles()
	if len(refs)+len(files) > maxFiles {
		return fmt.Errorf("%w: max %d images per record", galleryUsecase.ErrBulkTooManyFiles, maxFiles)
	}
	var total int64
	for _, f := range files {
		total += f.Size
	}
	if total > int64(s.limits.GetBulkUploadMaxTotalMB())*1024*1024 {
		return fmt.Errorf("%w: max %dMB total", galleryUsecase.ErrBulkTotalTooLarge, s.limits.GetBulkUploadMaxTotalMB())
	}
	return nil
}

// featureSlug is the system gallery folder that backs UMKM uploads.
const featureSlug = galleryUsecase.FeatureUMKM

// CreateUMKMInput represents the input for creating a UMKM
type CreateUMKMInput struct {
	Name           string
	Category       string
	Description    string
	Owner          *string
	Address        *string
	Phone          *string
	Email          *string
	Website        *string
	ImageFiles     []ImageInput // Raw multipart files; saved via gallery FileStore
	ImagesMediaIDs []string     // Pre-uploaded gallery media refs (POST /umkm/upload-media)
}

// ImageInput represents an image file input
type ImageInput struct {
	Filename    string
	Content     io.Reader
	Size        int64
	ContentType string
}

// UpdateUMKMInput represents the input for updating a UMKM
type UpdateUMKMInput struct {
	Name           string
	Category       string
	Description    string
	Owner          *string
	Address        *string
	Phone          *string
	Email          *string
	Website        *string
	ImageFiles     []ImageInput // For new file uploads
	ImagesMediaIDs []string     // Pre-uploaded gallery media refs (replaces the set)
}

// ListUMKMInput represents the input for listing UMKM
type ListUMKMInput struct {
	Query    *string // Optional search query (name, description, owner)
	Category *string // Optional filter by category
	Page     int
	Limit    int
}

// saveImageFiles uploads every input through the gallery FileStore and
// returns the resulting media UUIDs in input order. Files are validated
// upfront (per-file type/size caps via FileStore.ValidateFiles) so nothing
// is persisted for a batch with a bad member. A save failure aborts the
// batch and triggers best-effort cleanup of the already-saved siblings.
func (s *Service) saveImageFiles(ctx context.Context, files []ImageInput) ([]string, error) {
	if errs := s.fileStore.ValidateFiles(ctx, featureSlug, toFileInputs(files)); len(errs) > 0 {
		return nil, fmt.Errorf("invalid image: %w", errs[0])
	}
	mediaIDs := make([]string, 0, len(files))
	for i, img := range files {
		saved, err := s.fileStore.SaveImage(ctx, featureSlug, galleryUsecase.FileInput{
			OriginalName: img.Filename,
			Content:      img.Content,
			Size:         img.Size,
			ContentType:  img.ContentType,
		})
		if err != nil {
			for _, id := range mediaIDs {
				_ = s.fileStore.Delete(ctx, id)
			}
			return nil, fmt.Errorf("failed to save image #%d: %w", i, err)
		}
		mediaIDs = append(mediaIDs, saved.MediaID)
	}
	return mediaIDs, nil
}

func toFileInputs(files []ImageInput) []galleryUsecase.FileInput {
	out := make([]galleryUsecase.FileInput, len(files))
	for i, f := range files {
		out[i] = galleryUsecase.FileInput{
			OriginalName: f.Filename,
			ContentType:  f.ContentType,
			Size:         f.Size,
		}
	}
	return out
}

// deleteMediaIDs removes every media UUID from the gallery. Errors are
// swallowed on purpose: stale rows should not break the surrounding flow.
func (s *Service) deleteMediaIDs(ctx context.Context, ids []string) {
	for _, id := range ids {
		_ = s.fileStore.Delete(ctx, id)
	}
}

// mediaIDsNotIn returns the ids from old that are absent from keep — the
// media actually dropped by an update.
func mediaIDsNotIn(old, keep []string) []string {
	keepSet := make(map[string]struct{}, len(keep))
	for _, id := range keep {
		keepSet[id] = struct{}{}
	}
	var out []string
	for _, id := range old {
		if _, ok := keepSet[id]; !ok {
			out = append(out, id)
		}
	}
	return out
}

// validateMediaRefs checks that every pre-uploaded media ref still exists
// in the gallery. Refs were validated at upload time; this guards against
// GC'd or hand-crafted ids.
func (s *Service) validateMediaRefs(ctx context.Context, refs []string) error {
	for _, ref := range refs {
		if _, err := s.fileStore.Open(ctx, ref); err != nil {
			return fmt.Errorf("invalid media id %q: %w", ref, err)
		}
	}
	return nil
}

// Create creates a new UMKM.
// Returns concrete UMKM struct.
//
// Validates: Requirements 10.6, 11.3
func (s *Service) Create(ctx context.Context, input CreateUMKMInput) (*umkm.UMKM, error) {
	// Validate the category exists. The FK is enforced at the DB level,
	// but checking here provides a friendlier 400 response.
	cat, err := s.categoryRepo.FindByID(ctx, input.Category)
	if err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}
	name := cat.Name
	categoryName := &name

	// Validate pre-uploaded media refs before touching the file store.
	if err := s.validateMediaRefs(ctx, input.ImagesMediaIDs); err != nil {
		return nil, err
	}

	// Enforce gallery bulk caps (file count, total bytes) on the combined
	// image set before any file is saved.
	if err := s.enforceBulkLimits(input.ImagesMediaIDs, input.ImageFiles); err != nil {
		return nil, err
	}

	// Save any uploaded files via the gallery FileStore. The returned
	// media UUIDs are the only image identifier new rows carry.
	var imagesMediaIDs []string
	if len(input.ImageFiles) > 0 {
		saved, err := s.saveImageFiles(ctx, input.ImageFiles)
		if err != nil {
			return nil, err
		}
		imagesMediaIDs = append(input.ImagesMediaIDs, saved...)
	} else {
		imagesMediaIDs = input.ImagesMediaIDs
	}

	if imagesMediaIDs == nil {
		imagesMediaIDs = []string{}
	}

	// Create UMKM entity
	now := s.clock.Now()
	u := &umkm.UMKM{
		ID:             uuid.New().String(),
		Name:           input.Name,
		Category:       input.Category,
		CategoryName:   categoryName,
		Description:    input.Description,
		Owner:          input.Owner,
		Address:        input.Address,
		Phone:          input.Phone,
		Email:          input.Email,
		Website:        input.Website,
		ImagesMediaIDs: imagesMediaIDs,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Validate domain invariants
	if err := u.Validate(); err != nil {
		s.deleteMediaIDs(ctx, imagesMediaIDs)
		return nil, fmt.Errorf("invalid umkm data: %w", err)
	}

	// Persist UMKM
	if err := s.repo.Create(ctx, u); err != nil {
		s.deleteMediaIDs(ctx, imagesMediaIDs)
		return nil, fmt.Errorf("failed to create umkm: %w", err)
	}

	return u, nil
}

// GetByID retrieves a UMKM by ID.
// Returns concrete UMKM struct.
//
// Validates: Requirements 10.5, 11.2
func (s *Service) GetByID(ctx context.Context, id string) (*umkm.UMKM, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("umkm not found: %w", err)
	}

	u.ImagesMediaIDs, err = galleryUsecase.PruneDanglingMediaIDs(ctx, s.fileStore, u.ImagesMediaIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to prune media ids: %w", err)
	}

	return u, nil
}

// List retrieves paginated UMKM with optional filtering.
// Returns concrete UMKM structs and pagination metadata.
//
// Validates: Requirements 10.1, 10.2, 11.1, 17.2, 17.3
func (s *Service) List(ctx context.Context, input ListUMKMInput) ([]*umkm.UMKM, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Retrieve UMKM from repository with filters
	umkmList, total, err := s.repo.List(ctx, input.Query, input.Category, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list umkm: %w", err)
	}

	// Prune dangling media ids (Task 5.1) — deleted gallery rows must
	// never appear in API responses.
	for _, u := range umkmList {
		u.ImagesMediaIDs, err = galleryUsecase.PruneDanglingMediaIDs(ctx, s.fileStore, u.ImagesMediaIDs)
		if err != nil {
			return nil, pagination.Result{}, fmt.Errorf("failed to prune media ids: %w", err)
		}
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return umkmList, paginationResult, nil
}

// Update updates an existing UMKM.
// Returns concrete UMKM struct.
//
// Validates: Requirements 10.7, 11.4
func (s *Service) Update(ctx context.Context, id string, input UpdateUMKMInput) (*umkm.UMKM, error) {
	// Validate the new category exists before mutating the UMKM.
	if _, err := s.categoryRepo.FindByID(ctx, input.Category); err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}

	// Retrieve existing UMKM
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("umkm not found: %w", err)
	}

	// Track the gallery media IDs that existed before the update so we
	// can GC the orphans after a successful write.
	oldMediaIDs := append([]string(nil), u.ImagesMediaIDs...)

	// Handle image updates. Three modes:
	//   1. Pre-uploaded media refs supplied → validate, replace
	//      ImagesMediaIDs, drop the old media IDs.
	//   2. New files uploaded via multipart → save through FileStore,
	//      replace ImagesMediaIDs, drop the old media IDs.
	//   3. Neither → leave both arrays untouched.
	var newMediaIDs []string
	switch {
	case len(input.ImagesMediaIDs) > 0:
		if err := s.validateMediaRefs(ctx, input.ImagesMediaIDs); err != nil {
			return nil, err
		}
		if err := s.enforceBulkLimits(input.ImagesMediaIDs, nil); err != nil {
			return nil, err
		}
		u.ImagesMediaIDs = input.ImagesMediaIDs
	case len(input.ImageFiles) > 0:
		if err := s.enforceBulkLimits(nil, input.ImageFiles); err != nil {
			return nil, err
		}
		newMediaIDs, err = s.saveImageFiles(ctx, input.ImageFiles)
		if err != nil {
			return nil, err
		}
		u.ImagesMediaIDs = newMediaIDs
	}

	if u.ImagesMediaIDs == nil {
		u.ImagesMediaIDs = []string{}
	}

	// Update fields
	u.Name = input.Name
	u.Category = input.Category
	u.Description = input.Description
	u.Owner = input.Owner
	u.Address = input.Address
	u.Phone = input.Phone
	u.Email = input.Email
	u.Website = input.Website
	u.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := u.Validate(); err != nil {
		if len(input.ImageFiles) > 0 {
			s.deleteMediaIDs(ctx, newMediaIDs)
		}
		return nil, fmt.Errorf("invalid umkm data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, u); err != nil {
		if len(input.ImageFiles) > 0 {
			s.deleteMediaIDs(ctx, newMediaIDs)
		}
		return nil, fmt.Errorf("failed to update umkm: %w", err)
	}

	// GC only the IDs dropped by this update. The admin form resends
	// unchanged images via existing_images, so those must survive.
	if len(input.ImagesMediaIDs) > 0 || len(input.ImageFiles) > 0 {
		s.deleteMediaIDs(ctx, mediaIDsNotIn(oldMediaIDs, u.ImagesMediaIDs))
	}

	return u, nil
}

// Delete removes a UMKM.
//
// Validates: Requirements 10.8, 11.5
func (s *Service) Delete(ctx context.Context, id string) error {
	// Retrieve UMKM to ensure it exists and get its images
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("umkm not found: %w", err)
	}

	// Delete the row first so a transient gallery hiccup does not leave
	// orphaned media after the row is already gone.
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete umkm: %w", err)
	}

	s.deleteMediaIDs(ctx, u.ImagesMediaIDs)

	return nil
}
