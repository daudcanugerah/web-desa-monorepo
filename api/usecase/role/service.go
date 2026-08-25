package role

import (
	"context"
	"fmt"
	"time"
)

// Service implements role management business logic.
// It accepts the Enforcer interface and returns concrete structs.
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	enforcer Enforcer
	repo     Repository
}

// NewService creates a new role management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(enforcer Enforcer, repo Repository) *Service {
	return &Service{
		enforcer: enforcer,
		repo:     repo,
	}
}

// CreateRole creates a new role with specified permissions.
// Returns concrete Role struct.
//
// Validates: Requirements 6.1
func (s *Service) CreateRole(ctx context.Context, input CreateRoleInput) (*Role, error) {
	// Validate input
	if input.Name == "" {
		return nil, fmt.Errorf("role name is required")
	}

	// Add permissions to the role in Casbin
	for _, perm := range input.Permissions {
		if err := s.enforcer.AddPermissionForRole(ctx, input.Name, perm.Resource, perm.Action); err != nil {
			return nil, fmt.Errorf("failed to add permission to role: %w", err)
		}
	}

	// Return the created role with timestamps
	now := time.Now()
	return &Role{
		Name:        input.Name,
		Permissions: input.Permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ListRoles retrieves all available roles.
// Returns concrete Role structs.
//
// Validates: Requirements 6.2
func (s *Service) ListRoles(ctx context.Context) ([]Role, error) {
	// Get all roles from Casbin
	roleNames, err := s.enforcer.GetAllRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	// Build role list with permissions
	roles := make([]Role, 0, len(roleNames))
	now := time.Now()
	for _, roleName := range roleNames {
		// Get permissions for this role
		perms, err := s.enforcer.GetPermissionsForRole(ctx, roleName)
		if err != nil {
			return nil, fmt.Errorf("failed to get permissions for role %s: %w", roleName, err)
		}

		// Convert Casbin permission format to our Permission struct
		permissions := make([]Permission, 0, len(perms))
		for _, perm := range perms {
			// Casbin returns permissions as [][]string where each entry is [role, resource, action]
			if len(perm) >= 3 {
				permissions = append(permissions, Permission{
					Resource: perm[1],
					Action:   perm[2],
				})
			}
		}

		roles = append(roles, Role{
			Name:        roleName,
			Permissions: permissions,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	return roles, nil
}

// GetRolePermissions retrieves all permissions assigned to a role.
// Returns concrete Permission structs.
//
// Validates: Requirements 6.3
func (s *Service) GetRolePermissions(ctx context.Context, roleName string) ([]Permission, error) {
	// Get permissions from Casbin
	perms, err := s.enforcer.GetPermissionsForRole(ctx, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for role: %w", err)
	}

	// Convert Casbin permission format to our Permission struct
	permissions := make([]Permission, 0, len(perms))
	for _, perm := range perms {
		// Casbin returns permissions as [][]string where each entry is [role, resource, action]
		if len(perm) >= 3 {
			permissions = append(permissions, Permission{
				Resource: perm[1],
				Action:   perm[2],
			})
		}
	}

	return permissions, nil
}

// GetAvailablePermissions retrieves all system permissions.
// Returns concrete Permission structs.
//
// Validates: Requirements 6.4
func (s *Service) GetAvailablePermissions(ctx context.Context) ([]Permission, error) {
	// Get all permissions from Casbin
	perms, err := s.enforcer.GetAllPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available permissions: %w", err)
	}

	// Convert Casbin permission format to our Permission struct
	// Remove duplicates by using a map
	permMap := make(map[string]Permission)
	for _, perm := range perms {
		// Casbin returns permissions as [][]string where each entry is [role, resource, action]
		if len(perm) >= 3 {
			key := perm[1] + ":" + perm[2]
			permMap[key] = Permission{
				Resource: perm[1],
				Action:   perm[2],
			}
		}
	}

	// Convert map to slice
	permissions := make([]Permission, 0, len(permMap))
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// AddPermissionToRole adds a permission to a role.
// Updates the RBAC policy in Casbin.
//
// Validates: Requirements 6.5
func (s *Service) AddPermissionToRole(ctx context.Context, roleName string, perm Permission) error {
	// Validate input
	if roleName == "" {
		return fmt.Errorf("role name is required")
	}
	if perm.Resource == "" || perm.Action == "" {
		return fmt.Errorf("permission resource and action are required")
	}

	// Add permission to Casbin
	if err := s.enforcer.AddPermissionForRole(ctx, roleName, perm.Resource, perm.Action); err != nil {
		return fmt.Errorf("failed to add permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role.
// Updates the RBAC policy in Casbin.
//
// Validates: Requirements 6.6
func (s *Service) RemovePermissionFromRole(ctx context.Context, roleName string, perm Permission) error {
	// Validate input
	if roleName == "" {
		return fmt.Errorf("role name is required")
	}
	if perm.Resource == "" || perm.Action == "" {
		return fmt.Errorf("permission resource and action are required")
	}

	// Remove permission from Casbin
	if err := s.enforcer.DeletePermissionForRole(ctx, roleName, perm.Resource, perm.Action); err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	return nil
}

// DeleteRole deletes a role if no users are assigned to it.
// Returns error if users are still assigned to the role.
//
// Validates: Requirements 6.7, 6.8
func (s *Service) DeleteRole(ctx context.Context, roleName string) error {
	// Validate input
	if roleName == "" {
		return fmt.Errorf("role name is required")
	}

	// Check if any users are assigned to this role
	count, err := s.repo.CountUsersWithRole(ctx, roleName)
	if err != nil {
		return fmt.Errorf("failed to check users with role: %w", err)
	}

	// If users are assigned, return error (HTTP 400 Bad Request at handler level)
	if count > 0 {
		return fmt.Errorf("cannot delete role: %d user(s) are assigned to this role", count)
	}

	// Delete the role from Casbin
	if err := s.enforcer.DeleteRole(ctx, roleName); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// AssignRoleToUser assigns a role to a user.
// Adds user-role mapping in Casbin and persists to user_roles table.
//
// Validates: Requirements 6.1 (role assignment part)
func (s *Service) AssignRoleToUser(ctx context.Context, userID, roleName string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if roleName == "" {
		return fmt.Errorf("role name is required")
	}

	if err := s.enforcer.AddRoleForUser(ctx, userID, roleName); err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	if err := s.repo.AssignUserRole(ctx, userID, roleName); err != nil {
		return fmt.Errorf("failed to persist user role: %w", err)
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user.
// Removes user-role mapping in Casbin and from user_roles table.
//
// Validates: Requirements 6.1 (role removal part)
func (s *Service) RemoveRoleFromUser(ctx context.Context, userID, roleName string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if roleName == "" {
		return fmt.Errorf("role name is required")
	}

	if err := s.enforcer.DeleteRoleForUser(ctx, userID, roleName); err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	if err := s.repo.RemoveUserRole(ctx, userID, roleName); err != nil {
		return fmt.Errorf("failed to remove user role record: %w", err)
	}

	return nil
}

// GetUserRoles retrieves all roles assigned to a user.
// Returns role names as strings.
func (s *Service) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	// Validate input
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get roles from Casbin
	roles, err := s.enforcer.GetRolesForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}
