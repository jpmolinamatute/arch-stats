package apperror_test

import (
	"errors"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *apperror.AppError
		code     string
		contains string
	}{
		{"ErrNotFound", apperror.ErrNotFound, "NOT_FOUND", "not found"},
		{"ErrUnauthorized", apperror.ErrUnauthorized, "UNAUTHORIZED", "unauthorized"},
		{"ErrForbidden", apperror.ErrForbidden, "FORBIDDEN", "forbidden"},
		{"ErrConflict", apperror.ErrConflict, "CONFLICT", "conflict"},
		{"ErrValidation", apperror.ErrValidation, "VALIDATION", "validation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code() != tt.code {
				t.Errorf("Code() = %q, want %q", tt.err.Code(), tt.code)
			}
			if tt.err.Error() == "" {
				t.Error("Error() returned empty string")
			}
		})
	}
}

func TestWrap(t *testing.T) {
	wrapped := apperror.Wrap(apperror.ErrNotFound, "fetching archer by ID")

	if !errors.Is(wrapped, apperror.ErrNotFound) {
		t.Error("errors.Is() should match ErrNotFound")
	}

	var appErr *apperror.AppError
	if !errors.As(wrapped, &appErr) {
		t.Fatal("errors.As() should extract AppError")
	}
	if appErr.Code() != "NOT_FOUND" {
		t.Errorf("Code() = %q, want NOT_FOUND", appErr.Code())
	}
}

func TestWrapMessage(t *testing.T) {
	wrapped := apperror.Wrap(apperror.ErrValidation, "email is required")
	msg := wrapped.Error()

	if msg != "validation: email is required" {
		t.Errorf("Error() = %q, want %q", msg, "validation: email is required")
	}
}
