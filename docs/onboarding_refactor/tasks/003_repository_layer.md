# Task 003: Repository Layer for Auth, Archer, Bow, and Arrow

## Git Branch

`feature/003-repository-layer`

## Objective

Implement database access repositories using Squirrel query building and pgx/v5 for the
newly separated `auth` identity table, equipment tables (`bow` and `arrow`), and update the
`archer` repository to reflect the personal-profile-only schema.

## Dependencies

- Task 001 (Database migrations and schema)
- Task 002 (Domain model structs)

## Acceptance Criteria

- [ ] New repository `backend/internal/repository/bow.go` implements:
    - [ ] `Create(ctx context.Context, data model.BowCreate) (uuid.UUID, error)`
    - [ ] `FindByID(ctx context.Context, id uuid.UUID) (*model.BowRead, error)`
    - [ ] `FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.BowRead, error)`
    - [ ] `CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)`
- [ ] New repository `backend/internal/repository/arrow.go` implements:
    - [ ] `Create(ctx context.Context, data model.ArrowCreate) (uuid.UUID, error)` (inserts `status`
          , default `'in_use'`)
    - [ ] `CreateBatch(ctx context.Context, data model.ArrowBatchCreate) ([]model.ArrowRead, error)`
          (inserts `status`, default `'in_use'`)
    - [ ] `FindByID(ctx context.Context, id uuid.UUID) (*model.ArrowRead, error)`
    - [ ] `FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.ArrowRead, error)`
    - [ ] `CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)`
    - [ ] `CountInUseByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)` (for arrow
          ceiling)
    - [ ] `FindByArcherAndSet(ctx context.Context, archerID uuid.UUID, set int16)`
          `([]model.ArrowRead, error)`
- [ ] New repository `backend/internal/repository/auth_identity.go` implements:
    - [ ] `FindByGoogleSubject(ctx context.Context, sub string) (*model.AuthIdentityRead, error)`
    - [ ] `FindByID(ctx context.Context, id uuid.UUID) (*model.AuthIdentityRead, error)`
    - [ ] `Create(ctx context.Context, subject string, picture *string) (uuid.UUID, error)`
    - [ ] `UpdateLastLogin(ctx context.Context, id uuid.UUID, login time.Time, pic *string) error`
- [ ] Modified `backend/internal/repository/archer.go`:
    - [ ] Column selection updated: `archer_id`, `email`, `first_name`, `last_name`,
          `date_of_birth`, `gender`, `is_deleted`.
    - [ ] Scanning functions updated to omit legacy equipment/Google columns.
    - [ ] `Create` accepts `model.ArcherCreate` with `ArcherID` provided from `auth`.
- [ ] Repositories support transactions via `WithTx(tx pgx.Tx)`.
- [ ] Unit tests for all repositories using mock `DBTX` verify SQL query and argument construction.
- [ ] `cd backend && go test ./internal/repository/... -v` passes.
- [ ] `cd backend && golangci-lint run ./internal/repository/...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/repository/bow.go` |
| Create | `backend/internal/repository/bow_test.go` |
| Create | `backend/internal/repository/arrow.go` |
| Create | `backend/internal/repository/arrow_test.go` |
| Create | `backend/internal/repository/auth_identity.go` |
| Create | `backend/internal/repository/auth_identity_test.go` |
| Modify | `backend/internal/repository/archer.go` |
| Modify | `backend/internal/repository/archer_test.go` |

## Reference

- [PRD.md](../PRD.md)
- [base.go](../../../backend/internal/repository/base.go)
- [archer.go](../../../backend/internal/repository/archer.go)

## Steps

- [ ] **Step 1: Write failing query building tests for Bow and Arrow repos**

  Create `bow_test.go` and `arrow_test.go` verifying that `Create`, `FindByID`, `FindAllByArcherID`,
  and `CountByArcherID` generate valid SQL with Dollar placeholder format.

- [ ] **Step 2: Run tests to verify failure**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [ ] **Step 3: Implement `bow.go`**

  Implement `BowRepo` backed by `DBTX`. Write `FindAllByArcherID` ordered by `created_at ASC`, and
  `CountByArcherID` returning total active bows.

- [ ] **Step 4: Implement `arrow.go`**

  Implement `ArrowRepo`. Support single insertion and `CreateBatch` which inserts multiple rows
  in a single statement or transaction with `status` (defaulting to `'in_use'`). Implement
  `CountByArcherID`, `CountInUseByArcherID`, and `FindByArcherAndSet`.

- [ ] **Step 5: Implement `auth_identity.go`**

  Implement repository for the `auth` table. Provide lookup by `google_subject` and creation of new
  records when a user first signs in with Google OAuth.

- [ ] **Step 6: Update `archer.go`**

  Update column list in `archer.go` to strictly match personal fields:

  ```go
  var archerColumns = []string{
      "archer_id",
      "email",
      "first_name",
      "last_name",
      "date_of_birth",
      "gender",
      "is_deleted",
  }
  ```

  Update `scanArcher` and SQL generation for `Create`, `Update`, `FindByID`, and `FindByEmail`.

- [ ] **Step 7: Run repository unit tests**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  golangci-lint run ./internal/repository/...
  ```

- [ ] **Step 8: Commit changes**

  ```bash
  git add backend/internal/repository
  git commit -m "feat(repo): add bow, arrow, auth_identity repositories and update archer repo"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` exits with code 0.
- `cd backend && golangci-lint run ./internal/repository/...` passes with zero issues.
