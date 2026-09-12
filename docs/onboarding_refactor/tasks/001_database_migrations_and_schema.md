# Task 001: Database Migrations for Auth, Archer, Bow, and Arrow Schema

## Git Branch

`feature/001-db-migrations-auth-archer-bow-arrow`

## Objective

Refactor database schema migrations to cleanly separate authentication identity data (`auth`)
from archer personal profile data (`archer`), introduce dedicated equipment tables for bows
(`bow`) and optional arrows (`arrow`), and update session and shot foreign key references.

Per project guidelines, there is no requirement to maintain backward compatibility with legacy
migrations; migrations can be added, modified, or replaced to establish the canonical schema.

## Dependencies

None (foundational task for the onboarding refactor).

## Acceptance Criteria

- [x] PostgreSQL migration files in `backend/migrations/` define the canonical target schema:
    - [x] `auth` table:
        - `archer_id UUID PRIMARY KEY DEFAULT uuid_generate_v4()`
        - `google_subject TEXT NOT NULL UNIQUE`
        - `google_picture_url TEXT`
        - `last_login_at TIMESTAMPTZ NOT NULL DEFAULT now()`
        - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
    - [x] `archer` table:
        - `archer_id UUID PRIMARY KEY REFERENCES auth(archer_id) ON DELETE CASCADE`
        - `email VARCHAR(100) NOT NULL` (unique partial index: `uq_archer_email_active ON
          archer(lower(email)) WHERE is_deleted = FALSE`)
        - `first_name VARCHAR(100) NOT NULL CHECK (length(trim(first_name)) > 0)`
        - `last_name VARCHAR(100) NOT NULL CHECK (length(trim(last_name)) > 0)`
        - `date_of_birth DATE NOT NULL`
        - Trigger `trg_check_archer_age` `BEFORE INSERT OR UPDATE OF date_of_birth ON archer`
          enforcing `date_of_birth <= CURRENT_DATE - INTERVAL '10 years'`
        - `gender GENDER_TYPE NOT NULL`
        - `is_deleted BOOLEAN NOT NULL DEFAULT FALSE`
        - Legacy columns removed (`bowstyle`, `draw_weight`, `club_id`, `google_subject`,
          `google_picture_url`, `last_login_at`, `created_at`).
    - [x] `bow` table:
        - `bow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4()`
        - `archer_id UUID NOT NULL REFERENCES archer(archer_id) ON DELETE CASCADE`
        - `name VARCHAR(255) NOT NULL CHECK (length(trim(name)) > 0)`
        - `bowstyle BOWSTYLE_TYPE NOT NULL`
        - `draw_weight REAL NOT NULL CHECK (draw_weight > 0 AND draw_weight <= 200)`
        - `is_deleted BOOLEAN NOT NULL DEFAULT FALSE`
        - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
        - Composite unique constraint: `CONSTRAINT uq_bow_archer UNIQUE (archer_id, bow_id)`
        - Index on `bow(archer_id) WHERE is_deleted = FALSE`
    - [x] `arrow_status` enum:
        - `CREATE TYPE arrow_status AS ENUM ('in_use', 'damaged', 'lost')`
    - [x] `arrow` table:
        - `arrow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4()`
        - `archer_id UUID NOT NULL REFERENCES archer(archer_id) ON DELETE CASCADE`
        - `arrow_set SMALLINT NOT NULL`
        - `arrow_number SMALLINT NOT NULL CHECK (arrow_number >= 1)`
        - `spine REAL CHECK (spine IS NULL OR spine > 0)`
        - `length REAL CHECK (length IS NULL OR length > 0)`
        - `weight REAL CHECK (weight IS NULL OR weight > 0)`
        - `status arrow_status NOT NULL DEFAULT 'in_use'`
        - `is_deleted BOOLEAN NOT NULL DEFAULT FALSE`
        - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
        - Unique partial index: `UNIQUE (archer_id, arrow_set, arrow_number) WHERE
          is_deleted = FALSE`
        - Deferrable constraint trigger `trg_check_arrow_set_count` enforcing minimum 3 arrows per
          set
        - Index on `arrow(archer_id) WHERE status = 'in_use' AND is_deleted = FALSE`
    - [x] `session` table foreign key:
        - `(archer_id, bow_id) REFERENCES bow(archer_id, bow_id) ON DELETE RESTRICT`
    - [x] `shot` table foreign key:
        - `arrow_id UUID REFERENCES arrow(arrow_id) ON DELETE SET NULL`
    - [x] `auth_session` table (for server-side session token tracking):
        - `auth_id UUID PRIMARY KEY DEFAULT uuid_generate_v4()`
        - `archer_id UUID NOT NULL REFERENCES auth(archer_id) ON DELETE CASCADE`
        - `session_token_hash BYTEA NOT NULL UNIQUE`
        - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
        - `expires_at TIMESTAMPTZ NOT NULL`
        - `revoked_at TIMESTAMPTZ`
        - `ua TEXT`
        - `ip_inet INET`
        - Index on `auth_session(expires_at)`
- [x] `./scripts/run_migration_tests.bash` succeeds without errors.
- [x] Goose migrations apply cleanly (`goose up`), roll back cleanly (`goose down-to 0`), and
      re-apply cleanly (`goose up`).

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `backend/migrations/002_2025-09-26_archers_table.sql` |
| Modify | `backend/migrations/003_2025-09-26_authentication_session_table.sql` |
| Create | `backend/migrations/004_2026-03-10_bow_table.sql` |
| Modify | `backend/migrations/005_2025-10-28_arrow_table.sql` |
| Modify | `backend/migrations/006_2025-10-28_shot_table.sql` |
| Modify | `backend/migrations/scripts/run_migration_tests.bash` |

## Reference

- [PRD.md](../PRD.md)
- [README.md](../../../backend/migrations/README.md)
- [story_time.md](../../../backend/migrations/story_time.md)

## Steps

- [x] **Step 1: Inspect and update migrations for auth and archer separation**

  Update migration DDL to create `auth` first (containing `archer_id`, `google_subject`,
  `google_picture_url`, `last_login_at`, and `created_at`), then create `archer` with shared primary
  key `archer_id REFERENCES auth(archer_id)`, personal fields only, and `is_deleted BOOLEAN NOT NULL
  DEFAULT FALSE`. Create unique partial index on `lower(email)` for active archers. Add age trigger:

  ```sql
  CREATE OR REPLACE FUNCTION check_archer_age() RETURNS TRIGGER AS $$
  BEGIN
      IF NEW.date_of_birth > CURRENT_DATE - INTERVAL '10 years' THEN
          RAISE EXCEPTION 'Archer must be at least 10 years old (DOB: %)', NEW.date_of_birth
              USING ERRCODE = 'check_violation';
      END IF;
      RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  CREATE TRIGGER trg_check_archer_age
  BEFORE INSERT OR UPDATE OF date_of_birth ON archer
  FOR EACH ROW EXECUTE FUNCTION check_archer_age();
  ```

- [x] **Step 2: Create the `bow` table migration**

  Add migration creating table `bow`:

  ```sql
  CREATE TABLE IF NOT EXISTS bow (
      bow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      archer_id UUID NOT NULL REFERENCES archer(archer_id) ON DELETE CASCADE,
      name VARCHAR(255) NOT NULL,
      bowstyle BOWSTYLE_TYPE NOT NULL,
      draw_weight REAL NOT NULL CHECK (draw_weight > 0 AND draw_weight <= 200),
      is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      CONSTRAINT uq_bow_archer UNIQUE (archer_id, bow_id),
      CONSTRAINT bow_name_not_empty CHECK (length(trim(name)) > 0)
  );

  CREATE INDEX IF NOT EXISTS idx_bow_archer_active ON bow (archer_id) WHERE is_deleted = FALSE;
  ```

- [x] **Step 3: Update `arrow` table migration**

  Update `arrow` table schema to include `arrow_status` enum, `status` column, `SMALLINT` set and
  number, compact `REAL` specs, and composite partial unique constraint:

  ```sql
  DO $$
  BEGIN
      IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'arrow_status') THEN
          CREATE TYPE arrow_status AS ENUM ('in_use', 'damaged', 'lost');
      END IF;
  END$$;

  CREATE TABLE IF NOT EXISTS arrow (
      arrow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      archer_id UUID NOT NULL REFERENCES archer(archer_id) ON DELETE CASCADE,
      arrow_set SMALLINT NOT NULL,
      arrow_number SMALLINT NOT NULL CHECK (arrow_number >= 1),
      spine REAL CHECK (spine IS NULL OR spine > 0),
      length REAL CHECK (length IS NULL OR length > 0),
      weight REAL CHECK (weight IS NULL OR weight > 0),
      status arrow_status NOT NULL DEFAULT 'in_use',
      is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now()
  );

  CREATE UNIQUE INDEX IF NOT EXISTS uq_archer_arrow_set_number_active
  ON arrow (archer_id, arrow_set, arrow_number)
  WHERE is_deleted = FALSE;

  CREATE INDEX IF NOT EXISTS idx_arrow_archer_in_use
  ON arrow (archer_id)
  WHERE status = 'in_use' AND is_deleted = FALSE;
  ```

- [x] **Step 4: Create `auth_session` table and update session/shot foreign keys**

  Add migration for server-side token management:

  ```sql
  CREATE TABLE IF NOT EXISTS auth_session (
      auth_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      archer_id UUID NOT NULL REFERENCES auth(archer_id) ON DELETE CASCADE,
      session_token_hash BYTEA NOT NULL UNIQUE,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      expires_at TIMESTAMPTZ NOT NULL,
      revoked_at TIMESTAMPTZ,
      ua TEXT,
      ip_inet INET
  );

  CREATE INDEX IF NOT EXISTS idx_auth_session_expires_at ON auth_session (expires_at);
  CREATE INDEX IF NOT EXISTS idx_auth_session_archer_id ON auth_session (archer_id);
  ```

  Ensure `session` references `bow(archer_id, bow_id)` and `shot` references `arrow(arrow_id)`.

- [x] **Step 5: Run migration test script**

  ```bash
  ./backend/migrations/scripts/run_migration_tests.bash
  ```

  Verify `goose up`, rollback to 0, and re-run all pass cleanly.

- [x] **Step 6: Commit changes**

  ```bash
  git add backend/migrations
  git commit -m "feat(db): refactor schema for auth, archer, bow, and arrow separation"
  ```

## Verification

- `./backend/migrations/scripts/run_migration_tests.bash` exits with code 0.
- All tables, columns, check constraints, and foreign keys match `backend/migrations/README.md`.
