package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application-specific error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Cause   error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// HTTPStatus returns the appropriate HTTP status code for this error
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeValidation:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeUpstreamAPI:
		return http.StatusBadGateway
	case ErrCodeDatabase:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Error codes
const (
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeUpstreamAPI  = "UPSTREAM_API_ERROR"
	ErrCodeDatabase     = "DATABASE_ERROR"
	ErrCodeInternal     = "INTERNAL_ERROR"
)

// Predefined errors
var (
	ErrNotFound = &AppError{
		Code:    ErrCodeNotFound,
		Message: "Resource not found",
	}

	ErrValidationFailure = &AppError{
		Code:    ErrCodeValidation,
		Message: "Validation failed",
	}

	ErrUpstreamAPIFailure = &AppError{
		Code:    ErrCodeUpstreamAPI,
		Message: "External API request failed",
	}

	ErrDatabaseFailure = &AppError{
		Code:    ErrCodeDatabase,
		Message: "Database operation failed",
	}
)

// New creates a new AppError
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// WithDetails adds details to an existing AppError
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
		Cause:   e.Cause,
	}
}
