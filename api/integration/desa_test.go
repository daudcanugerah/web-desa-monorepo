package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertDesaShape asserts all fields of a desa response object.
// The profile no longer has id or created_at — it is stored as JSON in settings.
func assertDesaShape(t *testing.T, data map[string]any) {
	t.Helper()
	assert.NotEmpty(t, data["name"], "name must not be empty")
	assert.NotEmpty(t, data["updated_at"], "updated_at must not be empty")
	// description, address, phone, email, website, vision_mission are optional
}

// insertDesaSetting inserts a village profile directly into the settings table as JSON.
func insertDesaSetting(t *testing.T, ts *TestServer, name, address, phone, email, website, description string) {
	t.Helper()
	_, err := ts.DB.Exec(`
		INSERT INTO settings (key, value, updated_at)
		VALUES ('desa_profile', $1::jsonb, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, `{"name":"`+name+`","description":"`+description+`","address":"`+address+`","phone":"`+phone+`","email":"`+email+`","website":"`+website+`"}`)
	require.NoError(t, err)
}

func TestDesa_HappyPath(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	insertDesaSetting(t, ts, "Test Village", "Test Address", "081234567890",
		"test@village.com", "https://test.village.com", "Test Description")

	// Get Desa (public)
	resp, err := ts.MakeRequest("GET", "/api/v1/public/desa", nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	require.NotNil(t, data)
	assertDesaShape(t, data)
	assert.Equal(t, "Test Village", data["name"])
	assert.Equal(t, "Test Address", data["address"])
	assert.Equal(t, "081234567890", data["phone"])
	assert.Equal(t, "test@village.com", data["email"])
	assert.Equal(t, "https://test.village.com", data["website"])
	assert.Equal(t, "Test Description", data["description"])

	// Update Desa
	vm := "Vision: Maju. Mission: Melayani."
	updatePayload := map[string]any{
		"name":           "Updated Village",
		"address":        "Updated Address",
		"phone":          "081234567891",
		"email":          "updated@village.com",
		"website":        "https://updated.village.com",
		"description":    "Updated Description",
		"vision_mission": vm,
	}

	resp, err = ts.MakeRequest("PUT", "/api/v1/desa", updatePayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
	data = getDataField(body)
	require.NotNil(t, data)
	assertDesaShape(t, data)
	assert.Equal(t, "Updated Village", data["name"])
	assert.Equal(t, "Updated Address", data["address"])
	assert.Equal(t, "081234567891", data["phone"])
	assert.Equal(t, "updated@village.com", data["email"])
	assert.Equal(t, "https://updated.village.com", data["website"])
	assert.Equal(t, "Updated Description", data["description"])
	assert.Equal(t, vm, data["vision_mission"])

	// Verify update via GET
	resp, err = ts.MakeRequest("GET", "/api/v1/public/desa", nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
	data = getDataField(body)
	require.NotNil(t, data)
	assertDesaShape(t, data)
	assert.Equal(t, "Updated Village", data["name"])
	assert.Equal(t, vm, data["vision_mission"])
}

func TestDesa_UpdateWithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	insertDesaSetting(t, ts, "Test Village", "Test Address", "081234567890",
		"test@village.com", "https://test.village.com", "Test Description")

	resp, err := ts.MakeRequest("PUT", "/api/v1/desa",
		map[string]any{"name": "Updated Village"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestDesa_UpdateWithEmptyName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	insertDesaSetting(t, ts, "Test Village", "Test Address", "081234567890",
		"test@village.com", "https://test.village.com", "Test Description")

	resp, err := ts.MakeRequest("PUT", "/api/v1/desa",
		map[string]any{"name": ""}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestDesa_UpdateWithInvalidEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	insertDesaSetting(t, ts, "Test Village", "Test Address", "081234567890",
		"test@village.com", "https://test.village.com", "Test Description")

	resp, err := ts.MakeRequest("PUT", "/api/v1/desa",
		map[string]any{"name": "Updated Village", "email": "not-an-email"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestDesa_UpdateWithInvalidWebsite(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	insertDesaSetting(t, ts, "Test Village", "Test Address", "081234567890",
		"test@village.com", "https://test.village.com", "Test Description")

	resp, err := ts.MakeRequest("PUT", "/api/v1/desa",
		map[string]any{"name": "Updated Village", "website": "not-a-url"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}
