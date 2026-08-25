package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInfographicCategory_List_Success verifies the public list endpoint.
func TestInfographicCategory_List_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestInfographicCategory(t, "Statistik")
	ts.CreateTestInfographicCategory(t, "Keuangan")

	resp, err := ts.MakeRequest("GET", "/api/v1/public/infographic/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})
	assert.Len(t, cats, 2)
	// Sorted alphabetically
	assert.Equal(t, "Keuangan", cats[0].(map[string]interface{})["name"])
	assert.Equal(t, "Statistik", cats[1].(map[string]interface{})["name"])
}

// TestInfographicCategory_SearchAutocomplete verifies text search.
func TestInfographicCategory_SearchAutocomplete(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestInfographicCategory(t, "Statistik")
	ts.CreateTestInfographicCategory(t, "Statistik Pendidikan")
	ts.CreateTestInfographicCategory(t, "Keuangan")

	resp, err := ts.MakeRequest("GET", "/api/v1/public/infographic/categories?q=stat", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})
	assert.Len(t, cats, 2)
}

// TestInfographicCategory_Create_Success verifies create.
func TestInfographicCategory_Create_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic/categories",
		map[string]string{"name": "Custom"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestInfographicCategory_Create_WithoutAuth verifies the endpoint is protected.
func TestInfographicCategory_Create_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic/categories",
		map[string]string{"name": "Custom"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestInfographicCategory_Delete_InUse returns 409 when referenced.
func TestInfographicCategory_Delete_InUse(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	categoryID := ts.CreateTestInfographicCategory(t, "InUse")
	// Create an infographic referencing the category
	_, err := ts.DB.Exec(
		`INSERT INTO infographic (id, component_id, component_type, section_name, section_endpoint, category, state, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`,
		uuid.New().String(), 1, "dashboard", "Section", "/section", categoryID, true,
	)
	require.NoError(t, err)

	resp, err := ts.MakeRequest("DELETE", "/api/v1/infographic/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestInfographicCategory_Delete_Success verifies delete.
func TestInfographicCategory_Delete_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	categoryID := ts.CreateTestInfographicCategory(t, "Unused")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/infographic/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
