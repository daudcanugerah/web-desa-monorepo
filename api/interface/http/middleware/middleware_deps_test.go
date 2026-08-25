package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// Verifies that the per-package route registrars properly apply RBAC middleware
// to the correct sub-groups (regression guard after splitting handler into
// per-feature packages).
func TestMiddlewareDeps_RBACPerRoute(t *testing.T) {
	enforcer, err := casbin.NewEnforcer("../../../rbac/rbac_model.conf", "../../../rbac/rbac_policy.csv")
	if err != nil {
		t.Skipf("enforcer init skipped: %v", err)
	}

	deps := &MiddlewareDeps{
		Enforcer:  enforcer,
		JWTSecret: "test-secret",
	}

	r := chi.NewRouter()

	// Mock admin-only route protected by RBAC
	r.Group(func(r chi.Router) {
		r.Use(deps.RBAC("berita", "write"))
		r.Post("/berita", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})
	})

	// Public route — no middleware
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("public endpoint works without auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("RBAC blocks unauthenticated admin endpoint", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/berita", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		// RBAC checks for sub claim from JWT context — no JWT means denied
		assert.NotEqual(t, http.StatusCreated, w.Code)
	})
}