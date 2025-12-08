package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// SuccessResponse represents the standard success envelope
type SuccessResponse struct {
	Data interface{}            `json:"data"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// ErrorResponse represents the standard error envelope
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// CheckSuccess verifies a successful response and returns the parsed data
func CheckSuccess(t *testing.T, resp *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	t.Helper()

	if resp.Code != expectedStatus {
		t.Fatalf("Expected status %d, got %d. Body: %s", expectedStatus, resp.Code, resp.Body.String())
	}

	// Verify Content-Type
	contentType := resp.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var result SuccessResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v. Body: %s", err, resp.Body.String())
	}

	if result.Data == nil {
		t.Fatal("Expected 'data' field in success response, got nil")
	}

	// Convert to map for easier assertion
	dataJSON, _ := json.Marshal(result.Data)
	var dataMap map[string]interface{}
	json.Unmarshal(dataJSON, &dataMap)

	return dataMap
}

// CheckSuccessArray verifies a successful response with array data
func CheckSuccessArray(t *testing.T, resp *httptest.ResponseRecorder, expectedStatus int) []interface{} {
	t.Helper()

	if resp.Code != expectedStatus {
		t.Fatalf("Expected status %d, got %d. Body: %s", expectedStatus, resp.Code, resp.Body.String())
	}

	var result struct {
		Data []interface{}          `json:"data"`
		Meta map[string]interface{} `json:"meta,omitempty"`
	}

	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v. Body: %s", err, resp.Body.String())
	}

	if result.Data == nil {
		t.Fatal("Expected 'data' array in success response, got nil")
	}

	return result.Data
}

// CheckError verifies an error response and checks the error code
func CheckError(t *testing.T, resp *httptest.ResponseRecorder, expectedStatus int, expectedCode string) ErrorDetail {
	t.Helper()

	if resp.Code != expectedStatus {
		t.Fatalf("Expected status %d, got %d. Body: %s", expectedStatus, resp.Code, resp.Body.String())
	}

	// Verify Content-Type
	contentType := resp.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var result ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON error response: %v. Body: %s", err, resp.Body.String())
	}

	if result.Error.Code == "" {
		t.Fatal("Expected 'error.code' field in error response")
	}

	if result.Error.Code != expectedCode {
		t.Errorf("Expected error code '%s', got '%s'. Message: %s",
			expectedCode, result.Error.Code, result.Error.Message)
	}

	if result.Error.Message == "" {
		t.Error("Expected 'error.message' field to be non-empty")
	}

	return result.Error
}

// WithAuth adds Bearer token authentication to a request
func WithAuth(req *http.Request, token string) *http.Request {
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// WithSession adds session cookie to a request
func WithSession(req *http.Request, sessionID string) *http.Request {
	req.AddCookie(&http.Cookie{
		Name:  "cc_session", // Must match auth.SessionCookieName
		Value: sessionID,
	})
	return req
}

// JSONRequest creates a new HTTP request with JSON body
func JSONRequest(method, path string, body interface{}) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// AssertFieldEquals checks if a field in the data map equals expected value
func AssertFieldEquals(t *testing.T, data map[string]interface{}, field string, expected interface{}) {
	t.Helper()

	actual, exists := data[field]
	if !exists {
		t.Errorf("Expected field '%s' to exist in response data", field)
		return
	}

	if actual != expected {
		t.Errorf("Expected field '%s' to be '%v', got '%v'", field, expected, actual)
	}
}

// AssertFieldExists checks if a field exists in the data map
func AssertFieldExists(t *testing.T, data map[string]interface{}, field string) {
	t.Helper()

	if _, exists := data[field]; !exists {
		t.Errorf("Expected field '%s' to exist in response data", field)
	}
}
