package banner

import (
	"net/http"

	"webdesa/api/interface/http/middleware"
	galleryUsecase "webdesa/api/usecase/gallery"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
)

// BannerUploadHandler handles banner cover image uploads. Each upload is
// written to the system/banner gallery folder and returns the admin binary
// URL; public banner responses translate that to a feature-scoped public
// URL the anonymous site can stream.
type BannerUploadHandler struct {
	fileStore galleryUsecase.FileStore
	logger    middleware.Logger
}

// NewBannerUploadHandler creates a new banner upload handler.
func NewBannerUploadHandler(fileStore galleryUsecase.FileStore, logger middleware.Logger) *BannerUploadHandler {
	return &BannerUploadHandler{
		fileStore: fileStore,
		logger:    logger,
	}
}

// BannerUploadMediaResponse represents the upload response
type BannerUploadMediaResponse struct {
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`
	Filename string `json:"filename"`
}

// UploadMedia godoc
// @Summary      Upload banner cover media (admin)
// @Description  Multipart: file (image only). Validates against the banner feature allowlist (max size, MIME type). Returns admin content URL plus the media_id the front-end can persist alongside the banner. RBAC: banners:write.
// @Tags         banner-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Media file"
// @Success      200  {object} banner.BannerUploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /banners/upload-media [post]
func (h *BannerUploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
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

	// The multipart header is the source of truth; fall back to filename
	// inference so the FileStore sees a valid MIME for empty/octet-stream
	// headers.
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = handlerutil.InferContentType(fileHeader.Filename)
	}

	saved, err := h.fileStore.SaveImage(ctx, galleryUsecase.FeatureBanner, galleryUsecase.FileInput{
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

	h.logger.Info(ctx, "banner media uploaded successfully", "media_id", saved.MediaID, "filename", fileHeader.Filename, "size", fileHeader.Size)

	response.Success(w, http.StatusOK, BannerUploadMediaResponse{
		URL:      url,
		MediaID:  saved.MediaID,
		Filename: fileHeader.Filename,
	})
}
