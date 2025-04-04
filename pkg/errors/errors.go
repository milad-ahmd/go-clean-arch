package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput is returned when the input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when the user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user is forbidden from accessing a resource
	ErrForbidden = errors.New("forbidden")

	// ErrConflict is returned when there is a conflict
	ErrConflict = errors.New("conflict")

	// ErrInternal is returned when there is an internal server error
	ErrInternal = errors.New("internal server error")
)

// AppError represents an application error
type AppError struct {
	Err        error
	Message    string
	StatusCode int
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors"`
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string, errors map[string]interface{}) *ValidationError {
	return &ValidationError{
		Message: message,
		Errors:  errors,
	}
}

// NewAppError creates a new application error
func NewAppError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(entity string, id interface{}) *AppError {
	return &AppError{
		Err:        ErrNotFound,
		Message:    fmt.Sprintf("%s with ID %v not found", entity, id),
		StatusCode: http.StatusNotFound,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(entity, field string, value interface{}) *AppError {
	return &AppError{
		Err:        ErrConflict,
		Message:    fmt.Sprintf("%s with %s %v already exists", entity, field, value),
		StatusCode: http.StatusConflict,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "unauthorized access"
	}
	return &AppError{
		Err:        ErrUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "access forbidden"
	}
	return &AppError{
		Err:        ErrForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(err error) *AppError {
	return &AppError{
		Err:        err,
		Message:    "internal server error",
		StatusCode: http.StatusInternalServerError,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Err:        ErrInvalidInput,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// GetStatusCode returns the HTTP status code for an error
func GetStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}

	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return http.StatusBadRequest
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
