package desa

import (
	"net/http"

	"webdesa/api/interface/http/middleware"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
	galleryuc "webdesa/api/usecase/gallery"
)

// DesaUploadHandler handles village-header (kepala desa) photo uploads. Each
// upload is written to the system/desa gallery folder and returns the admin
// content URL plus the media_id the front-end persists on the desa profile.
type DesaUploadHandler struct {
	fileStore galleryuc.FileStore
	logger    middleware.Logger
}

// NewDesaUploadHandler creates a new desa upload handler.
func NewDesaUploadHandler(fileStore galleryuc.FileStore, logger middleware.Logger) *DesaUploadHandler {
	return &DesaUploadHandler{fileStore: fileStore, logger: logger}
}

// DesaUploadMediaResponse represents the upload response.
type DesaUploadMediaResponse struct {
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`
	Filename string `json:"filename"`
}

// UploadMedia godoc
// @Summary      Upload kepala desa photo (admin)
// @Description  Multipart: file (image only). Validates against the desa feature allowlist (max size, MIME type). Returns admin content URL plus the media_id the front-end persists on the desa profile. RBAC: desa:write.
// @Tags         desa-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Media file"
// @Success      200  {object} desa.DesaUploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /desa/upload-media [post]
func (h *DesaUploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		h.logger.Error(ctx, "failed to parse multipart form", "error", err.Error())
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error(ctx, "failed to get file from form", "error", err.Error())
		response.Error(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = handlerutil.InferContentType(fileHeader.Filename)
	}

	saved, err := h.fileStore.SaveImage(ctx, galleryuc.FeatureDesa, galleryuc.FileInput{
		OriginalName: fileHeader.Filename,
		Content:      file,
		Size:         fileHeader.Size,
		ContentType:  contentType,
	})
	if err != nil {
		h.logger.Error(ctx, "failed to save upload", "error", err.Error())
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to upload media", err)
		return
	}

	url := galleryuc.URLFor(galleryuc.URLScopeAdmin, "content", saved.MediaID)

	h.logger.Info(ctx, "desa media uploaded successfully", "media_id", saved.MediaID, "filename", fileHeader.Filename, "size", fileHeader.Size)

	response.Success(w, http.StatusOK, DesaUploadMediaResponse{
		URL:      url,
		MediaID:  saved.MediaID,
		Filename: fileHeader.Filename,
	})
}
