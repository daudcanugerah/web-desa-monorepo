package beritacategory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webdesa/api/domain/beritacategory"
	"webdesa/api/pkg/clock"
)

// mockRepository is a mock implementation of Repository for testing.
type mockRepository struct {
	categories  map[string]*beritacategory.Category
	usageByID   map[string]int
	createErr   error
	deleteErr   error
	findByIDErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		categories: make(map[string]*beritacategory.Category),
		usageByID:  make(map[string]int),
	}
}

func (m *mockRepository) Create(ctx context.Context, c *beritacategory.Category) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.categories[c.ID] = c
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.categories, id)
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*beritacategory.Category, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if c, ok := m.categories[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func (m *mockRepository) List(ctx context.Context, query *string, offset, limit int) ([]*beritacategory.Category, int, error) {
	var filtered []*beritacategory.Category
	for _, c := range m.categories {
		if query != nil && *query != "" {
			if !containsCI(c.Name, *query) {
				continue
			}
		}
		filtered = append(filtered, c)
	}
	total := len(filtered)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

func (m *mockRepository) CountByCategoryIDs(ctx context.Context, ids []string) (map[string]int, error) {
	result := make(map[string]int)
	for _, id := range ids {
		if count, ok := m.usageByID[id]; ok && count > 0 {
			result[id] = count
		}
	}
	return result, nil
}

// containsCI is a simple case-insensitive substring check for tests.
func containsCI(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			a, b := s[i+j], substr[j]
			if a >= 'A' && a <= 'Z' {
				a += 'a' - 'A'
			}
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// fixedClock returns a fixed time for testing.
func newTestClock() clock.Clock {
	return clock.NewFixedClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
}

// TestCreate_Success verifies happy-path category creation.
func TestBeritaCategory_Create_Success(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	c, err := svc.Create(context.Background(), "Pengumuman")
	require.NoError(t, err)
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "Pengumuman", c.Name)
}

// TestCreate_TrimsWhitespace verifies name is trimmed.
func TestBeritaCategory_Create_TrimsWhitespace(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	c, err := svc.Create(context.Background(), "  Spaced  ")
	require.NoError(t, err)
	assert.Equal(t, "Spaced", c.Name)
}

// TestCreate_EmptyName returns error.
func TestBeritaCategory_Create_EmptyName(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	_, err := svc.Create(context.Background(), "")
	assert.Error(t, err)

	_, err = svc.Create(context.Background(), "   ")
	assert.Error(t, err)
}

// TestCreate_TooLong returns error.
func TestBeritaCategory_Create_TooLong(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := svc.Create(context.Background(), string(longName))
	assert.Error(t, err)
}

// TestDelete_Success verifies happy-path deletion.
func TestBeritaCategory_Delete_Success(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	c, _ := svc.Create(context.Background(), "X")
	err := svc.Delete(context.Background(), c.ID)
	assert.NoError(t, err)
}

// TestDelete_NotFound returns CategoryNotFoundError.
func TestBeritaCategory_Delete_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	err := svc.Delete(context.Background(), "ghost-id")
	var notFound *CategoryNotFoundError
	assert.True(t, errors.As(err, &notFound), "expected CategoryNotFoundError, got %T", err)
}

// TestDelete_InUse returns CategoryInUseError.
func TestBeritaCategory_Delete_InUse(t *testing.T) {
	repo := newMockRepository()
	repo.usageByID["in-use-id"] = 5
	repo.categories["in-use-id"] = &beritacategory.Category{ID: "in-use-id", Name: "InUse"}

	svc := NewService(repo, newTestClock())
	err := svc.Delete(context.Background(), "in-use-id")
	var inUse *CategoryInUseError
	assert.True(t, errors.As(err, &inUse))
	assert.Equal(t, 5, inUse.UsageCount)
}

// TestList_WithUsageCount verifies the list enriches with usage counts.
func TestBeritaCategory_List_WithUsageCount(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	repo.categories["a"] = &beritacategory.Category{ID: "a", Name: "Alpha"}
	repo.categories["b"] = &beritacategory.Category{ID: "b", Name: "Beta"}
	repo.usageByID["a"] = 3
	repo.usageByID["b"] = 0

	result, pag, err := svc.List(context.Background(), nil, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, pag.Total)
	assert.Len(t, result, 2)

	alphaCount, betaCount := 0, 0
	for _, r := range result {
		if r.Name == "Alpha" {
			alphaCount = r.UsageCount
		}
		if r.Name == "Beta" {
			betaCount = r.UsageCount
		}
	}
	assert.Equal(t, 3, alphaCount)
	assert.Equal(t, 0, betaCount)
}

// TestList_WithQuery verifies text search filters results.
func TestBeritaCategory_List_WithQuery(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	repo.categories["a"] = &beritacategory.Category{ID: "a", Name: "Pengumuman"}
	repo.categories["b"] = &beritacategory.Category{ID: "b", Name: "Pemberitahuan"}
	repo.categories["c"] = &beritacategory.Category{ID: "c", Name: "Wisata"}

	// "Peng" matches "Pengumuman" and "Pemberitahuan" (the latter contains "peng" if you read it... wait, no)
	// "P-e-n" is in "Pengumuman" (positions 0-2) and "Pemberitahuan" (P-e-m, not Pen)
	// So "P-e-n" only matches "Pengumuman"
	q := "P-e-n-g" // = "Peng"
	_ = q
	// Actually use a more reliable test: substring "ng" matches "Pengumuman" only (Pemberitahuan has "mb" not "ng")
	// And matches "Wisata"? No, "Wisata" doesn't have "ng"
	// Let me just use "ngu" which is unique to Pengumuman
	q = "ngu"
	result, _, err := svc.List(context.Background(), &q, 1, 10)
	require.NoError(t, err)
	assert.Len(t, result, 1, "should match only Pengumuman")

	// No match
	q = "xyz"
	result, _, err = svc.List(context.Background(), &q, 1, 10)
	require.NoError(t, err)
	assert.Len(t, result, 0, "should match nothing")
}

// TestList_Pagination verifies pagination math.
func TestBeritaCategory_List_Pagination(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	for _, n := range []string{"A", "B", "C", "D", "E"} {
		repo.categories[n] = &beritacategory.Category{ID: n, Name: n}
	}

	result, pag, err := svc.List(context.Background(), nil, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, pag.Total)
	assert.Equal(t, 3, pag.TotalPages)
	assert.Len(t, result, 2)
}

// TestList_MaxLimitEnforced verifies MaxLimit is enforced.
func TestBeritaCategory_List_MaxLimitEnforced(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, newTestClock())

	_, _, err := svc.List(context.Background(), nil, 1, 200)
	assert.Error(t, err)
}
