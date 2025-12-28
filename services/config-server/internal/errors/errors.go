package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors - Resource lifecycle
var (
	// ErrNotFound is returned when a requested resource does not exist
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when attempting to create a resource that already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInUse is returned when attempting to delete a resource that is still referenced
	ErrInUse = errors.New("resource is in use and cannot be deleted")
)

// Sentinel errors - Validation and input
var (
	// ErrInvalidInput is returned for generic validation failures
	ErrInvalidInput = errors.New("invalid input")
)

// Sentinel errors - Relationships and integrity
var (
	// ErrForeignKeyViolation is returned when a referenced resource does not exist
	ErrForeignKeyViolation = errors.New("referenced resource does not exist")

	// ErrCircularReference is returned when a circular dependency is detected
	ErrCircularReference = errors.New("circular reference detected")

	// ErrCannotRemoveLastGroup is returned when attempting to remove the last group from a target
	ErrCannotRemoveLastGroup = errors.New("cannot remove last group from target")
)

// Sentinel errors - Bootstrap tokens
var (
	// ErrTokenExpired is returned when a bootstrap token has expired
	ErrTokenExpired = errors.New("token expired")

	// ErrTokenExhausted is returned when a bootstrap token has reached its usage limit
	ErrTokenExhausted = errors.New("token usage limit exceeded")

	// ErrInvalidToken is returned when a bootstrap token is invalid or not found
	ErrInvalidToken = errors.New("invalid token")
)

// Sentinel errors - Database operations
var (
	// ErrDuplicateKey is returned when a UNIQUE constraint is violated
	ErrDuplicateKey = errors.New("duplicate key violation")

	// ErrConstraintViolation is returned for generic constraint violations
	ErrConstraintViolation = errors.New("constraint violation")

	// ErrDatabaseError is returned for unexpected database errors
	ErrDatabaseError = errors.New("database error")
)

// Sentinel errors - Handler layer
var (
	// ErrBindingFailed is returned when request binding fails
	ErrBindingFailed = errors.New("request binding failed")
)

// ValidationError represents a field-level validation error
type ValidationError struct {
	Field   string
	Message string
	Cause   error
}

// Error implements the error interface for ValidationError
func (e *ValidationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Field, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Unwrap returns the wrapped error
func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// BindingError represents a request binding/parsing error
type BindingError struct {
	Message string
	Cause   error
}

// Error implements the error interface for BindingError
func (e *BindingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("binding error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("binding error: %s", e.Message)
}

// Unwrap returns the wrapped error
func (e *BindingError) Unwrap() error {
	return e.Cause
}

// NewBindingError creates a new binding error
func NewBindingError(err error) error {
	if err == nil {
		return nil
	}
	return &BindingError{
		Message: "failed to bind request",
		Cause:   err,
	}
}

// DatabaseError represents a database-specific error with details
type DatabaseError struct {
	Operation string
	Table     string
	Cause     error
}

// Error implements the error interface for DatabaseError
func (e *DatabaseError) Error() string {
	if e.Table != "" {
		return fmt.Sprintf("database error during %s on %s: %v", e.Operation, e.Table, e.Cause)
	}
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Cause)
}

// Unwrap returns the wrapped error
func (e *DatabaseError) Unwrap() error {
	return e.Cause
}

// Error type checking functions

// IsNotFound checks if an error is ErrNotFound
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if an error is ErrAlreadyExists
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsDuplicateKey checks if an error is ErrDuplicateKey
func IsDuplicateKey(err error) bool {
	return errors.Is(err, ErrDuplicateKey)
}

// IsConstraintViolation checks if an error is a constraint violation
func IsConstraintViolation(err error) bool {
	return errors.Is(err, ErrConstraintViolation) ||
		errors.Is(err, ErrDuplicateKey) ||
		errors.Is(err, ErrForeignKeyViolation)
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsBindingError checks if an error is a BindingError
func IsBindingError(err error) bool {
	var bindingErr *BindingError
	return errors.As(err, &bindingErr)
}

// IsForeignKeyViolation checks if an error is ErrForeignKeyViolation
func IsForeignKeyViolation(err error) bool {
	return errors.Is(err, ErrForeignKeyViolation)
}

// IsCircularReference checks if an error is ErrCircularReference
func IsCircularReference(err error) bool {
	return errors.Is(err, ErrCircularReference)
}

// IsInUse checks if an error is ErrInUse
func IsInUse(err error) bool {
	return errors.Is(err, ErrInUse)
}

// Error wrapping functions

// Wrap wraps an error with a message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}
