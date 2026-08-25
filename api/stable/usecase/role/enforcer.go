package role

import "context"

// Enforcer defines the interface for RBAC enforcement operations.
// This interface wraps Casbin methods needed by the role service.
// Following Go best practices, this interface is defined in the usecase layer
// where it is USED, not in the domain or interface layers.
//
// The interface will be implemented by interface/rbac/casbin_enforcer.go
// which wraps the external Casbin library.
type Enforcer interface {
	// Enforce checks if a subject has permission to perform an action on an object
	Enforce(ctx context.Context, subject, object, action string) (bool, error)

	// AddRoleForUser assigns a role to a user
	AddRoleForUser(ctx context.Context, user, role string) error

	// DeleteRoleForUser removes a role from a user
	DeleteRoleForUser(ctx context.Context, user, role string) error

	// GetRolesForUser retrieves all roles assigned to a user
	GetRolesForUser(ctx context.Context, user string) ([]string, error)

	// AddPermissionForRole adds a permission to a role
	AddPermissionForRole(ctx context.Context, role, object, action string) error

	// DeletePermissionForRole removes a permission from a role
	DeletePermissionForRole(ctx context.Context, role, object, action string) error

	// GetPermissionsForRole retrieves all permissions for a role
	GetPermissionsForRole(ctx context.Context, role string) ([][]string, error)

	// DeleteRole removes a role and all its permissions
	DeleteRole(ctx context.Context, role string) error

	// GetAllRoles retrieves all defined roles
	GetAllRoles(ctx context.Context) ([]string, error)

	// GetAllPermissions retrieves all available permissions in the system
	GetAllPermissions(ctx context.Context) ([][]string, error)
}
