package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertBeritaShape asserts all fields of a berita response object
func assertBeritaShape(t *testing.T, data map[string]interface{}) {
	t.Helper()
	assert.NotEmpty(t, data["id"], "id must not be empty")
	assert.NotEmpty(t, data["title"], "title must not be empty")
	assert.NotEmpty(t, data["content"], "content must not be empty")
	assert.NotEmpty(t, data["category"], "category must not be empty")
	// image_url is nullable — key may be absent or nil
	assert.NotEmpty(t, data["created_at"], "created_at must not be empty")
	assert.NotEmpty(t, data["updated_at"], "updated_at must not be empty")
}

// assertPaginationShape asserts all fields of a pagination object
func assertPaginationShape(t *testing.T, pagination map[string]interface{}) {
	t.Helper()
	assert.NotNil(t, pagination["page"], "pagination.page must be present")
	assert.NotNil(t, pagination["limit"], "pagination.limit must be present")
	assert.NotNil(t, pagination["total"], "pagination.total must be present")
	assert.NotNil(t, pagination["total_pages"], "pagination.total_pages must be present")
}

func TestBerita_HappyPath(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create categories first (category is now a UUID FK)
	pengumumanID := ts.CreateTestBeritaCategory(t, "Pengumuman")
	beritaID2 := ts.CreateTestBeritaCategory(t, "Berita")

	// Create Berita
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
		map[string]string{"title": "Test News", "content": "This is test news content", "category": pengumumanID},
		nil,
		token,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "title", "Test News")
	validator.AssertFieldValue(t, "content", "This is test news content")
	catObj, ok := validator.Data["category"].(map[string]interface{})
	require.True(t, ok, "category should be an object {id, name}")
	assert.Equal(t, pengumumanID, catObj["id"], "category.id should match the created category")
	validator.AssertFieldIsUUID(t, "id")
	validator.AssertFieldIsTimestamp(t, "created_at")
	validator.AssertFieldIsTimestamp(t, "updated_at")
	beritaID := validator.Data["id"].(string)

	// List Berita (public) - should return simple data without content
	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/list", nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "berita")
	validator.AssertFieldExists(t, "pagination")
	validator.AssertFieldType(t, "berita", "array")

	beritaList := validator.Data["berita"].([]interface{})
	assert.Len(t, beritaList, 1)
	listItem := beritaList[0].(map[string]interface{})
	assert.NotEmpty(t, listItem["id"])
	assert.NotEmpty(t, listItem["title"])
	assert.NotEmpty(t, listItem["category"])

// Get Berita by ID (public) - should return full content
	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/"+beritaID, nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", beritaID)
	validator.AssertFieldValue(t, "title", "Test News")
	validator.AssertFieldValue(t, "content", "This is test news content")
	catObj, ok = validator.Data["category"].(map[string]interface{})
	require.True(t, ok, "category should be an object {id, name}")
	assert.Equal(t, pengumumanID, catObj["id"], "category.id should match")

	// Update Berita
	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/berita/"+beritaID,
		map[string]string{"title": "Updated News", "content": "Updated news content", "category": beritaID2},
		nil,
		token,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", beritaID)
	validator.AssertFieldValue(t, "title", "Updated News")
	validator.AssertFieldValue(t, "content", "Updated news content")
	catObj, ok = validator.Data["category"].(map[string]interface{})
	require.True(t, ok, "category should be an object {id, name}")
	assert.Equal(t, beritaID2, catObj["id"], "category.id should match updated category")

	// Delete Berita
	resp, err = ts.MakeRequest("DELETE", "/api/v1/berita/"+beritaID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
}

func TestBerita_CreateWithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
		map[string]string{"title": "Test News", "content": "Content"},
		nil, "",
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestBerita_CreateWithEmptyTitle(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
		map[string]string{"title": "", "content": "Content"},
		nil, token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestBerita_CreateWithEmptyContent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
		map[string]string{"title": "Test News", "content": ""},
		nil, token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestBerita_GetNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/00000000-0000-0000-0000-000000000000", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestBerita_GetWithInvalidID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/invalid-uuid", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Public endpoint returns 404 for any lookup error including invalid UUID format.
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestBerita_UpdateNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeMultipartRequest("PUT", "/api/v1/berita/00000000-0000-0000-0000-000000000000",
		map[string]string{"title": "Updated News", "content": "Updated content"},
		nil, token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestBerita_DeleteNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/berita/00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestBerita_ListWithInvalidPage(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/list?page=0&limit=10", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestBerita_ListExceedsMaxLimit(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/list?page=1&limit=101", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

// TestBerita_CategoriesEndpoint verifies the new enhanced categories endpoint
// returns UUIDs and pagination metadata. See integration/beritacategory_test.go
// for comprehensive coverage of the category CRUD endpoints.
func TestBerita_CategoriesEndpoint(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestBeritaCategory(t, "Pengumuman")
	ts.CreateTestBeritaCategory(t, "Berita")
	ts.CreateTestBeritaCategory(t, "Laporan")

	// Public access (no auth)
	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))

	data := getDataField(body)
	require.NotNil(t, data)

	// New format: categories is an array of objects with id + name
	categories, ok := data["categories"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(categories), 3)

	for _, c := range categories {
		cm, ok := c.(map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, cm["id"], "category should have an id (UUID)")
		assert.NotEmpty(t, cm["name"], "category should have a name")
	}

	// Pagination metadata should be present
	_, hasPagination := data["pagination"]
	assert.True(t, hasPagination, "pagination metadata should be present")
}

// TestBerita_ListEndpointReturnsSimpleData verifies that list endpoint excludes content field
func TestBerita_ListEndpointReturnsSimpleData(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestBeritaCategory(t, "News")

	// Create berita with long content
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
		map[string]string{
			"title":    "Test Article",
			"content":  "This is a very long content that should not appear in list endpoint. Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			"category": cat0ID,
		},
		nil,
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	beritaID := data["id"].(string)

	// Get list endpoint (public)
	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/list", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	listData := getDataField(body)
	beritaList := listData["berita"].([]interface{})
	require.Len(t, beritaList, 1)

	listItem := beritaList[0].(map[string]interface{})

	// Verify list response does NOT include content field
	assert.NotContains(t, listItem, "content", "list endpoint should not include content field")
	assert.NotEmpty(t, listItem["id"])
	assert.NotEmpty(t, listItem["title"])
	assert.NotEmpty(t, listItem["category"])
	assert.NotEmpty(t, listItem["created_at"])
	assert.NotEmpty(t, listItem["updated_at"])

	// Get detail endpoint (public) - should include content
	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/"+beritaID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	detailData := getDataField(body)

	// Verify detail response INCLUDES content field
	assert.NotEmpty(t, detailData["content"], "detail endpoint should include content field")
	assert.Equal(t, "This is a very long content that should not appear in list endpoint. Lorem ipsum dolor sit amet, consectetur adipiscing elit.", detailData["content"])
}

// TestBerita_DetailEndpointReturnsFullContent verifies detail endpoint includes content
func TestBerita_DetailEndpointReturnsFullContent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestBeritaCategory(t, "Berita")

	fullContent := "This is the complete article content with all details and information that should only appear in the detail view."

	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
		map[string]string{
			"title":    "Full Content Article",
			"content":  fullContent,
			"category": cat0ID,
		},
		nil,
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	beritaID := data["id"].(string)

	// Get detail endpoint
	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/"+beritaID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	detailData := getDataField(body)

	// Verify all fields are present
	assert.Equal(t, beritaID, detailData["id"])
	assert.Equal(t, "Full Content Article", detailData["title"])
	assert.Equal(t, fullContent, detailData["content"])
	catObj, ok := detailData["category"].(map[string]interface{})
	require.True(t, ok, "category should be an object {id, name}")
	assert.Equal(t, "Berita", catObj["name"], "category.name should match")
	assert.NotEmpty(t, detailData["created_at"])
	assert.NotEmpty(t, detailData["updated_at"])
}

// TestBerita_TimestampsInResponses verifies all responses include created_at and updated_at
func TestBerita_TimestampsInResponses(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestBeritaCategory(t, "Test")

	// Create berita
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
		map[string]string{
			"title":    "Timestamp Test",
			"content":  "Testing timestamps",
			"category": cat0ID,
		},
		nil,
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify timestamps in creation response
	assert.NotEmpty(t, data["created_at"], "created_at must be present in creation response")
	assert.NotEmpty(t, data["updated_at"], "updated_at must be present in creation response")

	beritaID := data["id"].(string)
	createdAt := data["created_at"].(string)

	// Get berita
	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/"+beritaID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	// Verify timestamps in retrieval response
	assert.NotEmpty(t, data["created_at"], "created_at must be present in retrieval response")
	assert.NotEmpty(t, data["updated_at"], "updated_at must be present in retrieval response")

	// Parse and compare timestamps (handle both Z and +07:00 formats)
	createdAtParsed, err := time.Parse(time.RFC3339, createdAt)
	require.NoError(t, err, "created_at should be valid RFC3339 format")
	retrievedAtParsed, err := time.Parse(time.RFC3339, data["created_at"].(string))
	require.NoError(t, err, "retrieved created_at should be valid RFC3339 format")
	assert.True(t, createdAtParsed.Equal(retrievedAtParsed), "created_at should not change")

	// Update berita
	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/berita/"+beritaID,
		map[string]string{
			"title":    "Updated Title",
			"content":  "Updated content",
			"category": cat0ID,
		},
		nil,
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	// Verify timestamps in update response
	assert.NotEmpty(t, data["created_at"], "created_at must be present in update response")
	assert.NotEmpty(t, data["updated_at"], "updated_at must be present in update response")

	// Parse and compare created_at after update
	updatedAtParsed, err := time.Parse(time.RFC3339, data["created_at"].(string))
	require.NoError(t, err, "updated created_at should be valid RFC3339 format")
	assert.True(t, createdAtParsed.Equal(updatedAtParsed), "created_at should not change on update")
	// updated_at should be different (or at least present)
	assert.NotEmpty(t, data["updated_at"])
}

// TestBerita_List_FilterByQuery verifies the admin list endpoint filters by query string
// across both title and content (case-insensitive via ILIKE).
func TestBerita_List_FilterByQuery(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	catID := ts.CreateTestBeritaCategory(t, "Filter Query Test")

	// Create three berita: two with needle in title, one with needle only in content.
	mk := func(title, content string) {
		resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
			map[string]string{"title": title, "content": content, "category": catID},
			nil, token,
		)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	mk("Sunrise Event Tomorrow", "Generic news content body.")
	mk("Travel Diary", "The sunrise over the mountain was unforgettable.")
	mk("Cooking Recipes", "Soup recipes from grandma's kitchen.") // negative control

	// q matching title only
	resp, err := ts.MakeRequest("GET", "/api/v1/berita?q=Sunrise+Event", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	list := data["berita"].([]interface{})
	assert.Len(t, list, 1, "q matching title only should return 1 record")
	assert.Equal(t, "Sunrise Event Tomorrow", list[0].(map[string]interface{})["title"])

	// q case-insensitive on title
	resp, err = ts.MakeRequest("GET", "/api/v1/berita?q=sUnRiSe+eVeNt", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["berita"].([]interface{})
	assert.Len(t, list, 1, "ILIKE should be case-insensitive")
	assert.Equal(t, "Sunrise Event Tomorrow", list[0].(map[string]interface{})["title"])

	// q matching content only (across columns)
	resp, err = ts.MakeRequest("GET", "/api/v1/berita?q=grandma", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["berita"].([]interface{})
	assert.Len(t, list, 1, "q should match content column too")
	assert.Equal(t, "Cooking Recipes", list[0].(map[string]interface{})["title"])

	// q matching BOTH title and content → returns both
	resp, err = ts.MakeRequest("GET", "/api/v1/berita?q=Sunrise", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["berita"].([]interface{})
	assert.Len(t, list, 2, "q 'Sunrise' should match 2 records (1 by title, 1 by content)")

	// q with no matches → empty
	resp, err = ts.MakeRequest("GET", "/api/v1/berita?q=zzznonexistent", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["berita"].([]interface{})
	assert.Len(t, list, 0, "non-matching q should return empty")
	assert.Equal(t, float64(0), data["pagination"].(map[string]interface{})["total"])
}

// TestBerita_List_FilterByCategory verifies the admin list endpoint filters by category UUID.
func TestBerita_List_FilterByCategory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	catA := ts.CreateTestBeritaCategory(t, "Category A")
	catB := ts.CreateTestBeritaCategory(t, "Category B")

	mk := func(title, cat string) {
		resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
			map[string]string{"title": title, "content": "body", "category": cat},
			nil, token,
		)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	mk("A one", catA)
	mk("A two", catA)
	mk("B one", catB)

	// Filter by catA only
	resp, err := ts.MakeRequest("GET", "/api/v1/berita?category="+catA, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	list := data["berita"].([]interface{})
	assert.Len(t, list, 2)
	assert.Equal(t, float64(2), data["pagination"].(map[string]interface{})["total"])
	for _, item := range list {
		cat := item.(map[string]interface{})["category"].(map[string]interface{})
		assert.Equal(t, catA, cat["id"], "all results must have catA id")
	}

	// Filter by catB only
	resp, err = ts.MakeRequest("GET", "/api/v1/berita?category="+catB, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["berita"].([]interface{})
	assert.Len(t, list, 1)
	assert.Equal(t, "B one", list[0].(map[string]interface{})["title"])
}

// TestBerita_List_CombineQueryAndCategory verifies q+category combine with AND semantics.
func TestBerita_List_CombineQueryAndCategory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	catA := ts.CreateTestBeritaCategory(t, "Cat Combine A")
	catB := ts.CreateTestBeritaCategory(t, "Cat Combine B")

	mk := func(title, content, cat string) {
		resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
			map[string]string{"title": title, "content": content, "category": cat},
			nil, token,
		)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	mk("Festival A", "Festival news", catA)
	mk("Festival B", "Festival news", catB)
	mk("Other A", "no match in title", catA)

	// q=festival + category=catA → should return 1 (Festival A only)
	resp, err := ts.MakeRequest("GET", "/api/v1/berita?q=Festival&category="+catA, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	list := data["berita"].([]interface{})
	assert.Len(t, list, 1, "combined filter must intersect (AND) q and category")
	assert.Equal(t, "Festival A", list[0].(map[string]interface{})["title"])
}

// TestBerita_List_SortByDate verifies the sort/order query params
// enable a "recently news" tab on the frontend.
func TestBerita_List_SortByDate(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	catID := ts.CreateTestBeritaCategory(t, "News")

	mk := func(title string) string {
		resp, err := ts.MakeMultipartRequest("POST", "/api/v1/berita",
			map[string]string{"title": title, "content": "content", "category": catID},
			nil, token)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		body, _ := parseJSONResponse(resp.Body)
		return getDataField(body)["id"].(string)
	}

	id1 := mk("Alpha")
	time.Sleep(20 * time.Millisecond)
	_ = mk("Charlie")
	time.Sleep(20 * time.Millisecond)
	id3 := mk("Bravo")

	// Default order = newest first (desc by created_at)
	resp, err := ts.MakeRequest("GET", "/api/v1/berita?category="+catID, nil, token)
	require.NoError(t, err)
	body, _ := parseJSONResponse(resp.Body)
	list := getDataField(body)["berita"].([]interface{})
	require.GreaterOrEqual(t, len(list), 3, "need at least 3 records from this test")
	firstID := list[0].(map[string]interface{})["id"].(string)
	assert.Equal(t, id3, firstID, "newest (Bravo) should be first by default")
	lastID := list[len(list)-1].(map[string]interface{})["id"].(string)
	assert.Equal(t, id1, lastID, "oldest (Alpha) should be last by default")

	// Explicit sort=title, order=asc → alphabetical A,B,C
	resp, err = ts.MakeRequest("GET", "/api/v1/berita?category="+catID+"&sort=title&order=asc", nil, token)
	require.NoError(t, err)
	body, _ = parseJSONResponse(resp.Body)
	list = getDataField(body)["berita"].([]interface{})
	firstTitle := list[0].(map[string]interface{})["title"].(string)
	assert.Equal(t, "Alpha", firstTitle, "sort=title&order=asc → Alpha first")
	lastTitle := list[len(list)-1].(map[string]interface{})["title"].(string)
	assert.Equal(t, "Charlie", lastTitle, "sort=title&order=asc → Charlie last")

	// sort=created_at&order=asc → oldest first
	resp, err = ts.MakeRequest("GET", "/api/v1/berita?category="+catID+"&sort=created_at&order=asc", nil, token)
	require.NoError(t, err)
	body, _ = parseJSONResponse(resp.Body)
	list = getDataField(body)["berita"].([]interface{})
	firstID = list[0].(map[string]interface{})["id"].(string)
	assert.Equal(t, id1, firstID, "sort=created_at&order=asc → Alpha first")

	// Invalid sort value falls back to default (created_at desc)
	resp, err = ts.MakeRequest("GET", "/api/v1/berita?category="+catID+"&sort=password", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
