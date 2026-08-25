package integration

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertFasilitasShape asserts all fields of a fasilitas response object
func assertFasilitasShape(t *testing.T, data map[string]interface{}) {
	t.Helper()
	assert.NotEmpty(t, data["id"], "id must not be empty")
	assert.NotEmpty(t, data["name"], "name must not be empty")
	// category (object or nil) and description are nullable — keys may be absent or nil
	assert.NotNil(t, data["latitude"], "latitude must be present")
	assert.NotNil(t, data["longitude"], "longitude must be present")
	// images is an array field
	assert.NotNil(t, data["images"], "images field must be present")
	assert.NotEmpty(t, data["created_at"], "created_at must not be empty")
	assert.NotEmpty(t, data["updated_at"], "updated_at must not be empty")
}

// assertCategoryObject asserts that data["category"] is an object containing
// the expected id and name. If expectedID is "" (no category), the field
// must be absent or nil.
func assertCategoryObject(t *testing.T, data map[string]interface{}, expectedID, expectedName string) {
	t.Helper()
	if expectedID == "" {
		_, present := data["category"]
		if present {
			cat, _ := data["category"].(map[string]interface{})
			if cat != nil {
				t.Errorf("expected no category object, got %v", cat)
			}
		}
		return
	}
	cat, ok := data["category"].(map[string]interface{})
	require.True(t, ok, "category must be an object {id, name} when category_id is set")
	assert.Equal(t, expectedID, cat["id"], "category.id must match")
	assert.Equal(t, expectedName, cat["name"], "category.name must match")
}

func TestFasilitas_HappyPath(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	facilityType := "Health"
	description := "Test Facility Description"
	healthCatID := ts.CreateTestFasilitasCategory(t, facilityType)

	// Create Fasilitas
	createPayload := map[string]any{
		"name":        "Test Facility",
		"category":    healthCatID,
		"latitude":    -6.2088,
		"longitude":   106.8456,
		"description": description,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas", createPayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator := NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "name", "Test Facility")
	validator.AssertFieldValue(t, "category_id", healthCatID)
	validator.AssertFieldValue(t, "description", description)
	validator.AssertFieldIsUUID(t, "id")
	validator.AssertFieldIsTimestamp(t, "created_at")
	validator.AssertFieldIsTimestamp(t, "updated_at")
	assertCategoryObject(t, validator.Data, healthCatID, facilityType)
	fasilitasID := validator.Data["id"].(string)

	// List Fasilitas (admin)
	resp, err = ts.MakeRequest("GET", "/api/v1/fasilitas", nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldExists(t, "fasilitas")
	validator.AssertFieldExists(t, "pagination")
	validator.AssertFieldType(t, "fasilitas", "array")

	fasilitasList := validator.Data["fasilitas"].([]interface{})
	assert.Len(t, fasilitasList, 1)

	// Get Fasilitas by ID (admin)
	resp, err = ts.MakeRequest("GET", "/api/v1/fasilitas/"+fasilitasID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", fasilitasID)
	validator.AssertFieldValue(t, "name", "Test Facility")
	validator.AssertFieldValue(t, "category_id", healthCatID)
	assertCategoryObject(t, validator.Data, healthCatID, facilityType)

	// Update Fasilitas
	newType := "Education"
	eduCatID := ts.CreateTestFasilitasCategory(t, newType)
	updatePayload := map[string]any{
		"name":      "Updated Facility",
		"category":  eduCatID,
		"latitude":  -6.3000,
		"longitude": 106.9000,
	}

	resp, err = ts.MakeRequest("PUT", "/api/v1/fasilitas/"+fasilitasID, updatePayload, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
	validator.AssertFieldValue(t, "id", fasilitasID)
	validator.AssertFieldValue(t, "name", "Updated Facility")
	validator.AssertFieldValue(t, "category_id", eduCatID)
	assertCategoryObject(t, validator.Data, eduCatID, newType)

	// Delete Fasilitas
	resp, err = ts.MakeRequest("DELETE", "/api/v1/fasilitas/"+fasilitasID, nil, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	validator = NewResponseValidator(body)
	validator.AssertSuccess(t, true)
}

func TestFasilitas_CreateWithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas",
		map[string]any{"name": "Test Facility", "latitude": -6.2088, "longitude": 106.8456}, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestFasilitas_CreateWithEmptyName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas",
		map[string]any{"name": "", "latitude": -6.2088, "longitude": 106.8456}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestFasilitas_CreateWithInvalidCoordinates(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas",
		map[string]any{"name": "Test Facility", "latitude": 999.0, "longitude": 999.0}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestFasilitas_GetNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/fasilitas/00000000-0000-0000-0000-000000000000", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestFasilitas_GetWithInvalidID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/fasilitas/invalid-uuid", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestFasilitas_UpdateNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("PUT", "/api/v1/fasilitas/00000000-0000-0000-0000-000000000000",
		map[string]any{"name": "Updated Facility", "latitude": -6.2088, "longitude": 106.8456}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestFasilitas_DeleteNonExistent(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/fasilitas/00000000-0000-0000-0000-000000000000", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestFasilitas_ListWithInvalidPage(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/fasilitas/list?page=0&limit=10", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

func TestFasilitas_ListExceedsMaxLimit(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("GET", "/api/v1/fasilitas?page=1&limit=101", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, body["success"].(bool))
}

// TestFasilitas_ImagesArrayField verifies images array field is properly stored and retrieved.
// Images are sent via multipart/form-data because the JSON request path validates image MIME types.
func TestFasilitas_ImagesArrayField(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestFasilitasCategory(t, "school")

	// Create Fasilitas with one multipart image upload.
	resp, err := ts.MakeMultipartRequest("POST", "/api/v1/fasilitas",
		map[string]string{
			"name":      "School Building",
			"latitude":  "-6.2088",
			"longitude": "106.8456",
			"category":  cat0ID,
		},
		map[string][2]string{"images": {"school1.jpg", string(tinyJPEG())}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)

	// Verify images in creation response — server stores with new UUID names.
	imagesData, ok := data["images"].([]interface{})
	require.True(t, ok, "images must be an array")
	assert.Len(t, imagesData, 1)
	for _, img := range imagesData {
		s, ok := img.(string)
		require.True(t, ok, "image entry must be a string")
		assert.True(t, strings.HasPrefix(s, "/uploads/"), "image URL should be served from /uploads/, got %s", s)
	}

	fasilitasID := data["id"].(string)

	// Get Fasilitas and verify image is persisted.
	resp, err = ts.MakeRequest("GET", "/api/v1/fasilitas/"+fasilitasID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	imagesData, ok = data["images"].([]interface{})
	require.True(t, ok, "images must be an array in retrieval")
	assert.Len(t, imagesData, 1)

	// Update Fasilitas with a new image via multipart.
	resp, err = ts.MakeMultipartRequest("PUT", "/api/v1/fasilitas/"+fasilitasID,
		map[string]string{
			"name":      "School Building",
			"latitude":  "-6.2088",
			"longitude": "106.8456",
			"category":  cat0ID,
		},
		map[string][2]string{"images": {"school_new.jpg", string(tinyJPEG())}},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	// Verify updated images array has exactly one entry now.
	imagesData, ok = data["images"].([]interface{})
	require.True(t, ok, "images must be an array after update")
	assert.Len(t, imagesData, 1)
}

// TestFasilitas_EmptyImagesArray verifies empty images array is handled correctly
func TestFasilitas_EmptyImagesArray(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestFasilitasCategory(t, "health")

	// Create Fasilitas without images
	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas",
		map[string]any{
			"name":      "Health Center",
			"latitude":  -6.2100,
			"longitude": 106.8500,
			"category":  cat0ID,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	require.NotNil(t, data, "response data must not be nil")

	// Verify images field is present (can be empty or null)
	assert.NotNil(t, data["images"], "images field must be present")

	fasilitasID := data["id"].(string)

	// Get Fasilitas and verify images field is present
	resp, err = ts.MakeRequest("GET", "/api/v1/public/fasilitas/"+fasilitasID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	assert.NotNil(t, data["images"], "images field must be present in retrieval")
}

// TestFasilitas_TimestampsInResponses verifies all responses include created_at and updated_at
func TestFasilitas_TimestampsInResponses(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat0ID := ts.CreateTestFasilitasCategory(t, "mosque")

	// Create Fasilitas
	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas",
		map[string]any{
			"name":      "Mosque",
			"latitude":  -6.2150,
			"longitude": 106.8550,
			"category":  cat0ID,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	require.NotNil(t, data, "response data must not be nil")

	// Verify timestamps in creation response
	assert.NotEmpty(t, data["created_at"], "created_at must be present in creation response")
	assert.NotEmpty(t, data["updated_at"], "updated_at must be present in creation response")

	fasilitasID := data["id"].(string)
	createdAtStr := data["created_at"].(string)
	createdAt, err := parseTimestamp(createdAtStr)
	require.NoError(t, err)

	// Get Fasilitas
	resp, err = ts.MakeRequest("GET", "/api/v1/public/fasilitas/"+fasilitasID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data = getDataField(body)

	// Verify timestamps in retrieval response
	assert.NotEmpty(t, data["created_at"], "created_at must be present in retrieval response")
	assert.NotEmpty(t, data["updated_at"], "updated_at must be present in retrieval response")
	retrievedCreatedAt, err := parseTimestamp(data["created_at"].(string))
	require.NoError(t, err)
	assert.True(t, createdAt.Equal(retrievedCreatedAt), "created_at should not change")

	// Update Fasilitas
	resp, err = ts.MakeRequest("PUT", "/api/v1/fasilitas/"+fasilitasID,
		map[string]any{
			"name":      "Mosque Updated",
			"latitude":  -6.2150,
			"longitude": 106.8550,
			"category":  cat0ID,
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
}

// TestFasilitasList_FilterByQuery verifies the q query parameter filters fasilitas
// by ILIKE match on name.
func TestFasilitasList_FilterByQuery(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create three fasilitas with distinct names.
	for _, name := range []string{
		"Sekolah Dasar Negeri 1",
		"Pusat Kesehatan Masyarakat",
		"Masjid Al-Ikhlas",
	} {
		resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas",
			map[string]any{"name": name, "latitude": -6.2, "longitude": 106.8}, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}

	// Search for a substring present in exactly one name.
	resp, err := ts.MakeRequest("GET", "/api/v1/fasilitas?q=Masjid", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := getDataField(body)
	fasilitases := data["fasilitas"].([]interface{})
	require.Len(t, fasilitases, 1, "q=Masjid must match exactly one fasilitas")

	name, ok := fasilitases[0].(map[string]interface{})["name"].(string)
	require.True(t, ok, "name must be a string")
	assert.Contains(t, name, "Masjid")
}

// TestFasilitasList_FilterByCategory verifies the category query parameter filters
// fasilitas by their category UUID.
func TestFasilitasList_FilterByCategory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create two distinct categories.
	catSchool := ts.CreateTestFasilitasCategory(t, "School")
	catHealth := ts.CreateTestFasilitasCategory(t, "Health")

	// Create fasilitas across the two categories.
	for _, name := range []string{"SDN 1", "SMP 2"} {
		resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas",
			map[string]any{"name": name, "latitude": -6.2, "longitude": 106.8, "category": catSchool}, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}
	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas",
		map[string]any{"name": "Puskesmas", "latitude": -6.21, "longitude": 106.81, "category": catHealth}, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Filter by the School category.
	resp, err = ts.MakeRequest("GET", "/api/v1/fasilitas?category="+catSchool, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data := getDataField(body)
	fasilitases := data["fasilitas"].([]interface{})
	require.Len(t, fasilitases, 2, "category filter must return only School fasilitas")

	for _, f := range fasilitases {
		fm := f.(map[string]interface{})
		catID, _ := fm["category_id"].(string)
		assert.Equal(t, catSchool, catID, "each returned fasilitas must belong to the filtered category")
	}

	// Filter by the Health category.
	resp, err = ts.MakeRequest("GET", "/api/v1/fasilitas?category="+catHealth, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = parseJSONResponse(resp.Body)
	require.NoError(t, err)

	data = getDataField(body)
	fasilitases = data["fasilitas"].([]interface{})
	require.Len(t, fasilitases, 1, "category filter must return only the Health fasilitas")
	fm := fasilitases[0].(map[string]interface{})
	catID, _ := fm["category_id"].(string)
	assert.Equal(t, catHealth, catID)
}
