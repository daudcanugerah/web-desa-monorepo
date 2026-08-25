package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUMKMCategory_List_Success verifies the public list endpoint for UMKM.
func TestUMKMCategory_List_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUMKMCategory(t, "Kuliner")
	ts.CreateTestUMKMCategory(t, "Kerajinan")

	resp, err := ts.MakeRequest("GET", "/api/v1/public/umkm/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := parseJSONResponse(resp.Body)
	assert.True(t, body["success"].(bool))
	cats := getDataField(body)["categories"].([]interface{})
	assert.Len(t, cats, 2)
	// Sorted alphabetically
	assert.Equal(t, "Kerajinan", cats[0].(map[string]interface{})["name"])
	assert.Equal(t, "Kuliner", cats[1].(map[string]interface{})["name"])
}

// TestUMKMCategory_SearchAutocomplete verifies text search works for UMKM.
func TestUMKMCategory_SearchAutocomplete(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUMKMCategory(t, "Kuliner")
	ts.CreateTestUMKMCategory(t, "Kerajinan")
	ts.CreateTestUMKMCategory(t, "Pertanian")

	resp, err := ts.MakeRequest("GET", "/api/v1/public/umkm/categories?q=ker", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})
	assert.Len(t, cats, 1)
	assert.Equal(t, "Kerajinan", cats[0].(map[string]interface{})["name"])
}

// TestUMKMCategory_Create_Success verifies creating a UMKM category.
func TestUMKMCategory_Create_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/umkm/categories",
		map[string]string{"name": "Jasa"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestUMKMCategory_Create_WithoutAuth verifies the endpoint is protected.
func TestUMKMCategory_Create_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("POST", "/api/v1/umkm/categories",
		map[string]string{"name": "Jasa"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestUMKMCategory_Create_DuplicateName returns 409.
func TestUMKMCategory_Create_DuplicateName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.CreateTestUMKMCategory(t, "Jasa")
	resp, err := ts.MakeRequest("POST", "/api/v1/umkm/categories",
		map[string]string{"name": "Jasa"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestUMKMCategory_Delete_InUse returns 409 when in use.
func TestUMKMCategory_Delete_InUse(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create category and a UMKM that uses it.
	categoryID := ts.CreateTestUMKMCategory(t, "Kuliner")
	ts.MakeMultipartRequest("POST", "/api/v1/umkm",
		map[string]string{
			"name": "Warung Makan", "category": categoryID, "description": "Desc",
		},
		nil, token)

	// Attempt to delete the in-use category.
	resp, err := ts.MakeRequest("DELETE", "/api/v1/umkm/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestUMKMCategory_Delete_Success verifies deletion of an unused category.
func TestUMKMCategory_Delete_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	categoryID := ts.CreateTestUMKMCategory(t, "Unused")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/umkm/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestUMKMCategory_UsageCount verifies the usage count is accurate.
func TestUMKMCategory_UsageCount(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	kulinerID := ts.CreateTestUMKMCategory(t, "Kuliner")
	kerajinanID := ts.CreateTestUMKMCategory(t, "Kerajinan")

	// 2 UMKM in kuliner, 1 in kerajinan
	for i := 0; i < 2; i++ {
		ts.MakeMultipartRequest("POST", "/api/v1/umkm",
			map[string]string{
				"name":        "Test " + string(rune('A'+i)),
				"category":    kulinerID,
				"description": "Test",
			},
			nil, token)
	}
	ts.MakeMultipartRequest("POST", "/api/v1/umkm",
		map[string]string{
			"name": "Single", "category": kerajinanID, "description": "Test",
		},
		nil, token)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/umkm/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})

	counts := map[string]int{}
	for _, c := range cats {
		cm := c.(map[string]interface{})
		counts[cm["name"].(string)] = int(cm["usage_count"].(float64))
	}
	assert.Equal(t, 2, counts["Kuliner"])
	assert.Equal(t, 1, counts["Kerajinan"])
}
