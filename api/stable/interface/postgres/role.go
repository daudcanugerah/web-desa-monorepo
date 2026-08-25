package postgres

import (
	"context"
	"fmt"

	"braces.dev/errtrace"
	"github.com/jmoiron/sqlx"

	roleUsecase "webdesa/api/usecase/role"
)

// RoleRepository implements usecase/role.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/role where it's USED, not here where it's implemented.
type RoleRepository struct {
	db *sqlx.DB
}

// NewRoleRepository creates a new RoleRepository instance
func NewRoleRepository(db *sqlx.DB) roleUsecase.Repository {
	return &RoleRepository{db: db}
}

// CountUsersWithRole returns the number of users assigned to a role
// Used to prevent deletion of roles that are still in use
// Validates: Requirements 6.8
func (r *RoleRepository) CountUsersWithRole(ctx context.Context, roleName string) (int, error) {
	var count int

	query := `SELECT COUNT(DISTINCT user_id) FROM user_roles WHERE role = $1`

	err := r.db.GetContext(ctx, &count, query, roleName)
	if err != nil {
		return 0, errtrace.Wrap(fmt.Errorf("failed to count users with role: %w", err))
	}

	return count, nil
}

// AssignUserRole inserts a user-role record into the user_roles table.
func (r *RoleRepository) AssignUserRole(ctx context.Context, userID, roleName string) error {
	query := `
		INSERT INTO user_roles (user_id, role)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, userID, roleName)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to assign user role: %w", err))
	}
	return nil
}

// RemoveUserRole deletes a user-role record from the user_roles table.
func (r *RoleRepository) RemoveUserRole(ctx context.Context, userID, roleName string) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role = $2`
	_, err := r.db.ExecContext(ctx, query, userID, roleName)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to remove user role: %w", err))
	}
	return nil
}
