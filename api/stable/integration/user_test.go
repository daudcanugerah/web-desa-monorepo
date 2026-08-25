package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertUserShape validates all fields of a user response object
func assertUserShape(t *testing.T, u map[string]interface{}) {
	t.Helper()
	assert.NotEmpty(t, u["id"], "user.id must not be empty")
	assert.NotEmpty(t, u["name"], "user.name must not be empty")
	assert.NotEmpty(t, u["email"], "user.email must not be empty")
	assert.NotEmpty(t, u["created_at"], "user.created_at must not be empty")
	assert.NotEmpty(t, u["updated_at"], "user.updated_at must not be empty")
	// roles must be present and be a slice (never null)
	roles, ok := u["roles"]
	assert.True(t, ok, "user.roles field must be present")
	assert.NotNil(t, roles, "user.roles must not be null")
	_, isSlice := roles.([]interface{})
	assert.True(t, isSlice, "user.roles must be an array")
}

func TestCreateUser_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	createReq := map[string]interface{}{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/users", createReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assertUserShape(t, data)
	assert.Equal(t, "John Doe", data["name"])
	assert.Equal(t, "john@example.com", data["email"])
	roles := data["roles"].([]interface{})
	assert.Empty(t, roles, "newly created user should have no roles")
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	email := "duplicate@example.com"
	ts.CreateTestUser(t, "First User", email, "password123")

	createReq := map[string]interface{}{
		"name":     "Second User",
		"email":    email,
		"password": "password456",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/users", createReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Equal(t, "email already exists", body["error"])
}

func TestCreateUser_MissingFields(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	tests := []struct {
		name    string
		request map[string]interface{}
	}{
		{"missing name", map[string]interface{}{"email": "a@b.com", "password": "pass123"}},
		{"missing email", map[string]interface{}{"name": "Test", "password": "pass123"}},
		{"missing password", map[string]interface{}{"name": "Test", "email": "a@b.com"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := ts.MakeRequest("POST", "/api/v1/users", tc.request, token)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			body, err := parseJSONResponse(resp.Body)
			require.NoError(t, err)
			assert.False(t, body["success"].(bool))
			assert.NotEmpty(t, body["error"])
		})
	}
}

func TestListUsers_WithPagination(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	for i := 1; i <= 15; i++ {
		ts.CreateTestUser(t, fmt.Sprintf("User %d", i), fmt.Sprintf("user%d@example.com", i), "password123")
	}

	resp, err := ts.MakeRequest("GET", "/api/v1/users?page=1&limit=10", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)

	users := data["users"].([]interface{})
	assert.Len(t, users, 10)
	// assert every user has the full shape
	for _, u := range users {
		assertUserShape(t, u.(map[string]interface{}))
	}

	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(10), pagination["limit"])
	assert.Equal(t, float64(16), pagination["total"])
	assert.Equal(t, float64(2), pagination["total_pages"])
}

func TestListUsers_SecondPage(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	for i := 1; i <= 15; i++ {
		ts.CreateTestUser(t, fmt.Sprintf("User %d", i), fmt.Sprintf("user%d@example.com", i), "password123")
	}

	resp, err := ts.MakeRequest("GET", "/api/v1/users?page=2&limit=10", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := getDataField(body)
	users := data["users"].([]interface{})
	assert.Len(t, users, 6)
	for _, u := range users {
		assertUserShape(t, u.(map[string]interface{}))
	}

	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(2), pagination["page"])
	assert.Equal(t, float64(16), pagination["total"])
}

func TestListUsers_EmptyResult(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/users?page=1&limit=10", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := getDataField(body)
	users := data["users"].([]interface{})
	assert.Len(t, users, 1)

	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["total"])
	assert.Equal(t, float64(1), pagination["total_pages"])
}

func TestGetUser_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "Test User", "test@example.com", "password123")

	resp, err := ts.MakeRequest("GET", fmt.Sprintf("/api/v1/users/%s", userID), nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assertUserShape(t, data)
	assert.Equal(t, userID, data["id"])
	assert.Equal(t, "Test User", data["name"])
	assert.Equal(t, "test@example.com", data["email"])
	roles := data["roles"].([]interface{})
	assert.Empty(t, roles)
}

func TestGetUser_WithRole(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "Admin User", "admin@example.com", "password123")
	ts.AssignAdminRole(t, "admin@example.com")

	resp, err := ts.MakeRequest("GET", fmt.Sprintf("/api/v1/users/%s", userID), nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := getDataField(body)
	assertUserShape(t, data)
	roles := data["roles"].([]interface{})
	assert.NotEmpty(t, roles, "user with admin role should have roles populated")
	assert.Equal(t, "admin", roles[0].(string))
}

func TestGetUser_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/users/00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Equal(t, "User not found", body["error"])
}

func TestUpdateUser_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "Original Name", "original@example.com", "password123")

	updateReq := map[string]interface{}{
		"name":  "Updated Name",
		"email": "updated@example.com",
	}

	resp, err := ts.MakeRequest("PUT", fmt.Sprintf("/api/v1/users/%s", userID), updateReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assertUserShape(t, data)
	assert.Equal(t, userID, data["id"])
	assert.Equal(t, "Updated Name", data["name"])
	assert.Equal(t, "updated@example.com", data["email"])
}

func TestUpdateUser_DuplicateEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.CreateTestUser(t, "User One", "one@example.com", "password123")
	userTwoID := ts.CreateTestUser(t, "User Two", "two@example.com", "password123")

	updateReq := map[string]interface{}{
		"name":  "User Two",
		"email": "one@example.com",
	}

	resp, err := ts.MakeRequest("PUT", fmt.Sprintf("/api/v1/users/%s", userTwoID), updateReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Equal(t, "email already exists", body["error"])
}

func TestUpdateUser_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	updateReq := map[string]interface{}{
		"name":  "Ghost",
		"email": "ghost@example.com",
	}

	resp, err := ts.MakeRequest("PUT", "/api/v1/users/00000000-0000-0000-0000-000000000000", updateReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Equal(t, "User not found", body["error"])
}

func TestUpdateUser_MissingFields(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "Test", "test@example.com", "password123")

	tests := []struct {
		name    string
		request map[string]interface{}
	}{
		{"missing name", map[string]interface{}{"email": "a@b.com"}},
		{"missing email", map[string]interface{}{"name": "Test"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := ts.MakeRequest("PUT", fmt.Sprintf("/api/v1/users/%s", userID), tc.request, token)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			body, err := parseJSONResponse(resp.Body)
			require.NoError(t, err)
			assert.False(t, body["success"].(bool))
			assert.NotEmpty(t, body["error"])
		})
	}
}

func TestDeleteUser_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "To Delete", "delete@example.com", "password123")

	resp, err := ts.MakeRequest("DELETE", fmt.Sprintf("/api/v1/users/%s", userID), nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Equal(t, "User deleted successfully", data["message"])
}

func TestDeleteUser_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/users/00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Equal(t, "User not found", body["error"])
}

func TestGetUser_InvalidID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/users/not-a-uuid", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.NotEmpty(t, body["error"])
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/users", map[string]interface{}{
		"name": "Test", "email": "not-an-email", "password": "password123",
	}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
	assert.NotEmpty(t, body["error"])
}

func TestListUsers_InvalidPage(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/users?page=0", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestListUsers_ExceedsMaxLimit(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/users?limit=101", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestGetCurrentUser_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Me User", "me@example.com", "password123")
	token := ts.GetAuthToken(t, "me@example.com", "password123")

	resp, err := ts.MakeRequest("GET", "/api/v1/users/me", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assertUserShape(t, data)
	assert.Equal(t, "Me User", data["name"])
	assert.Equal(t, "me@example.com", data["email"])
	roles := data["roles"].([]interface{})
	assert.Empty(t, roles)
}

func TestGetCurrentUser_WithRoles(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin Me", "adminme@example.com", "password123")
	ts.AssignAdminRole(t, "adminme@example.com")
	token := ts.GetAuthToken(t, "adminme@example.com", "password123")

	resp, err := ts.MakeRequest("GET", "/api/v1/users/me", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := getDataField(body)
	assertUserShape(t, data)
	roles := data["roles"].([]interface{})
	assert.NotEmpty(t, roles)
	assert.Equal(t, "admin", roles[0].(string))
}

func TestGetCurrentUser_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/users/me", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
	assert.NotEmpty(t, body["error"])
}

func TestGetCurrentUser_WithInvalidToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/users/me", nil, "invalid.token.here")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestUpdateCurrentUserProfile_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Old Name", "profile@example.com", "password123")
	token := ts.GetAuthToken(t, "profile@example.com", "password123")

	updateReq := map[string]interface{}{
		"name":  "New Name",
		"email": "profile@example.com",
	}

	resp, err := ts.MakeRequest("PUT", "/api/v1/users/me", updateReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assertUserShape(t, data)
	assert.Equal(t, "New Name", data["name"])
}

func TestUpdateCurrentUserProfile_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("PUT", "/api/v1/users/me", map[string]interface{}{"name": "X", "email": "x@x.com"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

// TestUpdateCurrentUserProfile_WithProfileImage_PreservesName verifies that updating a user's profile
// with a profile image upload preserves the user's name. This test catches a regression where
// UpdateProfileImage was calling UpdateProfile with an empty string for the name, causing the
// user's name to be cleared.
func TestUpdateCurrentUserProfile_WithProfileImage_PreservesName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	originalName := "Original Name"
	ts.CreateTestUser(t, originalName, "imageprofile@example.com", "password123")
	token := ts.GetAuthToken(t, "imageprofile@example.com", "password123")

	// Create a minimal valid JPEG image (1x1 pixel)
	// This is the smallest valid JPEG file
	jpegData := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
		0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
		0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
		0x09, 0x08, 0x0A, 0x0C, 0x14, 0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12,
		0x13, 0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A, 0x1C, 0x1C, 0x20,
		0x24, 0x2E, 0x27, 0x20, 0x22, 0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29,
		0x2C, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39, 0x3D, 0x38, 0x32,
		0x3C, 0x2E, 0x33, 0x34, 0x32, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x01,
		0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xC4, 0x00, 0x1F, 0x00, 0x00,
		0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0xFF, 0xC4, 0x00, 0xB5, 0x10, 0x00, 0x02, 0x01, 0x03,
		0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7D,
		0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06,
		0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xA1, 0x08,
		0x23, 0x42, 0xB1, 0xC1, 0x15, 0x52, 0xD1, 0xF0, 0x24, 0x33, 0x62, 0x72,
		0x82, 0x09, 0x0A, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x25, 0x26, 0x27, 0x28,
		0x29, 0x2A, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x43, 0x44, 0x45,
		0x46, 0x47, 0x48, 0x49, 0x4A, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
		0x5A, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6A, 0x73, 0x74, 0x75,
		0x76, 0x77, 0x78, 0x79, 0x7A, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
		0x8A, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0xA2, 0xA3,
		0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6,
		0xB7, 0xB8, 0xB9, 0xBA, 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9,
		0xCA, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xE1, 0xE2,
		0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xF1, 0xF2, 0xF3, 0xF4,
		0xF5, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01,
		0x00, 0x00, 0x3F, 0x00, 0xFB, 0xD0, 0xFF, 0xD9,
	}

	// Update profile with image upload
	resp, err := ts.MakeMultipartRequest("PUT", "/api/v1/users/me", map[string]string{
		"name":  originalName,
		"email": "imageprofile@example.com",
	}, map[string][2]string{
		"profile_image": {"test_profile.jpg", string(jpegData)},
	}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assertUserShape(t, data)
	// CRITICAL: Verify name is preserved and not set to empty string
	assert.Equal(t, originalName, data["name"], "User name should be preserved after profile image update")
	assert.Equal(t, "imageprofile@example.com", data["email"])
	// Verify profile image URL is set
	assert.NotNil(t, data["profile_image_url"], "Profile image URL should be set after upload")
}

func TestUpdatePassword_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "Test User", "password@example.com", "oldpassword123")

	updateReq := map[string]interface{}{
		"old_password": "oldpassword123",
		"new_password": "newpassword456",
	}

	resp, err := ts.MakeRequest("PUT", fmt.Sprintf("/api/v1/users/%s/password", userID), updateReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Equal(t, "Password updated successfully", data["message"])
}

func TestUpdatePassword_InvalidOldPassword(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "Test User", "password@example.com", "correctpassword")

	updateReq := map[string]interface{}{
		"old_password": "wrongpassword",
		"new_password": "newpassword456",
	}

	resp, err := ts.MakeRequest("PUT", fmt.Sprintf("/api/v1/users/%s/password", userID), updateReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Equal(t, "Invalid old password", body["error"])
}

func TestUpdatePassword_MissingNewPassword(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "Test User", "password@example.com", "oldpassword")

	updateReq := map[string]interface{}{
		"old_password": "oldpassword",
	}

	resp, err := ts.MakeRequest("PUT", fmt.Sprintf("/api/v1/users/%s/password", userID), updateReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.NotEmpty(t, body["error"])
}

func TestUpdatePassword_NewPasswordTooShort(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "Test User", "password@example.com", "oldpassword")

	updateReq := map[string]interface{}{
		"old_password": "oldpassword",
		"new_password": "short",
	}

	resp, err := ts.MakeRequest("PUT", fmt.Sprintf("/api/v1/users/%s/password", userID), updateReq, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.NotEmpty(t, body["error"])
}
