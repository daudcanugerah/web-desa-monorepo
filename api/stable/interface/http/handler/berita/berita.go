package berita

import (
	"net/http"
	"strings"
	"time"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	"webdesa/api/pkg/response"
	"webdesa/api/usecase/berita"

	"webdesa/api/pkg/handlerutil"
)

// BeritaHandler handles HTTP requests for berita (news) management operations.
// It accepts the berita service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
type BeritaHandler struct {
	beritaService *berita.Service
}

// NewBeritaHandler creates a new berita handler with service dependency injected.
func NewBeritaHandler(beritaService *berita.Service) *BeritaHandler {
	return &BeritaHandler{
		beritaService: beritaService,
	}
}

// ptrStr returns a pointer to the given string.
func ptrStr(s string) *string {
	return &s
}

// buildCategoryInfo builds a response.CategoryInfo pointer from id and name pointers.
// Returns nil when id is nil (e.g. uncategorised record).
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

// BeritaResponse represents the response for berita data
type BeritaResponse struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	CategoryID *string                `json:"category_id,omitempty"`
	Category   *response.CategoryInfo `json:"category,omitempty"`
	ImageURL   *string                `json:"image_url,omitempty"`
	CreatedAt  string                 `json:"created_at"` // ISO 8601 format
	UpdatedAt  string                 `json:"updated_at"` // ISO 8601 format
}

// BeritaListResponse represents the response for berita list (without content)
type BeritaListResponse struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	CategoryID *string                `json:"category_id,omitempty"`
	Category   *response.CategoryInfo `json:"category,omitempty"`
	ImageURL   *string                `json:"image_url,omitempty"`
	CreatedAt  string                 `json:"created_at"` // ISO 8601 format
	UpdatedAt  string                 `json:"updated_at"` // ISO 8601 format
}

// BeritaPaginatedResponse represents a paginated list of berita
type BeritaPaginatedResponse struct {
	Berita     []BeritaListResponse   `json:"berita"`
	Pagination map[string]interface{} `json:"pagination"`
}

// CreateBerita godoc
// @Summary      Create news article (admin)
// @Description  Multipart: title (required), content (Quill delta JSON, required), category (UUID required), optional image. RBAC: berita:write.
// @Tags         berita
// @Accept       mpfd
// @Produce      json
// @Param        title      formData string true   "Title"
// @Param        content    formData string true   "Quill delta JSON"
// @Param        category   formData string true   "Category UUID"
// @Param        image	formData	file	false	"Cover image"
// @Success      201  {object} berita.BeritaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita [post]
func (h *BeritaHandler) CreateBerita(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string       `in:"form=title" validate:"required,min=1"`
		Content  string       `in:"form=content" validate:"required,min=1"`
		Category string       `in:"form=category" validate:"required,min=1"`
		Image    *httpin.File `in:"form=image"`
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

	// Get image file (optional)
	var beritaInput berita.CreateBeritaInput
	if input.Image != nil {
		imageFile, err := input.Image.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read image file")
			return
		}
		beritaInput = berita.CreateBeritaInput{
			Title:       input.Title,
			Content:     input.Content,
			Category:    input.Category,
			ImageFile:   imageFile,
			ImageName:   input.Image.Filename(),
			ImageSize:   input.Image.Size(),
			ContentType: input.Image.MIMEHeader().Get("Content-Type"),
		}
	} else {
		// No image provided
		beritaInput = berita.CreateBeritaInput{
			Title:     input.Title,
			Content:   input.Content,
			Category:  input.Category,
			ImageFile: nil,
		}
	}

	// Call berita service
	b, err := h.beritaService.Create(r.Context(), beritaInput)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to create berita", err)
		return
	}

	// Build response
	resp := BeritaResponse{
		ID:         b.ID,
		Title:      b.Title,
		Content:    b.Content,
		CategoryID: ptrStr(b.Category),
		Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
		ImageURL:   b.ImageURL,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListBerita godoc
// @Summary      List news (admin)
// @Description  Paginated list. Query: q, category (UUID), since, until (YYYY-MM-DD), sort (created_at|title), order (asc|desc). RBAC: berita:read.
// @Tags         berita
// @Produce      json
//
//	@Param	q	query	string	false	"Search query"
//	@Param	category	query	string	false	"Category UUID"
//	@Param	since	query	string	false	"Start date YYYY-MM-DD"
//	@Param	until	query	string	false	"End date YYYY-MM-DD"
//	@Param	sort	query	string	false	"Sort field: created_at or title (default created_at)"
//	@Param	order	query	string	false	"Sort direction: asc or desc (default desc - newest first)"
//	@Param	page	query	int	false	"Page default(1)"
//	@Param	limit	query	int	false	"Page size default(10)"
//
// @Success      200  {object} berita.BeritaPaginatedResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita [get]
func (h *BeritaHandler) ListBerita(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page     int     `in:"query=page;default=1" validate:"min=1"`
		Limit    int     `in:"query=limit;default=10" validate:"min=1"`
		Q        *string `in:"query=q"`
		Category *string `in:"query=category"`
		Since    *string `in:"query=since"`
		Until    *string `in:"query=until"`
		Sort     *string `in:"query=sort"`
		Order    *string `in:"query=order"`
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

	// Parse date range filters (optional)
	var since, until *time.Time
	if input.Since != nil {
		if t, err := time.Parse("2006-01-02", *input.Since); err == nil {
			since = &t
		}
	}

	if input.Until != nil {
		if t, err := time.Parse("2006-01-02", *input.Until); err == nil {
			until = &t
		}
	}

	// Call berita service
	serviceInput := berita.ListBeritaInput{
		Query:    input.Q,
		Category: input.Category,
		Since:    since,
		Until:    until,
		Sort:     "created_at", // default
		Order:    "desc",      // default
		Page:     input.Page,
		Limit:    input.Limit,
	}
	if input.Sort != nil && *input.Sort != "" {
		serviceInput.Sort = *input.Sort
	}
	if input.Order != nil && *input.Order != "" {
		serviceInput.Order = *input.Order
	}

	beritaList, paginationResult, err := h.beritaService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list berita", err)
		return
	}

	// Build response
	beritaResponses := make([]BeritaListResponse, len(beritaList))
	for i, b := range beritaList {
		beritaResponses[i] = BeritaListResponse{
			ID:         b.ID,
			Title:      b.Title,
			CategoryID: ptrStr(b.Category),
			Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
			ImageURL:   b.ImageURL,
			CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"berita":     beritaResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetBerita godoc
// @Summary      Get news by ID (admin)
// @Description  RBAC: berita:read.
// @Tags         berita
// @Produce      json
// @Param        id         path     string true   "Berita UUID"
// @Success      200  {object} berita.BeritaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita/{id} [get]
func (h *BeritaHandler) GetBerita(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Berita ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid berita ID format")
		return
	}

	// Call berita service
	b, err := h.beritaService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Berita not found")
		return
	}

	// Build response
	resp := BeritaResponse{
		ID:         b.ID,
		Title:      b.Title,
		Content:    b.Content,
		CategoryID: ptrStr(b.Category),
		Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
		ImageURL:   b.ImageURL,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateBerita godoc
// @Summary      Update news (admin)
// @Description  Partial update — missing fields fall back to existing values. RBAC: berita:write.
// @Tags         berita
// @Accept       mpfd
// @Produce      json
// @Param        id         path     string true   "Berita UUID"
// @Success      200  {object} berita.BeritaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita/{id} [put]
func (h *BeritaHandler) UpdateBerita(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Berita ID is required")
		return
	}

	// Check if berita exists first (before validating input)
	existing, err := h.beritaService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Berita not found")
		return
	}

	var input struct {
		Title    string       `in:"form=title"`
		Content  string       `in:"form=content"`
		Category string       `in:"form=category"`
		Image    *httpin.File `in:"form=image"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get form fields - use existing values as defaults
	title := input.Title
	if title == "" {
		title = existing.Title
	}

	content := input.Content
	if content == "" {
		content = existing.Content
	}

	category := input.Category
	if category == "" {
		category = existing.Category
	}

	// Get image file (optional for update)
	var beritaInput berita.UpdateBeritaInput
	if input.Image != nil {
		imageFile, err := input.Image.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read image file")
			return
		}
		beritaInput = berita.UpdateBeritaInput{
			Title:       title,
			Content:     content,
			Category:    category,
			ImageFile:   imageFile,
			ImageName:   input.Image.Filename(),
			ImageSize:   input.Image.Size(),
			ContentType: input.Image.MIMEHeader().Get("Content-Type"),
		}
	} else {
		// No new image provided
		beritaInput = berita.UpdateBeritaInput{
			Title:     title,
			Content:   content,
			Category:  category,
			ImageFile: nil,
		}
	}

	// Call berita service
	b, err := h.beritaService.Update(r.Context(), id, beritaInput)
	if err != nil {
		if strings.Contains(err.Error(), "berita not found") {
			response.Error(w, http.StatusNotFound, "Berita not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update berita", err)
		return
	}

	// Build response
	resp := BeritaResponse{
		ID:         b.ID,
		Title:      b.Title,
		Content:    b.Content,
		CategoryID: ptrStr(b.Category),
		Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
		ImageURL:   b.ImageURL,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// DeleteBerita godoc
// @Summary      Delete news (admin)
// @Description  Deletes article and its cover image. RBAC: berita:write.
// @Tags         berita
// @Produce      json
// @Param        id         path     string true   "Berita UUID"
// @Success      200  {object} response.MessageResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita/{id} [delete]
func (h *BeritaHandler) DeleteBerita(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Berita ID is required")
		return
	}

	// Call berita service
	err := h.beritaService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "berita not found") {
			response.Error(w, http.StatusNotFound, "Berita not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete berita", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Berita deleted successfully",
	})
}

// ListBeritaPublic godoc
// @Summary      List news (public)
// @Description  Public paginated list without content body.
// @Tags         berita
// @Produce      json
//
//	@Param	page	query	int	false	"Page default(1)"
//	@Param	limit	query	int	false	"Page size default(10)"
//	@Param	sort	query	string	false	"Sort field: created_at or title (default created_at)"
//	@Param	order	query	string	false	"Sort direction: asc or desc (default desc - newest first)"
//
// @Success      200  {object} berita.BeritaPaginatedResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Router       /public/berita/list [get]
func (h *BeritaHandler) ListBeritaPublic(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page  int     `in:"query=page;default=1"`
		Limit int     `in:"query=limit;default=10"`
		Sort  *string `in:"query=sort"`
		Order *string `in:"query=order"`
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

	serviceInput := berita.ListBeritaInput{
		Page:  input.Page,
		Limit: input.Limit,
		Sort:  "created_at",
		Order: "desc",
	}
	if input.Sort != nil && *input.Sort != "" {
		serviceInput.Sort = *input.Sort
	}
	if input.Order != nil && *input.Order != "" {
		serviceInput.Order = *input.Order
	}

	beritaList, paginationResult, err := h.beritaService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list berita", err)
		return
	}

	beritaResponses := make([]BeritaListResponse, len(beritaList))
	for i, b := range beritaList {
		beritaResponses[i] = BeritaListResponse{
			ID:         b.ID,
			Title:      b.Title,
			CategoryID: ptrStr(b.Category),
			Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
			ImageURL:   b.ImageURL,
			CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"berita":     beritaResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetBeritaPublic godoc
// @Summary      Get news (public)
// @Description  Public full detail.
// @Tags         berita
// @Produce      json
// @Param        id         path     string true   "Berita UUID"
// @Success      200  {object} berita.BeritaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Router       /public/berita/{id} [get]
func (h *BeritaHandler) GetBeritaPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Berita ID is required")
		return
	}

	b, err := h.beritaService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "berita not found") {
			response.Error(w, http.StatusNotFound, "Berita not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get berita", err)
		return
	}

	resp := BeritaResponse{
		ID:         b.ID,
		Title:      b.Title,
		Content:    b.Content,
		CategoryID: ptrStr(b.Category),
		Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
		ImageURL:   b.ImageURL,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}
