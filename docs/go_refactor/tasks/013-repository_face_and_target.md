# Task 013: Build Repository — Face and Target

## Git Branch

`refactor/013-repository-face-and-target`

## Objective

Implement the face and target repositories. The face repository is the highest-effort item
(14KB Python source). It manages target face definitions and their scoring zones. The target
repository manages target configurations per slot.

## Dependencies

- Task 008 (repository base patterns + DBTX interface)
- Task 007 (face and target model structs)

## Acceptance Criteria

- [x] `backend/internal/repository/face.go` implements `FaceRepo` with methods:
    - `FindByID(ctx, id) (*model.FaceRead, error)`
    - `FindAll(ctx) ([]model.FaceRead, error)`
    - `FindByType(ctx, faceType) ([]model.FaceRead, error)`
- [x] `backend/internal/repository/target.go` implements `TargetRepo` with methods:
    - `FindByID(ctx, id) (*model.TargetRead, error)`
    - `FindBySlotID(ctx, slotID) ([]model.TargetRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)`
    - `Update(ctx, data, filter) error`
    - `Delete(ctx, id) error`
- [x] All queries use squirrel.
- [x] Unit tests verify query building for all methods.
- [x] `go test ./internal/repository/...` passes.
- [x] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/repository/face.go` |
| Create | `backend/internal/repository/face_test.go` |
| Create | `backend/internal/repository/target.go` |
| Create | `backend/internal/repository/target_test.go` |

## Reference

- Python face data: [face_data.py](file:///home/juanpa/Projects/arch-stats/backend/src/core/face_data.py)
- Python face schema: [face_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/face_schema.py)
- Python target model: [target_model.py](file:///home/juanpa/Projects/arch-stats/backend/src/models/target_model.py)
- Python target schema: [target_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/target_schema.py)

## Steps

- [x] **Step 1: Write failing tests for FaceRepo**

  Create `backend/internal/repository/face_test.go`:
    - Test `FindAll` builds SELECT with correct columns
    - Test `FindByType` builds SELECT with WHERE face_type = $1

- [x] **Step 2: Write failing tests for TargetRepo**

  Create `backend/internal/repository/target_test.go`:
    - Test `FindBySlotID` builds SELECT with WHERE slot_id = $1
    - Test `Create` builds INSERT with face_id, slot_id

- [x] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [x] **Step 4: Implement `face.go` and `target.go`**

- [x] **Step 5: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/repository/... -v
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
  git commit -m "feat: add face and target repositories"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
