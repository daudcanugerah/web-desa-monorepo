package user

import (
	"context"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	"webdesa/api/interface/http/middleware"
	"webdesa/api/pkg/response"
	"webdesa/api/usecase/user"

	"webdesa/api/pkg/handlerutil")

// UserHandler handles HTTP requests for user management operations.
// It accepts the user service and role service from the usecase layer as dependencies.
// Dependencies point inward: interface/http → usecase → domain
type UserHandler struct {
	userService *user.Service
	roleService RoleService
	fileHandler user.FileHandler
}

// RoleService defines the interface for role operations needed by UserHandler.
// Following Go best practices, this interface is defined where it's used.
type RoleService interface {
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

// NewUserHandler creates a new user handler with service dependencies injected.
func NewUserHandler(userService *user.Service, roleService RoleService, fileHandler user.FileHandler) *UserHandler {
	return &UserHandler{
		userService: userService,
		roleService: roleService,
		fileHandler: fileHandler,
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

// UpdateProfileRequest represents the request body for updating profile
type UpdateProfileRequest struct {
	Name            string  `json:"name"`
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
}

// UserResponse represents the response for user data
type UserResponse struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	ProfileImageURL *string  `json:"profile_image_url,omitempty"`
	Roles           []string `json:"roles"`
	CreatedAt       string   `json:"created_at"` // ISO 8601 format
	UpdatedAt       string   `json:"updated_at"` // ISO 8601 format
}

// UserListResponse represents a paginated list of users.
type UserListResponse struct {
	Users      []UserResponse          `json:"users"`
	Pagination map[string]interface{}  `json:"pagination"`
}

// CreateUser godoc
// @Summary      Create a new user (admin)
// @Description  Create a new user account. Accepts multipart/form-data (with optional profile_image) OR application/json. RBAC: users:write.
// @Tags         users
// @Accept       json
// @Accept       mpfd
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
	// Determine if request is form data or JSON
	contentType := r.Header.Get("Content-Type")
	var name, email, password string
	var profileImageURL *string

	if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart/form-data with optional file upload
		var req struct {
			Name         string       `in:"form=name" validate:"required,min=1"`
			Email        string       `in:"form=email" validate:"required,email"`
			Password     string       `in:"form=password" validate:"required,min=6"`
			ProfileImage *httpin.File `in:"form=profile_image"`
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
		password = req.Password

		// Handle profile image upload if provided
		if req.ProfileImage != nil {
			imageFile, err := req.ProfileImage.Open()
			if err != nil {
				response.Error(w, http.StatusBadRequest, "Failed to read profile image")
				return
			}
			defer imageFile.Close()

			contentType := req.ProfileImage.MIMEHeader().Get("Content-Type")
			if contentType == "" {
				contentType = "image/jpeg"
			}

			imageURL, err := h.fileHandler.SaveImage(r.Context(), req.ProfileImage.Filename(), imageFile, req.ProfileImage.Size(), contentType)
			if err != nil {
				response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to upload profile image", err)
				return
			}
			profileImageURL = &imageURL
		}
	} else {
		// Handle JSON body
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

		name = req.Payload.Name
		email = req.Payload.Email
		password = req.Payload.Password
	}

	// Call user service
	input := user.CreateUserInput{
		Name:     name,
		Email:    email,
		Password: password,
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

	// Update profile image if provided
	if profileImageURL != nil {
		u.ProfileImageURL = profileImageURL
		// Persist the image URL update
		if err := h.userService.UpdateProfileImage(r.Context(), u.ID, *profileImageURL); err != nil {
			response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update profile image", err)
			return
		}
	}

	// Build response
	resp := UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		ProfileImageURL: u.ProfileImageURL,
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
			ProfileImageURL: u.ProfileImageURL,
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
		ProfileImageURL: u.ProfileImageURL,
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
	var profileImageURL *string

	if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart/form-data with optional file upload
		var req struct {
			Name         string       `in:"form=name" validate:"required,min=1"`
			Email        string       `in:"form=email" validate:"required,email"`
			ProfileImage *httpin.File `in:"form=profile_image"`
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

		// Handle profile image upload if provided
		if req.ProfileImage != nil {
			imageFile, err := req.ProfileImage.Open()
			if err != nil {
				response.Error(w, http.StatusBadRequest, "Failed to read profile image")
				return
			}
			defer imageFile.Close()

			contentType := req.ProfileImage.MIMEHeader().Get("Content-Type")
			if contentType == "" {
				contentType = "image/jpeg"
			}

			imageURL, err := h.fileHandler.SaveImage(r.Context(), req.ProfileImage.Filename(), imageFile, req.ProfileImage.Size(), contentType)
			if err != nil {
				response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to upload profile image", err)
				return
			}
			profileImageURL = &imageURL
		}
	} else {
		// Handle JSON body
		var req struct {
			Payload struct {
				Name  string `json:"name" validate:"required,min=1"`
				Email string `json:"email" validate:"required,email"`
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
	if profileImageURL != nil {
		// Delete old image if exists
		if u.ProfileImageURL != nil && *u.ProfileImageURL != "" {
			_ = h.fileHandler.Delete(r.Context(), *u.ProfileImageURL)
		}
		u.ProfileImageURL = profileImageURL
		// Persist the image URL update
		if err := h.userService.UpdateProfileImage(r.Context(), id, *profileImageURL); err != nil {
			response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update profile image", err)
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
		ProfileImageURL: u.ProfileImageURL,
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
// @Failure		400	{object} response.ErrorResponse "Invalid payload"
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
		ProfileImageURL: u.ProfileImageURL,
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
	var profileImageURL *string

	if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart/form-data with optional file upload
		var req struct {
			Name         string       `in:"form=name" validate:"required,min=1"`
			Email        string       `in:"form=email" validate:"required,email"`
			ProfileImage *httpin.File `in:"form=profile_image"`
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

		// Handle profile image upload if provided
		if req.ProfileImage != nil {
			imageFile, err := req.ProfileImage.Open()
			if err != nil {
				response.Error(w, http.StatusBadRequest, "Failed to read profile image")
				return
			}
			defer imageFile.Close()

			contentType := req.ProfileImage.MIMEHeader().Get("Content-Type")
			if contentType == "" {
				contentType = "image/jpeg"
			}

			imageURL, err := h.fileHandler.SaveImage(r.Context(), req.ProfileImage.Filename(), imageFile, req.ProfileImage.Size(), contentType)
			if err != nil {
				response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to upload profile image", err)
				return
			}
			profileImageURL = &imageURL
		}
	} else {
		// Handle JSON body
		var req struct {
			Payload struct {
				Name  string `json:"name" validate:"required,min=1"`
				Email string `json:"email" validate:"required,email"`
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
	if profileImageURL != nil {
		// Delete old image if exists
		if u.ProfileImageURL != nil && *u.ProfileImageURL != "" {
			_ = h.fileHandler.Delete(r.Context(), *u.ProfileImageURL)
		}
		u.ProfileImageURL = profileImageURL
		// Persist the image URL update
		if err := h.userService.UpdateProfileImage(r.Context(), userID, *profileImageURL); err != nil {
			response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update profile image", err)
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
		ProfileImageURL: u.ProfileImageURL,
		Roles:           roles,
		CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}
