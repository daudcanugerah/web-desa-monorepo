package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type galleryUploadFile struct {
	field       string
	filename    string
	contentType string
	content     []byte
}

func setupGalleryAdmin(t *testing.T) (*TestServer, string) {
	t.Helper()
	ts := SetupTestServer(t)
	ts.CreateTestUser(t, "Gallery Admin", "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")
	return ts, ts.AdminToken(t)
}

func tinyGalleryJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, A: 255})
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}))
	return buf.Bytes()
}

func tinyGalleryPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, A: 255})
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func galleryMultipartRequest(t *testing.T, ts *TestServer, method, path, token string, files ...galleryUploadFile) (*http.Response, error) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, file := range files {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, file.field, file.filename))
		header.Set("Content-Type", file.contentType)
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(file.content); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, ts.Server.URL+path, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	return http.DefaultClient.Do(req)
}

func galleryResponseData(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	data := getDataField(body)
	require.NotNil(t, data)
	return data
}

func createGalleryFolder(t *testing.T, ts *TestServer, token, name string, isPublic *bool) (string, map[string]interface{}) {
	t.Helper()
	payload := map[string]interface{}{"name": name}
	if isPublic != nil {
		payload["is_public"] = *isPublic
	}
	resp, err := ts.MakeRequest("POST", "/api/v1/gallery/folders", payload, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	data := galleryResponseData(t, resp)
	id, ok := data["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, id)
	return id, data
}

func uploadGalleryMedia(t *testing.T, ts *TestServer, folderID, token, field, filename, contentType string, content []byte) (string, map[string]interface{}) {
	t.Helper()
	resp, err := galleryMultipartRequest(t, ts, "POST", "/api/v1/gallery/folders/"+folderID+"/media", token, galleryUploadFile{
		field:       field,
		filename:    filename,
		contentType: contentType,
		content:     content,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	data := galleryResponseData(t, resp)
	uploaded, ok := data["uploaded"].([]interface{})
	require.True(t, ok)
	require.Len(t, uploaded, 1)
	media, ok := uploaded[0].(map[string]interface{})
	require.True(t, ok)
	id, ok := media["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, id)
	return id, media
}

func updateGalleryFolderVisibility(t *testing.T, ts *TestServer, token, folderID string, isPublic bool) map[string]interface{} {
	t.Helper()
	resp, err := ts.MakeRequest("PATCH", "/api/v1/gallery/folders/"+folderID+"/visibility", map[string]interface{}{"is_public": isPublic}, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return galleryResponseData(t, resp)
}

func updateGalleryMediaVisibility(t *testing.T, ts *TestServer, token, mediaID string, isPublic bool) map[string]interface{} {
	t.Helper()
	resp, err := ts.MakeRequest("PATCH", "/api/v1/gallery/media/"+mediaID+"/visibility", map[string]interface{}{"is_public": isPublic}, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return galleryResponseData(t, resp)
}

func deleteGalleryMedia(t *testing.T, ts *TestServer, token, mediaID string) {
	t.Helper()
	resp, err := ts.MakeRequest("DELETE", "/api/v1/gallery/media/"+mediaID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func adminGalleryFolder(t *testing.T, ts *TestServer, token, folderID string) map[string]interface{} {
	t.Helper()
	resp, err := ts.MakeRequest("GET", "/api/v1/gallery/folders/"+folderID, nil, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := galleryResponseData(t, resp)
	folder, ok := data["folder"].(map[string]interface{})
	require.True(t, ok)
	return folder
}

func galleryCoverID(t *testing.T, folder map[string]interface{}) string {
	t.Helper()
	value, ok := folder["cover_media_id"]
	if !ok || value == nil {
		return ""
	}
	cover, ok := value.(string)
	require.True(t, ok)
	return cover
}

func publicGalleryMediaRequest(t *testing.T, ts *TestServer, method, path string) (*http.Response, []byte) {
	t.Helper()
	resp, err := ts.MakeRequest(method, path, nil, "")
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	return resp, body
}

func assertGalleryPublicMedia(t *testing.T, ts *TestServer, mediaID string, expected []byte) {
	t.Helper()
	resp, body := publicGalleryMediaRequest(t, ts, "GET", "/api/v1/public/gallery/media/"+mediaID)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var envelope map[string]interface{}
	require.NoError(t, jsonUnmarshalGallery(body, &envelope))
	data := getDataField(envelope)
	require.NotNil(t, data)
	assert.Equal(t, mediaID, data["id"])

	resp, body = publicGalleryMediaRequest(t, ts, "GET", "/api/v1/public/gallery/media/"+mediaID+"/content")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expected, body)
	assert.Equal(t, "image/jpeg", resp.Header.Get("Content-Type"))

	resp, body = publicGalleryMediaRequest(t, ts, "GET", "/api/v1/public/gallery/media/"+mediaID+"/thumbnail")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, body)
	assert.Equal(t, "image/webp", resp.Header.Get("Content-Type"))
}

func assertGalleryPublicMediaHidden(t *testing.T, ts *TestServer, mediaID string) {
	t.Helper()
	paths := []string{
		"/api/v1/public/gallery/media/" + mediaID,
		"/api/v1/public/gallery/media/" + mediaID + "/content",
		"/api/v1/public/gallery/media/" + mediaID + "/thumbnail",
	}
	for _, path := range paths {
		resp, err := ts.MakeRequest("GET", path, nil, "")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, path)
	}
}

func galleryUploadResult(t *testing.T, data map[string]interface{}) map[string]interface{} {
	t.Helper()
	results, ok := data["results"].([]interface{})
	require.True(t, ok)
	require.Len(t, results, 1)
	result, ok := results[0].(map[string]interface{})
	require.True(t, ok)
	return result
}

func jsonUnmarshalGallery(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

func TestGalleryProtectedRoutesRequireAuthentication(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/gallery/folders"},
		{"GET", "/api/v1/gallery/folders/00000000-0000-0000-0000-000000000001"},
		{"GET", "/api/v1/gallery/media"},
		{"GET", "/api/v1/gallery/media/00000000-0000-0000-0000-000000000001"},
		{"GET", "/api/v1/gallery/media/00000000-0000-0000-0000-000000000001/content"},
		{"GET", "/api/v1/gallery/media/00000000-0000-0000-0000-000000000001/thumbnail"},
		{"POST", "/api/v1/gallery/folders"},
		{"PUT", "/api/v1/gallery/folders/00000000-0000-0000-0000-000000000001"},
		{"PATCH", "/api/v1/gallery/folders/00000000-0000-0000-0000-000000000001/visibility"},
		{"DELETE", "/api/v1/gallery/folders/00000000-0000-0000-0000-000000000001"},
		{"POST", "/api/v1/gallery/folders/00000000-0000-0000-0000-000000000001/media"},
		{"PATCH", "/api/v1/gallery/media/00000000-0000-0000-0000-000000000001/visibility"},
		{"POST", "/api/v1/gallery/folders/00000000-0000-0000-0000-000000000001/media/visibility"},
		{"DELETE", "/api/v1/gallery/media/00000000-0000-0000-0000-000000000001"},
		{"POST", "/api/v1/gallery/media/00000000-0000-0000-0000-000000000001/regenerate-thumbnail"},
	}
	for _, route := range routes {
		var resp *http.Response
		var err error
		if route.path == "/api/v1/gallery/folders/00000000-0000-0000-0000-000000000001/media" {
			resp, err = galleryMultipartRequest(t, ts, route.method, route.path, "", galleryUploadFile{field: "media[]", filename: "image.jpg", contentType: "image/jpeg", content: tinyGalleryJPEG(t)})
		} else {
			resp, err = ts.MakeRequest(route.method, route.path, nil, "")
		}
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, route.method+" "+route.path)
	}
}

func TestGalleryFolderCreateDefaultsPrivateDuplicateListFilterUpdateDelete(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, folder := createGalleryFolder(t, ts, token, "Case Filter Folder", nil)
	assert.False(t, folder["is_public"].(bool))
	assert.NotEmpty(t, folder["created_by"])

	resp, err := ts.MakeRequest("POST", "/api/v1/gallery/folders", map[string]interface{}{"name": "case filter folder"}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	public := true
	createGalleryFolder(t, ts, token, "Public Case Folder", &public)
	resp, err = ts.MakeRequest("GET", "/api/v1/gallery/folders?is_public=false&q=case&limit=100", nil, token)
	require.NoError(t, err)
	data := galleryResponseData(t, resp)
	folders, ok := data["folders"].([]interface{})
	require.True(t, ok)
	require.Len(t, folders, 1)
	filtered, ok := folders[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, folderID, filtered["id"])
	assert.False(t, filtered["is_public"].(bool))

	resp, err = ts.MakeRequest("PUT", "/api/v1/gallery/folders/"+folderID, map[string]interface{}{"name": "Updated Case Folder", "description": "Updated description"}, token)
	require.NoError(t, err)
	updated := galleryResponseData(t, resp)
	assert.Equal(t, "Updated Case Folder", updated["name"])
	assert.Equal(t, "Updated description", updated["description"])

	resp, err = ts.MakeRequest("DELETE", "/api/v1/gallery/folders/"+folderID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = ts.MakeRequest("GET", "/api/v1/gallery/folders/"+folderID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGalleryUploadAliasesThumbnailMIMEValidationAndPrivateDefaults(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Upload Validation", nil)
	jpegBytes := tinyGalleryJPEG(t)
	jpegID, jpegMedia := uploadGalleryMedia(t, ts, folderID, token, "media[]", "valid.jpg", "image/jpeg", jpegBytes)
	assert.Equal(t, "image", jpegMedia["media_type"])
	assert.False(t, jpegMedia["is_public"].(bool))
	assert.NotEmpty(t, jpegMedia["thumbnail_url"])
	assert.NotEmpty(t, jpegMedia["content_url"])

	pngID, pngMedia := uploadGalleryMedia(t, ts, folderID, token, "media", "valid.png", "image/png", tinyGalleryPNG(t))
	assert.Equal(t, "image", pngMedia["media_type"])
	assert.False(t, pngMedia["is_public"].(bool))
	assert.NotEmpty(t, pngMedia["thumbnail_url"])

	var thumbnailURL sql.NullString
	require.NoError(t, ts.DB.QueryRow("SELECT thumbnail_url FROM gallery_media WHERE id = $1", jpegID).Scan(&thumbnailURL))
	require.True(t, thumbnailURL.Valid)
	_, err := os.Stat(filepath.Join("test_uploads", "private", "gallery", "thumbnails", thumbnailURL.String))
	require.NoError(t, err)

	resp, err := galleryMultipartRequest(t, ts, "POST", "/api/v1/gallery/folders/"+folderID+"/media", token, galleryUploadFile{field: "media", filename: "unsupported.txt", contentType: "text/plain", content: []byte("not an image")})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	unsupported := galleryUploadResult(t, galleryResponseData(t, resp))
	assert.Equal(t, "error", unsupported["status"])
	assert.Equal(t, float64(http.StatusUnsupportedMediaType), unsupported["code"])

	resp, err = galleryMultipartRequest(t, ts, "POST", "/api/v1/gallery/folders/"+folderID+"/media", token, galleryUploadFile{field: "media[]", filename: "forged.jpg", contentType: "image/jpeg", content: tinyGalleryPNG(t)})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	forged := galleryUploadResult(t, galleryResponseData(t, resp))
	assert.Equal(t, "error", forged["status"])
	assert.Equal(t, float64(http.StatusUnsupportedMediaType), forged["code"])

	var mediaCount int
	require.NoError(t, ts.DB.QueryRow("SELECT COUNT(*) FROM gallery_media WHERE folder_id = $1", folderID).Scan(&mediaCount))
	assert.Equal(t, 2, mediaCount)
	assert.NotEmpty(t, pngID)
}

func TestGalleryPublicMatrix(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	// Public visibility is folder-level: a media item is visible whenever its
	// folder is public, regardless of the item's own is_public flag.
	cases := []struct {
		name         string
		folderPublic bool
		mediaPublic  bool
		visible      bool
	}{
		{"private_private", false, false, false},
		{"private_public", false, true, false},
		{"public_private", true, false, true},
		{"public_public", true, true, true},
	}
	jpegBytes := tinyGalleryJPEG(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			folderID, _ := createGalleryFolder(t, ts, token, "Matrix "+tc.name, &tc.folderPublic)
			mediaID, _ := uploadGalleryMedia(t, ts, folderID, token, "media[]", tc.name+".jpg", "image/jpeg", jpegBytes)
			if tc.mediaPublic {
				updateGalleryMediaVisibility(t, ts, token, mediaID, true)
			}
			if tc.visible {
				assertGalleryPublicMedia(t, ts, mediaID, jpegBytes)
				resp, err := ts.MakeRequest("GET", "/api/v1/public/gallery/folders/"+folderID, nil, "")
				require.NoError(t, err)
				resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			} else {
				assertGalleryPublicMediaHidden(t, ts, mediaID)
				resp, err := ts.MakeRequest("GET", "/api/v1/public/gallery/folders/"+folderID, nil, "")
				require.NoError(t, err)
				resp.Body.Close()
				assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			}
		})
	}
}

func TestGalleryFolderVisibilityHidesAndRepublishesMedia(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	public := true
	folderID, _ := createGalleryFolder(t, ts, token, "Visibility Toggle", &public)
	jpegBytes := tinyGalleryJPEG(t)
	mediaID, _ := uploadGalleryMedia(t, ts, folderID, token, "media[]", "toggle.jpg", "image/jpeg", jpegBytes)
	updateGalleryMediaVisibility(t, ts, token, mediaID, true)
	assertGalleryPublicMedia(t, ts, mediaID, jpegBytes)

	privateFolder := updateGalleryFolderVisibility(t, ts, token, folderID, false)
	assert.False(t, privateFolder["is_public"].(bool))
	assertGalleryPublicMediaHidden(t, ts, mediaID)
	adminMediaResp, err := ts.MakeRequest("GET", "/api/v1/gallery/media/"+mediaID, nil, token)
	require.NoError(t, err)
	adminMedia := galleryResponseData(t, adminMediaResp)
	assert.True(t, adminMedia["is_public"].(bool))

	publishedFolder := updateGalleryFolderVisibility(t, ts, token, folderID, true)
	assert.True(t, publishedFolder["is_public"].(bool))
	assertGalleryPublicMedia(t, ts, mediaID, jpegBytes)
}

func TestGalleryFolderVisibilityRequiresIsPublic(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Missing Folder Visibility", nil)
	resp, err := ts.MakeRequest("PATCH", "/api/v1/gallery/folders/"+folderID+"/visibility", map[string]interface{}{}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGalleryMediaVisibilityRequiresIsPublic(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Missing Media Visibility", nil)
	mediaID, _ := uploadGalleryMedia(t, ts, folderID, token, "media[]", "missing-media-visibility.jpg", "image/jpeg", tinyGalleryJPEG(t))
	resp, err := ts.MakeRequest("PATCH", "/api/v1/gallery/media/"+mediaID+"/visibility", map[string]interface{}{}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGalleryBulkVisibilityRequiresIsPublic(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Missing Bulk Visibility", nil)
	mediaID, _ := uploadGalleryMedia(t, ts, folderID, token, "media[]", "missing-bulk-visibility.jpg", "image/jpeg", tinyGalleryJPEG(t))
	resp, err := ts.MakeRequest("POST", "/api/v1/gallery/folders/"+folderID+"/media/visibility", map[string]interface{}{"media_ids": []string{mediaID}}, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGalleryMixedOrderMultipartResults(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Mixed Multipart Order", nil)
	resp, err := galleryMultipartRequest(t, ts, "POST", "/api/v1/gallery/folders/"+folderID+"/media", token,
		galleryUploadFile{field: "media[]", filename: "first.jpg", contentType: "image/jpeg", content: tinyGalleryJPEG(t)},
		galleryUploadFile{field: "media[]", filename: "forged.jpg", contentType: "image/jpeg", content: tinyGalleryPNG(t)},
		galleryUploadFile{field: "media[]", filename: "third.png", contentType: "image/png", content: tinyGalleryPNG(t)},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	data := galleryResponseData(t, resp)
	results, ok := data["results"].([]interface{})
	require.True(t, ok)
	require.Len(t, results, 3)

	first := results[0].(map[string]interface{})
	assert.Equal(t, "first.jpg", first["filename"])
	assert.Equal(t, "ok", first["status"])
	firstMedia := first["media"].(map[string]interface{})
	assert.NotEmpty(t, firstMedia["id"])

	forged := results[1].(map[string]interface{})
	assert.Equal(t, "forged.jpg", forged["filename"])
	assert.Equal(t, "error", forged["status"])
	assert.Equal(t, float64(http.StatusUnsupportedMediaType), forged["code"])

	third := results[2].(map[string]interface{})
	assert.Equal(t, "third.png", third["filename"])
	assert.Equal(t, "ok", third["status"])
	thirdMedia := third["media"].(map[string]interface{})
	assert.NotEmpty(t, thirdMedia["id"])

	uploaded, ok := data["uploaded"].([]interface{})
	require.True(t, ok)
	require.Len(t, uploaded, 2)
	assert.Equal(t, firstMedia["id"], uploaded[0].(map[string]interface{})["id"])
	assert.Equal(t, thirdMedia["id"], uploaded[1].(map[string]interface{})["id"])

	var mediaCount int
	require.NoError(t, ts.DB.QueryRow("SELECT COUNT(*) FROM gallery_media WHERE folder_id = $1", folderID).Scan(&mediaCount))
	assert.Equal(t, 2, mediaCount)
}

func TestGalleryPublicListDetailFiltersMediaAndHidesUploader(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	public := true
	folderID, _ := createGalleryFolder(t, ts, token, "Public Eligible", &public)
	jpegBytes := tinyGalleryJPEG(t)
	firstID, _ := uploadGalleryMedia(t, ts, folderID, token, "media[]", "first.jpg", "image/jpeg", jpegBytes)
	secondID, _ := uploadGalleryMedia(t, ts, folderID, token, "media", "second.jpg", "image/jpeg", jpegBytes)

	// A private folder is never listed, even when it holds a public media item.
	privateFolderID, _ := createGalleryFolder(t, ts, token, "Private With Public Media", nil)
	privateMediaID, _ := uploadGalleryMedia(t, ts, privateFolderID, token, "media[]", "private.jpg", "image/jpeg", jpegBytes)
	updateGalleryMediaVisibility(t, ts, token, privateMediaID, true)

	resp, err := ts.MakeRequest("GET", "/api/v1/public/gallery/folders", nil, "")
	require.NoError(t, err)
	data := galleryResponseData(t, resp)
	folders, ok := data["folders"].([]interface{})
	require.True(t, ok)
	require.Len(t, folders, 1)
	publicFolder, ok := folders[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, folderID, publicFolder["id"])
	assert.NotContains(t, publicFolder, "uploaded_by")

	// Folder-level gate: a public folder exposes ALL of its media, regardless
	// of each item's own is_public flag.
	resp, err = ts.MakeRequest("GET", "/api/v1/public/gallery/folders/"+folderID, nil, "")
	require.NoError(t, err)
	data = galleryResponseData(t, resp)
	mediaList, ok := data["media"].([]interface{})
	require.True(t, ok)
	require.Len(t, mediaList, 2)
	gotIDs := map[string]bool{}
	for _, raw := range mediaList {
		m := raw.(map[string]interface{})
		gotIDs[m["id"].(string)] = true
		assert.NotContains(t, m, "uploaded_by")
	}
	assert.True(t, gotIDs[firstID])
	assert.True(t, gotIDs[secondID])
}

func TestGalleryBulkMediaVisibility(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Bulk Visibility", nil)
	jpegBytes := tinyGalleryJPEG(t)
	firstID, _ := uploadGalleryMedia(t, ts, folderID, token, "media[]", "first.jpg", "image/jpeg", jpegBytes)
	secondID, _ := uploadGalleryMedia(t, ts, folderID, token, "media", "second.jpg", "image/jpeg", jpegBytes)

	resp, err := ts.MakeRequest("POST", "/api/v1/gallery/folders/"+folderID+"/media/visibility", map[string]interface{}{"media_ids": []string{firstID, secondID}, "is_public": true}, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := galleryResponseData(t, resp)
	mediaList, ok := data["media"].([]interface{})
	require.True(t, ok)
	require.Len(t, mediaList, 2)
	for _, item := range mediaList {
		media := item.(map[string]interface{})
		assert.True(t, media["is_public"].(bool))
	}

	var publicCount int
	require.NoError(t, ts.DB.QueryRow("SELECT COUNT(*) FROM gallery_media WHERE folder_id = $1 AND is_public = TRUE", folderID).Scan(&publicCount))
	assert.Equal(t, 2, publicCount)
}

func TestGalleryCoverRecomputesOnVisibilityAndDelete(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Cover Recompute", nil)
	jpegBytes := tinyGalleryJPEG(t)
	firstID, _ := uploadGalleryMedia(t, ts, folderID, token, "media[]", "first.jpg", "image/jpeg", jpegBytes)
	time.Sleep(10 * time.Millisecond)
	secondID, _ := uploadGalleryMedia(t, ts, folderID, token, "media", "second.jpg", "image/jpeg", jpegBytes)
	assert.Equal(t, secondID, galleryCoverID(t, adminGalleryFolder(t, ts, token, folderID)))

	// Folder-level gate: making the folder public keeps the latest media as
	// its cover — no media need to be individually public.
	updateGalleryFolderVisibility(t, ts, token, folderID, true)
	assert.Equal(t, secondID, galleryCoverID(t, adminGalleryFolder(t, ts, token, folderID)))

	updateGalleryFolderVisibility(t, ts, token, folderID, false)
	assert.Equal(t, secondID, galleryCoverID(t, adminGalleryFolder(t, ts, token, folderID)))
	updateGalleryFolderVisibility(t, ts, token, folderID, true)
	assert.Equal(t, secondID, galleryCoverID(t, adminGalleryFolder(t, ts, token, folderID)))

	deleteGalleryMedia(t, ts, token, secondID)
	assert.Equal(t, firstID, galleryCoverID(t, adminGalleryFolder(t, ts, token, folderID)))
	deleteGalleryMedia(t, ts, token, firstID)
	assert.Empty(t, galleryCoverID(t, adminGalleryFolder(t, ts, token, folderID)))
}

func TestGalleryPrivateBinaryAdminReadAndPublicDeny(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Private Binaries", nil)
	jpegBytes := tinyGalleryJPEG(t)
	mediaID, media := uploadGalleryMedia(t, ts, folderID, token, "media[]", "private.jpg", "image/jpeg", jpegBytes)
	assert.False(t, media["is_public"].(bool))

	resp, err := ts.MakeRequest("GET", "/api/v1/gallery/media/"+mediaID+"/content", nil, token)
	require.NoError(t, err)
	content, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, jpegBytes, content)
	assert.Equal(t, "image/jpeg", resp.Header.Get("Content-Type"))
	assert.Equal(t, "private, max-age=300", resp.Header.Get("Cache-Control"))

	resp, err = ts.MakeRequest("GET", "/api/v1/gallery/media/"+mediaID+"/thumbnail", nil, token)
	require.NoError(t, err)
	thumbnail, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, thumbnail)
	assert.Equal(t, "image/webp", resp.Header.Get("Content-Type"))

	assertGalleryPublicMediaHidden(t, ts, mediaID)
}

func TestGalleryFolderCascadeRemovesRowsAndFiles(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Cascade Files", nil)
	jpegBytes := tinyGalleryJPEG(t)
	firstID, _ := uploadGalleryMedia(t, ts, folderID, token, "media[]", "first.jpg", "image/jpeg", jpegBytes)
	secondID, _ := uploadGalleryMedia(t, ts, folderID, token, "media", "second.jpg", "image/jpeg", jpegBytes)

	paths := make([]string, 0, 4)
	for _, mediaID := range []string{firstID, secondID} {
		var fileURL string
		var thumbnailURL sql.NullString
		require.NoError(t, ts.DB.QueryRow("SELECT file_url, thumbnail_url FROM gallery_media WHERE id = $1", mediaID).Scan(&fileURL, &thumbnailURL))
		require.True(t, thumbnailURL.Valid)
		paths = append(paths,
			filepath.Join("test_uploads", "private", "gallery", "originals", fileURL),
			filepath.Join("test_uploads", "private", "gallery", "thumbnails", thumbnailURL.String),
		)
	}
	for _, path := range paths {
		_, err := os.Stat(path)
		require.NoError(t, err, path)
	}

	resp, err := ts.MakeRequest("DELETE", "/api/v1/gallery/folders/"+folderID, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var folderCount, mediaCount int
	require.NoError(t, ts.DB.QueryRow("SELECT COUNT(*) FROM gallery_folders WHERE id = $1", folderID).Scan(&folderCount))
	require.NoError(t, ts.DB.QueryRow("SELECT COUNT(*) FROM gallery_media WHERE folder_id = $1", folderID).Scan(&mediaCount))
	assert.Zero(t, folderCount)
	assert.Zero(t, mediaCount)
	for _, path := range paths {
		_, err := os.Stat(path)
		require.Error(t, err, path)
		assert.True(t, os.IsNotExist(err), path)
	}

	resp, err = ts.MakeRequest("GET", "/api/v1/gallery/folders/"+folderID, nil, token)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	for _, mediaID := range []string{firstID, secondID} {
		resp, err = ts.MakeRequest("GET", "/api/v1/gallery/media/"+mediaID, nil, token)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	}
}

func TestGalleryPublicURLsDoNotExposePrivateStorage(t *testing.T) {
	ts, token := setupGalleryAdmin(t)
	defer ts.Cleanup(t)

	folderID, _ := createGalleryFolder(t, ts, token, "Private Storage", nil)
	mediaID, media := uploadGalleryMedia(t, ts, folderID, token, "media[]", "storage.jpg", "image/jpeg", tinyGalleryJPEG(t))
	assert.NotContains(t, media["thumbnail_url"].(string), "test_uploads")
	assert.NotContains(t, media["content_url"].(string), "test_uploads")
	assert.True(t, strings.HasPrefix(media["thumbnail_url"].(string), "/api/v1/gallery/"))
	assert.True(t, strings.HasPrefix(media["content_url"].(string), "/api/v1/gallery/"))
	assert.NotEmpty(t, mediaID)
}
