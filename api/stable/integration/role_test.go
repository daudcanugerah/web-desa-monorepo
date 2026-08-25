package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertPermissionShape validates a permission object has resource+action fields
func assertPermissionShape(t *testing.T, p map[string]interface{}) {
	t.Helper()
	assert.NotEmpty(t, p["resource"], "permission.resource must not be empty")
	assert.NotEmpty(t, p["action"], "permission.action must not be empty")
}

// assertRoleShape validates a role response object
func assertRoleShape(t *testing.T, r map[string]interface{}) {
	t.Helper()
	assert.NotEmpty(t, r["name"], "role.name must not be empty")
	perms, ok := r["permissions"]
	assert.True(t, ok, "role.permissions field must be present")
	assert.NotNil(t, perms, "role.permissions must not be null")
	_, isSlice := perms.([]interface{})
	assert.True(t, isSlice, "role.permissions must be an array")
	assert.NotEmpty(t, r["created_at"], "role.created_at must not be empty")
	assert.NotEmpty(t, r["updated_at"], "role.updated_at must not be empty")
}

func TestListRoles_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/roles", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	roles := body["data"].([]interface{})
	assert.NotNil(t, roles)
	for _, r := range roles {
		assertRoleShape(t, r.(map[string]interface{}))
	}
}

func TestCreateRole_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{
		"name": "editor",
	}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assertRoleShape(t, data)
	assert.Equal(t, "editor", data["name"])
	perms := data["permissions"].([]interface{})
	assert.Empty(t, perms, "new role with no permissions should have empty array")
}

func TestCreateRole_WithPermissions(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{
		"name": "writer",
		"permissions": []map[string]interface{}{
			{"resource": "berita", "action": "write"},
		},
	}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := getDataField(body)
	assertRoleShape(t, data)
	perms := data["permissions"].([]interface{})
	assert.Len(t, perms, 1)
	assertPermissionShape(t, perms[0].(map[string]interface{}))
	assert.Equal(t, "berita", perms[0].(map[string]interface{})["resource"])
	assert.Equal(t, "write", perms[0].(map[string]interface{})["action"])
}

func TestGetRolePermissions_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// create role with a permission
	ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{
		"name":        "viewer",
		"permissions": []map[string]interface{}{{"resource": "berita", "action": "read"}},
	}, token)

	resp, err := ts.MakeRequest("GET", "/api/v1/roles/viewer/permissions", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	perms := body["data"].([]interface{})
	assert.NotNil(t, perms)
	for _, p := range perms {
		assertPermissionShape(t, p.(map[string]interface{}))
	}
}

func TestAddPermissionToRole_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{"name": "editor"}, token)

	resp, err := ts.MakeRequest("POST", "/api/v1/roles/editor/permissions", map[string]interface{}{
		"resource": "berita",
		"action":   "write",
	}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Equal(t, "Permission added to role successfully", data["message"])
}

func TestGetAvailablePermissions_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// seed some permissions first
	ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{
		"name": "seeder",
		"permissions": []map[string]interface{}{
			{"resource": "berita", "action": "read"},
			{"resource": "umkm", "action": "write"},
		},
	}, token)

	resp, err := ts.MakeRequest("GET", "/api/v1/permissions", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	perms := body["data"].([]interface{})
	assert.NotNil(t, perms)
	assert.NotEmpty(t, perms)
	for _, p := range perms {
		assertPermissionShape(t, p.(map[string]interface{}))
	}
}

func TestGetAvailablePermissions_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/permissions", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestAssignRoleToUser_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{"name": "editor"}, token)
	userID := ts.CreateTestUser(t, "Editor", "editor@test.com", "password123")

	resp, err := ts.MakeRequest("POST", "/api/v1/users/"+userID+"/roles", map[string]interface{}{"role": "editor"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Equal(t, "Role assigned to user successfully", data["message"])

	// verify the role shows up on the user
	getResp, err := ts.MakeRequest("GET", "/api/v1/users/"+userID, nil, token)
	require.NoError(t, err)
	defer getResp.Body.Close()

	getBody, _ := parseJSONResponse(getResp.Body)
	userData := getDataField(getBody)
	roles := userData["roles"].([]interface{})
	assert.Contains(t, roles, "editor")
}

func TestRemoveRoleFromUser_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{"name": "editor"}, token)
	userID := ts.CreateTestUser(t, "Editor", "editor@test.com", "password123")
	ts.MakeRequest("POST", "/api/v1/users/"+userID+"/roles", map[string]interface{}{"role": "editor"}, token)

	resp, err := ts.MakeRequest("DELETE", "/api/v1/users/"+userID+"/roles/editor", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Equal(t, "Role removed from user successfully", data["message"])

	// verify role is gone from user
	getResp, err := ts.MakeRequest("GET", "/api/v1/users/"+userID, nil, token)
	require.NoError(t, err)
	defer getResp.Body.Close()

	getBody, _ := parseJSONResponse(getResp.Body)
	userData := getDataField(getBody)
	roles := userData["roles"].([]interface{})
	assert.NotContains(t, roles, "editor")
}

func TestDeleteRole_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{"name": "temp"}, token)

	resp, err := ts.MakeRequest("DELETE", "/api/v1/roles/temp", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Equal(t, "Role deleted successfully", data["message"])
}

func TestCreateRole_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{"name": "x"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	assert.False(t, body["success"].(bool))
}

func TestCreateRole_EmptyName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{"name": ""}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	assert.False(t, body["success"].(bool))
	assert.NotEmpty(t, body["error"])
}

func TestAddPermission_MissingResource(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.MakeRequest("POST", "/api/v1/roles", map[string]interface{}{"name": "editor"}, token)

	resp, err := ts.MakeRequest("POST", "/api/v1/roles/editor/permissions", map[string]interface{}{
		"resource": "",
		"action":   "write",
	}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	assert.False(t, body["success"].(bool))
}

func TestListRoles_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/roles", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	assert.False(t, body["success"].(bool))
}

func TestAssignRole_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	userID := ts.CreateTestUser(t, "Test", "test@test.com", "password123")

	resp, err := ts.MakeRequest("POST", "/api/v1/users/"+userID+"/roles", map[string]interface{}{"role": "editor"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	assert.False(t, body["success"].(bool))
}

func TestAssignRole_EmptyRole(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	userID := ts.CreateTestUser(t, "Test", "test@test.com", "password123")

	resp, err := ts.MakeRequest("POST", "/api/v1/users/"+userID+"/roles", map[string]interface{}{"role": ""}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	assert.False(t, body["success"].(bool))
}
