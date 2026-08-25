package role

import "context"

// Repository defines the interface for role-related data operations.
// This interface is defined in the usecase layer where it is USED,
// following Go best practices for interface placement.
//
// The interface will be implemented by interface/repository/role_postgres.go
type Repository interface {
	// CountUsersWithRole returns the number of users assigned to a role
	CountUsersWithRole(ctx context.Context, roleName string) (int, error)

	// AssignUserRole inserts a user-role record into user_roles table
	AssignUserRole(ctx context.Context, userID, roleName string) error

	// RemoveUserRole deletes a user-role record from user_roles table
	RemoveUserRole(ctx context.Context, userID, roleName string) error
}
