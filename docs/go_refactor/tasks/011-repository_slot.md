# Task 011: Build Repository — Slot

## Git Branch

`refactor/011-repository-slot`

## Objective

Implement the slot repository for managing shooting slots within a session. This maps from the
Python `slot_model.py` and `slot_manager.py`, covering CRUD operations for slots including
their relationship to sessions and score tracking.

## Dependencies

- Task 008 (repository base patterns + DBTX interface)
- Task 007 (slot model structs)

## Acceptance Criteria

- [ ] `backend/internal/repository/slot.go` implements `SlotRepo` with methods:
    - `FindByID(ctx, id) (*model.SlotRead, error)`
    - `FindBySessionID(ctx, sessionID) ([]model.SlotRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)`
    - `Update(ctx, data, filter) error`
    - `Delete(ctx, id) error`
    - `CountBySessionID(ctx, sessionID) (int, error)`
- [ ] All queries use squirrel.
- [ ] Unit tests verify query building for all methods.
- [ ] `go test ./internal/repository/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/repository/slot.go` |
| Create | `backend/internal/repository/slot_test.go` |

## Reference

- Python slot model: [slot_model.py](file:///home/juanpa/Projects/arch-stats/backend/src/models/slot_model.py)
- Python slot manager: [slot_manager.py](file:///home/juanpa/Projects/arch-stats/backend/src/core/slot_manager.py)

## Steps

- [ ] **Step 1: Write failing tests**

  Create `backend/internal/repository/slot_test.go`:
    - Test `FindBySessionID` builds SELECT with WHERE session_id = $1 ORDER BY slot_number
    - Test `Create` builds INSERT with session_id, slot_number, distance, face_type, etc.
    - Test `Update` builds UPDATE with appropriate SET and WHERE clauses
    - Test `CountBySessionID` builds SELECT COUNT(*) with WHERE session_id

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [ ] **Step 3: Implement `slot.go`**

- [ ] **Step 4: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [ ] **Step 5: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "feat: add slot repository with session relationship queries"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
