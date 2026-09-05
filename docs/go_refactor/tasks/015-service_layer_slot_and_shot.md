# Task 015: Build Service Layer — Slot and Shot

## Git Branch

`refactor/015-service-layer-slot-and-shot`

## Objective

Build the service layer for slot and shot domains. These services enforce business rules such as
ensuring a slot belongs to an open session, shots belong to a valid slot, and score constraints
are respected.

## Dependencies

- Task 002 (apperror for domain errors)
- Task 011 (slot repository)
- Task 012 (shot repository)
- Task 010 (session repository — needed to validate session state)

## Acceptance Criteria

- [x] `backend/internal/service/slot.go` implements `SlotService` with methods:
    - `GetByID(ctx, id) (*model.SlotRead, error)`
    - `ListBySessionID(ctx, sessionID) ([]model.SlotRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)` — validates session is open before creating
    - `Update(ctx, id, data) error`
    - `Delete(ctx, id) error`
- [x] `backend/internal/service/shot.go` implements `ShotService` with methods:
    - `GetByID(ctx, id) (*model.ShotRead, error)`
    - `ListBySlotID(ctx, slotID) ([]model.ShotRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)` — validates slot exists
    - `Update(ctx, id, data) error`
    - `Delete(ctx, id) error`
- [x] Services accept repository interfaces via constructor injection.
- [x] Unit tests with mock repositories verify business logic.
- [x] `go test ./internal/service/...` passes.
- [x] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/service/slot.go` |
| Create | `backend/internal/service/slot_test.go` |
| Create | `backend/internal/service/shot.go` |
| Create | `backend/internal/service/shot_test.go` |

## Steps

- [x] **Step 1: Write failing tests for SlotService**

  Create `backend/internal/service/slot_test.go`:
    - Test `Create` returns `apperror.ErrValidation` when session is not open
    - Test `Create` succeeds when session is open
    - Test `GetByID` returns `apperror.ErrNotFound` when slot doesn't exist
    - Test `Delete` returns `apperror.ErrNotFound` when slot doesn't exist

- [x] **Step 2: Write failing tests for ShotService**

  Create `backend/internal/service/shot_test.go`:
    - Test `Create` returns `apperror.ErrNotFound` when slot doesn't exist
    - Test `Create` succeeds with valid slot
    - Test `GetByID` returns `apperror.ErrNotFound` when shot doesn't exist

- [x] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/service/... -v
  ```

- [x] **Step 4: Implement `slot.go` and `shot.go`**

- [x] **Step 5: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/service/... -v
  ```

- [x] **Step 6: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [x] **Step 7: Commit**

  ```bash
  git add -A
  git commit -m "feat: add slot and shot service layers with business validation"
  ```

## Verification

- `cd backend && go test ./internal/service/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
