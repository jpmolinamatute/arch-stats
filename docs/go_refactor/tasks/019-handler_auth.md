# Task 019: Build HTTP Handler — Auth Endpoints

## Git Branch

`refactor/019-handler-auth`

## Objective

Implement the auth HTTP handler in `internal/handler/`, porting the Python `auth_router.py`.
This handles Google One Tap login, registration, token refresh, and logout endpoints. The auth
handler is the most complex handler due to the multi-step login/registration flow.

## Dependencies

- Task 017 (auth service)
- Task 014 (archer service)
- Task 018 (middleware — auth, error mapper)
- Task 007 (auth model structs)

## Acceptance Criteria

- [ ] `backend/internal/handler/auth.go` implements `AuthHandler` with methods:
    - `Login(w, r)` — POST `/api/v0/auth/login` — Google One Tap credential → JWT + session
    - `Register(w, r)` — POST `/api/v0/auth/register` — complete registration for new archer
    - `Logout(w, r)` — POST `/api/v0/auth/logout` — invalidate session
    - `Me(w, r)` — GET `/api/v0/auth/me` — return current archer from session
- [ ] Handler sets HTTP-only cookies matching the Python implementation's cookie names and settings.
- [ ] Handler returns JSON responses matching the current API contract.
- [ ] Unit tests using `httptest` verify:
    - Login with valid credential returns 200 + sets cookie
    - Login with invalid credential returns 401
    - Register with missing fields returns 422
    - Logout clears the session cookie
    - Me returns the authenticated archer
- [ ] `go test ./internal/handler/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/handler/auth.go` |
| Create | `backend/internal/handler/auth_test.go` |
| Create | `backend/internal/handler/helpers.go` |
| Delete | `backend/internal/handler/.gitkeep` |

## Reference

- Python auth router: [auth_router.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/routers/v0/auth_router.py)

## Steps

- [ ] **Step 1: Create handler helpers**

  Create `backend/internal/handler/helpers.go` with shared utilities:
    - `writeJSON(w, status, data)` — write JSON response with status code
    - `readJSON(r, dst)` — decode JSON request body into struct
    - `writeError(w, status, message)` — write error JSON response

- [ ] **Step 2: Write failing tests for auth handler**

  Create `backend/internal/handler/auth_test.go` using `httptest.NewRecorder()`:
    - Test Login endpoint with mock auth service
    - Test Register endpoint with mock auth service
    - Test Logout endpoint clears cookies
    - Test Me endpoint returns archer data from context

- [ ] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/handler/... -v
  ```

- [ ] **Step 4: Implement `helpers.go` and `auth.go`**
- [ ] **Step 5: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/handler/... -v
  ```

- [ ] **Step 6: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [ ] **Step 7: Commit**

  ```bash
  rm -f backend/internal/handler/.gitkeep
  git add -A
  git commit -m "feat: add auth HTTP handler with login, register, logout, me endpoints"
  ```

## Verification

- `cd backend && go test ./internal/handler/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
