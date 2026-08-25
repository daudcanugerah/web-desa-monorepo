package file

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"webdesa/api/interface/http/middleware"
)

// FileHandler handles file access requests
type FileHandler struct {
	uploadsDir string
	logger     middleware.Logger
}

// NewFileHandler creates a new file handler
func NewFileHandler(uploadsDir string, logger middleware.Logger) *FileHandler {
	return &FileHandler{
		uploadsDir: uploadsDir,
		logger:     logger,
	}
}

// GetFile godoc
// @Summary      Serve uploaded file (public)
// @Description  Public endpoint with directory-traversal protection.
// @Tags         files
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Success      200      {object}  response.FileResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /files/{filename} [get]
// GetFile handles GET /api/v1/files/:filename
// Retrieves a file from the uploads directory with security validation
func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filename := r.PathValue("filename")

	// Validate filename to prevent directory traversal attacks
	if !isValidFilename(filename) {
		h.logger.Error(ctx, "invalid filename requested", "filename", filename)
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	// Construct safe file path
	filePath := filepath.Join(h.uploadsDir, filename)

	// Ensure the resolved path is still within uploads directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		h.logger.Error(ctx, "failed to resolve file path", "filename", filename, "error", err.Error())
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}

	absUploadsDir, err := filepath.Abs(h.uploadsDir)
	if err != nil {
		h.logger.Error(ctx, "failed to resolve uploads directory", "error", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Verify the file is within uploads directory
	if !strings.HasPrefix(absPath, absUploadsDir) {
		h.logger.Error(ctx, "directory traversal attempt detected", "filename", filename, "path", absPath)
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// Check if file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			h.logger.Info(ctx, "file not found", "filename", filename, "absPath", absPath)
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		h.logger.Error(ctx, "failed to stat file", "filename", filename, "error", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Ensure it's a file, not a directory
	if fileInfo.IsDir() {
		h.logger.Error(ctx, "attempted to access directory as file", "filename", filename)
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// Set appropriate Content-Type header
	contentType := getContentType(filename)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// Serve the file
	http.ServeFile(w, r, absPath)

	h.logger.Info(ctx, "file served successfully", "filename", filename, "size", fileInfo.Size())
}

// isValidFilename validates filename to prevent directory traversal attacks
// Allows alphanumeric characters, hyphens, underscores, and dots
// Prevents directory traversal attempts (..)
func isValidFilename(filename string) bool {
	if filename == "" {
		return false
	}

	// Check for directory traversal attempts
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return false
	}

	// Allow alphanumeric, hyphen, underscore, and dot
	// This pattern matches: tmp-20260317-uuid.jpg, image-12345.jpg, etc.
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	return pattern.MatchString(filename)
}

// getContentType returns the appropriate Content-Type for a file based on its extension
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".zip":  "application/zip",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}

	return "application/octet-stream"
}
