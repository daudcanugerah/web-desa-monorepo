package banner

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"webdesa/api/domain/bannercategory"
	bannercategorysvc "webdesa/api/usecase/bannercategory"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/pagination"
	"webdesa/api/pkg/response"

	"github.com/go-chi/chi/v5"
)

// BannerCategoryResponse is the JSON shape for a banner category.
type BannerCategoryResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UsageCount int    `json:"usage_count"`
	CreatedAt  string `json:"created_at"` // ISO 8601
	UpdatedAt  string `json:"updated_at"` // ISO 8601
}

// BannerCategoryListResponse represents a paginated list of banner categories.
type BannerCategoryListResponse struct {
	BannerCategories []BannerCategoryResponse `json:"banner_categories"`
	Pagination       map[string]interface{}   `json:"pagination"`
}

// BannerCategoryHandler serves /banner-category endpoints.
// It accepts the banner-category service as a dependency.
type BannerCategoryHandler struct {
	categoryService CategoryService
}

// CategoryService is the subset of the banner-category service this handler
// needs. Satisfied by *bannercategory.Service.
type CategoryService interface {
	Create(ctx context.Context, name string) (*bannercategory.Category, error)
	Delete(ctx context.Context, id string) error
	ListWithUsage(ctx context.Context, query string, page, limit int) ([]bannercategorysvc.ListItem, int, error)
	FindByID(ctx context.Context, id string) (*bannercategory.Category, error)
}

// NewBannerCategoryHandler creates a new handler.
func NewBannerCategoryHandler(svc CategoryService) *BannerCategoryHandler {
	return &BannerCategoryHandler{categoryService: svc}
}

// ListBannerCategories godoc
// @Summary      List banner categories (admin)
// @Description  Paginated list with usage counts and optional q search.
// @Tags         banner-categories
// @Produce      json
// @Param        q     query     string  false  "Search"
// @Param        page  query     int     false  "Page number"  default(1)
// @Param        limit query     int     false  "Page size"    default(10)
// @Success      200   {object}  banner.BannerCategoryListResponse
// @Failure      401   {object}  response.ErrorResponse
// @Failure      403   {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /banners/categories [get]
func (h *BannerCategoryHandler) ListBannerCategories(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	page := atoiOr(r.URL.Query().Get("page"), 1)
	limit := atoiOr(r.URL.Query().Get("limit"), 10)

	cats, total, err := h.categoryService.ListWithUsage(r.Context(), q, page, limit)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list categories", err)
		return
	}

	out := make([]BannerCategoryResponse, len(cats))
	for i, c := range cats {
		out[i] = BannerCategoryResponse{
			ID:         c.ID,
			Name:       c.Name,
			UsageCount: c.UsageCount,
			CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	resp := map[string]interface{}{
		"banner_categories": out,
		"pagination":        pagination.NewResult(page, limit, total),
	}
	response.Success(w, http.StatusOK, resp)
}

// CreateBannerCategory godoc
// @Summary      Create banner category (admin)
// @Description  RBAC: banners:write. 409 on duplicate name.
// @Tags         banner-categories
// @Accept       json
// @Produce      json
// @Param        request body      object{name=string}  true  "Category name"
// @Success      201   {object}  banner.BannerCategoryResponse
// @Failure      400   {object}  response.ErrorResponse
// @Failure      401   {object}  response.ErrorResponse
// @Failure      403   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /banners/categories [post]
func (h *BannerCategoryHandler) CreateBannerCategory(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name" validate:"required,min=1,max=100"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	c, err := h.categoryService.Create(r.Context(), input.Name)
	if err != nil {
	var dupErr *bannercategorysvc.DuplicateNameError
	if errors.As(err, &dupErr) {
		response.Error(w, http.StatusConflict, dupErr.Error())
		return
	}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to create category", err)
		return
	}

	resp := BannerCategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	response.Success(w, http.StatusCreated, resp)
}

// DeleteBannerCategory godoc
// @Summary      Delete banner category (admin)
// @Description  409 if category is in use by any banner.
// @Tags         banner-categories
// @Produce      json
// @Param        id  path  string  true  "Category UUID"
// @Success      204 {object}  response.MessageResponse
// @Failure      400 {object}  response.ErrorResponse
// @Failure      401 {object}  response.ErrorResponse
// @Failure      403 {object}  response.ErrorResponse
// @Failure      404 {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /banners/categories/{id} [delete]
func (h *BannerCategoryHandler) DeleteBannerCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	if err := h.categoryService.Delete(r.Context(), id); err != nil {
		var inUseErr *bannercategorysvc.CategoryInUseError
		var notFoundErr *bannercategorysvc.CategoryNotFoundError
		switch {
		case errors.As(err, &inUseErr):
			response.Error(w, http.StatusConflict, inUseErr.Error())
		case errors.As(err, &notFoundErr):
			response.Error(w, http.StatusNotFound, notFoundErr.Error())
		default:
			response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete category", err)
		}
		return
	}
	response.Success(w, http.StatusNoContent, nil)
}

func atoiOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}