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

- [x] `backend/internal/repository/base.go` defines:
    - A `DBTX` interface abstracting `pgxpool.Pool` for testability (methods: `Query`, `QueryRow`,
    `Exec`)
    - Common helper functions for scanning rows into structs
    - A `WithTx(ctx, pool, fn)` helper function that begins a transaction, executes `fn(tx)`,
      and commits or rolls back. This enables multi-repository operations to run atomically
      (e.g., creating a session + target + slot in one transaction). The `fn` callback receives
      a `pgx.Tx` which satisfies the `DBTX` interface, so repositories can accept either a pool
      or a transaction transparently.
- [x] `backend/internal/repository/archer.go` implements `ArcherRepo` with methods:
    - `FindByID(ctx, id) (*model.ArcherRead, error)`
    - `FindByEmail(ctx, email) (*model.ArcherRead, error)`
    - `FindByGoogleSubject(ctx, sub) (*model.ArcherRead, error)`
    - `FindAll(ctx, filter) ([]model.ArcherRead, error)`
    - `Create(ctx, data) (uuid.UUID, error)`
    - `Update(ctx, data, filter) error`
    - `Delete(ctx, id) error`
- [x] `backend/internal/repository/maintenance.go` implements `MaintenanceRepo` with methods:
    - `RefreshOpenParticipants(ctx) error` — executes
      `REFRESH MATERIALIZED VIEW CONCURRENTLY open_participants`. This must be called after
      slot creation/update/deletion and after session close.
    - `GetSchemaVersion(ctx) (int64, error)` — reads the current goose migration version from
      `goose_db_version` for startup logging and health checks.
- [x] `backend/internal/repository/reporting.go` implements `ReportingRepo` with read-only
  methods for cross-domain analytics queries. Initial methods (can be stubs that return
  placeholder data for now, to be fleshed out when reporting features are built):
    - `GetSessionSummary(ctx, sessionID) (*model.SessionSummaryReport, error)`
    - `GetArcherPerformance(ctx, archerID, from, to) ([]model.ScoringTrend, error)`
  These queries may use raw SQL (not squirrel) when the join/aggregation complexity warrants it.
- [x] All queries are built with `squirrel` (not raw SQL strings), except for `ReportingRepo`
  and `MaintenanceRepo` where raw SQL is acceptable for complex analytical queries and DDL.
- [x] Unit tests with a mock DBTX interface verify query building logic.
- [x] `go test ./internal/repository/...` passes.
- [x] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/repository/base.go` |
| Create | `backend/internal/repository/archer.go` |
| Create | `backend/internal/repository/archer_test.go` |
| Create | `backend/internal/repository/maintenance.go` |
| Create | `backend/internal/repository/maintenance_test.go` |
| Create | `backend/internal/repository/reporting.go` |
| Create | `backend/internal/repository/reporting_test.go` |
| Modify | `backend/go.mod` (add squirrel + uuid dependencies) |

## Reference

- Python base model: [parent_model.py](file:///home/juanpa/Projects/arch-stats/backend/src/models/parent_model.py)
- Python archer model: [archer_model.py](file:///home/juanpa/Projects/arch-stats/backend/src/models/archer_model.py)

## Steps

- [x] **Step 1: Add dependencies**

  ```bash
  cd backend
  go get github.com/Masterminds/squirrel
  go get github.com/google/uuid
  ```

- [x] **Step 2: Write failing tests for ArcherRepo query building**

  Create `backend/internal/repository/archer_test.go`:
    - Test that `FindByID` builds the correct SQL (use squirrel's `ToSql()` to verify)
    - Test that `Create` builds a valid INSERT statement
    - Test that `Update` builds a valid UPDATE with WHERE clause
    - Test that `FindAll` with filters builds correct WHERE clauses

- [x] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [x] **Step 4: Implement `base.go`**

  Define the `DBTX` interface and common helpers:

  ```go
  package repository

  import (
      "context"

      "github.com/jackc/pgx/v5"
      "github.com/jackc/pgx/v5/pgconn"
      "github.com/jackc/pgx/v5/pgxpool"
  )

  // DBTX abstracts the pgxpool.Pool interface for testability.
   type DBTX interface {
       Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
       QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
       Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
   }

   // WithTx wraps a function in a database transaction. The callback receives
   // a pgx.Tx which also satisfies the DBTX interface, so any repository method
   // that accepts DBTX can participate in the transaction transparently.
   func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
       tx, err := pool.Begin(ctx)
       if err != nil {
           return fmt.Errorf("beginning transaction: %w", err)
       }
       defer tx.Rollback(ctx) // no-op if already committed

       if err := fn(tx); err != nil {
           return err
       }
       return tx.Commit(ctx)
   }
  ```

- [x] **Step 5: Implement `archer.go`**

  Implement `ArcherRepo` struct with all CRUD methods using squirrel.
  Use `squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)` for PostgreSQL.

- [x] **Step 5b: Implement `maintenance.go`**

  Implement `MaintenanceRepo` with:
    - `RefreshOpenParticipants(ctx)` using `REFRESH MATERIALIZED VIEW CONCURRENTLY open_participants`.
    - `GetSchemaVersion(ctx)` using `goose.GetDBVersion()` or a direct query on `goose_db_version`.

- [x] **Step 5c: Implement `reporting.go`**

  Implement `ReportingRepo` with initial stub methods. These can return placeholder data
  or `apperror.ErrNotImplemented` for now. The key is establishing the pattern:
    - Separate from entity CRUD repos.
    - Accepts `DBTX` interface.
    - May use raw SQL for complex joins/aggregations instead of squirrel.

- [x] **Step 6: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [x] **Step 7: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [x] **Step 8: Commit**

  ```bash
  git add -A
  git commit -m "feat: add repository base patterns and archer repository with squirrel"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
