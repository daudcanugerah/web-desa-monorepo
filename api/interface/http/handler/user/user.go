package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	userdomain "webdesa/api/domain/user"
	galleryuc "webdesa/api/usecase/gallery"
	"webdesa/api/interface/http/middleware"
	"webdesa/api/pkg/response"
	"webdesa/api/usecase/user"

	"webdesa/api/pkg/handlerutil")

// UserHandler handles HTTP requests for user management operations.
// It accepts the user service and role service from the usecase layer as dependencies.
// Dependencies point inward: interface/http → usecase → domain
//
// Task 7.1.6: signedURL fields mirror the banner handler.
type UserHandler struct {
	userService      *user.Service
	roleService      RoleService
	signedURL        *galleryuc.SignedURLService
	signedURLsEnabled bool
}

// RoleService defines the interface for role operations needed by UserHandler.
// Following Go best practices, this interface is defined where it's used.
type RoleService interface {
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

// NewUserHandler creates a new user handler with service dependencies injected.
func NewUserHandler(userService *user.Service, roleService RoleService, signedURL *galleryuc.SignedURLService, signedURLsEnabled bool) *UserHandler {
	return &UserHandler{
		userService:      userService,
		roleService:      roleService,
		signedURL:        signedURL,
		signedURLsEnabled: signedURLsEnabled,
	}
}

// containsError checks if an error message contains a specific substring.
// This is useful for handling wrapped errors.
func containsError(errMsg, substr string) bool {
	return strings.Contains(strings.ToLower(errMsg), strings.ToLower(substr))
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdatePasswordRequest represents the request body for updating password
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password"`
}

// UserResponse represents the response for user data
type UserResponse struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	ProfileMedia    *response.MediaInfo `json:"profile_media,omitempty"`
	Roles           []string `json:"roles"`
	CreatedAt       string   `json:"created_at"` // ISO 8601 format
	UpdatedAt       string   `json:"updated_at"` // ISO 8601 format
}

// resolveProfileImageURL returns the gallery binary URL for a user's
// profile image when a media id is present, falling back to the legacy
// string URL.
func resolveProfileImageURL(u *userdomain.User) *string {
	if u.ProfileImageMediaID != nil && *u.ProfileImageMediaID != "" {
		url := galleryuc.URLFor(galleryuc.URLScopeAdmin, "content", *u.ProfileImageMediaID)
		return &url
	}
	return nil
}

// profileMediaForUser emits the shared media shape (Task 4.3) for a user's
// profile image. When signed URLs are enabled (Task 7.1.6) the URLs
// embed ?jwt= pointing at /api/v1/media/{id}/...?jwt=.
func (h *UserHandler) profileMediaForUser(u *userdomain.User) *response.MediaInfo {
	if u.ProfileImageMediaID == nil || *u.ProfileImageMediaID == "" {
		return nil
	}
	id := *u.ProfileImageMediaID
	if h.signedURLsEnabled && h.signedURL != nil {
		url := galleryuc.SignedURLPath("content", id) + h.signedURL.SignedURLQuery(galleryuc.ScopeAdmin, id, "user:admin", 0)
		thumb := galleryuc.SignedURLPath("thumbnail", id) + h.signedURL.SignedURLQuery(galleryuc.ScopeAdmin, id, "user:admin", 0)
		return &response.MediaInfo{MediaID: id, URL: url, ThumbnailURL: &thumb}
	}
	url := galleryuc.URLFor(galleryuc.URLScopeAdmin, "content", id)
	thumb := galleryuc.URLFor(galleryuc.URLScopeAdmin, "thumbnail", id)
	return &response.MediaInfo{MediaID: id, URL: url, ThumbnailURL: &thumb}
}

// UserListResponse represents a paginated list of users.
type UserListResponse struct {
	Users      []UserResponse          `json:"users"`
	Pagination map[string]interface{}  `json:"pagination"`
}

// CreateUser godoc
// @Summary      Create a new user (admin)
// @Description  Create a new user account. Avatar is uploaded separately via the user update flow. RBAC: users:write.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "User payload"
// @Success      201      {object} user.UserResponse
// @Failure		400	{object} response.ErrorResponse "Invalid request"
// @Failure		401	{object} response.ErrorResponse "Unauthorized"
// @Failure		403	{object} response.ErrorResponse "Forbidden"
// @Failure		409	{object} response.ErrorResponse "Email already in use"
// @Security     BearerAuth
// @Router       /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			Name     string `json:"name" validate:"required,min=1"`
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required,min=6"`
		} `in:"body=json"`
	}

	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	input := user.CreateUserInput{
		Name:     req.Payload.Name,
		Email:    req.Payload.Email,
		Password: req.Payload.Password,
	}

	u, err := h.userService.Create(r.Context(), input)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	// Build response
	resp := UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		ProfileMedia:    h.profileMediaForUser(u),
		Roles:           []string{},
		CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListUsers godoc
// @Summary      List users (admin)
// @Description  Paginated list of users. RBAC: users:read.
// @Tags         users
// @Produce      json
// @Param        page   query     int  false  "Page number"   default(1)
// @Param        limit  query     int  false  "Page size"     default(10)  maximum(100)
// @Success      200    {object} user.UserListResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page  int `in:"query=page;default=1" validate:"min=1"`
		Limit int `in:"query=limit;default=10" validate:"min=1"`
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

	// Call user service
	users, paginationResult, err := h.userService.List(r.Context(), input.Page, input.Limit)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list users", err)
		return
	}

	// Build response
	userResponses := make([]UserResponse, len(users))
	for i, u := range users {
		roles, _ := h.roleService.GetUserRoles(r.Context(), u.ID)
		if roles == nil {
			roles = []string{}
		}
		userResponses[i] = UserResponse{
			ID:              u.ID,
			Name:            u.Name,
			Email:           u.Email,
			ProfileMedia:    h.profileMediaForUser(u),
			Roles:           roles,
			CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"users":      userResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetUser godoc
// @Summary      Get user by ID (admin)
// @Description  Retrieve a single user by UUID. RBAC: users:read.
// @Tags         users
// @Produce      json
// @Param        id  path      string  true  "User UUID"
// @Success      200  {object} user.UserResponse
// @Failure		400	{object} response.ErrorResponse "Invalid UUID"
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse "User not found"
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "User ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// Call user service
	u, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	// Build response
	getUserRoles, _ := h.roleService.GetUserRoles(r.Context(), u.ID)
	if getUserRoles == nil {
		getUserRoles = []string{}
	}
	resp := UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		ProfileMedia:    h.profileMediaForUser(u),
		Roles:           getUserRoles,
		CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateUser godoc
// @Summary      Update user (admin)
// @Description  Update user name/email and optionally replace profile_image. RBAC: users:write.
// @Tags         users
// @Accept       json
// @Accept       mpfd
// @Produce      json
// @Param        id       path      string          true  "User UUID"
// @Param        request    body      map[string]interface{}  true  "User update payload"
// @Success      200      {object} user.UserResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse "Email already in use"
// @Security     BearerAuth
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "User ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// Determine if request is form data or JSON
	contentType := r.Header.Get("Content-Type")
	var name, email string
	var profileImage *httpin.File
	var profileImageMediaID string

	if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart/form-data with optional file upload
		var req struct {
			Name                string       `in:"form=name" validate:"required,min=1"`
			Email               string       `in:"form=email" validate:"required,email"`
			ProfileImage        *httpin.File `in:"form=profile_image"`
			ProfileImageMediaID *string      `in:"form=profile_image_media_id"`
		}

		if err := httpin.DecodeTo(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := handlerutil.ValidateStruct(req); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		name = req.Name
		email = req.Email
		profileImage = req.ProfileImage
		if req.ProfileImageMediaID != nil {
			profileImageMediaID = *req.ProfileImageMediaID
		}
	} else {
		// Handle JSON body
		var req struct {
			Payload struct {
				Name                string  `json:"name" validate:"required,min=1"`
				Email               string  `json:"email" validate:"required,email"`
				ProfileImageMediaID *string `json:"profile_image_media_id"`
			} `in:"body=json"`
		}

		if err := httpin.DecodeTo(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := handlerutil.ValidateStruct(req.Payload); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		name = req.Payload.Name
		email = req.Payload.Email
		if req.Payload.ProfileImageMediaID != nil {
			profileImageMediaID = *req.Payload.ProfileImageMediaID
		}
	}

	if profileImage != nil && profileImageMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either profile_image or profile_image_media_id, not both")
		return
	}

	// Call user service
	input := user.UpdateUserInput{
		Name:  name,
		Email: email,
	}

	u, err := h.userService.Update(r.Context(), id, input)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		// Check if error contains "user not found" (handles wrapped errors)
		if containsError(err.Error(), "user not found") {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update user", err)
		return
	}

	// Update profile image if provided
	if profileImage != nil {
		u, err = h.uploadProfileImage(r.Context(), id, profileImage)
		if err != nil {
			response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update profile image", err)
			return
		}
	} else if profileImageMediaID != "" {
		u, err = h.userService.UpdateProfileImageMediaID(r.Context(), id, profileImageMediaID)
		if err != nil {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to update profile image", err)
			return
		}
	}

	// Build response
	updateRoles, _ := h.roleService.GetUserRoles(r.Context(), u.ID)
	if updateRoles == nil {
		updateRoles = []string{}
	}
	resp := UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		ProfileMedia:    h.profileMediaForUser(u),
		Roles:           updateRoles,
		CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// DeleteUser godoc
// @Summary      Delete user (admin)
// @Description  Delete a user by UUID. Cascades to user_roles and password_reset_tokens. RBAC: users:write.
// @Tags         users
// @Produce      json
// @Param        id  path  string  true  "User UUID"
// @Success      200  {object} response.MessageResponse
// @Failure		400	{object} response.ErrorResponse "Invalid UUID"
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse "User not found"
// @Security     BearerAuth
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "User ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// Call user service
	err := h.userService.Delete(r.Context(), id)
	if err != nil {
		// Check if error contains "user not found" (handles wrapped errors)
		if containsError(err.Error(), "user not found") {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

// UpdatePassword godoc
// @Summary      Update user password (admin)
// @Description  Admin updates another user's password without old-password check. RBAC: users:write.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id       path      string               true  "User UUID"
// @Param        request    body      map[string]interface{}  true  "New password (admin: OldPassword can be empty)"
// @Success      200      {object} response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @ Security     BearerAuth
// @Router       /users/{id}/password [put]
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var req struct {
		Payload struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password" validate:"required,min=6"`
		} `in:"body=json"`
	}

	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user is admin (from context, set by RBAC middleware)
	// For now, we'll check if old password is provided to determine if admin
	isAdmin := req.Payload.OldPassword == ""

	// Call user service
	input := user.UpdatePasswordInput{
		OldPassword: req.Payload.OldPassword,
		NewPassword: req.Payload.NewPassword,
	}

	err := h.userService.UpdatePassword(r.Context(), id, input, isAdmin)
	if err != nil {
		if err.Error() == "invalid old password" {
			response.Error(w, http.StatusBadRequest, "Invalid old password")
			return
		}
		// Check if error contains "user not found" (handles wrapped errors)
		if containsError(err.Error(), "user not found") {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update password", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Password updated successfully",
	})
}

// GetCurrentUser godoc
// @Summary      Get current user
// @Description  Returns the JWT-authenticated user with their assigned roles.
// @Tags         users
// @Produce      json
// @Success      200  {object} user.UserResponse
// @Failure		401	{object} response.ErrorResponse
// @ Security     BearerAuth
// @Router       /users/me [get]
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Call user service
	u, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	// Get user roles - always return at least an empty slice, never nil
	roles := []string{}
	if userRoles, err := h.roleService.GetUserRoles(r.Context(), userID); err == nil {
		roles = userRoles
	}

	// Build response
	resp := UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		ProfileMedia:    h.profileMediaForUser(u),
		Roles:           roles,
		CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateCurrentUserProfile godoc
// @Summary      Update current user profile
// @Description  Update own profile (name + optional profile_image). Accepts multipart/form-data OR application/json.
// @Tags         users
// @Accept       json
// @Accept       mpfd
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "Profile update"
// @Success      200      {object} user.UserResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me [put]
func (h *UserHandler) UpdateCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Determine if request is form data or JSON
	contentType := r.Header.Get("Content-Type")
	var name, email string
	var profileImage *httpin.File
	var profileImageMediaID string

	if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart/form-data with optional file upload
		var req struct {
			Name                string       `in:"form=name" validate:"required,min=1"`
			Email               string       `in:"form=email" validate:"required,email"`
			ProfileImage        *httpin.File `in:"form=profile_image"`
			ProfileImageMediaID *string      `in:"form=profile_image_media_id"`
		}

		if err := httpin.DecodeTo(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := handlerutil.ValidateStruct(req); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		name = req.Name
		email = req.Email
		profileImage = req.ProfileImage
		if req.ProfileImageMediaID != nil {
			profileImageMediaID = *req.ProfileImageMediaID
		}
	} else {
		// Handle JSON body
		var req struct {
			Payload struct {
				Name                string  `json:"name" validate:"required,min=1"`
				Email               string  `json:"email" validate:"required,email"`
				ProfileImageMediaID *string `json:"profile_image_media_id"`
			} `in:"body=json"`
		}

		if err := httpin.DecodeTo(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := handlerutil.ValidateStruct(req.Payload); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		name = req.Payload.Name
		email = req.Payload.Email
		if req.Payload.ProfileImageMediaID != nil {
			profileImageMediaID = *req.Payload.ProfileImageMediaID
		}
	}

	if profileImage != nil && profileImageMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either profile_image or profile_image_media_id, not both")
		return
	}

	// Call user service
	input := user.UpdateUserInput{
		Name:  name,
		Email: email,
	}

	u, err := h.userService.Update(r.Context(), userID, input)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		if containsError(err.Error(), "user not found") {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	// Update profile image if provided
	if profileImage != nil {
		u, err = h.uploadProfileImage(r.Context(), userID, profileImage)
		if err != nil {
			response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update profile image", err)
			return
		}
	} else if profileImageMediaID != "" {
		u, err = h.userService.UpdateProfileImageMediaID(r.Context(), userID, profileImageMediaID)
		if err != nil {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to update profile image", err)
			return
		}
	}

	// Get user roles - always return at least an empty slice, never nil
	roles := []string{}
	if userRoles, err := h.roleService.GetUserRoles(r.Context(), userID); err == nil {
		roles = userRoles
	}

	// Build response
	resp := UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		ProfileMedia:    h.profileMediaForUser(u),
		Roles:           roles,
		CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// uploadProfileImage streams the multipart file into the gallery FileStore
// via the user service, returning the refreshed user.
func (h *UserHandler) uploadProfileImage(ctx context.Context, userID string, file *httpin.File) (*userdomain.User, error) {
	imageFile, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to read profile image: %w", err)
	}
	defer imageFile.Close()

	contentType := file.MIMEHeader().Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = handlerutil.InferContentType(file.Filename())
	}

	return h.userService.UpdateProfileImage(ctx, userID, galleryuc.FileInput{
		OriginalName: file.Filename(),
		Content:      imageFile,
		Size:         file.Size(),
		ContentType:  contentType,
	})
}