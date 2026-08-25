package casbin

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlxAdapter_Integration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Connect to test database
	dsn := "postgres://desauser:desapass@localhost:5432/desa?sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err, "Failed to connect to database")
	defer db.Close()

	// Create adapter
	adapter, err := NewSqlxAdapter(db)
	require.NoError(t, err, "Failed to create adapter")

	// Create enforcer
	enforcer, err := casbin.NewEnforcer("../../rbac/rbac_model.conf", adapter)
	require.NoError(t, err, "Failed to create enforcer")

	// Load policies
	err = enforcer.LoadPolicy()
	require.NoError(t, err, "Failed to load policies")

	t.Run("Admin has user create permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("admin", "user", "create")
		require.NoError(t, err)
		assert.True(t, allowed, "Admin should have user:create permission")
	})

	t.Run("Admin has user read permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("admin", "user", "read")
		require.NoError(t, err)
		assert.True(t, allowed, "Admin should have user:read permission")
	})

	t.Run("Admin has user update permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("admin", "user", "update")
		require.NoError(t, err)
		assert.True(t, allowed, "Admin should have user:update permission")
	})

	t.Run("Admin has user delete permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("admin", "user", "delete")
		require.NoError(t, err)
		assert.True(t, allowed, "Admin should have user:delete permission")
	})

	t.Run("Operator has berita create permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("operator", "berita", "create")
		require.NoError(t, err)
		assert.True(t, allowed, "Operator should have berita:create permission")
	})

	t.Run("Operator has berita update permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("operator", "berita", "update")
		require.NoError(t, err)
		assert.True(t, allowed, "Operator should have berita:update permission")
	})

	t.Run("Operator does NOT have user create permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("operator", "user", "create")
		require.NoError(t, err)
		assert.False(t, allowed, "Operator should NOT have user:create permission")
	})

	t.Run("Operator does NOT have user delete permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("operator", "user", "delete")
		require.NoError(t, err)
		assert.False(t, allowed, "Operator should NOT have user:delete permission")
	})

	t.Run("Operator does NOT have berita delete permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("operator", "berita", "delete")
		require.NoError(t, err)
		assert.False(t, allowed, "Operator should NOT have berita:delete permission")
	})

	t.Run("Admin has backup execute permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("admin", "backup", "execute")
		require.NoError(t, err)
		assert.True(t, allowed, "Admin should have backup:execute permission")
	})

	t.Run("Operator does NOT have backup execute permission", func(t *testing.T) {
		allowed, err := enforcer.Enforce("operator", "backup", "execute")
		require.NoError(t, err)
		assert.False(t, allowed, "Operator should NOT have backup:execute permission")
	})
}

func TestSqlxAdapter_AddAndRemovePolicy(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Connect to test database
	dsn := "postgres://desauser:desapass@localhost:5432/desa?sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err, "Failed to connect to database")
	defer db.Close()

	// Create adapter
	adapter, err := NewSqlxAdapter(db)
	require.NoError(t, err, "Failed to create adapter")

	// Create enforcer
	enforcer, err := casbin.NewEnforcer("../../rbac/rbac_model.conf", adapter)
	require.NoError(t, err, "Failed to create enforcer")

	// Load policies
	err = enforcer.LoadPolicy()
	require.NoError(t, err, "Failed to load policies")

	t.Run("Add new policy", func(t *testing.T) {
		// Add a new policy
		added, err := enforcer.AddPolicy("test_role", "test_resource", "test_action")
		require.NoError(t, err)
		assert.True(t, added, "Policy should be added")

		// Verify policy exists
		allowed, err := enforcer.Enforce("test_role", "test_resource", "test_action")
		require.NoError(t, err)
		assert.True(t, allowed, "New policy should be enforced")

		// Cleanup: Remove the test policy
		removed, err := enforcer.RemovePolicy("test_role", "test_resource", "test_action")
		require.NoError(t, err)
		assert.True(t, removed, "Policy should be removed")

		// Verify policy is removed
		allowed, err = enforcer.Enforce("test_role", "test_resource", "test_action")
		require.NoError(t, err)
		assert.False(t, allowed, "Removed policy should not be enforced")
	})
}
