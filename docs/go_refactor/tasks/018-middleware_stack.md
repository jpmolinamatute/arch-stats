# Task 018: Build HTTP Middleware Stack

## Git Branch

`refactor/018-middleware-stack`

## Objective

Build the HTTP middleware stack in `internal/middleware/`. This includes authentication
middleware (validate session token → attach user context), error-to-HTTP mapper, structured
request logging, CORS configuration, and panic recovery.

## Dependencies

- Task 002 (apperror for domain error → HTTP status mapping)
- Task 017 (auth service for session validation)
- Task 006 (logging setup)
- Task 003 (config for CORS and dev mode settings)

## Acceptance Criteria

- [ ] `backend/internal/middleware/auth.go` — validates session token from cookie/header, attaches
  archer ID to request context, returns 401 if invalid.
- [ ] `backend/internal/middleware/error_mapper.go` — catches `apperror.*` types from handlers and
  maps them to HTTP status codes:
    - `ErrNotFound` → 404
    - `ErrUnauthorized` → 401
    - `ErrForbidden` → 403
    - `ErrConflict` → 409
    - `ErrValidation` → 422
    - Unknown errors → 500
- [ ] `backend/internal/middleware/logging.go` — logs method, path, status, duration for each
  request using `slog`.
- [ ] `backend/internal/middleware/cors.go` — configures CORS headers based on dev/prod mode.
- [ ] `backend/internal/middleware/recovery.go` — recovers from panics and returns 500.
- [ ] `backend/internal/middleware/context.go` — helpers to get/set archer ID in request context.
- [ ] Unit tests cover:
    - Error mapper maps each sentinel to correct HTTP status
    - Auth middleware rejects requests without valid tokens
    - Auth middleware attaches archer ID to context on valid token
    - Context helpers round-trip correctly
- [ ] `go test ./internal/middleware/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/middleware/auth.go` |
| Create | `backend/internal/middleware/auth_test.go` |
| Create | `backend/internal/middleware/error_mapper.go` |
| Create | `backend/internal/middleware/error_mapper_test.go` |
| Create | `backend/internal/middleware/logging.go` |
| Create | `backend/internal/middleware/cors.go` |
| Create | `backend/internal/middleware/recovery.go` |
| Create | `backend/internal/middleware/context.go` |
| Create | `backend/internal/middleware/context_test.go` |
| Delete | `backend/internal/middleware/.gitkeep` |

## Steps

- [ ] **Step 1: Write failing tests for error mapper**

  Create `backend/internal/middleware/error_mapper_test.go`:
    - Test each sentinel → HTTP status mapping
    - Test unknown error → 500
    - Test error response body is valid JSON

- [ ] **Step 2: Write failing tests for auth middleware**

  Create `backend/internal/middleware/auth_test.go` using `httptest`:
    - Test request without cookie → 401
    - Test request with invalid token → 401
    - Test request with valid token → handler receives archer ID in context

- [ ] **Step 3: Write failing tests for context helpers**

  Create `backend/internal/middleware/context_test.go`:
    - Test set/get archer ID round-trip
    - Test getting archer ID from empty context returns error

- [ ] **Step 4: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/middleware/... -v
  ```

- [ ] **Step 5: Implement all middleware files**
- [ ] **Step 6: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/middleware/... -v
  ```

- [ ] **Step 7: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [ ] **Step 8: Commit**

  ```bash
  rm -f backend/internal/middleware/.gitkeep
  git add -A
  git commit -m "feat: add HTTP middleware stack (auth, error mapper, logging, CORS, recovery)"
  ```

## Verification

- `cd backend && go test ./internal/middleware/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
