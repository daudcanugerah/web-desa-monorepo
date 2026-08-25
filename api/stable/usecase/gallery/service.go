package gallery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"webdesa/api/config"
	"webdesa/api/domain/gallery"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

const (
	imageJpeg      = "image/jpeg"
	imagePng       = "image/png"
	imageWebp      = "image/webp"
	imageGif       = "image/gif"
	videoMp4       = "video/mp4"
	videoWebm      = "video/webm"
	videoQuicktime = "video/quicktime"

	thumbExt            = ".webp"
	thumbContentType    = imageWebp
	thumbFilenameSuffix = ".thumb.webp"

	videoHeaderSize = 32
)

var (
	ErrFolderNotFound     = errors.New("folder not found")
	ErrMediaNotFound      = errors.New("media not found")
	ErrInvalidMimeType    = errors.New("invalid mime type")
	ErrFileTooLarge       = errors.New("file too large")
	ErrBulkTooManyFiles   = errors.New("bulk upload: too many files")
	ErrBulkTotalTooLarge  = errors.New("bulk upload: total size exceeded")
	ErrInvalidFolderID    = errors.New("invalid folder id")
	ErrInvalidMediaID     = errors.New("invalid media id")
	ErrFolderNameInvalid  = errors.New("folder name invalid")
	ErrDescriptionTooLong = errors.New("folder description too long")
	ErrMediaFilenameEmpty = errors.New("media filename empty")
	ErrCreatedByMissing   = errors.New("created_by is required")
	ErrUploadedByMissing  = errors.New("uploaded_by is required")
)

var allowedImageMIME = map[string]bool{
	imageJpeg: true,
	imagePng:  true,
	imageWebp: true,
	imageGif:  true,
}

var allowedVideoMIME = map[string]bool{
	videoMp4:       true,
	videoWebm:      true,
	videoQuicktime: true,
}

func mediaTypeFromMIME(mt string) gallery.MediaType {
	low := strings.ToLower(mt)
	if allowedImageMIME[low] {
		return gallery.MediaTypeImage
	}
	if allowedVideoMIME[low] {
		return gallery.MediaTypeVideo
	}
	return ""
}

func extForMIME(mt string) string {
	switch strings.ToLower(mt) {
	case imageJpeg:
		return ".jpg"
	case imagePng:
		return ".png"
	case imageWebp:
		return ".webp"
	case imageGif:
		return ".gif"
	case videoMp4:
		return ".mp4"
	case videoWebm:
		return ".webm"
	case videoQuicktime:
		return ".mov"
	}
	if e, _ := mime.ExtensionsByType(mt); len(e) > 0 {
		return e[0]
	}
	return ""
}

func maxBytesFor(cfg config.GalleryConfig, mt gallery.MediaType) int64 {
	switch mt {
	case gallery.MediaTypeImage:
		return int64(cfg.GetImageMaxSizeMB()) * 1024 * 1024
	case gallery.MediaTypeVideo:
		return int64(cfg.GetVideoMaxSizeMB()) * 1024 * 1024
	}
	return 0
}

type thumbnailResult struct {
	path     *string
	width    int
	height   int
	duration float64
	failed   bool
}

type Service struct {
	repo            Repository
	originalStorage Storage
	thumbStorage    Storage
	imageProcessor  ImageProcessor
	videoProcessor  VideoProcessor
	cfg             config.GalleryConfig
	clock           clock.Clock
}

func NewService(
	repo Repository,
	originalStorage Storage,
	thumbStorage Storage,
	imageProcessor ImageProcessor,
	videoProcessor VideoProcessor,
	cfg config.GalleryConfig,
	clk clock.Clock,
) *Service {
	return &Service{
		repo:            repo,
		originalStorage: originalStorage,
		thumbStorage:    thumbStorage,
		imageProcessor:  imageProcessor,
		videoProcessor:  videoProcessor,
		cfg:             cfg,
		clock:           clk,
	}
}

type CreateFolderInput struct {
	Name        string
	Description string
	IsPublic    *bool
	CreatedBy   string
}

type UpdateFolderInput struct {
	Name        string
	Description string
}

type ListFoldersInput struct {
	Query    string
	IsPublic *bool
	Page     int
	Limit    int
}

type ListMediaInput struct {
	FolderID  string
	IsPublic  *bool
	MediaType string
	Query     string
	Page      int
	Limit     int
}

type UploadMediaInput struct {
	FolderID     string
	OriginalName string
	MimeType     string
	Size         int64
	Content      io.Reader
	UploadedBy   string
}

type UploadMediaResult struct {
	Media  *gallery.Media
	Err    error
	Failed bool
}

func (s *Service) CreateFolder(ctx context.Context, in CreateFolderInput) (*gallery.Folder, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || len(name) > 255 {
		return nil, ErrFolderNameInvalid
	}
	if len(in.Description) > 1000 {
		return nil, ErrDescriptionTooLong
	}
	if in.CreatedBy == "" {
		return nil, ErrCreatedByMissing
	}

	isPublic := false
	if in.IsPublic != nil {
		isPublic = *in.IsPublic
	}

	now := s.clock.Now()
	f := &gallery.Folder{
		ID:          uuid.NewString(),
		Name:        name,
		Description: in.Description,
		IsPublic:    isPublic,
		CreatedBy:   in.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := f.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFolderNameInvalid, err.Error())
	}

	if err := s.repo.CreateFolder(ctx, f); err != nil {
		return nil, fmt.Errorf("create folder: %w", err)
	}
	return f, nil
}

func (s *Service) UpdateFolder(ctx context.Context, id string, in UpdateFolderInput) (*gallery.Folder, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidFolderID
	}

	existing, err := s.repo.GetFolderByID(ctx, id)
	if err != nil {
		return nil, ErrFolderNotFound
	}

	name := strings.TrimSpace(in.Name)
	if name == "" || len(name) > 255 {
		return nil, ErrFolderNameInvalid
	}
	if len(in.Description) > 1000 {
		return nil, ErrDescriptionTooLong
	}

	existing.Name = name
	existing.Description = in.Description
	existing.UpdatedAt = s.clock.Now()

	if err := existing.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFolderNameInvalid, err.Error())
	}

	if err := s.repo.UpdateFolder(ctx, existing); err != nil {
		return nil, fmt.Errorf("update folder: %w", err)
	}
	return existing, nil
}

func (s *Service) UpdateFolderVisibility(ctx context.Context, id string, isPublic bool) (*gallery.Folder, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidFolderID
	}

	if _, err := s.repo.GetFolderByID(ctx, id); err != nil {
		return nil, ErrFolderNotFound
	}

	if err := s.repo.UpdateFolderVisibility(ctx, id, isPublic); err != nil {
		return nil, fmt.Errorf("update folder visibility: %w", err)
	}

	if err := s.repo.RecomputeFolderCover(ctx, id); err != nil {
		return nil, fmt.Errorf("recompute folder cover: %w", err)
	}

	return s.repo.GetFolderByID(ctx, id)
}

func (s *Service) DeleteFolder(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidFolderID
	}

	if _, err := s.repo.GetFolderByID(ctx, id); err != nil {
		return ErrFolderNotFound
	}

	mediaList, err := s.collectFolderMedia(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteFolder(ctx, id); err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}

	s.cleanupFolderFiles(mediaList)
	return nil
}

func (s *Service) collectFolderMedia(ctx context.Context, folderID string) ([]gallery.Media, error) {
	const pageSize = pagination.MaxLimit
	var all []gallery.Media
	for page := 1; ; page++ {
		media, total, err := s.repo.ListMediaByFolder(ctx, folderID, false, MediaListInput{
			FolderID: folderID,
			Page:     page,
			Limit:    pageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("list folder media (page %d): %w", page, err)
		}
		all = append(all, media...)
		if len(media) < pageSize || len(all) >= total {
			break
		}
	}
	return all, nil
}

func (s *Service) GetFolderByID(ctx context.Context, id string) (*gallery.Folder, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidFolderID
	}
	f, err := s.repo.GetFolderByID(ctx, id)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	return f, nil
}

func (s *Service) GetPublicFolderByID(ctx context.Context, id string) (*gallery.Folder, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidFolderID
	}
	f, err := s.repo.GetPublicFolderByID(ctx, id)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	return f, nil
}

func (s *Service) GetFolderByName(ctx context.Context, name string) (*gallery.Folder, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrFolderNameInvalid
	}
	return s.repo.GetFolderByName(ctx, name)
}

func (s *Service) ListFolders(ctx context.Context, in ListFoldersInput) ([]gallery.Folder, pagination.Result, error) {
	_, limit, err := pagination.Paginate(in.Page, in.Limit)
	if err != nil {
		return nil, pagination.Result{}, err
	}
	folders, total, err := s.repo.ListFolders(ctx, FolderListInput{
		Query:    strings.TrimSpace(in.Query),
		IsPublic: in.IsPublic,
		Page:     in.Page,
		Limit:    limit,
	})
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("list folders: %w", err)
	}
	return folders, pagination.NewResult(in.Page, limit, total), nil
}

func (s *Service) ListFoldersWithPublicMedia(ctx context.Context, in ListFoldersInput) ([]gallery.Folder, pagination.Result, error) {
	_, limit, err := pagination.Paginate(in.Page, in.Limit)
	if err != nil {
		return nil, pagination.Result{}, err
	}
	folders, total, err := s.repo.ListFoldersWithPublicMedia(ctx, FolderListInput{
		Query:    strings.TrimSpace(in.Query),
		IsPublic: in.IsPublic,
		Page:     in.Page,
		Limit:    limit,
	})
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("list public folders: %w", err)
	}
	return folders, pagination.NewResult(in.Page, limit, total), nil
}

func (s *Service) CreateMedia(ctx context.Context, in UploadMediaInput) (*gallery.Media, error) {
	if _, err := uuid.Parse(in.FolderID); err != nil {
		return nil, ErrInvalidFolderID
	}
	if strings.TrimSpace(in.OriginalName) == "" {
		return nil, ErrMediaFilenameEmpty
	}
	if in.UploadedBy == "" {
		return nil, ErrUploadedByMissing
	}

	mt := strings.ToLower(strings.TrimSpace(in.MimeType))
	mediaType := mediaTypeFromMIME(mt)
	if mediaType == "" {
		return nil, ErrInvalidMimeType
	}
	if err := s.checkSize(mediaType, in.Size); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetFolderByID(ctx, in.FolderID); err != nil {
		return nil, ErrFolderNotFound
	}

	id := uuid.NewString()
	originalName := id + extForMIME(mt)

	var (
		storedPath string
		thumb      thumbnailResult
		origWidth  int
		origHeight int
	)

	switch mediaType {
	case gallery.MediaTypeImage:
		buf, w, h, err := s.readAndValidateImage(mt, in.Content)
		if err != nil {
			return nil, err
		}
		origWidth, origHeight = w, h

		path, err := s.originalStorage.Save(ctx, originalName, bytes.NewReader(buf), int64(len(buf)), mt)
		if err != nil {
			return nil, fmt.Errorf("save original: %w", err)
		}
		storedPath = path

		thumb = s.generateImageThumbnail(ctx, storedPath, buf)

	case gallery.MediaTypeVideo:
		prefix, err := s.validateVideoHeaderStreamed(mt, in.Content)
		if err != nil {
			return nil, err
		}
		reader := io.MultiReader(bytes.NewReader(prefix), in.Content)
		path, err := s.originalStorage.Save(ctx, originalName, reader, in.Size, mt)
		if err != nil {
			return nil, fmt.Errorf("save original: %w", err)
		}
		storedPath = path

		thumb = s.generateVideoThumbnail(ctx, storedPath)

	default:
		return nil, ErrInvalidMimeType
	}

	now := s.clock.Now()
	m := &gallery.Media{
		ID:               id,
		FolderID:         in.FolderID,
		MediaType:        mediaType,
		FileURL:          storedPath,
		ThumbnailURL:     thumb.path,
		ThumbnailFailed:  thumb.failed,
		OriginalFilename: filepath.Base(in.OriginalName),
		MimeType:         mt,
		FileSize:         in.Size,
		Width:            intPtrNonZero(origWidth),
		Height:           intPtrNonZero(origHeight),
		IsPublic:         false,
		UploadedBy:       in.UploadedBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if mediaType == gallery.MediaTypeVideo && thumb.duration > 0 {
		d := thumb.duration
		m.DurationSeconds = &d
	}

	if err := m.Validate(); err != nil {
		s.rollbackMedia(storedPath, thumb.path)
		return nil, fmt.Errorf("invalid media: %w", err)
	}

	if err := s.repo.CreateMedia(ctx, m); err != nil {
		s.rollbackMedia(storedPath, thumb.path)
		return nil, fmt.Errorf("create media: %w", err)
	}

	if err := s.repo.RecomputeFolderCover(ctx, in.FolderID); err != nil {
		_ = s.repo.DeleteMedia(ctx, m.ID)
		s.rollbackMedia(storedPath, thumb.path)
		return nil, fmt.Errorf("recompute cover: %w", err)
	}
	return m, nil
}

func (s *Service) readAndValidateImage(mt string, content io.Reader) ([]byte, int, int, error) {
	maxBytes := maxBytesFor(s.cfg, gallery.MediaTypeImage)
	if maxBytes <= 0 {
		return nil, 0, 0, ErrInvalidMimeType
	}
	limited := io.LimitReader(content, maxBytes+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("read content: %w", err)
	}
	if int64(len(buf)) > maxBytes {
		return nil, 0, 0, ErrFileTooLarge
	}
	if len(buf) == 0 {
		return nil, 0, 0, ErrInvalidMimeType
	}
	return validateImageBytes(mt, buf)
}

func validateImageBytes(mt string, buf []byte) ([]byte, int, int, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		return nil, 0, 0, ErrInvalidMimeType
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, 0, 0, ErrInvalidMimeType
	}
	if int64(cfg.Width)*int64(cfg.Height) > MaxImagePixels {
		return nil, 0, 0, ErrInvalidMimeType
	}
	if _, _, err := image.Decode(bytes.NewReader(buf)); err != nil {
		return nil, 0, 0, ErrInvalidMimeType
	}
	if !imageFormatMatchesMIME(format, mt) {
		return nil, 0, 0, ErrInvalidMimeType
	}
	return buf, cfg.Width, cfg.Height, nil
}

func imageFormatMatchesMIME(format, mt string) bool {
	switch mt {
	case imageJpeg:
		return format == "jpeg"
	case imagePng:
		return format == "png"
	case imageWebp:
		return format == "webp"
	case imageGif:
		return format == "gif"
	}
	return false
}

func (s *Service) validateVideoHeaderStreamed(mt string, content io.Reader) ([]byte, error) {
	prefix := make([]byte, videoHeaderSize)
	n, err := io.ReadFull(content, prefix)
	if err != nil {
		return nil, ErrInvalidMimeType
	}
	if err := checkVideoHeader(mt, prefix[:n]); err != nil {
		return nil, err
	}
	return prefix[:n], nil
}

func checkVideoHeader(mt string, buf []byte) error {
	switch mt {
	case videoMp4, videoQuicktime:
		return validateMP4Header(buf)
	case videoWebm:
		return validateWebMHeader(buf)
	}
	return ErrInvalidMimeType
}

func validateMP4Header(buf []byte) error {
	if len(buf) < 12 {
		return ErrInvalidMimeType
	}
	if string(buf[4:8]) != "ftyp" {
		return ErrInvalidMimeType
	}
	return nil
}

func validateWebMHeader(buf []byte) error {
	if len(buf) < 4 {
		return ErrInvalidMimeType
	}
	if buf[0] != 0x1A || buf[1] != 0x45 || buf[2] != 0xDF || buf[3] != 0xA3 {
		return ErrInvalidMimeType
	}
	return nil
}

func (s *Service) BulkCreateMedia(ctx context.Context, folderID, uploadedBy string, files []UploadMediaInput) ([]UploadMediaResult, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > s.cfg.GetBulkUploadMaxFiles() {
		return nil, ErrBulkTooManyFiles
	}
	var total int64
	for _, f := range files {
		total += f.Size
	}
	if total > int64(s.cfg.GetBulkUploadMaxTotalMB())*1024*1024 {
		return nil, ErrBulkTotalTooLarge
	}

	results := make([]UploadMediaResult, 0, len(files))
	for _, f := range files {
		in := f
		in.FolderID = folderID
		in.UploadedBy = uploadedBy
		m, err := s.CreateMedia(ctx, in)
		if err != nil {
			results = append(results, UploadMediaResult{Err: err, Failed: true})
			continue
		}
		results = append(results, UploadMediaResult{Media: m})
	}
	return results, nil
}

func (s *Service) UpdateMediaVisibility(ctx context.Context, id string, isPublic bool) (*gallery.Media, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidMediaID
	}

	m, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}

	if err := s.repo.UpdateMediaVisibility(ctx, id, isPublic); err != nil {
		return nil, fmt.Errorf("update media visibility: %w", err)
	}

	if err := s.repo.RecomputeFolderCover(ctx, m.FolderID); err != nil {
		return m, fmt.Errorf("recompute cover: %w", err)
	}
	m.IsPublic = isPublic
	m.UpdatedAt = s.clock.Now()
	return m, nil
}

func (s *Service) BulkUpdateMediaVisibility(ctx context.Context, folderID string, mediaIDs []string, isPublic bool) error {
	if _, err := uuid.Parse(folderID); err != nil {
		return ErrInvalidFolderID
	}
	if len(mediaIDs) == 0 {
		return nil
	}
	for _, id := range mediaIDs {
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidMediaID, id)
		}
	}

	if err := s.repo.BulkUpdateMediaVisibility(ctx, folderID, mediaIDs, isPublic); err != nil {
		return fmt.Errorf("bulk update visibility: %w", err)
	}

	if err := s.repo.RecomputeFolderCover(ctx, folderID); err != nil {
		return fmt.Errorf("recompute cover: %w", err)
	}
	return nil
}

func (s *Service) DeleteMedia(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidMediaID
	}

	m, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return ErrMediaNotFound
	}

	if err := s.repo.DeleteMedia(ctx, id); err != nil {
		return fmt.Errorf("delete media: %w", err)
	}

	s.cleanupMediaFiles(m.FileURL, m.ThumbnailURL)

	if err := s.repo.RecomputeFolderCover(ctx, m.FolderID); err != nil {
		return fmt.Errorf("recompute cover: %w", err)
	}
	return nil
}

func (s *Service) GetMediaByID(ctx context.Context, id string) (*gallery.Media, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidMediaID
	}
	m, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	return m, nil
}

func (s *Service) GetPublicMediaByID(ctx context.Context, id string) (*gallery.Media, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidMediaID
	}
	m, err := s.repo.GetPublicMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	return m, nil
}

func (s *Service) ListMediaByFolder(ctx context.Context, folderID string, publicOnly bool, in ListMediaInput) ([]gallery.Media, pagination.Result, error) {
	if _, err := uuid.Parse(folderID); err != nil {
		return nil, pagination.Result{}, ErrInvalidFolderID
	}
	_, limit, err := pagination.Paginate(in.Page, in.Limit)
	if err != nil {
		return nil, pagination.Result{}, err
	}
	in.FolderID = folderID
	media, total, err := s.repo.ListMediaByFolder(ctx, folderID, publicOnly, MediaListInput{
		FolderID:  folderID,
		IsPublic:  in.IsPublic,
		MediaType: in.MediaType,
		Query:     strings.TrimSpace(in.Query),
		Page:      in.Page,
		Limit:     limit,
	})
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("list folder media: %w", err)
	}
	return media, pagination.NewResult(in.Page, limit, total), nil
}

func (s *Service) ListAllMedia(ctx context.Context, in ListMediaInput) ([]gallery.Media, pagination.Result, error) {
	_, limit, err := pagination.Paginate(in.Page, in.Limit)
	if err != nil {
		return nil, pagination.Result{}, err
	}
	media, total, err := s.repo.ListAllMedia(ctx, MediaListInput{
		FolderID:  in.FolderID,
		IsPublic:  in.IsPublic,
		MediaType: in.MediaType,
		Query:     strings.TrimSpace(in.Query),
		Page:      in.Page,
		Limit:     limit,
	})
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("list all media: %w", err)
	}
	return media, pagination.NewResult(in.Page, limit, total), nil
}

func (s *Service) RegenerateThumbnail(ctx context.Context, id string) (*gallery.Media, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidMediaID
	}
	m, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}

	oldThumb := m.ThumbnailURL
	thumb := s.regenerateFromStoredFile(ctx, m)

	m.ThumbnailURL = thumb.path
	m.ThumbnailFailed = thumb.failed
	if m.MediaType == gallery.MediaTypeVideo && thumb.duration > 0 {
		d := thumb.duration
		m.DurationSeconds = &d
	}
	m.UpdatedAt = s.clock.Now()

	if err := s.repo.UpdateMedia(ctx, m); err != nil {
		return nil, fmt.Errorf("update media: %w", err)
	}

	if oldThumb != nil && *oldThumb != "" && (thumb.path == nil || *oldThumb != *thumb.path) {
		_ = s.thumbStorage.Delete(ctx, *oldThumb)
	}

	if err := s.repo.RecomputeFolderCover(ctx, m.FolderID); err != nil {
		return m, fmt.Errorf("recompute cover: %w", err)
	}
	return m, nil
}

func (s *Service) GetMediaContent(ctx context.Context, id string) (*MediaBinary, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrMediaNotFound
	}
	m, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	if m.FileURL == "" {
		return nil, ErrMediaNotFound
	}
	return &MediaBinary{
		FilePath:    s.originalStorage.Path(m.FileURL),
		ContentType: m.MimeType,
		Filename:    m.OriginalFilename,
		Size:        m.FileSize,
	}, nil
}

func (s *Service) GetPublicMediaContent(ctx context.Context, id string) (*MediaBinary, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrMediaNotFound
	}
	m, err := s.repo.GetPublicMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	if m.FileURL == "" {
		return nil, ErrMediaNotFound
	}
	return &MediaBinary{
		FilePath:    s.originalStorage.Path(m.FileURL),
		ContentType: m.MimeType,
		Filename:    m.OriginalFilename,
		Size:        m.FileSize,
	}, nil
}

func (s *Service) GetMediaThumbnail(ctx context.Context, id string) (*MediaBinary, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrMediaNotFound
	}
	m, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	if m.ThumbnailURL == nil || *m.ThumbnailURL == "" || m.ThumbnailFailed {
		return nil, ErrMediaNotFound
	}
	return &MediaBinary{
		FilePath:    s.thumbStorage.Path(*m.ThumbnailURL),
		ContentType: thumbContentType,
		Filename:    thumbnailFilename(m.OriginalFilename),
		Size:        0,
	}, nil
}

func (s *Service) GetPublicMediaThumbnail(ctx context.Context, id string) (*MediaBinary, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrMediaNotFound
	}
	m, err := s.repo.GetPublicMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	if m.ThumbnailURL == nil || *m.ThumbnailURL == "" || m.ThumbnailFailed {
		return nil, ErrMediaNotFound
	}
	return &MediaBinary{
		FilePath:    s.thumbStorage.Path(*m.ThumbnailURL),
		ContentType: thumbContentType,
		Filename:    thumbnailFilename(m.OriginalFilename),
		Size:        0,
	}, nil
}

func thumbnailFilename(original string) string {
	base := filepath.Base(original)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	if stem == "" {
		stem = "thumbnail"
	}
	return stem + thumbFilenameSuffix
}

func (s *Service) generateImageThumbnail(ctx context.Context, storedPath string, imageBytes []byte) thumbnailResult {
	if s.imageProcessor == nil {
		return thumbnailResult{failed: true}
	}
	webp, w, h, err := s.imageProcessor.GenerateThumbnail(ctx, bytes.NewReader(imageBytes), s.cfg.GetThumbnailMaxWidth(), s.cfg.GetThumbnailMaxHeight(), s.cfg.GetThumbnailQuality())
	if err != nil {
		return thumbnailResult{failed: true}
	}
	thumbName := uuid.NewString() + thumbExt
	if _, err := s.thumbStorage.Save(ctx, thumbName, bytes.NewReader(webp), int64(len(webp)), imageWebp); err != nil {
		return thumbnailResult{failed: true}
	}
	return thumbnailResult{path: &thumbName, width: w, height: h, failed: false}
}

func (s *Service) generateVideoThumbnail(ctx context.Context, storedPath string) thumbnailResult {
	if s.videoProcessor == nil || !s.videoProcessor.Available() {
		return thumbnailResult{failed: true}
	}
	thumbName := uuid.NewString() + thumbExt
	thumbAbsPath := s.thumbStorage.Path(thumbName)
	srcAbsPath := s.originalStorage.Path(filepath.Base(storedPath))
	w, h, d, err := s.videoProcessor.GenerateThumbnail(ctx, srcAbsPath, thumbAbsPath)
	if err != nil {
		_ = os.Remove(thumbAbsPath)
		return thumbnailResult{failed: true}
	}
	return thumbnailResult{path: &thumbName, width: w, height: h, duration: d, failed: false}
}

func (s *Service) regenerateFromStoredFile(ctx context.Context, m *gallery.Media) thumbnailResult {
	srcPath := s.originalStorage.Path(filepath.Base(m.FileURL))
	switch m.MediaType {
	case gallery.MediaTypeImage:
		if s.imageProcessor == nil {
			return thumbnailResult{failed: true}
		}
		f, err := os.Open(srcPath)
		if err != nil {
			return thumbnailResult{failed: true}
		}
		defer f.Close()
		webp, w, h, err := s.imageProcessor.GenerateThumbnail(ctx, f, s.cfg.GetThumbnailMaxWidth(), s.cfg.GetThumbnailMaxHeight(), s.cfg.GetThumbnailQuality())
		if err != nil {
			return thumbnailResult{failed: true}
		}
		thumbName := uuid.NewString() + thumbExt
		if _, err := s.thumbStorage.Save(ctx, thumbName, bytes.NewReader(webp), int64(len(webp)), imageWebp); err != nil {
			return thumbnailResult{failed: true}
		}
		return thumbnailResult{path: &thumbName, width: w, height: h, failed: false}

	case gallery.MediaTypeVideo:
		return s.generateVideoThumbnail(ctx, m.FileURL)
	}
	return thumbnailResult{failed: true}
}

func (s *Service) checkSize(mt gallery.MediaType, size int64) error {
	if size <= 0 {
		return ErrFileTooLarge
	}
	switch mt {
	case gallery.MediaTypeImage:
		if size > int64(s.cfg.GetImageMaxSizeMB())*1024*1024 {
			return ErrFileTooLarge
		}
	case gallery.MediaTypeVideo:
		if size > int64(s.cfg.GetVideoMaxSizeMB())*1024*1024 {
			return ErrFileTooLarge
		}
	}
	return nil
}

func (s *Service) rollbackOriginal(storedPath string) {
	if storedPath == "" {
		return
	}
	_ = s.originalStorage.Delete(context.Background(), storedPath)
}

func (s *Service) rollbackMedia(storedPath string, thumbPath *string) {
	s.rollbackOriginal(storedPath)
	if thumbPath != nil && *thumbPath != "" {
		_ = s.thumbStorage.Delete(context.Background(), *thumbPath)
	}
}

func (s *Service) cleanupFolderFiles(media []gallery.Media) {
	for _, m := range media {
		s.cleanupMediaFiles(m.FileURL, m.ThumbnailURL)
	}
}

func (s *Service) cleanupMediaFiles(originalPath string, thumbPath *string) {
	if originalPath != "" {
		_ = s.originalStorage.Delete(context.Background(), originalPath)
	}
	if thumbPath != nil && *thumbPath != "" {
		_ = s.thumbStorage.Delete(context.Background(), *thumbPath)
	}
}

func intPtrNonZero(v int) *int {
	if v <= 0 {
		return nil
	}
	return &v
}
