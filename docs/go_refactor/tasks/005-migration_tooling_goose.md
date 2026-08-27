# Task 005: Integrate goose for Database Migrations

## Git Branch

`refactor/005-migration-tooling-goose`

## Objective

Replace Flyway with `pressly/goose` as the migration tool. Convert existing SQL migration files
from Flyway naming convention to goose format (add `-- +goose Up` / `-- +goose Down` markers).
Embed goose as a Go library so migrations can run programmatically on startup or via a CLI
subcommand (`arch-stats migrate`).

## Dependencies

- Task 004 (database connection pool)

## Acceptance Criteria

- [ ] All existing SQL migration files from `backend-old/migrations/` are converted to goose
  format under `backend/migrations/`:
    - `V001__2025-09-26_db_init.sql` → `001_db_init.sql`
    - `V002__2025-09-26_archers_table.sql` → `002_archers_table.sql`
    - `V003__2025-09-26_authentication_session_table.sql` → `003_authentication_session_table.sql`
    - `V004__2025-09-26_shooting_sessions_table.sql` → `004_shooting_sessions_table.sql`
    - `V005__2025-10-28_arrow_table.sql` → `005_arrow_table.sql`
    - `V006__2025-10-28_shot_table.sql` → `006_shot_table.sql`
- [ ] Each migration file has `-- +goose Up` and `-- +goose Down` markers.
- [ ] `backend/internal/repository/migrate.go` provides a `RunMigrations(ctx, pool, migrationsDir)`
  function that runs goose migrations using the pool's connection.
- [ ] `main.go` calls `RunMigrations()` on startup when `config.ApplyMigrationsOnStart` is true.
- [ ] A `migrate` subcommand is available: `arch-stats migrate` runs migrations independently.
- [ ] `go build ./cmd/arch-stats` compiles cleanly.
- [ ] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/migrations/001_db_init.sql` |
| Create | `backend/migrations/002_archers_table.sql` |
| Create | `backend/migrations/003_authentication_session_table.sql` |
| Create | `backend/migrations/004_shooting_sessions_table.sql` |
| Create | `backend/migrations/005_arrow_table.sql` |
| Create | `backend/migrations/006_shot_table.sql` |
| Create | `backend/internal/repository/migrate.go` |
| Create | `backend/internal/repository/migrate_test.go` |
| Modify | `backend/cmd/arch-stats/main.go` |
| Modify | `backend/go.mod` (add goose dependency) |

## Reference

Existing Flyway migration files are in
[backend-old/migrations/](file:///home/juanpa/Projects/arch-stats/backend-old/migrations/).

## Steps

- [ ] **Step 1: Add goose dependency**

  ```bash
  cd backend
  go get github.com/pressly/goose/v3
  ```

- [ ] **Step 2: Copy and convert migration files**

  For each migration file in `backend-old/migrations/`:
  1. Copy to `backend/migrations/` with goose naming convention (NNN_description.sql).
  2. Add `-- +goose Up` marker before the existing SQL.
  3. Add `-- +goose Down` marker with appropriate DROP/rollback statements.

  Example for `001_db_init.sql`:

  ```sql
  -- +goose Up
  -- Original Flyway: V001__2025-09-26_db_init.sql
  <existing SQL content>

  -- +goose Down
  -- Rollback for db_init
  ```

- [ ] **Step 3: Write the migration runner**

  Create `backend/internal/repository/migrate.go`:

  ```go
  package repository

  import (
      "context"
      "fmt"

      "github.com/jackc/pgx/v5/pgxpool"
      "github.com/jackc/pgx/v5/stdlib"
      "github.com/pressly/goose/v3"
  )

  // RunMigrations applies pending database migrations from the given directory.
  func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
      db := stdlib.OpenDBFromPool(pool)
      defer db.Close()

      goose.SetDialect("postgres")

      if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
          return fmt.Errorf("running migrations: %w", err)
      }

      return nil
  }
  ```

- [ ] **Step 4: Write a unit test for migration file validation**

  Create `backend/internal/repository/migrate_test.go` that:
    - Verifies all migration files in `../../migrations/` exist
    - Verifies each file contains `-- +goose Up` marker
    - Verifies filenames follow the NNN_name.sql pattern

  ```go
  package repository_test

  import (
      "os"
      "path/filepath"
      "regexp"
      "strings"
      "testing"
  )

  func TestMigrationFilesExist(t *testing.T) {
      migrationsDir := filepath.Join("..", "..", "migrations")
      entries, err := os.ReadDir(migrationsDir)
      if err != nil {
          t.Fatalf("reading migrations dir: %v", err)
      }

      if len(entries) == 0 {
          t.Fatal("migrations directory is empty")
      }

      pattern := regexp.MustCompile(`^\d{3}_\w+\.sql$`)
      for _, entry := range entries {
          if entry.IsDir() {
              continue
          }
          name := entry.Name()
          if !pattern.MatchString(name) {
              t.Errorf("migration file %q does not match pattern NNN_name.sql", name)
          }

          content, err := os.ReadFile(filepath.Join(migrationsDir, name))
          if err != nil {
              t.Errorf("reading %s: %v", name, err)
              continue
          }
          if !strings.Contains(string(content), "-- +goose Up") {
              t.Errorf("migration file %s missing '-- +goose Up' marker", name)
          }
      }
  }
  ```

- [ ] **Step 5: Update `main.go` with migrate subcommand and startup migration**

  Add logic to `main.go`:
    - If `os.Args` contains `migrate`, run migrations and exit.
    - Otherwise, if `config.ApplyMigrationsOnStart`, run migrations before starting the server.

- [ ] **Step 6: Run tests**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [ ] **Step 7: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./cmd/arch-stats
  ```

- [ ] **Step 8: Commit**

  ```bash
  git add -A
  git commit -m "feat: integrate goose migrations with startup and CLI subcommand"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` — migration file tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./cmd/arch-stats` — compiles without errors.
- `ls backend/migrations/*.sql` — all 6 migration files exist with goose markers.
