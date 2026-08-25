package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertUMKMShape asserts all fields of a UMKM response object
func assertUMKMShape(t *testing.T, data map[string]interface{}) {
	t.Helper()
	assert.NotEmpty(t, data["id"], "id must not be empty")
	assert.NotEmpty(t, data["name"], "name must not be empty")
	assert.NotEmpty(t, data["category"], "category must not be empty")
	assert.NotEmpty(t, data["description"], "description must not be empty")
	// owner, address, phone, email, website are nullable — keys may be absent or nil
	// images is an array field
	assert.NotNil(t, data["images"], "images field must be present")
	assert.NotEmpty(t, data["created_at"], "created_at must not be empty")
	assert.NotEmpty(t, data["updated_at"], "updated_at must not be empty")
}

func TestUMKM_HappyPath(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create categories first (category is now a UUID FK)
	kerajinanID := ts.CreateTestUMKMCategory(t, "Kerajinan")
	makananID := ts.CreateTestUMKMCategory(t, "Makanan")

	owner := "John Doe"
	address := "Test Address"
	phone := "081234567890"
	description := "A beautiful handmade craft business"

	// Create UMKM
	createPayload := map[string]any{
		"name":        "Test UMKM",
		"owner":       owner,
		"address":     address,
		"phone":       phone,
		"category":    kerajinanID,
		"description": description,
		"images":      []string{"/uploads/umkm1.jpg", "/uploads/umkm2.jpg"},
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/umkm", createPayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "name", "Test UMKM")
	validator.AssertFieldValue(t, "owner", owner)
	validator.AssertFieldValue(t, "address", address)
	validator.AssertFieldValue(t, "phone", phone)
	catObj, ok := validator.Data["category"].(map[string]interface{})
	require.True(t, ok, "category should be an object {id, name}")
	assert.Equal(t, kerajinanID, catObj["id"], "category.id should match")
	validator.AssertFieldValue(t, "description", description)
	validator.AssertFieldIsUUID(t, "id")
	validator.AssertFieldIsTimestamp(t, "created_at")
	validator.AssertFieldIsTimestamp(t, "updated_at")
	validator.AssertFieldType(t, "images", "array")
	validator.AssertFieldLength(t, "images", 2)
	umkmID := validator.Data["id"].(string)

// List UMKM (public)
	resp, err = ts.MakeRequest("GET", "/api/v1/public/umkm/list", nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "umkm")
	validator.AssertFieldExists(t, "pagination")
	validator.AssertFieldType(t, "umkm", "array")

	umkmList := validator.Data["umkm"].([]interface{})
	assert.Len(t, umkmList, 1)

	// Get UMKM by ID (public)
	resp, err = ts.MakeRequest("GET", "/api/v1/public/umkm/"+umkmID, nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", umkmID)
	validator.AssertFieldValue(t, "name", "Test UMKM")
	catObj, ok = validator.Data["category"].(map[string]interface{})
	require.True(t, ok, "category should be an object {id, name}")
	assert.Equal(t, kerajinanID, catObj["id"], "category.id should match")
	validator.AssertFieldValue(t, "description", description)

	// Update UMKM
	newOwner := "Jane Doe"
	updatePayload := map[string]any{
		"name":        "Updated UMKM",
		"owner":       newOwner,
		"category":    makananID,
		"description": "Updated description",
		"images":      []string{"/uploads/umkm3.jpg"},
	}

	resp, err = ts.MakeRequest("PUT", "/api/v1/umkm/"+umkmID, updatePayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", umkmID)
	validator.AssertFieldValue(t, "name", "Updated UMKM")
	validator.AssertFieldValue(t, "owner", newOwner)
	catObj, ok = validator.Data["category"].(map[string]interface{})
	require.True(t, ok, "category should be an object {id, name}")
	assert.Equal(t, makananID, catObj["id"], "category.id should match updated category")
	validator.AssertFieldValue(t, "description", "Updated description")
	validator.AssertFieldLength(t, "images", 1)

	// Delete UMKM
	resp, err = ts.MakeRequest("DELETE", "/api/v1/umkm/"+umkmID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
}

func TestUMKM_CreateWithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("POST", "/api/v1/umkm", map[string]any{"name": "Test UMKM"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestUMKM_CreateWithEmptyName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/umkm", map[string]any{"name": ""}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestUMKM_GetNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/umkm/00000000-0000-0000-0000-000000000000", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestUMKM_UpdateNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestUMKMCategory(t, "food")

	resp, err := ts.MakeRequest("PUT", "/api/v1/umkm/00000000-0000-0000-0000-000000000000",
		map[string]any{
			"name":        "Updated UMKM",
			"category":    cat0ID,
			"description": "Updated description",
		}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestUMKM_DeleteNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/umkm/00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestUMKM_ListWithInvalidPage(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/umkm/list?page=0&limit=10", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestUMKM_ListExceedsMaxLimit(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/umkm/list?page=1&limit=101", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestUMKM_CategoriesEndpoint(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create categories
	ts.CreateTestUMKMCategory(t, "Kerajinan")
	ts.CreateTestUMKMCategory(t, "Makanan")
	ts.CreateTestUMKMCategory(t, "Pertanian")

	// Get categories endpoint (public, enhanced with pagination)
	resp, err := ts.MakeRequest("GET", "/api/v1/public/umkm/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	require.NotNil(t, data)

	// New format: categories is an array of objects with id + name + usage_count
	categoriesData, ok := data["categories"].([]interface{})
	require.True(t, ok, "categories field must be an array")
	assert.GreaterOrEqual(t, len(categoriesData), 3, "should have at least 3 categories")

	for _, c := range categoriesData {
		cm, ok := c.(map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, cm["id"], "category should have an id (UUID)")
		assert.NotEmpty(t, cm["name"], "category should have a name")
	}

	// Pagination metadata should be present
	_, hasPagination := data["pagination"]
	assert.True(t, hasPagination, "pagination metadata should be present")
}

// TestUMKM_CategoryField verifies category field is properly stored and retrieved
func TestUMKM_CategoryField(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestUMKMCategory(t, "Kerajinan")

	// Create UMKM with category
	resp, err := ts.MakeRequest("POST", "/api/v1/umkm",
		map[string]any{
			"name":        "Craft Shop",
			"category":    cat0ID,
			"description": "Handmade crafts",
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify category in creation response
	catObj, ok := data["category"].(map[string]interface{})
	require.True(t, ok, "category should be an object {id, name}")
	assert.Equal(t, "Kerajinan", catObj["name"], "category.name should match")
	umkmID := data["id"].(string)

	// Get UMKM and verify category is persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/public/umkm/"+umkmID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	catObj, ok = data["category"].(map[string]interface{})
	require.True(t, ok, "category should be an object {id, name}")
	assert.Equal(t, "Kerajinan", catObj["name"], "category.name should be persisted")
}

// TestUMKM_DescriptionAndImagesFields verifies description and images array fields
func TestUMKM_DescriptionAndImagesFields(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestUMKMCategory(t, "Makanan")

	description := "A premium coffee shop serving specialty coffee and pastries"
	images := []string{"/uploads/umkm/coffee1.jpg", "/uploads/umkm/coffee2.jpg", "/uploads/umkm/coffee3.jpg"}

	// Create UMKM with description and images
	resp, err := ts.MakeRequest("POST", "/api/v1/umkm",
		map[string]any{
			"name":        "Coffee Shop",
			"category":    cat0ID,
			"description": description,
			"images":      images,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify description in creation response
	assert.Equal(t, description, data["description"])

	// Verify images array in creation response
	imagesData, ok := data["images"].([]interface{})
	require.True(t, ok, "images must be an array")
	assert.Len(t, imagesData, 3)
	for i, img := range imagesData {
		assert.Equal(t, images[i], img.(string))
	}

	umkmID := data["id"].(string)

	// Get UMKM and verify description and images are persisted
	resp, err = ts.MakeRequest("GET", "/api/v1/public/umkm/"+umkmID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assert.Equal(t, description, data["description"], "description should be persisted")

	imagesData, ok = data["images"].([]interface{})
	require.True(t, ok, "images must be an array in retrieval")
	assert.Len(t, imagesData, 3)
	for i, img := range imagesData {
		assert.Equal(t, images[i], img.(string))
	}

	// Update UMKM with new images
	newImages := []string{"/uploads/umkm/coffee_new1.jpg", "/uploads/umkm/coffee_new2.jpg"}
	resp, err = ts.MakeRequest("PUT", "/api/v1/umkm/"+umkmID,
		map[string]any{
			"name":        "Coffee Shop",
			"category":    cat0ID,
			"description": "Updated description",
			"images":      newImages,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	// Verify updated images
	imagesData, ok = data["images"].([]interface{})
	require.True(t, ok, "images must be an array after update")
	assert.Len(t, imagesData, 2)
	for i, img := range imagesData {
		assert.Equal(t, newImages[i], img.(string))
	}
}

// TestUMKM_TimestampsInResponses verifies all responses include created_at and updated_at
func TestUMKM_TimestampsInResponses(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestUMKMCategory(t, "Test")

	// Create UMKM
	resp, err := ts.MakeRequest("POST", "/api/v1/umkm",
		map[string]any{
			"name":        "Timestamp Test UMKM",
			"category":    cat0ID,
			"description": "Testing timestamps",
		},
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

	umkmID := data["id"].(string)
	createdAtStr := data["created_at"].(string)

	// Get UMKM
	resp, err = ts.MakeRequest("GET", "/api/v1/public/umkm/"+umkmID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	// Verify timestamps in retrieval response
	assert.NotEmpty(t, data["created_at"], "created_at must be present in retrieval response")
	assert.NotEmpty(t, data["updated_at"], "updated_at must be present in retrieval response")

	// Parse and compare timestamps to handle format differences
	origCreatedTime, _ := parseTimestamp(createdAtStr)
	retrievedCreatedTime, _ := parseTimestamp(data["created_at"].(string))
	assert.True(t, origCreatedTime.Equal(retrievedCreatedTime), "created_at should not change")

	// Update UMKM
	resp, err = ts.MakeRequest("PUT", "/api/v1/umkm/"+umkmID,
		map[string]any{
			"name":        "Updated UMKM",
			"category":    cat0ID,
			"description": "Updated description",
		},
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

	// Parse and compare timestamps to handle format differences
	updatedCreatedTime, _ := parseTimestamp(data["created_at"].(string))
	assert.True(t, origCreatedTime.Equal(updatedCreatedTime), "created_at should not change on update")
}

// TestUMKM_List_FilterByQuery verifies the admin list endpoint filters by query string
// across name, description, and owner (case-insensitive ILIKE).
func TestUMKM_List_FilterByQuery(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	catID := ts.CreateTestUMKMCategory(t, "Filter Query Test")

	mk := func(name, desc, owner, category string) {
		resp, err := ts.MakeRequest("POST", "/api/v1/umkm",
			map[string]any{
				"name":        name,
				"description": desc,
				"owner":       owner,
				"category":    category,
			}, token,
		)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// name-only match
	mk("Coffee Shop", "Generic brew description.", "Alice", catID)
	// description-only match
	mk("Bakery", "We roast our coffee beans on site every morning.", "Bob", catID)
	// owner-only match
	mk("Stationery", "Paper, pens, notebooks.", "Charlie", catID)
	// negative control
	mk("Tailor", "Suits, shirts, dresses.", "Dave", catID)

	// q matching name
	resp, err := ts.MakeRequest("GET", "/api/v1/umkm?q=Coffee+Shop", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	list := data["umkm"].([]interface{})
	assert.Len(t, list, 1, "name match should return 1")
	assert.Equal(t, "Coffee Shop", list[0].(map[string]interface{})["name"])

	// q matching description
	resp, err = ts.MakeRequest("GET", "/api/v1/umkm?q=roast", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["umkm"].([]interface{})
	assert.Len(t, list, 1, "q should match description column too")
	assert.Equal(t, "Bakery", list[0].(map[string]interface{})["name"])

	// q matching owner
	resp, err = ts.MakeRequest("GET", "/api/v1/umkm?q=Charlie", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["umkm"].([]interface{})
	assert.Len(t, list, 1, "q should match owner column too")
	assert.Equal(t, "Stationery", list[0].(map[string]interface{})["name"])

	// q with multiple matches (Coffee matches name AND description in same row)
	resp, err = ts.MakeRequest("GET", "/api/v1/umkm?q=Coffee", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["umkm"].([]interface{})
	assert.Len(t, list, 2, "q 'Coffee' should match name + description rows")

	// case-insensitive
	resp, err = ts.MakeRequest("GET", "/api/v1/umkm?q=cOfFeE", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["umkm"].([]interface{})
	assert.Len(t, list, 2, "ILIKE should be case-insensitive")

	// no match
	resp, err = ts.MakeRequest("GET", "/api/v1/umkm?q=zzznomatch", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["umkm"].([]interface{})
	assert.Len(t, list, 0, "non-matching q should return empty")
	assert.Equal(t, float64(0), data["pagination"].(map[string]interface{})["total"])
}

// TestUMKM_List_FilterByCategory verifies the admin list endpoint filters by category UUID.
func TestUMKM_List_FilterByCategory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	catA := ts.CreateTestUMKMCategory(t, "Filter A")
	catB := ts.CreateTestUMKMCategory(t, "Filter B")

	mk := func(name, cat string) {
		resp, err := ts.MakeRequest("POST", "/api/v1/umkm",
			map[string]any{
				"name":        name,
				"description": "placeholder",
				"category":    cat,
			}, token,
		)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	mk("UMKM A1", catA)
	mk("UMKM A2", catA)
	mk("UMKM B1", catB)

	// Filter by catA
	resp, err := ts.MakeRequest("GET", "/api/v1/umkm?category="+catA, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	list := data["umkm"].([]interface{})
	assert.Len(t, list, 2)
	assert.Equal(t, float64(2), data["pagination"].(map[string]interface{})["total"])
	for _, item := range list {
		cat := item.(map[string]interface{})["category"].(map[string]interface{})
		assert.Equal(t, catA, cat["id"], "all results must have catA id")
	}

	// Filter by catB
	resp, err = ts.MakeRequest("GET", "/api/v1/umkm?category="+catB, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)
	list = data["umkm"].([]interface{})
	assert.Len(t, list, 1)
	assert.Equal(t, "UMKM B1", list[0].(map[string]interface{})["name"])
}

// TestUMKM_List_CombineQueryAndCategory verifies q+category combine with AND semantics.
func TestUMKM_List_CombineQueryAndCategory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	catA := ts.CreateTestUMKMCategory(t, "Combine A")
	catB := ts.CreateTestUMKMCategory(t, "Combine B")

	mk := func(name, cat string) {
		resp, err := ts.MakeRequest("POST", "/api/v1/umkm",
			map[string]any{
				"name":        name,
				"description": "test",
				"category":    cat,
			}, token,
		)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	mk("Mie Ayam A", catA)   // match name "Mie", catA
	mk("Mie Ayam B", catB)   // match name "Mie", catB
	mk("Bakso A", catA)      // no match in name

	resp, err := ts.MakeRequest("GET", "/api/v1/umkm?q=Mie&category="+catA, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	list := data["umkm"].([]interface{})
	assert.Len(t, list, 1, "combined filter must intersect (AND) q and category")
	assert.Equal(t, "Mie Ayam A", list[0].(map[string]interface{})["name"])
}
