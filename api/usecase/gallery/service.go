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
	"time"

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

// allowedDocumentMIME is the document allowlist used by the PPID
// feature (and any future document upload). PDF, Word, Excel, PowerPoint
// and plain text are the supported office formats.
var allowedDocumentMIME = map[string]bool{
	"application/pdf":                              true,
	"application/msword":                           true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel":                                               true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":      true,
	"application/vnd.ms-powerpoint":                                          true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain":                                                             true,
	"application/octet-stream":                                               true,
}

func mediaTypeFromMIME(mt string) gallery.MediaType {
	low := strings.ToLower(mt)
	if allowedImageMIME[low] {
		return gallery.MediaTypeImage
	}
	if allowedVideoMIME[low] {
		return gallery.MediaTypeVideo
	}
	if allowedDocumentMIME[low] {
		return gallery.MediaTypeDocument
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
	hub             *MediaDeletionHub
	signedURL       *SignedURLService
}

func NewService(
	repo Repository,
	originalStorage Storage,
	thumbStorage Storage,
	imageProcessor ImageProcessor,
	videoProcessor VideoProcessor,
	cfg config.GalleryConfig,
	clk clock.Clock,
	hub *MediaDeletionHub,
	signedURL *SignedURLService,
) *Service {
	return &Service{
		repo:            repo,
		originalStorage: originalStorage,
		thumbStorage:    thumbStorage,
		imageProcessor:  imageProcessor,
		videoProcessor:  videoProcessor,
		cfg:             cfg,
		clock:           clk,
		hub:             hub,
		signedURL:       signedURL,
	}
}

// SignedURL exposes the media-URL signer so response mappers in other
// packages can attach a freshly-minted JWT to the URLs they emit.
// Returns nil when signed URLs are disabled (the caller should fall
// back to the legacy per-scope route pattern).
func (s *Service) SignedURL() *SignedURLService { return s.signedURL }

// Hub exposes the deletion hub so callers can register feature-repo
// listeners after the service is constructed.
func (s *Service) Hub() *MediaDeletionHub {
	return s.hub
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

// SetFolderCover pins a manual cover media for the folder. The media must
// belong to the folder. System folders are immutable via the gallery API.
func (s *Service) SetFolderCover(ctx context.Context, folderID, mediaID string) (*gallery.Folder, error) {
	if _, err := uuid.Parse(folderID); err != nil {
		return nil, ErrInvalidFolderID
	}
	if _, err := uuid.Parse(mediaID); err != nil {
		return nil, ErrInvalidMediaID
	}

	folder, err := s.repo.GetFolderByID(ctx, folderID)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	if folder.IsSystem {
		return nil, ErrSystemFolderImmutable
	}

	m, err := s.repo.GetMediaByID(ctx, mediaID)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	if m.FolderID != folderID {
		return nil, ErrMediaNotFound
	}

	if err := s.repo.SetFolderCover(ctx, folderID, &mediaID); err != nil {
		return nil, fmt.Errorf("set folder cover: %w", err)
	}

	return s.repo.GetFolderByID(ctx, folderID)
}

func (s *Service) UpdateFolderVisibility(ctx context.Context, id string, isPublic bool) (*gallery.Folder, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidFolderID
	}

	folder, err := s.repo.GetFolderByID(ctx, id)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	if folder.IsSystem {
		return nil, ErrSystemFolderImmutable
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

	folder, err := s.repo.GetFolderByID(ctx, id)
	if err != nil {
		return ErrFolderNotFound
	}
	if folder.IsSystem {
		return ErrSystemFolderImmutable
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

	var (
		storedPath  string
		thumb       thumbnailResult
		origWidth   int
		origHeight  int
		actualMime  string
		originalName string
	)

	switch mediaType {
	case gallery.MediaTypeImage:
		buf, detected, w, h, err := s.readAndValidateImage(mt, in.Content)
		if err != nil {
			return nil, err
		}
		origWidth, origHeight = w, h
		actualMime = detected
		originalName = id + extForMIME(detected)

		path, err := s.originalStorage.Save(ctx, originalName, bytes.NewReader(buf), int64(len(buf)), detected)
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
		originalName = id + extForMIME(mt)
		path, err := s.originalStorage.Save(ctx, originalName, reader, in.Size, mt)
		if err != nil {
			return nil, fmt.Errorf("save original: %w", err)
		}
		storedPath = path
		actualMime = mt

		thumb = s.generateVideoThumbnail(ctx, storedPath)

	case gallery.MediaTypeDocument:
		// Documents (PDF, DOC, XLS, ...) are stored verbatim — no decode,
		// no thumbnail. Size + MIME validation already ran.
		originalName = id + extForMIME(mt)
		path, err := s.originalStorage.Save(ctx, originalName, in.Content, in.Size, mt)
		if err != nil {
			return nil, fmt.Errorf("save original: %w", err)
		}
		storedPath = path
		actualMime = mt

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
		MimeType:         actualMime,
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

func (s *Service) readAndValidateImage(mt string, content io.Reader) ([]byte, string, int, int, error) {
	maxBytes := maxBytesFor(s.cfg, gallery.MediaTypeImage)
	if maxBytes <= 0 {
		return nil, "", 0, 0, ErrInvalidMimeType
	}
	limited := io.LimitReader(content, maxBytes+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", 0, 0, fmt.Errorf("read content: %w", err)
	}
	if int64(len(buf)) > maxBytes {
		return nil, "", 0, 0, ErrFileTooLarge
	}
	if len(buf) == 0 {
		return nil, "", 0, 0, ErrInvalidMimeType
	}
	return validateImageBytes(mt, buf)
}

func validateImageBytes(mt string, buf []byte) ([]byte, string, int, int, error) {
	// Validate against the ACTUAL bytes, not the client-declared
	// Content-Type. Browsers and third-party clients frequently label an
	// image with the wrong mime (e.g. sending a JPEG as image/webp), and
	// Go's x/image/webp decoder rejects animated WebP outright. Sniff the
	// real format and accept any supported image type.

	// WebP detection: RIFF + WEBP magic (includes animated VP8X).
	if len(buf) >= 12 && string(buf[0:4]) == "RIFF" && string(buf[8:12]) == "WEBP" {
		w, h, err := validateWebpHeader(buf)
		if err != nil {
			return nil, "", 0, 0, err
		}
		if int64(w)*int64(h) > MaxImagePixels {
			return nil, "", 0, 0, ErrInvalidMimeType
		}
		return buf, imageWebp, w, h, nil
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		return nil, "", 0, 0, ErrInvalidMimeType
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, "", 0, 0, ErrInvalidMimeType
	}
	if int64(cfg.Width)*int64(cfg.Height) > MaxImagePixels {
		return nil, "", 0, 0, ErrInvalidMimeType
	}
	if _, _, err := image.Decode(bytes.NewReader(buf)); err != nil {
		return nil, "", 0, 0, ErrInvalidMimeType
	}
	detected := mimeForFormat(format)
	if detected == "" {
		return nil, "", 0, 0, ErrInvalidMimeType
	}
	return buf, detected, cfg.Width, cfg.Height, nil
}

// mimeForFormat maps a Go image format string to its canonical MIME type.
func mimeForFormat(format string) string {
	switch format {
	case "jpeg":
		return imageJpeg
	case "png":
		return imagePng
	case "gif":
		return imageGif
	case "webp":
		return imageWebp
	}
	return ""
}

// validateWebpHeader checks that buf is a valid WebP container
// (RIFF + WEBP magic + at least one VP8/VP8L/VP8X image chunk) and
// returns its dimensions. Unlike image.Decode it accepts animated WebP
// (VP8X with an animation flag), which the standard library decoder
// rejects.
func validateWebpHeader(buf []byte) (width, height int, err error) {
	if len(buf) < 12 || string(buf[0:4]) != "RIFF" || string(buf[8:12]) != "WEBP" {
		return 0, 0, ErrInvalidMimeType
	}
	// Walk the RIFF chunks looking for VP8 / VP8L / VP8X.
	pos := 12
	found := false
	for pos+8 <= len(buf) {
		chunkID := string(buf[pos : pos+4])
		size := int(buf[pos+4]) | int(buf[pos+5])<<8 | int(buf[pos+6])<<16 | int(buf[pos+7])<<24
		if pos+8+size > len(buf) {
			return 0, 0, ErrInvalidMimeType
		}
		body := buf[pos+8 : pos+8+size]
		switch chunkID {
		case "VP8 ":
			// VP8 lossy: frame tag (3 bytes) + start code 9D 01 2A + 14-bit dims.
			if len(body) >= 10 && body[3] == 0x9d && body[4] == 0x01 && body[5] == 0x2a {
				width = int(body[6]) | int(body[7]&0x3f)<<8
				height = int(body[8]) | int(body[9]&0x3f)<<8
				found = true
			}
		case "VP8L":
			// VP8L lossless: 1-byte signature 0x2f + 14-bit dims in next 4 bytes.
			if len(body) >= 5 && body[0] == 0x2f {
				b := uint32(body[1]) | uint32(body[2])<<8 | uint32(body[3])<<16 | uint32(body[4])<<24
				width = int(b&0x3fff) + 1
				height = int((b>>14)&0x3fff) + 1
				found = true
			}
		case "VP8X":
			// VP8X extended (may be animated): 1 reserved byte + 24-bit dims-1.
			if len(body) >= 10 {
				width = int(body[4]) | int(body[5])<<8 | int(body[6])<<16
				height = int(body[7]) | int(body[8])<<8 | int(body[9])<<16
				width++
				height++
				found = true
			}
		}
		if found {
			break
		}
		pos += 8 + size
		if size%2 == 1 {
			pos++ // RIFF chunks are word-aligned
		}
	}
	if !found || width <= 0 || height <= 0 {
		return 0, 0, ErrInvalidMimeType
	}
	return width, height, nil
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
	if err := s.ensureMutableMedia(ctx, m); err != nil {
		return nil, err
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
	folder, err := s.repo.GetFolderByID(ctx, folderID)
	if err != nil {
		return ErrFolderNotFound
	}
	if folder.IsSystem {
		return ErrSystemFolderImmutable
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
	if err := s.ensureMutableMedia(ctx, m); err != nil {
		return err
	}

	return s.deleteMedia(ctx, m)
}

// DeleteMediaForFeature removes media on behalf of a feature (berita, umkm,
// ppid, ...). System folders are immutable via the gallery API, but feature
// services must be able to clean up the media they created.
func (s *Service) DeleteMediaForFeature(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidMediaID
	}

	m, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return ErrMediaNotFound
	}

	return s.deleteMedia(ctx, m)
}

func (s *Service) deleteMedia(ctx context.Context, m *gallery.Media) error {
	if err := s.repo.DeleteMedia(ctx, m.ID); err != nil {
		return fmt.Errorf("delete media: %w", err)
	}

	s.cleanupMediaFiles(m.FileURL, m.ThumbnailURL)

	// If the deleted media was a manually pinned cover, release the pin so
	// auto-recompute can pick the next eligible candidate.
	if folder, err := s.repo.GetFolderByID(ctx, m.FolderID); err == nil &&
		folder.CoverManual && folder.CoverMediaID != nil && *folder.CoverMediaID == m.ID {
		if err := s.repo.SetFolderCover(ctx, m.FolderID, nil); err != nil {
			return fmt.Errorf("clear manual cover: %w", err)
		}
	}

	if err := s.repo.RecomputeFolderCover(ctx, m.FolderID); err != nil {
		return fmt.Errorf("recompute cover: %w", err)
	}

	// Notify feature listeners so they can sweep the deleted UUID out of
	// their `*_media_ids` JSONB arrays (Task 5.1). Best-effort: the
	// hub swallows listener errors and on-read pruning covers any rows
	// the sweep missed.
	s.hub.Notify(ctx, m.ID)

	return nil
}

// ensureMutableMedia rejects mutations on media living in system folders.
func (s *Service) ensureMutableMedia(ctx context.Context, m *gallery.Media) error {
	folder, err := s.repo.GetFolderByID(ctx, m.FolderID)
	if err != nil {
		return ErrFolderNotFound
	}
	if folder.IsSystem {
		return ErrSystemFolderImmutable
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
	if err := s.ensureNonSystemMedia(ctx, m); err != nil {
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
	if err := s.ensureNonSystemMedia(ctx, m); err != nil {
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

// getSignedPublicMedia resolves a media id for a scope=public signed
// URL. It allows media that is either (a) in a public folder, or (b) in
// a feature system folder whose feature is public-exposed via the
// feature-scoped URLs (banner, berita, struktur, umkm, fasilitas).
// system/ppid and system/user media are NEVER served to public tokens.
func (s *Service) getSignedPublicMedia(ctx context.Context, id string) (*gallery.Media, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrMediaNotFound
	}
	m, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	folder, err := s.repo.GetFolderByID(ctx, m.FolderID)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	if folder.IsSystem {
		if folder.FeatureSlug == nil {
			return nil, ErrMediaNotFound
		}
		switch *folder.FeatureSlug {
		case FeatureBanner, FeatureBerita, FeatureStruktur, FeatureUMKM, FeatureFasilitas:
			// feature system folders exposed publicly via signed URLs
		default:
			// system/ppid, system/user and any future private feature
			return nil, ErrMediaNotFound
		}
	} else if !folder.IsPublic {
		// non-system folder that isn't public
		return nil, ErrMediaNotFound
	}
	return m, nil
}

// GetSignedPublicMediaContent serves the content binary for a scope=public
// signed URL. Only media the public scope is entitled to is streamed.
func (s *Service) GetSignedPublicMediaContent(ctx context.Context, id string) (*MediaBinary, error) {
	m, err := s.getSignedPublicMedia(ctx, id)
	if err != nil {
		return nil, err
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

// GetSignedPublicMediaThumbnail serves the thumbnail binary for a
// scope=public signed URL.
func (s *Service) GetSignedPublicMediaThumbnail(ctx context.Context, id string) (*MediaBinary, error) {
	m, err := s.getSignedPublicMedia(ctx, id)
	if err != nil {
		return nil, err
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

// ensureNonSystemMedia rejects public streaming of feature-bound media even
// if visibility flags were (accidentally) flipped. System folders are
// feature-private by design.
func (s *Service) ensureNonSystemMedia(ctx context.Context, m *gallery.Media) error {
	folder, err := s.repo.GetFolderByID(ctx, m.FolderID)
	if err != nil {
		return ErrFolderNotFound
	}
	if folder.IsSystem {
		return ErrSystemFolderImmutable
	}
	return nil
}

// GetFeatureMediaForPublic streams a media binary only when the media lives
// in the given feature's system folder. Feature-scoped public endpoints
// (banner covers, berita images) are the sole public path for system media;
// the generic public gallery endpoints reject them.
func (s *Service) GetFeatureMediaForPublic(ctx context.Context, feature, id string) (*MediaBinary, error) {
	m, err := s.getFeatureMedia(ctx, feature, id)
	if err != nil {
		return nil, err
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

// GetFeatureMediaThumbnailForPublic streams a thumbnail for a media that
// lives in the given feature's system folder.
func (s *Service) GetFeatureMediaThumbnailForPublic(ctx context.Context, feature, id string) (*MediaBinary, error) {
	m, err := s.getFeatureMedia(ctx, feature, id)
	if err != nil {
		return nil, err
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

// getFeatureMedia resolves a media id and verifies it belongs to the
// system folder backing the given feature slug.
func (s *Service) getFeatureMedia(ctx context.Context, feature, id string) (*gallery.Media, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrMediaNotFound
	}
	if systemFolderSpecFor(strings.ToLower(feature)) == nil {
		return nil, ErrMediaNotFound
	}
	m, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	folder, err := s.repo.GetFolderByID(ctx, m.FolderID)
	if err != nil || !folder.IsSystem || folder.FeatureSlug == nil || *folder.FeatureSlug != strings.ToLower(feature) {
		return nil, ErrMediaNotFound
	}
	return m, nil
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
	case gallery.MediaTypeDocument:
		if size > int64(s.cfg.GetDocumentMaxSizeMB())*1024*1024 {
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

// EnsureSystemFolders upserts every canonical system folder in specs.
// Existing rows are left in place so re-running the seed is a no-op; the
// feature slug uniqueness constraint (partial unique index) makes the
// upsert atomic via CreateSystemFolder.
func (s *Service) EnsureSystemFolders(ctx context.Context, specs []SystemFolderSpec, createdBy string) ([]*gallery.Folder, error) {
	if strings.TrimSpace(createdBy) == "" {
		return nil, ErrCreatedByMissing
	}
	out := make([]*gallery.Folder, 0, len(specs))
	for _, spec := range specs {
		slug := strings.ToLower(strings.TrimSpace(spec.Slug))
		if slug == "" {
			return nil, fmt.Errorf("%w: empty slug", ErrFolderNameInvalid)
		}
		now := s.clock.Now()
		f := &gallery.Folder{
			ID:          uuid.NewString(),
			Name:        spec.Name,
			Description: spec.Description,
			IsPublic:    false,
			IsSystem:    true,
			FeatureSlug: &slug,
			CreatedBy:   createdBy,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.repo.CreateSystemFolder(ctx, f); err != nil {
			return nil, fmt.Errorf("ensure system folder %s: %w", slug, err)
		}
		out = append(out, f)
	}
	return out, nil
}

// GetSystemFolderByFeature returns the seeded system folder for the given
// feature slug or ErrFolderNotFound. FileStore uses this to find the
// destination folder for every upload.
func (s *Service) GetSystemFolderByFeature(ctx context.Context, feature string) (*gallery.Folder, error) {	slug, err := normalizeFeature(feature)
	if err != nil {
		return nil, err
	}
	folder, err := s.repo.GetFolderByFeatureSlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("%w: feature %s", ErrFolderNotFound, slug)
	}
	return folder, nil
}

// BootstrapSystemFolderOwner returns a user id that can own system folders
// created on demand. Uses the oldest user as a stable bootstrap owner.
func (s *Service) BootstrapSystemFolderOwner(ctx context.Context) (string, error) {
	owner, err := s.repo.FindSystemFolderOwner(ctx)
	if err != nil {
		return "", fmt.Errorf("find system folder owner: %w", err)
	}
	return owner, nil
}

// FindOrphanMedia lists system-folder media unreferenced by any feature,
// created before the cutoff. Callers delete via DeleteMediaForFeature so
// the immutability guard does not block the sweep.
func (s *Service) FindOrphanMedia(ctx context.Context, olderThan time.Time, limit int) ([]gallery.Media, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.FindOrphanMedia(ctx, olderThan, limit)
}

// CreateDocument validates and stores a single document file. Documents
// are never thumbnailed; the storage key mirrors the image flow but
// skips any media-type-specific work. Exposed via FileStore.SaveDocument.
func (s *Service) CreateDocument(ctx context.Context, in UploadMediaInput) (*gallery.Media, error) {
	mt := strings.ToLower(strings.TrimSpace(in.MimeType))
	if mediaTypeFromMIME(mt) != gallery.MediaTypeDocument {
		return nil, ErrInvalidMimeType
	}
	return s.CreateMedia(ctx, in)
}
