# Task 008: Build Repository Layer — Base Patterns + Archer Repository

## Git Branch

`refactor/008-repository-base-and-archer`

## Objective

Build the foundation of the repository layer: define common repository interfaces/patterns and
implement the first concrete repository (Archer). Use `squirrel` for SQL query building and `pgx`
for database access, following the patterns established in the Python `parent_model.py`.

## Dependencies

- Task 004 (database connection pool)
- Task 007 (domain model structs)

## Acceptance Criteria

- [ ] `backend/internal/repository/base.go` defines:
    - A `DBTX` interface abstracting `pgxpool.Pool` for testability (methods: `Query`, `QueryRow`,
    `Exec`)
    - Common helper functions for scanning rows into structs
- [ ] `backend/internal/repository/archer.go` implements `ArcherRepo` with methods:
    - `FindByID(ctx, id) (*model.ArcherRead, error)`
    - `FindByEmail(ctx, email) (*model.ArcherRead, error)`
    - `FindByGoogleSubject(ctx, sub) (*model.ArcherRead, error)`
    - `FindAll(ctx, filter) ([]model.ArcherRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)`
    - `Update(ctx, data, filter) error`
    - `Delete(ctx, id) error`
- [ ] All queries are built with `squirrel` (not raw SQL strings).
- [ ] Unit tests with a mock DBTX interface verify query building logic.
- [ ] `go test ./internal/repository/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/repository/base.go` |
| Create | `backend/internal/repository/archer.go` |
| Create | `backend/internal/repository/archer_test.go` |
| Modify | `backend/go.mod` (add squirrel + uuid dependencies) |

## Reference

- Python base model: [parent_model.py](file:///home/juanpa/Projects/arch-stats/backend/src/models/parent_model.py)
- Python archer model: [archer_model.py](file:///home/juanpa/Projects/arch-stats/backend/src/models/archer_model.py)

## Steps

- [ ] **Step 1: Add dependencies**

  ```bash
  cd backend
  go get github.com/Masterminds/squirrel
  go get github.com/google/uuid
  ```

- [ ] **Step 2: Write failing tests for ArcherRepo query building**

  Create `backend/internal/repository/archer_test.go`:
    - Test that `FindByID` builds the correct SQL (use squirrel's `ToSql()` to verify)
    - Test that `Create` builds a valid INSERT statement
    - Test that `Update` builds a valid UPDATE with WHERE clause
    - Test that `FindAll` with filters builds correct WHERE clauses

- [ ] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [ ] **Step 4: Implement `base.go`**

  Define the `DBTX` interface and common helpers:

  ```go
  package repository

  import (
      "context"

      "github.com/jackc/pgx/v5"
      "github.com/jackc/pgx/v5/pgconn"
  )

  // DBTX abstracts the pgxpool.Pool interface for testability.
  type DBTX interface {
      Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
      QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
      Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
  }
  ```

- [ ] **Step 5: Implement `archer.go`**

  Implement `ArcherRepo` struct with all CRUD methods using squirrel.
  Use `squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)` for PostgreSQL.

- [ ] **Step 6: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [ ] **Step 7: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [ ] **Step 8: Commit**

  ```bash
  git add -A
  git commit -m "feat: add repository base patterns and archer repository with squirrel"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
