# Database Migrations for Auth, Archer, Bow, and Arrow Schema Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to
> implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor PostgreSQL database schema migrations to cleanly separate authentication
identity data (`auth`) from archer personal profile data (`archer`), introduce equipment tables for
bows (`bow`) and optional arrows (`arrow`), establish server-side token management (`auth_session`),
update cross-entity foreign keys, and expand the migration test suite.

**Architecture:** Database migrations in `backend/migrations/` managed with Goose
(`pressly/goose/v3`). Enforce declarative constraints, partial indexes for active entities,
deferrable trigger checks for multi-row operations (`trg_check_arrow_set_count`), age validation
triggers (`trg_check_archer_age`), and table/column SQL comments, ensuring full two-way lifecycle
integrity (`goose up` -> `goose down-to 0` -> `goose up`).

**Tech Stack:** PostgreSQL 17, Goose v3, Bash, Docker Compose, `pgx/v5`.

**Spec:** [docs/onboarding_refactor/tasks/001_database_migrations_and_schema.md][spec-doc]

[spec-doc]: ../onboarding_refactor/tasks/001_database_migrations_and_schema.md

## Global Constraints

- Database migrations must strictly follow Goose naming: `NNN_YYYY-MM-DD_description.sql` with
  `-- +goose Up` and `-- +goose Down` markers.
- Every migration must be losslessly reversible via `goose down-to 0`.
- All PL/pgSQL functions and multi-statement blocks must be wrapped in `-- +goose StatementBegin`
  and `-- +goose StatementEnd`.
- Every created or updated table must have `-- Table description` with `COMMENT ON TABLE`.
- Every column on created or updated tables must have `-- Column descriptions` with
  `COMMENT ON COLUMN`.
- DB access in tests and app uses parameterization; no ORM.
- Zero backward-compatibility requirement with legacy migration versions per project guidelines.
- Verification command: `./backend/migrations/scripts/run_migration_tests.bash` must exit 0 with all
  test cases passing.

## Migration Sequence & Versioning Resolution

In the existing codebase, `backend/migrations/` contains:

- `001_2025-09-26_db_init.sql` (extension `uuid-ossp`)
- `002_2025-09-26_archers_table.sql` (legacy `archer` table)
- `003_2025-09-26_authentication_session_table.sql` (legacy session table named `auth`)
- `004_2025-09-26_shooting_sessions_table.sql` (legacy `session`, `target`, `slot`)
- `005_2025-10-28_arrow_table.sql` (legacy `arrow`)
- `006_2025-10-28_shot_table.sql` (legacy `shot`)

Task 001 requests creating `backend/migrations/004_2026-03-10_bow_table.sql`. Because Goose strictly
forbids duplicate integer version numbers (both using prefix `004_`), and `session` declares a
foreign key referencing `bow`, `bow` must be executed before `session`. The canonical sequential
ordering is:

1. `001_2025-09-26_db_init.sql` (`uuid-ossp`)
2. `002_2025-09-26_archers_table.sql` (`auth` and `archer`)
3. `003_2025-09-26_authentication_session_table.sql` (`auth_session`)
4. `004_2026-03-10_bow_table.sql` (`bow`)
5. `005_2025-10-28_arrow_table.sql` (`arrow_status`, `arrow`, trigger)
6. `006_2025-09-26_shooting_sessions_table.sql` (`session` with FK to `bow`, `target`, `slot`)
7. `007_2025-10-28_shot_table.sql` (`shot` with FK to `arrow`)

This establishes a clean, forward-dependency order matching
[`backend/migrations/README.md`](../../backend/migrations/README.md):
`auth` -> `archer` -> `auth_session` -> `bow` -> `arrow` -> `session` -> `shot`.

---

### Task 1: Auth & Archer Separation Migration

**Files:**

- Modify: `backend/migrations/002_2025-09-26_archers_table.sql`
- Test: `backend/migrations/scripts/run_migration_tests.bash`

**Interfaces:**

- Consumes: `uuid-ossp` extension from `001_2025-09-26_db_init.sql`
- Produces:
    - Table `auth` (`archer_id UUID PK`, `google_subject TEXT UNIQUE`, `google_picture_url TEXT`,
      `last_login_at TIMESTAMPTZ`, `created_at TIMESTAMPTZ`)
    - Table `archer` (`archer_id UUID PK/FK`, `email VARCHAR(100)`, `first_name`, `last_name`,
      `date_of_birth`, `gender`, `is_deleted`)
    - Table and column comments on `auth` and `archer`
    - Index `uq_archer_email` on `archer(lower(email))` (unique across all records)
    - Function `check_archer_age()` and trigger `trg_check_archer_age` enforcing
      `date_of_birth <= CURRENT_DATE - INTERVAL '10 years'`
    - Enums `bowstyle_type`, `gender_type`, `face_type`

- [ ] **Step 1: Write failing test in `run_migration_tests.bash` for auth & archer separation**

Add test function `test_auth_and_archer_separation` into
`backend/migrations/scripts/run_migration_tests.bash`:

```bash
test_auth_and_archer_separation() {
    header "auth and archer separation (DDL, constraints, triggers)"

    # 1. Insert into auth table directly
    local suffix archer_id
    suffix="$(random_suffix)"
    archer_id="$(run_sql "
        INSERT INTO auth (google_subject, google_picture_url)
        VALUES ('gs-${suffix}', 'https://example.com/pic-${suffix}.jpg')
        RETURNING archer_id;
    ")"

    if [[ -n "$archer_id" ]]; then
        pass "Created auth record: $archer_id"
    else
        fail "Failed to create auth record"
        return
    fi

    # 2. auth.google_subject unique constraint
    if run_sql "
        INSERT INTO auth (google_subject)
        VALUES ('gs-${suffix}');
    " >/dev/null 2>&1; then
        fail "auth allowed duplicate google_subject"
    else
        pass "auth enforces UNIQUE(google_subject)"
    fi

    # 3. Create archer referencing auth(archer_id)
    local email="archer_${suffix}@example.com"
    if run_sql "
        INSERT INTO archer (archer_id, email, first_name, last_name, date_of_birth, gender)
        VALUES ('$archer_id', '$email', 'Robin', 'Hood', '1995-05-15', 'male');
    " >/dev/null 2>&1; then
        pass "Created archer profile referencing auth"
    else
        fail "Failed to create archer profile referencing auth"
    fi

    # 4. archer unique email index across all records
    local auth_id_2
    auth_id_2="$(run_sql "
        INSERT INTO auth (google_subject)
        VALUES ('gs2-${suffix}')
        RETURNING archer_id;
    ")"
    if run_sql "
        INSERT INTO archer (archer_id, email, first_name, last_name, date_of_birth, gender)
        VALUES ('$auth_id_2', '$email', 'John', 'Little', '1992-02-02', 'male');
    " >/dev/null 2>&1; then
        fail "archer allowed duplicate active email"
    else
        pass "archer enforces uq_archer_email"
    fi

    # 5. Soft-deleted archer still prohibits duplicate email
    run_sql "UPDATE archer SET is_deleted = TRUE WHERE archer_id = '$archer_id';"
    if run_sql "
        INSERT INTO archer (archer_id, email, first_name, last_name, date_of_birth, gender)
        VALUES ('$auth_id_2', '$email', 'John', 'Little', '1992-02-02', 'male');
    " >/dev/null 2>&1; then
        fail "archer allowed duplicate email when prior record is_deleted = TRUE"
    else
        pass "archer rejects duplicate email regardless of is_deleted status"
    fi

    # 6. Age check trigger (< 10 years old rejected)
    local auth_id_young young_dob
    young_dob="$(date -d '9 years ago' +%Y-%m-%d 2>/dev/null || date -v-9y +%Y-%m-%d)"
    auth_id_young="$(run_sql "
        INSERT INTO auth (google_subject)
        VALUES ('young-${suffix}')
        RETURNING archer_id;
    ")"
    if run_sql "
        INSERT INTO archer (
            archer_id, email, first_name, last_name, date_of_birth, gender
        )
        VALUES (
            '$auth_id_young',
            'young_${suffix}@example.com',
            'Young',
            'Archer',
            '$young_dob',
            'female'
        );
    " >/dev/null 2>&1; then
        fail "trg_check_archer_age allowed archer younger than 10 years old"
    else
        pass "trg_check_archer_age rejected archer younger than 10 years old"
    fi

    # 7. Name check constraints (empty strings rejected)
    local auth_id_empty
    auth_id_empty="$(run_sql "
        INSERT INTO auth (google_subject)
        VALUES ('empty-${suffix}')
        RETURNING archer_id;
    ")"
    if run_sql "
        INSERT INTO archer (
            archer_id, email, first_name, last_name, date_of_birth, gender
        )
        VALUES (
            '$auth_id_empty',
            'empty_${suffix}@example.com',
            '   ',
            'Valid',
            '1990-01-01',
            'female'
        );
    " >/dev/null 2>&1; then
        fail "archer allowed blank first_name"
    else
        pass "archer rejects blank first_name"
    fi

    # Clean up
    run_sql "DELETE FROM auth WHERE archer_id IN (
        '$archer_id', '$auth_id_2', '$auth_id_young', '$auth_id_empty'
    );"
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: FAIL with `relation "auth" does not exist` or column mismatch.

- [ ] **Step 3: Update `backend/migrations/002_2025-09-26_archers_table.sql`**

Replace `backend/migrations/002_2025-09-26_archers_table.sql` with canonical DDL including comments:

```sql
-- +goose Up
-- +goose StatementBegin
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
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS auth (
    archer_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    google_subject TEXT NOT NULL UNIQUE,
    google_picture_url TEXT,
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Table description
COMMENT ON TABLE auth IS 'Authentication identity credentials managed via Google OAuth.';

-- Column descriptions
COMMENT ON COLUMN auth.archer_id IS
    'Unique archer identifier auto-generated on first OAuth sign-in.';
COMMENT ON COLUMN auth.google_subject IS
    'Unique Google account subject identifier (sub claim).';
COMMENT ON COLUMN auth.google_picture_url IS
    'Profile avatar picture URL provided by Google OAuth.';
COMMENT ON COLUMN auth.last_login_at IS
    'Timestamp of the most recent successful sign-in.';
COMMENT ON COLUMN auth.created_at IS
    'Timestamp when the authentication record was first created.';

CREATE TABLE IF NOT EXISTS archer (
    archer_id UUID PRIMARY KEY REFERENCES auth (archer_id) ON DELETE CASCADE,
    email VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    gender GENDER_TYPE NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT archer_first_name_not_empty CHECK (length(trim(first_name)) > 0),
    CONSTRAINT archer_last_name_not_empty CHECK (length(trim(last_name)) > 0)
);

-- Table description
COMMENT ON TABLE archer IS
    'Personal profile details for an archer collected during onboarding.';

-- Column descriptions
COMMENT ON COLUMN archer.archer_id IS 'Shared primary key referencing auth.archer_id.';
COMMENT ON COLUMN archer.email IS
    'Contact email address; must be unique across all archers regardless of status.';
COMMENT ON COLUMN archer.first_name IS 'Archer given first name.';
COMMENT ON COLUMN archer.last_name IS 'Archer family last name.';
COMMENT ON COLUMN archer.date_of_birth IS
    'Date of birth; must be at least 10 years in the past.';
COMMENT ON COLUMN archer.gender IS 'Gender identity classification.';
COMMENT ON COLUMN archer.is_deleted IS
    'Soft-delete flag; true indicates profile is archived/deleted.';

CREATE UNIQUE INDEX IF NOT EXISTS uq_archer_email
ON archer (lower(email));

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_archer_age() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.date_of_birth > CURRENT_DATE - INTERVAL '10 years' THEN
        RAISE EXCEPTION 'Archer must be at least 10 years old (DOB: %)', NEW.date_of_birth
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_check_archer_age ON archer;
CREATE TRIGGER trg_check_archer_age
BEFORE INSERT OR UPDATE OF date_of_birth ON archer
FOR EACH ROW EXECUTE FUNCTION check_archer_age();

-- +goose Down
DROP TRIGGER IF EXISTS trg_check_archer_age ON archer;
DROP FUNCTION IF EXISTS check_archer_age();
DROP INDEX IF EXISTS uq_archer_email;
DROP TABLE IF EXISTS archer CASCADE;
DROP TABLE IF EXISTS auth CASCADE;
DROP TYPE IF EXISTS face_type CASCADE;
DROP TYPE IF EXISTS gender_type CASCADE;
DROP TYPE IF EXISTS bowstyle_type CASCADE;
```

Update `create_archer()` in `run_migration_tests.bash` to write to `auth` first, then `archer`:

```bash
create_archer() {
    local suffix
    suffix="$(random_suffix)"
    local email="migtest_${suffix}@example.com"
    local gsubj="gs-${suffix}"
    local sql_statement="""
        WITH new_auth AS (
            INSERT INTO auth (google_subject, google_picture_url)
            VALUES ('$gsubj', 'https://example.com/pic-${suffix}.jpg')
            RETURNING archer_id
        )
        INSERT INTO archer (
            archer_id,
            email,
            first_name,
            last_name,
            date_of_birth,
            gender
        )
        SELECT
            archer_id,
            '$email',
            'Test',
            '$suffix',
            '1990-01-01'::date,
            'male'::gender_type
        FROM new_auth
        RETURNING archer_id;
"""
    run_sql "${sql_statement}"
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: `test_auth_and_archer_separation` PASS.

- [ ] **Step 5: Commit**

```bash
cd backend/migrations
git add 002_2025-09-26_archers_table.sql scripts/run_migration_tests.bash
git commit -m "feat(db): separate auth and archer tables with age trigger, index, and comments"
```

---

### Task 2: Authentication Session Table Migration

**Files:**

- Modify: `backend/migrations/003_2025-09-26_authentication_session_table.sql`
- Test: `backend/migrations/scripts/run_migration_tests.bash`

**Interfaces:**

- Consumes: `auth(archer_id)` from `002_2025-09-26_archers_table.sql`
- Produces:
    - Table `auth_session` (`auth_id UUID PK`, `archer_id UUID REFERENCES auth(archer_id)`,
      `session_token_hash BYTEA UNIQUE`, `created_at`, `expires_at`, `revoked_at`, `ua`, `ip_inet`)
    - Table and column comments on `auth_session`
    - Indexes: `idx_auth_session_expires_at`, `idx_auth_session_archer_id`

- [ ] **Step 1: Write failing test in `run_migration_tests.bash` for `auth_session`**

Update `test_auth_table_constraints_and_cascade` in `run_migration_tests.bash`:

```bash
test_auth_table_constraints_and_cascade() {
    header "auth_session table: constraints, indexes, and cascade"

    local archer_id
    archer_id="$(create_archer)"

    # Insert auth_session referencing auth(archer_id)
    local auth_id
    auth_id="$(run_sql "
        INSERT INTO auth_session (
            archer_id,
            session_token_hash,
            expires_at,
            ua,
            ip_inet
        )
        VALUES (
            '$archer_id',
            decode('e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855', 'hex'),
            now() + interval '24 hours',
            'Mozilla/5.0 MigrationTest',
            '127.0.0.1'::inet
        )
        RETURNING auth_id;
    ")"

    if [[ -n "$auth_id" ]]; then
        pass "Created auth_session: $auth_id"
    else
        fail "Failed to create auth_session"
    fi

    # UNIQUE session_token_hash
    if run_sql "
        INSERT INTO auth_session (
            archer_id,
            session_token_hash,
            expires_at
        )
        VALUES (
            '$archer_id',
            decode('e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855', 'hex'),
            now() + interval '24 hours'
        );
    " >/dev/null 2>&1; then
        fail "auth_session allowed duplicate session_token_hash"
    else
        pass "auth_session.session_token_hash enforces UNIQUE"
    fi

    # Cascade delete on auth deletion
    run_sql "DELETE FROM auth WHERE archer_id = '$archer_id';"
    local count
    count="$(run_sql "SELECT count(*) FROM auth_session WHERE auth_id = '$auth_id';")"
    if [[ "$count" -eq 0 ]]; then
        pass "auth_session row cascaded when auth/archer was deleted"
    else
        fail "auth_session row was not cascaded on delete"
    fi
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: FAIL with `relation "auth_session" does not exist`.

- [ ] **Step 3: Update `backend/migrations/003_2025-09-26_authentication_session_table.sql`**

Replace `backend/migrations/003_2025-09-26_authentication_session_table.sql` with:

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS auth_session (
    auth_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID NOT NULL REFERENCES auth (archer_id) ON DELETE CASCADE,
    session_token_hash BYTEA NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    ua TEXT,
    ip_inet INET
);

-- Table description
COMMENT ON TABLE auth_session IS
    'Server-side session tokens for authenticated user sessions.';

-- Column descriptions
COMMENT ON COLUMN auth_session.auth_id IS
    'Primary key identifying the session token record.';
COMMENT ON COLUMN auth_session.archer_id IS
    'Foreign key referencing auth.archer_id.';
COMMENT ON COLUMN auth_session.session_token_hash IS
    'SHA-256 hash of the bearer session token.';
COMMENT ON COLUMN auth_session.created_at IS
    'Timestamp when the session token was issued.';
COMMENT ON COLUMN auth_session.expires_at IS
    'Expiration timestamp after which the session token is invalid.';
COMMENT ON COLUMN auth_session.revoked_at IS
    'Timestamp when session was explicitly revoked (NULL = active).';
COMMENT ON COLUMN auth_session.ua IS
    'User-Agent client header string recorded at session creation.';
COMMENT ON COLUMN auth_session.ip_inet IS
    'Client IP address recorded at session creation.';

CREATE INDEX IF NOT EXISTS idx_auth_session_expires_at
ON auth_session (expires_at);

CREATE INDEX IF NOT EXISTS idx_auth_session_archer_id
ON auth_session (archer_id);

-- +goose Down
DROP INDEX IF EXISTS idx_auth_session_archer_id;
DROP INDEX IF EXISTS idx_auth_session_expires_at;
DROP TABLE IF EXISTS auth_session CASCADE;
```

- [ ] **Step 4: Run test to verify it passes**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: `test_auth_table_constraints_and_cascade` PASS.

- [ ] **Step 5: Commit**

```bash
cd backend/migrations
git add 003_2025-09-26_authentication_session_table.sql scripts/run_migration_tests.bash
git commit -m "feat(db): rename session table to auth_session with indexes and comments"
```

---

### Task 3: Bow Table Migration

**Files:**

- Create: `backend/migrations/004_2026-03-10_bow_table.sql`
- Test: `backend/migrations/scripts/run_migration_tests.bash`

**Interfaces:**

- Consumes: `archer(archer_id)` and `bowstyle_type` from `002_2025-09-26_archers_table.sql`
- Produces:
    - Table `bow` (`bow_id UUID PK`, `archer_id UUID FK`, `name VARCHAR(255)`,
      `bowstyle BOWSTYLE_TYPE`, `draw_weight REAL`, `is_deleted BOOLEAN`, `created_at TIMESTAMPTZ`)
    - Table and column comments on `bow`
    - Constraint `uq_bow_archer UNIQUE (archer_id, bow_id)`
    - Constraint `bow_name_not_empty CHECK (length(trim(name)) > 0)`
    - Constraint `draw_weight CHECK (draw_weight > 0 AND draw_weight <= 200)`
    - Index `idx_bow_archer_active ON bow(archer_id) WHERE is_deleted = FALSE`

- [ ] **Step 1: Write failing test in `run_migration_tests.bash` for `bow` table**

Add helper `create_bow` and test function `test_bow_constraints`:

```bash
create_bow() {
    local archer_id="$1"
    local name="${2:-My Recurve}"
    local bowstyle="${3:-recurve}"
    local draw_weight="${4:-38.5}"
    local sql_statement="""
        INSERT INTO bow (archer_id, name, bowstyle, draw_weight)
        VALUES ('$archer_id', '$name', '$bowstyle'::bowstyle_type, $draw_weight)
        RETURNING bow_id;
"""
    run_sql "${sql_statement}"
}

test_bow_constraints() {
    header "bow table: constraints, composite unique, and soft deletes"

    local archer_id
    archer_id="$(create_archer)"

    # 1. Happy path
    local bow_id
    bow_id="$(create_bow "$archer_id" "Competition Bow" "barebow" 42.0)"
    if [[ -n "$bow_id" ]]; then
        pass "Created bow: $bow_id"
    else
        fail "Failed to create bow"
        return
    fi

    # 2. Check blank name rejected
    if run_sql "
        INSERT INTO bow (archer_id, name, bowstyle, draw_weight)
        VALUES ('$archer_id', '   ', 'recurve', 30.0);
    " >/dev/null 2>&1; then
        fail "bow allowed blank name"
    else
        pass "bow rejects blank name"
    fi

    # 3. Check draw weight bounds (0 < weight <= 200)
    if run_sql "
        INSERT INTO bow (archer_id, name, bowstyle, draw_weight)
        VALUES ('$archer_id', 'Low Weight', 'recurve', 0.0);
    " >/dev/null 2>&1; then
        fail "bow allowed draw_weight = 0"
    else
        pass "bow rejects draw_weight <= 0"
    fi

    if run_sql "
        INSERT INTO bow (archer_id, name, bowstyle, draw_weight)
        VALUES ('$archer_id', 'High Weight', 'compound', 201.0);
    " >/dev/null 2>&1; then
        fail "bow allowed draw_weight > 200"
    else
        pass "bow rejects draw_weight > 200"
    fi

    # 4. Composite unique constraint (archer_id, bow_id)
    local count
    count="$(run_sql "
        SELECT count(*) FROM pg_constraint
        WHERE conname = 'uq_bow_archer';
    ")"
    if [[ "$count" -ge 1 ]]; then
        pass "bow has composite unique constraint uq_bow_archer"
    else
        fail "bow missing composite unique constraint uq_bow_archer"
    fi

    # 5. Cascade delete on archer deletion
    run_sql "DELETE FROM auth WHERE archer_id = '$archer_id';"
    count="$(run_sql "SELECT count(*) FROM bow WHERE bow_id = '$bow_id';")"
    if [[ "$count" -eq 0 ]]; then
        pass "bow cascaded on archer deletion"
    else
        fail "bow was not cascaded when archer deleted"
    fi
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: FAIL with `relation "bow" does not exist`.

- [ ] **Step 3: Create `backend/migrations/004_2026-03-10_bow_table.sql`**

Write `backend/migrations/004_2026-03-10_bow_table.sql`:

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS bow (
    bow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    bowstyle BOWSTYLE_TYPE NOT NULL,
    draw_weight REAL NOT NULL CHECK (draw_weight > 0 AND draw_weight <= 200),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_bow_archer UNIQUE (archer_id, bow_id),
    CONSTRAINT bow_name_not_empty CHECK (length(trim(name)) > 0)
);

-- Table description
COMMENT ON TABLE bow IS 'Equipment inventory of bows registered by archers.';

-- Column descriptions
COMMENT ON COLUMN bow.bow_id IS 'Primary key identifying the bow.';
COMMENT ON COLUMN bow.archer_id IS 'Foreign key referencing the owner archer profile.';
COMMENT ON COLUMN bow.name IS 'Archer-given descriptive name or label for the bow.';
COMMENT ON COLUMN bow.bowstyle IS
    'Bow discipline classification (recurve, compound, barebow, longbow).';
COMMENT ON COLUMN bow.draw_weight IS
    'Peak draw weight in pounds (compact 4-byte float, 0 < weight <= 200).';
COMMENT ON COLUMN bow.is_deleted IS
    'Soft-delete flag; true indicates the bow is archived/deleted.';
COMMENT ON COLUMN bow.created_at IS 'Timestamp when the bow was registered.';

CREATE INDEX IF NOT EXISTS idx_bow_archer_active
ON bow (archer_id)
WHERE is_deleted = FALSE;

-- +goose Down
DROP INDEX IF EXISTS idx_bow_archer_active;
DROP TABLE IF EXISTS bow CASCADE;
```

Renumber existing migrations to preserve strict ordering:

- Rename `004_2025-09-26_shooting_sessions_table.sql` ->
  `006_2025-09-26_shooting_sessions_table.sql`
- Keep `005_2025-10-28_arrow_table.sql` as `005`
- Rename `006_2025-10-28_shot_table.sql` ->
  `007_2025-10-28_shot_table.sql`

- [ ] **Step 4: Run test to verify it passes**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: `test_bow_constraints` PASS.

- [ ] **Step 5: Commit**

```bash
cd backend/migrations
git add 004_2026-03-10_bow_table.sql 006_2025-09-26_shooting_sessions_table.sql \
    007_2025-10-28_shot_table.sql scripts/run_migration_tests.bash
git commit -m "feat(db): add bow table migration with constraints, index, and comments"
```

---

### Task 4: Arrow Table Migration with Set Constraint Trigger

**Files:**

- Modify: `backend/migrations/005_2025-10-28_arrow_table.sql`
- Test: `backend/migrations/scripts/run_migration_tests.bash`

**Interfaces:**

- Consumes: `archer(archer_id)` from `002_2025-09-26_archers_table.sql`
- Produces:
    - Enum `arrow_status` (`'in_use'`, `'damaged'`, `'lost'`)
    - Table `arrow` (`arrow_id UUID PK`, `archer_id UUID FK`, `arrow_set SMALLINT`,
      `arrow_number SMALLINT`, `spine REAL`, `length REAL`, `weight REAL`, `status arrow_status`,
      `is_deleted BOOLEAN`, `created_at TIMESTAMPTZ`)
    - Table and column comments on `arrow`
    - Partial unique index `uq_archer_arrow_set_number_active`
    - Index `idx_arrow_archer_in_use ON arrow(archer_id)`
      WHERE `status = 'in_use' AND is_deleted = FALSE`
    - Deferrable constraint trigger `trg_check_arrow_set_count` enforcing min 3 arrows per set

- [ ] **Step 1: Write failing test in `run_migration_tests.bash` for `arrow` table**

Update `test_arrow_and_shot_cascades_and_constraints`:

```bash
test_arrow_constraints_and_set_trigger() {
    header "arrow table: enum, constraints, partial index, and deferrable set trigger"

    local archer_id
    archer_id="$(create_archer)"

    # 1. Batch insert 3 arrows in set 1 (within a transaction) succeeds
    if run_sql "
        BEGIN;
        INSERT INTO arrow (archer_id, arrow_set, arrow_number, spine, length, weight)
        VALUES
            ('$archer_id', 1, 1, 500, 28.5, 320),
            ('$archer_id', 1, 2, 500, 28.5, 320),
            ('$archer_id', 1, 3, 500, 28.5, 320);
        COMMIT;
    " >/dev/null 2>&1; then
        pass "Inserting valid 3-arrow set succeeded"
    else
        fail "Failed to insert valid 3-arrow set"
    fi

    # 2. Inserting only 2 arrows in set 2 fails at commit
    if run_sql "
        BEGIN;
        INSERT INTO arrow (archer_id, arrow_set, arrow_number)
        VALUES
            ('$archer_id', 2, 1),
            ('$archer_id', 2, 2);
        COMMIT;
    " >/dev/null 2>&1; then
        fail "trg_check_arrow_set_count allowed set with fewer than 3 arrows"
    else
        pass "trg_check_arrow_set_count rejected set with only 2 arrows"
    fi

    # 3. Partial unique index prevents duplicate active arrow_number in set
    if run_sql "
        INSERT INTO arrow (archer_id, arrow_set, arrow_number)
        VALUES ('$archer_id', 1, 1);
    " >/dev/null 2>&1; then
        fail "arrow allowed duplicate active (archer_id, arrow_set, arrow_number)"
    else
        pass "arrow enforces uq_archer_arrow_set_number_active"
    fi

    # 4. Soft-deleting one arrow leaving 2 fails deferrable check
    if run_sql "
        UPDATE arrow SET is_deleted = TRUE
        WHERE archer_id = '$archer_id' AND arrow_set = 1 AND arrow_number = 3;
    " >/dev/null 2>&1; then
        fail "trg_check_arrow_set_count allowed soft delete leaving 2 active arrows"
    else
        pass "trg_check_arrow_set_count rejected soft delete leaving 2 active arrows"
    fi

    # 5. Soft-deleting all arrows in set succeeds (count = 0)
    if run_sql "
        UPDATE arrow SET is_deleted = TRUE
        WHERE archer_id = '$archer_id' AND arrow_set = 1;
    " >/dev/null 2>&1; then
        pass "trg_check_arrow_set_count allowed soft-deleting entire arrow set"
    else
        fail "trg_check_arrow_set_count rejected emptying arrow set"
    fi

    # Clean up
    run_sql "DELETE FROM auth WHERE archer_id = '$archer_id';"
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: FAIL with missing columns or trigger.

- [ ] **Step 3: Update `backend/migrations/005_2025-10-28_arrow_table.sql`**

Replace `backend/migrations/005_2025-10-28_arrow_table.sql` with:

```sql
-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'arrow_status') THEN
        CREATE TYPE arrow_status AS ENUM ('in_use', 'damaged', 'lost');
    END IF;
END$$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS arrow (
    arrow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE CASCADE,
    arrow_set SMALLINT NOT NULL,
    arrow_number SMALLINT NOT NULL CHECK (arrow_number >= 1),
    spine REAL CHECK (spine IS NULL OR spine > 0),
    length REAL CHECK (length IS NULL OR length > 0),
    weight REAL CHECK (weight IS NULL OR weight > 0),
    status arrow_status NOT NULL DEFAULT 'in_use',
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Table description
COMMENT ON TABLE arrow IS
    'Equipment inventory of individual arrows organized into sets.';

-- Column descriptions
COMMENT ON COLUMN arrow.arrow_id IS 'Primary key identifying the arrow.';
COMMENT ON COLUMN arrow.archer_id IS
    'Foreign key referencing the owner archer profile.';
COMMENT ON COLUMN arrow.arrow_set IS
    'Arrow set group number grouping at least 3 active arrows.';
COMMENT ON COLUMN arrow.arrow_number IS
    'Identifying sequential number of the arrow within its set (>= 1).';
COMMENT ON COLUMN arrow.spine IS
    'Arrow stiffness deflection rating in inches/thousandths.';
COMMENT ON COLUMN arrow.length IS 'Arrow shaft length in inches.';
COMMENT ON COLUMN arrow.weight IS 'Arrow total weight in grains.';
COMMENT ON COLUMN arrow.status IS
    'Availability condition: in_use, damaged, or lost.';
COMMENT ON COLUMN arrow.is_deleted IS
    'Soft-delete flag; true indicates arrow is archived/deleted.';
COMMENT ON COLUMN arrow.created_at IS
    'Timestamp when the arrow was registered.';

CREATE UNIQUE INDEX IF NOT EXISTS uq_archer_arrow_set_number_active
ON arrow (archer_id, arrow_set, arrow_number)
WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_arrow_archer_in_use
ON arrow (archer_id)
WHERE status = 'in_use' AND is_deleted = FALSE;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_arrow_set_count() RETURNS TRIGGER AS $$
DECLARE
    v_archer_id UUID;
    v_arrow_set SMALLINT;
    v_count INTEGER;
BEGIN
    IF TG_OP = 'DELETE' THEN
        v_archer_id := OLD.archer_id;
        v_arrow_set := OLD.arrow_set;
    ELSE
        v_archer_id := NEW.archer_id;
        v_arrow_set := NEW.arrow_set;
    END IF;

    -- If UPDATE changed arrow_set, also check OLD set
    IF TG_OP = 'UPDATE' AND OLD.arrow_set IS DISTINCT FROM NEW.arrow_set THEN
        SELECT count(*) INTO v_count
        FROM arrow
        WHERE archer_id = OLD.archer_id
          AND arrow_set = OLD.arrow_set
          AND status = 'in_use'
          AND is_deleted = FALSE;

        IF v_count > 0 AND v_count < 3 THEN
            RAISE EXCEPTION
                'Arrow set % for archer % must contain at least 3 active arrows, found %',
                OLD.arrow_set, OLD.archer_id, v_count
                USING ERRCODE = 'check_violation';
        END IF;
    END IF;

    SELECT count(*) INTO v_count
    FROM arrow
    WHERE archer_id = v_archer_id
      AND arrow_set = v_arrow_set
      AND status = 'in_use'
      AND is_deleted = FALSE;

    IF v_count > 0 AND v_count < 3 THEN
        RAISE EXCEPTION
            'Arrow set % for archer % must contain at least 3 active arrows, found %',
            v_arrow_set, v_archer_id, v_count
            USING ERRCODE = 'check_violation';
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_check_arrow_set_count ON arrow;
CREATE CONSTRAINT TRIGGER trg_check_arrow_set_count
AFTER INSERT OR UPDATE OR DELETE ON arrow
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION check_arrow_set_count();

-- +goose Down
DROP TRIGGER IF EXISTS trg_check_arrow_set_count ON arrow;
DROP FUNCTION IF EXISTS check_arrow_set_count();
DROP INDEX IF EXISTS idx_arrow_archer_in_use;
DROP INDEX IF EXISTS uq_archer_arrow_set_number_active;
DROP TABLE IF EXISTS arrow CASCADE;
DROP TYPE IF EXISTS arrow_status CASCADE;
```

- [ ] **Step 4: Run test to verify it passes**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: `test_arrow_constraints_and_set_trigger` PASS.

- [ ] **Step 5: Commit**

```bash
cd backend/migrations
git add 005_2025-10-28_arrow_table.sql scripts/run_migration_tests.bash
git commit -m "feat(db): update arrow schema with arrow_status, trigger, and comments"
```

---

### Task 5: Session and Shot Foreign Key Updates

**Files:**

- Modify: `backend/migrations/006_2025-09-26_shooting_sessions_table.sql`
- Modify: `backend/migrations/007_2025-10-28_shot_table.sql`
- Test: `backend/migrations/scripts/run_migration_tests.bash`

**Interfaces:**

- Consumes:
    - `bow(archer_id, bow_id)` from `004_2026-03-10_bow_table.sql`
    - `arrow(arrow_id)` from `005_2025-10-28_arrow_table.sql`
- Produces:
    - Table and column comments on `session`, `target`, `slot`, and `shot`
    - Foreign key on `session`:
      `CONSTRAINT fk_session_archer_bow FOREIGN KEY (archer_id, bow_id)`
      `REFERENCES bow(archer_id, bow_id) ON DELETE RESTRICT`
    - Foreign key on `shot`: `arrow_id UUID REFERENCES arrow(arrow_id) ON DELETE SET NULL`

- [ ] **Step 1: Write failing test in `run_migration_tests.bash` for FKs**

Add `test_session_bow_and_shot_arrow_fks`:

```bash
test_session_bow_and_shot_arrow_fks() {
    header "session-bow and shot-arrow foreign key integrity"

    local archer_a archer_b bow_a bow_b
    archer_a="$(create_archer)"
    archer_b="$(create_archer)"
    bow_a="$(create_bow "$archer_a" "Archer A Bow")"
    bow_b="$(create_bow "$archer_b" "Archer B Bow")"

    # 1. Session with matching archer_id and bow_id succeeds
    local session_id
    session_id="$(run_sql "
        INSERT INTO session (
            owner_archer_id, archer_id, bow_id, session_location, is_indoor, is_opened
        )
        VALUES ('$archer_a', '$archer_a', '$bow_a', 'Indoor Range', true, true)
        RETURNING session_id;
    ")"
    if [[ -n "$session_id" ]]; then
        pass "Created session referencing archer's own bow"
    else
        fail "Failed to create session with valid bow FK"
    fi

    # 2. Session with mismatched archer_id and bow_id fails FK check
    if run_sql "
        INSERT INTO session (
            owner_archer_id, archer_id, bow_id, session_location, is_indoor, is_opened
        )
        VALUES ('$archer_a', '$archer_a', '$bow_b', 'Indoor Range', true, true);
    " >/dev/null 2>&1; then
        fail "session allowed foreign bow owned by another archer"
    else
        pass "session enforces foreign key (archer_id, bow_id) REFERENCES bow"
    fi

    # 3. ON DELETE RESTRICT on bow prevents deleting bow while session exists
    if run_sql "DELETE FROM bow WHERE bow_id = '$bow_a';" >/dev/null 2>&1; then
        fail "bow was deleted despite active session reference (RESTRICT failed)"
    else
        pass "bow deletion restricted when referenced by session"
    fi

    # 4. Shot arrow FK ON DELETE SET NULL
    local target_id slot_id arrow_id shot_id
    target_id="$(create_target "$session_id" 18 1)"
    slot_id="$(assign_slot "$target_id" "$archer_a" "$session_id" "A" "true")"
    run_sql "
        INSERT INTO arrow (archer_id, arrow_set, arrow_number)
        VALUES ('$archer_a', 1, 1), ('$archer_a', 1, 2), ('$archer_a', 1, 3);
    "
    arrow_id="$(run_sql "
        SELECT arrow_id FROM arrow
        WHERE archer_id = '$archer_a' AND arrow_number = 1;
    ")"
    shot_id="$(run_sql "
        INSERT INTO shot (slot_id, x, y, score, arrow_id, is_x)
        VALUES ('$slot_id', 0.0, 0.0, 10, '$arrow_id', true)
        RETURNING shot_id;
    ")"
    run_sql "DELETE FROM arrow WHERE arrow_id = '$arrow_id';"
    local shot_arrow
    shot_arrow="$(run_sql "
        SELECT coalesce(arrow_id::text, 'NULL') FROM shot
        WHERE shot_id = '$shot_id';
    ")"
    if [[ "$shot_arrow" == "NULL" ]]; then
        pass "shot.arrow_id was SET NULL when arrow was deleted"
    else
        fail "shot.arrow_id was not set to NULL on arrow delete (got $shot_arrow)"
    fi

    # Clean up
    run_sql "DELETE FROM auth WHERE archer_id IN ('$archer_a', '$archer_b');"
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: FAIL due to missing `bow_id` column or constraint on `session`.

- [ ] **Step 3: Update `session` and `shot` migration DDL**

In `backend/migrations/006_2025-09-26_shooting_sessions_table.sql`:
Update `session` table definition to include `archer_id`, `bow_id`, composite FK, and comments:

```sql
CREATE TABLE IF NOT EXISTS session (
    session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID REFERENCES archer (archer_id) ON DELETE RESTRICT,
    bow_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at TIMESTAMPTZ,
    session_location VARCHAR(255) NOT NULL,
    is_indoor BOOLEAN NOT NULL,
    is_opened BOOLEAN NOT NULL,
    CONSTRAINT fk_session_archer_bow FOREIGN KEY (archer_id, bow_id)
        REFERENCES bow (archer_id, bow_id) ON DELETE RESTRICT,
    CONSTRAINT sessions_time_check CHECK (
        closed_at IS NULL OR closed_at > created_at
    )
);

-- Table description
COMMENT ON TABLE session IS 'Shooting session at the range.';

-- Column descriptions
COMMENT ON COLUMN session.session_id IS 'Primary key identifying the session.';
COMMENT ON COLUMN session.archer_id IS
    'Foreign key referencing the archer who shoots the session.';
COMMENT ON COLUMN session.bow_id IS
    'Foreign key referencing the bow used for the session.';
COMMENT ON COLUMN session.session_location IS 'Free-text location description.';
COMMENT ON COLUMN session.is_indoor IS 'True if shooting indoors, false if outdoors.';
COMMENT ON COLUMN session.is_opened IS 'True if session is open for recording shots.';
COMMENT ON COLUMN session.created_at IS 'Timestamp when the session was created.';
COMMENT ON COLUMN session.closed_at IS 'Timestamp when session closed (NULL if active).';
```

In the Down block, `DROP TABLE IF EXISTS session CASCADE;` automatically drops the table,
foreign key constraints, and all column comments.

In `backend/migrations/007_2025-10-28_shot_table.sql`:
Ensure comments are present for `shot`:

```sql
-- Table description
COMMENT ON TABLE shot IS 'Individual arrow shot recorded within a session slot.';

-- Column descriptions
COMMENT ON COLUMN shot.shot_id IS 'Primary key identifying the shot record.';
COMMENT ON COLUMN shot.slot_id IS 'Foreign key referencing slot(slot_id).';
COMMENT ON COLUMN shot.x IS 'Horizontal coordinate on target face in millimeters.';
COMMENT ON COLUMN shot.y IS 'Vertical coordinate on target face in millimeters.';
COMMENT ON COLUMN shot.score IS 'Score value from 0 through 10.';
COMMENT ON COLUMN shot.arrow_id IS
    'Foreign key referencing arrow(arrow_id) (NULL if untagged).';
COMMENT ON COLUMN shot.created_at IS 'Timestamp when the shot was recorded.';
COMMENT ON COLUMN shot.is_x IS 'True if shot hit the inner 10 (X) ring.';
```

- [ ] **Step 4: Run test to verify it passes**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected: `test_session_bow_and_shot_arrow_fks` PASS.

- [ ] **Step 5: Commit**

```bash
cd backend/migrations
git add 006_2025-09-26_shooting_sessions_table.sql 007_2025-10-28_shot_table.sql \
    scripts/run_migration_tests.bash
git commit -m "feat(db): establish session-bow and shot-arrow foreign keys with comments"
```

---

### Task 6: Full Migration Suite Refactoring & Two-Way Lifecycle Verification

**Files:**

- Modify: `backend/migrations/scripts/run_migration_tests.bash`
- Modify: `backend/migrations/README.md`

**Interfaces:**

- Consumes: All migrations in `backend/migrations/*.sql`
- Produces: 100% green test execution and clean migration rollback and re-apply

- [ ] **Step 1: Update all test helper calls in `run_migration_tests.bash`**

Audit and update functions in `backend/migrations/scripts/run_migration_tests.bash`:

1. Ensure `load_env` exports `PGPASSWORD="${POSTGRES_PASSWORD}"`.
2. Ensure `validate_migration_files` validates all `NNN_*.sql` files with no duplicate version
   numbers.
3. Replace obsolete `test_archer_constraints` with checks for the simplified `archer` entity
   (first_name, last_name, date_of_birth, gender, unconditional email uniqueness).
4. Verify `test_goose_rollback_and_reapply` runs `goose down-to 0` and `goose up` cleanly without
   orphaned types or objects.

- [ ] **Step 2: Run full migration test script**

Run: `PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash`
Expected output:

```text
✓ All migration files pass format validation
✓ Applied migrations via goose CLI
...
✓ goose down-to 0 executed successfully
✓ goose up re-applied successfully
✓ All migration tests passed (0 failures)
```

- [ ] **Step 3: Verify Go backend compiles and existing model unit tests pass**

Run: `cd backend && go test ./internal/model/... -v`
Expected: PASS.

Run: `cd backend && golangci-lint run ./...`
Expected: No errors.

- [ ] **Step 4: Commit**

```bash
cd backend/migrations
git add scripts/run_migration_tests.bash
git commit -m "test(db): update full migration test suite for canonical onboarding schema"
```

Update submodule pointer in root repository if required:

```bash
git add backend/migrations
git commit -m \
    "chore(migrations): update migrations submodule with auth, archer, bow, and arrow schema"
```

---

## Verification Plan

### Automated Tests

1. **Migration Suite Test**:

    ```bash
    PGPASSWORD=changeme ./backend/migrations/scripts/run_migration_tests.bash
    ```

    - Validates naming patterns (`NNN_*.sql`), goose markers (`-- +goose Up`, `-- +goose Down`).
    - Runs all assertion test suites covering `auth`, `archer`, `bow`, `arrow`, `auth_session`, FKs,
      and comment presence.
    - Tests rollback (`goose down-to 0`) and re-application (`goose up`).

2. **Go Backend Model Unit Tests**:

    ```bash
    cd backend && go test ./internal/model/... -v
    ```

    - Confirms backend model package remains healthy.

3. **Go Backend Linting**:

    ```bash
    cd backend && golangci-lint run ./...
    ```

    - Confirms no regressions or formatting issues.

### Manual Verification

- Verify database state directly with `psql` by querying `\d+ auth`, `\d+ archer`, `\d+ bow`,
  `\d+ arrow`, `\d+ auth_session`, `\d+ session`, `\d+ shot` to inspect table descriptions,
  column descriptions, constraints, foreign keys, triggers, and partial indexes.
