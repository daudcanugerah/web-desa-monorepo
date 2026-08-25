package umkm

import (
	"net/http"

	"webdesa/api/interface/http/middleware"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
	galleryUsecase "webdesa/api/usecase/gallery"
)

// UMKMUploadHandler handles UMKM image uploads. Each call stores one file
// in the system/umkm gallery folder and returns the admin content URL plus
// the media_id the caller attaches via images_media_ids on POST /umkm or
// PUT /umkm/{id}. Repeatable: call once per image.
type UMKMUploadHandler struct {
	fileStore galleryUsecase.FileStore
	logger    middleware.Logger
}

// NewUMKMUploadHandler creates a new umkm upload handler.
func NewUMKMUploadHandler(fileStore galleryUsecase.FileStore, logger middleware.Logger) *UMKMUploadHandler {
	return &UMKMUploadHandler{
		fileStore: fileStore,
		logger:    logger,
	}
}

// UMKMUploadMediaResponse represents the upload response
type UMKMUploadMediaResponse struct {
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`
	Filename string `json:"filename"`
}

// UploadMedia godoc
// @Summary      Upload UMKM image (admin)
// @Description  Multipart: file (image only). Validates against the umkm image allowlist (max size, MIME type). Returns admin content URL plus the media_id to attach via images_media_ids. RBAC: umkm:write.
// @Tags         umkm-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Image file"
// @Success      200  {object} umkm.UMKMUploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /umkm/upload-media [post]
func (h *UMKMUploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
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

	saved, err := h.fileStore.SaveImage(ctx, galleryUsecase.FeatureUMKM, galleryUsecase.FileInput{
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

	url := galleryUsecase.URLFor(galleryUsecase.URLScopeAdmin, "content", saved.MediaID)

	h.logger.Info(ctx, "umkm media uploaded successfully", "media_id", saved.MediaID, "filename", fileHeader.Filename, "size", fileHeader.Size)

	response.Success(w, http.StatusOK, UMKMUploadMediaResponse{
		URL:      url,
		MediaID:  saved.MediaID,
		Filename: fileHeader.Filename,
	})
}
