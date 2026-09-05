# Task 020: Build HTTP Handler — Archers Endpoints

## Git Branch

`refactor/020-handler-archers`

## Objective

Implement the archers HTTP handler in `internal/handler/`, porting the Python
`archer_router.py`. This provides CRUD endpoints for archer management.

## Dependencies

- Task 014 (archer service)
- Task 018 (middleware — auth, error mapper)
- Task 019 (handler helpers — `writeJSON`, `readJSON`, `writeError`)

## Acceptance Criteria

- [x] `backend/internal/handler/archer.go` implements `ArcherHandler` with methods:
    - `List(w, r)` — GET `/api/v0/archer/` — list all archers
    - `GetByID(w, r)` — GET `/api/v0/archer/{id}` — get single archer by UUID
    - `Create(w, r)` — POST `/api/v0/archer/` — create a new archer, returns 201
    - `Update(w, r)` — PATCH `/api/v0/archer/` — update archer fields
    - `Delete(w, r)` — DELETE `/api/v0/archer/{id}` — delete archer, returns 204
- [x] Handler delegates all business logic to `ArcherService`.
- [x] Error responses use the error mapper middleware (404, 422).
- [x] JSON response shapes match the Python API contract (snake_case fields).
- [x] Unit tests using `httptest` with mock service verify:
    - List returns 200 + JSON array
    - GetByID with valid ID returns 200 + archer JSON
    - GetByID with non-existent ID returns 404
    - Create with valid payload returns 201 + `{"archer_id": "..."}`
    - Delete with valid ID returns 204
- [x] `go test ./internal/handler/...` passes.
- [x] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/handler/archer.go` |
| Create | `backend/internal/handler/archer_test.go` |

## Reference

- Python router: [archer_router.py](file:///home/juanpa/Projects/arch-stats/backend/src/routers/v0/archer_router.py)
- Endpoints: `GET /`, `GET /{id}`, `POST /`, `PATCH /`, `DELETE /{id}`

## Steps

- [x] **Step 1: Write failing tests**

  Create `backend/internal/handler/archer_test.go` using `httptest.NewRecorder()`:
    - Define a mock `archerService` interface
    - Test each endpoint with mock responses

- [x] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [x] **Step 3: Implement `archer.go`**

  Implement `ArcherHandler` struct with constructor injection of the service interface.
  Parse UUID path params using `chi.URLParam(r, "id")`.

- [x] **Step 4: Run tests to verify they pass**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [x] **Step 5: Run go vet and build**

  ```bash
  cd backend && go vet ./... && go build ./...
  ```

- [x] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "feat: add archers HTTP handler with CRUD endpoints"
  ```

## Verification

- `cd backend && go test ./internal/handler/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
