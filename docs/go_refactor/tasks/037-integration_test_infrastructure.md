# Task 037: Integration Test Infrastructure with testcontainers-go

## Git Branch

`refactor/037-integration-test-infrastructure`

## Objective

Set up the integration test infrastructure using `testcontainers-go` to spin up a real
PostgreSQL container per test run. Create shared test helpers, fixtures, and a `TestMain`
function that boots a PostgreSQL container, runs goose migrations, and provides a clean
`pgxpool.Pool` to all integration tests. This mirrors the Python `conftest.py` approach
(real DB, truncate between tests).

## Dependencies

- Task 004 (database connection pool — `pgxpool`)
- Task 005 (goose migrations — migration files + runner)
- Task 008 (repository base — `DBTX` interface)

## Acceptance Criteria

- [ ] `backend/tests/integration/testmain_test.go` provides a `TestMain(m)` function that:
    - Starts a PostgreSQL 17 container via `testcontainers-go`
    - Waits for the container to be healthy
    - Creates a `pgxpool.Pool` connected to the container
    - Runs goose migrations against the test database
    - Exposes the pool to all tests via a package-level variable
    - Tears down the container after all tests complete
- [ ] `backend/tests/integration/helpers_test.go` provides shared utilities:
    - `truncateAll(ctx, pool)` — truncates all tables in FK-safe order (matching Python's
    `_truncate_all`)
    - `createTestArcher(ctx, pool, overrides)` — inserts a test archer and returns the record
    - `createTestSession(ctx, pool, archerID, overrides)` — inserts a test shooting session
    - `jwtForArcher(archerID, secret)` — generates a valid JWT for test requests
- [ ] `backend/tests/integration/smoke_test.go` contains a single smoke test that:
    - Verifies the pool is connected
    - Verifies migrations ran (tables exist)
    - Verifies `truncateAll` works without error
- [ ] `go test ./tests/integration/... -v` passes (smoke test connects, migrates, truncates).
- [ ] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/tests/integration/testmain_test.go` |
| Create | `backend/tests/integration/helpers_test.go` |
| Create | `backend/tests/integration/smoke_test.go` |
| Modify | `backend/go.mod` (add testcontainers-go dependency) |

## Reference

- Python test fixtures: [conftest.py](file:///home/juanpa/Projects/arch-stats/backend/tests/conftest.py)
- The truncation order from Python: `shot → slot → target → session → archer` (FK order)

## Steps

- [ ] **Step 1: Add testcontainers-go dependency**

  ```bash
  cd backend
  go get github.com/testcontainers/testcontainers-go
  go get github.com/testcontainers/testcontainers-go/modules/postgres
  ```

- [ ] **Step 2: Write `testmain_test.go`**

  Create `backend/tests/integration/testmain_test.go`:

  ```go
  package integration_test

  import (
      "context"
      "fmt"
      "os"
      "testing"
      "time"

      "github.com/jackc/pgx/v5/pgxpool"
      "github.com/testcontainers/testcontainers-go"
      "github.com/testcontainers/testcontainers-go/modules/postgres"
      "github.com/testcontainers/testcontainers-go/wait"

      "github.com/jpmolinamatute/arch-stats/backend/internal/repository"
  )

  var testPool *pgxpool.Pool

  func TestMain(m *testing.M) {
      ctx := context.Background()

      pgContainer, err := postgres.Run(ctx, "postgres:17",
          postgres.WithDatabase("arch-stats-test"),
          postgres.WithUsername("testuser"),
          postgres.WithPassword("testpass"),
          testcontainers.WithWaitStrategy(
              wait.ForLog("database system is ready to accept connections").
                  WithOccurrence(2).
                  WithStartupTimeout(30*time.Second),
          ),
      )
      if err != nil {
          fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
          os.Exit(1)
      }

      dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
      if err != nil {
          fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
          os.Exit(1)
      }

      testPool, err = repository.NewPool(ctx, dsn, 2, 5)
      if err != nil {
          fmt.Fprintf(os.Stderr, "failed to create pool: %v\n", err)
          os.Exit(1)
      }

      // Run goose migrations
      migrationsDir := "../../migrations"
      if err := repository.RunMigrations(ctx, testPool, migrationsDir); err != nil {
          fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
          os.Exit(1)
      }

      code := m.Run()

      testPool.Close()
      _ = pgContainer.Terminate(ctx)
      os.Exit(code)
  }
  ```

- [ ] **Step 3: Write `helpers_test.go`**

  Create `backend/tests/integration/helpers_test.go` with:
    - `truncateAll(ctx, pool)` executing TRUNCATE in FK-safe order
    - `createTestArcher(ctx, pool)` inserting a default archer via the repository
    - `createTestSession(ctx, pool, archerID)` inserting a default session
    - `jwtForArcher(archerID)` generating a test JWT

  ```go
  package integration_test

  import (
      "context"
      "time"

      "github.com/google/uuid"
      "github.com/jackc/pgx/v5/pgxpool"

      "github.com/jpmolinamatute/arch-stats/backend/internal/auth"
  )

  func truncateAll(ctx context.Context, pool *pgxpool.Pool) error {
      tables := []string{
          "shot",
          "slot",
          "target",
          "session",
          "authentication_session",
          "archer",
      }
      for _, table := range tables {
          if _, err := pool.Exec(ctx, "TRUNCATE "+table+" RESTART IDENTITY CASCADE"); err != nil {
              return err
          }
      }
      return nil
  }

  func jwtForArcher(archerID uuid.UUID, secret string) string {
      now := time.Now()
      token, _ := auth.BuildJWT(archerID, "test-sid", now, now.Add(time.Hour), secret, "HS256")
      return token
  }
  ```

- [ ] **Step 4: Write `smoke_test.go`**

  Create `backend/tests/integration/smoke_test.go`:

  ```go
  package integration_test

  import (
      "context"
      "testing"
  )

  func TestSmoke_PoolConnected(t *testing.T) {
      ctx := context.Background()
      if err := testPool.Ping(ctx); err != nil {
          t.Fatalf("pool not connected: %v", err)
      }
  }

  func TestSmoke_TablesExist(t *testing.T) {
      ctx := context.Background()
      var count int
      err := testPool.QueryRow(ctx,
          "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'archer'",
      ).Scan(&count)
      if err != nil {
          t.Fatalf("query failed: %v", err)
      }
      if count != 1 {
          t.Fatalf("archer table not found, count = %d", count)
      }
  }

  func TestSmoke_TruncateAll(t *testing.T) {
      ctx := context.Background()
      if err := truncateAll(ctx, testPool); err != nil {
          t.Fatalf("truncateAll failed: %v", err)
      }
  }
  ```

- [ ] **Step 5: Run tests**

  ```bash
  cd backend
  go test ./tests/integration/... -v -count=1
  ```

  Expected: all 3 smoke tests pass (container boots, migrations run, truncate works).

- [ ] **Step 6: Run go vet**

  ```bash
  cd backend
  go vet ./...
  ```

- [ ] **Step 7: Commit**

  ```bash
  git add -A
  git commit -m "feat: add integration test infrastructure with testcontainers-go"
  ```

## Verification

- `cd backend && go test ./tests/integration/... -v -count=1` — all smoke tests pass.
- `cd backend && go vet ./...` — clean.
- Docker must be running for testcontainers to work.
