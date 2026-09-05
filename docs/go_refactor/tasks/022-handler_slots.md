# Task 022: Build HTTP Handler — Slots Endpoints

## Git Branch

`refactor/022-handler-slots`

## Objective

Implement the slots HTTP handler in `internal/handler/`, porting the Python `slot_router.py`.
This provides endpoints for managing slot assignments within sessions: get current slot, get slot
by ID, join/re-join/leave a session. All endpoints require authentication.

## Dependencies

- Task 015 (slot service)
- Task 018 (middleware — auth, error mapper)
- Task 019 (handler helpers)

## Acceptance Criteria

- [x] `backend/internal/handler/slot.go` implements `SlotHandler` with methods:
    - `GetArcherCurrentSlot(w, r)` — GET `/api/v0/session/slot/archer/{archer_id}`
    - `GetSlot(w, r)` — GET `/api/v0/session/slot/{slot_id}`
    - `JoinSession(w, r)` — POST `/api/v0/session/slot`
    - `ReJoinSession(w, r)` — PATCH `/api/v0/session/slot/re-join/{slot_id}`
    - `LeaveSession(w, r)` — PATCH `/api/v0/session/slot/leave/{slot_id}`
- [x] All endpoints extract the authenticated archer ID from request context.
- [x] Handler delegates business logic to `SlotService`.
- [x] Error responses: 400 (bad request), 403 (forbidden), 404 (not found), 422 (validation).
- [x] Unit tests using `httptest` with mock service verify:
    - GetArcherCurrentSlot returns 200 + full slot info
    - JoinSession with valid payload returns 200 + slot join response
    - LeaveSession returns 200
    - GetSlot with non-existent ID returns 404
- [x] `go test ./internal/handler/...` passes.
- [x] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/handler/slot.go` |
| Create | `backend/internal/handler/slot_test.go` |

## Reference

- Python router: [slot_router.py](file:///home/juanpa/Projects/arch-stats/backend/src/routers/v0/slot_router.py)
- 5 endpoints total, all under `/api/v0/session/slot/` prefix

## Steps

- [x] **Step 1: Write failing tests**

  Create `backend/internal/handler/slot_test.go`:
    - Define mock `slotService` interface
    - Test join, re-join, leave, get endpoints

- [x] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [x] **Step 3: Implement `slot.go`**

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
  git commit -m "feat: add slots HTTP handler with join, re-join, leave endpoints"
  ```

## Verification

- `cd backend && go test ./internal/handler/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
