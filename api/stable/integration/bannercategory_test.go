package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a banner category via API and return the ID.
func (ts *TestServer) createBannerCategory(t *testing.T, name string) string {
	resp, err := ts.MakeRequest("POST", "/api/v1/banners/categories",
		map[string]string{"name": name}, ts.AdminToken(t))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create banner category %q", name)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	id, ok := data["id"].(string)
	require.True(t, ok, "banner category id missing: %+v", data)
	return id
}

// TestBannerCategory_HappyPath covers the standard CRUD + usage count flow.
func TestBannerCategory_HappyPath(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)
	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")
	_ = token

	// Use unique name to avoid 409 with seeded data
	id := ts.createBannerCategory(t, "HappyPathTestCat")

	// List (admin) — should include the new category
	resp, err := ts.MakeRequest("GET", "/api/v1/banners/categories", nil, ts.AdminToken(t))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	cats := body["data"].(map[string]interface{})["banner_categories"].([]interface{})
	require.GreaterOrEqual(t, len(cats), 1)
	// usage_count is 0 (no banners yet)
	found := false
	for _, c := range cats {
		m := c.(map[string]interface{})
		if m["id"] == id {
			assert.Equal(t, "HappyPathTestCat", m["name"])
			assert.Equal(t, float64(0), m["usage_count"])
			found = true
		}
	}
	assert.True(t, found, "created banner category not in list")

	// Public list
	resp, err = ts.MakeRequest("GET", "/api/v1/public/banner/categories", nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Delete (still 0 banners, so it should succeed)
	resp, err = ts.MakeRequest("DELETE", "/api/v1/banners/categories/"+id, nil, ts.AdminToken(t))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestBannerCategory_DuplicateName ensures the unique constraint is enforced
// at the application level — second Create with same name returns 409.
func TestBannerCategory_DuplicateName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)
	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")
	_ = ts.GetAuthToken(t, "admin@test.com", "password123")

	ts.createBannerCategory(t, "UniqueName")

	resp, err := ts.MakeRequest("POST", "/api/v1/banners/categories",
		map[string]string{"name": "UniqueName"}, ts.AdminToken(t))
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestBannerCategory_InUseDelete ensures 409 when banners still reference it.
func TestBannerCategory_InUseDelete(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)
	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")

	catID := ts.createBannerCategory(t, "InUseCategory")

	// Create a banner using this category
	bannerID := ts.CreateTestBannerWithCategory(t, "Test Banner", "Test desc", catID)
	_ = bannerID

	// Try to delete — should be 409
	resp, err := ts.MakeRequest("DELETE", "/api/v1/banners/categories/"+catID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestBannerCategory_NotFound ensures 404 on missing id.
func TestBannerCategory_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)
	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")
	_ = ts.GetAuthToken(t, "admin@test.com", "password123")

	resp, err := ts.MakeRequest("DELETE",
		"/api/v1/banners/categories/00000000-0000-0000-0000-000000000000",
		nil, ts.AdminToken(t))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestBannerCategory_PublicListNoAuth verifies the public list endpoint.
// We seed a category directly via the service so the test doesn't need auth.
func TestBannerCategory_PublicListNoAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Seed directly via service (no admin login required)
	_, err := ts.BannerCategoryService.Create(context.Background(), "PublicListTest")
	require.NoError(t, err)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/banner/categories", nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestBannerCategory_RequireAuth verifies admin endpoints require JWT.
func TestBannerCategory_RequireAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/banners/categories", nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestBanner_CreateWithCategory verifies FK validation in banner Create.
func TestBanner_CreateWithCategory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)
	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")
	_ = ts.GetAuthToken(t, "admin@test.com", "password123")

	catID := ts.createBannerCategory(t, "BannerTestCat")

	// Create banner with valid category — should succeed
	bannerID := ts.CreateTestBannerWithCategory(t, "Cat Banner", "Desc", catID)
	require.NotEmpty(t, bannerID)
}