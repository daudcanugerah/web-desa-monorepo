package role

import (
	"fmt"
	"net/http"

	"webdesa/api/pkg/response"
	"webdesa/api/usecase/role"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	"webdesa/api/pkg/handlerutil")

// RoleHandler handles HTTP requests for role management operations.
// It accepts the role service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
type RoleHandler struct {
	roleService *role.Service
}

// NewRoleHandler creates a new role handler with service dependency injected.
func NewRoleHandler(roleService *role.Service) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// CreateRoleRequest represents the request body for creating a role
type CreateRoleRequest struct {
	Name        string              `json:"name"`
	Permissions []PermissionRequest `json:"permissions"`
}

// PermissionRequest represents a permission in the request
type PermissionRequest struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// AddPermissionRequest represents the request body for adding a permission
type AddPermissionRequest struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// AssignRoleRequest represents the request body for assigning a role to a user
type AssignRoleRequest struct {
	Role string `json:"role"`
}

// RoleResponse represents the response for role data
type RoleResponse struct {
	Name        string               `json:"name"`
	Permissions []PermissionResponse `json:"permissions"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

// PermissionResponse represents a permission in the response
type PermissionResponse struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// CreateRole godoc
// @Summary      Create a role
// @Description  Create a new RBAC role with a list of permissions. RBAC: roles:write.
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "Role payload"
// @Success      201      {object} role.RoleResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @ Security     BearerAuth
// @Router       /roles [post]
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			Name        string              `json:"name" validate:"required,min=1"`
			Permissions []PermissionRequest `json:"permissions"`
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

	// Convert request permissions to service input
	permissions := make([]role.Permission, len(req.Payload.Permissions))
	for i, p := range req.Payload.Permissions {
		permissions[i] = role.Permission{
			Resource: p.Resource,
			Action:   p.Action,
		}
	}

	// Call role service
	input := role.CreateRoleInput{
		Name:        req.Payload.Name,
		Permissions: permissions,
	}

	createdRole, err := h.roleService.CreateRole(r.Context(), input)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to create role", err)
		return
	}

	// Build response
	permResponses := make([]PermissionResponse, len(createdRole.Permissions))
	for i, p := range createdRole.Permissions {
		permResponses[i] = PermissionResponse{
			Resource: p.Resource,
			Action:   p.Action,
		}
	}

	resp := RoleResponse{
		Name:        createdRole.Name,
		Permissions: permResponses,
		CreatedAt:   createdRole.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   createdRole.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListRoles godoc
// @Summary      List roles
// @Description  Returns all defined roles with their permissions. RBAC: roles:read.
// @Tags         roles
// @Produce      json
// @Success      200  {object} []role.RoleResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /roles [get]
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	// Call role service
	roles, err := h.roleService.ListRoles(r.Context())
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list roles", err)
		return
	}

	// Build response
	roleResponses := make([]RoleResponse, len(roles))
	for i, r := range roles {
		permResponses := make([]PermissionResponse, len(r.Permissions))
		for j, p := range r.Permissions {
			permResponses[j] = PermissionResponse{
				Resource: p.Resource,
				Action:   p.Action,
			}
		}

		roleResponses[i] = RoleResponse{
			Name:        r.Name,
			Permissions: permResponses,
			CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response.Success(w, http.StatusOK, roleResponses)
}

// GetRolePermissions godoc
// @Summary      List role permissions
// @Description  Returns all permissions assigned to a role. RBAC: roles:read.
// @Tags         roles
// @Produce      json
// @Param        role  path  string  true  "Role name"
// @Success      200   {object} []role.PermissionResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @ Security     BearerAuth
// @Router       /roles/{role}/permissions [get]
func (h *RoleHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	roleName := chi.URLParam(r, "role")
	if roleName == "" {
		response.Error(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Call role service
	permissions, err := h.roleService.GetRolePermissions(r.Context(), roleName)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get role permissions", err)
		return
	}

	// Build response
	permResponses := make([]PermissionResponse, len(permissions))
	for i, p := range permissions {
		permResponses[i] = PermissionResponse{
			Resource: p.Resource,
			Action:   p.Action,
		}
	}

	response.Success(w, http.StatusOK, permResponses)
}

// GetAvailablePermissions godoc
// @Summary      List available system permissions
// @Description  Returns the de-duplicated set of available (resource, action) permission pairs. RBAC: roles:read.
// @Tags         roles
// @Produce      json
// @Success      200  {object} []role.PermissionResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /permissions [get]
func (h *RoleHandler) GetAvailablePermissions(w http.ResponseWriter, r *http.Request) {
	// Call role service
	permissions, err := h.roleService.GetAvailablePermissions(r.Context())
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get available permissions", err)
		return
	}

	// Build response
	permResponses := make([]PermissionResponse, len(permissions))
	for i, p := range permissions {
		permResponses[i] = PermissionResponse{
			Resource: p.Resource,
			Action:   p.Action,
		}
	}

	response.Success(w, http.StatusOK, permResponses)
}

// AddPermissionToRole godoc
// @Summary      Add permission to role
// @Description  Grant (resource, action) permission to a role. RBAC: roles:write.
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        role     path      string                 true  "Role name"
// @Param        request    body      map[string]interface{}  true  "Permission payload"
// @Success      200      {object} response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @ Security     BearerAuth
// @Router       /roles/{role}/permissions [post]
func (h *RoleHandler) AddPermissionToRole(w http.ResponseWriter, r *http.Request) {
	roleName := chi.URLParam(r, "role")
	if roleName == "" {
		response.Error(w, http.StatusBadRequest, "Role name is required")
		return
	}

	var req struct {
		Payload struct {
			Resource string `json:"resource" validate:"required,min=1"`
			Action   string `json:"action" validate:"required,min=1"`
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

	// Call role service
	perm := role.Permission{
		Resource: req.Payload.Resource,
		Action:   req.Payload.Action,
	}

	err := h.roleService.AddPermissionToRole(r.Context(), roleName, perm)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to add permission to role", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Permission added to role successfully",
	})
}

// RemovePermissionFromRole godoc
// @Summary      Remove permission from role
// @Description  Revoke (resource, action) permission from a role. The permission identifier is passed as `resource:action` query parameter. RBAC: roles:write.
// @Tags         roles
// @Produce      json
// @Param        role        path     string  true  "Role name"
// @Param        permission  query    string  true  "Permission identifier in resource:action format"
// @Success      200         {object}  response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /roles/{role}/permissions [delete]
func (h *RoleHandler) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	roleName := chi.URLParam(r, "role")
	if roleName == "" {
		response.Error(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Permission is passed as "resource:action" in URL
	permissionStr := chi.URLParam(r, "permission")
	if permissionStr == "" {
		response.Error(w, http.StatusBadRequest, "Permission is required")
		return
	}

	// Parse permission string (format: "resource:action")
	var resource, action string
	if _, err := fmt.Sscanf(permissionStr, "%s:%s", &resource, &action); err != nil || resource == "" || action == "" {
		response.Error(w, http.StatusBadRequest, "Invalid permission format, expected 'resource:action'")
		return
	}

	// Call role service
	perm := role.Permission{
		Resource: resource,
		Action:   action,
	}

	err := h.roleService.RemovePermissionFromRole(r.Context(), roleName, perm)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to remove permission from role", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Permission removed from role successfully",
	})
}

// DeleteRole godoc
// @Summary      Delete role
// @Description  Delete a role by name. Fails with 400 if any user is still assigned. RBAC: roles:write.
// @Tags         roles
// @Produce      json
// @Param        role  path  string  true  "Role name"
// @Success      200   {object} response.MessageResponse
// @Failure		400	{object} response.ErrorResponse "Role in use"
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /roles/{role} [delete]
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleName := chi.URLParam(r, "role")
	if roleName == "" {
		response.Error(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Call role service
	err := h.roleService.DeleteRole(r.Context(), roleName)
	if err != nil {
		// Check if error is due to users being assigned to the role
		if err.Error() == "cannot delete role: users are assigned to this role" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete role", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Role deleted successfully",
	})
}

// AssignRoleToUser godoc
// @Summary      Assign role to user
// @Description  Double-writes to Casbin and user_roles table. RBAC: roles:write.
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "User UUID"
// @Param        request    body      map[string]interface{}  true  "Role name"
// @Success      200      {object} response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @ Security     BearerAuth
// @Router       /users/{id}/roles [post]
func (h *RoleHandler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		response.Error(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var req struct {
		Payload struct {
			Role string `json:"role" validate:"required,min=1"`
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

	// Call role service
	err := h.roleService.AssignRoleToUser(r.Context(), userID, req.Payload.Role)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to assign role to user", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Role assigned to user successfully",
	})
}

// RemoveRoleFromUser godoc
// @Summary      Remove role from user
// @Description  Double-removes from Casbin and user_roles table. RBAC: roles:write.
// @Tags         roles
// @Produce      json
// @Param        id    path  string  true  "User UUID"
// @Param        role  path  string  true  "Role name"
// @Success      200   {object} response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @ Security     BearerAuth
// @Router       /users/{id}/roles/{role} [delete]
func (h *RoleHandler) RemoveRoleFromUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		response.Error(w, http.StatusBadRequest, "User ID is required")
		return
	}

	roleName := chi.URLParam(r, "role")
	if roleName == "" {
		response.Error(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Call role service
	err := h.roleService.RemoveRoleFromUser(r.Context(), userID, roleName)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to remove role from user", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Role removed from user successfully",
	})
}
