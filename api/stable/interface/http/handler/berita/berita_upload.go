package berita

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"webdesa/api/interface/http/middleware"
	"webdesa/api/pkg/response"

	"github.com/google/uuid"
)

// BeritaUploadHandler handles media uploads for Quill editor
type BeritaUploadHandler struct {
	uploadsDir string
	logger     middleware.Logger
}

// NewBeritaUploadHandler creates a new berita upload handler
func NewBeritaUploadHandler(uploadsDir string, logger middleware.Logger) *BeritaUploadHandler {
	return &BeritaUploadHandler{
		uploadsDir: uploadsDir,
		logger:     logger,
	}
}

// UploadMediaRequest represents the upload request
type UploadMediaRequest struct {
	File []byte `json:"file"`
	Name string `json:"name"`
	Type string `json:"type"` // "image" or "video"
}

// UploadMediaResponse represents the upload response
type UploadMediaResponse struct {
	URL string `json:"url"`
}

// UploadMedia godoc
// @Summary      Upload media for Quill editor (admin)
// @Description  Multipart: type=image|video, file. Image max 10 MB, video max 50 MB. Returns tmp filename used in subsequent Berita save. RBAC: berita:write.
// @Tags         berita-upload
// @Accept       mpfd
// @Produce      json
//		@Param	type	formData	string	false	"image|video (default image)"
// @Param        file       formData file   true   "Media file"
// @Success      200  {object} berita.UploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita/upload-media [post]
func (h *BeritaUploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form (max 50MB for videos)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		h.logger.Error(ctx, "failed to parse multipart form", "error", err.Error())
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error(ctx, "failed to get file from form", "error", err.Error())
		response.Error(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Get media type from form
	mediaType := r.FormValue("type")
	if mediaType == "" {
		mediaType = "image"
	}

	// Validate media type
	if mediaType != "image" && mediaType != "video" {
		response.Error(w, http.StatusBadRequest, "Invalid media type. Must be 'image' or 'video'")
		return
	}

	// Validate file size
	fileSize := fileHeader.Size
	maxSize := int64(10 << 20) // 10MB for images
	if mediaType == "video" {
		maxSize = int64(50 << 20) // 50MB for videos
	}

	if fileSize > maxSize {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("File size exceeds maximum of %dMB", maxSize>>20))
		return
	}

	// Validate MIME type
	mimeType := fileHeader.Header.Get("Content-Type")
	if !isValidMediaMimeType(mimeType, mediaType) {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Invalid %s MIME type: %s", mediaType, mimeType))
		return
	}

	// Ensure uploads directory exists
	if err := os.MkdirAll(h.uploadsDir, 0755); err != nil {
		h.logger.Error(ctx, "failed to create uploads directory", "error", err.Error())
		response.Error(w, http.StatusInternalServerError, "Failed to create upload directory")
		return
	}

	// Generate unique filename with tmp- prefix
	filename := generateUniqueFilename(fileHeader.Filename)
	filePath := filepath.Join(h.uploadsDir, filename)

	// Save file to tmp directory
	dst, err := os.Create(filePath)
	if err != nil {
		h.logger.Error(ctx, "failed to create file", "error", err.Error())
		response.Error(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		h.logger.Error(ctx, "failed to copy file", "error", err.Error())
		os.Remove(filePath)
		response.Error(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Return just the filename (file access endpoint handles the uploads directory)
	url := filename

	h.logger.Info(ctx, "media uploaded successfully", "filename", filename, "type", mediaType, "size", fileSize)

	response.Success(w, http.StatusOK, UploadMediaResponse{URL: url})
}

// isValidMediaMimeType validates MIME type for media
func isValidMediaMimeType(mimeType string, mediaType string) bool {
	validImageTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	validVideoTypes := map[string]bool{
		"video/mp4":       true,
		"video/webm":      true,
		"video/ogg":       true,
		"video/quicktime": true,
	}

	if mediaType == "image" {
		return validImageTypes[mimeType]
	}
	if mediaType == "video" {
		return validVideoTypes[mimeType]
	}

	return false
}

// generateUniqueFilename generates a unique filename with format: tmp-yyyymmdd-<uuid>.<ext>
func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	// Generate UUID
	id := uuid.New().String()
	// Get current date in yyyymmdd format
	dateStr := time.Now().Format("20060102")
	// Format: tmp-yyyymmdd-uuid.ext
	return fmt.Sprintf("tmp-%s-%s%s", dateStr, id, ext)
}
