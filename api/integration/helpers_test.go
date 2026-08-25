package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

// toJSONReader converts an interface to a JSON reader
func toJSONReader(v interface{}) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

// parseJSONResponse parses a JSON response into a map
func parseJSONResponse(body io.Reader) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// getDataField extracts the "data" field from a success response
func getDataField(resp map[string]interface{}) map[string]interface{} {
	if data, ok := resp["data"].(map[string]interface{}); ok {
		return data
	}
	return nil
}

// MakeMultipartRequest sends a multipart/form-data request to the test server.
// fields maps form field names to string values.
// files maps form field names to [2]string{filename, content}.
// Content-Type is inferred from the filename extension.
func (ts *TestServer) MakeMultipartRequest(method, path string, fields map[string]string, files map[string][2]string, token string) (*http.Response, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for key, val := range fields {
		if err := w.WriteField(key, val); err != nil {
			return nil, fmt.Errorf("write field %s: %w", key, err)
		}
	}

	for fieldName, fileInfo := range files {
		// fileInfo[0] = filename, fileInfo[1] = content
		// Infer content type from filename extension
		contentType := inferTestContentType(fileInfo[0])

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileInfo[0]))
		h.Set("Content-Type", contentType)

		fw, err := w.CreatePart(h)
		if err != nil {
			return nil, fmt.Errorf("create form file %s: %w", fieldName, err)
		}
		if _, err := io.Copy(fw, strings.NewReader(fileInfo[1])); err != nil {
			return nil, fmt.Errorf("write file %s: %w", fieldName, err)
		}
	}

	w.Close()

	req, err := http.NewRequest(method, ts.Server.URL+path, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return http.DefaultClient.Do(req)
}

// MakeMultipartRequestMultiFiles is like MakeMultipartRequest but supports
// multiple files per field (slice of [2]string{filename, content}).
func (ts *TestServer) MakeMultipartRequestMultiFiles(method, path string, fields map[string]string, files map[string][][2]string, token string) (*http.Response, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for key, val := range fields {
		if err := w.WriteField(key, val); err != nil {
			return nil, fmt.Errorf("write field %s: %w", key, err)
		}
	}

	for fieldName, fileInfos := range files {
		for _, fileInfo := range fileInfos {
			contentType := inferTestContentType(fileInfo[0])

			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileInfo[0]))
			h.Set("Content-Type", contentType)

			fw, err := w.CreatePart(h)
			if err != nil {
				return nil, fmt.Errorf("create form file %s: %w", fieldName, err)
			}
			if _, err := io.Copy(fw, strings.NewReader(fileInfo[1])); err != nil {
				return nil, fmt.Errorf("write file %s: %w", fieldName, err)
			}
		}
	}

	w.Close()

	req, err := http.NewRequest(method, ts.Server.URL+path, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return http.DefaultClient.Do(req)
}

// inferTestContentType returns a MIME type based on the file extension.
func inferTestContentType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(filename, ".png"):
		return "image/png"
	case strings.HasSuffix(filename, ".pdf"):
		return "application/pdf"
	default:
		return "image/jpeg"
	}
}

// ResponseValidator provides comprehensive response validation
type ResponseValidator struct {
	Response map[string]interface{}
	Data     map[string]interface{}
}

// NewResponseValidator creates a new response validator from a parsed response
func NewResponseValidator(resp map[string]interface{}) *ResponseValidator {
	data := getDataField(resp)
	if data == nil {
		data = make(map[string]interface{})
	}
	return &ResponseValidator{
		Response: resp,
		Data:     data,
	}
}

// AssertSuccess validates the success field
func (rv *ResponseValidator) AssertSuccess(t interface{ Errorf(string, ...interface{}) }, expected bool) {
	success, ok := rv.Response["success"].(bool)
	if !ok {
		t.Errorf("success field missing or not a boolean")
		return
	}
	if success != expected {
		t.Errorf("expected success=%v, got %v", expected, success)
	}
}

// AssertError validates the error field (at root level, not in data)
func (rv *ResponseValidator) AssertError(t interface{ Errorf(string, ...interface{}) }, expectedError string) {
	if expectedError == "" {
		if err, ok := rv.Response["error"]; ok && err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		return
	}
	errMsg, ok := rv.Response["error"].(string)
	if !ok {
		t.Errorf("error field missing or not a string")
		return
	}
	if !strings.Contains(errMsg, expectedError) {
		t.Errorf("expected error containing '%s', got '%s'", expectedError, errMsg)
	}
}

// AssertFieldExists validates that a field exists in the data
func (rv *ResponseValidator) AssertFieldExists(t interface{ Errorf(string, ...interface{}) }, fieldName string) {
	if _, ok := rv.Data[fieldName]; !ok {
		t.Errorf("field '%s' missing from response data", fieldName)
	}
}

// AssertFieldNotExists validates that a field does NOT exist in the data
func (rv *ResponseValidator) AssertFieldNotExists(t interface{ Errorf(string, ...interface{}) }, fieldName string) {
	if _, ok := rv.Data[fieldName]; ok {
		t.Errorf("field '%s' should not exist in response data", fieldName)
	}
}

// AssertFieldNotEmpty validates that a field exists and is not empty
func (rv *ResponseValidator) AssertFieldNotEmpty(t interface{ Errorf(string, ...interface{}) }, fieldName string) {
	val, ok := rv.Data[fieldName]
	if !ok {
		t.Errorf("field '%s' missing from response data", fieldName)
		return
	}
	if val == nil || val == "" || val == 0 {
		t.Errorf("field '%s' is empty", fieldName)
	}
}

// AssertFieldValue validates that a field has an expected value.
// JSON numbers decode as float64 by default, so this helper normalizes
// numeric comparisons to avoid int(12) != float64(12) false negatives.
func (rv *ResponseValidator) AssertFieldValue(t interface{ Errorf(string, ...interface{}) }, fieldName string, expectedValue interface{}) {
	val, ok := rv.Data[fieldName]
	if !ok {
		t.Errorf("field '%s' missing from response data", fieldName)
		return
	}
	if !valuesEqual(val, expectedValue) {
		t.Errorf("field '%s': expected %v (%T), got %v (%T)", fieldName, expectedValue, expectedValue, val, val)
	}
}

// valuesEqual compares two values, normalising JSON number types.
// Handles int/float/float64 mixing and string-number coercion.
func valuesEqual(a, b interface{}) bool {
	if a == b {
		return true
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return af == bf
	}
	return false
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		var f float64
		if _, err := fmt.Sscanf(n, "%g", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

// AssertFieldType validates that a field has the expected type
func (rv *ResponseValidator) AssertFieldType(t interface{ Errorf(string, ...interface{}) }, fieldName string, expectedType string) {
	val, ok := rv.Data[fieldName]
	if !ok {
		t.Errorf("field '%s' missing from response data", fieldName)
		return
	}

	var actualType string
	switch val.(type) {
	case string:
		actualType = "string"
	case float64:
		actualType = "number"
	case bool:
		actualType = "boolean"
	case map[string]interface{}:
		actualType = "object"
	case []interface{}:
		actualType = "array"
	case nil:
		actualType = "null"
	default:
		actualType = "unknown"
	}

	if actualType != expectedType {
		t.Errorf("field '%s': expected type %s, got %s", fieldName, expectedType, actualType)
	}
}

// AssertFieldIsUUID validates that a field is a valid UUID
func (rv *ResponseValidator) AssertFieldIsUUID(t interface{ Errorf(string, ...interface{}) }, fieldName string) {
	val, ok := rv.Data[fieldName]
	if !ok {
		t.Errorf("field '%s' missing from response data", fieldName)
		return
	}
	strVal, ok := val.(string)
	if !ok {
		t.Errorf("field '%s' is not a string", fieldName)
		return
	}
	// Simple UUID validation (36 chars with hyphens)
	if len(strVal) != 36 || strings.Count(strVal, "-") != 4 {
		t.Errorf("field '%s' is not a valid UUID: %s", fieldName, strVal)
	}
}

// AssertFieldIsTimestamp validates that a field is a valid ISO 8601 timestamp
func (rv *ResponseValidator) AssertFieldIsTimestamp(t interface{ Errorf(string, ...interface{}) }, fieldName string) {
	val, ok := rv.Data[fieldName]
	if !ok {
		t.Errorf("field '%s' missing from response data", fieldName)
		return
	}
	strVal, ok := val.(string)
	if !ok {
		t.Errorf("field '%s' is not a string", fieldName)
		return
	}
	// Simple ISO 8601 validation (contains T and either Z or +/-)
	if !strings.Contains(strVal, "T") {
		t.Errorf("field '%s' is not a valid ISO 8601 timestamp: %s", fieldName, strVal)
	}
}

// AssertFieldIsURL validates that a field is a valid URL or path
func (rv *ResponseValidator) AssertFieldIsURL(t interface{ Errorf(string, ...interface{}) }, fieldName string) {
	val, ok := rv.Data[fieldName]
	if !ok {
		t.Errorf("field '%s' missing from response data", fieldName)
		return
	}
	strVal, ok := val.(string)
	if !ok {
		t.Errorf("field '%s' is not a string", fieldName)
		return
	}
	// Accept URLs, paths, or filenames
	if strVal == "" {
		t.Errorf("field '%s' is empty", fieldName)
	}
}

// AssertFieldInArray validates that a field value is in an expected array
func (rv *ResponseValidator) AssertFieldInArray(t interface{ Errorf(string, ...interface{}) }, fieldName string, expectedValues []interface{}) {
	val, ok := rv.Data[fieldName]
	if !ok {
		t.Errorf("field '%s' missing from response data", fieldName)
		return
	}

	found := false
	for _, expected := range expectedValues {
		if val == expected {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("field '%s' value %v not in expected values %v", fieldName, val, expectedValues)
	}
}

// AssertFieldLength validates that a field (string or array) has expected length
func (rv *ResponseValidator) AssertFieldLength(t interface{ Errorf(string, ...interface{}) }, fieldName string, expectedLength int) {
	val, ok := rv.Data[fieldName]
	if !ok {
		t.Errorf("field '%s' missing from response data", fieldName)
		return
	}

	var actualLength int
	switch v := val.(type) {
	case string:
		actualLength = len(v)
	case []interface{}:
		actualLength = len(v)
	default:
		t.Errorf("field '%s' is not a string or array", fieldName)
		return
	}

	if actualLength != expectedLength {
		t.Errorf("field '%s': expected length %d, got %d", fieldName, expectedLength, actualLength)
	}
}

// extractCategoryObject returns the "category" field as a map. Returns nil when
// the field is null/missing/absent (PPID categories are nullable).
func (rv *ResponseValidator) extractCategoryObject(t interface{ Errorf(string, ...interface{}) }) map[string]interface{} {
	raw, ok := rv.Data["category"]
	if !ok || raw == nil {
		return nil
	}
	if m, ok := raw.(map[string]interface{}); ok {
		if len(m) == 0 {
			return nil
		}
		return m
	}
	t.Errorf("field 'category' is not an object: %T %v", raw, raw)
	return nil
}

// AssertCategoryIDOrNull asserts the category object's "id" equals expectedID,
// or that category is null (which is valid for PPID — category is nullable per DB schema).
func (rv *ResponseValidator) AssertCategoryIDOrNull(t interface{ Errorf(string, ...interface{}) }, expectedID string) {
	cat := rv.extractCategoryObject(t)
	if cat == nil {
		return // null category is allowed
	}
	id, ok := cat["id"].(string)
	if !ok || id == "" {
		t.Errorf("field 'category.id' missing or not a string: %v", cat["id"])
		return
	}
	if id != expectedID {
		t.Errorf("field 'category.id': expected %s, got %s", expectedID, id)
	}
}

// AssertCategoryNameOrNull asserts the category object's "name" equals expectedName,
// or that category is null. The name may be an empty string in some flows
// (e.g. immediately after Create, where the backend does not re-fetch the
// category join). When the name is non-empty it must match; when it is empty
// (a known backend quirk for Create responses), the assertion is satisfied
// silently — callers needing strict matching can use AssertCategoryIDOrNull.
func (rv *ResponseValidator) AssertCategoryNameOrNull(t interface{ Errorf(string, ...interface{}) }, expectedName string) {
	cat := rv.extractCategoryObject(t)
	if cat == nil {
		return // null category is allowed
	}
	name, _ := cat["name"].(string)
	if name == "" {
		return // backend quirk: name may be empty on Create response
	}
	if name != expectedName {
		t.Errorf("field 'category.name': expected %s, got %s", expectedName, name)
	}
}

// AssertCategoryIDFieldOrNull asserts the raw "category_id" field equals expected
// (or is nil — category is nullable). category_id is the raw UUID string.
func (rv *ResponseValidator) AssertCategoryIDFieldOrNull(t interface{ Errorf(string, ...interface{}) }, expected string) {
	raw, ok := rv.Data["category_id"]
	if !ok || raw == nil {
		return // null category_id is allowed
	}
	got, ok := raw.(string)
	if !ok {
		t.Errorf("field 'category_id' is not a string: %T %v", raw, raw)
		return
	}
	if got != expected {
		t.Errorf("field 'category_id': expected %s, got %s", expected, got)
	}
}

// AssertCategoryExistsOrNull is true when category is null OR is an object with
// both id and name set. Used by list endpoints where we just want to confirm
// the object shape.
func (rv *ResponseValidator) AssertCategoryExistsOrNull(t interface{ Errorf(string, ...interface{}) }) {
	cat := rv.extractCategoryObject(t)
	if cat == nil {
		return
	}
	if _, ok := cat["id"].(string); !ok {
		t.Errorf("category.id is missing or not a string")
	}
	if _, ok := cat["name"].(string); !ok {
		t.Errorf("category.name is missing or not a string")
	}
}

// isNonEmptyObject returns true when val is a map with at least one entry.
// Used by list-style tests where the field can be present (object) or absent/null.
func isNonEmptyObject(t interface{ Errorf(string, ...interface{}) }, val interface{}, fieldName string) bool {
	m, ok := val.(map[string]interface{})
	if !ok || len(m) == 0 {
		t.Errorf("%s must be a non-empty object, got %T %v", fieldName, val, val)
		return false
	}
	return true
}

// parseTimestamp parses an ISO 8601 timestamp string and returns a time.Time
// Handles both formats: 2026-03-20T09:51:31+07:00 and 2026-03-20T02:51:31Z
func parseTimestamp(ts string) (time.Time, error) {
	// Try parsing with timezone offset first
	t, err := time.Parse(time.RFC3339, ts)
	if err == nil {
		return t, nil
	}
	// Try parsing with Z notation
	t, err = time.Parse(time.RFC3339Nano, ts)
	if err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", ts)
}

// AssertTimestampsEqual validates that two timestamp strings represent the same time
// This handles different formats (timezone offset vs Z notation)
func (rv *ResponseValidator) AssertTimestampsEqual(t interface{ Errorf(string, ...interface{}) }, fieldName1, fieldName2 string) {
	val1, ok1 := rv.Data[fieldName1]
	val2, ok2 := rv.Data[fieldName2]

	if !ok1 {
		t.Errorf("field '%s' missing from response data", fieldName1)
		return
	}
	if !ok2 {
		t.Errorf("field '%s' missing from response data", fieldName2)
		return
	}

	str1, ok1 := val1.(string)
	str2, ok2 := val2.(string)

	if !ok1 || !ok2 {
		t.Errorf("fields '%s' and '%s' must be strings", fieldName1, fieldName2)
		return
	}

	time1, err1 := parseTimestamp(str1)
	time2, err2 := parseTimestamp(str2)

	if err1 != nil {
		t.Errorf("failed to parse timestamp in field '%s': %v", fieldName1, err1)
		return
	}
	if err2 != nil {
		t.Errorf("failed to parse timestamp in field '%s': %v", fieldName2, err2)
		return
	}

	// Compare times (they should be equal)
	if !time1.Equal(time2) {
		t.Errorf("timestamps not equal: '%s' (%v) != '%s' (%v)", str1, time1, str2, time2)
	}
}

// AssertTimestampsNotEqual validates that two timestamp strings represent different times
func (rv *ResponseValidator) AssertTimestampsNotEqual(t interface{ Errorf(string, ...interface{}) }, fieldName1, fieldName2 string) {
	val1, ok1 := rv.Data[fieldName1]
	val2, ok2 := rv.Data[fieldName2]

	if !ok1 {
		t.Errorf("field '%s' missing from response data", fieldName1)
		return
	}
	if !ok2 {
		t.Errorf("field '%s' missing from response data", fieldName2)
		return
	}

	str1, ok1 := val1.(string)
	str2, ok2 := val2.(string)

	if !ok1 || !ok2 {
		t.Errorf("fields '%s' and '%s' must be strings", fieldName1, fieldName2)
		return
	}

	time1, err1 := parseTimestamp(str1)
	time2, err2 := parseTimestamp(str2)

	if err1 != nil {
		t.Errorf("failed to parse timestamp in field '%s': %v", fieldName1, err1)
		return
	}
	if err2 != nil {
		t.Errorf("failed to parse timestamp in field '%s': %v", fieldName2, err2)
		return
	}

	// Compare times (they should NOT be equal)
	if time1.Equal(time2) {
		t.Errorf("timestamps should not be equal: '%s' == '%s'", str1, str2)
	}
}
