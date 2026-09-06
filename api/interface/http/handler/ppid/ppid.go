package ppid

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"webdesa/api/interface/http/middleware"
	"webdesa/api/pkg/response"
	galleryuc "webdesa/api/usecase/gallery"
	"webdesa/api/usecase/ppid"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	"webdesa/api/pkg/handlerutil"
)

// PPIDHandler handles HTTP requests for PPID (public information disclosure) management operations.
// It accepts the PPID service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
//
// The uploadsDir argument is preserved for API back-compat with the existing
// wiring (cmd/serve, integration/setup_test). All file storage now goes
// through the gallery FileStore that the service owns.
type PPIDHandler struct {
	ppidService       *ppid.Service
	uploadsDir        string
	logger            middleware.Logger
	signedURL         *galleryuc.SignedURLService
	signedURLsEnabled bool
}

// NewPPIDHandler creates a new PPID handler with service dependency injected.
func NewPPIDHandler(ppidService *ppid.Service, uploadsDir string, logger middleware.Logger, signedURL *galleryuc.SignedURLService, signedURLsEnabled bool) *PPIDHandler {
	return &PPIDHandler{
		ppidService:       ppidService,
		uploadsDir:        uploadsDir,
		logger:            logger,
		signedURL:         signedURL,
		signedURLsEnabled: signedURLsEnabled,
	}
}

// formatTimePtr formats a time pointer to ISO 8601 string, returns nil if pointer is nil
func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format("2006-01-02T15:04:05Z07:00")
	return &formatted
}

// buildCategoryInfo assembles a response.CategoryInfo from raw id and name pointers.
// Returns nil when id is nil (no category linked).
func buildCategoryInfo(id *string, name *string) *response.CategoryInfo {
	if id == nil {
		return nil
	}
	n := ""
	if name != nil {
		n = *name
	}
	return &response.CategoryInfo{ID: *id, Name: n}
}

// PPIDResponse represents the response for PPID data
type PPIDResponse struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	CategoryID    *string                `json:"category_id,omitempty"`
	Category      *response.CategoryInfo `json:"category,omitempty"`
	Description   *string                `json:"description,omitempty"`
	Document      *response.MediaInfo    `json:"document,omitempty"`
	Thumbnail     *response.MediaInfo    `json:"thumbnail,omitempty"`
	PublicationAt *string                `json:"publication_at"` // ISO 8601 format
	CreatedAt     string                 `json:"created_at"`     // ISO 8601 format
	UpdatedAt     string                 `json:"updated_at"`     // ISO 8601 format
}

// PPIDResponsePublic represents the response for PPID data on public
// endpoints. Citizens browse metadata + thumbnail preview only — the
// document binary is delivered exclusively via the approval email's
// signed link (see feature-docs/ppid/concept.md).
type PPIDResponsePublic struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	CategoryID    *string                `json:"category_id,omitempty"`
	Category      *response.CategoryInfo `json:"category,omitempty"`
	Description   *string                `json:"description,omitempty"`
	Thumbnail     *response.MediaInfo    `json:"thumbnail,omitempty"`
	PublicationAt *string                `json:"publication_at"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// PPIDResponseTruncated represents the response for PPID data in list endpoints
// with descriptions truncated to approximately 100 words (500 characters)
type PPIDResponseTruncated struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	CategoryID    *string                `json:"category_id,omitempty"`
	Category      *response.CategoryInfo `json:"category,omitempty"`
	Description   *string                `json:"description"` // Truncated to ~100 words (500 chars) with "..." suffix, always included
	Thumbnail     *response.MediaInfo    `json:"thumbnail,omitempty"`
	PublicationAt *string                `json:"publication_at"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// PPIDResponsePrivate represents the response for PPID data in private/authenticated list endpoints
// without descriptions for performance
type PPIDResponsePrivate struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	CategoryID    *string                `json:"category_id,omitempty"`
	Category      *response.CategoryInfo `json:"category,omitempty"`
	PublicationAt *string                `json:"publication_at"`
	Document      *response.MediaInfo    `json:"document,omitempty"`
	Thumbnail     *response.MediaInfo    `json:"thumbnail,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// ppidMediaFor emits the shared media shape (Task 4.3) for a PPID
// document or thumbnail. PPID media stays private, so admin-scoped binary
// URLs are used. When signed URLs are enabled (Task 7.1.6), the URL
// embeds ?jwt= pointing at the unified /api/v1/media/{id}/...?jwt= route.
func (h *PPIDHandler) ppidMediaFor(mediaID *string) *response.MediaInfo {
	if mediaID == nil || *mediaID == "" {
		return nil
	}
	id := *mediaID
	if h.signedURLsEnabled && h.signedURL != nil {
		url := galleryuc.SignedURLPath("content", id) + h.signedURL.SignedURLQuery(galleryuc.ScopeAdmin, id, "user:admin", 0)
		return &response.MediaInfo{MediaID: id, URL: url}
	}
	url := galleryuc.URLFor(galleryuc.URLScopeAdmin, "content", id)
	return &response.MediaInfo{MediaID: id, URL: url}
}

// PPIDRequestResponse represents the response for PPID request data
type PPIDRequestResponse struct {
	ID             string  `json:"id"`
	PPIDId         string  `json:"ppid_id"`
	RequesterName  string  `json:"requester_name"`
	RequesterEmail string  `json:"requester_email"`
	Purpose        *string `json:"purpose,omitempty"`
	Status         string  `json:"status"` // pending, approved, revoked
	ApprovedAt     *string `json:"approved_at,omitempty"`
	ApprovedBy     *string `json:"approved_by,omitempty"`
	RevokedAt      *string `json:"revoked_at,omitempty"`
	RevokedBy      *string `json:"revoked_by,omitempty"`
	CreatedAt      string  `json:"created_at"` // ISO 8601 format
}

// PPIDPaginatedResponse represents a paginated list of PPID documents.
type PPIDPaginatedResponse struct {
	PPID       []PPIDResponse         `json:"ppid"`
	Pagination map[string]interface{} `json:"pagination"`
}

// PPIDRequestPaginatedResponse represents a paginated list of PPID requests.
type PPIDRequestPaginatedResponse struct {
	Requests   []PPIDRequestResponse  `json:"requests"`
	Pagination map[string]interface{} `json:"pagination"`
}

// CreatePPID godoc
// @Summary      Create PPID document (admin)
// @Description  Multipart. Required: title, document (PDF/DOC/DOCX/XLS/XLSX ≤50MB). Optional: category (UUID), description, publication_at (RFC3339), thumbnail. RBAC: ppid:write.
// @Tags         ppid
// @Accept       mpfd
// @Produce      json
// @Param        title      formData string true   "Title"
// @Param        document   formData file   true   "Document file"
// @Param        category   formData string  false  "Category UUID"
// @Param        description formData string  false  "Description"
// @Param        publication_at formData string  false  "RFC3339 datetime"
// @Param        thumbnail  formData file  false  "Thumbnail image"
// @Success      200  {object} ppid.PPIDResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid [post]
func (h *PPIDHandler) CreatePPID(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title            string       `in:"form=title" validate:"required,min=1"`
		Category         *string      `in:"form=category"`
		Description      *string      `in:"form=description"`
		PublicationAt    *string      `in:"form=publication_at"`
		Document         *httpin.File `in:"form=document"`
		Thumbnail        *httpin.File `in:"form=thumbnail"`
		DocumentMediaID  *string      `in:"form=document_media_id"`
		ThumbnailMediaID *string      `in:"form=thumbnail_media_id"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Validate required fields
	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	if input.Document == nil && (input.DocumentMediaID == nil || *input.DocumentMediaID == "") {
		response.Error(w, http.StatusBadRequest, "Document file or document_media_id is required")
		return
	}
	if input.Document != nil && input.DocumentMediaID != nil && *input.DocumentMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either document or document_media_id, not both")
		return
	}
	if input.Thumbnail != nil && input.ThumbnailMediaID != nil && *input.ThumbnailMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either thumbnail or thumbnail_media_id, not both")
		return
	}

	// Parse publication_at if provided
	var publicationAt *time.Time
	if input.PublicationAt != nil && *input.PublicationAt != "" {
		if t, err := time.Parse(time.RFC3339, *input.PublicationAt); err == nil {
			publicationAt = &t
		}
	}

	// Prepare service input
	serviceInput := ppid.CreatePPIDInput{
		Title:            input.Title,
		Category:         input.Category,
		Description:      input.Description,
		PublicationAt:    publicationAt,
		DocumentMediaID:  input.DocumentMediaID,
		ThumbnailMediaID: input.ThumbnailMediaID,
	}

	// Open document file (legacy inline path; pre-uploaded refs skip it)
	if input.Document != nil {
		file, err := input.Document.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read document file")
			return
		}
		defer file.Close()

		serviceInput.DocumentFile = file
		serviceInput.DocumentName = input.Document.Filename()
		serviceInput.DocumentSize = input.Document.Size()
		serviceInput.ContentType = input.Document.MIMEHeader().Get("Content-Type")
	}

	// Handle optional thumbnail
	if input.Thumbnail != nil {
		thumbFile, err := input.Thumbnail.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read thumbnail file")
			return
		}
		defer thumbFile.Close()

		serviceInput.ThumbnailFile = thumbFile
		serviceInput.ThumbnailName = input.Thumbnail.Filename()
		serviceInput.ThumbnailSize = input.Thumbnail.Size()
		serviceInput.ThumbnailType = input.Thumbnail.MIMEHeader().Get("Content-Type")
	}

	p, err := h.ppidService.Create(r.Context(), serviceInput)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to create PPID", err)
		return
	}

	// Build response
	resp := PPIDResponse{
		ID:            p.ID,
		Title:         p.Title,
		CategoryID:    p.Category,
		Category:      buildCategoryInfo(p.Category, p.CategoryName),
		Document:      h.ppidMediaFor(p.DocumentMediaID),
		Thumbnail:     h.ppidMediaFor(p.ThumbnailMediaID),
		Description:   p.Description,
		PublicationAt: formatTimePtr(p.PublicationAt),
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListPPID godoc
// @Summary      List PPID documents (admin)
// @Description  Description omitted from list response. RBAC: ppid:read.
// @Tags         ppid
// @Produce      json
// @Param        q          query    string  false  "Search title"
// @Param        category   query    string  false  "Category UUID"
// @Param        page       query    int  false  "Page"
// @Param        limit      query    int  false  "Page size"
// @Success      200  {object} ppid.PPIDPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid [get]
func (h *PPIDHandler) ListPPID(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page     int     `in:"query=page;default=1" validate:"min=1"`
		Limit    int     `in:"query=limit;default=10" validate:"min=1"`
		Category *string `in:"query=category"`
		Q        *string `in:"query=q"`
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

	// Call PPID service
	serviceInput := ppid.ListPPIDInput{
		Category: input.Category,
		Query:    input.Q,
		Page:     input.Page,
		Limit:    input.Limit,
	}

	ppidList, paginationResult, err := h.ppidService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list PPID", err)
		return
	}

	// Build response
	ppidResponses := make([]PPIDResponsePrivate, len(ppidList))
	for i, p := range ppidList {
		ppidResponses[i] = PPIDResponsePrivate{
			ID:            p.ID,
			Title:         p.Title,
			CategoryID:    p.Category,
			Category:      buildCategoryInfo(p.Category, p.CategoryName),
			Document:      h.ppidMediaFor(p.DocumentMediaID),
			Thumbnail:     h.ppidMediaFor(p.ThumbnailMediaID),
			PublicationAt: formatTimePtr(p.PublicationAt),
			CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"ppid":       ppidResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetPPID godoc
// @Summary      Get PPID by ID (admin)
// @Description  RBAC: ppid:read.
// @Tags         ppid
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Success      200      {object}  ppid.PPIDResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid/{id} [get]
// GetPPID handles GET /api/v1/ppid/:id
// Retrieves a PPID document by ID (public endpoint).
//
// Validates: Requirements 12.4, 16.1, 16.2
func (h *PPIDHandler) GetPPID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "PPID ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid PPID ID format")
		return
	}

	// Call PPID service
	p, err := h.ppidService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "PPID not found")
		return
	}

	// Build response
	resp := PPIDResponse{
		ID:            p.ID,
		Title:         p.Title,
		CategoryID:    p.Category,
		Category:      buildCategoryInfo(p.Category, p.CategoryName),
		Document:      h.ppidMediaFor(p.DocumentMediaID),
		Thumbnail:     h.ppidMediaFor(p.ThumbnailMediaID),
		Description:   p.Description,
		PublicationAt: formatTimePtr(p.PublicationAt),
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdatePPID godoc
// @Summary      Update PPID document (admin)
// @Description  Required: title. Optional document/thumbnail. RBAC: ppid:write.
// @Tags         ppid
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Param        request  body      object  false  "JSON payload"
// @Success      200      {object}  ppid.PPIDResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid/{id} [put]
// UpdatePPID handles PUT /api/v1/ppid/:id
// Updates an existing PPID document with optional document file update.
//
// Validates: Requirements 12.6, 16.1, 16.2
func (h *PPIDHandler) UpdatePPID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "PPID ID is required")
		return
	}

	var input struct {
		Title            string       `in:"form=title" validate:"required,min=1"`
		Category         *string      `in:"form=category"`
		Description      *string      `in:"form=description"`
		PublicationAt    *string      `in:"form=publication_at"`
		Document         *httpin.File `in:"form=document"`
		Thumbnail        *httpin.File `in:"form=thumbnail"`
		DocumentMediaID  *string      `in:"form=document_media_id"`
		ThumbnailMediaID *string      `in:"form=thumbnail_media_id"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Validate required fields
	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	if input.Document != nil && input.DocumentMediaID != nil && *input.DocumentMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either document or document_media_id, not both")
		return
	}
	if input.Thumbnail != nil && input.ThumbnailMediaID != nil && *input.ThumbnailMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either thumbnail or thumbnail_media_id, not both")
		return
	}

	// Parse publication_at if provided
	var publicationAt *time.Time
	var clearPublicationAt bool
	if input.PublicationAt != nil && *input.PublicationAt != "" {
		if t, err := time.Parse(time.RFC3339, *input.PublicationAt); err == nil {
			publicationAt = &t
		}
	} else if input.PublicationAt != nil {
		// Field is present but empty - clear it
		clearPublicationAt = true
	}

	// Prepare service input
	serviceInput := ppid.UpdatePPIDInput{
		Title:              input.Title,
		Category:           input.Category,
		Description:        input.Description,
		PublicationAt:      publicationAt,
		ClearPublicationAt: clearPublicationAt,
		DocumentMediaID:    input.DocumentMediaID,
		ThumbnailMediaID:   input.ThumbnailMediaID,
	}

	// Handle document file (optional for update)
	if input.Document != nil {
		file, err := input.Document.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read document file")
			return
		}
		defer file.Close()

		serviceInput.DocumentFile = file
		serviceInput.DocumentName = input.Document.Filename()
		serviceInput.DocumentSize = input.Document.Size()
		serviceInput.ContentType = input.Document.MIMEHeader().Get("Content-Type")
	}

	// Handle thumbnail file (optional for update)
	if input.Thumbnail != nil {
		thumbFile, err := input.Thumbnail.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read thumbnail file")
			return
		}
		defer thumbFile.Close()

		serviceInput.ThumbnailFile = thumbFile
		serviceInput.ThumbnailName = input.Thumbnail.Filename()
		serviceInput.ThumbnailSize = input.Thumbnail.Size()
		serviceInput.ThumbnailType = input.Thumbnail.MIMEHeader().Get("Content-Type")
	}

	// Call PPID service
	p, err := h.ppidService.Update(r.Context(), id, serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "ppid not found") {
			response.Error(w, http.StatusNotFound, "PPID not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to update PPID", err)
		return
	}

	// Build response
	resp := PPIDResponse{
		ID:            p.ID,
		Title:         p.Title,
		CategoryID:    p.Category,
		Category:      buildCategoryInfo(p.Category, p.CategoryName),
		Document:      h.ppidMediaFor(p.DocumentMediaID),
		Thumbnail:     h.ppidMediaFor(p.ThumbnailMediaID),
		Description:   p.Description,
		PublicationAt: formatTimePtr(p.PublicationAt),
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// DeletePPID godoc
// @Summary      Delete PPID document (admin)
// @Description  Removes file. RBAC: ppid:write.
// @Tags         ppid
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Success      200      {object}  response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid/{id} [delete]
// DeletePPID handles DELETE /api/v1/ppid/:id
// Deletes a PPID document and its associated file.
//
// Validates: Requirements 12.7, 16.1, 16.2
func (h *PPIDHandler) DeletePPID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "PPID ID is required")
		return
	}

	// Call PPID service
	err := h.ppidService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "ppid not found") {
			response.Error(w, http.StatusNotFound, "PPID not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete PPID", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "PPID deleted successfully",
	})
}

// CreatePPIDRequest godoc
// @Summary      Request a PPID document (public)
// @Description  Public endpoint. Creates a request with status=pending.
// @Tags         ppid-requests
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Param        request  body      object  false  "JSON payload"
// @Success      200      {object}  ppid.PPIDRequestResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/ppid/{id}/requests [post]
// CreatePPIDRequest handles POST /api/v1/ppid/:id/requests
// Creates a new PPID document request (public endpoint).
//
// Validates: Requirements 12.8, 16.1, 16.2
func (h *PPIDHandler) CreatePPIDRequest(w http.ResponseWriter, r *http.Request) {
	ppidID := chi.URLParam(r, "id")
	if ppidID == "" {
		response.Error(w, http.StatusBadRequest, "PPID ID is required")
		return
	}

	var payload struct {
		RequesterName  string  `json:"requester_name" validate:"required,min=1"`
		RequesterEmail string  `json:"requester_email" validate:"required,email"`
		Purpose        *string `json:"purpose"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if err := handlerutil.ValidateStruct(payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call PPID service
	input := ppid.CreateRequestInput{
		PPIDId:         ppidID,
		RequesterName:  payload.RequesterName,
		RequesterEmail: payload.RequesterEmail,
		Purpose:        payload.Purpose,
	}

	request, err := h.ppidService.CreateRequest(r.Context(), input)
	if err != nil {
		if strings.Contains(err.Error(), "ppid not found") {
			response.Error(w, http.StatusNotFound, "PPID not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to create PPID request", err)
		return
	}

	// Build response
	resp := PPIDRequestResponse{
		ID:             request.ID,
		PPIDId:         request.PPIDId,
		RequesterName:  request.RequesterName,
		RequesterEmail: request.RequesterEmail,
		Purpose:        request.Purpose,
		Status:         request.Status,
		ApprovedAt:     formatTimePtr(request.ApprovedAt),
		ApprovedBy:     request.ApprovedBy,
		RevokedAt:      formatTimePtr(request.RevokedAt),
		RevokedBy:      request.RevokedBy,
		CreatedAt:      request.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListPPIDRequests godoc
// @Summary      List PPID requests (admin)
// @Description  Filter by status (repeatable). RBAC: ppid:read.
// @Tags         ppid-requests
// @Produce      json
// @Param        status     query    string  false  "pending|approved|revoked (repeatable)"
// @Param        page       query    int  false  "Page"
// @Param        limit      query    int  false  "Page size"
// @Success      200  {object} ppid.PPIDRequestPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid/requests [get]
func (h *PPIDHandler) ListPPIDRequests(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Statuses []string `in:"query=status"`
		Page     int      `in:"query=page;default=1" validate:"min=1"`
		Limit    int      `in:"query=limit;default=10" validate:"min=1"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	serviceInput := ppid.ListRequestsInput{
		Statuses: input.Statuses,
		Page:     input.Page,
		Limit:    input.Limit,
	}

	requests, paginationResult, err := h.ppidService.ListRequests(r.Context(), serviceInput)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list PPID requests", err)
		return
	}

	// Build response
	requestResponses := make([]PPIDRequestResponse, len(requests))
	for i, req := range requests {
		requestResponses[i] = PPIDRequestResponse{
			ID:             req.ID,
			PPIDId:         req.PPIDId,
			RequesterName:  req.RequesterName,
			RequesterEmail: req.RequesterEmail,
			Purpose:        req.Purpose,
			Status:         req.Status,
			ApprovedAt:     formatTimePtr(req.ApprovedAt),
			ApprovedBy:     req.ApprovedBy,
			RevokedAt:      formatTimePtr(req.RevokedAt),
			RevokedBy:      req.RevokedBy,
			CreatedAt:      req.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"requests":   requestResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// ListPPIDPublic godoc
// @Summary      List PPID (public)
// @Description  Description truncated to 500 chars.
// @Tags         ppid
// @Produce      json
// @Param        q          query    string  false  "Search"
// @Param        category   query    string  false  "Category UUID"
// @Param        page       query    int  false  "Page"
// @Param        limit      query    int  false  "Page size"
// @Success      200  {object} ppid.PPIDPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/ppid/list [get]
func truncateDescription(desc *string) *string {
	if desc == nil || *desc == "" {
		return desc
	}

	maxChars := 500
	if len(*desc) <= maxChars {
		return desc
	}

	truncated := (*desc)[:maxChars]
	// Find the last space to avoid cutting off in the middle of a word
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}
	truncated += "..."
	return &truncated
}

func (h *PPIDHandler) ListPPIDPublic(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page  int `in:"query=page;default=1"`
		Limit int `in:"query=limit;default=10"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if input.Page <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid page parameter: must be a positive integer")
		return
	}

	if input.Limit <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid limit parameter: must be a positive integer")
		return
	}

	serviceInput := ppid.ListPPIDInput{
		Page:  input.Page,
		Limit: input.Limit,
	}

	ppidList, paginationResult, err := h.ppidService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list PPID", err)
		return
	}

	ppidResponses := make([]PPIDResponseTruncated, len(ppidList))
	for i, p := range ppidList {
		ppidResponses[i] = PPIDResponseTruncated{
			ID:            p.ID,
			Title:         p.Title,
			CategoryID:    p.Category,
			Category:      buildCategoryInfo(p.Category, p.CategoryName),
			Thumbnail:     h.ppidMediaFor(p.ThumbnailMediaID),
			Description:   truncateDescription(p.Description),
			PublicationAt: formatTimePtr(p.PublicationAt),
			CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"ppid":       ppidResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetPPIDPublic godoc
// @Summary      Get PPID (public)
// @Description  Public full detail with description.
// @Tags         ppid
// @Produce      json
// @Param        id         path     string true   "PPID UUID"
// @Success      200  {object} ppid.PPIDResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/ppid/{id} [get]
func (h *PPIDHandler) GetPPIDPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "PPID ID is required")
		return
	}

	p, err := h.ppidService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "ppid not found") {
			response.Error(w, http.StatusNotFound, "PPID not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get PPID", err)
		return
	}

	resp := PPIDResponsePublic{
		ID:            p.ID,
		Title:         p.Title,
		CategoryID:    p.Category,
		Category:      buildCategoryInfo(p.Category, p.CategoryName),
		Thumbnail:     h.ppidMediaFor(p.ThumbnailMediaID),
		Description:   p.Description,
		PublicationAt: formatTimePtr(p.PublicationAt),
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// ApproveRequest godoc
// @Summary      Approve PPID request (admin)
// @Description  Sends approval email with signed JWT download link. Email sent BEFORE DB update. RBAC: ppid:write.
// @Tags         ppid-requests
// @Produce      json
// @Param        id         path     string true   "Request UUID"
// @Success      200  {object} response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid/requests/{id}/approve [post]
func (h *PPIDHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		response.Error(w, http.StatusBadRequest, "Request ID is required")
		return
	}

	// Extract authenticated user ID from context using the proper constant
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		response.Error(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Create new context for service call (without request context values)
	ctx := context.Background()

	// Call service to approve request
	_, err := h.ppidService.ApproveRequest(ctx, requestID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "ppid request not found") {
			response.Error(w, http.StatusNotFound, "Request not found")
			return
		}
		if strings.Contains(err.Error(), "request already approved") {
			response.Error(w, http.StatusBadRequest, "Request already approved")
			return
		}
		if strings.Contains(err.Error(), "cannot approve a revoked request") {
			response.Error(w, http.StatusBadRequest, "Cannot approve a revoked request")
			return
		}
		if strings.Contains(err.Error(), "failed to send approval email") {
			response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to send approval email", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to approve request", err)
		return
	}

	// Build response with status and message only
	resp := map[string]interface{}{
		"status":  "approved",
		"message": "Request approved. Download link has been sent to requester's email.",
	}

	response.Success(w, http.StatusOK, resp)
}

// RevokeRequest godoc
// @Summary      Revoke PPID request (admin)
// @Description  Marks the request as revoked. RBAC: ppid:write.
// @Tags         ppid-requests
// @Produce      json
// @Param        id         path     string true   "Request UUID"
// @Success      200  {object} ppid.PPIDRequestResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /ppid/requests/{id}/revoke [post]
func (h *PPIDHandler) RevokeRequest(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		response.Error(w, http.StatusBadRequest, "Request ID is required")
		return
	}

	// Extract authenticated user ID from context using the proper constant
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		response.Error(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Call service to revoke request
	req, err := h.ppidService.RevokeRequest(r.Context(), requestID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "ppid request not found") {
			response.Error(w, http.StatusNotFound, "Request not found")
			return
		}
		if strings.Contains(err.Error(), "request already revoked") {
			response.Error(w, http.StatusBadRequest, "Request already revoked")
			return
		}
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to revoke request", err)
		return
	}

	// Build response
	resp := map[string]interface{}{
		"id":              req.ID,
		"ppid_id":         req.PPIDId,
		"requester_name":  req.RequesterName,
		"requester_email": req.RequesterEmail,
		"purpose":         req.Purpose,
		"status":          req.Status,
		"created_at":      req.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"revoked_at":      req.RevokedAt.Format("2006-01-02T15:04:05Z07:00"),
		"revoked_by":      req.RevokedBy,
	}

	response.Success(w, http.StatusOK, resp)
}

// (Task 7.4: removed DownloadDocument and streamDocumentBinary.
// Citizens now download via the unified /api/v1/media/{id}/...?jwt=
// route, using the signed URL the server embeds in the approval email.)
