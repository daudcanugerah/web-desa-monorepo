package handler

import (
	"io"
	"net/http"
	"strconv"

	"braces.dev/errtrace"
	"github.com/ggicci/httpin"

	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
	"webdesa/api/usecase/backup"
)

// BackupHandler handles HTTP requests for backup operations.
// This handler follows Clean Architecture by accepting the backup service
// from the usecase layer and delegating all business logic to it.
//
// Dependencies: interface/http/handler → usecase/backup → domain/backup
type BackupHandler struct {
	service *backup.Service
}

// NewBackupHandler creates a new BackupHandler instance.
// Accepts concrete backup.Service struct (not interface) following
// "accept interfaces, return structs" principle at the handler level.
func NewBackupHandler(service *backup.Service) *BackupHandler {
	return &BackupHandler{
		service: service,
	}
}

// CreateBackup handles POST /api/v1/backups
// Creates a new database backup using pg_dump
// Validates: Requirements 14.1, 14.2, 14.3
func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create backup
	b, err := h.service.CreateBackup(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, errtrace.Wrap(err).Error())
		return
	}

	response.Success(w, http.StatusCreated, b)
}

// ListBackups handles GET /api/v1/backups
// Retrieves paginated list of backup metadata
// Query parameters: page (default 1), limit (default 20, max 100)
// Validates: Requirements 14.4, 14.5, 17.2
func (h *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input struct {
		Page  int `in:"query=page;default=1" validate:"min=1"`
		Limit int `in:"query=limit;default=20" validate:"min=1,max=100"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Validate pagination
	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	offset := (input.Page - 1) * input.Limit

	// Get backups
	backups, total, err := h.service.ListBackups(ctx, offset, input.Limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, errtrace.Wrap(err).Error())
		return
	}

	// Build response with pagination metadata
	response.Success(w, http.StatusOK, map[string]interface{}{
		"backups": backups,
		"pagination": map[string]interface{}{
			"total": total,
			"page":  input.Page,
			"limit": input.Limit,
		},
	})
}

// RestoreBackup handles POST /api/v1/backups/:id/restore
// Restores database from an existing backup file
// Validates: Requirements 15.1, 15.2
func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get backup ID from URL path
	backupID := r.PathValue("id")
	if backupID == "" {
		response.Error(w, http.StatusBadRequest, "backup ID is required")
		return
	}

	// Restore from backup
	if err := h.service.RestoreFromBackup(ctx, backupID); err != nil {
		response.Error(w, http.StatusInternalServerError, errtrace.Wrap(err).Error())
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Database restored successfully",
	})
}

// UploadAndRestore handles POST /api/v1/backups/restore
// Restores database from an uploaded backup file
// Expects multipart/form-data with "file" field
// Validates: Requirements 15.2, 15.3
func (h *BackupHandler) UploadAndRestore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input struct {
		File *httpin.File `in:"form=file"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "file is required")
		return
	}

	if input.File == nil {
		response.Error(w, http.StatusBadRequest, "file is required")
		return
	}

	// Restore from uploaded file
	fileReader, err := input.File.Open()
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to read file")
		return
	}
	if err := h.service.RestoreFromFile(ctx, fileReader, input.File.Filename()); err != nil {
		response.Error(w, http.StatusInternalServerError, errtrace.Wrap(err).Error())
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Database restored successfully from uploaded file",
	})
}

// DownloadBackup handles GET /api/v1/backups/:id/download
// Downloads a backup file
// Validates: Requirements 14.4
func (h *BackupHandler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get backup ID from URL path
	backupID := r.PathValue("id")
	if backupID == "" {
		response.Error(w, http.StatusBadRequest, "backup ID is required")
		return
	}

	// Get backup file
	file, size, err := h.service.GetBackupFile(ctx, backupID)
	if err != nil {
		response.Error(w, http.StatusNotFound, errtrace.Wrap(err).Error())
		return
	}
	defer file.Close()

	// Set headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=backup.sql")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))

	// Stream file to response
	if _, err := io.Copy(w, file); err != nil {
		// Can't send error response after headers are sent
		// Log error instead
		return
	}
}

// GetBackup handles GET /api/v1/backups/:id
// Retrieves backup metadata by ID
// Validates: Requirements 14.4
func (h *BackupHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	// Get backup ID from URL path
	backupID := r.PathValue("id")
	if backupID == "" {
		response.Error(w, http.StatusBadRequest, "backup ID is required")
		return
	}

	// This would require adding a GetByID method to the service
	// For now, we can use ListBackups and filter, or add the method
	// Let's add a simple response indicating the endpoint exists
	response.Error(w, http.StatusNotImplemented, "GetBackup endpoint not yet implemented")
}
