package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertStrukturShape asserts all fields of a struktur response object
func assertStrukturShape(t *testing.T, data map[string]interface{}) {
	t.Helper()
	assert.NotEmpty(t, data["id"], "id must not be empty")
	assert.NotEmpty(t, data["name"], "name must not be empty")
	// position, email, phone, description, profile_image_url are nullable — keys may be absent or nil
	assert.NotEmpty(t, data["created_at"], "created_at must not be empty")
	assert.NotEmpty(t, data["updated_at"], "updated_at must not be empty")
}

func TestStruktur_HappyPath(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create Struktur
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/struktur",
		map[string]string{
			"name":     "Test Official",
			"position": "Village Head",
			"email":    "official@village.com",
			"phone":    "081234567890",
		},
		nil, token,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	require.NotNil(t, data)
	assertStrukturShape(t, data)
	assert.Equal(t, "Test Official", data["name"])
	assert.Equal(t, "Village Head", data["position"])
	assert.Equal(t, "official@village.com", data["email"])
	assert.Equal(t, "081234567890", data["phone"])
	strukturID := data["id"].(string)

	// List Struktur (public)
	resp, err = ts.MakeRequest("GET", "/api/v1/public/struktur/list", nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
	listData := getDataField(body)
	require.NotNil(t, listData)
	strukturList, ok := listData["struktur"].([]interface{})
	require.True(t, ok, "struktur field must be an array")
	assert.Len(t, strukturList, 1)
	assertStrukturShape(t, strukturList[0].(map[string]interface{}))
	pagination, ok := listData["pagination"].(map[string]interface{})
	require.True(t, ok, "pagination field must be present")
	assertPaginationShape(t, pagination)

	// Get Struktur by ID (public)
	resp, err = ts.MakeRequest("GET", "/api/v1/public/struktur/"+strukturID, nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
	data = getDataField(body)
	require.NotNil(t, data)
	assertStrukturShape(t, data)
	assert.Equal(t, strukturID, data["id"])
	assert.Equal(t, "Test Official", data["name"])

	// Update Struktur
	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/struktur/"+strukturID,
		map[string]string{
			"name":     "Updated Official",
			"position": "Deputy Village Head",
		},
		nil, token,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
	data = getDataField(body)
	require.NotNil(t, data)
	assertStrukturShape(t, data)
	assert.Equal(t, strukturID, data["id"])
	assert.Equal(t, "Updated Official", data["name"])
	assert.Equal(t, "Deputy Village Head", data["position"])

	// Delete Struktur
	resp, err = ts.MakeRequest("DELETE", "/api/v1/struktur/"+strukturID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
}

func TestStruktur_List_FilterByQuery(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	rows := []map[string]string{
		{"name": "Budi Santoso", "position": "Kepala Desa"},
		{"name": "Siti Aminah", "position": "Sekretaris Desa"},
		{"name": "Agus Wibowo", "position": "Bendahara Desa"},
		{"name": "Dewi Lestari", "position": "Kaur Pemerintahan"},
	}
	for _, r := range rows {
		resp, err := ts.MakeMultipartRequest("POST", "/api/v1/struktur", r, nil, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	resp, err := ts.MakeRequest("GET", "/api/v1/struktur?q=Budi", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list := validator.Data["struktur"].([]interface{})
	assert.Len(t, list, 1, "q=Budi should match exactly 1 row by name")
	first := list[0].(map[string]interface{})
	assert.Equal(t, "Budi Santoso", first["name"])

	resp, err = ts.MakeRequest("GET", "/api/v1/struktur?q=Desa", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list = validator.Data["struktur"].([]interface{})
	assert.Len(t, list, 3, "q=Desa should match positions containing 'Desa' (Kepala/Sekretaris/Bendahara)")

	resp, err = ts.MakeRequest("GET", "/api/v1/struktur?q=NoSuchName", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	rawList := validator.Data["struktur"]
	if rawList == nil {
		rawList = []interface{}{}
	}
	list = rawList.([]interface{})
	assert.Len(t, list, 0, "non-matching q should return empty list")

	resp, err = ts.MakeRequest("GET", "/api/v1/struktur?q=budi", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list = validator.Data["struktur"].([]interface{})
	assert.Len(t, list, 1, "q should be case-insensitive")
}
func TestStruktur_CreateWithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/struktur",
		map[string]string{"name": "Test Official"}, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestStruktur_CreateWithEmptyName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/struktur",
		map[string]string{"name": ""}, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestStruktur_GetNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/struktur/00000000-0000-0000-0000-000000000000", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestStruktur_GetWithInvalidID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/struktur/invalid-uuid", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestStruktur_UpdateNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeMultipartRequest("PUT", "/api/v1/struktur/00000000-0000-0000-0000-000000000000",
		map[string]string{"name": "Updated Official"}, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestStruktur_DeleteNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/struktur/00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}
