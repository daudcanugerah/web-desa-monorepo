package gallery

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	"webdesa/api/config"
	"webdesa/api/domain/gallery"
	"webdesa/api/interface/http/middleware"
	usecaseGallery "webdesa/api/usecase/gallery"

	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/pagination"
	"webdesa/api/pkg/response"
)

type GalleryService interface {
	CreateFolder(ctx context.Context, in usecaseGallery.CreateFolderInput) (*gallery.Folder, error)
	UpdateFolder(ctx context.Context, id string, in usecaseGallery.UpdateFolderInput) (*gallery.Folder, error)
	UpdateFolderVisibility(ctx context.Context, id string, isPublic bool) (*gallery.Folder, error)
	SetFolderCover(ctx context.Context, folderID, mediaID string) (*gallery.Folder, error)
	DeleteFolder(ctx context.Context, id string) error
	GetFolderByID(ctx context.Context, id string) (*gallery.Folder, error)
	GetPublicFolderByID(ctx context.Context, id string) (*gallery.Folder, error)
	ListFolders(ctx context.Context, in usecaseGallery.ListFoldersInput) ([]gallery.Folder, pagination.Result, error)
	ListFoldersWithPublicMedia(ctx context.Context, in usecaseGallery.ListFoldersInput) ([]gallery.Folder, pagination.Result, error)

	CreateMedia(ctx context.Context, in usecaseGallery.UploadMediaInput) (*gallery.Media, error)
	BulkCreateMedia(ctx context.Context, folderID, uploadedBy string, files []usecaseGallery.UploadMediaInput) ([]usecaseGallery.UploadMediaResult, error)
	UpdateMediaVisibility(ctx context.Context, id string, isPublic bool) (*gallery.Media, error)
	BulkUpdateMediaVisibility(ctx context.Context, folderID string, mediaIDs []string, isPublic bool) error
	DeleteMedia(ctx context.Context, id string) error
	GetMediaByID(ctx context.Context, id string) (*gallery.Media, error)
	GetPublicMediaByID(ctx context.Context, id string) (*gallery.Media, error)
	ListMediaByFolder(ctx context.Context, folderID string, publicOnly bool, in usecaseGallery.ListMediaInput) ([]gallery.Media, pagination.Result, error)
	ListAllMedia(ctx context.Context, in usecaseGallery.ListMediaInput) ([]gallery.Media, pagination.Result, error)

	RegenerateThumbnail(ctx context.Context, id string) (*gallery.Media, error)

	GetMediaContent(ctx context.Context, id string) (*usecaseGallery.MediaBinary, error)
	GetPublicMediaContent(ctx context.Context, id string) (*usecaseGallery.MediaBinary, error)
	GetMediaThumbnail(ctx context.Context, id string) (*usecaseGallery.MediaBinary, error)
	GetPublicMediaThumbnail(ctx context.Context, id string) (*usecaseGallery.MediaBinary, error)
}

type GalleryHandler struct {
	galleryService    GalleryService
	galleryConfig     config.GalleryConfig
	signedURL         *usecaseGallery.SignedURLService
	signedURLsEnabled bool
}

// NewGalleryHandler wires the handler. signedURL may be nil — when
// SignedURLsEnabled is false the handler falls back to the legacy
// per-scope URL builders so flipping the flag is a no-op for older
// deployments. signedURLsEnabled=true with nil signedURL produces 503
// from the unified route (intentional misconfig surface).
func NewGalleryHandler(s GalleryService, cfg config.GalleryConfig, signedURL *usecaseGallery.SignedURLService, signedURLsEnabled bool) *GalleryHandler {
	return &GalleryHandler{
		galleryService:    s,
		galleryConfig:     cfg,
		signedURL:         signedURL,
		signedURLsEnabled: signedURLsEnabled,
	}
}

type pending struct {
	fileIdx int
	input   usecaseGallery.UploadMediaInput
}

// CreateFolderRequest represents the JSON body for creating a gallery folder.
type CreateFolderRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255" example:"Panen Raya 2026"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000" example:"Dokumentasi panen"`
	IsPublic    *bool  `json:"is_public,omitempty" example:"false"`
}

// UpdateFolderRequest represents the JSON body for updating a gallery folder.
type UpdateFolderRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255" example:"Panen Raya 2026"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000" example:"Dokumentasi panen"`
}

// FolderVisibilityRequest represents the JSON body for toggling a folder's visibility.
type FolderVisibilityRequest struct {
	IsPublic bool `json:"is_public" validate:"required" example:"true"`
}

// MediaVisibilityRequest represents the JSON body for toggling a media item's visibility.
type MediaVisibilityRequest struct {
	IsPublic bool `json:"is_public" validate:"required" example:"true"`
}

// BulkMediaVisibilityRequest represents the JSON body for bulk media visibility updates.
type BulkMediaVisibilityRequest struct {
	MediaIDs []string `json:"media_ids" validate:"required,min=1"`
	IsPublic bool     `json:"is_public" validate:"required" example:"true"`
}

// GalleryBinaryResponse describes the streamed binary payload returned by content and thumbnail endpoints.
type GalleryBinaryResponse struct {
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
}

// signedContentURL builds the unified /api/v1/media/{id}/content?jwt=...
// URL when signed URLs are enabled. Falls back to the legacy
// per-scope admin path when the flag is off so the admin gallery
// client keeps working without changes (Task 7.4+ gallery refactor).
func (h *GalleryHandler) signedContentURL(mediaID, sub string) string {
	if mediaID == "" {
		return ""
	}
	if h.signedURLsEnabled && h.signedURL != nil {
		return usecaseGallery.SignedURLPath("content", mediaID) +
			h.signedURL.SignedURLQuery(usecaseGallery.ScopeAdmin, mediaID, sub, 0)
	}
	return fmt.Sprintf(adminMediaContentPath, mediaID)
}

// signedThumbnailURL is the thumbnail counterpart of signedContentURL.
func (h *GalleryHandler) signedThumbnailURL(mediaID, sub string) string {
	if mediaID == "" {
		return ""
	}
	if h.signedURLsEnabled && h.signedURL != nil {
		return usecaseGallery.SignedURLPath("thumbnail", mediaID) +
			h.signedURL.SignedURLQuery(usecaseGallery.ScopeAdmin, mediaID, sub, 0)
	}
	return fmt.Sprintf(adminMediaThumbnailPath, mediaID)
}

// gallerySub is the placeholder sub used in admin signed URLs until
// 7.3 hardens the real user id. When the admin client runs through
// the unified route, scope=admin gates the route; tightening the sub
// to the actual user id happens in a follow-up so revoking a user
// can also revoke outstanding URLs.
const gallerySub = "user:admin"

const (
	publicMediaContentPath   = "/api/v1/public/gallery/media/%s/content"
	publicMediaThumbnailPath = "/api/v1/public/gallery/media/%s/thumbnail"
	adminMediaContentPath    = "/api/v1/gallery/media/%s/content"
	adminMediaThumbnailPath  = "/api/v1/gallery/media/%s/thumbnail"
)

func publicThumbnailURL(mediaID string) string {
	if mediaID == "" {
		return ""
	}
	return fmt.Sprintf(publicMediaThumbnailPath, mediaID)
}

func adminThumbnailURL(mediaID string) string {
	if mediaID == "" {
		return ""
	}
	return fmt.Sprintf(adminMediaThumbnailPath, mediaID)
}

func publicContentURL(mediaID string) string {
	if mediaID == "" {
		return ""
	}
	return fmt.Sprintf(publicMediaContentPath, mediaID)
}

func adminContentURL(mediaID string) string {
	if mediaID == "" {
		return ""
	}
	return fmt.Sprintf(adminMediaContentPath, mediaID)
}

func ptrStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

const iso8601 = "2006-01-02T15:04:05Z07:00"

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(iso8601)
}

const (
	errFolderNotFound      = "folder not found"
	errMediaNotFound       = "media not found"
	errDuplicateFolderName = "duplicate folder name"
	errInvalidMimeType     = "invalid mime type"
	errFileTooLarge        = "file too large"
	errTooManyFiles        = "too many files"
	errTotalSizeExceeded   = "total size exceeded"
	errInvalidFolderID     = "invalid folder id"
	errInvalidMediaID      = "invalid media id"
	errFolderNameInvalid   = "folder name invalid"
	errSystemFolderImmutable = "system folder is immutable"
)

func mapServiceError(err error, fallback int) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, errFolderNotFound):
		return http.StatusNotFound, "Folder not found"
	case strings.Contains(msg, errMediaNotFound):
		return http.StatusNotFound, "Media not found"
	case strings.Contains(msg, errDuplicateFolderName):
		return http.StatusConflict, "A folder with this name already exists"
	case strings.Contains(msg, errInvalidMimeType):
		return http.StatusUnsupportedMediaType, "Unsupported media type"
	case strings.Contains(msg, errFileTooLarge):
		return http.StatusRequestEntityTooLarge, "File size exceeds limit"
	case strings.Contains(msg, errTotalSizeExceeded):
		return http.StatusRequestEntityTooLarge, "Total upload size exceeds limit"
	case strings.Contains(msg, errTooManyFiles):
		return http.StatusBadRequest, "Too many files in upload"
	case strings.Contains(msg, errInvalidFolderID), strings.Contains(msg, errInvalidMediaID):
		return http.StatusBadRequest, "Invalid id"
	case strings.Contains(msg, errFolderNameInvalid):
		return http.StatusBadRequest, "Invalid folder name"
	case strings.Contains(msg, errSystemFolderImmutable):
		return http.StatusForbidden, "System folder is immutable"
	}
	return fallback, ""
}

func (h *GalleryHandler) requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	uid, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || uid == "" {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return "", false
	}
	return uid, true
}

// ListFoldersAdmin godoc
// @Summary      List gallery folders (admin)
// @Description  Query: q, is_public, page, limit. RBAC: gallery:read.
// @Tags         gallery
// @Produce      json
// @Param        q         query   string  false  "Search name/description"
// @Param        is_public query   boolean false  "Filter by public flag"
// @Param        page      query   int     false  "Page"
// @Param        limit     query   int     false  "Page size"
// @Success      200       {object} gallery.FolderListResponse
// @Failure      400       {object} response.ErrorResponse
// @Failure      401       {object} response.ErrorResponse
// @Failure      403       {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/folders [get]
func (h *GalleryHandler) ListFoldersAdmin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page     int     `in:"query=page;default=1"`
		Limit    int     `in:"query=limit;default=10"`
		Q        *string `in:"query=q"`
		IsPublic *bool   `in:"query=is_public"`
	}
	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	page, lim, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid pagination parameters")
		return
	}

	folders, pg, err := h.galleryService.ListFolders(r.Context(), usecaseGallery.ListFoldersInput{
		Query:    stringPtr(input.Q),
		IsPublic: input.IsPublic,
		Page:     page,
		Limit:    lim,
	})
	if err != nil {
		if status, msg := mapServiceError(err, http.StatusInternalServerError); status != http.StatusInternalServerError {
			response.Error(w, status, msg)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list folders", err)
		return
	}

	resp := map[string]interface{}{
		"folders":    h.toAdminFolderResponses(folders),
		"pagination": pg,
	}
	response.Success(w, http.StatusOK, resp)
}

// GetFolderAdmin godoc
// @Summary      Get gallery folder by ID (admin)
// @Description  Returns folder with all media. RBAC: gallery:read.
// @Tags         gallery
// @Produce      json
// @Param        id   path  string  true  "Folder UUID"
// @Success      200  {object} gallery.FolderDetailResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/folders/{id} [get]
func (h *GalleryHandler) GetFolderAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Folder ID is required")
		return
	}

	folder, err := h.galleryService.GetFolderByID(r.Context(), id)
	if err != nil {
		if status, _ := mapServiceError(err, http.StatusNotFound); status == http.StatusNotFound {
			response.Error(w, status, "Folder not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get folder", err)
		return
	}

	page, lim, _ := pagination.Paginate(1, 100)
	media, pg, err := h.galleryService.ListMediaByFolder(r.Context(), id, false, usecaseGallery.ListMediaInput{
		FolderID: id,
		Page:     page,
		Limit:    lim,
	})
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list folder media", err)
		return
	}

	resp := FolderDetailResponse{
		Folder:     h.toAdminFolderResponse(folder),
		Media:      h.toAdminMediaResponses(media),
		Pagination: pg,
	}
	response.Success(w, http.StatusOK, resp)
}

// CreateFolder godoc
// @Summary      Create gallery folder (admin)
// @Description  Body: name (required), description, is_public. Defaults to private. RBAC: gallery:write.
// @Tags         gallery
// @Accept       json
// @Produce      json
// @Param        request  body  gallery.CreateFolderRequest  true  "Folder payload"
// @Success      201  {object} gallery.FolderResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/folders [post]
func (h *GalleryHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	creator, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	var req struct {
		Payload struct {
			Name        string `json:"name" validate:"required,min=1,max=255"`
			Description string `json:"description" validate:"omitempty,max=1000"`
			IsPublic    *bool  `json:"is_public"`
		} `in:"body=json"`
	}
	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}
	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	folder, err := h.galleryService.CreateFolder(r.Context(), usecaseGallery.CreateFolderInput{
		Name:        strings.TrimSpace(req.Payload.Name),
		Description: req.Payload.Description,
		IsPublic:    req.Payload.IsPublic,
		CreatedBy:   creator,
	})
	if err != nil {
		if status, msg := mapServiceError(err, http.StatusInternalServerError); status != http.StatusInternalServerError {
			response.Error(w, status, msg)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to create folder", err)
		return
	}

	response.Success(w, http.StatusCreated, h.toAdminFolderResponse(folder))
}

// UpdateFolder godoc
// @Summary      Update gallery folder (admin)
// @Description  Update name (required) and/or description. RBAC: gallery:write.
// @Tags         gallery
// @Accept       json
// @Produce      json
// @Param        id       path   string  true  "Folder UUID"
// @Param        request  body  gallery.UpdateFolderRequest  true  "Folder update payload"
// @Success      200  {object} gallery.FolderResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/folders/{id} [put]
func (h *GalleryHandler) UpdateFolder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Folder ID is required")
		return
	}

	var req struct {
		Payload struct {
			Name        string `json:"name" validate:"required,min=1,max=255"`
			Description string `json:"description" validate:"omitempty,max=1000"`
		} `in:"body=json"`
	}
	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}
	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	folder, err := h.galleryService.UpdateFolder(r.Context(), id, usecaseGallery.UpdateFolderInput{
		Name:        strings.TrimSpace(req.Payload.Name),
		Description: req.Payload.Description,
	})
	if err != nil {
		if status, _ := mapServiceError(err, http.StatusInternalServerError); status == http.StatusNotFound {
			response.Error(w, status, "Folder not found")
			return
		}
		if status, msg := mapServiceError(err, http.StatusInternalServerError); status != http.StatusInternalServerError {
			response.Error(w, status, msg)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update folder", err)
		return
	}
	response.Success(w, http.StatusOK, h.toAdminFolderResponse(folder))
}

// UpdateFolderVisibility godoc
// @Summary      Toggle folder visibility (admin)
// @Description  Flips folder.is_public. Preserves all media flags. RBAC: gallery:write.
// @Tags         gallery
// @Accept       json
// @Produce      json
// @Param        id       path   string  true  "Folder UUID"
// @Param        request  body  gallery.FolderVisibilityRequest  true  "Folder visibility payload"
// @Success      200  {object} gallery.FolderResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/folders/{id}/visibility [patch]
func (h *GalleryHandler) UpdateFolderVisibility(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Folder ID is required")
		return
	}
	var req struct {
		Payload struct {
			IsPublic *bool `json:"is_public" validate:"required"`
		} `in:"body=json"`
	}
	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}
	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	folder, err := h.galleryService.UpdateFolderVisibility(r.Context(), id, *req.Payload.IsPublic)
	if err != nil {
		if status, _ := mapServiceError(err, http.StatusInternalServerError); status == http.StatusNotFound {
			response.Error(w, status, "Folder not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update folder visibility", err)
		return
	}
	response.Success(w, http.StatusOK, h.toAdminFolderResponse(folder))
}

// SetFolderCover godoc
// @Summary      Set manual folder cover (admin)
// @Description  Pins cover_media_id to the given media (must belong to the folder) and marks it manual; auto-recompute no longer clobbers it until the cover media is deleted or a new manual cover is set. RBAC: gallery:write.
// @Tags         gallery
// @Accept       json
// @Produce      json
// @Param        id   path  string  true  "Folder UUID"
// @Param        body body  object  true  "{\"media_id\": \"...\"}"
// @Success      200  {object} response.MessageResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/folders/{id}/cover [patch]
func (h *GalleryHandler) SetFolderCover(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Folder ID is required")
		return
	}
	var req struct {
		Payload struct {
			MediaID string `json:"media_id" validate:"required"`
		} `in:"body=json"`
	}
	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}
	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	folder, err := h.galleryService.SetFolderCover(r.Context(), id, req.Payload.MediaID)
	if err != nil {
		if status, _ := mapServiceError(err, http.StatusInternalServerError); status != http.StatusInternalServerError {
			response.Error(w, status, "Failed to set folder cover")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to set folder cover", err)
		return
	}
	response.Success(w, http.StatusOK, h.toAdminFolderResponse(folder))
}

// DeleteFolder godoc
// @Summary      Delete gallery folder (admin)
// @Description  Cascades to contained media + files. RBAC: gallery:write.
// @Tags         gallery
// @Produce      json
// @Param        id   path  string  true  "Folder UUID"
// @Success      200  {object} response.MessageResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/folders/{id} [delete]
func (h *GalleryHandler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Folder ID is required")
		return
	}
	if err := h.galleryService.DeleteFolder(r.Context(), id); err != nil {
		if status, _ := mapServiceError(err, http.StatusInternalServerError); status == http.StatusNotFound {
			response.Error(w, status, "Folder not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete folder", err)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "Folder deleted successfully"})
}

// ListMediaAdmin godoc
// @Summary      List gallery media (admin, flat)
// @Description  Query: folder_id, is_public, media_type, q, page, limit. RBAC: gallery:read.
// @Tags         gallery
// @Produce      json
// @Param        folder_id  query   string  false  "Filter by folder UUID"
// @Param        is_public  query   boolean false  "Filter by public flag"
// @Param        media_type query   string  false  "image|video"
// @Param        q          query   string  false  "Search filename"
// @Param        page       query   int     false  "Page"
// @Param        limit      query   int     false  "Page size"
// @Success      200        {object} gallery.MediaListResponse
// @Failure      400        {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/media [get]
func (h *GalleryHandler) ListMediaAdmin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page      int     `in:"query=page;default=1"`
		Limit     int     `in:"query=limit;default=10"`
		FolderID  *string `in:"query=folder_id"`
		IsPublic  *bool   `in:"query=is_public"`
		MediaType *string `in:"query=media_type"`
		Q         *string `in:"query=q"`
	}
	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	page, lim, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid pagination parameters")
		return
	}

	media, pg, err := h.galleryService.ListAllMedia(r.Context(), usecaseGallery.ListMediaInput{
		FolderID:  stringPtr(input.FolderID),
		IsPublic:  input.IsPublic,
		MediaType: stringPtr(input.MediaType),
		Query:     stringPtr(input.Q),
		Page:      page,
		Limit:     lim,
	})
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list media", err)
		return
	}
	resp := map[string]interface{}{
		"media":      h.toAdminMediaResponses(media),
		"pagination": pg,
	}
	response.Success(w, http.StatusOK, resp)
}

// GetMediaAdmin godoc
// @Summary      Get media by ID (admin)
// @Description  Metadata at any visibility. RBAC: gallery:read.
// @Tags         gallery
// @Produce      json
// @Param        id   path  string  true  "Media UUID"
// @Success      200  {object} gallery.MediaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/media/{id} [get]
func (h *GalleryHandler) GetMediaAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Media ID is required")
		return
	}
	m, err := h.galleryService.GetMediaByID(r.Context(), id)
	if err != nil {
		if status, _ := mapServiceError(err, http.StatusNotFound); status == http.StatusNotFound {
			response.Error(w, status, "Media not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get media", err)
		return
	}
	response.Success(w, http.StatusOK, h.toAdminMediaResponse(m))
}

// UploadMedia godoc
// @Summary      Bulk upload media into folder (admin)
// @Description  Multipart media[] files (also accepts 'media'). Per-file validation; partial failures reported in results. RBAC: gallery:write.
// @Tags         gallery
// @Accept       mpfd
// @Produce      json
// @Param        id    path     string  true   "Folder UUID"
// @Param        media formData file    true   "One or more media files"
// @Success      201   {object} gallery.BulkUploadResponse
// @Failure      400   {object} response.ErrorResponse
// @Failure      401   {object} response.ErrorResponse
// @Failure      403   {object} response.ErrorResponse
// @Failure      404   {object} response.ErrorResponse
// @Failure      413   {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/folders/{id}/media [post]
func (h *GalleryHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	folderID := chi.URLParam(r, "id")
	if folderID == "" {
		response.Error(w, http.StatusBadRequest, "Folder ID is required")
		return
	}
	uploader, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	maxFiles := h.galleryConfig.GetBulkUploadMaxFiles()
	maxTotal := int64(h.galleryConfig.GetBulkUploadMaxTotalMB()) << 20
	const multipartOverheadBytes = 1 << 20
	const maxMemoryCap int64 = 32 << 20
	memThreshold := maxTotal
	if memThreshold > maxMemoryCap {
		memThreshold = maxMemoryCap
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxTotal+multipartOverheadBytes)

	if err := r.ParseMultipartForm(memThreshold); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) || isTooLargeErr(err) {
			response.Error(w, http.StatusRequestEntityTooLarge, "Total upload size exceeds limit")
			return
		}
		response.Error(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}
	if r.MultipartForm == nil {
		response.Error(w, http.StatusBadRequest, "Multipart form is required")
		return
	}

	files := pickMediaFiles(r.MultipartForm.File)
	if len(files) == 0 {
		response.Error(w, http.StatusBadRequest, "No media files provided (use form field 'media' or 'media[]')")
		return
	}
	if len(files) > maxFiles {
		response.Error(w, http.StatusBadRequest,
			fmt.Sprintf("Too many files in upload: max %d", maxFiles))
		return
	}

	var totalBytes int64
	for _, fh := range files {
		totalBytes += fh.Size
	}
	if totalBytes > maxTotal {
		response.Error(w, http.StatusRequestEntityTooLarge, "Total upload size exceeds limit")
		return
	}

	results := make([]BulkUploadResult, len(files))
	pendingList := make([]pending, 0, len(files))

	for i, fh := range files {
		ct := fh.Header.Get("Content-Type")
		if !isAllowedGalleryMime(ct) {
			results[i] = BulkUploadResult{
				Filename: fh.Filename,
				Status:   "error",
				Error:    "unsupported media type",
				Code:     http.StatusUnsupportedMediaType,
			}
			continue
		}
		if fh.Size > maxFileSizeBytes(ct, h.galleryConfig) {
			results[i] = BulkUploadResult{
				Filename: fh.Filename,
				Status:   "error",
				Error:    "file size exceeds limit",
				Code:     http.StatusRequestEntityTooLarge,
			}
			continue
		}
		pendingList = append(pendingList, pending{
			fileIdx: i,
			input: usecaseGallery.UploadMediaInput{
				FolderID:     folderID,
				OriginalName: fh.Filename,
				MimeType:     ct,
				Size:         fh.Size,
				UploadedBy:   uploader,
			},
		})
	}

	var svcResults []usecaseGallery.UploadMediaResult
	if len(pendingList) > 0 {
		inputs := make([]usecaseGallery.UploadMediaInput, 0, len(pendingList))
		opened := make([]multipart.File, 0, len(pendingList))
		defer func() {
			for _, f := range opened {
				if f != nil {
					_ = f.Close()
				}
			}
		}()
		for i := range pendingList {
			p := &pendingList[i]
			f, err := files[p.fileIdx].Open()
			if err != nil {
				p.input.Content = errReader{err: err}
				inputs = append(inputs, p.input)
				continue
			}
			p.input.Content = f
			opened = append(opened, f)
			inputs = append(inputs, p.input)
		}
		svcResults, _ = h.galleryService.BulkCreateMedia(r.Context(), folderID, uploader, inputs)
	}

	uploaded := make([]MediaResponse, 0, len(pendingList))
	for i, p := range pendingList {
		var res usecaseGallery.UploadMediaResult
		if i < len(svcResults) {
			res = svcResults[i]
		}
		if res.Err != nil || res.Failed || res.Media == nil {
			code := http.StatusInternalServerError
			msg := "failed to create media"
			if res.Err != nil {
				if status, mapped := mapServiceError(res.Err, http.StatusInternalServerError); status != http.StatusInternalServerError {
					code = status
					msg = strings.ToLower(mapped)
				} else if handlerutil.IsDuplicateKeyError(res.Err) {
					code = http.StatusConflict
					msg = "duplicate"
				} else {
					msg = res.Err.Error()
				}
			}
			results[p.fileIdx] = BulkUploadResult{
				Filename: files[p.fileIdx].Filename,
				Status:   "error",
				Error:    msg,
				Code:     code,
			}
			continue
		}
		mediaResp := h.toAdminMediaResponse(res.Media)
		results[p.fileIdx] = BulkUploadResult{
			Filename: files[p.fileIdx].Filename,
			Status:   "ok",
			Media:    &mediaResp,
		}
		uploaded = append(uploaded, mediaResp)
	}

	resp := BulkUploadResponse{
		Uploaded: uploaded,
		Results:  results,
	}
	response.Success(w, http.StatusCreated, resp)
}

// UpdateMediaVisibility godoc
// @Summary      Toggle media visibility (admin)
// @Description  Flips media.is_public. RBAC: gallery:write.
// @Tags         gallery
// @Accept       json
// @Produce      json
// @Param        id       path   string  true  "Media UUID"
// @Param        request  body  gallery.MediaVisibilityRequest  true  "Media visibility payload"
// @Success      200  {object} gallery.MediaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/media/{id}/visibility [patch]
func (h *GalleryHandler) UpdateMediaVisibility(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Media ID is required")
		return
	}
	var req struct {
		Payload struct {
			IsPublic *bool `json:"is_public" validate:"required"`
		} `in:"body=json"`
	}
	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}
	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	m, err := h.galleryService.UpdateMediaVisibility(r.Context(), id, *req.Payload.IsPublic)
	if err != nil {
		if status, _ := mapServiceError(err, http.StatusNotFound); status == http.StatusNotFound {
			response.Error(w, status, "Media not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update media visibility", err)
		return
	}
	response.Success(w, http.StatusOK, h.toAdminMediaResponse(m))
}

// BulkUpdateMediaVisibility godoc
// @Summary      Bulk media visibility toggle (admin)
// @Description  Flips is_public for many media in one folder. RBAC: gallery:write.
// @Tags         gallery
// @Accept       json
// @Produce      json
// @Param        id       path   string  true  "Folder UUID"
// @Param        request  body  gallery.BulkMediaVisibilityRequest  true  "Bulk media visibility payload"
// @Success      200  {object} gallery.MediaListResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/folders/{id}/media/visibility [post]
func (h *GalleryHandler) BulkUpdateMediaVisibility(w http.ResponseWriter, r *http.Request) {
	folderID := chi.URLParam(r, "id")
	if folderID == "" {
		response.Error(w, http.StatusBadRequest, "Folder ID is required")
		return
	}
	var req struct {
		Payload struct {
			MediaIDs []string `json:"media_ids" validate:"required,min=1"`
			IsPublic *bool    `json:"is_public" validate:"required"`
		} `in:"body=json"`
	}
	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}
	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	if err := h.galleryService.BulkUpdateMediaVisibility(r.Context(), folderID, req.Payload.MediaIDs, *req.Payload.IsPublic); err != nil {
		if status, _ := mapServiceError(err, http.StatusInternalServerError); status == http.StatusNotFound {
			response.Error(w, status, "Folder not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update media visibility", err)
		return
	}

	page, lim, _ := pagination.Paginate(1, pagination.MaxLimit)
	media, pg, err := h.galleryService.ListMediaByFolder(r.Context(), folderID, false, usecaseGallery.ListMediaInput{
		FolderID: folderID,
		Page:     page,
		Limit:    lim,
	})
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to read folder media", err)
		return
	}
	resp := map[string]interface{}{
		"media":      h.toAdminMediaResponses(media),
		"pagination": pg,
	}
	response.Success(w, http.StatusOK, resp)
}

// DeleteMedia godoc
// @Summary      Delete media (admin)
// @Description  Removes media row + files. RBAC: gallery:write.
// @Tags         gallery
// @Produce      json
// @Param        id   path  string  true  "Media UUID"
// @Success      200  {object} response.MessageResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/media/{id} [delete]
func (h *GalleryHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Media ID is required")
		return
	}
	if err := h.galleryService.DeleteMedia(r.Context(), id); err != nil {
		if status, _ := mapServiceError(err, http.StatusNotFound); status == http.StatusNotFound {
			response.Error(w, status, "Media not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete media", err)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "Media deleted successfully"})
}

// RegenerateMediaThumbnail godoc
// @Summary      Regenerate media thumbnail (admin)
// @Description  Re-runs the thumbnail pipeline. RBAC: gallery:write.
// @Tags         gallery
// @Produce      json
// @Param        id   path  string  true  "Media UUID"
// @Success      200  {object} gallery.MediaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /gallery/media/{id}/regenerate-thumbnail [post]
func (h *GalleryHandler) RegenerateMediaThumbnail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Media ID is required")
		return
	}
	m, err := h.galleryService.RegenerateThumbnail(r.Context(), id)
	if err != nil {
		if status, _ := mapServiceError(err, http.StatusNotFound); status == http.StatusNotFound {
			response.Error(w, status, "Media not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to regenerate thumbnail", err)
		return
	}
	response.Success(w, http.StatusOK, h.toAdminMediaResponse(m))
}

type MediaResponse struct {
	ID               string   `json:"id"`
	FolderID         string   `json:"folder_id"`
	MediaType        string   `json:"media_type"`
	OriginalFilename string   `json:"original_filename"`
	MimeType         string   `json:"mime_type"`
	FileSize         int64    `json:"file_size"`
	Width            *int     `json:"width,omitempty"`
	Height           *int     `json:"height,omitempty"`
	DurationSeconds  *float64 `json:"duration_seconds,omitempty"`
	IsPublic         bool     `json:"is_public"`
	ThumbnailURL     *string  `json:"thumbnail_url,omitempty"`
	ThumbnailFailed  bool     `json:"thumbnail_failed"`
	ContentURL       *string  `json:"content_url,omitempty"`
	UploadedBy       string   `json:"uploaded_by,omitempty"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

type FolderResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsPublic          bool   `json:"is_public"`
	CoverMediaID      string `json:"cover_media_id,omitempty"`
	CoverThumbnailURL string `json:"cover_thumbnail_url,omitempty"`
	MediaCount        int    `json:"media_count,omitempty"`
	PublicMediaCount  int    `json:"public_media_count,omitempty"`
	CreatedBy         string `json:"created_by,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type FolderListResponse struct {
	Folders    []FolderResponse  `json:"folders"`
	Pagination pagination.Result `json:"pagination"`
}

type FolderDetailResponse struct {
	Folder     FolderResponse    `json:"folder"`
	Media      []MediaResponse   `json:"media"`
	Pagination pagination.Result `json:"pagination"`
}

type PublicFolderListResponse = FolderListResponse

type PublicFolderDetailResponse = FolderDetailResponse

type MediaListResponse struct {
	Media      []MediaResponse   `json:"media"`
	Pagination pagination.Result `json:"pagination"`
}

type BulkUploadResult struct {
	Filename string         `json:"filename"`
	Status   string         `json:"status"`
	Media    *MediaResponse `json:"media,omitempty"`
	Error    string         `json:"error,omitempty"`
	Code     int            `json:"code,omitempty"`
}

type BulkUploadResponse struct {
	Uploaded []MediaResponse    `json:"uploaded"`
	Results  []BulkUploadResult `json:"results"`
}

func coverMediaID(f *gallery.Folder) string {
	if f == nil || f.CoverMediaID == nil {
		return ""
	}
	return *f.CoverMediaID
}

func (h *GalleryHandler) toAdminFolderResponse(f *gallery.Folder) FolderResponse {
	if f == nil {
		return FolderResponse{}
	}
	cover := coverMediaID(f)
	return FolderResponse{
		ID:                f.ID,
		Name:              f.Name,
		Description:       f.Description,
		IsPublic:          f.IsPublic,
		CoverMediaID:      cover,
		CoverThumbnailURL: h.signedThumbnailURL(cover, gallerySub),
		MediaCount:        f.MediaCount,
		CreatedBy:         f.CreatedBy,
		CreatedAt:         formatTime(f.CreatedAt),
		UpdatedAt:         formatTime(f.UpdatedAt),
	}
}

func (h *GalleryHandler) toAdminFolderResponses(in []gallery.Folder) []FolderResponse {
	out := make([]FolderResponse, 0, len(in))
	for i := range in {
		out = append(out, h.toAdminFolderResponse(&in[i]))
	}
	return out
}

func toPublicFolderResponse(f *gallery.Folder) FolderResponse {
	if f == nil {
		return FolderResponse{}
	}
	cover := coverMediaID(f)
	return FolderResponse{
		ID:                f.ID,
		Name:              f.Name,
		Description:       f.Description,
		IsPublic:          true,
		CoverMediaID:      cover,
		CoverThumbnailURL: publicThumbnailURL(cover),
		CreatedAt:         formatTime(f.CreatedAt),
		UpdatedAt:         formatTime(f.UpdatedAt),
	}
}

func toPublicFolderResponses(in []gallery.Folder) []FolderResponse {
	out := make([]FolderResponse, 0, len(in))
	for i := range in {
		out = append(out, toPublicFolderResponse(&in[i]))
	}
	return out
}

func (h *GalleryHandler) thumbnailResponse(m *gallery.Media, scope string) *string {
	if m.ThumbnailURL == nil || *m.ThumbnailURL == "" || m.ThumbnailFailed {
		return nil
	}
	if scope == "public" {
		if h.signedURLsEnabled && h.signedURL != nil {
			return ptrStr(usecaseGallery.SignedURLPath("thumbnail", m.ID) +
				h.signedURL.SignedURLQuery(usecaseGallery.ScopePublic, m.ID, "anonymous", 0))
		}
	}
	return ptrStr(h.signedThumbnailURL(m.ID, gallerySub))
}

func (h *GalleryHandler) contentResponse(m *gallery.Media, scope string) *string {
	if m.ID == "" {
		return nil
	}
	if scope == "public" {
		if h.signedURLsEnabled && h.signedURL != nil {
			return ptrStr(usecaseGallery.SignedURLPath("content", m.ID) +
				h.signedURL.SignedURLQuery(usecaseGallery.ScopePublic, m.ID, "anonymous", 0))
		}
	}
	return ptrStr(h.signedContentURL(m.ID, gallerySub))
}

func (h *GalleryHandler) toAdminMediaResponse(m *gallery.Media) MediaResponse {
	if m == nil {
		return MediaResponse{}
	}
	return MediaResponse{
		ID:               m.ID,
		FolderID:         m.FolderID,
		MediaType:        string(m.MediaType),
		OriginalFilename: m.OriginalFilename,
		MimeType:         m.MimeType,
		FileSize:         m.FileSize,
		Width:            m.Width,
		Height:           m.Height,
		DurationSeconds:  m.DurationSeconds,
		IsPublic:         m.IsPublic,
		ThumbnailURL:     h.thumbnailResponse(m, "admin"),
		ThumbnailFailed:  m.ThumbnailFailed,
		ContentURL:       h.contentResponse(m, "admin"),
		UploadedBy:       m.UploadedBy,
		CreatedAt:        formatTime(m.CreatedAt),
		UpdatedAt:        formatTime(m.UpdatedAt),
	}
}

func (h *GalleryHandler) toAdminMediaResponses(in []gallery.Media) []MediaResponse {
	out := make([]MediaResponse, 0, len(in))
	for i := range in {
		out = append(out, h.toAdminMediaResponse(&in[i]))
	}
	return out
}

func (h *GalleryHandler) toPublicMediaResponse(m *gallery.Media) MediaResponse {
	if m == nil {
		return MediaResponse{}
	}
	return MediaResponse{
		ID:               m.ID,
		FolderID:         m.FolderID,
		MediaType:        string(m.MediaType),
		OriginalFilename: m.OriginalFilename,
		MimeType:         m.MimeType,
		FileSize:         m.FileSize,
		Width:            m.Width,
		Height:           m.Height,
		DurationSeconds:  m.DurationSeconds,
		IsPublic:         true,
		ThumbnailURL:     h.thumbnailResponse(m, "public"),
		ThumbnailFailed:  m.ThumbnailFailed,
		ContentURL:       h.contentResponse(m, "public"),
		CreatedAt:        formatTime(m.CreatedAt),
		UpdatedAt:        formatTime(m.UpdatedAt),
	}
}

func (h *GalleryHandler) toPublicMediaResponses(in []gallery.Media) []MediaResponse {
	out := make([]MediaResponse, 0, len(in))
	for i := range in {
		out = append(out, h.toPublicMediaResponse(&in[i]))
	}
	return out
}

func streamBinary(w http.ResponseWriter, r *http.Request, bin *usecaseGallery.MediaBinary, publicAccess bool) {
	if bin == nil || bin.FilePath == "" {
		response.Error(w, http.StatusNotFound, "Media not found")
		return
	}
	if strings.Contains(bin.FilePath, "..") {
		response.Error(w, http.StatusNotFound, "Media not found")
		return
	}
	info, err := os.Stat(bin.FilePath)
	if err != nil || info.IsDir() {
		response.Error(w, http.StatusNotFound, "Media not found")
		return
	}
	ct := bin.ContentType
	if ct == "" {
		ct = inferContentTypeFromExt(filepath.Ext(bin.FilePath))
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	if publicAccess {
		w.Header().Set("Cache-Control", "public, max-age=300")
	} else {
		w.Header().Set("Cache-Control", "private, max-age=300")
	}
	http.ServeFile(w, r, bin.FilePath)
}

func pickMediaFiles(form map[string][]*multipart.FileHeader) []*multipart.FileHeader {
	if v, ok := form["media[]"]; ok && len(v) > 0 {
		return v
	}
	return form["media"]
}

type errReader struct{ err error }

func (e errReader) Read([]byte) (int, error) { return 0, e.err }

func maxFileSizeBytes(contentType string, cfg config.GalleryConfig) int64 {
	if strings.HasPrefix(contentType, "video/") {
		return int64(cfg.GetVideoMaxSizeMB()) << 20
	}
	return int64(cfg.GetImageMaxSizeMB()) << 20
}

func isAllowedGalleryMime(contentType string) bool {
	switch strings.ToLower(contentType) {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return true
	case "video/mp4", "video/webm", "video/quicktime":
		return true
	}
	return false
}

func isTooLargeErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "request body too large") ||
		strings.Contains(msg, "http: request body too large")
}

func inferContentTypeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	}
	return "application/octet-stream"
}
