package models

import "time"

// APIResponse provides a consistent response structure
type APIResponse[T any] struct {
	Success bool      `json:"success"`
	Data    T         `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
	Meta    *Meta     `json:"meta,omitempty"`
}

// APIError provides structured error information
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta provides response metadata
type Meta struct {
	Timestamp  time.Time          `json:"timestamp"`
	Pagination *PaginationMeta    `json:"pagination,omitempty"`
	Filters    map[string]string  `json:"filters,omitempty"`
	RequestID  string             `json:"request_id,omitempty"`
}

// PaginationMeta provides pagination information
type PaginationMeta struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Total      int    `json:"total"`
	TotalPages int    `json:"total_pages"`
	HasNext    bool   `json:"has_next"`
	HasPrev    bool   `json:"has_prev"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// Common error codes
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeConflict      = "CONFLICT"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeRateLimit     = "RATE_LIMIT_EXCEEDED"
	ErrCodeBadRequest    = "BAD_REQUEST"
)

// Helper functions for creating responses
func NewSuccessResponse[T any](data T) APIResponse[T] {
	return APIResponse[T]{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Timestamp: time.Now(),
		},
	}
}

func NewErrorResponse(code, message, details string) APIResponse[any] {
	return APIResponse[any]{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: &Meta{
			Timestamp: time.Now(),
		},
	}
}

func NewPaginatedResponse[T any](data T, pagination PaginationMeta) APIResponse[T] {
	return APIResponse[T]{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Timestamp:  time.Now(),
			Pagination: &pagination,
		},
	}
}