# Task 004: Set Up pgxpool Database Connection Pool

## Git Branch

`refactor/004-database-connection-pool`

## Objective

Set up the PostgreSQL connection pool using `pgxpool` in the application entry point. Create a
reusable database initialization function that accepts a `Config` and returns a `*pgxpool.Pool`.
Wire this into `main.go` with graceful shutdown on SIGINT/SIGTERM.

## Dependencies

- Task 003 (config package with `DatabaseURL()` method)

## Acceptance Criteria

- [x] `backend/internal/repository/pool.go` provides a `NewPool(ctx, cfg) (*pgxpool.Pool, error)`
  function that:
    - Parses the connection string from `cfg.DatabaseURL()`
    - Configures min/max connections from config
    - Returns a connected pool or an error
- [x] `main.go` is updated to:
    - Load config via `config.Load()`
    - Create the pool via `repository.NewPool()`
    - Defer `pool.Close()`
    - Listen for OS signals (SIGINT, SIGTERM) for graceful shutdown
    - Log pool connection status
- [x] Unit test for `NewPool` configuration parsing exists (does not require a real DB).
- [x] `go build ./cmd/arch-stats` compiles cleanly.
- [x] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/repository/pool.go` |
| Create | `backend/internal/repository/pool_test.go` |
| Modify | `backend/cmd/arch-stats/main.go` |
| Delete | `backend/internal/repository/.gitkeep` |
| Modify | `backend/go.mod` (add pgx dependency) |

## Steps

- [x] **Step 1: Add pgx dependency**

  ```bash
  cd backend
  go get github.com/jackc/pgx/v5
  go get github.com/jackc/pgx/v5/pgxpool
  ```

- [x] **Step 2: Write the failing test**

  Create `backend/internal/repository/pool_test.go`:

  ```go
  package repository_test

  import (
      "testing"

      "github.com/jpmolinamatute/arch-stats/backend/internal/repository"
  )

  func TestParsePoolConfig(t *testing.T) {
      dsn := "postgresql://testuser:testpass@localhost:5432/testdb?sslmode=disable"
      poolCfg, err := repository.ParsePoolConfig(dsn, 2, 10)
      if err != nil {
          t.Fatalf("ParsePoolConfig() error: %v", err)
      }

      if poolCfg.MinConns != 2 {
          t.Errorf("MinConns = %d, want 2", poolCfg.MinConns)
      }
      if poolCfg.MaxConns != 10 {
          t.Errorf("MaxConns = %d, want 10", poolCfg.MaxConns)
      }
  }

  func TestParsePoolConfig_InvalidDSN(t *testing.T) {
      _, err := repository.ParsePoolConfig("not-a-valid-dsn://", 1, 5)
      if err == nil {
          t.Error("ParsePoolConfig() should fail with invalid DSN")
      }
  }
  ```

- [x] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [x] **Step 4: Implement `pool.go`**

  Create `backend/internal/repository/pool.go`:

  ```go
  // Package repository provides database access functions.
  package repository

  import (
      "context"
      "fmt"

      "github.com/jackc/pgx/v5/pgxpool"
  )

  // ParsePoolConfig parses a DSN and returns a pgxpool.Config with the
  // specified min/max connection counts.
  func ParsePoolConfig(dsn string, minConns, maxConns int32) (*pgxpool.Config, error) {
      cfg, err := pgxpool.ParseConfig(dsn)
      if err != nil {
          return nil, fmt.Errorf("parsing pool config: %w", err)
      }
      cfg.MinConns = minConns
      cfg.MaxConns = maxConns
      return cfg, nil
  }

  // NewPool creates a new pgxpool.Pool from the given DSN and pool size parameters.
  // The caller is responsible for calling pool.Close() when done.
  func NewPool(ctx context.Context, dsn string, minConns, maxConns int32) (*pgxpool.Pool, error) {
      poolCfg, err := ParsePoolConfig(dsn, minConns, maxConns)
      if err != nil {
          return nil, err
      }

      pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
      if err != nil {
          return nil, fmt.Errorf("connecting to database: %w", err)
      }

      if err := pool.Ping(ctx); err != nil {
          pool.Close()
          return nil, fmt.Errorf("pinging database: %w", err)
      }

      return pool, nil
  }
  ```

- [x] **Step 5: Update `main.go` with pool initialization and graceful shutdown**

  Update `backend/cmd/arch-stats/main.go` to:
    - Load config
    - Build DSN from config
    - Create pool
    - Set up OS signal handling for graceful shutdown
    - Defer pool.Close()

- [x] **Step 6: Run tests**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

  Expected: all tests PASS.

- [x] **Step 7: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./cmd/arch-stats
  ```

- [x] **Step 8: Commit**

  ```bash
  rm -f backend/internal/repository/.gitkeep
  git add -A
  git commit -m "feat: add pgxpool connection pool with config and graceful shutdown"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./cmd/arch-stats` — compiles without errors.
- Running the binary without a DB should print a clear connection error (not panic).
