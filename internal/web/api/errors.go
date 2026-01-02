package api

import (
	"encoding/json"
	"net/http"
)

// Common error codes
const (
	ErrCodeNotFound     = "not_found"
	ErrCodeBadRequest   = "bad_request"
	ErrCodeInternal     = "internal_error"
	ErrCodeUnauthorized = "unauthorized"
)

// WriteError writes an error response with the given status code.
func WriteError(w http.ResponseWriter, code int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: message,
		},
	})
}

// WriteJSON writes a successful JSON response.
func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    data,
	})
}

// WritePaginated writes a paginated response.
func WritePaginated(w http.ResponseWriter, items interface{}, total, offset, limit int) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data: PaginatedResponse{
			Items:   items,
			Total:   total,
			Offset:  offset,
			Limit:   limit,
			HasMore: offset+limit < total,
		},
	})
}

// NotFound writes a 404 response.
func NotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, ErrCodeNotFound, message)
}

// BadRequest writes a 400 response.
func BadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
}

// InternalError writes a 500 response.
func InternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, ErrCodeInternal, message)
}
