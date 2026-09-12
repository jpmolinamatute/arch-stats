# PRD: Archer Onboarding Flow & Auth/Archer Refactor

## Background

This PRD covers **two lean, interrelated workstreams**:

1. **Refactoring existing code** - the current archer model mixes authentication data, personal
   information, and equipment preferences (bowstyle, draw weight) in a single entity. This must be
   separated into distinct, purpose-built entities: `auth` (system-managed credentials), `archer`
   (personal info only), `bow` (equipment), and `arrow` (optional equipment).
2. **Archer Onboarding Flow** - a streamlined first-time registration flow guiding newly
   authenticated users from Google Sign-In through:
   - Completing their personal profile (mandatory)
   - Registering their first bow (mandatory)
   - Registering their arrows (optional step before completing onboarding)
   - Landing on the app ready to shoot.

> [!NOTE]
> **Scope Boundary & Lean Focus:**
> A dedicated **Archer Profile Page** (for viewing/editing personal details post-onboarding,
> managing multiple bows, or editing/deleting arrows) is **deferred to a future feature**. To keep
> this workstream lean, focused, and immediately deliverable, this PRD concentrates strictly on the
> **onboarding flow** (including optional arrow registration) and the required backend refactoring
> to support it.

The new single-archer model introduces a clean separation of concerns:

- **Auth** - authentication credentials managed by the system from Google OAuth (never user-edited)
- **Archer** - personal information collected during onboarding (`POST /api/v0/archers`)
- **Bow** - equipment registered during onboarding (at least one required before shooting)
- **Arrow** - optional equipment registered during onboarding (enables per-arrow diagnostics)

### Source Documents

- [story_time.md](../../backend/migrations/story_time.md) the
  definitive user-facing specification (sections: Archer Profile, Onboarding).
- [README.md](../../backend/migrations/README.md) the technical
  database schema specification (entities: auth, archer, bow, arrow).

## 1. Auth / Archer Separation

### 1.1 Current State

The current [archer model](../../backend/internal/model/archer.go)
mixes authentication and profile data in one struct:

```go
type ArcherCreate struct {
    FirstName        string      // personal
    LastName         string      // personal
    Email            string      // personal
    DateOfBirth      string      // personal
    Gender           Gender      // personal
    Bowstyle         Bowstyle    // equipment (moving to Bow)
    DrawWeight       float64     // equipment (moving to Bow)
    ClubID           *uuid.UUID  // club (removing)
    GooglePictureURL *string     // auth (moving to Auth)
    GoogleSubject    string      // auth (moving to Auth)
}
```

### 1.2 Target State: Auth Entity

A new `auth` entity stores authentication and identity data from the Google OAuth provider. It is
**system-managed** and never edited directly by the archer.

From the [database spec](../../backend/migrations/README.md#auth):

| Field | Type | Nullable | Default | Description |
| ----- | ---- | -------- | ------- | ----------- |
| archer_id | UUID (PK) | NO | auto | Primary key (auto-generated on first OAuth sign-in) |
| google_subject | TEXT | NO | | Google account identifier (unique) |
| google_picture_url | TEXT | YES | | Profile photo URL from Google |
| last_login_at | TIMESTAMPTZ | NO | now() | Timestamp of most recent sign-in |
| created_at | TIMESTAMPTZ | NO | now() | Timestamp of account creation |

**Constraints:**

- `google_subject` is UNIQUE

**Relationships:**

- Has one `archer` (1:1 via `archer_id`, populated when the archer completes onboarding)
- Has many `auth_session` (1:many via `archer_id` for hashed session token tracking)

**What moves here from the current archer model:**

- `google_subject` → `auth.google_subject`
- `google_picture_url` → `auth.google_picture_url`
- `last_login_at` → `auth.last_login_at`

### 1.3 Target State: Archer Entity

The archer entity becomes **purely personal information**, provided by the user during the
onboarding flow. Editing profile fields post-onboarding is deferred to the future Profile Page
feature. The `archer_id` is supplied by the application using the `archer_id` from the existing
`auth` record.

From the [database spec](../../backend/migrations/README.md#archer):

| Field | Type | Nullable | Default | Description |
| ----- | ---- | -------- | ------- | ----------- |
| archer_id | UUID (PK, FK) | NO | | Shared PK; references `auth(archer_id)` (provided by app) |
| email | VARCHAR(100) | NO | | Contact email (unique) |
| first_name | VARCHAR(100) | NO | | First name |
| last_name | VARCHAR(100) | NO | | Last name |
| date_of_birth | DATE | NO | | Date of birth |
| gender | GENDER_TYPE | NO | | Gender identity |
| is_deleted | BOOLEAN | NO | false | Soft delete flag |

**Constraints:**

- `email` is UNIQUE
- `date_of_birth <= current_date - interval '10 years'` (archer must be at least 10 years old)

**Relationships:**

- `archer_id` references `auth(archer_id)` ON DELETE CASCADE (1:1, shared primary key)
- Has many `bow` (at least 1 required at application level)
- Has many `arrow` (optional)
- Has many `session`

**Fields removed from the current archer model:**

| Field | Reason |
| ----- | ------ |
| `bowstyle` | Moves to the new `bow` entity |
| `draw_weight` | Moves to the new `bow` entity |
| `club_id` | Clubs are out of scope for the personal app |
| `google_subject` | Moves to the new `auth` entity |
| `google_picture_url` | Moves to the new `auth` entity |
| `last_login_at` | Moves to the new `auth` entity |
| `created_at` | Moves to the new `auth` entity (account creation timestamp) |

## 2. New Entity: Bow

Archers register the bows they own. **At least one bow must be registered** before the archer can
start a session. Each session references exactly one bow.

From the [database spec](../../backend/migrations/README.md#bow):

| Field | Type | Nullable | Default | Description |
| ----- | ---- | -------- | ------- | ----------- |
| bow_id | UUID (PK) | NO | auto | Primary key |
| archer_id | UUID (FK) | NO | | Owner archer |
| name | VARCHAR(255) | NO | | Archer-given label (e.g., "My competition recurve") |
| bowstyle | BOWSTYLE_TYPE | NO | | Type of bow: recurve, compound, barebow, longbow |
| draw_weight | REAL | NO | | Draw weight in pounds (compact 4-byte float) |
| is_deleted | BOOLEAN | NO | false | Soft delete flag |
| created_at | TIMESTAMPTZ | NO | now() | When the bow was registered |

**Constraints:**

- `draw_weight > 0 AND draw_weight <= 200`
- `length(trim(name)) > 0`
- `UNIQUE (archer_id, bow_id)` (enables declarative cross-entity ownership foreign key from session)

**Relationships:**

- `archer_id` references `archer(archer_id)` ON DELETE CASCADE
- Referenced by `session(archer_id, bow_id)` each session uses exactly one bow owned by the archer

### 2.1 Bow Operations for Onboarding

To support the onboarding flow and enable session creation, the following bow operations are in
scope:

| Operation | Endpoint | Notes |
| --------- | -------- | ----- |
| Create | POST /api/v0/bows | Archer registers at least one bow during onboarding step 4 |
| List | GET /api/v0/bows | List all bows for the authenticated archer (used when choosing bow for a session) |

> [!NOTE]
> Additional bow management operations (updating bow details via `PATCH`, deleting bows via `DELETE`
> , or single bow view `GET /api/v0/bows/{bow_id}`) are **deferred to the future Profile & Equipment
> feature** to keep onboarding lean.

## 3. Entity: Arrow (Optional Onboarding Step)

Archers have the option to register their individual arrows after registering their bow(s) and
before finishing onboarding. Registering arrows is **optional**; archers can choose to register them
or skip this step. Arrows are grouped into **arrow sets** (e.g., grouping 3 or more arrows used with
different bows or for indoor vs. outdoor shooting).

From the [database spec](../../backend/migrations/README.md#arrow):

| Field | Type | Nullable | Default | Description |
| ----- | ---- | -------- | ------- | ----------- |
| arrow_id | UUID (PK) | NO | auto | Primary key |
| archer_id | UUID (FK) | NO | | Owner archer |
| arrow_set | SMALLINT | NO | | Arrow set / group identifier (groups 3+ arrows, numbered 1, 2, ...) |
| arrow_number | SMALLINT | NO | | Identifying number within the set |
| spine | REAL | YES | | Arrow stiffness rating (compact 4-byte float) |
| length | REAL | YES | | Arrow length in inches (compact 4-byte float) |
| weight | REAL | YES | | Arrow weight in grains (compact 4-byte float) |
| status | ARROW_STATUS | NO | 'in_use' | Availability status (in_use, damaged, lost) |
| is_deleted | BOOLEAN | NO | false | Soft delete flag |
| created_at | TIMESTAMPTZ | NO | now() | When the arrow was registered |

> [!NOTE]
> **Arrow Status during Onboarding:**
> All arrows registered during onboarding are automatically assigned `status = 'in_use'`. The
> onboarding wizard does not prompt the archer for arrow status.

**Constraints:**

- `(archer_id, arrow_set, arrow_number)` is UNIQUE WHERE is_deleted = FALSE (an archer can have the
  same arrow number as long as they belong to different arrow sets)
- An arrow set must group at least 3 arrows together (application-level validation)

**Relationships:**

- `archer_id` references `archer(archer_id)` ON DELETE CASCADE
- Referenced by `shot(arrow_id)` ON DELETE SET NULL

### 3.1 Arrow Operations for Onboarding

To support registering arrows during onboarding and retrieving them for sessions:

| Operation | Endpoint | Notes |
| --------- | -------- | ----- |
| Create | POST /api/v0/arrows | Archer registers arrows (specifying `arrow_set`, `arrow_number`, etc.) during the optional onboarding step |
| List | GET /api/v0/arrows | List all arrows for authenticated archer (used in session setup & shot tagging) |

> [!NOTE]
> Additional arrow management operations (updating arrow specs via `PATCH`, deleting arrows via
> `DELETE`, or single arrow view `GET /api/v0/arrows/{arrow_id}`) are **deferred to the future
> Profile & Equipment feature** to keep onboarding lean.

### 3.2 Impact on Session & Shot Workflows

Registering arrows during onboarding activates specific session behaviors:

- **Arrow ceiling rule:**
    - If the archer **has registered arrows**: `shots_per_end >= 3 AND shots_per_end <=
      number_of_in_use_arrows`.
    - If the archer **skipped arrow registration** (0 arrows): `shots_per_end >= 3` with no arrow
      count ceiling.
- **Arrow tagging rule:**
    - If the archer **has registered arrows**: archers identify arrows by `arrow_set` and
      `arrow_number` in the UI; the frontend translates `arrow_set` + `arrow_number` to `arrow_id`
      for persistence. Only `in_use` arrows are presented for tagging. Arrow tagging is
      **mandatory** on all scored shots.
    - If the archer **skipped arrow registration**: `shot.arrow_id` is stored as `NULL`, and arrow
      tagging UI is omitted.

## 4. Onboarding Flow

When an archer signs in with Google for the first time, the system must guide them through a
multi-step onboarding before they can start shooting.

### 4.1 Flow Steps

```text
Google Sign-In
    │
    ▼
┌──────────────────────────────┐
│ 1. Auth record created       │  System creates auth row with google_subject,
│    (system-managed)          │  google_picture_url, last_login_at
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│ 2. AuthStatus check          │  API returns "needs_registration" if no
│                              │  archer profile exists for this archer_id
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│ 3. Complete profile           │  Archer provides: email, first_name,
│    (mandatory)               │  last_name, date_of_birth, gender
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│ 4. Register at least one bow │  Archer provides: name, bowstyle,
│    (mandatory)               │  draw_weight
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│ 5. Register arrows           │  Optional: archer can register arrows
│    (optional - can skip)     │  (arrow_set, arrow_number, spine,
│                              │  length, weight; defaults to in_use)
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│ 6. Open shooting session     │  Archer is authenticated and ready
│                              │  to create a shooting session
└──────────────────────────────┘
```

### 4.2 Auth Status Response

The existing `AuthStatus` enum stays:

- `authenticated` archer profile exists, ready to use the app
- `needs_registration` auth record exists but no archer profile yet; redirect to onboarding

### 4.3 Profile Completeness Check

Before allowing session creation, the application must verify:

1. Archer profile exists (all required personal fields filled)
2. At least one bow is registered

> [!NOTE]
> Arrows are **optional**. An archer who skipped arrow registration has 0 arrows registered, which
> is completely valid and allows session creation. If arrows are registered, they can immediately be
> used for arrow tagging.

## 5. Model Changes Summary

### 5.1 Files to Create (In Scope)

| Layer | File | Purpose |
| ----- | ---- | ------- |
| Model | `internal/model/bow.go` | BowCreate, BowRead, BowSet, BowFilter structs |
| Model | `internal/model/arrow.go` | ArrowCreate, ArrowRead, ArrowSet, ArrowFilter structs |
| Handler | `internal/handler/bow.go` | Bow Create and List endpoints |
| Handler | `internal/handler/arrow.go` | Arrow Create and List endpoints |
| Service | `internal/service/bow.go` | Bow creation and retrieval business logic |
| Service | `internal/service/arrow.go` | Arrow creation, retrieval, and arrow ceiling/count logic |
| Repository | `internal/repository/bow.go` | Bow database queries |
| Repository | `internal/repository/arrow.go` | Arrow database queries |

### 5.2 Files to Modify

| Layer | File | Change |
| ----- | ---- | ------ |
| Model | [archer.go](../../backend/internal/model/archer.go) | Remove bowstyle, draw_weight, club_id, google_subject, google_picture_url, last_login_at, created_at |
| Model | [auth.go](../../backend/internal/model/auth.go) | Add google_subject, google_picture_url, last_login_at, created_at fields |
| Handler | [auth.go](../../backend/internal/handler/auth.go) | Update auth flow to create auth record separately from archer profile |
| Handler | [archer.go](../../backend/internal/handler/archer.go) | Update create payload to match simplified archer model |
| Service | [archer.go](../../backend/internal/service/archer.go) | Add profile completeness check (profile + at least 1 bow) |
| Repository | [archer.go](../../backend/internal/repository/archer.go) | Update queries to match new schema |

### 5.3 Migration Files

New SQL migrations needed for this workstream:

1. Create the `bow` table (`bow_id`, `archer_id`, `name`, `bowstyle`, `draw_weight`, `is_deleted`,
   `created_at`)
2. Create the `arrow` table (`arrow_id`, `archer_id`, `arrow_set`, `arrow_number`, `spine`, `length`
   , `weight`, `status`, `is_deleted`, `created_at`)
3. Move auth fields from `archer` to `auth` table
4. Update `archer` table (add `is_deleted` column; remove legacy `bowstyle`, `draw_weight`,
   `club_id` columns)
5. Add foreign key from `session.bow_id` to `bow(bow_id)`
6. Add foreign key from `shot.arrow_id` to `arrow(arrow_id)` ON DELETE SET NULL

## 6. API Surface

### 6.1 In Scope (Onboarding & Auth Refactor)

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | /api/v0/auth/status | Returns `needs_registration` or `authenticated` |
| POST | /api/v0/archers | Creates initial archer profile during onboarding step 3 |
| GET | /api/v0/archers/{archer_id} | Retrieves archer profile |
| POST | /api/v0/bows | Registers first bow during onboarding step 4 |
| GET | /api/v0/bows | Lists bows for authenticated archer (used in session creation) |
| POST | /api/v0/arrows | Registers an arrow during optional onboarding step 5 |
| GET | /api/v0/arrows | Lists arrows for authenticated archer (used for ceiling & tagging) |

### 6.2 Deferred to Future Features

| Method | Path | Deferred Feature | Reason |
| ------ | ---- | ---------------- | ------ |
| PATCH | /api/v0/archers/{archer_id} | Archer Profile Page | Editing profile info post-onboarding |
| GET | /api/v0/bows/{bow_id} | Profile & Equipment | Individual bow details view |
| PATCH | /api/v0/bows/{bow_id} | Profile & Equipment | Updating bow specifications |
| DELETE | /api/v0/bows/{bow_id} | Profile & Equipment | Bow deletion management |
| GET | /api/v0/arrows/{arrow_id} | Profile & Equipment | Individual arrow details view |
| PATCH | /api/v0/arrows/{arrow_id} | Profile & Equipment | Updating arrow specifications |
| DELETE | /api/v0/arrows/{arrow_id} | Profile & Equipment | Arrow deletion management |

## 7. Overlap & Dependency with Session Flow Refactoring

This feature interacts directly with the session flow refactoring documented in
[PRD: Removing Obsolete Code & Refactoring the Session Flow](../simplify_refactor/PRD.md):

| Touchpoint | Interaction / Dependency |
| ---------- | ------------------------ |
| **Bows in Sessions** | Sessions require a `bow_id` FK pointing to `bow(bow_id)` registered duringonboarding. At least one bow is guaranteed by the onboarding completeness check. |
| **Arrows in Sessions & Shots** | If the archer registered arrows during onboarding, the session enforces `shots_per_end <= number_of_registered_arrows` and mandatory shot tagging. If skipped, sessions default to `arrow_id = NULL` and `shots_per_end >= 3`. |
| **UI Arrow Tagging** | The frontend uses the archer's registered arrows (`GET /api/v0/arrows`) to allow selection by `arrow_number`, translating to `arrow_id` for shot persistence. |
| **Auth & Archer Migration** | Database migration coordinates splitting `auth`/`archer` and creating `bow`/`arrow` tables alongside dropping obsolete `slot`, `target`, and `club` tables. |

> [!NOTE]
> Including optional arrow registration during onboarding cleanly fulfills the arrow prerequisites
> for session refactoring without needing a full profile management UI.
