package service

import (
	"errors"
	"fmt"
)

// Common service errors
var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrForeignKeyViolation = errors.New("referenced resource does not exist")
	ErrCircularReference = errors.New("circular reference detected")
	ErrInUse             = errors.New("resource is in use and cannot be deleted")
)

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
