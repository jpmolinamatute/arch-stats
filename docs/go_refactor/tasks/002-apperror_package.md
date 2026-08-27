# Task 002: Create `internal/apperror/` Domain Error Types

## Git Branch

`refactor/002-apperror-package`

## Objective

Create the `internal/apperror/` package with sentinel domain error types and a helper function
for wrapping errors with context. These errors will be used throughout the repository, service,
and handler layers to represent domain-level failures (not found, unauthorized, conflict,
validation) independently of HTTP status codes.

## Dependencies

- Task 001 (Go module scaffold exists)

## Acceptance Criteria

- [ ] `backend/internal/apperror/errors.go` defines the following sentinel errors:
    - `ErrNotFound` — resource does not exist
    - `ErrUnauthorized` — authentication required or failed
    - `ErrForbidden` — authenticated but insufficient permissions
    - `ErrConflict` — duplicate resource or state conflict
    - `ErrValidation` — input validation failure
- [ ] Each sentinel error is a distinct type implementing the `error` interface with a
  `Code() string` method for programmatic identification.
- [ ] An `Is(err, target)` check works correctly with `errors.Is()` for each sentinel.
- [ ] A `Wrap(sentinel, msg string) error` helper exists that wraps a sentinel with context
  while preserving `errors.Is()` behavior.
- [ ] Unit tests in `backend/internal/apperror/errors_test.go` cover:
    - Each sentinel's `Error()` message
    - Each sentinel's `Code()` value
    - `errors.Is()` matching for wrapped errors
    - `errors.As()` extracting the typed error
- [ ] `go test ./internal/apperror/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/apperror/errors.go` |
| Create | `backend/internal/apperror/errors_test.go` |
| Delete | `backend/internal/apperror/.gitkeep` |

## Steps

- [ ] **Step 1: Write the failing tests**

  Create `backend/internal/apperror/errors_test.go`:

  ```go
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
  ```

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/apperror/...
  ```

  Expected: compilation errors (types don't exist yet).

- [ ] **Step 3: Implement the `apperror` package**

  Create `backend/internal/apperror/errors.go`:

  ```go
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
  ```

- [ ] **Step 4: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/apperror/... -v
  ```

  Expected: all tests PASS.

- [ ] **Step 5: Run go vet**

  ```bash
  cd backend
  go vet ./...
  ```

  Expected: no issues.

- [ ] **Step 6: Remove .gitkeep and commit**

  ```bash
  rm -f backend/internal/apperror/.gitkeep
  git add -A
  git commit -m "feat: add internal/apperror domain error types with tests"
  ```

## Verification

- `cd backend && go test ./internal/apperror/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles without errors.
