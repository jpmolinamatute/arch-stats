// Package apperror defines domain error types used across all layers.
//
// These errors are independent of HTTP — the middleware layer maps them
// to appropriate HTTP status codes.
package apperror

import "fmt"

// AppError represents a domain-level error with a machine-readable code.
type AppError struct {
	code    string
	message string
}

func (e *AppError) Error() string {
	return e.message
}

// Code returns the machine-readable error code (e.g., "NOT_FOUND").
func (e *AppError) Code() string {
	return e.code
}

// Is supports errors.Is() matching by comparing error codes.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.code == t.code
}

// Sentinel domain errors.
var (
	ErrNotFound     = &AppError{code: "NOT_FOUND", message: "not found"}
	ErrUnauthorized = &AppError{code: "UNAUTHORIZED", message: "unauthorized"}
	ErrForbidden    = &AppError{code: "FORBIDDEN", message: "forbidden"}
	ErrConflict     = &AppError{code: "CONFLICT", message: "conflict"}
	ErrValidation   = &AppError{code: "VALIDATION", message: "validation"}
)

// Wrap creates a new AppError with the same code as the sentinel but a
// contextual message. The returned error matches the sentinel via errors.Is().
func Wrap(sentinel *AppError, msg string) *AppError {
	return &AppError{
		code:    sentinel.code,
		message: fmt.Sprintf("%s: %s", sentinel.message, msg),
	}
}
