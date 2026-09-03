# Task 010: Build Repository — Shooting Session

## Git Branch

`refactor/010-repository-shooting-session`

## Objective

Implement the shooting session repository for managing archery shooting sessions. This maps
from the Python `session_model.py` and `session_manager.py`, covering CRUD operations for
shooting sessions with their status transitions (open/closed).

## Dependencies

- Task 008 (repository base patterns + DBTX interface)
- Task 007 (session model structs)

## Acceptance Criteria

- [x] `backend/internal/repository/session.go` implements `SessionRepo` with methods:
    - `FindByID(ctx, id) (*model.SessionRead, error)`
    - `FindAll(ctx, filter) ([]model.SessionRead, error)`
    - `FindOpen(ctx, archerID) (*model.SessionRead, error)` — find the currently open session
    - `Create(ctx, data) (uuid.UUID, error)`
    - `Update(ctx, data, filter) error`
    - `Close(ctx, id) error` — set status to closed, set ended_at timestamp
    - `Delete(ctx, id) error`
- [x] All queries use squirrel.
- [x] Unit tests verify query building for all methods.
- [x] `go test ./internal/repository/...` passes.
- [x] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/repository/session.go` |
| Create | `backend/internal/repository/session_test.go` |

## Reference

- Python session model: [session_model.py](file:///home/juanpa/Projects/arch-stats/backend/src/models/session_model.py)
- Python session manager: [session_manager.py](file:///home/juanpa/Projects/arch-stats/backend/src/core/session_manager.py)

## Steps

- [x] **Step 1: Write failing tests**

  Create `backend/internal/repository/session_test.go`:
    - Test `FindByID` builds SELECT with correct columns and WHERE clause
    - Test `FindOpen` builds SELECT with WHERE status = 'open' AND archer_id = $1
    - Test `Create` builds INSERT with all required columns
    - Test `Close` builds UPDATE setting status and ended_at

- [x] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [x] **Step 3: Implement `session.go`**

  Implement `SessionRepo` struct with all methods using squirrel and the `DBTX` interface.

- [x] **Step 4: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [x] **Step 5: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [x] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "feat: add shooting session repository with open/close logic"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
