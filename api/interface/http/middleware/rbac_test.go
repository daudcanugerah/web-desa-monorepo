package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Mock adapter for Casbin
type mockAdapter struct{}

func (a *mockAdapter) LoadPolicy(model model.Model) error {
	return nil
}

func (a *mockAdapter) SavePolicy(model model.Model) error {
	return nil
}

func (a *mockAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (a *mockAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (a *mockAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

var _ persist.Adapter = (*mockAdapter)(nil)

func setupTestEnforcer(t *testing.T) *casbin.Enforcer {
	// Create model
	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
	if err != nil {
		t.Fatal(err)
	}

	// Create enforcer with mock adapter
	enforcer, err := casbin.NewEnforcer(m, &mockAdapter{})
	if err != nil {
		t.Fatal(err)
	}

	return enforcer
}

func TestRBACMiddleware_Success(t *testing.T) {
	enforcer := setupTestEnforcer(t)
	userID := uuid.New().String()

	// Add role and policy
	enforcer.AddRoleForUser(userID, "admin")
	enforcer.AddPolicy("admin", "users", "read")

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := RBACMiddleware(enforcer, "users", "read")
	handler := middleware(nextHandler)

	// Create request with user context
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
}

func TestRBACMiddleware_NoPermission(t *testing.T) {
	enforcer := setupTestEnforcer(t)
	userID := uuid.New().String()

	// Add role but no policy for the action
	enforcer.AddRoleForUser(userID, "user")
	enforcer.AddPolicy("user", "users", "read")

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := RBACMiddleware(enforcer, "users", "write")
	handler := middleware(nextHandler)

	// Create request with user context
	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRBACMiddleware_NoUserInContext(t *testing.T) {
	enforcer := setupTestEnforcer(t)

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := RBACMiddleware(enforcer, "users", "read")
	handler := middleware(nextHandler)

	// Create request without user context
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRBACMiddleware_MultipleRoles(t *testing.T) {
	enforcer := setupTestEnforcer(t)
	userID := uuid.New().String()

	// Add multiple roles
	enforcer.AddRoleForUser(userID, "user")
	enforcer.AddRoleForUser(userID, "admin")

	// Only admin has permission
	enforcer.AddPolicy("admin", "users", "delete")

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := RBACMiddleware(enforcer, "users", "delete")
	handler := middleware(nextHandler)

	// Create request with user context
	req := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWithRBAC(t *testing.T) {
	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resource := r.Context().Value("rbac_resource")
		action := r.Context().Value("rbac_action")

		assert.Equal(t, "users", resource)
		assert.Equal(t, "read", action)

		w.WriteHeader(http.StatusOK)
	})

	middleware := WithRBAC("users", "read")
	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRBACMiddlewareWithContext_Success(t *testing.T) {
	enforcer := setupTestEnforcer(t)
	userID := uuid.New().String()

	// Add role and policy
	enforcer.AddRoleForUser(userID, "admin")
	enforcer.AddPolicy("admin", "users", "read")

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := RBACMiddlewareWithContext(enforcer)
	handler := middleware(nextHandler)

	// Create request with user and RBAC context
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	ctx = context.WithValue(ctx, "rbac_resource", "users")
	ctx = context.WithValue(ctx, "rbac_action", "read")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRBACMiddlewareWithContext_MissingContext(t *testing.T) {
	enforcer := setupTestEnforcer(t)
	userID := uuid.New()

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RBACMiddlewareWithContext(enforcer)
	handler := middleware(nextHandler)

	// Create request without RBAC context
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
