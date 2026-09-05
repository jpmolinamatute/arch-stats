# Task 014: Build Service Layer — Archer and Shooting Session

## Git Branch

`refactor/014-service-layer-archer-and-session`

## Objective

Build the service (business logic) layer for archer and shooting session domains. The service
layer sits between handlers and repositories, encapsulating business rules, validation, and
orchestration of repository calls.

## Dependencies

- Task 002 (apperror for domain errors)
- Task 008 (archer repository)
- Task 010 (session repository)

## Acceptance Criteria

- [x] `backend/internal/service/archer.go` implements `ArcherService` with methods:
    - `GetByID(ctx, id) (*model.ArcherRead, error)` — returns `apperror.ErrNotFound` if missing
    - `List(ctx, filter) ([]model.ArcherRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)` — validates required fields
    - `Update(ctx, id, data) error` — returns `apperror.ErrNotFound` if missing
- [x] `backend/internal/service/session.go` implements `SessionService` with methods:
    - `GetByID(ctx, id) (*model.SessionRead, error)`
    - `GetOpen(ctx, archerID) (*model.SessionRead, error)`
    - `List(ctx, filter) ([]model.SessionRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)` — validates no other open session exists
    - `Close(ctx, id) error` — validates session is currently open
- [x] Services accept repository interfaces via constructor injection (testable with mocks).
- [x] Unit tests use mock repositories to verify business logic.
- [x] `go test ./internal/service/...` passes.
- [x] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/service/archer.go` |
| Create | `backend/internal/service/archer_test.go` |
| Create | `backend/internal/service/session.go` |
| Create | `backend/internal/service/session_test.go` |
| Delete | `backend/internal/service/.gitkeep` |

## Steps

- [x] **Step 1: Define repository interfaces for mock injection**

  In each service file, define an interface that the repository must satisfy. This decouples
  the service from the concrete repository implementation.

  ```go
  // In archer.go
  type archerRepository interface {
      FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
      FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
      Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
      Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error
  }
  ```

- [x] **Step 2: Write failing tests for ArcherService**

  Create `backend/internal/service/archer_test.go`:
    - Test `GetByID` returns `apperror.ErrNotFound` when repo returns nil
    - Test `GetByID` returns the archer when repo finds one
    - Test `Create` validates required fields before calling repo

- [x] **Step 3: Write failing tests for SessionService**

  Create `backend/internal/service/session_test.go`:
    - Test `Create` returns `apperror.ErrConflict` when an open session already exists
    - Test `Close` returns `apperror.ErrNotFound` when session doesn't exist
    - Test `Close` returns `apperror.ErrValidation` when session is already closed

- [x] **Step 4: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/service/... -v
  ```

- [x] **Step 5: Implement `archer.go` and `session.go`**

- [x] **Step 6: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/service/... -v
  ```

- [x] **Step 7: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [x] **Step 8: Commit**

  ```bash
  rm -f backend/internal/service/.gitkeep
  git add -A
  git commit -m "feat: add archer and session service layers with business logic"
  ```

## Verification

- `cd backend && go test ./internal/service/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
