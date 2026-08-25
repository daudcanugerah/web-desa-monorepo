package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// genUUID returns a fresh UUID for direct SQL inserts in tests that bypass
// the API (e.g. seeding FK relations without uploading files).
func genUUID(t *testing.T, salt string) string {
	t.Helper()
	// Salt keeps IDs deterministic when a test inserts multiple rows and wants
	// readable IDs during debugging; collisions across runs don't matter.
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(salt)).String()
}

// TestPPIDCategory_List_Success verifies the public list endpoint for PPID.
func TestPPIDCategory_List_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestPPIDCategory(t, "Anggaran")
	ts.CreateTestPPIDCategory(t, "Peraturan")

	resp, err := ts.MakeRequest("GET", "/api/v1/public/ppid/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := parseJSONResponse(resp.Body)
	assert.True(t, body["success"].(bool))
	cats := getDataField(body)["categories"].([]interface{})
	assert.GreaterOrEqual(t, len(cats), 2)
	// Sorted alphabetically
	assert.Equal(t, "Anggaran", cats[0].(map[string]interface{})["name"])
	assert.Equal(t, "Peraturan", cats[1].(map[string]interface{})["name"])
}

// TestPPIDCategory_SearchAutocomplete verifies text search.
func TestPPIDCategory_SearchAutocomplete(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestPPIDCategory(t, "Anggaran")
	ts.CreateTestPPIDCategory(t, "Laporan")
	ts.CreateTestPPIDCategory(t, "Peraturan")

	resp, err := ts.MakeRequest("GET", "/api/v1/public/ppid/categories?q=ang", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})
	assert.Len(t, cats, 1)
	assert.Equal(t, "Anggaran", cats[0].(map[string]interface{})["name"])
}

// TestPPIDCategory_Pagination verifies page/limit.
func TestPPIDCategory_Pagination(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	for _, n := range []string{"A", "B", "C", "D", "E"} {
		ts.CreateTestPPIDCategory(t, n)
	}

	resp, err := ts.MakeRequest("GET", "/api/v1/public/ppid/categories?page=2&limit=2", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := parseJSONResponse(resp.Body)
	data := getDataField(body)
	cats := data["categories"].([]interface{})
	assert.Len(t, cats, 2)
	assert.Equal(t, "C", cats[0].(map[string]interface{})["name"])
	assert.Equal(t, "D", cats[1].(map[string]interface{})["name"])
}

// TestPPIDCategory_Create_Success verifies create.
func TestPPIDCategory_Create_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	resp, err := ts.MakeRequest("POST", "/api/v1/ppid/categories",
		map[string]string{"name": "Dokumen"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestPPIDCategory_Create_WithoutAuth verifies the endpoint is protected.
func TestPPIDCategory_Create_WithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("POST", "/api/v1/ppid/categories",
		map[string]string{"name": "Dokumen"}, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestPPIDCategory_Create_DuplicateName returns 409.
func TestPPIDCategory_Create_DuplicateName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	ts.CreateTestPPIDCategory(t, "Dokumen")
	resp, err := ts.MakeRequest("POST", "/api/v1/ppid/categories",
		map[string]string{"name": "Dokumen"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestPPIDCategory_Delete_InUse returns 409 when referenced.
func TestPPIDCategory_Delete_InUse(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	categoryID := ts.CreateTestPPIDCategory(t, "InUse")
	// Create a PPID referencing the category (skip the file upload — we just need
	// the row to exist for the FK check).
	_, err := ts.DB.Exec(
		`INSERT INTO ppid (id, title, category, file_url, created_at, updated_at) VALUES ($1, $2, $3, '/tmp/dummy.pdf', NOW(), NOW())`,
		genUUID(t, "11111111-1111-1111-1111-111111111111"), "Test PPID", categoryID,
	)
	require.NoError(t, err)

	resp, err := ts.MakeRequest("DELETE", "/api/v1/ppid/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestPPIDCategory_Delete_Success verifies delete.
func TestPPIDCategory_Delete_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	categoryID := ts.CreateTestPPIDCategory(t, "Unused")

	resp, err := ts.MakeRequest("DELETE", "/api/v1/ppid/categories/"+categoryID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestPPIDCategory_UsageCount verifies the count is accurate.
func TestPPIDCategory_UsageCount(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	cat1 := ts.CreateTestPPIDCategory(t, "WithRefs")
	_ = ts.CreateTestPPIDCategory(t, "NoRefs")

	// Insert 2 PPIDs referencing cat1
	for _, salt := range []string{"p1-salt", "p2-salt"} {
		_, err := ts.DB.Exec(
			`INSERT INTO ppid (id, title, category, file_url, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())`,
			genUUID(t, salt), "Title "+salt, cat1, "/tmp/dummy.pdf",
		)
		require.NoError(t, err)
	}

	resp, err := ts.MakeRequest("GET", "/api/v1/public/ppid/categories", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := parseJSONResponse(resp.Body)
	cats := getDataField(body)["categories"].([]interface{})

	counts := map[string]int{}
	for _, c := range cats {
		cm := c.(map[string]interface{})
		counts[cm["name"].(string)] = int(cm["usage_count"].(float64))
	}
	assert.Equal(t, 2, counts["WithRefs"])
	assert.Equal(t, 0, counts["NoRefs"])
}
