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

- [x] New repository `backend/internal/repository/bow.go` implements:
    - [x] `Create(ctx context.Context, data model.BowCreate) (uuid.UUID, error)`
    - [x] `FindByID(ctx context.Context, id uuid.UUID) (*model.BowRead, error)`
    - [x] `FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.BowRead, error)`
    - [x] `CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)`
- [x] New repository `backend/internal/repository/arrow.go` implements:
    - [x] `Create(ctx context.Context, data model.ArrowCreate) (uuid.UUID, error)` (inserts `status`
          , default `'in_use'`)
    - [x] `CreateBatch(ctx context.Context, data model.ArrowBatchCreate) ([]model.ArrowRead, error)`
          (inserts `status`, default `'in_use'`)
    - [x] `FindByID(ctx context.Context, id uuid.UUID) (*model.ArrowRead, error)`
    - [x] `FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.ArrowRead, error)`
    - [x] `CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)`
    - [x] `CountInUseByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)` (for arrow
          ceiling)
    - [x] `FindByArcherAndSet(ctx context.Context, archerID uuid.UUID, set int16)`
          `([]model.ArrowRead, error)`
- [x] New repository `backend/internal/repository/auth_identity.go` implements:
    - [x] `FindByGoogleSubject(ctx context.Context, sub string) (*model.AuthIdentityRead, error)`
    - [x] `FindByID(ctx context.Context, id uuid.UUID) (*model.AuthIdentityRead, error)`
    - [x] `Create(ctx context.Context, subject string, picture *string) (uuid.UUID, error)`
    - [x] `UpdateLastLogin(ctx context.Context, id uuid.UUID, login time.Time, pic *string) error`
- [x] Modified `backend/internal/repository/archer.go`:
    - [x] Column selection updated: `archer_id`, `email`, `first_name`, `last_name`,
          `date_of_birth`, `gender`, `is_deleted`.
    - [x] Scanning functions updated to omit legacy equipment/Google columns.
    - [x] `Create` accepts `model.ArcherCreate` with `ArcherID` provided from `auth`.
- [x] Repositories support transactions via `WithTx(tx pgx.Tx)`.
- [x] Unit tests for all repositories using mock `DBTX` verify SQL query and argument construction.
- [x] `cd backend && go test ./internal/repository/... -v` passes.
- [x] `cd backend && golangci-lint run ./internal/repository/...` reports no issues.

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

- [x] **Step 1: Write failing query building tests for Bow and Arrow repos**

  Create `bow_test.go` and `arrow_test.go` verifying that `Create`, `FindByID`, `FindAllByArcherID`,
  and `CountByArcherID` generate valid SQL with Dollar placeholder format.

- [x] **Step 2: Run tests to verify failure**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [x] **Step 3: Implement `bow.go`**

  Implement `BowRepo` backed by `DBTX`. Write `FindAllByArcherID` ordered by `created_at ASC`, and
  `CountByArcherID` returning total active bows.

- [x] **Step 4: Implement `arrow.go`**

  Implement `ArrowRepo`. Support single insertion and `CreateBatch` which inserts multiple rows
  in a single statement or transaction with `status` (defaulting to `'in_use'`). Implement
  `CountByArcherID`, `CountInUseByArcherID`, and `FindByArcherAndSet`.

- [x] **Step 5: Implement `auth_identity.go`**

  Implement repository for the `auth` table. Provide lookup by `google_subject` and creation of new
  records when a user first signs in with Google OAuth.

- [x] **Step 6: Update `archer.go`**

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

- [x] **Step 7: Run repository unit tests**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  golangci-lint run ./internal/repository/...
  ```

- [x] **Step 8: Commit changes**

  ```bash
  git add backend/internal/repository
  git commit -m "feat(repo): add bow, arrow, auth_identity repositories and update archer repo"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` exits with code 0.
- `cd backend && golangci-lint run ./internal/repository/...` passes with zero issues.
