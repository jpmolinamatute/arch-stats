# Task 005: Integrate goose for Database Migrations (Two-Repository Architecture)

## Git Branches

- **Migrations Repository (`arch-stats-migrations`)**: `feature/goose-migration-format`
- **Main Application Repository (`arch-stats`)**: `refactor/005-migration-tooling-goose`

## Objective

Replace Flyway with `pressly/goose/v3` as the database migration tool.

The SQL migration files (`*.sql`) live in a separate private git repository
([git@github.com:jpmolinamatute/arch-stats-migrations.git](file:///home/juanpa/Projects/arch-stats/backend-old/migrations)),
which is mounted as a git submodule in the application (`backend/migrations/`,previously
`backend-old/migrations/`).

This separation must be preserved. Consequently, this task requires **at least two separate commits
across two repositories**:

1. **Commit 1 (in `arch-stats-migrations` repository)**: Convert existing SQL migration files from
   Flyway format (`V00X__...sql`) to goose format (`00X_...sql`), add `-- +goose Up` and
   `-- +goose Down` rollback annotations, and update `scripts/run_migration_tests.bash` to validate
   migration application and rollback behavior.
2. **Commit 2 (in `arch-stats` main repository)**: Update the git submodule path to
   `backend/migrations/` and stage the updated submodule commit SHA, embed goose in the Go backend
   (`backend/internal/repository/migrate.go`), provide unit tests (`migrate_test.go`), and add
   startup auto-migration and a CLI subcommand (`arch-stats migrate`) in
   `backend/cmd/arch-stats/main.go`.

## Dependencies

- Task 004 (database connection pool via `pgxpool`)

## Repository Architecture & Two-Commit Workflow

```text
arch-stats (Main Repo)
├── .gitmodules                     --> Tracks backend/migrations submodule
└── backend/
    ├── cmd/arch-stats/main.go      --> Handles 'migrate' CLI subcommand & startup flag
    ├── go.mod                      --> Dependency: github.com/pressly/goose/v3
    ├── internal/repository/
    │   ├── migrate.go              --> RunMigrations(ctx, pool, migrationsDir)
    │   └── migrate_test.go         --> Validates migration files presence & markers
    └── migrations/ # arch-stats-migrations (Submodule Repo)
        ├── 001_db_init.sql                 --> goose Up/Down
        ├── 002_archers_table.sql           --> goose Up/Down
        ├── 003_authentication_session_table.sql --> goose Up/Down
        ├── 004_shooting_sessions_table.sql --> goose Up/Down
        ├── 005_arrow_table.sql             --> goose Up/Down
        ├── 006_shot_table.sql              --> goose Up/Down
        ├── README.md                       --> goose CLI documentation
        └── scripts/
            └── run_migration_tests.bash    --> Schema and migration regression tests
```

## Acceptance Criteria

### Repository 1: `arch-stats-migrations` (Submodule)

- [x] All existing SQL migration files are renamed from Flyway format to goose format:
    - `V001__2025-09-26_db_init.sql` → `001_2025-09-26_db_init.sql`
    - `V002__2025-09-26_archers_table.sql` → `002_2025-09-26_archers_table.sql`
    - `V003__2025-09-26_authentication_session_table.sql` → `003_2025-09-26_authentication_session_table.sql`
    - `V004__2025-09-26_shooting_sessions_table.sql` → `004_2025-09-26_shooting_sessions_table.sql`
    - `V005__2025-10-28_arrow_table.sql` → `005_2025-10-28_arrow_table.sql`
    - `V006__2025-10-28_shot_table.sql` → `006_2025-10-28_shot_table.sql`
- [x] Each migration file contains explicit `-- +goose Up` and `-- +goose Down` markers.
- [x] Every `-- +goose Down` block contains valid, reverse-dependency rollback statements (dropping
  tables, views, materialized views, triggers, functions, types, and extensions cleanly).
- [x] `scripts/run_migration_tests.bash` validates all 21+ schema test cases, function/view
  behaviors, triggers, and constraint checks against PostgreSQL.
- [x] `scripts/run_migration_tests.bash` includes verification for goose migration application (`up`)
  and rollback (`down`).
- [x] `scripts/run_migration_tests.bash` dynamically discovers all `*.sql` migration files rather
  than hardcoding file names or counts, so that newly added migrations are automatically validated.
- [x] `README.md` in the migrations repository is updated to document:
    - Goose usage (CLI commands, embedded usage).
    - **Schema evolution conventions**: how to author new migrations, including:
        - Naming convention: `NNN_YYYY-MM-DD_description.sql` with sequential numbering.
        - Mandatory `-- +goose Up` and `-- +goose Down` markers.
        - Rules for `Down` blocks: must be losslessly reversible or explicitly marked as destructive.
        - When to use `-- +goose StatementBegin` / `-- +goose StatementEnd` for multi-statement
          PL/pgSQL blocks.
        - Guidelines for **data migrations** (not just DDL): when and how to transform existing row
          data as columns change meaning or type.
        - Review process: all migration PRs must pass the full `up → test → down → up` cycle in CI.
- [x] Changes in `arch-stats-migrations` are committed to git.

### Repository 2: `arch-stats` (Main Application)

- [x] Git submodule configuration (`.gitmodules`) is updated so that `backend/migrations` tracks
  `git@github.com:jpmolinamatute/arch-stats-migrations.git` at the new commit SHA.
- [x] `backend/internal/repository/migrate.go` provides `RunMigrations(ctx context.Context, pool
  *pgxpool.Pool, migrationsDir string) error` using `pressly/goose/v3` and `pgx/v5/stdlib`.
- [x] `backend/internal/repository/migrate_test.go` provides unit tests that verify:
    - All migration files in `../../migrations/` exist.
    - Filenames match the goose naming pattern `^\d{3}_\w+\.sql$`.
    - Every file contains `-- +goose Up` and `-- +goose Down` markers.
- [x] `backend/cmd/arch-stats/main.go` supports:
    - Standalone migration subcommand: `arch-stats migrate` runs migrations and exits.
    - Startup migration check: if `config.ApplyMigrationsOnStart` is true, runs migrations before
      listening on HTTP.
    - After migrations run (or on startup if skipped), the application logs the current schema
      version via `goose.GetDBVersion(db)` using `slog.Info`.
- [x] `backend/go.mod` and `backend/go.sum` include `github.com/pressly/goose/v3`.
- [x] `go test ./internal/repository/... -v` passes.
- [x] `go vet ./...` reports no issues.
- [x] `go build ./cmd/arch-stats` compiles cleanly.
- [x] Changes in `arch-stats` are committed to the `refactor/005-migration-tooling-goose` branch.

## Files to Create / Modify

### In `arch-stats-migrations` Repository (`backend/migrations/`)

| Action | Path | Description |
| ------ | ---- | ----------- |
| Modify | `001_2025-09-26_db_init.sql` | Up: UUID extension; Down: drop extension |
| Modify | `002_2025-09-26_archers_table.sql` | Up: enum types & archer table; Down: drop table & enum types |
| Modify | `003_2025-09-26_authentication_session_table.sql` | Up: auth table & index; Down: drop index & auth table |
| Modify | `004_2025-09-26_shooting_sessions_table.sql` | Up: session, target, slot tables, views, functions; Down: drop all in reverse order |
| Modify | `005_2025-10-28_arrow_table.sql` | Up: arrow table; Down: drop arrow table |
| Modify | `006_2025-10-28_shot_table.sql` | Up: shot table, triggers, functions, view; Down: drop view, triggers, functions, shot table |
| Modify | `scripts/run_migration_tests.bash` | Add goose execution, migration up/down lifecycle tests, and verify all constraints |
| Modify | `README.md` | Update instructions from Flyway to goose |

### In `arch-stats` Repository (Main Application Root)

| Action | Path | Description |
| ------ | ---- | ----------- |
| Modify | `.gitmodules` | Point `backend/migrations` to `git@github.com:jpmolinamatute/arch-stats-migrations.git` |
| Create | `backend/internal/repository/migrate.go` | `RunMigrations()` implementation with goose |
| Create | `backend/internal/repository/migrate_test.go` | Unit tests for migration file structure and markers |
| Modify | `backend/cmd/arch-stats/main.go` | Add `migrate` CLI subcommand and startup migration hook |
| Modify | `backend/go.mod` | Add `github.com/pressly/goose/v3` dependency |

## Migration Files Specification

### 1. `001_2025-09-26_db_init.sql`

```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- +goose Down
DROP EXTENSION IF EXISTS "uuid-ossp";
```

### 2. `002_2025-09-26_archers_table.sql`

```sql
-- +goose Up
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bowstyle_type') THEN
        CREATE TYPE bowstyle_type AS ENUM (
            'recurve',
            'compound',
            'barebow',
            'longbow'
        );
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender_type') THEN
        CREATE TYPE gender_type AS ENUM (
            'male',
            'female',
            'non_binary',
            'other',
            'unspecified'
        );
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'face_type') THEN
        CREATE TYPE face_type AS ENUM (
            'wa_40cm_full',
            'wa_60cm_full',
            'wa_80cm_full',
            'wa_122cm_full',
            'wa_40cm_6rings',
            'wa_60cm_6rings',
            'wa_80cm_6rings',
            'wa_122cm_6rings',
            'wa_40cm_triple_vertical',
            'wa_60cm_triple_triangular',
            'none'
        );
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS archer (
    archer_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(100) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL CHECK (date_of_birth <= current_date),
    gender GENDER_TYPE NOT NULL,
    bowstyle BOWSTYLE_TYPE NOT NULL,
    draw_weight FLOAT NOT NULL CHECK (draw_weight > 0),
    club_id UUID,
    google_picture_url TEXT,
    google_subject TEXT NOT NULL UNIQUE,
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS archer CASCADE;
DROP TYPE IF EXISTS face_type CASCADE;
DROP TYPE IF EXISTS gender_type CASCADE;
DROP TYPE IF EXISTS bowstyle_type CASCADE;
```

### 3. `003_2025-09-26_authentication_session_table.sql`

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS auth (
    auth_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE CASCADE,
    session_token_hash BYTEA NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    ua TEXT,
    ip_inet INET
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at
ON auth (expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_auth_sessions_expires_at;
DROP TABLE IF EXISTS auth CASCADE;
```

### 4. `004_2025-09-26_shooting_sessions_table.sql`

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS session (
    session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at TIMESTAMPTZ,
    session_location VARCHAR(255) NOT NULL,
    is_indoor BOOLEAN NOT NULL,
    is_opened BOOLEAN NOT NULL,
    CONSTRAINT sessions_time_check CHECK (
        closed_at IS NULL OR closed_at > created_at
    )
);

CREATE TABLE IF NOT EXISTS target (
    target_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES session (session_id) ON DELETE CASCADE,
    distance INTEGER NOT NULL,
    lane INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT targets_one_per_session UNIQUE (session_id, lane),
    CONSTRAINT targets_lane_bounds CHECK (lane BETWEEN 1 AND 100),
    CONSTRAINT targets_distance_bounds CHECK (distance BETWEEN 1 AND 100)
);

CREATE TABLE IF NOT EXISTS slot (
    slot_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    target_id UUID NOT NULL REFERENCES target (target_id) ON DELETE CASCADE,
    archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE RESTRICT,
    session_id UUID NOT NULL REFERENCES session (session_id) ON DELETE CASCADE,
    interval_seconds INT NOT NULL DEFAULT 20,
    face_type FACE_TYPE NOT NULL,
    slot_letter CHAR(1) NOT NULL,
    is_shooting BOOLEAN NOT NULL,
    bowstyle BOWSTYLE_TYPE NOT NULL,
    draw_weight FLOAT NOT NULL CHECK (draw_weight > 0),
    club_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    shot_per_round INTEGER CHECK (shot_per_round >= 3 AND shot_per_round <= 10),
    CONSTRAINT uq_archer_per_session UNIQUE (archer_id, session_id),
    CONSTRAINT slot_letter_valid CHECK (slot_letter IN ('A', 'B', 'C', 'D')),
    CONSTRAINT uq_slot_on_target UNIQUE (target_id, slot_letter, archer_id),
    CONSTRAINT interval_seconds_check CHECK (interval_seconds >= 1 AND interval_seconds <= 100)
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_open_session_per_archer
ON session (owner_archer_id)
WHERE is_opened = TRUE;

CREATE OR REPLACE FUNCTION get_next_lane(p_session_id UUID)
RETURNS INTEGER AS $$
    SELECT COALESCE(MAX(lane), 0) + 1
    FROM target
    WHERE session_id = p_session_id
    AND EXISTS (
        SELECT 1
        FROM session
        WHERE session.session_id = p_session_id
        AND session.is_opened IS TRUE
    );
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION get_available_targets(p_session_id UUID, p_distance INTEGER)
RETURNS TABLE (
    target_id UUID,
    session_id UUID,
    distance INTEGER,
    lane INTEGER,
    occupied INTEGER,
    created_at TIMESTAMPTZ
) AS
$$
SELECT
    target.target_id,
    target.session_id,
    target.distance,
    target.lane,
    COUNT(slot.slot_letter) AS occupied,
    target.created_at
FROM target
LEFT JOIN slot
    ON target.target_id = slot.target_id
WHERE target.session_id = p_session_id
    AND target.distance = p_distance
GROUP BY target.target_id, target.session_id, target.distance, target.lane
HAVING COUNT(slot.slot_letter) < 4
ORDER BY target.lane ASC
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION get_slot_with_lane(p_slot_id UUID)
RETURNS TABLE (
    slot_id UUID,
    target_id UUID,
    archer_id UUID,
    session_id UUID,
    face_type FACE_TYPE,
    slot_letter CHAR(1),
    created_at TIMESTAMPTZ,
    is_shooting BOOLEAN,
    bowstyle BOWSTYLE_TYPE,
    draw_weight FLOAT,
    interval_seconds INTEGER,
    shot_per_round INTEGER,
    slot VARCHAR(4)
) AS
$$
SELECT
    slot.slot_id,
    slot.target_id,
    slot.archer_id,
    slot.session_id,
    slot.face_type,
    slot.slot_letter,
    slot.created_at,
    slot.is_shooting,
    slot.bowstyle,
    slot.draw_weight,
    slot.interval_seconds,
    slot.shot_per_round,
    CONCAT(target.lane, slot.slot_letter) AS slot
FROM slot
INNER JOIN target
    ON slot.target_id = target.target_id
WHERE slot.slot_id = p_slot_id
$$ LANGUAGE sql STABLE;

CREATE MATERIALIZED VIEW IF NOT EXISTS open_participants AS
SELECT
    slot.slot_id,
    slot.target_id,
    slot.archer_id,
    slot.session_id,
    slot.face_type,
    slot.slot_letter,
    slot.created_at,
    slot.is_shooting,
    slot.bowstyle,
    slot.draw_weight,
    slot.club_id,
    slot.interval_seconds,
    slot.shot_per_round,
    target.lane,
    target.distance,
    concat(target.lane, slot.slot_letter) AS slot
FROM slot
INNER JOIN session
    ON slot.session_id = session.session_id
INNER JOIN target
    ON slot.target_id = target.target_id
INNER JOIN archer
    ON slot.archer_id = archer.archer_id
WHERE
    session.is_opened IS TRUE
    AND slot.is_shooting IS TRUE
WITH NO DATA;

CREATE UNIQUE INDEX IF NOT EXISTS uq_open_participants_slot_id ON open_participants (slot_id);
CREATE INDEX IF NOT EXISTS idx_open_participants_session_id ON open_participants (session_id);
CREATE INDEX IF NOT EXISTS idx_open_participants_archer_id ON open_participants (archer_id);

REFRESH MATERIALIZED VIEW open_participants;

CREATE OR REPLACE FUNCTION get_active_slot_id(p_archer_id UUID)
RETURNS UUID AS $$
DECLARE
    v_slot_id UUID;
BEGIN
    SELECT op.slot_id INTO v_slot_id
    FROM open_participants op
    WHERE op.archer_id = p_archer_id
    LIMIT 1;
    RETURN v_slot_id;
END;
$$ LANGUAGE plpgsql STABLE;

-- +goose Down
DROP FUNCTION IF EXISTS get_active_slot_id(UUID);
DROP MATERIALIZED VIEW IF EXISTS open_participants CASCADE;
DROP FUNCTION IF EXISTS get_slot_with_lane(UUID);
DROP FUNCTION IF EXISTS get_available_targets(UUID, INTEGER);
DROP FUNCTION IF EXISTS get_next_lane(UUID);
DROP INDEX IF EXISTS unique_open_session_per_archer;
DROP TABLE IF EXISTS slot CASCADE;
DROP TABLE IF EXISTS target CASCADE;
DROP TABLE IF EXISTS session CASCADE;
```

### 5. `005_2025-10-28_arrow_table.sql`

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS arrow (
    arrow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE CASCADE,
    spine FLOAT,
    length FLOAT,
    weight FLOAT,
    arrow_number INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS arrow CASCADE;
```

### 6. `006_2025-10-28_shot_table.sql`

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS shot (
    shot_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slot_id UUID NOT NULL REFERENCES slot (slot_id) ON DELETE CASCADE,
    x DOUBLE PRECISION,
    y DOUBLE PRECISION,
    score INTEGER,
    arrow_id UUID REFERENCES arrow (arrow_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_x BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT shot_score_nonnegative CHECK (score >= 0 AND score <= 10),
    CONSTRAINT shot_is_x_requires_ten CHECK (
        NOT is_x OR coalesce(score = 10, FALSE)
    ),
    CONSTRAINT shot_coords_score_all_or_none CHECK (
        (x IS NULL AND y IS NULL AND score IS NULL)
        OR (x IS NOT NULL AND y IS NOT NULL AND score IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_shot_slot_id ON shot (slot_id);

CREATE OR REPLACE FUNCTION notify_shot_insert() RETURNS TRIGGER AS $$
DECLARE
    slot_record RECORD;
BEGIN
    FOR slot_record IN
        SELECT slot_id, json_agg(row_to_json(new_table)) as shots
        FROM new_table
        GROUP BY slot_id
    LOOP
        PERFORM pg_notify(
            'shot_insert_' || slot_record.slot_id::text,
            slot_record.shots::text
        );
    END LOOP;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS shot_insert ON shot;
CREATE TRIGGER shot_insert
AFTER INSERT ON shot
REFERENCING NEW TABLE AS new_table
FOR EACH STATEMENT EXECUTE FUNCTION notify_shot_insert();

CREATE OR REPLACE FUNCTION check_shot_arrow_ownership() RETURNS TRIGGER AS $$
DECLARE
    v_slot_archer_id UUID;
    v_session_is_opened BOOLEAN;
BEGIN
    IF NEW.arrow_id IS NULL THEN
        RETURN NEW;
    END IF;

    SELECT slot.archer_id, session.is_opened
    INTO v_slot_archer_id, v_session_is_opened
    FROM slot
    JOIN session ON session.session_id = slot.session_id
    WHERE slot.slot_id = NEW.slot_id;

    IF v_session_is_opened IS DISTINCT FROM TRUE THEN
        RAISE EXCEPTION 'Cannot record shot: session is not open for slot %', NEW.slot_id
        USING ERRCODE = '23514';
    END IF;

    PERFORM 1
    FROM arrow
    WHERE arrow.arrow_id = NEW.arrow_id
    AND arrow.archer_id = v_slot_archer_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Arrow % does not belong to the archer in slot %',
            NEW.arrow_id, NEW.slot_id
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS shot_arrow_ownership ON shot;
CREATE CONSTRAINT TRIGGER shot_arrow_ownership
AFTER INSERT OR UPDATE OF arrow_id, slot_id
ON shot
DEFERRABLE INITIALLY IMMEDIATE
FOR EACH ROW
WHEN (new.arrow_id IS NOT NULL)
EXECUTE FUNCTION check_shot_arrow_ownership();

CREATE OR REPLACE VIEW live_stat_by_slot_id AS
SELECT
    shot.slot_id,
    coalesce(avg(shot.score)::DOUBLE PRECISION, 0) AS mean,
    count(shot.slot_id) * 10 AS max_score,
    count(shot.slot_id) AS number_of_shots,
    count(*) FILTER (WHERE shot.is_x) AS number_of_x,
    coalesce(sum(shot.score), 0) AS total_score
FROM shot
GROUP BY shot.slot_id;

-- +goose Down
DROP VIEW IF EXISTS live_stat_by_slot_id;
DROP TRIGGER IF EXISTS shot_arrow_ownership ON shot;
DROP FUNCTION IF EXISTS check_shot_arrow_ownership();
DROP TRIGGER IF EXISTS shot_insert ON shot;
DROP FUNCTION IF EXISTS notify_shot_insert();
DROP INDEX IF EXISTS idx_shot_slot_id;
DROP TABLE IF EXISTS shot CASCADE;
```

## Detailed Steps

### Part 1: Migrations Repository Work (`arch-stats-migrations`)

- [x] **Step 1: Convert migration files to goose format**

  In the `backend/migrations/` (submodule) directory:
  1. Rename files from `V00X__...sql` to `00X_...sql`.
  2. Add `-- +goose Up` before existing creation DDL.
  3. Add `-- +goose Down` with appropriate drop/rollback DDL.

- [x] **Step 2: Update `scripts/run_migration_tests.bash`**

  Ensure the test script:
    - Dynamically discovers all `*.sql` migration files (no hardcoded file names or counts).
    - Applies goose migrations up.
    - Executes all constraint, trigger, view, and function regression tests.
    - Executes `-- +goose Down` rollback to verify clean teardown.
    - Re-applies `-- +goose Up` to ensure clean idempotency.
    - New migrations added in the future are automatically picked up without script changes.

- [x] **Step 3: Run migration tests**

  ```bash
  cd backend/migrations
  ./scripts/run_migration_tests.bash
  ```

- [x] **Step 4: Commit to `arch-stats-migrations`**

  ```bash
  cd backend/migrations
  git add 001_db_init.sql 002_archers_table.sql 003_authentication_session_table.sql \
          004_shooting_sessions_table.sql 005_arrow_table.sql 006_shot_table.sql \
          scripts/run_migration_tests.bash README.md
  git commit -m "feat: convert SQL migrations to goose format with Up and Down markers"
  git push origin HEAD
  ```

### Part 2: Main Application Repository Work (`arch-stats`)

- [x] **Step 5: Add goose dependency**

  ```bash
  cd backend
  go get github.com/pressly/goose/v3
  go get github.com/jackc/pgx/v5/stdlib
  ```

- [x] **Step 6: Update `.gitmodules` submodule configuration**

  Update `.gitmodules` so `backend/migrations` points to `git@github.com:jpmolinamatute/arch-stats-migrations.git`:

  ```ini
  [submodule "backend/migrations"]
    path = backend/migrations
    url = git@github.com:jpmolinamatute/arch-stats-migrations.git
  ```

- [x] **Step 7: Implement the migration runner (`backend/internal/repository/migrate.go`)**

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

      if err := goose.SetDialect("postgres"); err != nil {
          return fmt.Errorf("setting goose dialect: %w", err)
      }

      if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
          return fmt.Errorf("running goose up migrations: %w", err)
      }

      return nil
  }

  // RollbackMigration rolls back the last applied migration.
  func RollbackMigration(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
      db := stdlib.OpenDBFromPool(pool)
      defer db.Close()

      if err := goose.SetDialect("postgres"); err != nil {
          return fmt.Errorf("setting goose dialect: %w", err)
      }

      if err := goose.DownContext(ctx, db, migrationsDir); err != nil {
          return fmt.Errorf("running goose down migration: %w", err)
      }

      return nil
  }
  ```

- [x] **Step 8: Write migration unit tests (`backend/internal/repository/migrate_test.go`)**

  ```go
  package repository_test

  import (
      "os"
      "path/filepath"
      "regexp"
      "strings"
      "testing"
  )

  func TestMigrationFilesExistAndHaveGooseMarkers(t *testing.T) {
      migrationsDir := filepath.Join("..", "..", "migrations")
      entries, err := os.ReadDir(migrationsDir)
      if err != nil {
          t.Fatalf("reading migrations dir: %v", err)
      }

      pattern := regexp.MustCompile(`^\d{3}_\w+\.sql$`)
      sqlCount := 0

      for _, entry := range entries {
          if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
              continue
          }
          sqlCount++
          name := entry.Name()
          if !pattern.MatchString(name) {
              t.Errorf("migration file %q does not match pattern NNN_name.sql", name)
          }

          content, err := os.ReadFile(filepath.Join(migrationsDir, name))
          if err != nil {
              t.Errorf("reading %s: %v", name, err)
              continue
          }
          strContent := string(content)
          if !strings.Contains(strContent, "-- +goose Up") {
              t.Errorf("migration file %s missing '-- +goose Up' marker", name)
          }
          if !strings.Contains(strContent, "-- +goose Down") {
              t.Errorf("migration file %s missing '-- +goose Down' marker", name)
          }
      }

      if sqlCount == 0 {
          t.Fatal("no SQL migration files found in migrations directory")
      }
  }
  ```

- [x] **Step 9: Update `main.go` with CLI subcommand, startup migration, and version logging**

  Update `backend/cmd/arch-stats/main.go`:
    - Handle `migrate` subcommand:

      ```go
      if len(os.Args) > 1 && os.Args[1] == "migrate" {
          if err := repository.RunMigrations(ctx, pool, "migrations"); err != nil {
              log.Fatalf("migration failed: %v", err)
          }
          log.Println("migrations applied successfully")
          return
      }
      ```

    - Auto-run on startup if configured:

      ```go
      if cfg.ApplyMigrationsOnStart {
          if err := repository.RunMigrations(ctx, pool, "migrations"); err != nil {
              log.Fatalf("startup migration failed: %v", err)
          }
      }
      ```

    - Log the current schema version on startup (regardless of whether migrations ran):

      ```go
      version, err := repository.GetSchemaVersion(ctx, pool)
      if err != nil {
          slog.Warn("could not read schema version", "error", err)
      } else {
          slog.Info("database schema version", "version", version)
      }
      ```

- [x] **Step 10: Run tests, vet, and build**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  go vet ./...
  go build ./cmd/arch-stats
  ```

- [x] **Step 11: Commit to `arch-stats` main repository**

  ```bash
  git add .gitmodules backend/go.mod backend/go.sum backend/internal/repository/migrate.go \
          backend/internal/repository/migrate_test.go backend/cmd/arch-stats/main.go \
          backend/migrations
  git commit -m "feat: integrate goose migrations with startup and CLI subcommand (submodule updated)"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` migration file tests pass.
- `cd backend && go vet ./...` clean.
- `cd backend && go build ./cmd/arch-stats` compiles without errors.
- `cd backend/migrations && ./scripts/run_migration_tests.bash` all schema regression tests pass.
- `git status` in `backend/migrations` and root `arch-stats` both confirm clean working trees with
  appropriate branch tracking.
