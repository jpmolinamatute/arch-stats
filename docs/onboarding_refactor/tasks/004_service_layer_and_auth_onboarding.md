# Task 004: Service Layer and Onboarding Business Logic

## Git Branch

`feature/004-service-layer-onboarding`

## Objective

Implement service layer business logic for bow management, arrow management (enforcing the
minimum 3 arrows per set rule), archer profile completeness evaluation (verifying personal profile
and at least 1 registered bow), and update the Google OAuth authentication service to handle the
new onboarding lifecycle.

## Dependencies

- Task 002 (Domain model structs)
- Task 003 (Repository layer)

## Acceptance Criteria

- [ ] New service `backend/internal/service/bow.go` implements `BowService`:
    - [ ] `CreateBow(ctx context.Context, data model.BowCreate) (*model.BowRead, error)`
          validating positive draw weight and recognized bowstyle.
    - [ ] `ListBows(ctx context.Context, archerID uuid.UUID) ([]model.BowRead, error)`
    - [ ] `GetBowCount(ctx context.Context, archerID uuid.UUID) (int, error)`
- [ ] New service `backend/internal/service/arrow.go` implements `ArrowService`:
    - [ ] `CreateArrow(ctx context.Context, data model.ArrowCreate) (*model.ArrowRead, error)`
          defaulting `Status` to `ArrowStatusInUse` if not specified.
    - [ ] `CreateArrowBatch(ctx context.Context, data model.ArrowBatchCreate)`
          `([]model.ArrowRead, error)` enforcing the application rule that an arrow set must group
          at least 3 arrows, and defaulting `Status` to `ArrowStatusInUse`.
    - [ ] `ListArrows(ctx context.Context, archerID uuid.UUID) ([]model.ArrowRead, error)`
    - [ ] `GetArrowCount(ctx context.Context, archerID uuid.UUID) (int, error)`
    - [ ] `GetInUseArrowCount(ctx context.Context, archerID uuid.UUID) (int, error)` (for session
          arrow ceiling)
- [ ] Modified `backend/internal/service/archer.go`:
    - [ ] `CreateArcher(ctx context.Context, data model.ArcherCreate) (*model.ArcherRead, error)`
          validating archer is at least 10 years old and email is valid.
    - [ ] `IsProfileComplete(ctx context.Context, archerID uuid.UUID) (bool, error)` returning true
          if and only if personal profile exists AND at least 1 bow is registered (arrows are
          optional).
- [ ] Modified `backend/internal/auth/service.go`:
    - [ ] `LoginWithGoogle` looks up Google subject in `auth` identity table; creates record if
          first-time sign-in.
    - [ ] If archer profile or mandatory bow is missing, returns status `needs_registration` with
          `AuthNeedsRegistration` payload containing `google_subject` and Google profile info.
    - [ ] If profile is complete (profile + >= 1 bow), returns status `authenticated` with
          session token / JWT.
- [ ] Comprehensive unit tests with mocks in `service_test.go` and `auth_test.go`.
- [ ] `cd backend && go test ./internal/service/... ./internal/auth/... -v` passes.
- [ ] `cd backend && golangci-lint run ./internal/service/... ./internal/auth/...` passes.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/service/bow.go` |
| Create | `backend/internal/service/bow_test.go` |
| Create | `backend/internal/service/arrow.go` |
| Create | `backend/internal/service/arrow_test.go` |
| Modify | `backend/internal/service/archer.go` |
| Modify | `backend/internal/service/archer_test.go` |
| Modify | `backend/internal/auth/service.go` |
| Modify | `backend/internal/auth/service_test.go` |

## Reference

- [PRD.md](../PRD.md)
- [service.go](../../../backend/internal/auth/service.go)
- [archer.go](../../../backend/internal/service/archer.go)

## Steps

- [ ] **Step 1: Write failing tests for BowService and ArrowService**

  Create unit tests checking:
    - `BowService.CreateBow` rejects draw_weight <= 0.
    - `ArrowService.CreateArrowBatch` rejects batches with fewer than 3 arrows.
    - `ArcherService.IsProfileComplete` returns false when 0 bows exist, and true when >= 1 bow
      exists.

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/service/... -v
  ```

- [ ] **Step 3: Implement `bow.go`**

  Define interface and struct:

  ```go
  type BowRepo interface {
      Create(ctx context.Context, data model.BowCreate) (uuid.UUID, error)
      FindByID(ctx context.Context, id uuid.UUID) (*model.BowRead, error)
      FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.BowRead, error)
      CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)
  }

  type BowService struct {
      repo BowRepo
  }
  ```

  Implement business validations for draw weight and name trimming.

- [ ] **Step 4: Implement `arrow.go`**

  Implement `ArrowService`. For `CreateArrowBatch`, validate:

  ```go
  if data.Count < 3 {
      return nil, apperror.Wrap(
          apperror.ErrValidation,
          "arrow set must group at least 3 arrows",
      )
  }
  if data.Status == nil {
      inUse := model.ArrowStatusInUse
      data.Status = &inUse
  }
  ```

- [ ] **Step 5: Implement profile completeness check in `archer.go`**

  In `ArcherService`, inject `BowRepo` (or `BowCountProvider` interface):

  ```go
  func (s *ArcherService) IsProfileComplete(
      ctx context.Context,
      archerID uuid.UUID,
  ) (bool, error) {
      archer, err := s.repo.FindByID(ctx, archerID)
      if err != nil || archer == nil {
          return false, err
      }
      bowCount, err := s.bows.CountByArcherID(ctx, archerID)
      if err != nil {
          return false, err
      }
      return bowCount >= 1, nil
  }
  ```

- [ ] **Step 6: Update `internal/auth/service.go`**

  Update `LoginWithGoogle` to coordinate with `auth_identity` repo and `IsProfileComplete`. Return
  appropriate status based on completeness check.

- [ ] **Step 7: Run all unit tests**

  ```bash
  cd backend
  go test ./internal/service/... ./internal/auth/... -v
  golangci-lint run ./internal/service/... ./internal/auth/...
  ```

- [ ] **Step 8: Commit changes**

  ```bash
  git add backend/internal/service backend/internal/auth
  git commit -m "feat(service): implement bow, arrow, and onboarding completeness logic"
  ```

## Verification

- `cd backend && go test ./internal/service/... ./internal/auth/... -v` exits with code 0.
- `cd backend && golangci-lint run ./internal/service/... ./internal/auth/...` passes cleanly.
- Profile completeness requires profile and >= 1 bow; arrows are verified as optional.
