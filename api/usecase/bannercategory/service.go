package bannercategory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"webdesa/api/domain/bannercategory"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, c *bannercategory.Category) error
	FindByID(ctx context.Context, id string) (*bannercategory.Category, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, query string, offset, limit int) ([]bannercategory.Category, int, error)
	CountByCategoryIDs(ctx context.Context, ids []string) (map[string]int, error)
}

// CategoryInUseError is returned when Delete is attempted on a category that
// banners still reference.
type CategoryInUseError struct {
	CategoryID string
	UsageCount int
}

func (e *CategoryInUseError) Error() string {
	return fmt.Sprintf("banner category %s is in use by %d banner(s)", e.CategoryID, e.UsageCount)
}

// CategoryNotFoundError is returned when the category UUID doesn't exist.
type CategoryNotFoundError struct {
	CategoryID string
}

func (e *CategoryNotFoundError) Error() string {
	return fmt.Sprintf("banner category not found: %s", e.CategoryID)
}

// DuplicateNameError is returned when the unique name constraint is hit.
type DuplicateNameError struct {
	Name string
}

func (e *DuplicateNameError) Error() string {
	return fmt.Sprintf("banner category name already exists: %s", e.Name)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name string) (*bannercategory.Category, error) {
	name = strings.TrimSpace(name)
	c := &bannercategory.Category{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: nowFunc(),
		UpdatedAt: nowFunc(),
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, c); err != nil {
		if isUniqueViolation(err) {
			return nil, &DuplicateNameError{Name: name}
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		var notFound *CategoryNotFoundError
		if errors.As(err, &notFound) {
			return notFound
		}
		// repo.FindByID may return its own not-found wrapper; unwrap
		return &CategoryNotFoundError{CategoryID: id}
	}

	usage, err := s.repo.CountByCategoryIDs(ctx, []string{id})
	if err != nil {
		return err
	}
	if usage[id] > 0 {
		return &CategoryInUseError{CategoryID: id, UsageCount: usage[id]}
	}
	return s.repo.Delete(ctx, id)
}

// ListItem is the enriched category shape returned by ListWithUsage
// (a category row + the number of banners referencing it).
type ListItem struct {
	bannercategory.Category
	UsageCount int
}

func (s *Service) ListWithUsage(ctx context.Context, query string, page, limit int) ([]ListItem, int, error) {
	offset := (page - 1) * limit
	cats, total, err := s.repo.List(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	if len(cats) == 0 {
		return []ListItem{}, total, nil
	}
	ids := make([]string, len(cats))
	for i, c := range cats {
		ids[i] = c.ID
	}
	usage, err := s.repo.CountByCategoryIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ListItem, len(cats))
	for i, c := range cats {
		out[i] = ListItem{Category: c, UsageCount: usage[c.ID]}
	}
	return out, total, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*bannercategory.Category, error) {
	return s.repo.FindByID(ctx, id)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "23505")
}

// nowFunc is package-level so tests can override it.
var nowFunc = func() time.Time { return time.Now() }