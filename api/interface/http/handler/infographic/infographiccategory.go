package infographic

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"webdesa/api/pkg/response"
	"webdesa/api/usecase/infographiccategory"

	"webdesa/api/pkg/handlerutil")

// InfographicCategoryHandler handles HTTP requests for Infographic category management.
//
// Per the design, categories are a separate, first-class resource with their
// own create + delete endpoints. There is intentionally no Update handler;
// to rename, delete and re-create.
type InfographicCategoryHandler struct {
	categoryService *infographiccategory.Service
}

// NewInfographicCategoryHandler creates a new Infographic category handler.
func NewInfographicCategoryHandler(categoryService *infographiccategory.Service) *InfographicCategoryHandler {
	return &InfographicCategoryHandler{categoryService: categoryService}
}

// InfographicCategoryResponse is the JSON shape returned for a single category.
type InfographicCategoryResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UsageCount int    `json:"usage_count"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// InfographicCategoryPaginatedResponse is a paginated list of categories
type InfographicCategoryPaginatedResponse struct {
	Categories []InfographicCategoryResponse `json:"categories"`
	Pagination map[string]interface{}       `json:"pagination"`
}

// ListInfographicCategories godoc
// @Summary      List infographic categories (public)
// @Description  With q search for autocomplete.
// @Tags         infographic-categories
// @Produce      json
// @Param        q          query    string        false   "Search"
// @Param        page       query    int  false  "Page"
// @Param        limit      query    int  false  "Page size"
// @Success      200  {object}  infographic.InfographicCategoryPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/infographic/categories [get]
func (h *InfographicCategoryHandler) ListInfographicCategories(w http.ResponseWriter, r *http.Request) {
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

	resp := make([]InfographicCategoryResponse, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, InfographicCategoryResponse{
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

// CreateInfographicCategory godoc
// @Summary      Create infographic category (admin)
// @Description  RBAC: infographic:write.
// @Tags         infographic-categories
// @Accept       json
// @Produce      json
// (body params removed - swag cannot resolve inline body types; use OpenAPI override if needed)
// @Success      201  {object}  infographic.InfographicCategoryResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic/categories [post]
func (h *InfographicCategoryHandler) CreateInfographicCategory(w http.ResponseWriter, r *http.Request) {
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

	resp := InfographicCategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	response.Success(w, http.StatusCreated, resp)
}

// DeleteInfographicCategory godoc
// @Summary      Delete infographic category (admin)
// @Description  Returns 409 if category is in use. RBAC: infographic:write.
// @Tags         infographic-categories
// @Produce      json
// @Param        id         path     string      true   "Category UUID"
// @Success      200  {object}  response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic/categories/{id} [delete]
func (h *InfographicCategoryHandler) DeleteInfographicCategory(w http.ResponseWriter, r *http.Request) {
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

	var notFound *infographiccategory.CategoryNotFoundError
	if errors.As(err, &notFound) {
		response.Error(w, http.StatusNotFound, "Category not found")
		return
	}

	var inUse *infographiccategory.CategoryInUseError
	if errors.As(err, &inUse) {
		response.Error(w, http.StatusConflict, inUse.Error())
		return
	}

	response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete category", err)
}
