package rbac

import (
	"context"
	"fmt"

	"webdesa/api/usecase/role"

	"braces.dev/errtrace"
	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
)

// CasbinEnforcer wraps the Casbin enforcer to implement the usecase/role.Enforcer interface.
// This adapter allows the use case layer to remain independent of the Casbin library.
// Following Clean Architecture, dependencies point inward: interface/rbac → usecase/role.
type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

// NewCasbinEnforcer creates a new Casbin enforcer backed by PostgreSQL via sqlx.
// Policies are persisted in the casbin_rule table, surviving restarts.
func NewCasbinEnforcer(modelPath string, db *sqlx.DB) (role.Enforcer, error) {
	adapter := newPGAdapter(db)

	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to create casbin enforcer: %w", err))
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to load casbin policy: %w", err))
	}

	return &CasbinEnforcer{enforcer: enforcer}, nil
}

// Enforce checks if a subject has permission to perform an action on an object.
// Returns true if the permission is granted, false otherwise.
func (c *CasbinEnforcer) Enforce(ctx context.Context, subject, object, action string) (bool, error) {
	allowed, err := c.enforcer.Enforce(subject, object, action)
	if err != nil {
		return false, errtrace.Wrap(fmt.Errorf("failed to enforce policy: %w", err))
	}
	return allowed, nil
}

// AddRoleForUser assigns a role to a user.
func (c *CasbinEnforcer) AddRoleForUser(ctx context.Context, user, roleName string) error {
	_, err := c.enforcer.AddRoleForUser(user, roleName)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to add role for user: %w", err))
	}
	return nil
}

// DeleteRoleForUser removes a role from a user.
func (c *CasbinEnforcer) DeleteRoleForUser(ctx context.Context, user, roleName string) error {
	_, err := c.enforcer.DeleteRoleForUser(user, roleName)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete role for user: %w", err))
	}
	return nil
}

// GetRolesForUser retrieves all roles assigned to a user.
// Returns a slice of role names.
func (c *CasbinEnforcer) GetRolesForUser(ctx context.Context, user string) ([]string, error) {
	roles, err := c.enforcer.GetRolesForUser(user)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get roles for user: %w", err))
	}
	return roles, nil
}

// AddPermissionForRole adds a permission to a role.
func (c *CasbinEnforcer) AddPermissionForRole(ctx context.Context, roleName, object, action string) error {
	_, err := c.enforcer.AddPolicy(roleName, object, action)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to add permission for role: %w", err))
	}
	return nil
}

// DeletePermissionForRole removes a permission from a role.
func (c *CasbinEnforcer) DeletePermissionForRole(ctx context.Context, roleName, object, action string) error {
	_, err := c.enforcer.RemovePolicy(roleName, object, action)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete permission for role: %w", err))
	}
	return nil
}

// GetPermissionsForRole retrieves all permissions for a role.
// Returns a slice of permission tuples [role, object, action].
func (c *CasbinEnforcer) GetPermissionsForRole(ctx context.Context, roleName string) ([][]string, error) {
	permissions, err := c.enforcer.GetFilteredPolicy(0, roleName)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get permissions for role: %w", err))
	}
	return permissions, nil
}

// DeleteRole removes a role and all its permissions.
func (c *CasbinEnforcer) DeleteRole(ctx context.Context, roleName string) error {
	// Remove all policies for this role
	_, err := c.enforcer.RemoveFilteredPolicy(0, roleName)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete role: %w", err))
	}
	return nil
}

// GetAllRoles retrieves all defined roles.
// Returns a slice of unique role names from the policy.
func (c *CasbinEnforcer) GetAllRoles(ctx context.Context) ([]string, error) {
	allPolicies, err := c.enforcer.GetPolicy()
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get all policies: %w", err))
	}

	// Extract unique roles from policies
	roleMap := make(map[string]bool)
	for _, policy := range allPolicies {
		if len(policy) > 0 {
			roleMap[policy[0]] = true
		}
	}

	roles := make([]string, 0, len(roleMap))
	for roleName := range roleMap {
		roles = append(roles, roleName)
	}

	return roles, nil
}

// GetAllPermissions retrieves all available permissions in the system.
// Returns a slice of permission tuples [role, object, action].
func (c *CasbinEnforcer) GetAllPermissions(ctx context.Context) ([][]string, error) {
	permissions, err := c.enforcer.GetPolicy()
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get all permissions: %w", err))
	}
	return permissions, nil
}

// GetEnforcer returns the underlying Casbin enforcer.
// This is needed for middleware that requires direct access to the enforcer.
func (c *CasbinEnforcer) GetEnforcer() *casbin.Enforcer {
	return c.enforcer
}
