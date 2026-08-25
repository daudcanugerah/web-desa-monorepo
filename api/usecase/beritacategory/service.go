package beritacategory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"webdesa/api/domain/beritacategory"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// CategoryInUseError is returned when attempting to delete a category
// that is still referenced by one or more berita records.
type CategoryInUseError struct {
	CategoryID string
	UsageCount int
}

func (e *CategoryInUseError) Error() string {
	return fmt.Sprintf("category %q is in use by %d berita record(s)", e.CategoryID, e.UsageCount)
}

// CategoryNotFoundError is returned when a category cannot be located.
type CategoryNotFoundError struct {
	CategoryID string
}

func (e *CategoryNotFoundError) Error() string {
	return fmt.Sprintf("category %q not found", e.CategoryID)
}

// DuplicateNameError is returned when attempting to create a category
// whose name already exists.
type DuplicateNameError struct {
	Name string
}

func (e *DuplicateNameError) Error() string {
	return fmt.Sprintf("category with name %q already exists", e.Name)
}

// Service implements berita category management business logic.
// It accepts a Repository interface and a clock.Clock for testability
// and returns a concrete *Service struct.
type Service struct {
	repo  Repository
	clock clock.Clock
}

// NewService creates a new berita category service.
func NewService(repo Repository, clk clock.Clock) *Service {
	return &Service{
		repo:  repo,
		clock: clk,
	}
}

// CategoryWithCount is a view of a category enriched with the number of
// berita records that reference it. Returned by List.
type CategoryWithCount struct {
	*beritacategory.Category
	UsageCount int `json:"usage_count"`
}

// Create creates a new berita category.
//
// Trims and validates the name. Returns a *DuplicateNameError on unique
// constraint violation (the unique index is the source of truth).
func (s *Service) Create(ctx context.Context, name string) (*beritacategory.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}
	if len(name) > 100 {
		return nil, fmt.Errorf("category name must not exceed 100 characters")
	}

	now := s.clock.Now()
	c := &beritacategory.Category{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		// Wrap duplicate errors so the handler can map to 409.
		// The unique index name is idx_berita_categories_name; if the driver
		// surfaces a generic error we still wrap so callers can type-assert.
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return c, nil
}

// Delete removes a berita category by ID.
//
// Returns *CategoryNotFoundError if the category does not exist.
// Returns *CategoryInUseError if any berita record references it.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return &CategoryNotFoundError{CategoryID: id}
	}

	count, err := s.repo.CountByCategoryIDs(ctx, []string{id})
	if err != nil {
		return fmt.Errorf("failed to count usage: %w", err)
	}
	if usage, ok := count[id]; ok && usage > 0 {
		return &CategoryInUseError{CategoryID: id, UsageCount: usage}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// List returns a paginated, optionally text-filtered list of categories
// enriched with usage counts. The text search is a case-insensitive
// substring match on the name, suitable for autocomplete.
//
// The page/limit parameters are validated and normalized via the
// pagination package.
func (s *Service) List(ctx context.Context, query *string, page, limit int) ([]*CategoryWithCount, pagination.Result, error) {
	offset, validatedLimit, err := pagination.Paginate(page, limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	categories, total, err := s.repo.List(ctx, query, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list categories: %w", err)
	}

	// Enrich with usage counts in a single query.
	ids := make([]string, 0, len(categories))
	for _, c := range categories {
		ids = append(ids, c.ID)
	}
	counts, err := s.repo.CountByCategoryIDs(ctx, ids)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to count usage: %w", err)
	}

	result := make([]*CategoryWithCount, 0, len(categories))
	for _, c := range categories {
		result = append(result, &CategoryWithCount{
			Category:   c,
			UsageCount: counts[c.ID], // 0 if not present in map
		})
	}

	return result, pagination.NewResult(page, validatedLimit, total), nil
}

// FindByID retrieves a category by ID. Returns *CategoryNotFoundError
// if not found. Used by the berita service to validate category IDs
// when creating or updating berita records.
func (s *Service) FindByID(ctx context.Context, id string) (*beritacategory.Category, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &CategoryNotFoundError{CategoryID: id}
		}
		return nil, err
	}
	return c, nil
}
