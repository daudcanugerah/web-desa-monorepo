package gallery

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMediaTypeValid(t *testing.T) {
	assert.True(t, MediaTypeImage.Valid())
	assert.True(t, MediaTypeVideo.Valid())
	assert.False(t, MediaType("audio").Valid())
	assert.False(t, MediaType("").Valid())
}

func TestMediaValidate(t *testing.T) {
	now := time.Now()

	validMedia := Media{
		ID:               "m0000000-0000-0000-0000-000000000001",
		FolderID:         "f0000000-0000-0000-0000-000000000001",
		MediaType:        MediaTypeImage,
		FileURL:          "originals/uuid.jpg",
		OriginalFilename: "DSC_1234.jpg",
		MimeType:         "image/jpeg",
		FileSize:         4521984,
		UploadedBy:       "u0000000-0000-0000-0000-000000000001",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	t.Run("valid media passes", func(t *testing.T) {
		assert.NoError(t, validMedia.Validate())
	})

	t.Run("valid media with thumbnail and dimensions passes", func(t *testing.T) {
		m := validMedia
		thumb := "thumbnails/uuid.webp"
		m.ThumbnailURL = &thumb
		w, h := 1920, 1080
		m.Width = &w
		m.Height = &h
		dur := 12.5
		m.DurationSeconds = &dur
		assert.NoError(t, m.Validate())
	})

	missingID := validMedia
	missingID.ID = ""
	assert.Error(t, missingID.Validate())

	missingFolder := validMedia
	missingFolder.FolderID = ""
	assert.Error(t, missingFolder.Validate())

	badType := validMedia
	badType.MediaType = MediaType("audio")
	assert.Error(t, badType.Validate())

	missingFileURL := validMedia
	missingFileURL.FileURL = ""
	assert.Error(t, missingFileURL.Validate())

	longFileURL := validMedia
	longFileURL.FileURL = strings.Repeat("a", mediaFileURLMaxLen+1)
	assert.Error(t, longFileURL.Validate())

	emptyThumb := validMedia
	thumb := "  "
	emptyThumb.ThumbnailURL = &thumb
	assert.Error(t, emptyThumb.Validate())

	longThumb := validMedia
	thumb = strings.Repeat("a", mediaThumbnailURLMaxLen+1)
	longThumb.ThumbnailURL = &thumb
	assert.Error(t, longThumb.Validate())

	missingFilename := validMedia
	missingFilename.OriginalFilename = ""
	assert.Error(t, missingFilename.Validate())

	longFilename := validMedia
	longFilename.OriginalFilename = strings.Repeat("a", mediaOriginalNameMaxLen+1)
	assert.Error(t, longFilename.Validate())

	missingMime := validMedia
	missingMime.MimeType = ""
	assert.Error(t, missingMime.Validate())

	longMime := validMedia
	longMime.MimeType = strings.Repeat("a", mediaMimeTypeMaxLen+1)
	assert.Error(t, longMime.Validate())

	zeroSize := validMedia
	zeroSize.FileSize = 0
	assert.Error(t, zeroSize.Validate())

	negativeSize := validMedia
	negativeSize.FileSize = -1
	assert.Error(t, negativeSize.Validate())

	missingUploader := validMedia
	missingUploader.UploadedBy = ""
	assert.Error(t, missingUploader.Validate())

	missingCreatedAt := validMedia
	missingCreatedAt.CreatedAt = time.Time{}
	assert.Error(t, missingCreatedAt.Validate())

	missingUpdatedAt := validMedia
	missingUpdatedAt.UpdatedAt = time.Time{}
	assert.Error(t, missingUpdatedAt.Validate())
}
