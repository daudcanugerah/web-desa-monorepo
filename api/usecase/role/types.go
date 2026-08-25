package role

import "time"

// Role represents a role in the system with its permissions.
// This is a concrete struct returned by the service.
type Role struct {
	Name        string
	Permissions []Permission
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Permission represents a permission (resource + action).
// This is a concrete struct returned by the service.
type Permission struct {
	Resource string
	Action   string
}

// CreateRoleInput represents the input for creating a role.
type CreateRoleInput struct {
	Name        string
	Permissions []Permission
}
