# Task 016: Build Service Layer — Face and Target

## Git Branch

`refactor/016-service-layer-face-and-target`

## Objective

Build the service layer for face and target domains. Face data is largely read-only (target face
definitions); the target service manages target assignments per slot.

## Dependencies

- Task 002 (apperror for domain errors)
- Task 013 (face and target repositories)

## Acceptance Criteria

- [ ] `backend/internal/service/face.go` implements `FaceService` with methods:
    - `GetByID(ctx, id) (*model.FaceRead, error)`
    - `ListAll(ctx) ([]model.FaceRead, error)`
    - `ListByType(ctx, faceType) ([]model.FaceRead, error)`
- [ ] `backend/internal/service/target.go` implements `TargetService` with methods:
    - `GetByID(ctx, id) (*model.TargetRead, error)`
    - `ListBySlotID(ctx, slotID) ([]model.TargetRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)`
    - `Update(ctx, id, data) error`
    - `Delete(ctx, id) error`
- [ ] Services accept repository interfaces via constructor injection.
- [ ] Unit tests with mock repositories verify business logic.
- [ ] `go test ./internal/service/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/service/face.go` |
| Create | `backend/internal/service/face_test.go` |
| Create | `backend/internal/service/target.go` |
| Create | `backend/internal/service/target_test.go` |

## Steps

- [ ] **Step 1: Write failing tests for FaceService**

  Create `backend/internal/service/face_test.go`:
    - Test `GetByID` returns `apperror.ErrNotFound` when face doesn't exist
    - Test `ListAll` returns all faces from repository

- [ ] **Step 2: Write failing tests for TargetService**

  Create `backend/internal/service/target_test.go`:
    - Test `Create` validates face_id exists before inserting
    - Test `GetByID` returns `apperror.ErrNotFound` when target doesn't exist

- [ ] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/service/... -v
  ```

- [ ] **Step 4: Implement `face.go` and `target.go`**

- [ ] **Step 5: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/service/... -v
  ```

- [ ] **Step 6: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [ ] **Step 7: Commit**

  ```bash
  git add -A
  git commit -m "feat: add face and target service layers"
  ```

## Verification

- `cd backend && go test ./internal/service/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
