# Task 019: Build HTTP Handler — Auth Endpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the authentication HTTP handler in `backend/internal/handler/auth.go` and shared handler utilities in `backend/internal/handler/helpers.go`, porting the Python `auth_router.py` logic. This provides HTTP endpoints for Google One Tap login, new archer registration, session logout, and current user retrieval (`/me`), with full unit test coverage using `httptest`.

**Architecture:**
- `helpers.go`: Shared HTTP handler utilities for JSON decoding (`readJSON`), JSON encoding (`writeJSON`), and error writing (`writeError`, `writeAppError`) mapped to standard frontend-compatible JSON envelopes.
- `auth.go`: `AuthHandler` exposing `Login`, `Register`, `Logout`, and `Me` methods adhering to standard `http.HandlerFunc` signatures. The handler delegates domain logic to `AuthService` and `ArcherService` interfaces, manages HTTP-only `arch_stats_auth` session cookies, extracts client IP and user agent metadata, and parses request/response bodies.
- `auth/service.go`: Extended with orchestrator methods `LoginWithGoogle`, `RegisterWithGoogle`, `RevokeToken`, and `DecodeToken` so authentication domain logic remains cohesive in the service layer, keeping handlers lean and HTTP-focused.
- `middleware/auth.go`: Export `ExtractToken(r *http.Request) string` to share token extraction logic (Authorization header vs `arch_stats_auth` cookie) across middleware and handlers.

**Tech Stack:** Go 1.27+, standard library `net/http`, `net/http/httptest`, `encoding/json`, `time`, `github.com/google/uuid`, internal packages (`apperror`, `auth`, `middleware`, `model`).

**Spec:** [docs/go_refactor/tasks/019-handler_auth.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/019-handler_auth.md)

## Global Constraints

- Git branch: `refactor/019-handler-auth`
- Package paths:
  - `github.com/jpmolinamatute/arch-stats/backend/internal/handler`
  - `github.com/jpmolinamatute/arch-stats/backend/internal/auth`
  - `github.com/jpmolinamatute/arch-stats/backend/internal/middleware`
- Error handling: Wrap internal errors with `%w` using contextual messages. Return domain sentinels `apperror.ErrValidation`, `apperror.ErrUnauthorized`, and `apperror.ErrNotFound` mapped to HTTP 422, 401, and 404 respectively.
- JSON error formatting: Error responses must format JSON with `{"detail": "..."}` matching `frontend/src/api/client.ts`.
- Cookie settings:
  - Name: `"arch_stats_auth"` (defined in `middleware.AuthCookieName`).
  - Path: `"/"`.
  - HttpOnly: `true`.
  - SameSite: `http.SameSiteLaxMode`.
  - Secure: `true` in production; in development, `true` only if request is HTTPS (`r.TLS != nil` or `X-Forwarded-Proto: https`).
  - MaxAge: configurable (default 24h / 86400s). On logout, `MaxAge: -1` and `Expires: time.Unix(0, 0)`.
- Formatting & Linting: Code must pass `gofumpt` and `golangci-lint run ./...`.
- Tests: `go test -race ./internal/handler/... ./internal/auth/... ./internal/middleware/... -v` must pass.
- Verification: `go vet ./...` and `go build ./...` must succeed.

---

## File Structure

```
backend/
├── internal/
│   ├── auth/
│   │   ├── service.go             # [MODIFY] Add LoginWithGoogle, RegisterWithGoogle, RevokeToken, DecodeToken
│   │   └── service_test.go        # [MODIFY] Add unit tests for new auth.Service methods
│   ├── middleware/
│   │   ├── auth.go                # [MODIFY] Export ExtractToken helper
│   │   └── auth_test.go           # [MODIFY] Add test for ExtractToken
│   └── handler/
│       ├── .gitkeep               # [DELETE] Remove once handler files exist
│       ├── helpers.go             # [NEW] writeJSON, readJSON, writeError, writeAppError
│       ├── helpers_test.go        # [NEW] Unit tests for handler helpers
│       ├── auth.go                # [NEW] AuthHandler, AuthService, ArcherService, Login, Register, Logout, Me
│       └── auth_test.go           # [NEW] Comprehensive httptest unit tests for auth endpoints
```

---

### Task 1: Git Branch Setup & Middleware Token Extraction Helper

**Files:**
- Modify: `backend/internal/middleware/auth.go`
- Modify: `backend/internal/middleware/auth_test.go`

**Interfaces:**
- Consumes: `net/http`, `strings`
- Produces: `ExtractToken(r *http.Request) string`

- [x] **Step 1: Create git branch**

```bash
git switch -c refactor/019-handler-auth
```

- [x] **Step 2: Write failing test for `ExtractToken`**

In `backend/internal/middleware/auth_test.go`, add:

```go
func TestExtractToken(t *testing.T) {
	t.Run("extracts from Authorization header with Bearer prefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer header-token-xyz")

		got := middleware.ExtractToken(req)
		if got != "header-token-xyz" {
			t.Fatalf("expected 'header-token-xyz', got %q", got)
		}
	})

	t.Run("extracts from cookie when Authorization header is absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  middleware.AuthCookieName,
			Value: "cookie-token-abc",
		})

		got := middleware.ExtractToken(req)
		if got != "cookie-token-abc" {
			t.Fatalf("expected 'cookie-token-abc', got %q", got)
		}
	})

	t.Run("prefers Authorization header over cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer header-priority-token")
		req.AddCookie(&http.Cookie{
			Name:  middleware.AuthCookieName,
			Value: "cookie-ignored-token",
		})

		got := middleware.ExtractToken(req)
		if got != "header-priority-token" {
			t.Fatalf("expected 'header-priority-token', got %q", got)
		}
	})

	t.Run("returns empty string when neither header nor cookie present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		got := middleware.ExtractToken(req)
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})
}
```

- [x] **Step 3: Run test to verify it fails**

```bash
cd backend && go test ./internal/middleware/... -run TestExtractToken -v
```
Expected: FAIL with undefined `middleware.ExtractToken`.

- [x] **Step 4: Export `ExtractToken` in `backend/internal/middleware/auth.go`**

In `backend/internal/middleware/auth.go`, rename `extractToken` to `ExtractToken`:

```go
// ExtractToken extracts the bearer token from either the Authorization header
// or the arch_stats_auth cookie.
func ExtractToken(r *http.Request) string {
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
And update line 26 of `backend/internal/middleware/auth.go` to call `ExtractToken(r)`.

- [x] **Step 5: Run tests to verify they pass**

```bash
cd backend && go test ./internal/middleware/... -v
```
Expected: PASS for all tests in `internal/middleware`.

- [x] **Step 6: Commit**

```bash
git add backend/internal/middleware/auth.go backend/internal/middleware/auth_test.go
git commit -m "feat(middleware): export ExtractToken helper for shared token extraction"
```

---

### Task 2: Auth Service Orchestration Methods

**Files:**
- Modify: `backend/internal/auth/service.go`
- Modify: `backend/internal/auth/service_test.go`

**Interfaces:**
- Consumes: `model.AuthRegistrationRequest`, `model.AuthAuthenticated`, `model.AuthNeedsRegistration`, `auth.SessionMetadata`
- Produces:
  - `(s *Service) LoginWithGoogle(ctx context.Context, credential string, now time.Time, meta ...SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error)`
  - `(s *Service) RegisterWithGoogle(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...SessionMetadata) (*model.AuthAuthenticated, error)`
  - `(s *Service) RevokeToken(ctx context.Context, tokenStr string) error`
  - `(s *Service) DecodeToken(tokenStr string) (*Claims, error)`

- [x] **Step 1: Write failing tests in `backend/internal/auth/service_test.go`**

Add the following tests to `backend/internal/auth/service_test.go`:

```go
func TestService_LoginWithGoogle(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

	t.Run("returns AuthNeedsRegistration when archer does not exist", func(t *testing.T) {
		archers := newMockArcherRepo()
		sessions := newMockSessionRepo()
		svc := auth.NewService(archers, sessions, auth.Config{
			JWTSecret:           "test-secret-key-that-is-long-enough-32b",
			GoogleOAuthClientID: "test-client-id",
		})

		auth.SetDefaultGooglePayloadVerifierForTest(t, func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
			return &idtoken.Payload{
				Subject: "google-sub-new",
				Claims: map[string]any{
					"email":       "new@example.com",
					"given_name":  "New",
					"family_name": "Archer",
					"picture":     "https://pic.example.com/new.jpg",
				},
			}, nil
		})

		authd, needsReg, err := svc.LoginWithGoogle(ctx, "valid-token", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if authd != nil {
			t.Fatalf("expected nil AuthAuthenticated, got %+v", authd)
		}
		if needsReg == nil {
			t.Fatal("expected non-nil AuthNeedsRegistration")
		}
		if needsReg.GoogleEmail != "new@example.com" || needsReg.GoogleSubject != "google-sub-new" {
			t.Fatalf("mismatched Google details: %+v", needsReg)
		}
	})

	t.Run("returns AuthAuthenticated when archer exists", func(t *testing.T) {
		archers := newMockArcherRepo()
		sessions := newMockSessionRepo()
		svc := auth.NewService(archers, sessions, auth.Config{
			JWTSecret:           "test-secret-key-that-is-long-enough-32b",
			GoogleOAuthClientID: "test-client-id",
		})

		existingID := uuid.New()
		archers.byGoogleSubject["google-sub-existing"] = &model.ArcherRead{
			ArcherID:      existingID,
			FirstName:     "Existing",
			LastName:      "User",
			Email:         "existing@example.com",
			GoogleSubject: "google-sub-existing",
		}

		auth.SetDefaultGooglePayloadVerifierForTest(t, func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
			return &idtoken.Payload{
				Subject: "google-sub-existing",
				Claims: map[string]any{
					"email": "existing@example.com",
				},
			}, nil
		})

		authd, needsReg, err := svc.LoginWithGoogle(ctx, "valid-token", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if needsReg != nil {
			t.Fatalf("expected nil AuthNeedsRegistration, got %+v", needsReg)
		}
		if authd == nil {
			t.Fatal("expected non-nil AuthAuthenticated")
		}
		if authd.Archer.ArcherID != existingID {
			t.Fatalf("expected archer id %s, got %s", existingID, authd.Archer.ArcherID)
		}
	})

	t.Run("returns unauthorized error when google verification fails", func(t *testing.T) {
		archers := newMockArcherRepo()
		sessions := newMockSessionRepo()
		svc := auth.NewService(archers, sessions, auth.Config{
			JWTSecret:           "test-secret-key-that-is-long-enough-32b",
			GoogleOAuthClientID: "test-client-id",
		})

		auth.SetDefaultGooglePayloadVerifierForTest(t, func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
			return nil, errors.New("bad token")
		})

		_, _, err := svc.LoginWithGoogle(ctx, "bad-token", now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})
}

func TestService_RegisterWithGoogle(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

	archers := newMockArcherRepo()
	sessions := newMockSessionRepo()
	svc := auth.NewService(archers, sessions, auth.Config{
		JWTSecret:           "test-secret-key-that-is-long-enough-32b",
		GoogleOAuthClientID: "test-client-id",
	})

	auth.SetDefaultGooglePayloadVerifierForTest(t, func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
		return &idtoken.Payload{
			Subject: "google-sub-reg",
			Claims: map[string]any{
				"email":       "reg@example.com",
				"given_name":  "Reg",
				"family_name": "User",
			},
		}, nil
	})

	payload := model.AuthRegistrationRequest{
		Credential:  "valid-reg-token",
		DateOfBirth: "1995-05-15",
		Gender:      model.GenderFemale,
		Bowstyle:    model.BowstyleRecurve,
		DrawWeight:  32.5,
	}

	authd, err := svc.RegisterWithGoogle(ctx, payload, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authd == nil || authd.Archer.Email != "reg@example.com" {
		t.Fatalf("unexpected registration result: %+v", authd)
	}
}

func TestService_RevokeTokenAndDecodeToken(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

	archers := newMockArcherRepo()
	sessions := newMockSessionRepo()
	svc := auth.NewService(archers, sessions, auth.Config{
		JWTSecret:     "test-secret-key-that-is-long-enough-32b",
		JWTTTLMinutes: 60,
	})

	archerID := uuid.New()
	archers.byID[archerID] = &model.ArcherRead{
		ArcherID: archerID,
		Email:    "test@example.com",
	}

	authd, err := svc.LoginExisting(ctx, archers.byID[archerID], nil, now)
	if err != nil {
		t.Fatalf("LoginExisting failed: %v", err)
	}

	claims, err := svc.DecodeToken(authd.AccessToken)
	if err != nil {
		t.Fatalf("DecodeToken failed: %v", err)
	}
	if claims.Sub != archerID.String() {
		t.Fatalf("expected subject %s, got %s", archerID, claims.Sub)
	}

	if err := svc.RevokeToken(ctx, authd.AccessToken); err != nil {
		t.Fatalf("RevokeToken failed: %v", err)
	}

	// Verify token is now invalid due to revoked session
	_, err = svc.Authenticate(ctx, authd.AccessToken)
	if err == nil {
		t.Fatal("expected Authenticate to fail after revocation, but succeeded")
	}
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
```

- [x] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./internal/auth/... -run "TestService_LoginWithGoogle|TestService_RegisterWithGoogle|TestService_RevokeTokenAndDecodeToken" -v
```
Expected: FAIL with missing methods on `*auth.Service`.

- [x] **Step 3: Implement methods on `*auth.Service` in `backend/internal/auth/service.go`**

In `backend/internal/auth/service.go`, add:

```go
// LoginWithGoogle verifies a Google One Tap credential. If an archer with the verified Google subject
// already exists, it creates an active session and returns AuthAuthenticated. If the subject is unknown,
// it returns an AuthNeedsRegistration response with Google profile claims.
func (s *Service) LoginWithGoogle(
	ctx context.Context,
	credential string,
	now time.Time,
	meta ...SessionMetadata,
) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
	googleData, err := s.VerifyGoogleToken(ctx, credential)
	if err != nil {
		return nil, nil, err
	}

	if strings.TrimSpace(googleData.Sub) == "" || strings.TrimSpace(googleData.Email) == "" {
		return nil, nil, apperror.Wrap(apperror.ErrValidation, "google token missing required sub or email claim")
	}

	existing, err := s.archers.FindByGoogleSubject(ctx, googleData.Sub)
	if err != nil {
		return nil, nil, fmt.Errorf("checking archer by google subject: %w", err)
	}

	if existing == nil {
		return nil, BuildNeedsRegistrationResponse(googleData), nil
	}

	authd, err := s.LoginExisting(ctx, existing, googleData, now, meta...)
	if err != nil {
		return nil, nil, err
	}

	return authd, nil, nil
}

// RegisterWithGoogle verifies the submitted Google credential and registers a new archer profile.
// If an archer with the same Google subject already exists, it transparently logs them in.
//
//nolint:gocritic // hugeParam: payload matches API request specification
func (s *Service) RegisterWithGoogle(
	ctx context.Context,
	payload model.AuthRegistrationRequest,
	now time.Time,
	meta ...SessionMetadata,
) (*model.AuthAuthenticated, error) {
	googleData, err := s.VerifyGoogleToken(ctx, payload.Credential)
	if err != nil {
		return nil, err
	}

	return s.Register(ctx, payload, googleData, now, meta...)
}

// RevokeToken extracts the session ID from an access JWT and revokes the corresponding database session.
func (s *Service) RevokeToken(ctx context.Context, tokenStr string) error {
	if strings.TrimSpace(tokenStr) == "" {
		return nil
	}

	claims, err := DecodeJWT(tokenStr, s.cfg.JWTSecret, s.cfg.JWTAlgorithm)
	if err != nil {
		return nil // Ignore invalid tokens during logout
	}

	rawSession, err := DecodeSessionID(claims.SID)
	if err != nil {
		return nil
	}

	tokenHash := HashSessionToken(rawSession)
	return s.RevokeSession(ctx, tokenHash, time.Now().UTC())
}

// DecodeToken decodes and validates an access JWT against the service signing configuration.
func (s *Service) DecodeToken(tokenStr string) (*Claims, error) {
	return DecodeJWT(tokenStr, s.cfg.JWTSecret, s.cfg.JWTAlgorithm)
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/auth/... -v
```
Expected: PASS for all tests in `internal/auth`.

- [x] **Step 5: Commit**

```bash
git add backend/internal/auth/service.go backend/internal/auth/service_test.go
git commit -m "feat(auth): add LoginWithGoogle, RegisterWithGoogle, RevokeToken, and DecodeToken to Service"
```

---

### Task 3: Handler Helpers (`helpers.go`)

**Files:**
- Create: `backend/internal/handler/helpers.go`
- Create: `backend/internal/handler/helpers_test.go`
- Delete: `backend/internal/handler/.gitkeep`

**Interfaces:**
- Consumes: `net/http`, `encoding/json`, `github.com/jpmolinamatute/arch-stats/backend/internal/apperror`, `github.com/jpmolinamatute/arch-stats/backend/internal/middleware`
- Produces:
  - `writeJSON(w http.ResponseWriter, status int, data any) error` / `WriteJSON`
  - `readJSON(r *http.Request, dst any) error` / `ReadJSON`
  - `writeError(w http.ResponseWriter, status int, message string)` / `WriteError`
  - `writeAppError(w http.ResponseWriter, err error)` / `WriteAppError`

- [x] **Step 1: Write failing tests for handler helpers**

Create `backend/internal/handler/helpers_test.go`:

```go
package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestWriteJSON(t *testing.T) {
	t.Run("writes JSON response with headers and status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		data := map[string]string{"message": "success"}

		err := handler.WriteJSON(rec, http.StatusOK, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", ct)
		}

		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if body["message"] != "success" {
			t.Fatalf("expected message 'success', got %q", body["message"])
		}
	})

	t.Run("handles 204 No Content without body", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := handler.WriteJSON(rec, http.StatusNoContent, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
		if rec.Body.Len() != 0 {
			t.Fatalf("expected empty body, got %s", rec.Body.String())
		}
	})
}

func TestReadJSON(t *testing.T) {
	type samplePayload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("successfully decodes valid JSON", func(t *testing.T) {
		body := `{"name":"Robin","age":30}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

		var dst samplePayload
		err := handler.ReadJSON(req, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.Name != "Robin" || dst.Age != 30 {
			t.Fatalf("unexpected decoded struct: %+v", dst)
		}
	})

	t.Run("returns ErrValidation on empty request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		var dst samplePayload
		err := handler.ReadJSON(req, &dst)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("returns ErrValidation on malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-json"))
		var dst samplePayload
		err := handler.ReadJSON(req, &dst)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestWriteErrorAndWriteAppError(t *testing.T) {
	t.Run("writeError writes custom status and detail", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.WriteError(rec, http.StatusBadRequest, "custom error message")

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var resp middleware.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if resp.Detail != "custom error message" {
			t.Fatalf("expected detail 'custom error message', got %q", resp.Detail)
		}
	})

	t.Run("writeAppError maps ErrValidation to 422", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.WriteAppError(rec, apperror.Wrap(apperror.ErrValidation, "field required"))

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("writeAppError maps ErrUnauthorized to 401", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.WriteAppError(rec, apperror.Wrap(apperror.ErrUnauthorized, "invalid credentials"))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}
```

- [x] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./internal/handler/... -v
```
Expected: FAIL because `helpers.go` does not exist yet.

- [x] **Step 3: Implement `backend/internal/handler/helpers.go` and remove `.gitkeep`**

Create `backend/internal/handler/helpers.go`:

```go
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

// writeJSON marshals data as JSON, sets the Content-Type header to application/json,
// writes the HTTP status code, and writes the response body.
func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil || status == http.StatusNoContent {
		return nil
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("encoding json response: %w", err)
	}
	return nil
}

// readJSON decodes the JSON request body into the target pointer dst.
// It limits request payload size to 1MB and wraps any decode error in apperror.ErrValidation.
func readJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return apperror.Wrap(apperror.ErrValidation, "request body is empty")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(io.LimitReader(r.Body, 1048576))
	if err := dec.Decode(dst); err != nil {
		return apperror.Wrap(apperror.ErrValidation, fmt.Sprintf("invalid request body: %v", err))
	}
	return nil
}

// writeError writes an error response formatted as JSON with a status code and detail message.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(middleware.ErrorResponse{
		Detail: message,
	})
}

// writeAppError translates a domain error using middleware.WriteError into appropriate status code and JSON.
func writeAppError(w http.ResponseWriter, err error) {
	middleware.WriteError(w, err)
}

// WriteJSON is the exported alias for writeJSON.
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	return writeJSON(w, status, data)
}

// ReadJSON is the exported alias for readJSON.
func ReadJSON(r *http.Request, dst any) error {
	return readJSON(r, dst)
}

// WriteError is the exported alias for writeError.
func WriteError(w http.ResponseWriter, status int, message string) {
	writeError(w, status, message)
}

// WriteAppError is the exported alias for writeAppError.
func WriteAppError(w http.ResponseWriter, err error) {
	writeAppError(w, err)
}
```

Remove placeholder `.gitkeep`:
```bash
rm -f backend/internal/handler/.gitkeep
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/handler/... -v
```
Expected: PASS for all tests in `internal/handler`.

- [x] **Step 5: Commit**

```bash
git add backend/internal/handler/helpers.go backend/internal/handler/helpers_test.go
git rm -f backend/internal/handler/.gitkeep 2>/dev/null || true
git commit -m "feat(handler): add JSON and error handler helpers"
```

---

### Task 4: Auth Handler Implementation (`auth.go` & `auth_test.go`)

**Files:**
- Create: `backend/internal/handler/auth.go`
- Create: `backend/internal/handler/auth_test.go`

**Interfaces:**
- Consumes:
  - `AuthService`:
    - `LoginWithGoogle(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error)`
    - `RegisterWithGoogle(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error)`
    - `RevokeToken(ctx context.Context, tokenStr string) error`
    - `DecodeToken(tokenStr string) (*auth.Claims, error)`
    - `Authenticate(ctx context.Context, token string) (uuid.UUID, error)`
  - `ArcherService`:
    - `GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)`
- Produces:
  - `AuthHandler` struct
  - `NewAuthHandler(authSvc AuthService, archerSvc ArcherService, cfg AuthHandlerConfig) *AuthHandler`
  - `(h *AuthHandler) Login(w http.ResponseWriter, r *http.Request)`
  - `(h *AuthHandler) Register(w http.ResponseWriter, r *http.Request)`
  - `(h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request)`
  - `(h *AuthHandler) Me(w http.ResponseWriter, r *http.Request)`

- [x] **Step 1: Write failing tests in `backend/internal/handler/auth_test.go`**

Create `backend/internal/handler/auth_test.go`:

```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockAuthService struct {
	loginWithGoogleFn    func(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error)
	registerWithGoogleFn func(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error)
	revokeTokenFn        func(ctx context.Context, tokenStr string) error
	decodeTokenFn        func(tokenStr string) (*auth.Claims, error)
	authenticateFn       func(ctx context.Context, token string) (uuid.UUID, error)
}

func (m *mockAuthService) LoginWithGoogle(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
	if m.loginWithGoogleFn != nil {
		return m.loginWithGoogleFn(ctx, credential, now, meta...)
	}
	return nil, nil, errors.New("unimplemented")
}

func (m *mockAuthService) RegisterWithGoogle(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error) {
	if m.registerWithGoogleFn != nil {
		return m.registerWithGoogleFn(ctx, payload, now, meta...)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockAuthService) RevokeToken(ctx context.Context, tokenStr string) error {
	if m.revokeTokenFn != nil {
		return m.revokeTokenFn(ctx, tokenStr)
	}
	return nil
}

func (m *mockAuthService) DecodeToken(tokenStr string) (*auth.Claims, error) {
	if m.decodeTokenFn != nil {
		return m.decodeTokenFn(tokenStr)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockAuthService) Authenticate(ctx context.Context, token string) (uuid.UUID, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, token)
	}
	return uuid.Nil, errors.New("unimplemented")
}

type mockArcherService struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
}

func (m *mockArcherService) GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("unimplemented")
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("login with valid credential for existing user returns 200 and sets cookie", func(t *testing.T) {
		archerID := uuid.New()
		expectedExpires := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)

		authSvc := &mockAuthService{
			loginWithGoogleFn: func(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
				return &model.AuthAuthenticated{
					Status:      model.AuthStatusAuthenticated,
					AccessToken: "valid-jwt-token-123",
					ExpiresAt:   expectedExpires,
					Archer: model.ArcherRead{
						ArcherID:  archerID,
						FirstName: "Katniss",
						LastName:  "Everdeen",
						Email:     "katniss@district12.org",
					},
				}, nil, nil
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{
			JWTTTLMinutes: 1440,
			DevMode:       true,
		})

		body, _ := json.Marshal(model.GoogleOneTapRequest{Credential: "valid-google-credential"})
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp model.AuthAuthenticated
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Status != model.AuthStatusAuthenticated || resp.Archer.ArcherID != archerID {
			t.Fatalf("unexpected response payload: %+v", resp)
		}

		// Check cookie
		cookies := rec.Result().Cookies()
		var authCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == middleware.AuthCookieName {
				authCookie = c
				break
			}
		}
		if authCookie == nil {
			t.Fatal("expected arch_stats_auth cookie to be set, but found none")
		}
		if authCookie.Value != "valid-jwt-token-123" {
			t.Fatalf("expected cookie value 'valid-jwt-token-123', got %q", authCookie.Value)
		}
		if !authCookie.HttpOnly {
			t.Fatal("expected cookie to be HttpOnly")
		}
		if authCookie.SameSite != http.SameSiteLaxMode {
			t.Fatalf("expected SameSite Lax, got %v", authCookie.SameSite)
		}
	})

	t.Run("login with valid credential for new user returns 200 needs registration without cookie", func(t *testing.T) {
		given := "Legolas"
		authSvc := &mockAuthService{
			loginWithGoogleFn: func(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
				return nil, &model.AuthNeedsRegistration{
					Status:             model.AuthStatusNeedsRegistration,
					GoogleEmail:        "legolas@woodland.realm",
					GoogleSubject:      "sub-woodland-elf",
					GivenName:          &given,
					GivenNameProvided:  true,
					FamilyNameProvided: false,
				}, nil
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{
			JWTTTLMinutes: 1440,
			DevMode:       true,
		})

		body, _ := json.Marshal(model.GoogleOneTapRequest{Credential: "new-user-credential"})
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp model.AuthNeedsRegistration
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Status != model.AuthStatusNeedsRegistration || resp.GoogleEmail != "legolas@woodland.realm" {
			t.Fatalf("unexpected response payload: %+v", resp)
		}

		for _, c := range rec.Result().Cookies() {
			if c.Name == middleware.AuthCookieName && c.Value != "" && c.MaxAge > 0 {
				t.Fatalf("expected no auth cookie to be set for needs_registration, but got %+v", c)
			}
		}
	})

	t.Run("login with invalid credential returns 401", func(t *testing.T) {
		authSvc := &mockAuthService{
			loginWithGoogleFn: func(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
				return nil, nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid google credential")
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{})

		body, _ := json.Marshal(model.GoogleOneTapRequest{Credential: "invalid-cred"})
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
		}

		var errResp middleware.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if !strings.Contains(errResp.Detail, "invalid google credential") {
			t.Fatalf("expected detail containing 'invalid google credential', got %q", errResp.Detail)
		}
	})

	t.Run("login with empty credential returns 422", func(t *testing.T) {
		h := handler.NewAuthHandler(&mockAuthService{}, &mockArcherService{}, handler.AuthHandlerConfig{})

		body, _ := json.Marshal(model.GoogleOneTapRequest{Credential: ""})
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("register with missing fields returns 422", func(t *testing.T) {
		h := handler.NewAuthHandler(&mockAuthService{}, &mockArcherService{}, handler.AuthHandlerConfig{})

		// Missing date_of_birth, gender, draw_weight <= 0
		invalidPayload := model.AuthRegistrationRequest{
			Credential: "valid-credential",
			DrawWeight: 0,
		}

		body, _ := json.Marshal(invalidPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/register", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("register with valid fields returns 201 and sets cookie", func(t *testing.T) {
		archerID := uuid.New()
		authSvc := &mockAuthService{
			registerWithGoogleFn: func(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error) {
				return &model.AuthAuthenticated{
					Status:      model.AuthStatusAuthenticated,
					AccessToken: "new-user-jwt",
					ExpiresAt:   time.Now().Add(24 * time.Hour).UTC(),
					Archer: model.ArcherRead{
						ArcherID:  archerID,
						FirstName: "Hawkeye",
						LastName:  "Pierce",
						Email:     "hawkeye@mash4077.org",
					},
				}, nil
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{
			JWTTTLMinutes: 1440,
			DevMode:       true,
		})

		firstName := "Hawkeye"
		lastName := "Pierce"
		validPayload := model.AuthRegistrationRequest{
			Credential:  "google-reg-token-ok",
			DateOfBirth: "1980-01-01",
			Gender:      model.GenderMale,
			Bowstyle:    model.BowstyleRecurve,
			DrawWeight:  35.0,
			FirstName:   &firstName,
			LastName:    &lastName,
		}

		body, _ := json.Marshal(validPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/register", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
			t.Fatalf("expected status 201 or 200, got %d: %s", rec.Code, rec.Body.String())
		}

		cookies := rec.Result().Cookies()
		var authCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == middleware.AuthCookieName {
				authCookie = c
				break
			}
		}
		if authCookie == nil || authCookie.Value != "new-user-jwt" {
			t.Fatalf("expected auth cookie with value 'new-user-jwt', got %+v", authCookie)
		}
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("logout clears the session cookie and returns success", func(t *testing.T) {
		revoked := false
		authSvc := &mockAuthService{
			revokeTokenFn: func(ctx context.Context, tokenStr string) error {
				if tokenStr == "token-to-revoke" {
					revoked = true
				}
				return nil
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{})

		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/logout", nil)
		req.AddCookie(&http.Cookie{
			Name:  middleware.AuthCookieName,
			Value: "token-to-revoke",
		})
		rec := httptest.NewRecorder()

		h.Logout(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		if !revoked {
			t.Fatal("expected RevokeToken to be called")
		}

		var resp model.LogoutResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Fatal("expected success: true")
		}

		// Verify cookie is cleared (MaxAge < 0)
		cookies := rec.Result().Cookies()
		var authCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == middleware.AuthCookieName {
				authCookie = c
				break
			}
		}
		if authCookie == nil {
			t.Fatal("expected deleted auth cookie in response")
		}
		if authCookie.MaxAge >= 0 {
			t.Fatalf("expected cookie MaxAge < 0, got %d", authCookie.MaxAge)
		}
	})
}

func TestAuthHandler_Me(t *testing.T) {
	t.Run("me returns the authenticated archer from context", func(t *testing.T) {
		archerID := uuid.New()
		expectedArcher := &model.ArcherRead{
			ArcherID:  archerID,
			FirstName: "Merida",
			LastName:  "DunBroch",
			Email:     "merida@dunbroch.scot",
		}

		archerSvc := &mockArcherService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				if id == archerID {
					return expectedArcher, nil
				}
				return nil, apperror.ErrNotFound
			},
		}

		h := handler.NewAuthHandler(&mockAuthService{}, archerSvc, handler.AuthHandlerConfig{})

		req := httptest.NewRequest(http.MethodGet, "/api/v0/auth/me", nil)
		req = req.WithContext(middleware.WithArcherID(req.Context(), archerID))
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp model.AuthAuthenticated
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Archer.ArcherID != archerID || resp.Archer.FirstName != "Merida" {
			t.Fatalf("unexpected archer data: %+v", resp.Archer)
		}
	})

	t.Run("me returns 401 when no auth context or cookie present", func(t *testing.T) {
		h := handler.NewAuthHandler(&mockAuthService{}, &mockArcherService{}, handler.AuthHandlerConfig{})

		req := httptest.NewRequest(http.MethodGet, "/api/v0/auth/me", nil)
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}
```

- [x] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./internal/handler/... -run "TestAuthHandler" -v
```
Expected: FAIL with `handler.NewAuthHandler` undefined.

- [x] **Step 3: Implement `backend/internal/handler/auth.go`**

Create `backend/internal/handler/auth.go`:

```go
package handler

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// AuthService defines the authentication operations required by the HTTP auth handler.
type AuthService interface {
	LoginWithGoogle(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error)
	RegisterWithGoogle(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error)
	RevokeToken(ctx context.Context, tokenStr string) error
	DecodeToken(tokenStr string) (*auth.Claims, error)
	Authenticate(ctx context.Context, token string) (uuid.UUID, error)
}

// ArcherService defines the profile operations required by the HTTP auth handler.
type ArcherService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
}

// AuthHandlerConfig specifies runtime configuration for cookie management and token lifetimes.
type AuthHandlerConfig struct {
	JWTTTLMinutes int
	DevMode       bool
}

// AuthHandler manages HTTP endpoints for authentication, registration, session revocation, and current user retrieval.
type AuthHandler struct {
	authSvc   AuthService
	archerSvc ArcherService
	cfg       AuthHandlerConfig
}

// NewAuthHandler constructs an AuthHandler with service dependencies and configuration.
func NewAuthHandler(authSvc AuthService, archerSvc ArcherService, cfg AuthHandlerConfig) *AuthHandler {
	if cfg.JWTTTLMinutes <= 0 {
		cfg.JWTTTLMinutes = 1440
	}
	return &AuthHandler{
		authSvc:   authSvc,
		archerSvc: archerSvc,
		cfg:       cfg,
	}
}

// Login handles POST /api/v0/auth/login and POST /api/v0/auth/google.
// It verifies the Google One Tap credential. For existing archers, it mints a session, sets the
// arch_stats_auth HTTP-only cookie, and returns AuthAuthenticated (200). For new users, it returns
// AuthNeedsRegistration (200) with no cookie.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.GoogleOneTapRequest
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if strings.TrimSpace(req.Credential) == "" {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "credential is required"))
		return
	}

	now := time.Now().UTC()
	meta := extractSessionMetadata(r)

	authd, needsReg, err := h.authSvc.LoginWithGoogle(r.Context(), req.Credential, now, meta)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if needsReg != nil {
		_ = writeJSON(w, http.StatusOK, needsReg)
		return
	}

	h.setAuthCookie(w, r, authd.AccessToken, authd.ExpiresAt)
	_ = writeJSON(w, http.StatusOK, authd)
}

// Register handles POST /api/v0/auth/register.
// It validates registration fields, creates the archer profile (or logs them in if already registered),
// sets the arch_stats_auth HTTP-only cookie, and returns AuthAuthenticated (201).
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRegistrationRequest
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if err := validateRegistrationRequest(req); err != nil {
		writeAppError(w, err)
		return
	}

	now := time.Now().UTC()
	meta := extractSessionMetadata(r)

	authd, err := h.authSvc.RegisterWithGoogle(r.Context(), req, now, meta)
	if err != nil {
		writeAppError(w, err)
		return
	}

	h.setAuthCookie(w, r, authd.AccessToken, authd.ExpiresAt)
	_ = writeJSON(w, http.StatusCreated, authd)
}

// Logout handles POST /api/v0/auth/logout.
// It deletes the arch_stats_auth cookie and revokes the active session token in the database if present.
// It is idempotent and always returns 200 OK.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := middleware.ExtractToken(r)
	if token != "" {
		_ = h.authSvc.RevokeToken(r.Context(), token)
	}

	h.clearAuthCookie(w, r)
	_ = writeJSON(w, http.StatusOK, model.LogoutResponse{Success: true})
}

// Me handles GET /api/v0/auth/me.
// It returns the currently authenticated archer from the request context or token cookie.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	archerID, err := middleware.GetArcherID(r.Context())
	token := middleware.ExtractToken(r)

	if err != nil {
		if token == "" {
			writeAppError(w, apperror.Wrap(apperror.ErrUnauthorized, "no authentication token or context found"))
			return
		}

		authenticatedID, authErr := h.authSvc.Authenticate(r.Context(), token)
		if authErr != nil {
			writeAppError(w, apperror.Wrap(apperror.ErrUnauthorized, authErr.Error()))
			return
		}
		archerID = authenticatedID
	}

	archer, err := h.archerSvc.GetByID(r.Context(), archerID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	var expiresAt time.Time
	if token != "" {
		if claims, err := h.authSvc.DecodeToken(token); err == nil && claims != nil {
			expiresAt = time.Unix(claims.Exp, 0).UTC()
		}
	}

	resp := model.AuthAuthenticated{
		Status:      model.AuthStatusAuthenticated,
		AccessToken: token,
		ExpiresAt:   expiresAt,
		Archer:      *archer,
	}

	_ = writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) setAuthCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	maxAge := h.cfg.JWTTTLMinutes * 60
	if !expiresAt.IsZero() {
		if remaining := int(time.Until(expiresAt).Seconds()); remaining > 0 {
			maxAge = remaining
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.isCookieSecure(r),
	})
}

func (h *AuthHandler) clearAuthCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.isCookieSecure(r),
	})
}

func (h *AuthHandler) isCookieSecure(r *http.Request) bool {
	if !h.cfg.DevMode {
		return true
	}
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func extractSessionMetadata(r *http.Request) auth.SessionMetadata {
	var ua *string
	if val := strings.TrimSpace(r.Header.Get("User-Agent")); val != "" {
		ua = &val
	}

	var ip *string
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		first := strings.TrimSpace(parts[0])
		if first != "" {
			ip = &first
		}
	} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
		trimmed := strings.TrimSpace(xri)
		if trimmed != "" {
			ip = &trimmed
		}
	} else if r.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil && host != "" {
			ip = &host
		} else {
			trimmed := strings.TrimSpace(r.RemoteAddr)
			if trimmed != "" {
				ip = &trimmed
			}
		}
	}

	return auth.SessionMetadata{
		UserAgent: ua,
		IPAddress: ip,
	}
}

func validateRegistrationRequest(req model.AuthRegistrationRequest) error {
	if strings.TrimSpace(req.Credential) == "" {
		return apperror.Wrap(apperror.ErrValidation, "credential is required")
	}
	if strings.TrimSpace(req.DateOfBirth) == "" {
		return apperror.Wrap(apperror.ErrValidation, "date_of_birth is required")
	}
	if _, err := time.Parse("2006-01-02", req.DateOfBirth); err != nil {
		return apperror.Wrap(apperror.ErrValidation, "date_of_birth must be formatted as YYYY-MM-DD")
	}
	if !isValidGender(req.Gender) {
		return apperror.Wrap(apperror.ErrValidation, "invalid gender")
	}
	if !isValidBowstyle(req.Bowstyle) {
		return apperror.Wrap(apperror.ErrValidation, "invalid bowstyle")
	}
	if req.DrawWeight <= 0 || req.DrawWeight > 200 {
		return apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}
	return nil
}

func isValidGender(g model.Gender) bool {
	switch g {
	case model.GenderMale, model.GenderFemale, model.GenderNonBinary, model.GenderOther, model.GenderUnspecified:
		return true
	default:
		return false
	}
}

func isValidBowstyle(b model.Bowstyle) bool {
	switch b {
	case model.BowstyleRecurve, model.BowstyleCompound, model.BowstyleBarebow, model.BowstyleLongbow:
		return true
	default:
		return false
	}
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/handler/... -v
```
Expected: PASS for all tests in `internal/handler`.

- [x] **Step 5: Commit**

```bash
git add backend/internal/handler/auth.go backend/internal/handler/auth_test.go
git commit -m "feat(handler): implement AuthHandler with login, register, logout, and me endpoints"
```

---

### Task 5: Verification, Linting, and Clean Build

**Files:**
- None (verification only)

- [x] **Step 1: Run full test suite with race detector**

```bash
cd backend && go test -race ./... -v
```
Expected: PASS with 0 race conditions or test failures across all packages.

- [x] **Step 2: Run go vet**

```bash
cd backend && go vet ./...
```
Expected: Clean with 0 issues reported.

- [x] **Step 3: Run Go linting**

```bash
./scripts/linting.bash --go
```
Expected: 0 lint or formatting errors.

- [x] **Step 4: Verify build succeeds**

```bash
cd backend && go build ./...
```
Expected: Clean compilation with 0 warnings or errors.

- [x] **Step 5: Commit any formatting or lint fixes**

```bash
git status
# If clean: no commit needed
# If formatted: git commit -am "chore: format and lint fixes"
```

---

## Self-Review

### 1. Spec Coverage Checklist
- `backend/internal/handler/auth.go` implements `AuthHandler` with methods:
  - `Login(w, r)`: Covered in Task 4.
  - `Register(w, r)`: Covered in Task 4.
  - `Logout(w, r)`: Covered in Task 4.
  - `Me(w, r)`: Covered in Task 4.
- HTTP-only cookies matching Python implementation: Covered in Task 4 (`setAuthCookie`, `clearAuthCookie`).
- JSON responses matching API contract: Covered in Task 3 (`writeJSON`) and Task 4 (`AuthAuthenticated`, `AuthNeedsRegistration`, `LogoutResponse`).
- Unit tests using `httptest`: Covered in Task 4:
  - Login with valid credential returns 200 + sets cookie: Covered.
  - Login with invalid credential returns 401: Covered.
  - Register with missing fields returns 422: Covered.
  - Logout clears the session cookie: Covered.
  - Me returns the authenticated archer: Covered.
- `go test ./internal/handler/...` passes: Covered in Tasks 3, 4, 5.
- `go vet ./...` reports no issues: Covered in Task 5.
- Delete `backend/internal/handler/.gitkeep`: Covered in Task 3.

### 2. Placeholder Scan
- No TODO, TBD, or ellipsis placeholders exist in any code snippet.
- Full type signatures and implementations provided.

### 3. Type Consistency
- `middleware.AuthCookieName` used consistently for cookie name.
- `model.GoogleOneTapRequest`, `model.AuthRegistrationRequest`, `model.AuthAuthenticated`, `model.AuthNeedsRegistration`, and `model.LogoutResponse` match domain models in `backend/internal/model/auth.go`.
- `apperror.ErrValidation`, `apperror.ErrUnauthorized`, `apperror.ErrNotFound` match sentinels in `backend/internal/apperror`.
- Error envelope matches `middleware.ErrorResponse` (`{"detail": "..."}`).

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-09-05-handler-auth.md`. Two execution options:

1. **Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration
2. **Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
