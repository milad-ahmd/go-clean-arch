package domain

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrInternalServer = errors.New("internal server error")
	ErrConflict       = errors.New("conflict")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
)

// NotFoundError represents a not found error
type NotFoundError struct {
	Entity string
	ID     interface{}
}

// Error returns the error message
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %v not found", e.Entity, e.ID)
}

// Is checks if the error is of the given type
func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// ConflictError represents a conflict error
type ConflictError struct {
	Entity string
	Field  string
	Value  interface{}
}

// Error returns the error message
func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s with %s %v already exists", e.Entity, e.Field, e.Value)
}

// Is checks if the error is of the given type
func (e *ConflictError) Is(target error) bool {
	return target == ErrConflict
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s %s", e.Field, e.Message)
}

// Is checks if the error is of the given type
func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}
