package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertBannerShape validates all required fields of a banner response object
func assertBannerShape(t *testing.T, b map[string]interface{}) {
	t.Helper()
	assert.NotEmpty(t, b["id"], "banner.id must not be empty")
	assert.NotEmpty(t, b["title"], "banner.title must not be empty")
	assert.NotEmpty(t, b["image_url"], "banner.image_url must not be empty")
	assert.NotEmpty(t, b["status"], "banner.status must not be empty")
	assert.NotEmpty(t, b["created_at"], "banner.created_at must not be empty")
	assert.NotEmpty(t, b["updated_at"], "banner.updated_at must not be empty")
	// Metadata field should be present (can be null or empty object)
	assert.NotNil(t, b["metadata"], "banner.metadata must be present")
	// Description field should be present (can be empty string)
	assert.NotNil(t, b["description"], "banner.description must be present")
}

// assertBannerHasDescription validates that a banner has the expected description
func assertBannerHasDescription(t *testing.T, b map[string]interface{}, expectedDesc string) {
	t.Helper()
	desc, ok := b["description"].(string)
	require.True(t, ok, "description must be a string")
	assert.Equal(t, expectedDesc, desc)
}

func TestCreateBanner_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Test Banner",
			"description": "Test banner description",
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	// Comprehensive response validation
	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "id")
	validator.AssertFieldIsUUID(t, "id")
	validator.AssertFieldValue(t, "title", "Test Banner")
	validator.AssertFieldValue(t, "description", "Test banner description")
	validator.AssertFieldValue(t, "status", "inactive")
	validator.AssertFieldNotEmpty(t, "image_url")
	validator.AssertFieldNotEmpty(t, "created_at")
	validator.AssertFieldIsTimestamp(t, "created_at")
	validator.AssertFieldNotEmpty(t, "updated_at")
	validator.AssertFieldIsTimestamp(t, "updated_at")
	validator.AssertFieldExists(t, "metadata")
}

func TestGetBanner_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	createResp, _ := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "My Banner"},
		map[string][2]string{"image": {"b.jpg", "data"}},
		token,
	)
	createBody, _ := parseJSONResponse(createResp.Body)
	bannerID := getDataField(createBody)["id"].(string)

	resp, err := ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", bannerID)
	validator.AssertFieldValue(t, "title", "My Banner")
	validator.AssertFieldValue(t, "status", "inactive")
	validator.AssertFieldNotEmpty(t, "image_url")
	validator.AssertFieldIsTimestamp(t, "created_at")
	validator.AssertFieldIsTimestamp(t, "updated_at")
}

func TestListBanners_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Banner 1"},
		map[string][2]string{"image": {"b1.jpg", "data"}},
		token,
	)

	resp, err := ts.MakeRequest("GET", "/api/v1/banners", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "banners")
	validator.AssertFieldExists(t, "pagination")
	validator.AssertFieldType(t, "banners", "array")
	validator.AssertFieldType(t, "pagination", "object")

	// Validate pagination structure
	pagination := validator.Data["pagination"].(map[string]interface{})
	assert.NotNil(t, pagination["page"])
	assert.NotNil(t, pagination["limit"])
	assert.NotNil(t, pagination["total"])
	assert.NotNil(t, pagination["total_pages"])

	// Validate banners array
	banners := validator.Data["banners"].([]interface{})
	assert.NotEmpty(t, banners)
	for _, b := range banners {
		assertBannerShape(t, b.(map[string]interface{}))
	}
}

func TestGetActiveBanners_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Active Banner"},
		map[string][2]string{"image": {"b.jpg", "data"}},
		token,
	)

	resp, err := ts.MakeRequest("GET", "/api/v1/banners/active", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	// GetActiveBanners returns the array directly as data (not wrapped in {banners: [...]})
	rawData := body["data"]
	banners, ok := rawData.([]interface{})
	require.True(t, ok, "data must be an array of banners")
	for _, b := range banners {
		assertBannerShape(t, b.(map[string]interface{}))
	}
}

func TestUpdateBanner_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	createResp, _ := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Old Title",
			"description": "Old description",
		},
		map[string][2]string{"image": {"b.jpg", "data"}},
		token,
	)
	createBody, _ := parseJSONResponse(createResp.Body)
	bannerID := getDataField(createBody)["id"].(string)

	resp, err := ts.MakeMultipartRequest("PUT", "/api/v1/banners/"+bannerID,
		map[string]string{
			"title":       "New Title",
			"description": "New description",
		},
		map[string][2]string{"image": {"b2.jpg", "newdata"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", bannerID)
	validator.AssertFieldValue(t, "title", "New Title")
	validator.AssertFieldValue(t, "description", "New description")
	validator.AssertFieldIsTimestamp(t, "created_at")
	validator.AssertFieldIsTimestamp(t, "updated_at")
}

func TestDeleteBanner_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	createResp, _ := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "To Delete"},
		map[string][2]string{"image": {"b.jpg", "data"}},
		token,
	)
	createBody, _ := parseJSONResponse(createResp.Body)
	bannerID := getDataField(createBody)["id"].(string)

	resp, err := ts.MakeRequest("DELETE", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Equal(t, "Banner deleted successfully", data["message"])
}

func TestCreateBanner_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Test"},
		map[string][2]string{"image": {"b.jpg", "data"}},
		"",
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	assert.False(t, body["success"].(bool))
}

func TestCreateBanner_EmptyTitle(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": ""},
		map[string][2]string{"image": {"b.jpg", "data"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, false)
	// Error is at root level, not in data
	errMsg, ok := body["error"].(string)
	require.True(t, ok, "error field must be present at root level")
	assert.NotEmpty(t, errMsg)
}

func TestGetBanner_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/banners/00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, false)
	validator.AssertError(t, "Banner not found")
}

func TestGetBanner_InvalidID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/banners/invalid-uuid", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	assert.False(t, body["success"].(bool))
}

func TestDeleteBanner_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/banners/00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, _ := parseJSONResponse(resp.Body)
	assert.False(t, body["success"].(bool))
}

func TestBannerMetadata_PersistenceAndRetrieval(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with metadata
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Banner with Metadata"},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	bannerID := data["id"].(string)

	// Verify metadata field is present in creation response
	assert.NotNil(t, data["metadata"], "metadata field must be present in creation response")

	// Retrieve banner and verify metadata is persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data = getDataField(body)

	// Verify metadata is persisted and retrievable
	assert.NotNil(t, data["metadata"], "metadata field must be persisted and retrievable")
	assertBannerShape(t, data)
}

// TestBannerMetadata_CreateWithMetadata verifies metadata can be sent and retrieved
func TestBannerMetadata_CreateWithMetadata(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with metadata
	metadataJSON := `{"html_content":"<p>Welcome to our village</p>","color":"#FF5733","link":"https://example.com"}`

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":    "Banner with Metadata",
			"metadata": metadataJSON,
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)

	// Verify metadata is present in creation response
	assert.NotNil(t, data["metadata"], "metadata field must be present")
	metadata, ok := data["metadata"].(map[string]interface{})
	require.True(t, ok, "metadata must be an object")

	// Verify metadata content
	assert.Equal(t, "<p>Welcome to our village</p>", metadata["html_content"])
	assert.Equal(t, "#FF5733", metadata["color"])
	assert.Equal(t, "https://example.com", metadata["link"])

	bannerID := data["id"].(string)

	// Retrieve banner and verify metadata is persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data = getDataField(body)

	// Verify metadata is persisted and retrievable
	assert.NotNil(t, data["metadata"], "metadata field must be persisted")
	metadata, ok = data["metadata"].(map[string]interface{})
	require.True(t, ok, "metadata must be an object in retrieval")

	assert.Equal(t, "<p>Welcome to our village</p>", metadata["html_content"])
	assert.Equal(t, "#FF5733", metadata["color"])
	assert.Equal(t, "https://example.com", metadata["link"])
}

// TestBannerMetadata_UpdateMetadata verifies metadata can be updated
func TestBannerMetadata_UpdateMetadata(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with initial metadata
	initialMetadata := `{"version":"1.0"}`

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":    "Banner to Update",
			"metadata": initialMetadata,
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	bannerID := data["id"].(string)

	// Update banner with new metadata
	updatedMetadata := `{"version":"2.0","updated":true,"timestamp":"2024-01-15T10:00:00Z"}`

	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/banners/"+bannerID,
		map[string]string{
			"title":    "Banner Updated",
			"metadata": updatedMetadata,
		},
		map[string][2]string{"image": {"banner_new.jpg", "new-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data = getDataField(body)

	// Verify metadata is updated
	assert.NotNil(t, data["metadata"], "metadata field must be present after update")
	metadata, ok := data["metadata"].(map[string]interface{})
	require.True(t, ok, "metadata must be an object")

	assert.Equal(t, "2.0", metadata["version"])
	assert.Equal(t, true, metadata["updated"])
	assert.Equal(t, "2024-01-15T10:00:00Z", metadata["timestamp"])

	// Retrieve and verify updated metadata is persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	metadata, ok = data["metadata"].(map[string]interface{})
	require.True(t, ok, "metadata must be an object in retrieval after update")

	assert.Equal(t, "2.0", metadata["version"])
	assert.Equal(t, true, metadata["updated"])
}

// TestBannerMetadata_EmptyMetadata verifies empty metadata is handled correctly
func TestBannerMetadata_EmptyMetadata(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner without metadata
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Banner without Metadata"},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify metadata field is present (can be null or empty object)
	assert.NotNil(t, data["metadata"], "metadata field must be present even when not provided")

	bannerID := data["id"].(string)

	// Retrieve and verify metadata field is present
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assert.NotNil(t, data["metadata"], "metadata field must be present in retrieval")
}

// TestBannerMetadata_InvalidMetadataJSON verifies invalid JSON is rejected
func TestBannerMetadata_InvalidMetadataJSON(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Try to create banner with invalid metadata JSON
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":    "Banner with Bad Metadata",
			"metadata": `{"invalid": json}`, // Invalid JSON
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.NotEmpty(t, body["error"])
	assert.Contains(t, body["error"].(string), "Invalid metadata JSON format")
}

// TestBannerMetadata_ListIncludesMetadata verifies list endpoint includes metadata
func TestBannerMetadata_ListIncludesMetadata(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create multiple banners with different metadata
	metadata1 := `{"type":"announcement"}`
	metadata2 := `{"type":"promotion","discount":"20%"}`

	ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":    "Banner 1",
			"metadata": metadata1,
		},
		map[string][2]string{"image": {"b1.jpg", "data"}},
		token,
	)

	ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":    "Banner 2",
			"metadata": metadata2,
		},
		map[string][2]string{"image": {"b2.jpg", "data"}},
		token,
	)

	// List banners
	resp, err := ts.MakeRequest("GET", "/api/v1/banners", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	banners := data["banners"].([]interface{})

	// Verify all banners have metadata field
	for _, b := range banners {
		banner := b.(map[string]interface{})
		assert.NotNil(t, banner["metadata"], "metadata field must be present in list response")
	}
}

// TestBannerStatusUpdate_Success verifies banner status can be updated successfully
func TestBannerStatusUpdate_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner (default status is inactive)
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Test Banner"},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify initial status is inactive
	assert.Equal(t, "inactive", data["status"])
	bannerID := data["id"].(string)

	// Update banner status to active
	resp, err = ts.MakeRequest("PATCH", "/api/v1/banners/"+bannerID+"/status",
		map[string]string{"status": "active"},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data = getDataField(body)

	// Verify status is now active
	assert.Equal(t, "active", data["status"])
	assert.Equal(t, bannerID, data["id"])

	// Verify status is persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assert.Equal(t, "active", data["status"], "status should be persisted")
}

// TestBannerStatusUpdate_InvalidStatus verifies invalid status is rejected
func TestBannerStatusUpdate_InvalidStatus(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner
	resp, _ := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Test Banner"},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	body, _ := parseJSONResponse(resp.Body)
	bannerID := getDataField(body)["id"].(string)

	// Try to update with invalid status
	resp, err := ts.MakeRequest("PATCH", "/api/v1/banners/"+bannerID+"/status",
		map[string]string{"status": "invalid_status"},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestBannerStatusUpdate_NotFound verifies 404 for non-existent banner
func TestBannerStatusUpdate_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Try to update non-existent banner
	resp, err := ts.MakeRequest("PATCH", "/api/v1/banners/00000000-0000-0000-0000-000000000000/status",
		map[string]string{"status": "active"},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestBannerStatusUpdate_WithoutAuth verifies unauthenticated request is rejected
func TestBannerStatusUpdate_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Try to update without authentication
	resp, err := ts.MakeRequest("PATCH", "/api/v1/banners/00000000-0000-0000-0000-000000000000/status",
		map[string]string{"status": "active"},
		"",
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestBannerDescription_CreateWithDescription verifies description can be set on creation
func TestBannerDescription_CreateWithDescription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with description
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Banner with Description",
			"description": "This is a detailed description of the banner campaign",
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)

	// Verify description is present in creation response
	assert.Equal(t, "This is a detailed description of the banner campaign", data["description"])
	assertBannerShape(t, data)

	bannerID := data["id"].(string)

	// Retrieve banner and verify description is persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data = getDataField(body)

	// Verify description is persisted and retrievable
	assert.Equal(t, "This is a detailed description of the banner campaign", data["description"])
}

// TestBannerDescription_UpdateDescription verifies description can be updated
func TestBannerDescription_UpdateDescription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner without description
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Banner to Update"},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	bannerID := data["id"].(string)

	// Initially description should be empty
	assert.Equal(t, "", data["description"])

	// Update banner with new description
	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/banners/"+bannerID,
		map[string]string{
			"title":       "Banner Updated",
			"description": "Updated description with new content",
		},
		map[string][2]string{"image": {"banner_new.jpg", "new-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data = getDataField(body)

	// Verify description is updated
	assert.Equal(t, "Updated description with new content", data["description"])

	// Retrieve and verify updated description is persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assert.Equal(t, "Updated description with new content", data["description"])
}

// TestBannerDescription_ListIncludesDescription verifies list endpoint includes description
func TestBannerDescription_ListIncludesDescription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banners with different descriptions
	ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Banner 1",
			"description": "First banner description",
		},
		map[string][2]string{"image": {"b1.jpg", "data"}},
		token,
	)

	ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Banner 2",
			"description": "Second banner description",
		},
		map[string][2]string{"image": {"b2.jpg", "data"}},
		token,
	)

	// List banners
	resp, err := ts.MakeRequest("GET", "/api/v1/banners", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	banners := data["banners"].([]interface{})

	// Verify all banners have description field
	foundDescriptions := make(map[string]bool)
	for _, b := range banners {
		banner := b.(map[string]interface{})
		assert.NotNil(t, banner["description"], "description field must be present in list response")
		desc, ok := banner["description"].(string)
		require.True(t, ok, "description must be a string")
		foundDescriptions[desc] = true
	}

	// Verify both descriptions are present
	assert.True(t, foundDescriptions["First banner description"], "First description should be in list")
	assert.True(t, foundDescriptions["Second banner description"], "Second description should be in list")
}

// TestBannerDescription_EmptyDescription verifies empty description is handled correctly
func TestBannerDescription_EmptyDescription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner without description
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Banner without Description"},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify description field is present and empty
	assert.NotNil(t, data["description"], "description field must be present even when not provided")
	assert.Equal(t, "", data["description"])

	bannerID := data["id"].(string)

	// Retrieve and verify description field is present
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assert.NotNil(t, data["description"], "description field must be present in retrieval")
	assert.Equal(t, "", data["description"])
}

// TestBannerDescription_LongDescription verifies long description is handled correctly
func TestBannerDescription_LongDescription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with long description
	longDescription := "This is a very long description that contains multiple sentences. " +
		"It should be stored correctly in the database and retrieved without any issues. " +
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Banner with Long Description",
			"description": longDescription,
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify long description is stored correctly
	assert.Equal(t, longDescription, data["description"])

	bannerID := data["id"].(string)

	// Retrieve and verify long description is persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assert.Equal(t, longDescription, data["description"], "long description should be persisted correctly")
}

// TestBannerDescription_UpdatePreservesStatus verifies updating description preserves status
func TestBannerDescription_UpdatePreservesStatus(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Banner to Update"},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	bannerID := data["id"].(string)

	// Verify initial status is inactive
	assert.Equal(t, "inactive", data["status"])

	// Update status to active
	resp, err = ts.MakeRequest("PATCH", "/api/v1/banners/"+bannerID+"/status",
		map[string]string{"status": "active"},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))

	// Now update description
	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/banners/"+bannerID,
		map[string]string{
			"title":       "Banner Updated",
			"description": "New description",
		},
		map[string][2]string{"image": {"banner_new.jpg", "new-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	// Verify description is updated but status is preserved
	assert.Equal(t, "New description", data["description"])
	assert.Equal(t, "active", data["status"], "status should be preserved during description update")
}

// TestBannerDescription_SpecialCharacters verifies description with special characters
func TestBannerDescription_SpecialCharacters(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with special characters in description
	specialDesc := "Description with special chars: @#$%^&*()_+-=[]{}|;:',.<>?/~`"

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Banner with Special Chars",
			"description": specialDesc,
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify special characters are preserved
	assertBannerHasDescription(t, data, specialDesc)

	bannerID := data["id"].(string)

	// Retrieve and verify special characters are persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assertBannerHasDescription(t, data, specialDesc)
}

// TestBannerDescription_UnicodeCharacters verifies description with unicode characters
func TestBannerDescription_UnicodeCharacters(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with unicode characters
	unicodeDesc := "Deskripsi dengan karakter Unicode: 你好世界 🌍 مرحبا العالم"

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Banner with Unicode",
			"description": unicodeDesc,
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify unicode characters are preserved
	assertBannerHasDescription(t, data, unicodeDesc)

	bannerID := data["id"].(string)

	// Retrieve and verify unicode characters are persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assertBannerHasDescription(t, data, unicodeDesc)
}

// TestBannerDescription_WhitespaceHandling verifies whitespace in description is preserved
func TestBannerDescription_WhitespaceHandling(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with various whitespace
	descWithWhitespace := "Line 1\nLine 2\n\nLine 4 with\ttabs\tand   spaces"

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Banner with Whitespace",
			"description": descWithWhitespace,
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify whitespace is preserved
	assertBannerHasDescription(t, data, descWithWhitespace)

	bannerID := data["id"].(string)

	// Retrieve and verify whitespace is persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assertBannerHasDescription(t, data, descWithWhitespace)
}

// TestBannerDescription_ClearDescription verifies description can be cleared
func TestBannerDescription_ClearDescription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with description
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title":       "Banner with Description",
			"description": "Initial description",
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	bannerID := data["id"].(string)

	// Update banner to clear description
	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/banners/"+bannerID,
		map[string]string{
			"title":       "Banner Updated",
			"description": "", // Clear description
		},
		map[string][2]string{"image": {"banner_new.jpg", "new-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	// Verify description is cleared
	assertBannerHasDescription(t, data, "")

	// Retrieve and verify description is cleared
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assertBannerHasDescription(t, data, "")
}

// TestBannerTimestamps_CreatedAtAndUpdatedAt verifies timestamps are set correctly
func TestBannerTimestamps_CreatedAtAndUpdatedAt(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Banner for Timestamps"},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	validator := NewResponseValidator(body)

	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "id")
	validator.AssertFieldExists(t, "created_at")
	validator.AssertFieldExists(t, "updated_at")
	validator.AssertFieldIsTimestamp(t, "created_at")
	validator.AssertFieldIsTimestamp(t, "updated_at")

	data := validator.Data
	bannerID := data["id"].(string)
	createdAtStr := data["created_at"].(string)
	updatedAtStr := data["updated_at"].(string)

	// Add a delay to ensure timestamps will be different
	// Using 1 second to account for database timestamp precision
	time.Sleep(1 * time.Second)

	// Update banner
	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/banners/"+bannerID,
		map[string]string{"title": "Updated Banner"},
		map[string][2]string{"image": {"banner_new.jpg", "new-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	validator = NewResponseValidator(body)

	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "created_at")
	validator.AssertFieldExists(t, "updated_at")

	data = validator.Data
	newCreatedAtStr := data["created_at"].(string)
	newUpdatedAtStr := data["updated_at"].(string)

	// Parse timestamps for comparison
	origCreatedTime, _ := parseTimestamp(createdAtStr)
	currCreatedTime, _ := parseTimestamp(newCreatedAtStr)
	origUpdatedTime, _ := parseTimestamp(updatedAtStr)
	newUpdatedTime, _ := parseTimestamp(newUpdatedAtStr)

	// Verify created_at is unchanged
	assert.True(t, origCreatedTime.Equal(currCreatedTime), "created_at should not change on update")

	// Verify updated_at is changed
	assert.False(t, origUpdatedTime.Equal(newUpdatedTime), "updated_at should change on update")
}

// TestBannerLink_CreateWithLink verifies link can be set on creation
func TestBannerLink_CreateWithLink(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with link as form field
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title": "Banner with Link",
			"link":  "https://example.com/promo",
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok, "response should have data object")

	// Verify link field is present at top level
	link, ok := data["link"].(string)
	require.True(t, ok, "link should be at top level")
	assert.Equal(t, "https://example.com/promo", link)
}

// TestBannerLink_GetBannerWithLink verifies link is returned in get banner response
func TestBannerLink_GetBannerWithLink(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with link
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title": "Banner Link Test",
			"link":  "https://example.com/details",
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := body["data"].(map[string]interface{})
	bannerID := data["id"].(string)

	// Get banner by ID
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/"+bannerID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data = body["data"].(map[string]interface{})

	// Verify link is returned
	link, ok := data["link"].(string)
	require.True(t, ok, "link should be present")
	assert.Equal(t, "https://example.com/details", link)
}

// TestBannerLink_ListBannersWithLink verifies link is included in list response
func TestBannerLink_ListBannersWithLink(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with link
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title": "Banner for List",
			"link":  "https://example.com/list",
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	// List banners
	resp, err = ts.MakeRequest("GET", "/api/v1/banners?page=1&limit=10", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	bannersData, ok := body["data"].(map[string]interface{})
	require.True(t, ok, "response should have data object")

	banners, ok := bannersData["banners"].([]interface{})
	require.True(t, ok, "banners should be an array")
	require.Greater(t, len(banners), 0, "should have at least one banner")

	// Check first banner has link
	firstBanner := banners[0].(map[string]interface{})
	link, ok := firstBanner["link"].(string)
	require.True(t, ok, "link should be present in list")
	assert.Equal(t, "https://example.com/list", link)
}

// TestBannerLink_UpdateBannerLink verifies link can be updated
func TestBannerLink_UpdateBannerLink(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner without link
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{"title": "Banner Before Update"},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := body["data"].(map[string]interface{})
	bannerID := data["id"].(string)

	// Update banner with new link
	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/banners/"+bannerID,
		map[string]string{
			"title": "Banner After Update",
			"link":  "https://example.com/updated",
		},
		map[string][2]string{"image": {"banner_new.jpg", "new-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data = body["data"].(map[string]interface{})

	// Verify link is updated
	link, ok := data["link"].(string)
	require.True(t, ok, "link should be present")
	assert.Equal(t, "https://example.com/updated", link)
}

// TestBannerLink_ActiveBannersIncludesLink verifies active banners endpoint includes link
func TestBannerLink_ActiveBannersIncludesLink(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create banner with link and activate it
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/banners",
		map[string]string{
			"title": "Active Banner with Link",
			"link":  "https://example.com/active",
		},
		map[string][2]string{"image": {"banner.jpg", "fake-image-content"}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := body["data"].(map[string]interface{})
	bannerID := data["id"].(string)

	// Activate banner
	resp, err = ts.MakeRequest("PATCH", "/api/v1/banners/"+bannerID+"/status",
		map[string]string{"status": "active"}, token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Get active banners (public endpoint)
	resp, err = ts.MakeRequest("GET", "/api/v1/banners/active", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	banners, ok := body["data"].([]interface{})
	require.True(t, ok, "response should have banners array")
	require.Greater(t, len(banners), 0, "should have at least one active banner")

	// Find our banner and verify link
	var found bool
	for _, b := range banners {
		banner := b.(map[string]interface{})
		if banner["id"] == bannerID {
			link, ok := banner["link"].(string)
			require.True(t, ok, "link should be present in active banner")
			assert.Equal(t, "https://example.com/active", link)
			found = true
			break
		}
	}
	assert.True(t, found, "should find the created banner in active banners")
}

// TestBannerList_FilterByQuery verifies the q query parameter filters banners
// by ILIKE match on title and description.
func TestBannerList_FilterByQuery(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create three banners with distinct titles.
	ts.CreateTestBannerWithCategory(t, "Annual Village Festival", "Annual celebration with food stalls", "")
	ts.CreateTestBannerWithCategory(t, "Health Campaign", "Free vaccination drive for villagers", "")
	ts.CreateTestBannerWithCategory(t, "Education Workshop", "Free coding class for teenagers", "")

	// Search for a substring only present in one title.
	resp, err := ts.MakeRequest("GET", "/api/v1/banners?q=Festival", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := getDataField(body)
	banners := data["banners"].([]interface{})
	require.Len(t, banners, 1, "q=Festival must match exactly one banner")

	title, ok := banners[0].(map[string]interface{})["title"].(string)
	require.True(t, ok, "title must be a string")
	assert.Contains(t, title, "Festival", "matched banner must contain the search term")

	// Search for a substring only present in one description.
	resp, err = ts.MakeRequest("GET", "/api/v1/banners?q=vaccination", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data = getDataField(body)
	banners = data["banners"].([]interface{})
	require.Len(t, banners, 1, "q=vaccination must match exactly one banner")
	desc, ok := banners[0].(map[string]interface{})["description"].(string)
	require.True(t, ok, "description must be a string")
	assert.Contains(t, desc, "vaccination")
}

// TestBannerList_FilterByCategory verifies the category query parameter filters
// banners by their category UUID.
func TestBannerList_FilterByCategory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create two distinct categories.
	catEvents := ts.createBannerCategory(t, "Events")
	catHealth := ts.createBannerCategory(t, "Health")

	// Create banners across the two categories.
	ts.CreateTestBannerWithCategory(t, "Festival Banner", "Celebration", catEvents)
	ts.CreateTestBannerWithCategory(t, "Vaccination Banner", "Health", catHealth)
	ts.CreateTestBannerWithCategory(t, "Workshop Banner", "Coding", catEvents)

	// Filter by the Events category.
	resp, err := ts.MakeRequest("GET", "/api/v1/banners?category="+catEvents, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := getDataField(body)
	banners := data["banners"].([]interface{})
	require.Len(t, banners, 2, "category filter must return only banners in Events")

	for _, b := range banners {
		bm := b.(map[string]interface{})
		catID, _ := bm["category_id"].(string)
		assert.Equal(t, catEvents, catID, "each returned banner must belong to the filtered category")
	}

	// Filter by the Health category.
	resp, err = ts.MakeRequest("GET", "/api/v1/banners?category="+catHealth, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data = getDataField(body)
	banners = data["banners"].([]interface{})
	require.Len(t, banners, 1, "category filter must return only the Health banner")
	bm := banners[0].(map[string]interface{})
	catID, _ := bm["category_id"].(string)
	assert.Equal(t, catHealth, catID)
}
