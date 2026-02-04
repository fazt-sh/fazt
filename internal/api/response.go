package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// SuccessEnvelope represents a successful API response
// Success responses ONLY contain data (and optional meta), never error fields
type SuccessEnvelope struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta,omitempty"`
}

// ErrorEnvelope represents an error API response
// Error responses ONLY contain error details, never data fields
type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents structured error information
type ErrorDetail struct {
	Code    string                 `json:"code"`              // Machine-readable error code (UPPERCASE_SNAKE_CASE)
	Message string                 `json:"message"`           // Human-readable error message
	Details map[string]interface{} `json:"details,omitempty"` // Optional field-level details
}

// Success writes a successful JSON response with data
// Use this for single resources or when no pagination is needed
//
// Example:
//
//	api.Success(w, http.StatusOK, map[string]string{"username": "admin"})
//	// Returns: {"data": {"username": "admin"}}
func Success(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(SuccessEnvelope{Data: data}); err != nil {
		log.Printf("Failed to encode success response: %v", err)
	}
}

// SuccessWithMeta writes a successful JSON response with data and metadata
// Use this for paginated lists or when you need to include additional context
//
// Example:
//
//	meta := map[string]interface{}{"total": 100, "limit": 20, "offset": 0}
//	api.SuccessWithMeta(w, http.StatusOK, events, meta)
//	// Returns: {"data": [...], "meta": {"total": 100, "limit": 20, "offset": 0}}
func SuccessWithMeta(w http.ResponseWriter, status int, data interface{}, meta interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(SuccessEnvelope{Data: data, Meta: meta}); err != nil {
		log.Printf("Failed to encode success response with meta: %v", err)
	}
}

// Error writes an error JSON response with structured error details
// Use this for custom error codes with optional details
//
// Example:
//
//	details := map[string]interface{}{"field": "site_name", "constraint": "required"}
//	api.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Site name is required", details)
func Error(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(ErrorEnvelope{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// Common Error Response Helpers
// These provide shortcuts for standard HTTP error codes

// BadRequest returns a 400 Bad Request error
// Use for malformed requests, invalid JSON, or generic client errors
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message, nil)
}

// ValidationError returns a 400 Bad Request with field-level validation details
// Use for form validation errors where specific fields are invalid
//
// Example:
//
//	api.ValidationError(w, "Site name must be lowercase", "site_name", "pattern")
func ValidationError(w http.ResponseWriter, message, field, constraint string) {
	Error(w, http.StatusBadRequest, "VALIDATION_FAILED", message, map[string]interface{}{
		"field":      field,
		"constraint": constraint,
	})
}

// InvalidJSON returns a 400 Bad Request for JSON parsing errors
func InvalidJSON(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "INVALID_JSON", message, nil)
}

// MissingField returns a 400 Bad Request for missing required fields
func MissingField(w http.ResponseWriter, field string) {
	Error(w, http.StatusBadRequest, "MISSING_FIELD",
		"Required field is missing: "+field,
		map[string]interface{}{"field": field})
}

// Unauthorized returns a 401 Unauthorized error
// Use when authentication is required but missing or invalid
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// InvalidCredentials returns a 401 Unauthorized for login failures
func InvalidCredentials(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid username or password", nil)
}

// SessionExpired returns a 401 Unauthorized for expired sessions
func SessionExpired(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "SESSION_EXPIRED", "Your session has expired. Please log in again.", nil)
}

// InvalidAPIKey returns a 401 Unauthorized for API key authentication failures
func InvalidAPIKey(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid or expired API key", nil)
}

// Forbidden returns a 403 Forbidden error
// Use when user is authenticated but lacks permission
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message, nil)
}

// NotFound returns a 404 Not Found error with a custom error code
// Use with specific codes like SITE_NOT_FOUND, REDIRECT_NOT_FOUND
//
// Example:
//
//	api.NotFound(w, "SITE_NOT_FOUND", "Site 'blog' does not exist")
func NotFound(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusNotFound, code, message, nil)
}

// ResourceNotFound returns a generic 404 Not Found error
func ResourceNotFound(w http.ResponseWriter, resourceType, resourceID string) {
	Error(w, http.StatusNotFound, "NOT_FOUND",
		resourceType+" '"+resourceID+"' not found", nil)
}

// Conflict returns a 409 Conflict error
// Use when a resource already exists (e.g., duplicate site name)
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message, nil)
}

// PayloadTooLarge returns a 413 Payload Too Large error
// Use when uploads exceed size limits
func PayloadTooLarge(w http.ResponseWriter, maxSize string) {
	Error(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE",
		"Request payload exceeds maximum size of "+maxSize, nil)
}

// RateLimitExceeded returns a 429 Too Many Requests error
// Use when rate limits are exceeded
func RateLimitExceeded(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message, nil)
}

// InternalError returns a 500 Internal Server Error
// Use for unexpected server errors, database failures, panics
// The actual error is logged but not exposed to the client for security
func InternalError(w http.ResponseWriter, err error) {
	// Log the actual error for debugging
	if err != nil {
		log.Printf("Internal error: %v", err)
	}

	// Return generic message to client (don't leak internal details)
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR",
		"An unexpected error occurred. Please try again later.", nil)
}

// ServiceUnavailable returns a 503 Service Unavailable error
// Use when database is locked, maintenance mode, etc.
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, nil)
}

// ============================================================================
// Pagination Helpers
// ============================================================================

// Pagination represents pagination parameters and metadata
type Pagination struct {
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// ParsePagination extracts pagination parameters from request query string
// Defaults: offset=0, limit=20, max limit=100
func ParsePagination(r *http.Request) (offset, limit int) {
	offset = 0
	limit = 20

	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := parseInt(v); err == nil && n >= 0 {
			offset = n
		}
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			limit = n
			if limit > 100 {
				limit = 100
			}
		}
	}

	return offset, limit
}

// parseInt is a simple int parser helper
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// PaginatedSuccess writes a paginated response with data and pagination metadata
func PaginatedSuccess(w http.ResponseWriter, status int, data interface{}, offset, limit, total int) {
	meta := Pagination{
		Offset:  offset,
		Limit:   limit,
		Total:   total,
		HasMore: offset+limit < total,
	}
	SuccessWithMeta(w, status, data, meta)
}

// ============================================================================
// DEPRECATED: Backward compatibility aliases for gradual migration
// These will be removed once all handlers are migrated
// ============================================================================

// JSON is deprecated. Use Success() or SuccessWithMeta() instead.
func JSON(w http.ResponseWriter, status int, data interface{}, meta interface{}) {
	if meta != nil {
		SuccessWithMeta(w, status, data, meta)
	} else {
		Success(w, status, data)
	}
}

// ErrorResponse is deprecated. Use specific error helpers (BadRequest, NotFound, etc.) instead.
func ErrorResponse(w http.ResponseWriter, status int, code, message string, field string) {
	details := make(map[string]interface{})
	if field != "" {
		details["field"] = field
	}
	Error(w, status, code, message, details)
}

// ServerError is deprecated. Use InternalError() instead.
func ServerError(w http.ResponseWriter, err error) {
	InternalError(w, err)
}
