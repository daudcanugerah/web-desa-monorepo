package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBeritaCategory_List_Success verifies the public autocomplete-friendly
// list endpoint returns categories in alphabetical order with usage counts.
func TestBeritaCategory_List_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Seed some categories
	ts.CreateTestBeritaCategory(t, "Berita Desa")
	ts.CreateTestBeritaCategory(t, "Pengumuman")
	ts.CreateTestBeritaCategory(t, "Kegiatan")

	// Public access (no auth) — should still work
	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))

	data := getDataField(body)
	require.NotNil(t, data)
	categories, ok := data["categories"].([]interface{})
	require.True(t, ok, "categories must be an array")
	assert.Len(t, categories, 3)

	// Verify each category has the expected shape
	for _, c := range categories {
		cat, ok := c.(map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, cat["id"], "id is required")
		assert.NotEmpty(t, cat["name"], "name is required")
		// usage_count should be present (zero is fine)
		assert.Contains(t, cat, "usage_count")
		assert.Contains(t, cat, "created_at")
		assert.Contains(t, cat, "updated_at")
	}

	// Should be sorted alphabetically (Berita Desa, Kegiatan, Pengumuman)
	assert.Equal(t, "Berita Desa", categories[0].(map[string]interface{})["name"])
	assert.Equal(t, "Kegiatan", categories[1].(map[string]interface{})["name"])
	assert.Equal(t, "Pengumuman", categories[2].(map[string]interface{})["name"])

	// Pagination metadata should be present
	pagination, ok := data["pagination"].(map[string]interface{})
	require.True(t, ok, "pagination must be present")
	assert.Contains(t, pagination, "page")
	assert.Contains(t, pagination, "limit")
	assert.Contains(t, pagination, "total")
	assert.Equal(t, float64(3), pagination["total"])
}

// TestBeritaCategory_SearchAutocomplete verifies the `q` query parameter
// does case-insensitive substring matching for autocomplete UIs.
func TestBeritaCategory_SearchAutocomplete(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestBeritaCategory(t, "Berita Desa")
	ts.CreateTestBeritaCategory(t, "Pengumuman")
	ts.CreateTestBeritaCategory(t, "Penyaluran")
	ts.CreateTestBeritaCategory(t, "Pemberitahuan")

	// Prefix search "pen" should match Pengumuman + Pemberitahuan
	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/categories?q=pen", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := parseJSONResponse(resp.Body)
	categories := getDataField(body)["categories"].([]interface{})
	assert.Len(t, categories, 2)

	// Substring "ber" should match Berita Desa + Pemberitahuan
	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/categories?q=ber", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ = parseJSONResponse(resp.Body)
	categories = getDataField(body)["categories"].([]interface{})
	assert.Len(t, categories, 2)

	// No matches
	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/categories?q=nonexistent", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ = parseJSONResponse(resp.Body)
	categories = getDataField(body)["categories"].([]interface{})
	assert.Len(t, categories, 0)
}

// TestBeritaCategory_Pagination verifies page/limit query params.
func TestBeritaCategory_Pagination(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	for _, n := range []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo"} {
		ts.CreateTestBeritaCategory(t, n)
	}

	// Page 1, limit 2 — should return Alpha, Bravo
	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/categories?page=1&limit=2", nil, "")
	require.NoError(t, err)
	body, _ := parseJSONResponse(resp.Body)
	data := getDataField(body)
	cats := data["categories"].([]interface{})
	assert.Len(t, cats, 2)
	assert.Equal(t, "Alpha", cats[0].(map[string]interface{})["name"])
	assert.Equal(t, "Bravo", cats[1].(map[string]interface{})["name"])

	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(2), pagination["limit"])
	assert.Equal(t, float64(5), pagination["total"])
	assert.Equal(t, float64(3), pagination["total_pages"])

	// Page 3, limit 2 — should return Echo
	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/categories?page=3&limit=2", nil, "")
	require.NoError(t, err)
	body, _ = parseJSONResponse(resp.Body)
	cats = getDataField(body)["categories"].([]interface{})
	assert.Len(t, cats, 1)
	assert.Equal(t, "Echo", cats[0].(map[string]interface{})["name"])
}

// TestBeritaCategory_Create_Success verifies the create endpoint with auth.
func TestBeritaCategory_Create_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/berita/categories",
		map[string]string{"name": "Konser"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, _ := parseJSONResponse(resp.Body)
	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Equal(t, "Konser", data["name"])
	assert.NotEmpty(t, data["id"])
}

// TestBeritaCategory_Create_WithoutAuth verifies the create endpoint is protected.
func TestBeritaCategory_Create_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("POST", "/api/v1/berita/categories",
		map[string]string{"name": "Konser"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestBeritaCategory_Create_DuplicateName returns 409 on duplicate names.
func TestBeritaCategory_Create_DuplicateName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.CreateTestBeritaCategory(t, "Konser")

	resp, err := ts.MakeRequest("POST", "/api/v1/berita/categories",
		map[string]string{"name": "Konser"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestBeritaCategory_Create_EmptyName returns 400 on empty name.
func TestBeritaCategory_Create_EmptyName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/berita/categories",
		map[string]string{"name": ""}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestBeritaCategory_Delete_InUse returns 409 when category is referenced.
func TestBeritaCategory_Delete_InUse(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create a category and a berita that uses it
	categoryID := ts.CreateTestBeritaCategory(t, "Pengumuman")
	ts.MakeMultipartRequest("POST", "/api/v1/berita",
		map[string]string{"title": "Test", "content": "Content", "category": categoryID},
		nil, token)

	// Attempt to delete the in-use category
	resp, err := ts.MakeRequest("DELETE", "/api/v1/berita/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestBeritaCategory_Delete_Success verifies deletion of an unused category.
func TestBeritaCategory_Delete_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	categoryID := ts.CreateTestBeritaCategory(t, "Unused")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/berita/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's gone
	resp, err = ts.MakeRequest("DELETE", "/api/v1/berita/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestBeritaCategory_Delete_NotFound returns 404 for unknown IDs.
func TestBeritaCategory_Delete_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/berita/categories/00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestBeritaCategory_Delete_InvalidID returns 400 for non-UUID IDs.
func TestBeritaCategory_Delete_InvalidID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/berita/categories/not-a-uuid", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestBeritaCategory_UsageCount verifies the count is correct after creating berita.
func TestBeritaCategory_UsageCount(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat1 := ts.CreateTestBeritaCategory(t, "Pengumuman")
	_ = ts.CreateTestBeritaCategory(t, "Berita")

	// Create 2 berita in cat1, 1 in cat2
	for i := 0; i < 2; i++ {
		ts.MakeMultipartRequest("POST", "/api/v1/berita",
			map[string]string{"title": "Test", "content": "Content", "category": cat1},
			nil, token)
	}
	cat2 := ts.CreateTestBeritaCategory(t, "Kegiatan")
	ts.MakeMultipartRequest("POST", "/api/v1/berita",
		map[string]string{"title": "Test", "content": "Content", "category": cat2},
		nil, token)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})

	counts := map[string]int{}
	for _, c := range cats {
		cm := c.(map[string]interface{})
		counts[cm["name"].(string)] = int(cm["usage_count"].(float64))
	}
	assert.Equal(t, 2, counts["Pengumuman"])
	assert.Equal(t, 0, counts["Berita"])
	assert.Equal(t, 1, counts["Kegiatan"])
}
