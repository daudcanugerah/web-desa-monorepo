package fasilitas

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"webdesa/api/pkg/response"
	"webdesa/api/usecase/fasilitascategory"

	"webdesa/api/pkg/handlerutil")

// FasilitasCategoryHandler handles HTTP requests for Fasilitas category management.
//
// Per the design, categories are a separate, first-class resource with their
// own create + delete endpoints. There is intentionally no Update handler;
// to rename, delete and re-create.
type FasilitasCategoryHandler struct {
	categoryService *fasilitascategory.Service
}

// NewFasilitasCategoryHandler creates a new Fasilitas category handler.
func NewFasilitasCategoryHandler(categoryService *fasilitascategory.Service) *FasilitasCategoryHandler {
	return &FasilitasCategoryHandler{categoryService: categoryService}
}

// FasilitasCategoryResponse is the JSON shape returned for a single category.
type FasilitasCategoryResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UsageCount int    `json:"usage_count"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// FasilitasCategoryListResponse represents a paginated list of fasilitas categories.
type FasilitasCategoryListResponse struct {
	FasilitasCategories []FasilitasCategoryResponse `json:"categories"`
	Pagination          map[string]interface{}      `json:"pagination"`
}

// ListFasilitasCategories godoc
// @Summary      List fasilitas categories (public)
// @Description  With q search for autocomplete.
// @Tags         fasilitas-categories
// @Produce      json
//		@Param	q	query	string	false	"Search"
//		@Param	page	query	int	false	"Page"
//		@Param	limit	query	int	false	"Page size"
// @Success      200  {object} fasilitas.FasilitasCategoryListResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/fasilitas/categories [get]
func (h *FasilitasCategoryHandler) ListFasilitasCategories(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Query *string `in:"query=q"`
		Page  int     `in:"query=page"`
		Limit int     `in:"query=limit"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse query parameters")
		return
	}

	if input.Query != nil {
		trimmed := strings.TrimSpace(*input.Query)
		if trimmed == "" {
			input.Query = nil
		} else {
			input.Query = &trimmed
		}
	}

	categories, paginationResult, err := h.categoryService.List(r.Context(), input.Query, input.Page, input.Limit)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list categories", err)
		return
	}

	resp := make([]FasilitasCategoryResponse, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, FasilitasCategoryResponse{
			ID:         c.ID,
			Name:       c.Name,
			UsageCount: c.UsageCount,
			CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"categories": resp,
		"pagination": paginationResult,
	})
}

// CreateFasilitasCategory godoc
// @Summary      Create fasilitas category (admin)
// @Description  RBAC: fasilitas:write.
// @Tags         fasilitas-categories
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "{name}"
// @Success      201  {object} fasilitas.FasilitasCategoryResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /fasilitas/categories [post]
func (h *FasilitasCategoryHandler) CreateFasilitasCategory(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Payload struct {
			Name string `json:"name" validate:"required,min=1,max=100"`
		} `in:"body=json"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}

	if err := handlerutil.ValidateStruct(input.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	c, err := h.categoryService.Create(r.Context(), input.Payload.Name)
	if err != nil {
		if handlerutil.IsDuplicateKeyError(err) {
			response.Error(w, http.StatusConflict, "A category with this name already exists")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to create category", err)
		return
	}

	resp := FasilitasCategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	response.Success(w, http.StatusCreated, resp)
}

// DeleteFasilitasCategory godoc
// @Summary      Delete fasilitas category (admin)
// @Description  Returns 409 if category is in use. RBAC: fasilitas:write.
// @Tags         fasilitas-categories
// @Produce      json
// @Param        id         path     string true   "Category UUID"
// @Success      204
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /fasilitas/categories/{id} [delete]
func (h *FasilitasCategoryHandler) DeleteFasilitasCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Category ID is required")
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid category ID format")
		return
	}

	err := h.categoryService.Delete(r.Context(), id)
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var notFound *fasilitascategory.CategoryNotFoundError
	if errors.As(err, &notFound) {
		response.Error(w, http.StatusNotFound, "Category not found")
		return
	}

	var inUse *fasilitascategory.CategoryInUseError
	if errors.As(err, &inUse) {
		response.Error(w, http.StatusConflict, inUse.Error())
		return
	}

	response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete category", err)
}
