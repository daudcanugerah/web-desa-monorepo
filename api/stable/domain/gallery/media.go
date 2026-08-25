package gallery

import (
	"fmt"
	"strings"
	"time"
)

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

func (m MediaType) Valid() bool {
	return m == MediaTypeImage || m == MediaTypeVideo
}

type Media struct {
	ID               string    `db:"id"`
	FolderID         string    `db:"folder_id"`
	MediaType        MediaType `db:"media_type"`
	FileURL          string    `db:"file_url"`
	ThumbnailURL     *string   `db:"thumbnail_url"`
	ThumbnailFailed  bool      `db:"thumbnail_failed"`
	OriginalFilename string    `db:"original_filename"`
	MimeType         string    `db:"mime_type"`
	FileSize         int64     `db:"file_size"`
	Width            *int      `db:"width"`
	Height           *int      `db:"height"`
	DurationSeconds  *float64  `db:"duration_seconds"`
	IsPublic         bool      `db:"is_public"`
	UploadedBy       string    `db:"uploaded_by"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

const (
	mediaFileURLMaxLen      = 500
	mediaThumbnailURLMaxLen = 500
	mediaOriginalNameMaxLen = 255
	mediaMimeTypeMaxLen     = 100
)

func (m *Media) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("media ID is required")
	}
	if m.FolderID == "" {
		return fmt.Errorf("folder_id is required")
	}
	if !m.MediaType.Valid() {
		return fmt.Errorf("media_type must be one of: image, video")
	}
	if err := validateMediaFileURL(m.FileURL); err != nil {
		return err
	}
	if m.ThumbnailURL != nil {
		if err := validateMediaThumbnailURL(*m.ThumbnailURL); err != nil {
			return err
		}
	}
	if err := validateMediaOriginalFilename(m.OriginalFilename); err != nil {
		return err
	}
	if err := validateMediaMimeType(m.MimeType); err != nil {
		return err
	}
	if err := validateMediaFileSize(m.FileSize); err != nil {
		return err
	}
	if m.UploadedBy == "" {
		return fmt.Errorf("uploaded_by is required")
	}
	if m.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}
	if m.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at timestamp is required")
	}
	return nil
}

func validateMediaFileURL(fileURL string) error {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return fmt.Errorf("file_url is required")
	}
	if len(fileURL) > mediaFileURLMaxLen {
		return fmt.Errorf("file_url must not exceed %d characters", mediaFileURLMaxLen)
	}
	return nil
}

func validateMediaThumbnailURL(thumbnailURL string) error {
	thumbnailURL = strings.TrimSpace(thumbnailURL)
	if thumbnailURL == "" {
		return fmt.Errorf("thumbnail_url cannot be empty string")
	}
	if len(thumbnailURL) > mediaThumbnailURLMaxLen {
		return fmt.Errorf("thumbnail_url must not exceed %d characters", mediaThumbnailURLMaxLen)
	}
	return nil
}

func validateMediaOriginalFilename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("original_filename is required")
	}
	if len(name) > mediaOriginalNameMaxLen {
		return fmt.Errorf("original_filename must not exceed %d characters", mediaOriginalNameMaxLen)
	}
	return nil
}

func validateMediaMimeType(mimeType string) error {
	mimeType = strings.TrimSpace(mimeType)
	if mimeType == "" {
		return fmt.Errorf("mime_type is required")
	}
	if len(mimeType) > mediaMimeTypeMaxLen {
		return fmt.Errorf("mime_type must not exceed %d characters", mediaMimeTypeMaxLen)
	}
	return nil
}

func validateMediaFileSize(size int64) error {
	if size <= 0 {
		return fmt.Errorf("file_size must be positive")
	}
	return nil
}
