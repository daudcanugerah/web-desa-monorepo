package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfile_HappyPath(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create Profile
	createPayload := map[string]any{
		"content":          "Informasi keuangan desa tahun 2024",
		"section_name":     "keuangan",
		"section_endpoint": "/profile/keuangan",
		"state":            true,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/profile", createPayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "content", "Informasi keuangan desa tahun 2024")
	validator.AssertFieldValue(t, "section_name", "keuangan")
	validator.AssertFieldValue(t, "section_endpoint", "/profile/keuangan")
	validator.AssertFieldValue(t, "state", true)
	validator.AssertFieldIsUUID(t, "id")
	validator.AssertFieldIsTimestamp(t, "created_at")
	validator.AssertFieldIsTimestamp(t, "updated_at")
	profileID := validator.Data["id"].(string)

	// Get Profile by ID
	resp, err = ts.MakeRequest("GET", "/api/v1/profile/"+profileID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", profileID)
	validator.AssertFieldValue(t, "content", "Informasi keuangan desa tahun 2024")
	validator.AssertFieldValue(t, "section_name", "keuangan")

	// List Profiles
	resp, err = ts.MakeRequest("GET", "/api/v1/profile", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "profile")
	validator.AssertFieldExists(t, "pagination")
	validator.AssertFieldType(t, "profile", "array")

	profileList := validator.Data["profile"].([]interface{})
	assert.Len(t, profileList, 1)

	// Update Profile
	updatePayload := map[string]any{
		"content":          "Informasi keuangan desa tahun 2025",
		"section_name":     "keuangan",
		"section_endpoint": "/profile/keuangan",
		"state":            false,
	}

	resp, err = ts.MakeRequest("PUT", "/api/v1/profile/"+profileID, updatePayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "content", "Informasi keuangan desa tahun 2025")
	validator.AssertFieldValue(t, "state", false)

	// Delete Profile
	resp, err = ts.MakeRequest("DELETE", "/api/v1/profile/"+profileID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deletion
	resp, err = ts.MakeRequest("GET", "/api/v1/profile/"+profileID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProfile_GetSectionNames(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create multiple profiles with different section names
	sections := []string{"keuangan", "lokasi_desa", "potensi"}
	for _, section := range sections {
		payload := map[string]any{
			"content":          "Test content for " + section,
			"section_name":     section,
			"section_endpoint": "/profile/" + section,
			"state":            true,
		}
		resp, err := ts.MakeRequest("POST", "/api/v1/profile", payload, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get section names
	resp, err := ts.MakeRequest("GET", "/api/v1/profile/sections/names", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "section_names")
	validator.AssertFieldType(t, "section_names", "array")

	names := validator.Data["section_names"].([]interface{})
	assert.Len(t, names, 3)
}

func TestProfile_ListWithFilters(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create profiles with different states
	for i := 0; i < 3; i++ {
		payload := map[string]any{
			"content":          fmt.Sprintf("Test content filter %d", i),
			"section_name":     "keuangan",
			"section_endpoint": "/profile/keuangan",
			"state":            i%2 == 0, // Alternate true/false
		}
		resp, err := ts.MakeRequest("POST", "/api/v1/profile", payload, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List with state filter
	resp, err := ts.MakeRequest("GET", "/api/v1/profile?state=true", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	profileList := validator.Data["profile"].([]interface{})
	// Should have at least 2 items with state=true from this test
	assert.GreaterOrEqual(t, len(profileList), 2)
}

func TestProfile_List_FilterByQuery(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	payloads := []map[string]any{
		{
			"content":          "Informasi anggaran dan belanja desa tahun 2024",
			"section_name":     "keuangan",
			"section_endpoint": "/profile/keuangan",
			"state":            true,
		},
		{
			"content":          "Peta dan lokasi geografis desa",
			"section_name":     "lokasi_desa",
			"section_endpoint": "/profile/lokasi_desa",
			"state":            true,
		},
		{
			"content":          "Daftar potensi unggulan desa",
			"section_name":     "potensi",
			"section_endpoint": "/profile/potensi",
			"state":            true,
		},
	}
	for _, p := range payloads {
		resp, err := ts.MakeRequest("POST", "/api/v1/profile", p, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	resp, err := ts.MakeRequest("GET", "/api/v1/profile?q=lokasi", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list := validator.Data["profile"].([]interface{})
	assert.GreaterOrEqual(t, len(list), 1, "q=lokasi should match at least 1 row")
	found := false
	for _, item := range list {
		if item.(map[string]interface{})["section_name"] == "lokasi_desa" {
			found = true
			break
		}
	}
	assert.True(t, found, "lokasi_desa should be in results")

	resp, err = ts.MakeRequest("GET", "/api/v1/profile?q=anggaran", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list = validator.Data["profile"].([]interface{})
	assert.GreaterOrEqual(t, len(list), 1, "q=anggaran should match at least 1 row")
	found = false
	for _, item := range list {
		if item.(map[string]interface{})["section_name"] == "keuangan" {
			found = true
			break
		}
	}
	assert.True(t, found, "keuangan should be in results")

	resp, err = ts.MakeRequest("GET", "/api/v1/profile?q=desa", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list = validator.Data["profile"].([]interface{})
	assert.GreaterOrEqual(t, len(list), 3, "q=desa should match at least 3 rows (substring present in every section_name and content)")

	resp, err = ts.MakeRequest("GET", "/api/v1/profile?q=nonexistent-xyz", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	rawList := validator.Data["profile"]
	if rawList == nil {
		rawList = []interface{}{}
	}
	list = rawList.([]interface{})
	assert.Len(t, list, 0, "non-matching q should return empty list")
}
func TestProfile_CreateWithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	payload := map[string]any{
		"content":          "Test content",
		"section_name":     "keuangan",
		"section_endpoint": "/profile/keuangan",
		"state":            true,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/profile", payload, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestProfile_CreateWithEmptyContent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	payload := map[string]any{
		"content":          "",
		"section_name":     "keuangan",
		"section_endpoint": "/profile/keuangan",
		"state":            true,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/profile", payload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestProfile_GetNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/profile/nonexistent-id", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
