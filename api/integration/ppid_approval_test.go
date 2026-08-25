package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPPIDApprovalWorkflow tests the complete PPID request approval workflow
func TestPPIDApprovalWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Setup: Create admin user and get token
	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestPPIDCategory(t, "Financial")

	// Step 1: Create a PPID document
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/ppid",
		map[string]string{
			"title":       "Budget Report 2024",
			"category":    cat0ID,
			"description": "Annual budget report for 2024",
		},
		map[string][2]string{"document": {"budget.pdf", "fake-pdf-content"}},
		adminToken,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := body["data"].(map[string]interface{})
	ppidID := data["id"].(string)

	// Step 2: Create a PPID request (public endpoint)
	requestPayload := map[string]any{
		"requester_name":  "John Doe",
		"requester_email": "john@example.com",
		"purpose":         "Need budget information",
	}

	resp, err = ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = body["data"].(map[string]interface{})
	requestID := data["id"].(string)
	assert.Equal(t, "pending", data["status"])

	// Step 3: Approve the request (admin endpoint)
	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/approve", nil, adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = body["data"].(map[string]interface{})

	// Verify approval response - the approve endpoint returns only
	// {status, message}; the access_token is delivered via email, not the body.
	assert.Equal(t, "approved", data["status"], "status must be approved")
	assert.NotEmpty(t, data["message"], "message must exist")
	assert.Contains(t, data["message"].(string), "email", "email should be sent")

	// Step 5: Revoke the request
	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/revoke", nil, adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = body["data"].(map[string]interface{})
	assert.Equal(t, "revoked", data["status"])
	assert.NotEmpty(t, data["revoked_at"], "revoked_at must exist")
	assert.NotEmpty(t, data["revoked_by"], "revoked_by must exist")
}

// TestPPIDRequestCreation tests creating a PPID request
func TestPPIDRequestCreation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestPPIDCategory(t, "Test")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/ppid",
		map[string]string{"title": "Test Document", "category": cat0ID},
		map[string][2]string{"document": {"test.pdf", "fake-pdf-content"}},
		adminToken,
	)
	require.NoError(t, err)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := body["data"].(map[string]interface{})
	ppidID := data["id"].(string)

	requestPayload := map[string]any{
		"requester_name":  "Jane Doe",
		"requester_email": "jane@example.com",
		"purpose":         "Testing request creation",
	}

	resp, err = ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = body["data"].(map[string]interface{})
	assert.Equal(t, "Jane Doe", data["requester_name"])
	assert.Equal(t, "jane@example.com", data["requester_email"])
	assert.Equal(t, "pending", data["status"])
}

// NOTE: TestInvalidTokenDownload removed — the
// /api/v1/ppid/document/{id}/download?token=... route was removed in Task 7.4
// in favor of the unified signed-URL endpoint
// /api/v1/media/{id}/content?jwt=... See SignedURLService.Sign() in
// usecase/gallery/signed_url.go for the replacement test surface.

// TestApproveNonPendingRequest tests that approving a non-pending request fails
func TestApproveNonPendingRequest(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestPPIDCategory(t, "Test")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/ppid",
		map[string]string{"title": "Test Document", "category": cat0ID},
		map[string][2]string{"document": {"test.pdf", "fake-pdf-content"}},
		adminToken,
	)
	require.NoError(t, err)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := body["data"].(map[string]interface{})
	ppidID := data["id"].(string)

	requestPayload := map[string]any{
		"requester_name":  "Test User",
		"requester_email": "test@example.com",
		"purpose":         "Testing",
	}
	resp, err = ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload, "")
	require.NoError(t, err)
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = body["data"].(map[string]interface{})
	requestID := data["id"].(string)

	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/approve", nil, adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/approve", nil, adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestRevokeRequest tests revoking a request
func TestRevokeRequest(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestPPIDCategory(t, "Test")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/ppid",
		map[string]string{"title": "Test Document", "category": cat0ID},
		map[string][2]string{"document": {"test.pdf", "fake-pdf-content"}},
		adminToken,
	)
	require.NoError(t, err)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := body["data"].(map[string]interface{})
	ppidID := data["id"].(string)

	requestPayload := map[string]any{
		"requester_name":  "Test User",
		"requester_email": "test@example.com",
		"purpose":         "Testing",
	}
	resp, err = ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload, "")
	require.NoError(t, err)
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = body["data"].(map[string]interface{})
	requestID := data["id"].(string)

	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/revoke", nil, adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = body["data"].(map[string]interface{})
	assert.Equal(t, "revoked", data["status"])
}

// TestRevokeAlreadyRevokedRequest tests that revoking an already revoked request fails
func TestRevokeAlreadyRevokedRequest(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestPPIDCategory(t, "Test")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/ppid",
		map[string]string{"title": "Test Document", "category": cat0ID},
		map[string][2]string{"document": {"test.pdf", "fake-pdf-content"}},
		adminToken,
	)
	require.NoError(t, err)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	ppidID := body["data"].(map[string]interface{})["id"].(string)

	requestPayload := map[string]any{
		"requester_name":  "Test User",
		"requester_email": "test@example.com",
		"purpose":         "Testing",
	}
	resp, _ = ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload, "")
	body, _ = parseJSONResponse(resp.Body)
	requestID := body["data"].(map[string]interface{})["id"].(string)

	// First revoke succeeds
	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/revoke", nil, adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Second revoke fails
	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/revoke", nil, adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
	assert.Contains(t, body["error"].(string), "already revoked")
}

// TestCannotApproveRevokedRequest tests that approving a revoked request fails
func TestCannotApproveRevokedRequest(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestPPIDCategory(t, "Test")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/ppid",
		map[string]string{"title": "Test Document", "category": cat0ID},
		map[string][2]string{"document": {"test.pdf", "fake-pdf-content"}},
		adminToken,
	)
	require.NoError(t, err)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	ppidID := body["data"].(map[string]interface{})["id"].(string)

	requestPayload := map[string]any{
		"requester_name":  "Test User",
		"requester_email": "test@example.com",
		"purpose":         "Testing",
	}
	resp, _ = ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload, "")
	body, _ = parseJSONResponse(resp.Body)
	requestID := body["data"].(map[string]interface{})["id"].(string)

	// Revoke the request
	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/revoke", nil, adminToken)
	require.NoError(t, err)

	// Try to approve a revoked request - should fail
	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/approve", nil, adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
	assert.Contains(t, body["error"].(string), "revoked")
}

// NOTE: TestDownloadWithToken, TestDownloadWithRequestToken,
// TestDownloadWithoutToken, and TestExpiredTokenDownload removed — the
// /api/v1/ppid/document/{id}/download?token=... route was removed in Task 7.4
// in favor of the unified signed-URL endpoint
// /api/v1/media/{id}/content?jwt=... See SignedURLService.Sign() in
// usecase/gallery/signed_url.go for the replacement test surface.

// TestApproveWithoutPermission tests that non-admin cannot approve requests
func TestApproveWithoutPermission(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create non-admin user
	ts.CreateTestUser(t, "Regular User", "user@test.com", "password123")
	userToken := ts.GetAuthToken(t, "user@test.com", "password123")

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestPPIDCategory(t, "Test")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/ppid",
		map[string]string{"title": "Test Document", "category": cat0ID},
		map[string][2]string{"document": {"test.pdf", "fake-pdf-content"}},
		adminToken,
	)
	require.NoError(t, err)
	body, _ := parseJSONResponse(resp.Body)
	ppidID := body["data"].(map[string]interface{})["id"].(string)

	requestPayload := map[string]any{
		"requester_name":  "Test User",
		"requester_email": "test@example.com",
		"purpose":         "Testing",
	}
	resp, _ = ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload, "")
	body, _ = parseJSONResponse(resp.Body)
	requestID := body["data"].(map[string]interface{})["id"].(string)

	// Try to approve as non-admin user
	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/approve", nil, userToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

// TestRevokeWithoutPermission tests that non-admin cannot revoke requests
func TestRevokeWithoutPermission(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create non-admin user
	ts.CreateTestUser(t, "Regular User", "user@test.com", "password123")
	userToken := ts.GetAuthToken(t, "user@test.com", "password123")

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestPPIDCategory(t, "Test")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/ppid",
		map[string]string{"title": "Test Document", "category": cat0ID},
		map[string][2]string{"document": {"test.pdf", "fake-pdf-content"}},
		adminToken,
	)
	require.NoError(t, err)
	body, _ := parseJSONResponse(resp.Body)
	ppidID := body["data"].(map[string]interface{})["id"].(string)

	requestPayload := map[string]any{
		"requester_name":  "Test User",
		"requester_email": "test@example.com",
		"purpose":         "Testing",
	}
	resp, _ = ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload, "")
	body, _ = parseJSONResponse(resp.Body)
	requestID := body["data"].(map[string]interface{})["id"].(string)

	// Try to revoke as non-admin user
	resp, err = ts.MakeRequest("POST", "/api/v1/ppid/requests/"+requestID+"/revoke", nil, userToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

// TestListPPIDRequests tests listing all PPID requests (admin endpoint)
func TestListPPIDRequests(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestPPIDCategory(t, "Test")

	// Create PPID and request
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/ppid",
		map[string]string{"title": "Test Document", "category": cat0ID},
		map[string][2]string{"document": {"test.pdf", "fake-pdf-content"}},
		adminToken,
	)
	require.NoError(t, err)
	body, _ := parseJSONResponse(resp.Body)
	ppidID := body["data"].(map[string]interface{})["id"].(string)

	requestPayload := map[string]any{
		"requester_name":  "Requester One",
		"requester_email": "requester1@example.com",
		"purpose":         "Testing 1",
	}
	ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload, "")

	requestPayload2 := map[string]any{
		"requester_name":  "Requester Two",
		"requester_email": "requester2@example.com",
		"purpose":         "Testing 2",
	}
	ts.MakeRequest("POST", "/api/v1/public/ppid/"+ppidID+"/requests", requestPayload2, "")

	// List all requests
	resp, err = ts.MakeRequest("GET", "/api/v1/ppid/requests", nil, adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))

	responseData := body["data"].(map[string]interface{})
	require.NotNil(t, responseData, "data must not be nil")
	data := responseData["requests"].([]interface{})
	assert.Len(t, data, 2, "should have 2 requests")

	// CRITICAL: Verify ALL fields in EACH request item
	for i, reqItem := range data {
		item := reqItem.(map[string]interface{})
		assert.NotEmpty(t, item["id"], "request %d: id must exist", i)
		assert.NotEmpty(t, item["ppid_id"], "request %d: ppid_id must exist", i)
		assert.NotEmpty(t, item["requester_name"], "request %d: requester_name must exist", i)
		assert.NotEmpty(t, item["requester_email"], "request %d: requester_email must exist", i)
		assert.NotEmpty(t, item["status"], "request %d: status must exist", i)
		assert.NotEmpty(t, item["created_at"], "request %d: created_at must exist", i)
		// All timestamp fields must exist (can be null)
		_, hasApprovedAt := item["approved_at"]
		_, hasApprovedBy := item["approved_by"]
		_, hasRevokedAt := item["revoked_at"]
		_, hasRevokedBy := item["revoked_by"]
		assert.True(t, hasApprovedAt || item["approved_at"] == nil, "request %d: approved_at must exist", i)
		assert.True(t, hasApprovedBy || item["approved_by"] == nil, "request %d: approved_by must exist", i)
		assert.True(t, hasRevokedAt || item["revoked_at"] == nil, "request %d: revoked_at must exist", i)
		assert.True(t, hasRevokedBy || item["revoked_by"] == nil, "request %d: revoked_by must exist", i)
	}
}

// TestListPPIDRequestsWithoutAuth tests that unauthenticated request to list is rejected
func TestListPPIDRequestsWithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// List requests without auth
	resp, err := ts.MakeRequest("GET", "/api/v1/ppid/requests", nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

// NOTE: TestDownloadNonExistentDocument removed — the
// /api/v1/ppid/document/{id}/download?token=... route was removed in Task 7.4
// in favor of the unified signed-URL endpoint
// /api/v1/media/{id}/content?jwt=... See SignedURLService.Sign() in
// usecase/gallery/signed_url.go for the replacement test surface.
