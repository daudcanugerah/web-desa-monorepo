package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFasilitasCategory_List_Success verifies the public list endpoint.
func TestFasilitasCategory_List_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestFasilitasCategory(t, "School")
	ts.CreateTestFasilitasCategory(t, "Health Center")

	resp, err := ts.MakeRequest("GET", "/api/v1/public/fasilitas/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})
	assert.Len(t, cats, 2)
	// Sorted alphabetically
	assert.Equal(t, "Health Center", cats[0].(map[string]interface{})["name"])
	assert.Equal(t, "School", cats[1].(map[string]interface{})["name"])
}

// TestFasilitasCategory_SearchAutocomplete verifies text search.
func TestFasilitasCategory_SearchAutocomplete(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestFasilitasCategory(t, "School")
	ts.CreateTestFasilitasCategory(t, "Health Center")
	ts.CreateTestFasilitasCategory(t, "Mosque")

	resp, err := ts.MakeRequest("GET", "/api/v1/public/fasilitas/categories?q=sch", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})
	assert.Len(t, cats, 1)
	assert.Equal(t, "School", cats[0].(map[string]interface{})["name"])
}

// TestFasilitasCategory_Create_Success verifies create.
func TestFasilitasCategory_Create_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas/categories",
		map[string]string{"name": "Office"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestFasilitasCategory_Create_WithoutAuth verifies the endpoint is protected.
func TestFasilitasCategory_Create_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas/categories",
		map[string]string{"name": "Office"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestFasilitasCategory_Delete_InUse returns 409 when referenced.
func TestFasilitasCategory_Delete_InUse(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	categoryID := ts.CreateTestFasilitasCategory(t, "InUse")
	// Create a Fasilitas referencing the category
	_, err := ts.DB.Exec(
		`INSERT INTO fasilitas (id, name, category, latitude, longitude, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
		uuid.New().String(), "Test Facility", categoryID, -6.2088, 106.8456,
	)
	require.NoError(t, err)

	resp, err := ts.MakeRequest("DELETE", "/api/v1/fasilitas/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestFasilitasCategory_Delete_Success verifies delete.
func TestFasilitasCategory_Delete_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	categoryID := ts.CreateTestFasilitasCategory(t, "Unused")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/fasilitas/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestFasilitasCategory_UsageCount verifies the count.
func TestFasilitasCategory_UsageCount(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat1 := ts.CreateTestFasilitasCategory(t, "WithRefs")
	_ = ts.CreateTestFasilitasCategory(t, "NoRefs")

	// Insert 2 Fasilitas referencing cat1
	for _, id := range []string{uuid.New().String(), uuid.New().String()} {
		_, err := ts.DB.Exec(
			`INSERT INTO fasilitas (id, name, category, latitude, longitude, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
			id, "Test "+id, cat1, -6.2088, 106.8456,
		)
		require.NoError(t, err)
	}

	resp, err := ts.MakeRequest("GET", "/api/v1/public/fasilitas/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})

	counts := map[string]int{}
	for _, c := range cats {
		cm := c.(map[string]interface{})
		counts[cm["name"].(string)] = int(cm["usage_count"].(float64))
	}
	assert.Equal(t, 2, counts["WithRefs"])
	assert.Equal(t, 0, counts["NoRefs"])
}
