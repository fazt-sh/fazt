package api

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard JSON response format
type Envelope struct {
	Data  interface{} `json:"data"`
	Meta  interface{} `json:"meta,omitempty"`
	Error *Error      `json:"error,omitempty"`
}

// Error represents an API error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Meta represents pagination metadata
type Meta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// JSON sends a successful JSON response
func JSON(w http.ResponseWriter, status int, data interface{}, meta interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	json.NewEncoder(w).Encode(Envelope{
		Data: data,
		Meta: meta,
		Error: nil,
	})
}

// ErrorResponse sends an error JSON response
func ErrorResponse(w http.ResponseWriter, status int, code, message string, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	json.NewEncoder(w).Encode(Envelope{
		Data: nil,
		Meta: nil,
		Error: &Error{
			Code:    code,
			Message: message,
			Field:   field,
		},
	})
}

// Common Errors

func ServerError(w http.ResponseWriter, err error) {
	ErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", "")
}

func BadRequest(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusBadRequest, "BAD_REQUEST", message, "")
}

func NotFound(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusNotFound, "NOT_FOUND", message, "")
}

func Unauthorized(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", message, "")
}
