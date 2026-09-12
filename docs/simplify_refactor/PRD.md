# PRD: Removing Obsolete Code & Refactoring the Session Flow

## Background

Arch-Stats was originally designed as a **club-oriented, multi-archer** application where a session
owner (range officer) created sessions, archers joined via slots on shared targets/lanes, and
real-time WebSocket updates kept everyone in sync.

The application has since been redefined as a **personal tool for a single archer**. One archer logs
in, tracks their own shooting, and reviews their own history. There are no clubs, no shared sessions
, no managing other people.

This pivot makes several subsystems in the current Go backend obsolete. They add complexity, surface
area for bugs, and cognitive overhead without serving any user need. This PRD documents what must be
removed and how the session flow must be restructured to match the new single-archer model.

### Source Documents

- [story_time.md](../../backend/migrations/story_time.md) the definitive user-facing specification.
- [README.md](../../backend/migrations/README.md) the technical database schema specification.
- [high_level_refactoring_plan.md](../../docs/go_refactor/high_level_refactoring_plan.md) the Go
  refactoring plan from the Python-to-Go port.

## 1. Obsolete Subsystems to Remove

The following subsystems exist in the current codebase but have been **deliberately excluded** from
the product scope (see the "Out of Scope" table in
[README.md](../../backend/migrations/README.md#out-of-scope)).

### 1.1 Slot System

The slot system allowed multiple archers to join a session at different target positions (A/B/C/D
lane divisions). A solo archer doesn't need lane slots.

**Files to remove:**

| Layer | File | Lines |
| ----- | ---- | ----- |
| Model | [slot.go](../../backend/internal/model/slot.go) | 126 |
| Handler | [slot.go](../../backend/internal/handler/slot.go) | ~200 |
| Handler test | [slot_test.go](../../backend/internal/handler/slot_test.go) | ~620 |
| Service | [slot.go](../../backend/internal/service/slot.go) | ~210 |
| Service test | [slot_test.go](../../backend/internal/service/slot_test.go) | ~520 |
| Repository | [slot.go](../../backend/internal/repository/slot.go) | ~215 |
| Repository test | [slot_test.go](../../backend/internal/repository/slot_test.go) | ~580 |

**Enums to remove from [enums.go](../../backend/internal/model/enums.go):**

- `SlotLetter` type and its constants (`SlotLetterA` through `SlotLetterD`)

### 1.2 Target System

The target system represented physical target butts with lane numbers and distances. In the new
model, distance and face type are properties of the session itself there is one distance and one
face per session, no need for a separate target entity.

**Files to remove:**

| Layer | File |
| ----- | ---- |
| Model | [target.go](../../backend/internal/model/target.go) |
| Handler | [slot.go](../../backend/internal/handler/slot.go) (if targets are handled here) |
| Service | [target.go](../../backend/internal/service/target.go) (if exists) |
| Repository | [target.go](../../backend/internal/repository/target.go) |
| Repository test | [target_test.go](../../backend/internal/repository/target_test.go) |

### 1.3 WebSocket / Real-Time Infrastructure

With a single-user app, the person entering scores is viewing them. There is no need for WebSocket
push or PostgreSQL LISTEN/NOTIFY. Standard HTTP request/response is sufficient.

**Files to remove:**

| Layer | File |
| ----- | ---- |
| WebSocket hub | [hub.go](../../backend/internal/websocket/hub.go) |
| WebSocket hub test | [hub_test.go](../../backend/internal/websocket/hub_test.go) |
| WebSocket client | [client.go](../../backend/internal/websocket/client.go) |
| WebSocket client test | [client_test.go](../../backend/internal/websocket/client_test.go) |
| Live stats handler | [live_stats.go](../../backend/internal/handler/live_stats.go) |
| Live stats handler test | [live_stats_test.go](../../backend/internal/handler/live_stats_test.go) |

**Enums to remove from [enums.go](../../backend/internal/model/enums.go):**

- `WSContentType` type and its constants (`WSContentTypeShotCreated`, `WSContentTypeShotDeleted`,
  `WSContentTypeArrowCreated`, `WSContentTypeArrowDeleted`)

### 1.4 Materialized View Maintenance

The `open_participants` materialized view was for the multi-archer model. With no slots or shared
sessions, it is no longer needed.

**Files to remove:**

| Layer | File |
| ----- | ---- |
| Repository | [maintenance.go](../../backend/internal/repository/maintenance.go) |
| Repository test | [maintenance_test.go](../../backend/internal/repository/maintenance_test.go) |
| Repository | [reporting.go](../../backend/internal/repository/reporting.go) |
| Repository test | [reporting_test.go](../../backend/internal/repository/reporting_test.go) |

### 1.5 Live Stats Model

The current [live_stats.go](../../backend/internal/model/live_stats.go)
model will be replaced. Live stats become a standard HTTP endpoint (not WebSocket-driven). The
computed view `live_stat_by_session_id` in the database remains, but the delivery mechanism changes
from WebSocket push to HTTP GET.

### 1.6 Obsolete Registration Code and Endpoints

The legacy monolithic registration endpoint `POST /api/v0/auth/register` and its coupled request
struct `AuthRegistrationRequest` (which mixed personal demographics with equipment and clubs)
have been superseded by the multi-step onboarding flow (`POST /api/v0/archers`, `POST /api/v0/bows`,
and `POST /api/v0/arrows`). This obsolete endpoint, its supporting auth service logic, frontend
dead code, and legacy tests are removed during this refactoring.

**Code to remove:**

| Layer | Item | Location |
| ----- | ---- | -------- |
| Model | `AuthRegistrationRequest` struct | [auth.go](../../backend/internal/model/auth.go) |
| Handler | `Register` handler & docs | [auth.go](../../backend/internal/handler/auth.go) |
| Service | `RegisterWithGoogle`, `Register` | [service.go](../../backend/internal/auth/service.go) |
| Router | `/auth/register` route wiring | [main.go](../../backend/cmd/arch-stats/main.go) |
| Frontend | `registerNewArcher` composable method | [useAuth.ts](../../frontend/src/composables/useAuth.ts) |

## 2. Session Model Refactoring

The current session model is thin it stores only location, indoor/outdoor, and open/closed state.
Most session configuration (bow, face type, distance, shots per end) currently lives on
the **slot**, because the old model supported multiple archers with different settings per slot.

In the new single-archer model, all of these become **session-level properties** since there is
exactly one archer per session.

### 2.1 Current Session Model

From [session.go](../../backend/internal/model/session.go):

```go
type SessionCreate struct {
    OwnerArcherID   uuid.UUID
    SessionLocation string
    IsIndoor        bool
    IsOpened        bool
}
```

### 2.2 Target Session Model

From the [database spec](../../backend/migrations/README.md#session):

| Field | Type | Source | Notes |
| ----- | ---- | ------ | ----- |
| session_id | UUID (PK) | Auto | Unchanged |
| archer_id | UUID (FK) | Replaces `owner_archer_id` | Simpler name; the archer IS the owner |
| bow_id | UUID (FK) | **New** (absorbed from slot) | Which bow the archer is using |
| session_location | VARCHAR(255) | Existing | Unchanged |
| is_indoor | BOOLEAN | Existing | Unchanged |
| distance | SMALLINT | **New** (absorbed from slot/target) | Target distance in meters (1-100) |
| face_type | FACE_TYPE | **New** (absorbed from slot) | Target face; `none` = volume session |
| shots_per_end | SMALLINT | **New** (absorbed from slot) | Arrows per end (>= 3; upper limit <= registered arrows if any) |
| interval_seconds | SMALLINT | **New** (absorbed from slot) | Seconds between shot timestamps (1-100, default 20) |
| goal | TEXT | **New** | Optional session goal (up to 2000 chars) |
| was_goal_achieved | BOOLEAN | **New** | Set on close; only when goal was set |
| did_well | TEXT | **New** | Reflection: what went well (up to 2000 chars) |
| need_work | TEXT | **New** | Reflection: what to improve (up to 2000 chars) |
| status | SESSION_STATUS | **New** | Session rating: default 'not_rated'; mandatory 'bad', 'neutral', or 'good' on close |
| is_deleted | BOOLEAN | **New** | Soft delete flag (default false) |
| created_at | TIMESTAMPTZ | Existing | Unchanged |
| closed_at | TIMESTAMPTZ | Existing | NULL = open session |

**Fields removed:**

- `owner_archer_id` → renamed to `archer_id` (the "owner" concept is redundant for a single-archer
  app)
- `is_opened` → removed entirely; open/closed state is derived from `closed_at IS NULL`

**Key constraints:**

- Unique partial index: `(archer_id) WHERE closed_at IS NULL AND is_deleted = FALSE` enforces one
  open session per archer
- `closed_at IS NULL OR closed_at > created_at`
- `closed_at IS NULL OR status != 'not_rated'` (closed sessions must be rated)
- `distance BETWEEN 1 AND 100`
- `shots_per_end >= 3`:
    - When the archer **has registered arrows**: `shots_per_end >= 3 AND shots_per_end <=
      number_of_in_use_arrows` (enforced by session DB table and application logic on session
      creation)
    - When the archer **does NOT have registered arrows**: `shots_per_end >= 3` (enforced via
      database CHECK constraint)
- `archer_id` references `archer(archer_id)` ON DELETE RESTRICT
- `(archer_id, bow_id)` references `bow(archer_id, bow_id)` ON DELETE RESTRICT (declaratively
  guarantees bow ownership)
- Arrow ceiling rule is enforced at the database layer via `trg_check_session_arrow_ceiling` trigger
- Closed session immutability is enforced at the database layer via `trg_protect_closed_session`
  trigger

### 2.3 Session Type Derivation

There is no explicit session type field. The session type is **derived from `face_type`**:

- Any face type other than `none` → **Scored session** (shots have x/y coordinates and calculated
  scores)
- `face_type = none` → **Volume session** (only arrow counts are tracked, no scoring)

### 2.4 Session Lifecycle

1. **Create**: Archer provides bow, location, indoor/outdoor, distance, face_type, shots_per_end,
   interval_seconds, and optional goal. System sets session_id, archer_id, created_at, and
   status='not_rated'.
2. **Shoot**: Archer records ends (scored or volume). Session must be open (`closed_at IS NULL`).
3. **Close**: System sets `closed_at`. The archer is asked "How was the session?" and must select
   one of `bad`, `neutral`, `good` (`status` is mandatory). Optionally, archer provides reflection
   (`was_goal_achieved`, `did_well`, `need_work`). Session becomes **read-only**.
4. **Delete (soft)**: Sets `is_deleted = true`. Session is hidden from lists and analytics but data
   is preserved.

## 3. Shot Model Refactoring

### 3.1 Current Shot Model

From [shot.go](../../backend/internal/model/shot.go), shots
currently reference a **slot_id** (because in the old model, each archer's position in a session was
a slot):

```go
type ShotCreate struct {
    SlotID    uuid.UUID
    X         *float64
    Y         *float64
    IsX       bool
    Score     *int
    ArrowID   *uuid.UUID
    CreatedAt *time.Time
}
```

### 3.2 Target Shot Model

From the [database spec](../../backend/migrations/README.md#shot):

| Field | Type | Notes |
| ----- | ---- | ----- |
| shot_id | UUID (PK) | Unchanged |
| session_id | UUID (FK) | **Replaces `slot_id`** shots belong directly to sessions |
| arrow_id | UUID (FK) | Nullable; conditional on arrow registration |
| x | REAL | Nullable; NULL for volume sessions (compact 4-byte float) |
| y | REAL | Nullable; NULL for volume sessions (compact 4-byte float) |
| score | SMALLINT | Nullable (0-10); NULL for volume sessions (compact 2-byte int) |
| is_x | BOOLEAN | True if inner 10 ring; default false |
| created_at | TIMESTAMPTZ | Spaced by session's interval_seconds |

**Key change:** `slot_id` → `session_id`. Since there is no slot indirection, shots link directly to
the session.

**Constraints:**

- All-or-none: `(x IS NULL AND y IS NULL AND score IS NULL) OR (x IS NOT NULL AND y IS NOT NULL AND
  score IS NOT NULL)`
- `score >= 0 AND score <= 10`
- `is_x` can only be true when `score = 10`
- `session_id` references `session(session_id)` ON DELETE CASCADE
- `arrow_id` references `arrow(arrow_id)` ON DELETE SET NULL

**Database-Enforced Logic & Rules:**

- **Arrow tagging translation & DB validation**:
    - Archers identify arrows by `arrow_number` (e.g., 1, 2, 3) within their registered `arrow_set`
      (SMALLINT).
    - The frontend translates `arrow_set` + `arrow_number` to `arrow_id` before sending the payload.
    - **Trigger `trg_check_shot_insert`** strictly enforces at the DB level:
        - If the archer has registered arrows: tagging is **mandatory** on every scored shot, and
          `arrow_id` must belong to the session archer and have `status = 'in_use'` and
          `is_deleted = FALSE`.
        - If the archer has NOT registered arrows: `arrow_id` must be NULL on all shots.
        - Shots can only be inserted into open sessions (`closed_at IS NULL`).
    - **Trigger `trg_protect_closed_session_shots`** prevents modifying or deleting shots once a
      session is closed.

### 3.3 End Grouping

There is **no separate "end" entity**. Ends are derived implicitly: with a fixed `shots_per_end` of
N, the first N shots (ordered by `created_at`) form end 1, the next N form end 2, and so on.

## 4. Live Stats (Delivery Change)

The computed view `live_stat_by_session_id` remains in the database. However, the delivery mechanism
changes:

| Aspect | Old (remove) | New |
| ------ | ------------ | --- |
| Transport | WebSocket push | Standard HTTP GET |
| Trigger | PostgreSQL NOTIFY on shot insert/delete | Client polls or fetches on demand |
| Model | [live_stats.go](../../backend/internal/model/live_stats.go) | Needs update to match new view schema |

The live stats view returns:

| Column | Type | Description |
| ------ | ---- | ----------- |
| session_id | UUID | The session |
| mean | DOUBLE PRECISION | Average score per shot |
| max_score | INTEGER | Maximum achievable score (shot count * 10) |
| number_of_shots | INTEGER | Total shots recorded |
| number_of_x | INTEGER | Count of X shots (inner 10) |
| total_score | INTEGER | Sum of all scores |

## 5. Enum Cleanup

### Remove entirely

| Enum | Reason |
| ---- | ------ |
| `SlotLetter` (A/B/C/D) | No slots |
| `WSContentType` | No WebSocket |

### Keep / Add enums

| Enum | Values | Notes |
| ---- | ------ | ----- |
| `Gender` | male, female, non_binary, other, unspecified | Unchanged |
| `Bowstyle` | recurve, compound, barebow, longbow | Unchanged |
| `FaceType` | wa_40cm_full, wa_60cm_full, wa_80cm_full, wa_122cm_full, wa_40cm_6rings, wa_60cm_6rings, wa_80cm_6rings, wa_122cm_6rings, wa_40cm_triple_vertical, wa_60cm_triple_triangular, none | Unchanged |
| `AuthStatus` | authenticated, needs_registration | Unchanged |
| `ArrowStatus` | in_use, damaged, lost | Added: arrow availability condition (default in_use) |
| `SessionStatus` | not_rated, bad, neutral, good | Replaces legacy open/closed enum; rating on close |

## 6. Handler / Route Cleanup

### Routes to remove

| Route prefix | Handler file | Reason |
| ------------ | ------------ | ------ |
| `/api/v0/slots/*` | `handler/slot.go` | No slots |
| `/api/v0/live-stats/*` (WebSocket) | `handler/live_stats.go` | WebSocket removed |
| `/api/v0/auth/register` | `handler/auth.go` | Replaced by multi-step onboarding endpoints (`/archers`, `/bows`, `/arrows`) |

### Routes to update

| Route | Change |
| ----- | ------ |
| Session CRUD | Absorb slot fields into session create/update payloads |
| Shot CRUD | Replace `slot_id` references with `session_id` |
| Live stats | New HTTP GET endpoint replacing WebSocket delivery |

## 7. Overlap & Dependency with Archer Onboarding & Auth Refactor

There is direct domain, schema, and workflow overlap between this session flow refactoring and
the [Archer Onboarding Flow & Auth/Archer Refactor PRD](../onboarding_refactor/PRD.md). When
moving to execution, these dependencies are coordinated as follows:

| Overlapping Area | Relationship / Dependency | Implementation & Planning Consideration |
| ---------------- | ------------------------- | --------------------------------------- |
| **`session.bow_id`** | Session references `bow(bow_id)`, registered during onboarding | The `bow` table and Create/List endpoints must exist before session creation can link to a bow |
| **`shot.arrow_id` & Tagging** | Arrow registration is an **optional step** during onboarding | If the archer registers arrows, scored shots require `arrow_number` selection (translated by frontend to `arrow_id`). If the archer skips arrow registration, `arrow_id = NULL` on all shots. |
| **`shots_per_end` constraint** | Conditional on arrow registration | Enforces `shots_per_end <= number_of_registered_arrows` if arrows were registered; defaults to `shots_per_end >= 3` if skipped. |
| **Identity / Auth Cleanup** | Both PRDs standardize on single-archer ownership (`archer_id`) | Database migrations coordinate splitting `auth`/`archer` and creating `bow`/`arrow` tables alongside dropping obsolete `slot`, `target`, and `club` tables. |

> [!NOTE]
> **Implementation Sequencing Strategy**:
>
> 1. **Schema Migration**: Establish `bow` and `arrow` tables, separate `auth`/`archer`, and drop
>    obsolete `slot`/`target`/`club` structures.
> 2. **Onboarding Flow**: Implement Google auth separation, archer profile creation
>    (`POST /api/v0/archers`), first bow registration (`POST /api/v0/bows`), and optional arrow
>    registration (`POST /api/v0/arrows`).
> 3. **Session & Shot Refactor**: Implement the simplified Session and Shot flows linking
>    to `bow_id` and conditionally handling `arrow_id` and `shots_per_end` based on registered
>    arrows.
> 4. **Future Extension**: Dedicated Archer Profile Page and full equipment editing/management
>    UI can be added as a separate feature later.
