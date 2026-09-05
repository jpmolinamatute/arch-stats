# Task 018: Build HTTP Middleware Stack Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the HTTP middleware stack in `backend/internal/middleware/` providing authentication, error-to-HTTP mapping, structured request logging with slog, CORS handling for dev/prod modes, and panic recovery, with helpers for request context propagation.

**Architecture:**
- `context.go`: Type-safe request context helpers for storing and retrieving the authenticated archer UUID.
- `error_mapper.go`: Centralized domain error mapper translating `apperror.*` sentinels (`ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrValidation`) to standard HTTP status codes (404, 401, 403, 409, 422, 500) and formatting JSON responses (`{"detail": "...", "code": "..."}`).
- `auth.go`: HTTP authentication middleware extracting JWT tokens from either the `Authorization: Bearer <token>` header or `arch_stats_auth` cookie, delegating validation to `auth.TokenAuthenticator`, and injecting the archer ID into the request context.
- `logging.go`: Structured request logging middleware using standard library `log/slog` recording method, URL path, response status, client IP, user agent, and request duration.
- `cors.go`: CORS middleware configuring `Access-Control-*` headers and handling preflight `OPTIONS` requests based on development (`DevMode=true`) vs production (`DevMode=false`) runtime configuration.
- `recovery.go`: Panic recovery middleware catching unhandled runtime panics, logging stack traces, and serving standardized 500 Internal Server Error JSON responses.

**Tech Stack:** Go 1.27+, Chi-compatible HTTP middleware signatures (`func(http.Handler) http.Handler`), standard library `net/http`, `log/slog`, `runtime/debug`, `crypto/subtle`, `github.com/google/uuid`, internal packages (`apperror`, `auth`, `model`).

**Spec:** [docs/go_refactor/tasks/018-middleware_stack.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/018-middleware_stack.md)

## Global Constraints

- Git branch: `refactor/018-middleware-stack`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/middleware`
- Error handling: Wrap internal errors with `%w` using contextual descriptive messages. Return domain sentinels `apperror.ErrNotFound`, `apperror.ErrUnauthorized`, `apperror.ErrForbidden`, `apperror.ErrConflict`, and `apperror.Wrap(apperror.ErrValidation, ...)`.
- Interface design: Middleware functions must adhere to the standard Chi/HTTP signature `func(http.Handler) http.Handler`. `Auth` middleware must accept an interface `TokenAuthenticator` defined in `middleware` so that `*auth.Service` satisfies it implicitly without tight circular coupling.
- JSON error formatting: Error responses must output JSON with a `"detail"` string field matching frontend expectations (`frontend/src/api/client.ts`) and an optional `"code"` string field.
- Formatting must adhere to `gofumpt` and linting must pass `golangci-lint run ./...`.
- `go test -race ./internal/middleware/... ./internal/auth/... -v` must pass.
- `go vet ./...` must report no issues.
- `go build ./...` must compile cleanly.

---

## File Structure

```
backend/
├── internal/
│   ├── auth/
│   │   ├── service.go            # [MODIFY] Add Authenticate(ctx, tokenStr) method to Service
│   │   └── service_test.go       # [MODIFY] Add unit tests for Authenticate
│   └── middleware/
│       ├── .gitkeep              # [DELETE] Remove placeholder once files are added
│       ├── context.go            # [NEW] WithArcherID and GetArcherID context helpers
│       ├── context_test.go       # [NEW] Unit tests for context helpers
│       ├── error_mapper.go       # [NEW] MapError, WriteError, ErrorResponse, and ErrorMapper middleware
│       ├── error_mapper_test.go  # [NEW] Unit tests for sentinel status mapping and JSON output
│       ├── auth.go               # [NEW] TokenAuthenticator interface and Auth middleware
│       ├── auth_test.go          # [NEW] Unit tests for cookie/header extraction and 401 handling
│       ├── logging.go            # [NEW] RequestLogger middleware and statusRecorder
│       ├── logging_test.go       # [NEW] Unit tests for slog request logging
│       ├── cors.go               # [NEW] CORS middleware and CORSOptions for dev/prod modes
│       ├── cors_test.go          # [NEW] Unit tests for CORS headers and OPTIONS preflights
│       ├── recovery.go           # [NEW] Recovery middleware for catching panics
│       └── recovery_test.go      # [NEW] Unit tests for panic recovery and 500 response
```

---

### Task 1: Git Branch Setup & Context Helpers

**Files:**
- Create: `backend/internal/middleware/context.go`
- Create: `backend/internal/middleware/context_test.go`

**Interfaces:**
- Consumes: `github.com/google/uuid`, `github.com/jpmolinamatute/arch-stats/backend/internal/apperror`
- Produces:
  - `WithArcherID(ctx context.Context, archerID uuid.UUID) context.Context`
  - `GetArcherID(ctx context.Context) (uuid.UUID, error)`

- [x] **Step 1: Create git branch**

```bash
git switch -c refactor/018-middleware-stack
```

- [x] **Step 2: Write failing tests for context helpers**

Create `backend/internal/middleware/context_test.go`:

```go
package middleware_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestContext_ArcherIDRoundTrip(t *testing.T) {
	expectedID := uuid.New()
	ctx := context.Background()

	ctxWithArcher := middleware.WithArcherID(ctx, expectedID)

	gotID, err := middleware.GetArcherID(ctxWithArcher)
	if err != nil {
		t.Fatalf("GetArcherID() returned unexpected error: %v", err)
	}
	if gotID != expectedID {
		t.Errorf("GetArcherID() = %v, want %v", gotID, expectedID)
	}
}

func TestContext_GetArcherID_EmptyContextReturnsError(t *testing.T) {
	ctx := context.Background()

	gotID, err := middleware.GetArcherID(ctx)
	if err == nil {
		t.Fatalf("GetArcherID(emptyCtx) expected error, got nil (id=%v)", gotID)
	}
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Errorf("GetArcherID(emptyCtx) error = %v, want errors.Is(err, apperror.ErrUnauthorized)", err)
	}
	if gotID != uuid.Nil {
		t.Errorf("GetArcherID(emptyCtx) id = %v, want uuid.Nil", gotID)
	}
}

func TestContext_GetArcherID_NilUUIDReturnsError(t *testing.T) {
	ctx := middleware.WithArcherID(context.Background(), uuid.Nil)

	gotID, err := middleware.GetArcherID(ctx)
	if err == nil {
		t.Fatalf("GetArcherID(nilUUID) expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Errorf("GetArcherID(nilUUID) error = %v, want errors.Is(err, apperror.ErrUnauthorized)", err)
	}
	if gotID != uuid.Nil {
		t.Errorf("GetArcherID(nilUUID) id = %v, want uuid.Nil", gotID)
	}
}
```

- [x] **Step 3: Run test to verify it fails**

```bash
cd backend && go test ./internal/middleware/... -v
```
Expected: Compilation failure because `middleware.WithArcherID` and `middleware.GetArcherID` are not defined.

- [x] **Step 4: Write minimal context implementation**

Create `backend/internal/middleware/context.go`:

```go
package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

type contextKey string

const archerIDContextKey contextKey = "arch_stats_archer_id"

// WithArcherID injects the authenticated archer UUID into the request context.
func WithArcherID(ctx context.Context, archerID uuid.UUID) context.Context {
	return context.WithValue(ctx, archerIDContextKey, archerID)
}

// GetArcherID extracts the authenticated archer UUID from the request context.
// Returns apperror.ErrUnauthorized if the ID is missing or is uuid.Nil.
func GetArcherID(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(archerIDContextKey)
	if val == nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "archer id not found in context")
	}

	id, ok := val.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid archer id in context")
	}

	return id, nil
}
```

- [x] **Step 5: Run tests to verify they pass**

```bash
cd backend && go test ./internal/middleware/... -v
```
Expected: PASS.

- [x] **Step 6: Commit**

```bash
git add backend/internal/middleware/context.go backend/internal/middleware/context_test.go
git commit -m "feat(middleware): add archer ID context helpers and unit tests"
```

---

### Task 2: Error Mapper

**Files:**
- Create: `backend/internal/middleware/error_mapper.go`
- Create: `backend/internal/middleware/error_mapper_test.go`

**Interfaces:**
- Consumes: `github.com/jpmolinamatute/arch-stats/backend/internal/apperror`
- Produces:
  - `type ErrorResponse struct { Detail string; Code string }`
  - `MapError(err error) (int, ErrorResponse)`
  - `WriteError(w http.ResponseWriter, err error)`
  - `ErrorMapper(next http.Handler) http.Handler`

- [x] **Step 1: Write failing tests for error mapper**

Create `backend/internal/middleware/error_mapper_test.go`:

```go
package middleware_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantStatus   int
		wantCode     string
		wantDetailSub string
	}{
		{
			name:         "ErrNotFound",
			err:          apperror.ErrNotFound,
			wantStatus:   http.StatusNotFound,
			wantCode:     "NOT_FOUND",
			wantDetailSub: "not found",
		},
		{
			name:         "Wrapped ErrNotFound",
			err:          apperror.Wrap(apperror.ErrNotFound, "archer 123"),
			wantStatus:   http.StatusNotFound,
			wantCode:     "NOT_FOUND",
			wantDetailSub: "archer 123",
		},
		{
			name:         "ErrUnauthorized",
			err:          apperror.ErrUnauthorized,
			wantStatus:   http.StatusUnauthorized,
			wantCode:     "UNAUTHORIZED",
			wantDetailSub: "unauthorized",
		},
		{
			name:         "Wrapped ErrUnauthorized",
			err:          apperror.Wrap(apperror.ErrUnauthorized, "session expired"),
			wantStatus:   http.StatusUnauthorized,
			wantCode:     "UNAUTHORIZED",
			wantDetailSub: "session expired",
		},
		{
			name:         "ErrForbidden",
			err:          apperror.ErrForbidden,
			wantStatus:   http.StatusForbidden,
			wantCode:     "FORBIDDEN",
			wantDetailSub: "forbidden",
		},
		{
			name:         "Wrapped ErrForbidden",
			err:          apperror.Wrap(apperror.ErrForbidden, "not your session"),
			wantStatus:   http.StatusForbidden,
			wantCode:     "FORBIDDEN",
			wantDetailSub: "not your session",
		},
		{
			name:         "ErrConflict",
			err:          apperror.ErrConflict,
			wantStatus:   http.StatusConflict,
			wantCode:     "CONFLICT",
			wantDetailSub: "conflict",
		},
		{
			name:         "Wrapped ErrConflict",
			err:          apperror.Wrap(apperror.ErrConflict, "slot already taken"),
			wantStatus:   http.StatusConflict,
			wantCode:     "CONFLICT",
			wantDetailSub: "slot already taken",
		},
		{
			name:         "ErrValidation",
			err:          apperror.ErrValidation,
			wantStatus:   http.StatusUnprocessableEntity,
			wantCode:     "VALIDATION",
			wantDetailSub: "validation",
		},
		{
			name:         "Wrapped ErrValidation",
			err:          apperror.Wrap(apperror.ErrValidation, "invalid arrow score"),
			wantStatus:   http.StatusUnprocessableEntity,
			wantCode:     "VALIDATION",
			wantDetailSub: "invalid arrow score",
		},
		{
			name:         "Unknown standard error",
			err:          errors.New("database connection refused"),
			wantStatus:   http.StatusInternalServerError,
			wantCode:     "INTERNAL_ERROR",
			wantDetailSub: "internal server error",
		},
		{
			name:         "Nil error",
			err:          nil,
			wantStatus:   http.StatusOK,
			wantCode:     "",
			wantDetailSub: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, resp := middleware.MapError(tc.err)
			if status != tc.wantStatus {
				t.Errorf("MapError status = %d, want %d", status, tc.wantStatus)
			}
			if resp.Code != tc.wantCode {
				t.Errorf("MapError code = %q, want %q", resp.Code, tc.wantCode)
			}
			if tc.wantDetailSub != "" && !errorsContains(resp.Detail, tc.wantDetailSub) {
				t.Errorf("MapError detail = %q, want substring %q", resp.Detail, tc.wantDetailSub)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := apperror.Wrap(apperror.ErrNotFound, "archer not found")

	middleware.WriteError(rec, err)

	if rec.Code != http.StatusNotFound {
		t.Errorf("WriteError status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("WriteError Content-Type = %q, want application/json", ct)
	}

	var resp middleware.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("WriteError response is not valid JSON: %v", err)
	}
	if resp.Code != "NOT_FOUND" {
		t.Errorf("WriteError response Code = %q, want NOT_FOUND", resp.Code)
	}
	if resp.Detail != "not found: archer not found" {
		t.Errorf("WriteError response Detail = %q, want %q", resp.Detail, "not found: archer not found")
	}
}

func TestErrorMapper_MiddlewarePassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	wrapped := middleware.ErrorMapper(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("ErrorMapper status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func errorsContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || stringContains(s, sub))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/middleware/... -v
```
Expected: Compilation failure because `middleware.MapError`, `middleware.WriteError`, `middleware.ErrorResponse`, and `middleware.ErrorMapper` are not defined.

- [x] **Step 3: Write minimal implementation for error mapper**

Create `backend/internal/middleware/error_mapper.go`:

```go
package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

// ErrorResponse represents an HTTP error response body matching frontend expectations.
type ErrorResponse struct {
	Detail string `json:"detail"`
	Code   string `json:"code,omitempty"`
}

// MapError maps domain errors to their corresponding HTTP status codes and response bodies.
func MapError(err error) (int, ErrorResponse) {
	if err == nil {
		return http.StatusOK, ErrorResponse{}
	}

	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code() {
		case apperror.ErrNotFound.Code():
			return http.StatusNotFound, ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		case apperror.ErrUnauthorized.Code():
			return http.StatusUnauthorized, ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		case apperror.ErrForbidden.Code():
			return http.StatusForbidden, ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		case apperror.ErrConflict.Code():
			return http.StatusConflict, ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		case apperror.ErrValidation.Code():
			return http.StatusUnprocessableEntity, ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		}
	}

	return http.StatusInternalServerError, ErrorResponse{
		Detail: "internal server error",
		Code:   "INTERNAL_ERROR",
	}
}

// WriteError serializes and writes an error response formatted as JSON.
func WriteError(w http.ResponseWriter, err error) {
	status, resp := MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// ErrorMapper is an HTTP middleware that serves as a boundary filter.
func ErrorMapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/middleware/... -v
```
Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add backend/internal/middleware/error_mapper.go backend/internal/middleware/error_mapper_test.go
git commit -m "feat(middleware): add error-to-HTTP mapper and JSON error writer"
```

---

### Task 3: Auth Domain Service Token Validation Extension

**Files:**
- Modify: `backend/internal/auth/service.go`
- Modify: `backend/internal/auth/service_test.go`

**Interfaces:**
- Consumes: `auth.Service`, `DecodeJWT`, `DecodeSessionID`, `HashSessionToken`, `ValidateSession`
- Produces: `(s *Service) Authenticate(ctx context.Context, tokenStr string) (uuid.UUID, error)`

- [x] **Step 1: Write failing tests for Service.Authenticate**

Append to `backend/internal/auth/service_test.go`:

```go
func TestService_Authenticate_Success(t *testing.T) {
	archerID := uuid.New()
	secret := "test-secret-key-minimum-length"
	cfg := auth.Config{
		JWTSecret:    secret,
		JWTAlgorithm: "HS256",
	}

	rawSession := []byte("12345678901234567890123456789012")
	sid := auth.EncodeSessionID(rawSession)
	tokenHash := auth.HashSessionToken(rawSession)

	now := time.Now().UTC()
	jwtToken, err := auth.BuildJWT(archerID, sid, now, now.Add(time.Hour), secret, "HS256")
	if err != nil {
		t.Fatalf("BuildJWT() error: %v", err)
	}

	sessionRepo := &mockSessionRepo{
		findByTokenHashFn: func(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
			if !bytes.Equal(hash, tokenHash) {
				return nil, nil
			}
			return &model.AuthSessionRead{
				ArcherID:         archerID,
				SessionTokenHash: tokenHash,
				CreatedAt:        now,
				ExpiresAt:        now.Add(time.Hour),
			}, nil
		},
	}

	svc := auth.NewService(&mockArcherRepo{}, sessionRepo, cfg)

	gotID, err := svc.Authenticate(context.Background(), jwtToken)
	if err != nil {
		t.Fatalf("Authenticate() unexpected error: %v", err)
	}
	if gotID != archerID {
		t.Errorf("Authenticate() got archerID %v, want %v", gotID, archerID)
	}
}

func TestService_Authenticate_ExpiredJWT(t *testing.T) {
	archerID := uuid.New()
	secret := "test-secret-key-minimum-length"
	cfg := auth.Config{
		JWTSecret:    secret,
		JWTAlgorithm: "HS256",
	}

	past := time.Now().UTC().Add(-2 * time.Hour)
	jwtToken, err := auth.BuildJWT(archerID, "sid", past, past.Add(time.Hour), secret, "HS256")
	if err != nil {
		t.Fatalf("BuildJWT() error: %v", err)
	}

	svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, cfg)

	_, err = svc.Authenticate(context.Background(), jwtToken)
	if err == nil {
		t.Fatal("Authenticate() expected error for expired JWT, got nil")
	}
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Errorf("Authenticate() error = %v, want ErrUnauthorized", err)
	}
}

func TestService_Authenticate_RevokedSession(t *testing.T) {
	archerID := uuid.New()
	secret := "test-secret-key-minimum-length"
	cfg := auth.Config{
		JWTSecret:    secret,
		JWTAlgorithm: "HS256",
	}

	rawSession := []byte("12345678901234567890123456789012")
	sid := auth.EncodeSessionID(rawSession)
	tokenHash := auth.HashSessionToken(rawSession)

	now := time.Now().UTC()
	jwtToken, err := auth.BuildJWT(archerID, sid, now, now.Add(time.Hour), secret, "HS256")
	if err != nil {
		t.Fatalf("BuildJWT() error: %v", err)
	}

	revokedAt := now.Add(-5 * time.Minute)
	sessionRepo := &mockSessionRepo{
		findByTokenHashFn: func(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
			return &model.AuthSessionRead{
				ArcherID:         archerID,
				SessionTokenHash: tokenHash,
				CreatedAt:        now,
				ExpiresAt:        now.Add(time.Hour),
				RevokedAt:        &revokedAt,
			}, nil
		},
	}

	svc := auth.NewService(&mockArcherRepo{}, sessionRepo, cfg)

	_, err = svc.Authenticate(context.Background(), jwtToken)
	if err == nil {
		t.Fatal("Authenticate() expected error for revoked session, got nil")
	}
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Errorf("Authenticate() error = %v, want ErrUnauthorized", err)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/auth/... -v -run TestService_Authenticate
```
Expected: Compilation failure because `svc.Authenticate` is not defined.

- [x] **Step 3: Implement Service.Authenticate**

Add method to `backend/internal/auth/service.go`:

```go
// Authenticate verifies and decodes an access JWT, validates the underlying session in the database,
// ensures the session matches the archer, and returns the authenticated archer UUID.
func (s *Service) Authenticate(ctx context.Context, tokenStr string) (uuid.UUID, error) {
	claims, err := DecodeJWT(tokenStr, s.cfg.JWTSecret, s.cfg.JWTAlgorithm)
	if err != nil {
		return uuid.Nil, err
	}

	rawSession, err := DecodeSessionID(claims.SID)
	if err != nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid session id in token")
	}

	tokenHash := HashSessionToken(rawSession)
	session, err := s.ValidateSession(ctx, tokenHash)
	if err != nil {
		return uuid.Nil, err
	}

	archerID, err := claims.ArcherID()
	if err != nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid archer id in token")
	}

	if session.ArcherID != archerID {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "session does not belong to archer")
	}

	return archerID, nil
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/auth/... -v -run TestService_Authenticate
```
Expected: PASS.

- [x] **Step 5: Run full auth test suite**

```bash
cd backend && go test ./internal/auth/... -v
```
Expected: All auth tests PASS.

- [x] **Step 6: Commit**

```bash
git add backend/internal/auth/service.go backend/internal/auth/service_test.go
git commit -m "feat(auth): add Authenticate token validation method to auth.Service"
```

---

### Task 4: Auth Middleware

**Files:**
- Create: `backend/internal/middleware/auth.go`
- Create: `backend/internal/middleware/auth_test.go`

**Interfaces:**
- Consumes:
  - `WithArcherID(ctx, id)`
  - `WriteError(w, err)`
- Produces:
  - `type TokenAuthenticator interface { Authenticate(ctx context.Context, token string) (uuid.UUID, error) }`
  - `const AuthCookieName = "arch_stats_auth"`
  - `Auth(auth TokenAuthenticator) func(http.Handler) http.Handler`

- [x] **Step 1: Write failing tests for auth middleware**

Create `backend/internal/middleware/auth_test.go`:

```go
package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

type mockAuthenticator struct {
	authenticateFn func(ctx context.Context, token string) (uuid.UUID, error)
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, token string) (uuid.UUID, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, token)
	}
	return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "unimplemented mock")
}

func TestAuth_MissingTokenReturns401(t *testing.T) {
	authMw := middleware.Auth(&mockAuthenticator{})
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var errResp middleware.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if errResp.Code != "UNAUTHORIZED" {
		t.Errorf("error code = %q, want UNAUTHORIZED", errResp.Code)
	}
}

func TestAuth_InvalidTokenReturns401(t *testing.T) {
	authMw := middleware.Auth(&mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid token signature")
		},
	})
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", nil)
	req.AddCookie(&http.Cookie{Name: middleware.AuthCookieName, Value: "bad-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ValidCookieAttachesArcherID(t *testing.T) {
	expectedID := uuid.New()
	authMw := middleware.Auth(&mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			if token == "valid-cookie-token" {
				return expectedID, nil
			}
			return uuid.Nil, errors.New("unexpected token")
		},
	})

	var receivedID uuid.UUID
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := middleware.GetArcherID(r.Context())
		if err != nil {
			t.Errorf("GetArcherID returned error: %v", err)
		}
		receivedID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", nil)
	req.AddCookie(&http.Cookie{Name: middleware.AuthCookieName, Value: "valid-cookie-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if receivedID != expectedID {
		t.Errorf("receivedID = %v, want %v", receivedID, expectedID)
	}
}

func TestAuth_ValidAuthorizationHeaderAttachesArcherID(t *testing.T) {
	expectedID := uuid.New()
	authMw := middleware.Auth(&mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			if token == "valid-header-token" {
				return expectedID, nil
			}
			return uuid.Nil, errors.New("unexpected token")
		},
	})

	var receivedID uuid.UUID
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := middleware.GetArcherID(r.Context())
		if err != nil {
			t.Errorf("GetArcherID returned error: %v", err)
		}
		receivedID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", nil)
	req.Header.Set("Authorization", "Bearer valid-header-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if receivedID != expectedID {
		t.Errorf("receivedID = %v, want %v", receivedID, expectedID)
	}
}

func TestAuth_HeaderTakesPrecedenceOverCookie(t *testing.T) {
	headerID := uuid.New()
	authMw := middleware.Auth(&mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			if token == "header-token" {
				return headerID, nil
			}
			return uuid.Nil, errors.New("unexpected token")
		},
	})

	var receivedID uuid.UUID
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedID, _ = middleware.GetArcherID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", nil)
	req.Header.Set("Authorization", "Bearer header-token")
	req.AddCookie(&http.Cookie{Name: middleware.AuthCookieName, Value: "cookie-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if receivedID != headerID {
		t.Errorf("receivedID = %v, want %v", receivedID, headerID)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/middleware/... -v -run TestAuth
```
Expected: Compilation failure because `middleware.Auth` and `middleware.AuthCookieName` are not defined.

- [x] **Step 3: Implement Auth middleware**

Create `backend/internal/middleware/auth.go`:

```go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

// AuthCookieName defines the cookie key for authenticated sessions.
const AuthCookieName = "arch_stats_auth"

// TokenAuthenticator defines the contract for authenticating access tokens.
type TokenAuthenticator interface {
	Authenticate(ctx context.Context, token string) (uuid.UUID, error)
}

// Auth constructs an HTTP middleware that extracts the authentication token
// from either the Authorization header or the arch_stats_auth cookie, validates it
// with the authenticator, and injects the authenticated archer ID into the request context.
func Auth(authenticator TokenAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				WriteError(w, apperror.Wrap(apperror.ErrUnauthorized, "missing authentication token"))
				return
			}

			archerID, err := authenticator.Authenticate(r.Context(), token)
			if err != nil {
				WriteError(w, apperror.Wrap(apperror.ErrUnauthorized, err.Error()))
				return
			}

			ctx := WithArcherID(r.Context(), archerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if trimmed := strings.TrimSpace(parts[1]); trimmed != "" {
				return trimmed
			}
		}
	}

	if cookie, err := r.Cookie(AuthCookieName); err == nil {
		if trimmed := strings.TrimSpace(cookie.Value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/middleware/... -v -run TestAuth
```
Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add backend/internal/middleware/auth.go backend/internal/middleware/auth_test.go
git commit -m "feat(middleware): add Auth authentication middleware and unit tests"
```

---

### Task 5: Request Logging Middleware

**Files:**
- Create: `backend/internal/middleware/logging.go`
- Create: `backend/internal/middleware/logging_test.go`

**Interfaces:**
- Consumes: `log/slog`, `net/http`, `time`
- Produces:
  - `RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler`
  - `Logging(next http.Handler) http.Handler`

- [x] **Step 1: Write failing tests for RequestLogger**

Create `backend/internal/middleware/logging_test.go`:

```go
package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestRequestLogger_LogsFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	loggingMw := middleware.RequestLogger(logger)

	handler := loggingMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v0/shot", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("User-Agent", "ArchStatsClient/1.0")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("log output is not valid JSON: %v, raw: %s", err, buf.String())
	}

	if logEntry["msg"] != "http request" {
		t.Errorf("msg = %v, want 'http request'", logEntry["msg"])
	}
	if logEntry["method"] != "POST" {
		t.Errorf("method = %v, want 'POST'", logEntry["method"])
	}
	if logEntry["path"] != "/api/v0/shot" {
		t.Errorf("path = %v, want '/api/v0/shot'", logEntry["path"])
	}
	if int(logEntry["status"].(float64)) != http.StatusCreated {
		t.Errorf("status = %v, want %d", logEntry["status"], http.StatusCreated)
	}
	if _, ok := logEntry["duration_ms"]; !ok {
		t.Errorf("missing duration_ms in log entry")
	}
}

func TestRequestLogger_DefaultStatus200(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	loggingMw := middleware.RequestLogger(logger)

	handler := loggingMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok without explicit WriteHeader"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("log output is not JSON: %v", err)
	}
	if int(logEntry["status"].(float64)) != http.StatusOK {
		t.Errorf("status = %v, want %d", logEntry["status"], http.StatusOK)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/middleware/... -v -run TestRequestLogger
```
Expected: Compilation failure because `middleware.RequestLogger` is not defined.

- [x] **Step 3: Implement logging middleware**

Create `backend/internal/middleware/logging.go`:

```go
package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status      int
	written     bool
	bytesWritten int64
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.written {
		r.status = status
		r.written = true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.status = http.StatusOK
		r.written = true
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += int64(n)
	return n, err
}

// RequestLogger creates an HTTP middleware that records request execution details
// (method, path, status, duration, IP, user-agent) with slog.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			activeLogger := logger
			if activeLogger == nil {
				activeLogger = slog.Default()
			}

			start := time.Now()
			rec := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				clientIP = r.RemoteAddr
			}

			activeLogger.InfoContext(r.Context(), "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int64("duration_ms", duration.Milliseconds()),
				slog.Duration("duration", duration),
				slog.Int64("bytes", rec.bytesWritten),
				slog.String("ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// Logging is a convenience middleware using the default slog logger.
func Logging(next http.Handler) http.Handler {
	return RequestLogger(nil)(next)
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/middleware/... -v -run TestRequestLogger
```
Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add backend/internal/middleware/logging.go backend/internal/middleware/logging_test.go
git commit -m "feat(middleware): add structured request logging middleware"
```

---

### Task 6: CORS Middleware

**Files:**
- Create: `backend/internal/middleware/cors.go`
- Create: `backend/internal/middleware/cors_test.go`

**Interfaces:**
- Consumes: `net/http`
- Produces:
  - `type CORSOptions struct { ... }`
  - `GetCORSOptions(devMode bool) CORSOptions`
  - `CORS(devMode bool) func(http.Handler) http.Handler`

- [x] **Step 1: Write failing tests for CORS**

Create `backend/internal/middleware/cors_test.go`:

```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestCORS_DevModePreflight(t *testing.T) {
	corsMw := middleware.CORS(true)
	handler := corsMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("preflight OPTIONS request should not reach downstream handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v0/session", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:5173" {
		t.Errorf("Allow-Origin = %q, want http://localhost:5173", origin)
	}
	if creds := rec.Header().Get("Access-Control-Allow-Credentials"); creds != "true" {
		t.Errorf("Allow-Credentials = %q, want 'true'", creds)
	}
	if methods := rec.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("Allow-Methods is empty")
	}
	if headers := rec.Header().Get("Access-Control-Allow-Headers"); headers == "" {
		t.Error("Allow-Headers is empty")
	}
}

func TestCORS_DevModeActualRequest(t *testing.T) {
	corsMw := middleware.CORS(true)
	handlerReached := false
	handler := corsMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerReached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerReached {
		t.Error("actual GET request did not reach downstream handler")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:5173" {
		t.Errorf("Allow-Origin = %q, want http://localhost:5173", origin)
	}
	if creds := rec.Header().Get("Access-Control-Allow-Credentials"); creds != "true" {
		t.Errorf("Allow-Credentials = %q, want 'true'", creds)
	}
}

func TestCORS_ProdModeNoDevOrigins(t *testing.T) {
	corsMw := middleware.CORS(false)
	handlerReached := false
	handler := corsMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerReached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", nil)
	req.Header.Set("Origin", "http://malicious.example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerReached {
		t.Error("expected GET request to reach downstream handler in prod")
	}
	// In strict prod mode without cross-origin whitelist, cross-origin requests should not be reflected
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin == "http://malicious.example.com" {
		t.Errorf("Allow-Origin reflected untrusted origin in prod: %q", origin)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/middleware/... -v -run TestCORS
```
Expected: Compilation failure because `middleware.CORS` is not defined.

- [x] **Step 3: Implement CORS middleware**

Create `backend/internal/middleware/cors.go`:

```go
package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOptions holds configurable settings for CORS.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// GetCORSOptions returns default CORS options depending on devMode.
func GetCORSOptions(devMode bool) CORSOptions {
	if devMode {
		return CORSOptions{
			AllowedOrigins: []string{
				"http://localhost:5173",
				"http://127.0.0.1:5173",
				"http://localhost:8000",
				"http://localhost:3000",
			},
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
				http.MethodHead,
			},
			AllowedHeaders: []string{
				"Accept",
				"Authorization",
				"Content-Type",
				"X-CSRF-Token",
				"Origin",
			},
			AllowCredentials: true,
			MaxAge:           300,
		}
	}

	return CORSOptions{
		AllowedOrigins: []string{},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"Origin",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

// CORS constructs an HTTP middleware that configures Cross-Origin Resource Sharing headers.
func CORS(devMode bool) func(http.Handler) http.Handler {
	opts := GetCORSOptions(devMode)
	allowedMethods := strings.Join(opts.AllowedMethods, ", ")
	allowedHeaders := strings.Join(opts.AllowedHeaders, ", ")
	maxAgeStr := strconv.Itoa(opts.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			isAllowed := false
			if devMode {
				// In development mode, allow localhost/127.0.0.1 origins or any if devMode
				isAllowed = true
			} else {
				for _, allowed := range opts.AllowedOrigins {
					if allowed == origin {
						isAllowed = true
						break
					}
				}
			}

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if opts.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			if r.Method == http.MethodOptions {
				if isAllowed {
					w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
					w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
					w.Header().Set("Access-Control-Max-Age", maxAgeStr)
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/middleware/... -v -run TestCORS
```
Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add backend/internal/middleware/cors.go backend/internal/middleware/cors_test.go
git commit -m "feat(middleware): add CORS configuration middleware for dev and prod modes"
```

---

### Task 7: Panic Recovery Middleware

**Files:**
- Create: `backend/internal/middleware/recovery.go`
- Create: `backend/internal/middleware/recovery_test.go`

**Interfaces:**
- Consumes: `WriteError(w, err)`, `log/slog`, `runtime/debug`
- Produces: `Recovery(next http.Handler) http.Handler`

- [x] **Step 1: Write failing tests for Recovery**

Create `backend/internal/middleware/recovery_test.go`:

```go
package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestRecovery_CatchesPanicAndReturns500(t *testing.T) {
	handler := middleware.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went critically wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	// Should not crash the test process
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp middleware.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Code != "INTERNAL_ERROR" {
		t.Errorf("response Code = %q, want INTERNAL_ERROR", resp.Code)
	}
	if resp.Detail != "internal server error" {
		t.Errorf("response Detail = %q, want 'internal server error'", resp.Detail)
	}
}

func TestRecovery_NormalHandlerUnaffected(t *testing.T) {
	handler := middleware.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthy"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthy", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "healthy" {
		t.Errorf("body = %q, want 'healthy'", rec.Body.String())
	}
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/middleware/... -v -run TestRecovery
```
Expected: Compilation failure because `middleware.Recovery` is not defined.

- [x] **Step 3: Implement Recovery middleware**

Create `backend/internal/middleware/recovery.go`:

```go
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery returns an HTTP middleware that catches unhandled panics,
// logs the stack trace with slog.Error, and writes a standardized 500 JSON error response.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := string(debug.Stack())
				slog.ErrorContext(r.Context(), "panic recovered in HTTP handler",
					slog.Any("panic", rec),
					slog.String("stack", stack),
					slog.String("path", r.URL.Path),
					slog.String("method", r.Method),
				)

				WriteError(w, fmt.Errorf("panic: %v", rec))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/middleware/... -v -run TestRecovery
```
Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add backend/internal/middleware/recovery.go backend/internal/middleware/recovery_test.go
git commit -m "feat(middleware): add panic recovery middleware with slog stack trace"
```

---

### Task 8: Clean-up Placeholder, Vet, Lint, and Final Verification

**Files:**
- Delete: `backend/internal/middleware/.gitkeep`

**Interfaces:**
- Consumes: all middleware packages
- Produces: clean code repository passing tests, vet, and lint

- [x] **Step 1: Delete .gitkeep**

```bash
rm -f backend/internal/middleware/.gitkeep
```

- [x] **Step 2: Run middleware unit tests**

```bash
cd backend && go test -race ./internal/middleware/... -v
```
Expected: All tests PASS.

- [x] **Step 3: Run all backend unit tests**

```bash
cd backend && go test -race ./... -v
```
Expected: All tests in all packages PASS.

- [x] **Step 4: Run go vet**

```bash
cd backend && go vet ./...
```
Expected: 0 issues.

- [x] **Step 5: Run golangci-lint**

```bash
cd backend && golangci-lint run ./...
```
Expected: 0 issues.

- [x] **Step 6: Run go build**

```bash
cd backend && go build ./...
```
Expected: Compiles cleanly.

- [x] **Step 7: Commit**

```bash
git add -A
git commit -m "chore(middleware): remove .gitkeep and verify middleware stack"
```

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-09-05-middleware-stack.md`. Two execution options:

1. **Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. **Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
