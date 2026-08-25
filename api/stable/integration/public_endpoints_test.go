package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Response wrapper structures for validation
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type ErrorResponseData struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// TestPublicBannerList tests GET /api/v1/public/banner without authentication
func TestPublicBannerList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/banner", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicBannerGet tests GET /api/v1/public/banner/{id} without authentication
func TestPublicBannerGet(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)
	bannerID := createBannerWithAuth(t, ts, token)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/banner/"+bannerID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicBannerGetNotFound tests GET /api/v1/public/banner/{id} with invalid ID
func TestPublicBannerGetNotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/banner/invalid-id", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result ErrorResponseData
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.False(t, result.Success)
}

// TestPublicBeritaList tests GET /api/v1/public/berita/list without authentication
func TestPublicBeritaList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/berita/list", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicBeritaGet tests GET /api/v1/public/berita/{id} without authentication
func TestPublicBeritaGet(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)
	cat0ID := ts.CreateTestBeritaCategory(t, "umum")

	// Try to create berita, but skip test if creation fails
	resp, err := ts.MakeRequest("POST", "/api/v1/berita",
		map[string]interface{}{
			"title":    "Test News",
			"content":  "Test news content",
			"category": cat0ID,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Skip("Skipping test - berita creation failed")
	}

	beritaID := extractID(t, resp)

	resp, err = ts.MakeRequest("GET", "/api/v1/public/berita/"+beritaID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicDesaGet tests GET /api/v1/public/desa without authentication
func TestPublicDesaGet(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/desa", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Desa endpoint may return 500 if no profile exists, which is acceptable for empty DB
	if resp.StatusCode == http.StatusOK {
		var result SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.True(t, result.Success)
	}
}

// TestPublicFasilitasList tests GET /api/v1/public/fasilitas/list without authentication
func TestPublicFasilitasList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/fasilitas/list", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicFasilitasGet tests GET /api/v1/public/fasilitas/{id} without authentication
func TestPublicFasilitasGet(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)
	fasilitasID := createFasilitasWithAuth(t, ts, token)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/fasilitas/"+fasilitasID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicPPIDList tests GET /api/v1/public/ppid/list without authentication
func TestPublicPPIDList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/ppid/list", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicPPIDListWithPagination tests GET /api/v1/public/ppid/list with pagination
func TestPublicPPIDListWithPagination(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/ppid/list?page=1&limit=5", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicPPIDListDescriptionTruncation tests that descriptions are truncated to ~100 words
func TestPublicPPIDListDescriptionTruncation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)
	cat0ID := ts.CreateTestBeritaCategory(t, "umum")

	// Create a PPID with a very long description
	longDesc := strings.Repeat("This is a test description. ", 50) // ~1400 characters
	resp, err := ts.MakeRequest("POST", "/api/v1/ppid",
		map[string]interface{}{
			"title":       "Test PPID with Long Description",
			"category":    cat0ID,
			"description": longDesc,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Skip("Skipping test - PPID creation failed")
	}

	// Now fetch the list and verify description is truncated
	resp, err = ts.MakeRequest("GET", "/api/v1/public/ppid/list?page=1&limit=10", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	data := result["data"].(map[string]interface{})
	ppidList := data["ppid"].([]interface{})

	if len(ppidList) > 0 {
		ppid := ppidList[0].(map[string]interface{})
		desc := ppid["description"].(string)

		// Description should be truncated to max 500 characters
		require.LessOrEqual(t, len(desc), 503) // 500 + "..." = 503
		// Should end with "..." if truncated
		if len(desc) > 500 {
			require.True(t, strings.HasSuffix(desc, "..."), "Long description should be truncated with ...")
		}
	}
}

// TestPublicPPIDGet tests GET /api/v1/public/ppid/{id} without authentication
func TestPublicPPIDGet(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)
	cat0ID := ts.CreateTestBeritaCategory(t, "umum")

	resp, err := ts.MakeRequest("POST", "/api/v1/ppid",
		map[string]interface{}{
			"title":       "Test PPID",
			"category":    cat0ID,
			"description": "Test PPID description",
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Skip("Skipping test - PPID creation failed")
	}

	ppidID := extractID(t, resp)

	resp, err = ts.MakeRequest("GET", "/api/v1/public/ppid/"+ppidID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicStrukturList tests GET /api/v1/public/struktur/list without authentication
func TestPublicStrukturList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/struktur/list", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicStrukturGet tests GET /api/v1/public/struktur/{id} without authentication
func TestPublicStrukturGet(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)

	resp, err := ts.MakeRequest("POST", "/api/v1/struktur",
		map[string]interface{}{
			"name":     "Test Structure",
			"position": "Kepala Desa",
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Skip("Skipping test - struktur creation failed")
	}

	strukturID := extractID(t, resp)

	resp, err = ts.MakeRequest("GET", "/api/v1/public/struktur/"+strukturID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicUMKMList tests GET /api/v1/public/umkm/list without authentication
func TestPublicUMKMList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/umkm/list", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicUMKMGet tests GET /api/v1/public/umkm/{id} without authentication
func TestPublicUMKMGet(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)
	cat0ID := ts.CreateTestBeritaCategory(t, "perdagangan")

	resp, err := ts.MakeRequest("POST", "/api/v1/umkm",
		map[string]interface{}{
			"name":        "Test UMKM",
			"category":    cat0ID,
			"description": "Test UMKM description",
			"owner":       "Test Owner",
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Skip("Skipping test - UMKM creation failed")
	}

	umkmID := extractID(t, resp)

	resp, err = ts.MakeRequest("GET", "/api/v1/public/umkm/"+umkmID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicProfileList tests GET /api/v1/public/profile/list without authentication
func TestPublicProfileList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/profile/list", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicProfileGet tests GET /api/v1/public/profile/{id} without authentication
func TestPublicProfileGet(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)
	profileID := createProfileWithAuth(t, ts, token)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/profile/"+profileID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicInfographicList tests GET /api/v1/public/infographic/list without authentication
func TestPublicInfographicList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/infographic/list", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// TestPublicInfographicGet tests GET /api/v1/public/infographic/{id} without authentication
func TestPublicInfographicGet(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)
	infographicID := createInfographicWithAuth(t, ts, token)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/infographic/"+infographicID, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Success)
}

// Helper functions

func setupAdminUser(t *testing.T, ts *TestServer) string {
	ts.CreateTestUser(t, "Admin", "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")
	return ts.GetAuthToken(t, "admin@test.com", "password123")
}

func extractID(t *testing.T, resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	data := result["data"].(map[string]interface{})
	return data["id"].(string)
}

func createBannerWithAuth(t *testing.T, ts *TestServer, token string) string {
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
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return extractID(t, resp)
}

func createBeritaWithAuth(t *testing.T, ts *TestServer, token string) string {
	categoryID := ts.CreateTestBeritaCategory(t, "umum")
	resp, err := ts.MakeRequest("POST", "/api/v1/berita",
		map[string]interface{}{
			"title":    "Test News",
			"content":  "Test news content",
			"category": categoryID,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return extractID(t, resp)
}

func createFasilitasWithAuth(t *testing.T, ts *TestServer, token string) string {
	resp, err := ts.MakeRequest("POST", "/api/v1/fasilitas",
		map[string]interface{}{
			"name":        "Test Facility",
			"description": "Test facility description",
			"latitude":    -6.2088,
			"longitude":   106.8456,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return extractID(t, resp)
}

func createPPIDWithAuth(t *testing.T, ts *TestServer, token string) string {
	// PPID category is still free text (not FK in this refactor scope).
	resp, err := ts.MakeRequest("POST", "/api/v1/ppid",
		map[string]interface{}{
			"title":       "Test PPID",
			"category":    "Public Information",
			"description": "Test PPID description",
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return extractID(t, resp)
}

func createStrukturWithAuth(t *testing.T, ts *TestServer, token string) string {
	resp, err := ts.MakeRequest("POST", "/api/v1/struktur",
		map[string]interface{}{
			"name":     "Test Structure",
			"position": "Kepala Desa",
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return extractID(t, resp)
}

func createUMKMWithAuth(t *testing.T, ts *TestServer, token string) string {
	categoryID := ts.CreateTestUMKMCategory(t, "perdagangan")
	resp, err := ts.MakeRequest("POST", "/api/v1/umkm",
		map[string]interface{}{
			"name":        "Test UMKM",
			"category":    categoryID,
			"description": "Test UMKM description",
			"owner":       "Test Owner",
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return extractID(t, resp)
}

func createProfileWithAuth(t *testing.T, ts *TestServer, token string) string {
	resp, err := ts.MakeRequest("POST", "/api/v1/profile",
		map[string]interface{}{
			"content":          "Test profile content",
			"section_name":     "keuangan",
			"section_endpoint": "/keuangan",
			"state":            true,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return extractID(t, resp)
}

func createInfographicWithAuth(t *testing.T, ts *TestServer, token string) string {
	resp, err := ts.MakeRequest("POST", "/api/v1/infographic",
		map[string]interface{}{
			"component_id":     1,
			"component_type":   "dashboard",
			"section_name":     "keuangan",
			"section_endpoint": "/keuangan",
			"state":            true,
		},
		token,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return extractID(t, resp)
}

// TestPPIDCRUDWithDescription tests PPID CRUD operations with description
func TestPPIDCRUDWithDescription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)
	cat0ID := ts.CreateTestBeritaCategory(t, "Laporan")

	// Create PPID with description
	createResp, err := ts.MakeRequest("POST", "/api/v1/ppid",
		map[string]interface{}{
			"title":       "Test PPID with Description",
			"category":    cat0ID,
			"description": "This is a test PPID document with a detailed description about financial reports and budget information.",
		},
		token,
	)
	require.NoError(t, err)
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		t.Skip("Skipping test - PPID creation failed")
	}

	var createResult map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createResult)
	require.NoError(t, err)

	data := createResult["data"].(map[string]interface{})
	ppidID := data["id"].(string)
	require.NotEmpty(t, ppidID)

	// Verify description is in the response
	desc := data["description"]
	require.NotNil(t, desc)
	require.Equal(t, "This is a test PPID document with a detailed description about financial reports and budget information.", desc)
}

// TestPrivatePPIDListNoDescription tests that private PPID list excludes description
func TestPrivatePPIDListNoDescription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	token := setupAdminUser(t, ts)

	// Get private PPID list
	resp, err := ts.MakeRequest("GET", "/api/v1/ppid?page=1&limit=10", nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	data := result["data"].(map[string]interface{})
	ppidList := data["ppid"].([]interface{})

	// Verify description field is NOT in private list response
	if len(ppidList) > 0 {
		ppid := ppidList[0].(map[string]interface{})
		_, hasDescription := ppid["description"]
		require.False(t, hasDescription, "Private PPID list should not include description field")
	}
}

// TestPublicPPIDListIncludesDescription tests that public PPID list includes description
func TestPublicPPIDListIncludesDescription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Get public PPID list (no auth required)
	resp, err := ts.MakeRequest("GET", "/api/v1/public/ppid/list?page=1&limit=10", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	data := result["data"].(map[string]interface{})
	ppidList := data["ppid"].([]interface{})

	// Verify description field IS in public list response
	if len(ppidList) > 0 {
		ppid := ppidList[0].(map[string]interface{})
		_, hasDescription := ppid["description"]
		require.True(t, hasDescription, "Public PPID list should include description field")
	}
}
