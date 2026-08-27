# Task 023: Build HTTP Handler — Shots Endpoints

## Git Branch

`refactor/023-handler-shots`

## Objective

Implement the shots HTTP handler in `internal/handler/`, porting the Python `shot_router.py`.
This provides endpoints for recording arrow shots: create (single or batch), list by slot, and
count by slot. All endpoints require authentication.

## Dependencies

- Task 015 (shot service)
- Task 018 (middleware — auth, error mapper)
- Task 019 (handler helpers)

## Acceptance Criteria

- [ ] `backend/internal/handler/shot.go` implements `ShotHandler` with methods:
    - `Create(w, r)` — POST `/api/v0/shot` — accepts single shot or array of shots, returns 201
    - `GetBySlot(w, r)` — GET `/api/v0/shot/by-slot/{slot_id}` — list shots for a slot
    - `CountBySlot(w, r)` — GET `/api/v0/shot/count-by-slot/{slot_id}` — count shots in a slot
- [ ] The `Create` endpoint handles both `ShotCreate` and `[]ShotCreate` payloads (matching
  the Python union type `ShotCreate | list[ShotCreate]`).
- [ ] All endpoints extract authenticated archer ID from request context.
- [ ] Unit tests using `httptest` with mock service verify:
    - Create single shot returns 201 + shot ID
    - Create batch returns 201 + array of shot IDs
    - GetBySlot returns 200 + array of shots
    - CountBySlot returns 200 + integer count
- [ ] `go test ./internal/handler/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/handler/shot.go` |
| Create | `backend/internal/handler/shot_test.go` |

## Reference

- Python router: [shot_router.py](file:///home/juanpa/Projects/arch-stats/backend/src/routers/v0/shot_router.py)
- 3 endpoints total, note the union type for create

## Steps

- [ ] **Step 1: Write failing tests**

  Create `backend/internal/handler/shot_test.go`:
    - Define mock `shotService` interface
    - Test single and batch create
    - Test list and count by slot

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 3: Implement `shot.go`**

  For the union type (single vs. batch create), use `json.RawMessage` to peek at the payload
  structure before unmarshaling:

  ```go
  // Peek: if starts with '[', it's a batch
  raw, _ := io.ReadAll(r.Body)
  if len(raw) > 0 && raw[0] == '[' {
      // unmarshal as []model.ShotCreate
  } else {
      // unmarshal as model.ShotCreate
  }
  ```

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
  git commit -m "feat: add shots HTTP handler with single/batch create"
  ```

## Verification

- `cd backend && go test ./internal/handler/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
