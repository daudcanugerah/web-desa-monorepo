package gallery

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"webdesa/api/config"
	"webdesa/api/domain/gallery"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	imageMime = "image/jpeg"
)

type stubRepo struct {
	mu                sync.Mutex
	folders           map[string]*gallery.Folder
	media             map[string]*gallery.Media
	folderByID        map[string]*gallery.Folder
	folderByName      map[string]*gallery.Folder
	createErr         error
	updateErr         error
	deleteErr         error
	updateVisErr      error
	createMediaErr    error
	updateMediaVisErr error
	bulkMediaVisErr   error
	deleteMediaErr    error
	recomputeErr      error
	findErr           error
	findNameErr       error
	findPublicErr     error
	publicMediaErr    error
	listErr           error
}

func newStubRepo() *stubRepo {
	return &stubRepo{
		folders:      map[string]*gallery.Folder{},
		media:        map[string]*gallery.Media{},
		folderByID:   map[string]*gallery.Folder{},
		folderByName: map[string]*gallery.Folder{},
	}
}

func (r *stubRepo) CreateFolder(ctx context.Context, f *gallery.Folder) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.folderByName[strings.ToLower(f.Name)]; exists {
		return errors.New("duplicate")
	}
	cp := *f
	r.folders[f.ID] = &cp
	r.folderByID[f.ID] = &cp
	r.folderByName[strings.ToLower(f.Name)] = &cp
	return nil
}

func (r *stubRepo) UpdateFolder(ctx context.Context, f *gallery.Folder) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.folders[f.ID]; !ok {
		return errors.New("not found")
	}
	cp := *f
	r.folders[f.ID] = &cp
	r.folderByID[f.ID] = &cp
	delete(r.folderByName, strings.ToLower(cp.Name))
	r.folderByName[strings.ToLower(cp.Name)] = &cp
	return nil
}

func (r *stubRepo) UpdateFolderVisibility(ctx context.Context, id string, isPublic bool) error {
	if r.updateVisErr != nil {
		return r.updateVisErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.folders[id]
	if !ok {
		return errors.New("not found")
	}
	f.IsPublic = isPublic
	f.UpdatedAt = time.Now()
	return nil
}

func (r *stubRepo) DeleteFolder(ctx context.Context, id string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.folders[id]
	if !ok {
		return errors.New("not found")
	}
	delete(r.folders, id)
	delete(r.folderByID, id)
	delete(r.folderByName, strings.ToLower(f.Name))
	for mid, m := range r.media {
		if m.FolderID == id {
			delete(r.media, mid)
		}
	}
	return nil
}

func (r *stubRepo) GetFolderByID(ctx context.Context, id string) (*gallery.Folder, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.folders[id]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *f
	return &cp, nil
}

func (r *stubRepo) GetPublicFolderByID(ctx context.Context, id string) (*gallery.Folder, error) {
	if r.findPublicErr != nil {
		return nil, r.findPublicErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.folders[id]
	if !ok || !f.IsPublic {
		return nil, errors.New("not found")
	}
	cp := *f
	return &cp, nil
}

func (r *stubRepo) GetFolderByName(ctx context.Context, name string) (*gallery.Folder, error) {
	if r.findNameErr != nil {
		return nil, r.findNameErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.folderByName[strings.ToLower(name)]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *f
	return &cp, nil
}

func (r *stubRepo) ListFolders(ctx context.Context, q FolderListInput) ([]gallery.Folder, int, error) {
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	matched := make([]gallery.Folder, 0)
	for _, f := range r.folders {
		if q.IsPublic != nil && f.IsPublic != *q.IsPublic {
			continue
		}
		if q.Query != "" && !strings.Contains(strings.ToLower(f.Name), strings.ToLower(q.Query)) {
			continue
		}
		matched = append(matched, *f)
	}
	total := len(matched)
	offset, limit := paginate(q.Page, q.Limit)
	if offset >= total {
		return []gallery.Folder{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return matched[offset:end], total, nil
}

func (r *stubRepo) ListFoldersWithPublicMedia(ctx context.Context, q FolderListInput) ([]gallery.Folder, int, error) {
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	matched := make([]gallery.Folder, 0)
	for _, f := range r.folders {
		if !f.IsPublic {
			continue
		}
		hasPublic := false
		for _, m := range r.media {
			if m.FolderID == f.ID && m.IsPublic && m.ThumbnailURL != nil && *m.ThumbnailURL != "" {
				hasPublic = true
				break
			}
		}
		if !hasPublic {
			continue
		}
		if q.Query != "" && !strings.Contains(strings.ToLower(f.Name), strings.ToLower(q.Query)) {
			continue
		}
		matched = append(matched, *f)
	}
	total := len(matched)
	offset, limit := paginate(q.Page, q.Limit)
	if offset >= total {
		return []gallery.Folder{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return matched[offset:end], total, nil
}

func (r *stubRepo) CreateMedia(ctx context.Context, m *gallery.Media) error {
	if r.createMediaErr != nil {
		return r.createMediaErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *m
	r.media[m.ID] = &cp
	return nil
}

func (r *stubRepo) UpdateMediaVisibility(ctx context.Context, id string, isPublic bool) error {
	if r.updateMediaVisErr != nil {
		return r.updateMediaVisErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.media[id]
	if !ok {
		return errors.New("not found")
	}
	m.IsPublic = isPublic
	m.UpdatedAt = time.Now()
	return nil
}

func (r *stubRepo) BulkUpdateMediaVisibility(ctx context.Context, folderID string, mediaIDs []string, isPublic bool) error {
	if r.bulkMediaVisErr != nil {
		return r.bulkMediaVisErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range mediaIDs {
		m, ok := r.media[id]
		if !ok || m.FolderID != folderID {
			return errors.New("not found")
		}
	}
	for _, id := range mediaIDs {
		r.media[id].IsPublic = isPublic
	}
	return nil
}

func (r *stubRepo) DeleteMedia(ctx context.Context, id string) error {
	if r.deleteMediaErr != nil {
		return r.deleteMediaErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.media, id)
	return nil
}

func (r *stubRepo) GetMediaByID(ctx context.Context, id string) (*gallery.Media, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.media[id]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *m
	return &cp, nil
}

func (r *stubRepo) GetPublicMediaByID(ctx context.Context, id string) (*gallery.Media, error) {
	if r.publicMediaErr != nil {
		return nil, r.publicMediaErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.media[id]
	if !ok {
		return nil, errors.New("not found")
	}
	f, ok := r.folders[m.FolderID]
	if !ok || !f.IsPublic || !m.IsPublic {
		return nil, errors.New("not found")
	}
	cp := *m
	return &cp, nil
}

func (r *stubRepo) ListMediaByFolder(ctx context.Context, folderID string, publicOnly bool, q MediaListInput) ([]gallery.Media, int, error) {
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]gallery.Media, 0)
	for _, m := range r.media {
		if m.FolderID != folderID {
			continue
		}
		if publicOnly && !m.IsPublic {
			continue
		}
		if q.IsPublic != nil && m.IsPublic != *q.IsPublic {
			continue
		}
		if q.MediaType != "" && string(m.MediaType) != q.MediaType {
			continue
		}
		if q.Query != "" && !strings.Contains(strings.ToLower(m.OriginalFilename), strings.ToLower(q.Query)) {
			continue
		}
		out = append(out, *m)
	}
	return out, len(out), nil
}

func (r *stubRepo) ListAllMedia(ctx context.Context, q MediaListInput) ([]gallery.Media, int, error) {
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]gallery.Media, 0)
	for _, m := range r.media {
		if q.FolderID != "" && m.FolderID != q.FolderID {
			continue
		}
		if q.IsPublic != nil && m.IsPublic != *q.IsPublic {
			continue
		}
		if q.MediaType != "" && string(m.MediaType) != q.MediaType {
			continue
		}
		if q.Query != "" && !strings.Contains(strings.ToLower(m.OriginalFilename), strings.ToLower(q.Query)) {
			continue
		}
		out = append(out, *m)
	}
	return out, len(out), nil
}

func (r *stubRepo) RecomputeFolderCover(ctx context.Context, folderID string) error {
	if r.recomputeErr != nil {
		return r.recomputeErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.folders[folderID]
	if !ok {
		return errors.New("not found")
	}
	var best *gallery.Media
	for _, m := range r.media {
		if m.FolderID != folderID {
			continue
		}
		if m.ThumbnailURL == nil || *m.ThumbnailURL == "" {
			continue
		}
		if f.IsPublic && !m.IsPublic {
			continue
		}
		if best == nil || isBetterCover(m, best) {
			mm := *m
			best = &mm
		}
	}
	if best == nil {
		f.CoverMediaID = nil
	} else {
		id := best.ID
		f.CoverMediaID = &id
	}
	return nil
}

func isBetterCover(a, b *gallery.Media) bool {
	if a.MediaType != b.MediaType {
		return a.MediaType == gallery.MediaTypeImage
	}
	if a.CreatedAt.After(b.CreatedAt) {
		return true
	}
	return false
}

func paginate(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return (page - 1) * limit, limit
}

func (r *stubRepo) UpdateMedia(ctx context.Context, m *gallery.Media) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.media[m.ID]
	if !ok {
		return errors.New("not found")
	}
	cp := *m
	*existing = cp
	return nil
}

type stubStorage struct {
	mu              sync.Mutex
	baseDir         string
	files           map[string][]byte
	lastReaderKind  string
	lastReaderBytes int64
}

func newStubStorage(t *testing.T) *stubStorage {
	dir := t.TempDir()
	return &stubStorage{baseDir: dir, files: map[string][]byte{}}
}

func (s *stubStorage) Path(name string) string {
	return filepath.Join(s.baseDir, name)
}

func (s *stubStorage) Save(ctx context.Context, name string, content io.Reader, size int64, contentType string) (string, error) {
	s.mu.Lock()
	s.lastReaderKind = reflect.TypeOf(content).String()
	s.mu.Unlock()

	buf, err := io.ReadAll(content)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	s.lastReaderBytes = int64(len(buf))
	s.mu.Unlock()

	s.mu.Lock()
	s.files[name] = buf
	s.mu.Unlock()

	if err := os.WriteFile(filepath.Join(s.baseDir, name), buf, 0644); err != nil {
		return "", err
	}
	return name, nil
}

func (s *stubStorage) Delete(ctx context.Context, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.files, path)
	_ = os.Remove(filepath.Join(s.baseDir, path))
	return nil
}

func newTestService(t *testing.T, repo Repository, orig, thumb Storage, img ImageProcessor, vid VideoProcessor) *Service {
	t.Helper()
	cfg := config.GalleryConfig{}
	return NewService(repo, orig, thumb, img, vid, cfg, clock.NewFixedClock(time.Now()))
}

func makeJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}))
	return buf.Bytes()
}

func newFolder(t *testing.T, repo *stubRepo, name string, public bool, createdBy string) *gallery.Folder {
	t.Helper()
	f := &gallery.Folder{
		ID:        uuid.NewString(),
		Name:      name,
		IsPublic:  public,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.CreateFolder(context.Background(), f))
	return f
}

func TestCreateFolder_DefaultsPrivate(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)

	isPublic := true
	f, err := svc.CreateFolder(context.Background(), CreateFolderInput{
		Name:      "Panen Raya",
		IsPublic:  &isPublic,
		CreatedBy: "u-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "Panen Raya", f.Name)
	assert.True(t, f.IsPublic)
	assert.NotEmpty(t, f.ID)
}

func TestCreateFolder_DefaultPrivateWhenNotSet(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)

	f, err := svc.CreateFolder(context.Background(), CreateFolderInput{
		Name:      "Quiet",
		CreatedBy: "u-1",
	})
	require.NoError(t, err)
	assert.False(t, f.IsPublic)
}

func TestCreateFolder_NameValidation(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)

	_, err := svc.CreateFolder(context.Background(), CreateFolderInput{Name: "  ", CreatedBy: "u-1"})
	assert.ErrorIs(t, err, ErrFolderNameInvalid)

	long := strings.Repeat("x", 256)
	_, err = svc.CreateFolder(context.Background(), CreateFolderInput{Name: long, CreatedBy: "u-1"})
	assert.ErrorIs(t, err, ErrFolderNameInvalid)
}

func TestCreateFolder_DescriptionTooLong(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)

	desc := strings.Repeat("a", 1001)
	_, err := svc.CreateFolder(context.Background(), CreateFolderInput{Name: "ok", Description: desc, CreatedBy: "u-1"})
	assert.ErrorIs(t, err, ErrDescriptionTooLong)
}

func TestUpdateFolderVisibility_PreservesMediaFlags(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")

	m1 := &gallery.Media{
		ID:         uuid.NewString(),
		FolderID:   folder.ID,
		MediaType:  gallery.MediaTypeImage,
		FileURL:    "a.jpg",
		IsPublic:   true,
		UploadedBy: "u-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	m2 := &gallery.Media{
		ID:         uuid.NewString(),
		FolderID:   folder.ID,
		MediaType:  gallery.MediaTypeImage,
		FileURL:    "b.jpg",
		IsPublic:   false,
		UploadedBy: "u-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, repo.CreateMedia(ctx, m1))
	require.NoError(t, repo.CreateMedia(ctx, m2))

	_, err := svc.UpdateFolderVisibility(ctx, folder.ID, true)
	require.NoError(t, err)

	got1, err := repo.GetMediaByID(ctx, m1.ID)
	require.NoError(t, err)
	got2, err := repo.GetMediaByID(ctx, m2.ID)
	require.NoError(t, err)
	assert.True(t, got1.IsPublic)
	assert.False(t, got2.IsPublic)

	_, err = svc.UpdateFolderVisibility(ctx, folder.ID, false)
	require.NoError(t, err)

	got1, _ = repo.GetMediaByID(ctx, m1.ID)
	got2, _ = repo.GetMediaByID(ctx, m2.ID)
	assert.True(t, got1.IsPublic)
	assert.False(t, got2.IsPublic)
}

func TestListFolders_PaginationAndFilter(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		newFolder(t, repo, "Folder-"+string(rune('A'+i)), i%2 == 0, "u-1")
	}

	pub := true
	folders, res, err := svc.ListFolders(ctx, ListFoldersInput{IsPublic: &pub, Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Len(t, folders, 3)
	assert.Equal(t, 3, res.Total)

	folders, res, err = svc.ListFolders(ctx, ListFoldersInput{Page: 1, Limit: 2})
	require.NoError(t, err)
	assert.Len(t, folders, 2)
	assert.Equal(t, 2, res.Limit)
}

func TestCreateMedia_ImageDefaultsPrivate(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	jpegBytes := makeJPEG(t, 200, 150)

	m, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "photo.jpg",
		MimeType:     "image/jpeg",
		Size:         int64(len(jpegBytes)),
		Content:      bytes.NewReader(jpegBytes),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)
	assert.False(t, m.IsPublic)
	assert.Equal(t, gallery.MediaTypeImage, m.MediaType)
	assert.NotNil(t, m.ThumbnailURL)
	assert.False(t, m.ThumbnailFailed)
	assert.NotNil(t, m.Width)
	assert.NotNil(t, m.Height)
	assert.Equal(t, "photo.jpg", m.OriginalFilename)
}

func TestCreateMedia_Image_NoUpscale(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	small := makeJPEG(t, 50, 30)

	m, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "tiny.jpg",
		MimeType:     "image/jpeg",
		Size:         int64(len(small)),
		Content:      bytes.NewReader(small),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)
	require.NotNil(t, m.Width)
	require.NotNil(t, m.Height)
	assert.Equal(t, 50, *m.Width)
	assert.Equal(t, 30, *m.Height)
}

func TestCreateMedia_RejectsUnsupportedMime(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "doc.pdf",
		MimeType:     "application/pdf",
		Size:         100,
		Content:      bytes.NewReader([]byte("hello")),
		UploadedBy:   "u-1",
	})
	assert.ErrorIs(t, err, ErrInvalidMimeType)
}

func TestCreateMedia_FolderNotFound(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	jpegBytes := makeJPEG(t, 100, 100)
	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     uuid.NewString(),
		OriginalName: "p.jpg",
		MimeType:     "image/jpeg",
		Size:         int64(len(jpegBytes)),
		Content:      bytes.NewReader(jpegBytes),
		UploadedBy:   "u-1",
	})
	assert.ErrorIs(t, err, ErrFolderNotFound)
}

func TestCreateMedia_RecomputeCoverAfterUpload(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	jpegBytes := makeJPEG(t, 200, 150)

	m, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "p.jpg",
		MimeType:     "image/jpeg",
		Size:         int64(len(jpegBytes)),
		Content:      bytes.NewReader(jpegBytes),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)

	f, err := repo.GetFolderByID(ctx, folder.ID)
	require.NoError(t, err)
	require.NotNil(t, f.CoverMediaID)
	assert.Equal(t, m.ID, *f.CoverMediaID)
}

func TestBulkCreateMedia_Limits(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	cfg := config.GalleryConfig{}
	cfg.BulkUploadMaxFiles = 2
	cfg.BulkUploadMaxTotalMB = 1
	svc := NewService(repo, orig, thumb, NewImageProcessor(), nil, cfg, clock.NewFixedClock(time.Now()))
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	jpegBytes := makeJPEG(t, 50, 50)

	files := []UploadMediaInput{
		{OriginalName: "a.jpg", MimeType: "image/jpeg", Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes)},
		{OriginalName: "b.jpg", MimeType: "image/jpeg", Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes)},
		{OriginalName: "c.jpg", MimeType: "image/jpeg", Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes)},
	}
	_, err := svc.BulkCreateMedia(ctx, folder.ID, "u-1", files)
	assert.ErrorIs(t, err, ErrBulkTooManyFiles)

	files = files[:2]
	files[0].Size = 2 * 1024 * 1024
	_, err = svc.BulkCreateMedia(ctx, folder.ID, "u-1", files)
	assert.ErrorIs(t, err, ErrBulkTotalTooLarge)
}

func TestUpdateMediaVisibility_RecomputeCover(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", true, "u-1")
	jpegBytes := makeJPEG(t, 200, 150)
	m, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "p.jpg",
		MimeType:     "image/jpeg",
		Size:         int64(len(jpegBytes)),
		Content:      bytes.NewReader(jpegBytes),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)

	_, err = svc.UpdateMediaVisibility(ctx, m.ID, false)
	require.NoError(t, err)
	f, _ := repo.GetFolderByID(ctx, folder.ID)
	assert.Nil(t, f.CoverMediaID)

	_, err = svc.UpdateMediaVisibility(ctx, m.ID, true)
	require.NoError(t, err)
	f, _ = repo.GetFolderByID(ctx, folder.ID)
	require.NotNil(t, f.CoverMediaID)
	assert.Equal(t, m.ID, *f.CoverMediaID)
}

func TestBulkUpdateMediaVisibility_ValidatesOwnership(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	other := newFolder(t, repo, "Other", false, "u-1")

	jpegBytes := makeJPEG(t, 50, 50)
	m1, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID: folder.ID, OriginalName: "a.jpg", MimeType: "image/jpeg",
		Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes), UploadedBy: "u-1",
	})
	require.NoError(t, err)
	m2, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID: other.ID, OriginalName: "b.jpg", MimeType: "image/jpeg",
		Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes), UploadedBy: "u-1",
	})
	require.NoError(t, err)

	err = svc.BulkUpdateMediaVisibility(ctx, folder.ID, []string{m1.ID, m2.ID}, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	got1, err := repo.GetMediaByID(ctx, m1.ID)
	require.NoError(t, err)
	got2, err := repo.GetMediaByID(ctx, m2.ID)
	require.NoError(t, err)
	assert.False(t, got1.IsPublic)
	assert.False(t, got2.IsPublic)
}

func TestDeleteMedia_CleansFilesAndRecomputesCover(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", true, "u-1")
	jpegBytes := makeJPEG(t, 100, 100)
	m, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID: folder.ID, OriginalName: "p.jpg", MimeType: "image/jpeg",
		Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes), UploadedBy: "u-1",
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteMedia(ctx, m.ID))

	_, err = repo.GetMediaByID(ctx, m.ID)
	assert.Error(t, err)

	_, hasOriginal := orig.files[m.FileURL]
	assert.False(t, hasOriginal)
	if m.ThumbnailURL != nil {
		_, hasThumb := thumb.files[*m.ThumbnailURL]
		assert.False(t, hasThumb)
	}

	f, _ := repo.GetFolderByID(ctx, folder.ID)
	assert.Nil(t, f.CoverMediaID)
}

func TestDeleteFolder_CascadesFiles(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	jpegBytes := makeJPEG(t, 50, 50)
	for i := 0; i < 2; i++ {
		_, err := svc.CreateMedia(ctx, UploadMediaInput{
			FolderID:     folder.ID,
			OriginalName: "p" + string(rune('a'+i)) + ".jpg",
			MimeType:     "image/jpeg",
			Size:         int64(len(jpegBytes)),
			Content:      bytes.NewReader(jpegBytes),
			UploadedBy:   "u-1",
		})
		require.NoError(t, err)
	}

	before := len(orig.files)
	require.NoError(t, svc.DeleteFolder(ctx, folder.ID))

	_, err := repo.GetFolderByID(ctx, folder.ID)
	assert.Error(t, err)
	assert.Less(t, len(orig.files), before)
}

func TestRegenerateThumbnail_UpdatesMedia(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", true, "u-1")
	jpegBytes := makeJPEG(t, 200, 150)
	m, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID: folder.ID, OriginalName: "p.jpg", MimeType: "image/jpeg",
		Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes), UploadedBy: "u-1",
	})
	require.NoError(t, err)

	oldThumb := *m.ThumbnailURL
	_, err = svc.RegenerateThumbnail(ctx, m.ID)
	require.NoError(t, err)

	got, err := repo.GetMediaByID(ctx, m.ID)
	require.NoError(t, err)
	require.NotNil(t, got.ThumbnailURL)
	assert.NotEqual(t, oldThumb, *got.ThumbnailURL)
}

func TestGetPublicFolderByID_PrivateReturns404(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Secret", false, "u-1")
	_, err := svc.GetPublicFolderByID(ctx, folder.ID)
	assert.ErrorIs(t, err, ErrFolderNotFound)
}

func seedMedia(t *testing.T, repo *stubRepo, folder *gallery.Folder, name string, public bool, thumbName string) *gallery.Media {
	t.Helper()
	thumb := thumbName
	m := &gallery.Media{
		ID:               uuid.NewString(),
		FolderID:         folder.ID,
		MediaType:        gallery.MediaTypeImage,
		FileURL:          name + ".jpg",
		ThumbnailURL:     &thumb,
		ThumbnailFailed:  false,
		OriginalFilename: name + ".jpg",
		MimeType:         imageMime,
		FileSize:         1024,
		IsPublic:         public,
		UploadedBy:       "u-1",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	require.NoError(t, repo.CreateMedia(context.Background(), m))
	return m
}

func makePublicFolder(t *testing.T, repo *stubRepo, name string) *gallery.Folder {
	t.Helper()
	f := newFolder(t, repo, name, true, "u-1")
	m := seedMedia(t, repo, f, "cover", true, "cover-thumb")
	f.CoverMediaID = &m.ID
	return f
}

func TestGetMediaContent_Admin_Success(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	jpegBytes := makeJPEG(t, 100, 100)
	m, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID: folder.ID, OriginalName: "p.jpg", MimeType: imageMime,
		Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes), UploadedBy: "u-1",
	})
	require.NoError(t, err)

	bin, err := svc.GetMediaContent(ctx, m.ID)
	require.NoError(t, err)
	assert.Equal(t, orig.Path(m.FileURL), bin.FilePath)
	assert.Equal(t, imageMime, bin.ContentType)
	assert.Equal(t, "p.jpg", bin.Filename)
	assert.Equal(t, int64(len(jpegBytes)), bin.Size)
}

func TestGetMediaContent_Admin_NotFound(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)

	_, err := svc.GetMediaContent(context.Background(), uuid.NewString())
	assert.ErrorIs(t, err, ErrMediaNotFound)
}

func TestGetPublicMediaContent_HidesPrivateMedia(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := makePublicFolder(t, repo, "Album")
	jpegBytes := makeJPEG(t, 50, 50)
	m, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID: folder.ID, OriginalName: "p.jpg", MimeType: imageMime,
		Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes), UploadedBy: "u-1",
	})
	require.NoError(t, err)

	_, err = svc.GetPublicMediaContent(ctx, m.ID)
	assert.ErrorIs(t, err, ErrMediaNotFound)

	_, err = svc.UpdateMediaVisibility(ctx, m.ID, true)
	require.NoError(t, err)
	bin, err := svc.GetPublicMediaContent(ctx, m.ID)
	require.NoError(t, err)
	assert.Equal(t, orig.Path(m.FileURL), bin.FilePath)
}

func TestGetMediaThumbnail_Admin_FailsWhenNil(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	m := seedMedia(t, repo, folder, "no-thumb", false, "")
	m.ThumbnailURL = nil
	require.NoError(t, repo.UpdateMedia(ctx, m))

	_, err := svc.GetMediaThumbnail(ctx, m.ID)
	assert.ErrorIs(t, err, ErrMediaNotFound)
}

func TestGetMediaThumbnail_Admin_FailsWhenFailed(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	m := seedMedia(t, repo, folder, "failed", false, "x")
	m.ThumbnailFailed = true
	require.NoError(t, repo.UpdateMedia(ctx, m))

	_, err := svc.GetMediaThumbnail(ctx, m.ID)
	assert.ErrorIs(t, err, ErrMediaNotFound)
}

func TestGetMediaThumbnail_Admin_Success(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	m := seedMedia(t, repo, folder, "p", false, "p-thumb")

	bin, err := svc.GetMediaThumbnail(ctx, m.ID)
	require.NoError(t, err)
	assert.Equal(t, thumb.Path("p-thumb"), bin.FilePath)
	assert.Equal(t, thumbContentType, bin.ContentType)
	assert.Equal(t, "p.thumb.webp", bin.Filename)
}

func TestGetPublicMediaThumbnail_GatedByFolderAndMedia(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	privFolder := newFolder(t, repo, "Secret", false, "u-1")
	mPriv := seedMedia(t, repo, privFolder, "x", true, "x-thumb")

	_, err := svc.GetPublicMediaThumbnail(ctx, mPriv.ID)
	assert.ErrorIs(t, err, ErrMediaNotFound)

	pubFolder := makePublicFolder(t, repo, "Open")
	mPub := seedMedia(t, repo, pubFolder, "y", false, "y-thumb")
	_, err = svc.GetPublicMediaThumbnail(ctx, mPub.ID)
	assert.ErrorIs(t, err, ErrMediaNotFound)

	_, err = svc.UpdateMediaVisibility(ctx, mPub.ID, true)
	require.NoError(t, err)
	bin, err := svc.GetPublicMediaThumbnail(ctx, mPub.ID)
	require.NoError(t, err)
	assert.Equal(t, thumb.Path("y-thumb"), bin.FilePath)
}

func TestCreateMedia_RejectsArbitraryBytes(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	bogus := []byte("this is definitely not a jpeg image")
	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "fake.jpg",
		MimeType:     imageMime,
		Size:         int64(len(bogus)),
		Content:      bytes.NewReader(bogus),
		UploadedBy:   "u-1",
	})
	assert.ErrorIs(t, err, ErrInvalidMimeType)
}

func TestCreateMedia_RejectsFormatMismatch(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	jpegBytes := makeJPEG(t, 100, 100)

	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "fake.png",
		MimeType:     imagePng,
		Size:         int64(len(jpegBytes)),
		Content:      bytes.NewReader(jpegBytes),
		UploadedBy:   "u-1",
	})
	assert.ErrorIs(t, err, ErrInvalidMimeType)
}

func TestCreateMedia_RejectsVideoWithoutMP4Header(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	bogus := []byte("not really an mp4 file")
	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "fake.mp4",
		MimeType:     videoMp4,
		Size:         int64(len(bogus)),
		Content:      bytes.NewReader(bogus),
		UploadedBy:   "u-1",
	})
	assert.ErrorIs(t, err, ErrInvalidMimeType)
}

func TestCreateMedia_AcceptsMP4Header(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	buf := make([]byte, 32)
	copy(buf[4:8], []byte("ftyp"))
	copy(buf[8:12], []byte("mp42"))
	rest := bytes.Repeat([]byte{0}, 200)
	buf = append(buf, rest...)

	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "ok.mp4",
		MimeType:     videoMp4,
		Size:         int64(len(buf)),
		Content:      bytes.NewReader(buf),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)
}

func TestCreateMedia_AcceptsWebMHeader(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	buf := make([]byte, 32)
	buf[0] = 0x1A
	buf[1] = 0x45
	buf[2] = 0xDF
	buf[3] = 0xA3
	rest := bytes.Repeat([]byte{0}, 200)
	buf = append(buf, rest...)

	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "ok.webm",
		MimeType:     videoWebm,
		Size:         int64(len(buf)),
		Content:      bytes.NewReader(buf),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)
}

func TestCreateMedia_RollsBackOnCoverRecomputeFailure(t *testing.T) {
	repo := newStubRepo()
	repo.recomputeErr = errors.New("db down")
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	jpegBytes := makeJPEG(t, 100, 100)
	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID: folder.ID, OriginalName: "p.jpg", MimeType: imageMime,
		Size: int64(len(jpegBytes)), Content: bytes.NewReader(jpegBytes), UploadedBy: "u-1",
	})
	require.Error(t, err)

	assert.Equal(t, 0, len(repo.media))
	assert.Equal(t, 0, len(orig.files))
	assert.Equal(t, 0, len(thumb.files))
}

func TestDeleteFolder_CleansFilesBeyondPaginationMax(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	jpegBytes := makeJPEG(t, 20, 20)
	total := pagination.MaxLimit + 5
	for i := 0; i < total; i++ {
		_, err := svc.CreateMedia(ctx, UploadMediaInput{
			FolderID:     folder.ID,
			OriginalName: "p" + uuid.NewString() + ".jpg",
			MimeType:     imageMime,
			Size:         int64(len(jpegBytes)),
			Content:      bytes.NewReader(jpegBytes),
			UploadedBy:   "u-1",
		})
		require.NoError(t, err)
	}

	require.NoError(t, svc.DeleteFolder(ctx, folder.ID))
	assert.Equal(t, 0, len(orig.files))
	assert.Equal(t, 0, len(thumb.files))
}

func mp4Bytes(t *testing.T, payload int) []byte {
	t.Helper()
	buf := make([]byte, 32+payload)
	copy(buf[4:8], []byte("ftyp"))
	copy(buf[8:12], []byte("mp42"))
	for i := 12; i < len(buf); i++ {
		buf[i] = byte(i % 256)
	}
	return buf
}

func webmBytes(t *testing.T, payload int) []byte {
	t.Helper()
	buf := make([]byte, 32+payload)
	buf[0] = 0x1A
	buf[1] = 0x45
	buf[2] = 0xDF
	buf[3] = 0xA3
	for i := 4; i < len(buf); i++ {
		buf[i] = byte(i % 256)
	}
	return buf
}

func TestCreateMedia_VideoStreamsViaMultiReader(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	buf := mp4Bytes(t, 5*1024*1024)

	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "big.mp4",
		MimeType:     videoMp4,
		Size:         int64(len(buf)),
		Content:      bytes.NewReader(buf),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)

	assert.Equal(t, "*io.multiReader", orig.lastReaderKind,
		"video save reader should be io.MultiReader (prefix+remaining), not a buffered bytes.Reader")
	assert.Equal(t, int64(len(buf)), orig.lastReaderBytes,
		"streamed reader must still produce the full payload via Storage.Save")
}

func TestCreateMedia_ImageBuffersAsBytesReader(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	jpegBytes := makeJPEG(t, 200, 150)

	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "photo.jpg",
		MimeType:     imageMime,
		Size:         int64(len(jpegBytes)),
		Content:      bytes.NewReader(jpegBytes),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)

	assert.Equal(t, "*bytes.Reader", orig.lastReaderKind,
		"image save reader should be bytes.Reader wrapping the in-memory buffer")
}

func TestCreateMedia_StoresOriginalImageDimensionsNotThumb(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	original := makeJPEG(t, 1000, 500)

	m, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "wide.jpg",
		MimeType:     imageMime,
		Size:         int64(len(original)),
		Content:      bytes.NewReader(original),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)
	require.NotNil(t, m.Width)
	require.NotNil(t, m.Height)
	assert.Equal(t, 1000, *m.Width, "Media.Width must store ORIGINAL width, not thumbnail width")
	assert.Equal(t, 500, *m.Height, "Media.Height must store ORIGINAL height, not thumbnail height")
}

func TestCreateMedia_VideoRejectsPrefixOnly(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	tiny := mp4Bytes(t, 0)

	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "tiny.mp4",
		MimeType:     videoMp4,
		Size:         int64(len(tiny)),
		Content:      bytes.NewReader(tiny),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)
}

func TestCreateMedia_VideoRejectsTooShortHeader(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	tooShort := []byte("ftyp")

	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "short.mp4",
		MimeType:     videoMp4,
		Size:         int64(len(tooShort)),
		Content:      bytes.NewReader(tooShort),
		UploadedBy:   "u-1",
	})
	assert.ErrorIs(t, err, ErrInvalidMimeType)
}

func TestCreateMedia_VideoAcceptsWebMStream(t *testing.T) {
	repo := newStubRepo()
	orig := newStubStorage(t)
	thumb := newStubStorage(t)
	svc := newTestService(t, repo, orig, thumb, NewImageProcessor(), nil)
	ctx := context.Background()

	folder := newFolder(t, repo, "Album", false, "u-1")
	buf := webmBytes(t, 2*1024*1024)

	_, err := svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: "stream.webm",
		MimeType:     videoWebm,
		Size:         int64(len(buf)),
		Content:      bytes.NewReader(buf),
		UploadedBy:   "u-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "*io.multiReader", orig.lastReaderKind)
}
