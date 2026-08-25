package ppid

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"

	"webdesa/api/interface/http/middleware"
	galleryUsecase "webdesa/api/usecase/gallery"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
)

// PPIDUploadHandler handles PPID document and thumbnail uploads. Uploads
// are written to the system/ppid gallery folder and return the admin
// binary URL plus the media_id the caller can attach via
// document_media_id / thumbnail_media_id on POST /ppid or PUT /ppid/{id}.
//
// Three endpoints are exposed:
//   - POST /ppid/upload-media     → upload a single document
//   - POST /ppid/upload-thumbnail → upload a single thumbnail image
//   - POST /ppid/upload           → upload document + thumbnail together
//     (the combined endpoint; the two singles are kept as aliases for
//     existing clients).
type PPIDUploadHandler struct {
	fileStore galleryUsecase.FileStore
	logger    middleware.Logger
}

// NewPPIDUploadHandler creates a new ppid upload handler.
func NewPPIDUploadHandler(fileStore galleryUsecase.FileStore, logger middleware.Logger) *PPIDUploadHandler {
	return &PPIDUploadHandler{
		fileStore: fileStore,
		logger:    logger,
	}
}

// PPIDUploadMediaResponse represents the upload response
type PPIDUploadMediaResponse struct {
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`
	Filename string `json:"filename"`
}

// PPIDUploadResponse is the combined response for POST /ppid/upload.
// Document is always present; thumbnail is present when a thumbnail file
// was supplied.
type PPIDUploadResponse struct {
	Document  *PPIDUploadMediaResponse `json:"document,omitempty"`
	Thumbnail *PPIDUploadMediaResponse `json:"thumbnail,omitempty"`
}

// UploadDocument godoc
// @Summary      Upload PPID document (admin)
// @Description  Multipart: file (PDF/DOC/DOCX/XLS/XLSX). Validates against the ppid document allowlist (max size, MIME type). Returns admin content URL plus the media_id to attach via document_media_id. RBAC: ppid:write.
// @Tags         ppid-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Document file"
// @Success      200  {object} ppid.PPIDUploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid/upload-media [post]
func (h *PPIDUploadHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	h.uploadMedia(w, r, h.fileStore.SaveDocument)
}

// UploadThumbnail godoc
// @Summary      Upload PPID thumbnail (admin)
// @Description  Multipart: file (image only). Validates against the ppid thumbnail allowlist (max size, MIME type). Returns admin content URL plus the media_id to attach via thumbnail_media_id. RBAC: ppid:write.
// @Tags         ppid-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Thumbnail image"
// @Success      200  {object} ppid.PPIDUploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid/upload-thumbnail [post]
func (h *PPIDUploadHandler) UploadThumbnail(w http.ResponseWriter, r *http.Request) {
	h.uploadMedia(w, r, h.fileStore.SaveImage)
}

// Upload godoc
// @Summary      Upload PPID document + thumbnail (admin)
// @Description  Multipart: document (PDF/DOC/DOCX/XLS/XLSX) and optional thumbnail (image). Validates each file against its own allowlist. Returns both admin URLs + media_ids to attach via document_media_id / thumbnail_media_id. At least one of document/thumbnail is required. RBAC: ppid:write.
// @Tags         ppid-upload
// @Accept       mpfd
// @Produce      json
// @Param        document    formData file  false  "Document file (PDF/DOC/DOCX/XLS/XLSX)"
// @Param        thumbnail   formData file  false  "Thumbnail image"
// @Success      200  {object} ppid.PPIDUploadResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid/upload [post]
func (h *PPIDUploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		h.logger.Error(ctx, "failed to parse multipart form", "error", err.Error())
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	resp := &PPIDUploadResponse{}

	docFile, docHeader, docErr := r.FormFile("document")
	if docErr == nil {
		saved, err := h.saveOne(ctx, h.fileStore.SaveDocument, docFile, docHeader)
		docFile.Close()
		if err != nil {
			h.logger.Error(ctx, "failed to save document", "error", err.Error())
			response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to upload document", err)
			return
		}
		resp.Document = h.toResponse(saved, docHeader.Filename)
	} else if docErr != http.ErrMissingFile {
		h.logger.Error(ctx, "failed to get document from form", "error", docErr.Error())
		response.Error(w, http.StatusBadRequest, "Failed to read document file")
		return
	}

	thumbFile, thumbHeader, thumbErr := r.FormFile("thumbnail")
	if thumbErr == nil {
		saved, err := h.saveOne(ctx, h.fileStore.SaveImage, thumbFile, thumbHeader)
		thumbFile.Close()
		if err != nil {
			h.logger.Error(ctx, "failed to save thumbnail", "error", err.Error())
			response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to upload thumbnail", err)
			return
		}
		resp.Thumbnail = h.toResponse(saved, thumbHeader.Filename)
	} else if thumbErr != http.ErrMissingFile {
		h.logger.Error(ctx, "failed to get thumbnail from form", "error", thumbErr.Error())
		response.Error(w, http.StatusBadRequest, "Failed to read thumbnail file")
		return
	}

	if resp.Document == nil && resp.Thumbnail == nil {
		response.Error(w, http.StatusBadRequest, "At least one of document or thumbnail is required")
		return
	}

	response.Success(w, http.StatusOK, resp)
}

func (h *PPIDUploadHandler) uploadMedia(w http.ResponseWriter, r *http.Request, save func(ctx context.Context, feature string, in galleryUsecase.FileInput) (galleryUsecase.SavedFile, error)) {
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

	saved, err := h.saveOne(ctx, save, file, fileHeader)
	if err != nil {
		h.logger.Error(ctx, "failed to save upload", "error", err.Error())
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to upload media", err)
		return
	}

	h.logger.Info(ctx, "ppid media uploaded successfully", "media_id", saved.MediaID, "filename", fileHeader.Filename, "size", fileHeader.Size)

	response.Success(w, http.StatusOK, h.toResponse(saved, fileHeader.Filename))
}

// saveOne saves a single uploaded file via the given FileStore method and
// infers the Content-Type when the multipart header is empty or generic.
func (h *PPIDUploadHandler) saveOne(ctx context.Context, save func(ctx context.Context, feature string, in galleryUsecase.FileInput) (galleryUsecase.SavedFile, error), content io.Reader, fileHeader *multipart.FileHeader) (galleryUsecase.SavedFile, error) {
	return save(ctx, galleryUsecase.FeaturePPID, galleryUsecase.FileInput{
		OriginalName: fileHeader.Filename,
		Content:      content,
		Size:         fileHeader.Size,
		ContentType:  contentTypeFor(fileHeader),
	})
}

// toResponse converts a SavedFile into the API response shape.
func (h *PPIDUploadHandler) toResponse(saved galleryUsecase.SavedFile, filename string) *PPIDUploadMediaResponse {
	return &PPIDUploadMediaResponse{
		URL:      galleryUsecase.URLFor(galleryUsecase.URLScopeAdmin, "content", saved.MediaID),
		MediaID:  saved.MediaID,
		Filename: filename,
	}
}

// contentTypeFor returns the file's declared Content-Type, falling back to
// filename inference so the FileStore sees a valid MIME for empty or
// octet-stream headers.
func contentTypeFor(fileHeader *multipart.FileHeader) string {
	ct := fileHeader.Header.Get("Content-Type")
	if ct == "" || ct == "application/octet-stream" {
		ct = handlerutil.InferContentType(fileHeader.Filename)
	}
	return ct
}
