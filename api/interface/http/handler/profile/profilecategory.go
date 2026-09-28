package profile

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
	"webdesa/api/usecase/profilecategory"
)

// ProfileCategoryHandler handles HTTP requests for Profile category management.
//
// Categories are a separate, first-class resource with their own create +
// delete endpoints. There is intentionally no Update handler; to rename,
// delete and re-create.
type ProfileCategoryHandler struct {
	categoryService *profilecategory.Service
}

// NewProfileCategoryHandler creates a new Profile category handler.
func NewProfileCategoryHandler(categoryService *profilecategory.Service) *ProfileCategoryHandler {
	return &ProfileCategoryHandler{categoryService: categoryService}
}

// ProfileCategoryResponse is the JSON shape returned for a single category.
type ProfileCategoryResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SortOrder  int    `json:"sort_order"`
	UsageCount int    `json:"usage_count"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ProfileCategoryPaginatedResponse is a paginated list of categories.
type ProfileCategoryPaginatedResponse struct {
	Categories []ProfileCategoryResponse `json:"categories"`
	Pagination map[string]interface{}    `json:"pagination"`
}

// ListProfileCategories godoc
// @Summary      List profile categories (public)
// @Description  With q search for autocomplete.
// @Tags         profile-categories
// @Produce      json
// @Param        q          query    string        false   "Search"
// @Param        page       query    int  false  "Page"
// @Param        limit      query    int  false  "Page size"
// @Success      200  {object}  profile.ProfileCategoryPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/profile/categories [get]
func (h *ProfileCategoryHandler) ListProfileCategories(w http.ResponseWriter, r *http.Request) {
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

	resp := make([]ProfileCategoryResponse, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, ProfileCategoryResponse{
			ID:         c.ID,
			Name:       c.Name,
			SortOrder:  c.SortOrder,
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

// CreateProfileCategory godoc
// @Summary      Create profile category (admin)
// @Description  RBAC: profile:write.
// @Tags         profile-categories
// @Accept       json
// @Produce      json
// @Success      201  {object}  profile.ProfileCategoryResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /profile/categories [post]
func (h *ProfileCategoryHandler) CreateProfileCategory(w http.ResponseWriter, r *http.Request) {
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

	response.Success(w, http.StatusCreated, ProfileCategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		SortOrder: c.SortOrder,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteProfileCategory godoc
// @Summary      Delete profile category (admin)
// @Description  Returns 409 if category is in use. RBAC: profile:write.
// @Tags         profile-categories
// @Produce      json
// @Param        id         path     string      true   "Category UUID"
// @Success      204
// @Failure		400	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /profile/categories/{id} [delete]
func (h *ProfileCategoryHandler) DeleteProfileCategory(w http.ResponseWriter, r *http.Request) {
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

	var notFound *profilecategory.CategoryNotFoundError
	if errors.As(err, &notFound) {
		response.Error(w, http.StatusNotFound, "Category not found")
		return
	}

	var inUse *profilecategory.CategoryInUseError
	if errors.As(err, &inUse) {
		response.Error(w, http.StatusConflict, inUse.Error())
		return
	}

	response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete category", err)
}
