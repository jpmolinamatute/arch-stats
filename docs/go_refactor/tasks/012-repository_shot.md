# Task 012: Build Repository — Shot

## Git Branch

`refactor/012-repository-shot`

## Objective

Implement the shot repository for managing individual arrow shots within a slot. This maps
from the Python `shot_model.py` and `shot_manager.py`, covering CRUD operations for shots
including score recording and position tracking.

## Dependencies

- Task 008 (repository base patterns + DBTX interface)
- Task 007 (shot model structs)

## Acceptance Criteria

- [ ] `backend/internal/repository/shot.go` implements `ShotRepo` with methods:
    - `FindByID(ctx, id) (*model.ShotRead, error)`
    - `FindBySlotID(ctx, slotID) ([]model.ShotRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)`
    - `Update(ctx, data, filter) error`
    - `Delete(ctx, id) error`
- [ ] All queries use squirrel.
- [ ] Unit tests verify query building for all methods.
- [ ] `go test ./internal/repository/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/repository/shot.go` |
| Create | `backend/internal/repository/shot_test.go` |

## Reference

- Python shot model: [shot_model.py](file:///home/juanpa/Projects/arch-stats/backend/src/models/shot_model.py)
- Python shot manager: [shot_manager.py](file:///home/juanpa/Projects/arch-stats/backend/src/core/shot_manager.py)

## Steps

- [ ] **Step 1: Write failing tests**

  Create `backend/internal/repository/shot_test.go`:
    - Test `FindBySlotID` builds SELECT with WHERE slot_id = $1 ORDER BY arrow_number
    - Test `Create` builds INSERT with slot_id, arrow_number, score, x_position, y_position
    - Test `Update` builds UPDATE with appropriate SET and WHERE

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [ ] **Step 3: Implement `shot.go`**

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
  git commit -m "feat: add shot repository with slot relationship queries"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
