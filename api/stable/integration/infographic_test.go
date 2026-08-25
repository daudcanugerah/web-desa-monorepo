package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfographic_HappyPath(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create Infographic
	createPayload := map[string]any{
		"component_id":     12,
		"component_type":   "dashboard",
		"section_name":     "keuangan",
		"section_endpoint": "/profile/keuangan",
		"state":            true,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic", createPayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "component_id", float64(12))
	validator.AssertFieldValue(t, "component_type", "dashboard")
	validator.AssertFieldValue(t, "section_name", "keuangan")
	validator.AssertFieldValue(t, "section_endpoint", "/profile/keuangan")
	validator.AssertFieldValue(t, "state", true)
	validator.AssertFieldIsUUID(t, "id")
	validator.AssertFieldIsTimestamp(t, "created_at")
	validator.AssertFieldIsTimestamp(t, "updated_at")
	// Token should NOT be present in create response
	validator.AssertFieldNotExists(t, "token")
	infographicID := validator.Data["id"].(string)

	// Get Infographic by ID
	resp, err = ts.MakeRequest("GET", "/api/v1/infographic/"+infographicID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", infographicID)
	validator.AssertFieldValue(t, "component_id", "12")
	validator.AssertFieldValue(t, "component_type", "dashboard")
	validator.AssertFieldValue(t, "section_name", "keuangan")
	// Token SHOULD be present in GET response
	validator.AssertFieldExists(t, "token")

	// List Infographics
	resp, err = ts.MakeRequest("GET", "/api/v1/infographic", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "infographic")
	validator.AssertFieldExists(t, "pagination")
	validator.AssertFieldType(t, "infographic", "array")

	infographicList := validator.Data["infographic"].([]interface{})
	assert.Len(t, infographicList, 1)

	// Update Infographic
	updatePayload := map[string]any{
		"component_id":     456,
		"component_type":   "question",
		"section_name":     "keuangan",
		"section_endpoint": "/profile/keuangan",
		"state":            false,
	}

	resp, err = ts.MakeRequest("PUT", "/api/v1/infographic/"+infographicID, updatePayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "component_id", float64(456))
	validator.AssertFieldValue(t, "component_type", "question")
	validator.AssertFieldValue(t, "state", false)

	// Delete Infographic
	resp, err = ts.MakeRequest("DELETE", "/api/v1/infographic/"+infographicID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deletion
	resp, err = ts.MakeRequest("GET", "/api/v1/infographic/"+infographicID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestInfographic_GetSectionNames(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create multiple infographics with different section names
	sections := []string{"keuangan", "lokasi_desa", "potensi"}
	for i, section := range sections {
		payload := map[string]any{
			"component_id":     int64(i + 1),
			"component_type":   "dashboard",
			"section_name":     section,
			"section_endpoint": "/profile/" + section,
			"state":            true,
		}
		resp, err := ts.MakeRequest("POST", "/api/v1/infographic", payload, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get section names
	resp, err := ts.MakeRequest("GET", "/api/v1/infographic/sections/names", nil, token)
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

func TestInfographic_ListWithFilters(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create infographics with different states
	for i := 0; i < 3; i++ {
		payload := map[string]any{
			"component_id":     int64(100 + i),
			"component_type":   "dashboard",
			"section_name":     "keuangan",
			"section_endpoint": "/profile/keuangan",
			"state":            i%2 == 0, // Alternate true/false
		}
		resp, err := ts.MakeRequest("POST", "/api/v1/infographic", payload, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List with state filter
	resp, err := ts.MakeRequest("GET", "/api/v1/infographic?state=true", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	infographicList := validator.Data["infographic"].([]interface{})
	// Should have at least 2 items with state=true from this test
	assert.GreaterOrEqual(t, len(infographicList), 2)

	// Verify token is NOT present in list responses
	for _, item := range infographicList {
		itemMap := item.(map[string]interface{})
		_, hasToken := itemMap["token"]
		assert.False(t, hasToken, "Token should not be present in list responses")
	}
}

func TestInfographic_List_FilterByQuery(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	payloads := []map[string]any{
		{
			"component_id":     int64(911001),
			"component_type":   "dashboard",
			"section_name":     "keuangan",
			"section_endpoint": "/profile/keuangan",
			"state":            true,
		},
		{
			"component_id":     int64(911002),
			"component_type":   "question",
			"section_name":     "lokasi_desa",
			"section_endpoint": "/profile/lokasi_desa",
			"state":            true,
		},
		{
			"component_id":     int64(777003),
			"component_type":   "dashboard",
			"section_name":     "potensi",
			"section_endpoint": "/profile/potensi",
			"state":            true,
		},
	}
	for _, p := range payloads {
		resp, err := ts.MakeRequest("POST", "/api/v1/infographic", p, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	resp, err := ts.MakeRequest("GET", "/api/v1/infographic?q=911", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list := validator.Data["infographic"].([]interface{})
	assert.Len(t, list, 2, "q=911 should match two infographics via component_id substring")
	for _, item := range list {
		itemMap := item.(map[string]interface{})
		cidF, ok := itemMap["component_id"].(float64)
		require.True(t, ok, "component_id must serialize as number, got %T", itemMap["component_id"])
		cidStr := fmt.Sprintf("%d", int64(cidF))
		assert.Contains(t, cidStr, "911", "matched component_id should contain '911'")
	}

	resp, err = ts.MakeRequest("GET", "/api/v1/infographic?q=potensi", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list = validator.Data["infographic"].([]interface{})
	assert.Len(t, list, 1, "q=potensi should match only one infographic via section_name")
	first := list[0].(map[string]interface{})
	assert.Equal(t, "potensi", first["section_name"])

	resp, err = ts.MakeRequest("GET", "/api/v1/infographic?q=nonexistent-xyz", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	rawList := validator.Data["infographic"]
	if rawList == nil {
		rawList = []interface{}{}
	}
	list = rawList.([]interface{})
	assert.Len(t, list, 0, "non-matching q should return empty list")
}

func TestInfographic_List_FilterByCategory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat1ID := ts.CreateTestInfographicCategory(t, "Ekonomi")
	cat2ID := ts.CreateTestInfographicCategory(t, "Sosial")

	makePayload := func(cid int, catID string) map[string]any {
		return map[string]any{
			"component_id":     int64(cid),
			"component_type":   "dashboard",
			"section_name":     "keuangan",
			"section_endpoint": "/profile/keuangan",
			"category":         catID,
			"state":            true,
		}
	}
	for _, p := range []map[string]any{makePayload(101, cat1ID), makePayload(102, cat1ID), makePayload(103, cat2ID)} {
		resp, err := ts.MakeRequest("POST", "/api/v1/infographic", p, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	resp, err := ts.MakeRequest("GET", "/api/v1/infographic?category="+cat1ID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list := validator.Data["infographic"].([]interface{})
	assert.Len(t, list, 2, "filtering by cat1 should return 2 infographics")

	for _, item := range list {
		itemMap := item.(map[string]interface{})
		catObj, ok := itemMap["category"].(map[string]interface{})
		require.True(t, ok, "category should be an object")
		assert.Equal(t, cat1ID, catObj["id"], "returned category.id must match filter")
		assert.Equal(t, "Ekonomi", catObj["name"])
	}

	resp, err = ts.MakeRequest("GET", "/api/v1/infographic?category="+cat2ID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	list = validator.Data["infographic"].([]interface{})
	assert.Len(t, list, 1, "filtering by cat2 should return 1 infographic")
	only := list[0].(map[string]interface{})
	catObj, ok := only["category"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, cat2ID, catObj["id"])

	resp, err = ts.MakeRequest("GET", "/api/v1/infographic?category=00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	rawList := validator.Data["infographic"]
	if rawList == nil {
		rawList = []interface{}{}
	}
	list = rawList.([]interface{})
	assert.Len(t, list, 0, "unknown category should return empty list")
}
func TestInfographic_CreateWithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	payload := map[string]any{
		"component_id":     12,
		"component_type":   "dashboard",
		"section_name":     "keuangan",
		"section_endpoint": "/profile/keuangan",
		"state":            true,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic", payload, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestInfographic_CreateWithEmptyComponentID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	payload := map[string]any{
		"component_id":     0,
		"component_type":   "dashboard",
		"section_name":     "keuangan",
		"section_endpoint": "/profile/keuangan",
		"state":            true,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic", payload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestInfographic_GetNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/infographic/nonexistent-id", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestInfographic_MetabaseTokenGeneration(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create infographic with question type
	createPayload := map[string]any{
		"component_id":     42,
		"component_type":   "question",
		"section_name":     "keuangan",
		"section_endpoint": "/profile/keuangan",
		"state":            true,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic", createPayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	infographicID := validator.Data["id"].(string)

	// Get infographic and verify token is present
	resp, err = ts.MakeRequest("GET", "/api/v1/infographic/"+infographicID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "token")

	// Token should be a valid JWT format (header.payload.signature)
	tokenValue := validator.Data["token"].(string)
	parts := len(tokenValue)
	assert.Greater(t, parts, 0, "Token should not be empty")
	// JWT tokens have 3 parts separated by dots
	dotCount := 0
	for _, c := range tokenValue {
		if c == '.' {
			dotCount++
		}
	}
	assert.Equal(t, 2, dotCount, "Token should have 2 dots (3 parts)")
}

func TestInfographic_GeneratePreviewToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Generate preview token for dashboard
	previewPayload := map[string]any{
		"component_id":   12,
		"component_type": "dashboard",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic/preview/token", previewPayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "component_id", float64(12))
	validator.AssertFieldValue(t, "component_type", "dashboard")
	validator.AssertFieldExists(t, "token")

	// Verify token format
	tokenValue := validator.Data["token"].(string)
	dotCount := 0
	for _, c := range tokenValue {
		if c == '.' {
			dotCount++
		}
	}
	assert.Equal(t, 2, dotCount, "Token should have 2 dots (3 parts)")
}

func TestInfographic_GeneratePreviewToken_Question(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Generate preview token for question
	previewPayload := map[string]any{
		"component_id":   99,
		"component_type": "question",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic/preview/token", previewPayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "component_id", float64(99))
	validator.AssertFieldValue(t, "component_type", "question")
	validator.AssertFieldExists(t, "token")
}

func TestInfographic_GeneratePreviewToken_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	previewPayload := map[string]any{
		"component_id":   12,
		"component_type": "dashboard",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic/preview/token", previewPayload, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestInfographic_GeneratePreviewToken_InvalidComponentType(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	previewPayload := map[string]any{
		"component_id":   12,
		"component_type": "invalid",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic/preview/token", previewPayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestInfographic_GeneratePreviewToken_EmptyComponentID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	previewPayload := map[string]any{
		"component_id":   0,
		"component_type": "dashboard",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/infographic/preview/token", previewPayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestInfographic_PublicAccessLog verifies the audit log is written for
// every public token issuance and is queryable via the admin endpoint.
func TestInfographic_PublicAccessLog(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	createPayload := map[string]any{
		"component_id":     42,
		"component_type":   "dashboard",
		"section_name":     "stats",
		"section_endpoint": "/stats",
		"state":            true,
	}
	resp, err := ts.MakeRequest("POST", "/api/v1/infographic", createPayload, adminToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	infID := getDataField(body)["id"].(string)

	// Hit the public endpoint 3 times — should produce 3 audit rows
	for i := 0; i < 3; i++ {
		resp, err := ts.MakeRequest("GET", "/api/v1/public/infographic/"+infID, nil, "")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Admin can read the audit log
	resp, err = ts.MakeRequest("GET", "/api/v1/infographic/access-logs?infographic_id="+infID, nil, adminToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ = parseJSONResponse(resp.Body)

	// Also query DB directly to see what's there
	var dbCount int
	if err := ts.DB.Get(&dbCount, "SELECT COUNT(*) FROM infographic_access_log WHERE infographic_id = $1", infID); err == nil {
		t.Logf("DB has %d access log rows for %s", dbCount, infID)
	} else {
		t.Logf("DB query error: %v", err)
	}
	rows, _ := ts.DB.Query("SELECT id, infographic_id, endpoint, user_agent FROM infographic_access_log WHERE infographic_id = $1", infID)
	defer rows.Close()
	for rows.Next() {
		var id, iid, ep, ua string
		_ = rows.Scan(&id, &iid, &ep, &ua)
		t.Logf("row: id=%s inf_id=%s endpoint=%s ua=%q", id, iid, ep, ua)
	}

	data := getDataField(body)
	t.Logf("response body: %+v", body)
	logsRaw, hasLogs := data["logs"]
	if !hasLogs {
		t.Logf("body: %+v", body)
		t.Fatal("logs key missing in access log response")
	}
	logs := logsRaw.([]interface{})
	t.Logf("got %d access log entries", len(logs))
	assert.GreaterOrEqual(t, len(logs), 3, "at least 3 public access entries should be logged")

	// Each entry has expected fields
	first := logs[0].(map[string]interface{})
	assert.Equal(t, infID, first["infographic_id"])
	assert.Equal(t, "public_detail", first["endpoint"])
	assert.NotEmpty(t, first["ip_address"], "ip_address must be recorded")
}

// TestInfographic_PerComponentRateLimit verifies that hitting the public
// endpoint >30 times in an hour from the same IP for the same infographic
// returns 429.
func TestInfographic_PerComponentRateLimit(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	adminToken := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	createPayload := map[string]any{
		"component_id":     7,
		"component_type":   "dashboard",
		"section_name":     "rl",
		"section_endpoint": "/rl",
		"state":            true,
	}
	resp, err := ts.MakeRequest("POST", "/api/v1/infographic", createPayload, adminToken)
	require.NoError(t, err)
	body, _ := parseJSONResponse(resp.Body)
	infID := getDataField(body)["id"].(string)

	// Hit 30 times — should all succeed
	for i := 0; i < 30; i++ {
		resp, err := ts.MakeRequest("GET", "/api/v1/public/infographic/"+infID, nil, "")
		require.NoError(t, err)
		require.Equalf(t, http.StatusOK, resp.StatusCode, "request %d should succeed", i+1)
	}

	// 31st should be rate-limited
	resp, err = ts.MakeRequest("GET", "/api/v1/public/infographic/"+infID, nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
