# Task 001: Database Migrations for Simplified Session and Shot Schema

## Git Branch

`feature/001-db-migrations-simplify-session-shot`

## Objective

Refactor database schema migrations to transition from the legacy multi-archer slot and target
model to the canonical single-archer session and shot model. Absorb slot configuration fields
into `session`, update `shot` to reference `session_id` directly, establish the computed view
`live_stat_by_session_id`, drop obsolete tables (`slot`, `target`, `open_participants`), remove
obsolete database functions and WebSocket LISTEN/NOTIFY triggers, and update migration test suites.

Per project guidelines, there is no requirement to maintain backward compatibility with legacy
migrations; migrations can be added, modified, or replaced to establish the canonical schema.

## Dependencies

- Completion of onboarding schema foundation (tables `auth`, `archer`, `bow`, and `arrow`).

## Acceptance Criteria

- [ ] PostgreSQL migration files in `backend/migrations/` define the canonical target schema:
    - [ ] `session` table:
        - `session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4()`
        - `archer_id UUID NOT NULL REFERENCES archer(archer_id) ON DELETE RESTRICT`
        - `bow_id UUID NOT NULL`
        - Foreign key: `FOREIGN KEY (archer_id, bow_id) REFERENCES bow(archer_id, bow_id) ON DELETE
          RESTRICT`
        - `session_location VARCHAR(255) NOT NULL`
        - `is_indoor BOOLEAN NOT NULL`
        - `distance SMALLINT NOT NULL CHECK (distance BETWEEN 1 AND 100)`
        - `face_type FACE_TYPE NOT NULL`
        - `shots_per_end SMALLINT NOT NULL CHECK (shots_per_end >= 3)`
        - `interval_seconds SMALLINT NOT NULL DEFAULT 20 CHECK (interval_seconds BETWEEN 1 AND 100)`
        - `goal TEXT CHECK (goal IS NULL OR length(goal) <= 2000)`
        - `was_goal_achieved BOOLEAN`
        - `did_well TEXT CHECK (did_well IS NULL OR length(did_well) <= 2000)`
        - `need_work TEXT CHECK (need_work IS NULL OR length(need_work) <= 2000)`
        - `status SESSION_STATUS NOT NULL DEFAULT 'not_rated'`
        - `is_deleted BOOLEAN NOT NULL DEFAULT FALSE`
        - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
        - `closed_at TIMESTAMPTZ`
        - Constraint: `CHECK (closed_at IS NULL OR closed_at > created_at)`
        - Constraint: `CHECK (closed_at IS NULL OR status != 'not_rated')` (closed sessions must be
          rated)
        - Constraint: `CHECK (was_goal_achieved IS NULL OR goal IS NOT NULL)`
        - Unique partial index: `(archer_id) WHERE closed_at IS NULL AND is_deleted = FALSE`
          enforcing one open session per archer.
        - Trigger `trg_check_session_arrow_ceiling` enforcing `shots_per_end <= in_use arrows`.
        - Trigger `trg_protect_closed_session` enforcing immutability of closed sessions.
        - Index on `session(archer_id, created_at DESC) WHERE is_deleted = FALSE`.
        - Index on `session(archer_id, bow_id, distance, face_type, is_indoor) WHERE
          is_deleted = FALSE`.
    - [ ] `shot` table:
        - `shot_id UUID PRIMARY KEY DEFAULT uuid_generate_v4()`
        - `session_id UUID NOT NULL REFERENCES session(session_id) ON DELETE CASCADE`
        - `arrow_id UUID REFERENCES arrow(arrow_id) ON DELETE SET NULL`
        - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
        - `x REAL` (compact 4-byte float)
        - `y REAL` (compact 4-byte float)
        - `score SMALLINT` (compact 2-byte int)
        - `is_x BOOLEAN NOT NULL DEFAULT FALSE`
        - Constraint: `CHECK ((x IS NULL AND y IS NULL AND score IS NULL) OR`
          `(x IS NOT NULL AND y IS NOT NULL AND score IS NOT NULL))`
        - Constraint: `CHECK (score >= 0 AND score <= 10)`
        - Constraint: `CHECK (NOT is_x OR coalesce(score = 10, FALSE))`
        - Constraint: `CHECK (x IS NULL OR (x BETWEEN -1500 AND 1500 AND y BETWEEN -1500 AND 1500))`
        - Trigger `trg_check_shot_insert`: validates open session, arrow ownership, in_use status,
          mandatory tagging on scored sessions when arrows registered, and prohibits arrow_id when
          none registered.
        - Trigger `trg_protect_closed_session_shots`: prevents modifying or deleting shots in closed
          sessions.
        - Index `idx_shot_session_scoring ON shot(session_id) INCLUDE (score, is_x)` for zero-heap
          Index-Only live stats.
        - Index `idx_shot_session_created ON shot(session_id, created_at ASC)` for chronological end
          retrieval.
        - Index `idx_shot_arrow_id ON shot(arrow_id) WHERE arrow_id IS NOT NULL` for arrow
          diagnostics.
    - [ ] Computed view `live_stat_by_session_id`:
        - Left-joins `shot` on `session WHERE face_type != 'none'` so 0-shot sessions return
          initialized 0 values.
        - Calculates `mean`, `max_score`, `number_of_shots`, `number_of_x`, `total_score`
    - [ ] Obsolete database objects dropped:
        - Tables `slot` and `target` dropped
        - View `live_stat_by_slot_id` dropped
        - Materialized view `open_participants` dropped
        - Functions `get_next_lane`, `get_available_targets`, `get_slot_with_lane`,
          `get_active_slot_id` dropped
        - Trigger `shot_insert` and trigger function `notify_shot_insert` dropped
- [ ] `./backend/migrations/scripts/run_migration_tests.bash` updated to test the new schema:
    - [ ] Session creation with bow, distance, face_type, shots_per_end, interval_seconds, goal
    - [ ] Single open session constraint per archer
    - [ ] Scored shots with coordinates, score, and valid arrow ownership
    - [ ] Volume shots with all NULLs for coordinates and score
    - [ ] Live stats view computation matches expected values for session
    - [ ] Obsolete tests for slots, targets, and materialized views removed
- [ ] Goose migrations apply cleanly (`goose up`), roll back cleanly (`goose down-to 0`),
      and re-apply cleanly (`goose up`).

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `backend/migrations/004_2025-09-26_shooting_sessions_table.sql` |
| Modify | `backend/migrations/006_2025-10-28_shot_table.sql` |
| Modify | `backend/migrations/scripts/run_migration_tests.bash` |

## Reference

- [PRD.md](../PRD.md)
- [README.md](../../../backend/migrations/README.md)
- [story_time.md](../../../backend/migrations/story_time.md)

## Steps

- [ ] **Step 1: Refactor session migration in `004_2025-09-26_shooting_sessions_table.sql`**

  Update migration DDL to define the canonical `session` table absorbing slot configuration,
  with composite bow ownership foreign key, triggers, and query indexes:

  ```sql
  -- +goose Up
  DO $$
  BEGIN
      IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'session_status') THEN
          CREATE TYPE session_status AS ENUM ('not_rated', 'bad', 'neutral', 'good');
      END IF;
  END$$;

  CREATE TABLE IF NOT EXISTS session (
      session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE RESTRICT,
      bow_id UUID NOT NULL,
      session_location VARCHAR(255) NOT NULL,
      is_indoor BOOLEAN NOT NULL,
      distance SMALLINT NOT NULL CHECK (distance BETWEEN 1 AND 100),
      face_type FACE_TYPE NOT NULL,
      shots_per_end SMALLINT NOT NULL CHECK (shots_per_end >= 3),
      interval_seconds SMALLINT NOT NULL DEFAULT 20 CHECK (
          interval_seconds BETWEEN 1 AND 100
      ),
      goal TEXT CHECK (goal IS NULL OR length(goal) <= 2000),
      was_goal_achieved BOOLEAN,
      did_well TEXT CHECK (did_well IS NULL OR length(did_well) <= 2000),
      need_work TEXT CHECK (need_work IS NULL OR length(need_work) <= 2000),
      status SESSION_STATUS NOT NULL DEFAULT 'not_rated',
      is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      closed_at TIMESTAMPTZ,
      CONSTRAINT fk_session_archer_bow FOREIGN KEY (archer_id, bow_id)
          REFERENCES bow (archer_id, bow_id) ON DELETE RESTRICT,
      CONSTRAINT sessions_time_check CHECK (
          closed_at IS NULL OR closed_at > created_at
      ),
      CONSTRAINT session_rating_check CHECK (
          closed_at IS NULL OR status != 'not_rated'
      ),
      CONSTRAINT session_goal_achieved_requires_goal CHECK (
          was_goal_achieved IS NULL OR goal IS NOT NULL
      )
  );

  CREATE UNIQUE INDEX IF NOT EXISTS unique_open_session_per_archer
  ON session (archer_id)
  WHERE closed_at IS NULL AND is_deleted = FALSE;

  CREATE INDEX IF NOT EXISTS idx_session_archer_history
  ON session (archer_id, created_at DESC)
  WHERE is_deleted = FALSE;

  CREATE INDEX IF NOT EXISTS idx_session_analytics
  ON session (archer_id, bow_id, distance, face_type, is_indoor)
  WHERE is_deleted = FALSE;

  -- Arrow ceiling trigger
  CREATE OR REPLACE FUNCTION check_session_arrow_ceiling() RETURNS TRIGGER AS $$
  DECLARE
      v_arrow_count INTEGER;
  BEGIN
      SELECT count(*) INTO v_arrow_count
      FROM arrow
      WHERE archer_id = NEW.archer_id AND status = 'in_use' AND is_deleted = FALSE;

      IF v_arrow_count > 0 AND NEW.shots_per_end > v_arrow_count THEN
          RAISE EXCEPTION 'shots_per_end (%) cannot exceed available in_use arrows (%)',
              NEW.shots_per_end, v_arrow_count
              USING ERRCODE = 'check_violation';
      END IF;

      RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  CREATE TRIGGER trg_check_session_arrow_ceiling
  BEFORE INSERT OR UPDATE OF shots_per_end, archer_id ON session
  FOR EACH ROW EXECUTE FUNCTION check_session_arrow_ceiling();

  -- Closed session immutability trigger
  CREATE OR REPLACE FUNCTION protect_closed_session() RETURNS TRIGGER AS $$
  BEGIN
      IF OLD.closed_at IS NOT NULL THEN
          IF NEW.is_deleted IS DISTINCT FROM OLD.is_deleted
             AND NEW.archer_id = OLD.archer_id
             AND NEW.bow_id = OLD.bow_id
             AND NEW.session_location = OLD.session_location
             AND NEW.is_indoor = OLD.is_indoor
             AND NEW.distance = OLD.distance
             AND NEW.face_type = OLD.face_type
             AND NEW.shots_per_end = OLD.shots_per_end
             AND NEW.interval_seconds = OLD.interval_seconds
             AND NEW.goal IS NOT DISTINCT FROM OLD.goal
             AND NEW.was_goal_achieved IS NOT DISTINCT FROM OLD.was_goal_achieved
             AND NEW.did_well IS NOT DISTINCT FROM OLD.did_well
             AND NEW.need_work IS NOT DISTINCT FROM OLD.need_work
             AND NEW.status = OLD.status
             AND NEW.created_at = OLD.created_at
             AND NEW.closed_at = OLD.closed_at THEN
              RETURN NEW;
          END IF;
          RAISE EXCEPTION 'Closed sessions are immutable' USING ERRCODE = 'check_violation';
      END IF;
      RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  CREATE TRIGGER trg_protect_closed_session
  BEFORE UPDATE ON session
  FOR EACH ROW EXECUTE FUNCTION protect_closed_session();
  ```

  Remove definitions for `target`, `slot`, `get_next_lane`, `get_available_targets`,
  `get_slot_with_lane`, `get_active_slot_id`, and `open_participants`. Update `-- +goose Down`.

- [ ] **Step 2: Refactor shot migration in `006_2025-10-28_shot_table.sql`**

  Update migration DDL with compact types (`REAL`, `SMALLINT`), DB triggers, and optimized indexes:

  ```sql
  -- +goose Up
  CREATE TABLE IF NOT EXISTS shot (
      shot_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      session_id UUID NOT NULL REFERENCES session (session_id) ON DELETE CASCADE,
      arrow_id UUID REFERENCES arrow (arrow_id) ON DELETE SET NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      x REAL,
      y REAL,
      score SMALLINT,
      is_x BOOLEAN NOT NULL DEFAULT FALSE,
      CONSTRAINT shot_score_bounds CHECK (score >= 0 AND score <= 10),
      CONSTRAINT shot_is_x_requires_ten CHECK (
          NOT is_x OR coalesce(score = 10, FALSE)
      ),
      CONSTRAINT shot_coords_score_all_or_none CHECK (
          (x IS NULL AND y IS NULL AND score IS NULL)
          OR (x IS NOT NULL AND y IS NOT NULL AND score IS NOT NULL)
      ),
      CONSTRAINT shot_coords_bounds CHECK (
          x IS NULL OR (x BETWEEN -1500 AND 1500 AND y BETWEEN -1500 AND 1500)
      )
  );

  CREATE INDEX IF NOT EXISTS idx_shot_session_scoring ON shot (session_id) INCLUDE (score, is_x);
  CREATE INDEX IF NOT EXISTS idx_shot_session_created ON shot (session_id, created_at ASC);
  CREATE INDEX IF NOT EXISTS idx_shot_arrow_id ON shot (arrow_id) WHERE arrow_id IS NOT NULL;

  CREATE OR REPLACE VIEW live_stat_by_session_id AS
  SELECT
      s.session_id,
      coalesce(avg(sh.score)::DOUBLE PRECISION, 0) AS mean,
      count(sh.shot_id) * 10 AS max_score,
      count(sh.shot_id) AS number_of_shots,
      count(*) FILTER (WHERE sh.is_x) AS number_of_x,
      coalesce(sum(sh.score), 0) AS total_score
  FROM session s
  LEFT JOIN shot sh ON s.session_id = sh.session_id
  WHERE s.face_type != 'none'
  GROUP BY s.session_id;

  -- Shot insertion validation trigger (open session, mandatory arrow tagging, ownership)
  CREATE OR REPLACE FUNCTION check_shot_insert_rules() RETURNS TRIGGER AS $$
  DECLARE
      v_archer_id UUID;
      v_closed_at TIMESTAMPTZ;
      v_face_type face_type;
      v_arrow_count INTEGER;
      v_arrow_valid BOOLEAN;
  BEGIN
      SELECT archer_id, closed_at, face_type
      INTO v_archer_id, v_closed_at, v_face_type
      FROM session
      WHERE session_id = NEW.session_id;

      IF NOT FOUND THEN
          RAISE EXCEPTION 'Session % does not exist', NEW.session_id
          USING ERRCODE = 'foreign_key_violation';
      END IF;

      IF v_closed_at IS NOT NULL THEN
          RAISE EXCEPTION 'Cannot record shot in a closed session (%)', NEW.session_id
          USING ERRCODE = 'check_violation';
      END IF;

      SELECT count(*) INTO v_arrow_count
      FROM arrow
      WHERE archer_id = v_archer_id AND status = 'in_use' AND is_deleted = FALSE;

      IF v_arrow_count > 0 THEN
          IF v_face_type != 'none' AND NEW.arrow_id IS NULL THEN
              RAISE EXCEPTION 'Arrow tagging is mandatory for scored sessions when archer has
              registered arrows'
              USING ERRCODE = 'check_violation';
          END IF;

          IF NEW.arrow_id IS NOT NULL THEN
              SELECT EXISTS (
                  SELECT 1 FROM arrow
                  WHERE arrow_id = NEW.arrow_id
                    AND archer_id = v_archer_id
                    AND status = 'in_use'
                    AND is_deleted = FALSE
              ) INTO v_arrow_valid;

              IF NOT v_arrow_valid THEN
                  RAISE EXCEPTION 'Arrow % does not belong to session archer or is not in_use',
                  NEW.arrow_id
                  USING ERRCODE = 'check_violation';
              END IF;
          END IF;
      ELSE
          IF NEW.arrow_id IS NOT NULL THEN
              RAISE EXCEPTION 'Arrow tagging is disabled because archer has no registered arrows'
                  USING ERRCODE = 'check_violation';
          END IF;
      END IF;

      RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  CREATE TRIGGER trg_check_shot_insert
  BEFORE INSERT ON shot
  FOR EACH ROW EXECUTE FUNCTION check_shot_insert_rules();

  -- Shot closed session protection trigger
  CREATE OR REPLACE FUNCTION protect_closed_session_shots() RETURNS TRIGGER AS $$
  DECLARE
      v_closed_at TIMESTAMPTZ;
  BEGIN
      SELECT closed_at INTO v_closed_at FROM session WHERE session_id = OLD.session_id;
      IF v_closed_at IS NOT NULL THEN
          RAISE EXCEPTION 'Cannot modify or delete shots in a closed session'
          USING ERRCODE = 'check_violation';
      END IF;
      RETURN COALESCE(NEW, OLD);
  END;
  $$ LANGUAGE plpgsql;

  CREATE TRIGGER trg_protect_closed_session_shots
  BEFORE UPDATE OR DELETE ON shot
  FOR EACH ROW EXECUTE FUNCTION protect_closed_session_shots();
  ```

- [ ] **Step 3: Update `run_migration_tests.bash`**

  Rewrite test cases in `backend/migrations/scripts/run_migration_tests.bash`:
    - Test session happy path with all configuration columns
    - Test `unique_open_session_per_archer` partial unique index
    - Test shot constraints: bounds, is_x requires 10, all-or-none coordinates
    - Test arrow ownership check against session owner
    - Test `live_stat_by_session_id` view aggregation for scored and volume shots
    - Remove all obsolete tests calling dropped functions and slot tables

- [ ] **Step 4: Execute migration test suite**

  ```bash
  ./backend/migrations/scripts/run_migration_tests.bash
  ```

  Verify all tests pass and exit code is 0.

- [ ] **Step 5: Commit changes**

  ```bash
  git add backend/migrations/
  git commit -m "feat(db): simplify session and shot schema to single-archer model"
  ```

## Verification

- `./backend/migrations/scripts/run_migration_tests.bash` succeeds with 0 failures.
- No references to `slot`, `target`, or `open_participants` remain in active migrations.
