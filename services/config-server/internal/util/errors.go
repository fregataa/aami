package util

import (
	"errors"
	"fmt"
)

// Common application errors
var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrCircularReference = errors.New("circular reference detected")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenExhausted    = errors.New("token usage limit exceeded")
	ErrInvalidToken      = errors.New("invalid token")
	ErrDatabaseError     = errors.New("database error")
)

// AppError represents an application error with additional context
type AppError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NotFoundError creates a not found error
func NotFoundError(resource string) *AppError {
	return NewAppError("NOT_FOUND", fmt.Sprintf("%s not found", resource), ErrNotFound)
}

// AlreadyExistsError creates an already exists error
func AlreadyExistsError(resource string) *AppError {
	return NewAppError("ALREADY_EXISTS", fmt.Sprintf("%s already exists", resource), ErrAlreadyExists)
}

// ValidationError creates a validation error
func ValidationError(message string) *AppError {
	return NewAppError("VALIDATION_ERROR", message, ErrInvalidInput)
}
