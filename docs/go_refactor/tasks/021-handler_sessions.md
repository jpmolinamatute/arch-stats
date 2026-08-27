# Task 021: Build HTTP Handler — Sessions Endpoints

## Git Branch

`refactor/021-handler-sessions`

## Objective

Implement the sessions HTTP handler in `internal/handler/`, porting the Python
`session_router.py`. This provides endpoints for managing shooting sessions including
create, get, list open/closed, close, and re-open operations. All session endpoints
require authentication.

## Dependencies

- Task 014 (session service)
- Task 018 (middleware — auth, error mapper)
- Task 019 (handler helpers)

## Acceptance Criteria

- [ ] `backend/internal/handler/session.go` implements `SessionHandler` with methods:
    - `GetOpenForArcher(w, r)` — GET `/api/v0/session/archer/{archer_id}/open-session`
    - `GetClosedForArcher(w, r)` — GET `/api/v0/session/archer/{archer_id}/close-session`
    - `GetParticipating(w, r)` — GET `/api/v0/session/archer/{archer_id}/participating`
    - `ListAllOpen(w, r)` — GET `/api/v0/session/open`
    - `Create(w, r)` — POST `/api/v0/session` — returns 201
    - `GetByID(w, r)` — GET `/api/v0/session/{id}`
    - `ReOpen(w, r)` — PATCH `/api/v0/session/re-open`
    - `Close(w, r)` — PATCH `/api/v0/session/close`
- [ ] All endpoints extract the authenticated archer ID from the request context
  (set by auth middleware).
- [ ] Handler delegates business logic to `SessionService`.
- [ ] Error responses: 404 (not found), 409 (conflict — e.g., already open), 422 (validation).
- [ ] Unit tests using `httptest` with mock service verify key flows:
    - Create returns 201 + session ID
    - Close returns 200
    - GetByID with non-existent ID returns 404
    - Create when open session exists returns 409
- [ ] `go test ./internal/handler/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/handler/session.go` |
| Create | `backend/internal/handler/session_test.go` |

## Reference

- Python router: [session_router.py](file:///home/juanpa/Projects/arch-stats/backend/src/routers/v0/session_router.py)
- 8 endpoints total, all requiring auth

## Steps

- [ ] **Step 1: Write failing tests**

  Create `backend/internal/handler/session_test.go`:
    - Define mock `sessionService` interface
    - Test create, close, get, list endpoints
    - Inject authenticated archer ID into request context for tests

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 3: Implement `session.go`**

  Implement `SessionHandler` struct. Use `middleware.GetArcherID(r.Context())` to extract
  the authenticated archer from each request.

- [ ] **Step 4: Run tests to verify they pass**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 5: Run go vet and build**

  ```bash
  cd backend && go vet ./... && go build ./...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "feat: add sessions HTTP handler with create, close, re-open endpoints"
  ```

## Verification

- `cd backend && go test ./internal/handler/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
