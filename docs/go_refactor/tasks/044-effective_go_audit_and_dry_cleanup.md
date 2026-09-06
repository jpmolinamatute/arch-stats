# Task 044: Effective Go Audit and DRY Cleanup

## Git Branch

`refactor/044-effective-go-audit-and-dry-cleanup`

## Objective

Audit the entire Go backend for compliance with
[Effective Go](https://go.dev/doc/effective_go) and the project's own
[backend-go-coding](file:///home/juanpa/Projects/arch-stats/.agent/skills/backend-go-coding/SKILL.md)
skill. Eliminate duplicated code by extracting shared patterns into reusable
helpers (DRY principle), and enforce consistent naming conventions across every
package. After this task, every backend file should be free of code duplication,
follow idiomatic Go naming, and pass linting with zero findings.

## Dependencies

- Task 043 (linting and formatting — `golangci-lint` and `gofumpt` are
  configured and passing)

## Scope

This task covers three audit pillars:

1. **DRY violations** — identical or near-identical code blocks repeated across
   packages.
2. **Effective Go compliance** — naming, control flow, interface design, error
   handling, and concurrency patterns.
3. **Naming consistency** — uniform conventions for types, functions, variables,
   error messages, and comments across all packages.

---

## Audit Findings and Remediation Plan

### A. Repository Layer — DRY Violations

#### A1. Duplicated `Delete` boilerplate

Every CRUD repository (`archer.go`, `session.go`, `slot.go`, `shot.go`,
`target.go`) contains an identical `Delete` method with the same structure:
build DELETE query → exec → check `RowsAffected` → return `ErrNotFound`. The
only differences are the table name and the primary key column name.

**Remediation**: Extract a generic `deleteByID` helper in `base.go`:

```go
// deleteByID builds and executes a DELETE WHERE pk = id statement.
// Returns apperror.ErrNotFound if no row was deleted.
func deleteByID(ctx context.Context, db DBTX, table, pkColumn string, id uuid.UUID) error {
    sql, args, err := StmtBuilder.Delete(table).
        Where(squirrel.Eq{pkColumn: id}).
        ToSql()
    if err != nil {
        return fmt.Errorf("building delete %s query: %w", table, err)
    }

    tag, err := db.Exec(ctx, sql, args...)
    if err != nil {
        return fmt.Errorf("executing delete %s: %w", table, err)
    }

    if tag.RowsAffected() == 0 {
        return apperror.ErrNotFound
    }

    return nil
}
```

Then refactor each repository's `Delete` method to call:

```go
func (r *ArcherRepo) Delete(ctx context.Context, id uuid.UUID) error {
    return deleteByID(ctx, r.db, "archer", "archer_id", id)
}
```

**Files affected**:
- [base.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/base.go)
  (add helper)
- [archer.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/archer.go)
- [session.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/session.go)
- [slot.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/slot.go)
- [shot.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/shot.go)
- [target.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/target.go)

#### A2. Duplicated `Update` tail (exec + RowsAffected check)

Every `Update` method shares the same trailing code block after building the
query: `ToSql` → `Exec` → check `RowsAffected() == 0` → return
`apperror.ErrNotFound`. This exact 12-line block is repeated in 5 repositories.

**Remediation**: Extract `execUpdate` helper in `base.go`:

```go
// execUpdate executes a squirrel update builder and returns apperror.ErrNotFound
// if no rows were affected.
func execUpdate(ctx context.Context, db DBTX, q squirrel.UpdateBuilder) error {
    sql, args, err := q.ToSql()
    if err != nil {
        return fmt.Errorf("building update query: %w", err)
    }

    tag, err := db.Exec(ctx, sql, args...)
    if err != nil {
        return fmt.Errorf("executing update: %w", err)
    }

    if tag.RowsAffected() == 0 {
        return apperror.ErrNotFound
    }

    return nil
}
```

**Files affected**: Same repository files as A1.

#### A3. Duplicated `FindByID` pattern

Every repository's `FindByID` follows: build `SELECT ... WHERE pk = id` →
`QueryRow` → `ScanOne`. Only the table name, PK column, columns list, and scan
function differ.

**Remediation**: Extract `findByID` generic helper in `base.go`:

```go
// findByID builds a SELECT WHERE pk = id query and scans a single result.
func findByID[T any](
    ctx context.Context,
    db DBTX,
    table string,
    pkColumn string,
    columns []string,
    id uuid.UUID,
    scanFn func(pgx.Row) (T, error),
) (*T, error) {
    sql, args, err := StmtBuilder.Select(columns...).
        From(table).
        Where(squirrel.Eq{pkColumn: id}).
        ToSql()
    if err != nil {
        return nil, fmt.Errorf("building find %s by id query: %w", table, err)
    }

    row := db.QueryRow(ctx, sql, args...)
    return ScanOne(row, scanFn)
}
```

**Files affected**: Same repository files as A1.

#### A4. Duplicated `Create` returning-ID pattern

All `Create` methods follow: build `INSERT ... RETURNING pk` → `QueryRow.Scan`
→ return UUID. The structure is identical across all 5 CRUD repos.

**Remediation**: Extract `createReturningID` helper in `base.go`:

```go
// createReturningID builds an INSERT ... RETURNING pk query and scans the result.
func createReturningID(
    ctx context.Context,
    db DBTX,
    builder squirrel.InsertBuilder,
    pkColumn string,
) (uuid.UUID, error) {
    sql, args, err := builder.Suffix("RETURNING " + pkColumn).ToSql()
    if err != nil {
        return uuid.Nil, fmt.Errorf("building create query: %w", err)
    }

    var newID uuid.UUID
    if err := db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
        return uuid.Nil, fmt.Errorf("inserting record: %w", err)
    }

    return newID, nil
}
```

**Files affected**: Same repository files as A1.

---

### B. Handler Layer — DRY Violations

#### B1. Duplicated UUID-from-URL parsing boilerplate

Handlers repeatedly parse URL parameters into UUIDs with the same 6+ line
block: `getURLParam` → try fallback param → `uuid.Parse` → `writeAppError`.
This pattern appears at least 12 times across handler files.

**Remediation**: Extract `parseUUIDParam` in `helpers.go`:

```go
// parseUUIDParam extracts a UUID from the URL path, trying paramNames in order.
// Writes a 422 validation error response and returns uuid.Nil, false on failure.
func parseUUIDParam(w http.ResponseWriter, r *http.Request, paramNames ...string) (uuid.UUID, bool) {
    var raw string
    for _, name := range paramNames {
        if raw = getURLParam(r, name); raw != "" {
            break
        }
    }

    id, err := uuid.Parse(raw)
    if err != nil {
        writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid "+paramNames[0]+" is required"))
        return uuid.Nil, false
    }

    return id, true
}
```

Then refactor handler call sites, for example:

```go
// Before (repeated everywhere):
idStr := getURLParam(r, "slot_id")
if idStr == "" { idStr = getURLParam(r, "slot") }
if idStr == "" { idStr = getURLParam(r, "id") }
slotID, err := uuid.Parse(idStr)
if err != nil {
    writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid slot_id is required"))
    return
}

// After:
slotID, ok := parseUUIDParam(w, r, "slot_id", "slot", "id")
if !ok { return }
```

**Files affected**:
- [helpers.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/helpers.go)
  (add helper)
- [archer.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/archer.go)
- [session.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/session.go)
- [slot.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/slot.go)
- [shot.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/shot.go)

#### B2. Duplicated auth-then-ownership guard

Multiple session and slot handlers repeat the same 3-step boilerplate:
`middleware.GetArcherID` → parse archer UUID from URL → compare and reject if
mismatched. This exact pattern appears in `GetOpenForArcher`,
`GetClosedForArcher`, `GetParticipating`, and `GetArcherCurrentSlot`.

**Remediation**: Extract `requireOwnership` in `helpers.go`:

```go
// requireOwnership verifies the authenticated archer matches the archer_id in the URL.
// Writes the error response and returns uuid.Nil, false on failure.
func requireOwnership(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
    authArcherID, err := middleware.GetArcherID(r.Context())
    if err != nil {
        writeAppError(w, err)
        return uuid.Nil, false
    }

    archerID, ok := parseUUIDParam(w, r, "archer_id", "id")
    if !ok {
        return uuid.Nil, false
    }

    if authArcherID != archerID {
        writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
        return uuid.Nil, false
    }

    return archerID, true
}
```

**Files affected**: Same handler files as B1.

#### B3. Duplicated `ErrorResponse` type definition

`ErrorResponse` is defined both in
[model/base.go](file:///home/juanpa/Projects/arch-stats/backend/internal/model/base.go)
(`model.ErrorResponse`) and in
[middleware/error_mapper.go](file:///home/juanpa/Projects/arch-stats/backend/internal/middleware/error_mapper.go)
(`middleware.ErrorResponse`). Additionally,
[handler/helpers.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/helpers.go)
creates a type alias `ErrorResponse = model.ErrorResponse` but then uses
`middleware.ErrorResponse` in `writeError`.

**Remediation**: Keep `ErrorResponse` only in `model/base.go` (the canonical
domain layer). Update `middleware/error_mapper.go` to import and use
`model.ErrorResponse`. Remove the redundant type alias from `handler/helpers.go`.

**Files affected**:
- [model/base.go](file:///home/juanpa/Projects/arch-stats/backend/internal/model/base.go)
  (keep, canonical)
- [middleware/error_mapper.go](file:///home/juanpa/Projects/arch-stats/backend/internal/middleware/error_mapper.go)
  (import `model.ErrorResponse`, remove local struct)
- [handler/helpers.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/helpers.go)
  (remove alias, use `model.ErrorResponse` directly in `writeError`)

#### B4. Redundant exported/unexported function pairs in helpers.go

`helpers.go` defines both private (`writeJSON`, `readJSON`, `writeError`,
`writeAppError`) and public (`WriteJSON`, `ReadJSON`, `WriteError`,
`WriteAppError`) versions that are simple wrappers. Since all handler files
are in the same `handler` package, the exported versions are only needed if
consumed outside the package. If no external consumer exists, remove the
exported wrappers to simplify the API surface. If external consumers exist,
remove the unexported variants and use only the exported ones everywhere.

**Remediation**: Audit usages. If only internal: keep only unexported. If
external consumers exist (tests, middleware): keep only exported and update all
call sites.

**Files affected**:
- [handler/helpers.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/helpers.go)

#### B5. Duplicated nil-to-empty-slice guard

Multiple handlers repeat the same nil-to-empty-slice pattern:

```go
if sessions == nil {
    sessions = []model.SessionRead{}
}
```

This appears in `ListAllOpen`, `GetClosedForArcher`,
`List` (archer), `GetBySlot` (shot).

**Remediation**: The service layer already does this for `ArcherService.List`
(returns `[]model.ArcherRead{}` when nil). Apply the same pattern in all service
`List`/`FindAll` methods so handlers never receive nil slices, then remove
the nil-guard from handler code.

**Files affected**:
- [service/session.go](file:///home/juanpa/Projects/arch-stats/backend/internal/service/session.go)
- [service/shot.go](file:///home/juanpa/Projects/arch-stats/backend/internal/service/shot.go)
- All handler files that contain the nil-guard.

---

### C. Naming Consistency

#### C1. Error message casing inconsistency

Error messages use inconsistent casing:
- `"ERROR: session_id wasn't provided"` (ALL CAPS prefix, session handler Close)
- `"ERROR: user not allowed to open a session..."` (ALL CAPS prefix, session
  handler Create)
- `"Forbidden"` (capitalized, session/slot handlers)
- `"archer id is required"` (lowercase, archer service)
- `"first_name is required"` (lowercase, archer service)

Per Effective Go: error strings should not be capitalized (unless beginning with
proper nouns or acronyms) and should not end with punctuation.

**Remediation**: Standardize all error messages to lowercase, without `ERROR:`
prefix. Examples:
- `"session_id wasn't provided"` → `"session_id is required"`
- `"ERROR: user not allowed to open a session for another archer"` →
  `"forbidden: cannot open session for another archer"`
- `"Forbidden"` → `"forbidden"`

**Files affected**:
- [handler/session.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/session.go)
- [handler/slot.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/slot.go)
- [handler/shot.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/shot.go)

#### C2. Inconsistent guard-clause error message format in Update safety check

The "at least one filter" guard error message differs:
- `"update requires at least one filter condition to prevent unrestricted
  updates"` (archer, session, slot, shot)
- `"at least one filter criterion is required for update"` (target)

**Remediation**: Unify to a single phrasing. Add a package-level sentinel or
constant:

```go
var errUpdateRequiresFilter = errors.New("update requires at least one filter condition")
```

**Files affected**: All repository files with `Update` methods.

#### C3. Inconsistent `fmt.Errorf` context prefixes

Error wrapping messages are inconsistent in style:
- `"building find by id query"` (archer) vs `"building find slot by id query"`
  (slot) vs `"building find shot by id query"` (shot)
- `"executing delete"` (archer) vs `"executing delete session"` (session) vs
  `"executing delete slot"` (slot)

**Remediation**: Establish a consistent format:
`"<operation> <entity>: %w"`. Example: `"finding archer by id: %w"`,
`"deleting session: %w"`, `"updating slot: %w"`.

**Files affected**: All repository files.

#### C4. Mixed response styles for write operations

Handlers return different shapes for successful mutations:
- `model.SessionID{SessionID: &id}` for Create (session)
- `model.ArcherID{ArcherID: id}` for Create (archer) — note: not a pointer
- `model.ShotID{ShotID: id}` for Create (shot) — not a pointer
- `map[string]string{"status": "closed"}` for Close (session) — ad-hoc map
- `w.WriteHeader(http.StatusNoContent)` for Delete (archer) — no body
- `w.WriteHeader(http.StatusOK)` for LeaveSession (slot) — no body

**Remediation**: Establish conventions:
- `201 Created` → return `{<entity>_id: uuid}` consistently (pointer vs value
  should match across all entities).
- `200 OK` for state mutations → return `{"status": "<action>"}` or the updated
  entity.
- `204 No Content` for destructive operations (delete, leave) → no body.

Document the convention in a code comment in `helpers.go`.

**Files affected**: All handler files.

#### C5. `viewBox` column uses camelCase instead of snake_case

In [face.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/face.go),
the `faceColumns` array includes `"viewBox"` which breaks the snake_case
convention used by all other column names. This appears to originate from the
database schema.

**Remediation**: If the database column is `viewBox`, add a comment explaining
the deviation. If it can be renamed via migration, add a migration to rename it
to `view_box` and update the repository and model accordingly.

**Files affected**:
- [repository/face.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/face.go)
- Potentially a new migration file.

---

### D. Effective Go Compliance

#### D1. Unnecessary `_ = ctx` pattern in FaceRepo

[face.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/face.go)
uses `_ = ctx` to suppress "unused parameter" warnings. Per Effective Go, if
the parameter is part of an interface contract but not used in the current
implementation, the idiomatic approach is to keep the parameter name (for
documentation) and rely on the linter configuration to suppress the warning, or
use a blank identifier in the signature `_ context.Context` only if the method
is not expected to ever use it.

**Remediation**: Since `FaceRepo` is backed by an in-memory catalog and may
transition to database-backed queries later, keep the parameter named `ctx` and
remove the `_ = ctx` lines. Configure `golangci-lint` to ignore unused
parameters in this specific case, or accept the linter directive.

**Files affected**:
- [repository/face.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/face.go)

#### D2. FaceRepo constructor uses variadic for optional dependency

`NewFaceRepo(db ...DBTX)` uses a variadic to make the `DBTX` parameter optional.
This is an anti-pattern per Effective Go — variadic arguments are for homogeneous
lists, not optional parameters. This also breaks constructor parity with other
repos.

**Remediation**: Change to `NewFaceRepo(db DBTX)` matching all other repository
constructors. If `db` can be nil for the in-memory catalog case, accept nil
explicitly rather than hiding optionality behind variadic syntax.

**Files affected**:
- [repository/face.go](file:///home/juanpa/Projects/arch-stats/backend/internal/repository/face.go)
- Any call site creating `FaceRepo`.

#### D3. Missing `Routes` method consistency

`ArcherHandler.Routes` is defined at the bottom of the file (line 184), while
all other handlers define `Routes` near the top (after the constructor). This
inconsistency makes the codebase harder to navigate.

**Remediation**: Move `ArcherHandler.Routes` to immediately after
`NewArcherHandler`, matching the pattern in session, slot, shot handlers.

**Files affected**:
- [handler/archer.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/archer.go)

#### D4. `getURLParam` defined in `archer.go` instead of `helpers.go`

The shared `getURLParam` function is defined at the bottom of `archer.go`
(line 193), even though it's used by every handler file. It belongs in
`helpers.go` alongside other shared handler utilities.

**Remediation**: Move `getURLParam` to `helpers.go`.

**Files affected**:
- [handler/archer.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/archer.go)
  (remove)
- [handler/helpers.go](file:///home/juanpa/Projects/arch-stats/backend/internal/handler/helpers.go)
  (add)

---

## Acceptance Criteria

- [ ] A `deleteByID` helper exists in `repository/base.go` and all repository
  `Delete` methods delegate to it.
- [ ] An `execUpdate` helper exists in `repository/base.go` and all repository
  `Update` methods use it for the tail exec + affected-rows check.
- [ ] A `findByID` generic helper exists in `repository/base.go` and all
  repository `FindByID` methods delegate to it.
- [ ] A `createReturningID` helper exists in `repository/base.go` and all
  repository `Create` methods delegate to it.
- [ ] A `parseUUIDParam` helper exists in `handler/helpers.go` and all UUID
  parsing from URL params uses it.
- [ ] A `requireOwnership` helper exists in `handler/helpers.go` and all
  auth + ownership checks use it.
- [ ] `ErrorResponse` is defined only in `model/base.go`. No duplicate
  definition exists in `middleware/error_mapper.go` or `handler/helpers.go`.
- [ ] The redundant exported/unexported function pairs in `handler/helpers.go`
  are consolidated into a single set of functions.
- [ ] Nil-to-empty-slice normalization happens in the service layer, not in
  handlers.
- [ ] All `fmt.Errorf` context messages follow the `"<operation> <entity>: %w"`
  format consistently.
- [ ] All error strings are lowercase without `ERROR:` prefix.
- [ ] The update-filter guard uses a shared sentinel error.
- [ ] `FaceRepo` constructor uses `NewFaceRepo(db DBTX)`, not variadic.
- [ ] `_ = ctx` lines are removed from `FaceRepo` methods.
- [ ] `getURLParam` is in `handler/helpers.go`, not `handler/archer.go`.
- [ ] `ArcherHandler.Routes` is positioned after the constructor.
- [ ] Handler response shapes for mutations are consistent and documented.
- [ ] `golangci-lint run ./...` passes with zero findings.
- [ ] `gofumpt -l .` reports no unformatted files.
- [ ] `go test ./... -v -count=1` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `backend/internal/repository/base.go` |
| Modify | `backend/internal/repository/archer.go` |
| Modify | `backend/internal/repository/session.go` |
| Modify | `backend/internal/repository/slot.go` |
| Modify | `backend/internal/repository/shot.go` |
| Modify | `backend/internal/repository/target.go` |
| Modify | `backend/internal/repository/face.go` |
| Modify | `backend/internal/handler/helpers.go` |
| Modify | `backend/internal/handler/archer.go` |
| Modify | `backend/internal/handler/session.go` |
| Modify | `backend/internal/handler/slot.go` |
| Modify | `backend/internal/handler/shot.go` |
| Modify | `backend/internal/middleware/error_mapper.go` |
| Modify | `backend/internal/service/session.go` |
| Modify | `backend/internal/service/shot.go` |
| Modify | `backend/internal/repository/archer_test.go` |
| Modify | `backend/internal/repository/session_test.go` |
| Modify | `backend/internal/repository/slot_test.go` |
| Modify | `backend/internal/repository/shot_test.go` |
| Modify | `backend/internal/repository/target_test.go` |
| Modify | `backend/internal/repository/base_test.go` |
| Modify | `backend/internal/handler/archer_test.go` |
| Modify | `backend/internal/handler/session_test.go` |
| Modify | `backend/internal/handler/slot_test.go` |
| Modify | `backend/internal/handler/shot_test.go` |
| Modify | `backend/internal/handler/helpers_test.go` |
| Modify | `backend/internal/middleware/error_mapper_test.go` |

## Reference

- Effective Go: https://go.dev/doc/effective_go
- Project Go coding standards:
  [backend-go-coding SKILL.md](file:///home/juanpa/Projects/arch-stats/.agent/skills/backend-go-coding/SKILL.md)
- Project instructions:
  [instructions.md](file:///home/juanpa/Projects/arch-stats/.agent/rules/instructions.md)
- Linting task:
  [task 043](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/043-implement_linting_and_formatter.md)

## Steps

- [ ] **Step 1: Add repository base helpers (`deleteByID`, `execUpdate`,
  `findByID`, `createReturningID`)**

  Add the four generic helpers to `backend/internal/repository/base.go` as
  described in findings A1–A4. Write unit tests for each helper in
  `backend/internal/repository/base_test.go`.

  ```bash
  cd backend && go test ./internal/repository/... -v -run TestDeleteByID
  cd backend && go test ./internal/repository/... -v -run TestExecUpdate
  cd backend && go test ./internal/repository/... -v -run TestFindByID
  cd backend && go test ./internal/repository/... -v -run TestCreateReturningID
  ```

- [ ] **Step 2: Refactor all repository CRUD methods to use base helpers**

  Refactor `Delete`, `Update` tail, `FindByID`, and `Create` methods in all
  5 CRUD repositories (archer, session, slot, shot, target) to delegate to the
  new helpers. Update existing tests to match the new error message format.

  ```bash
  cd backend && go test ./internal/repository/... -v
  ```

- [ ] **Step 3: Unify the update-filter guard error**

  Add `errUpdateRequiresFilter` sentinel to `base.go`. Replace all inline
  `errors.New("update requires...")` and
  `errors.New("at least one filter...")` calls in repository `Update` methods
  with this sentinel.

  ```bash
  cd backend && go test ./internal/repository/... -v
  ```

- [ ] **Step 4: Standardize `fmt.Errorf` context messages**

  Audit all repository files and normalize error wrapping messages to the
  `"<operation> <entity>: %w"` format. Update any tests that assert on error
  message strings.

  ```bash
  cd backend && go test ./internal/repository/... -v
  ```

- [ ] **Step 5: Fix FaceRepo anti-patterns (D1, D2)**

  Change `NewFaceRepo(db ...DBTX)` to `NewFaceRepo(db DBTX)`. Remove `_ = ctx`
  lines from `FindAll`, `FindByType`, and `FindByID`. Update any call sites
  and tests.

  ```bash
  cd backend && go test ./internal/repository/... -v
  ```

- [ ] **Step 6: Add handler helpers (`parseUUIDParam`, `requireOwnership`)**

  Add both helpers to `handler/helpers.go` as described in B1 and B2. Move
  `getURLParam` from `archer.go` to `helpers.go` (D4). Write unit tests in
  `handler/helpers_test.go`.

  ```bash
  cd backend && go test ./internal/handler/... -v -run TestParseUUIDParam
  cd backend && go test ./internal/handler/... -v -run TestRequireOwnership
  ```

- [ ] **Step 7: Refactor handler UUID parsing and ownership checks**

  Replace all inline UUID-from-URL parsing and auth-then-ownership boilerplate
  in handlers with calls to `parseUUIDParam` and `requireOwnership`. Update
  handler tests accordingly.

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 8: Consolidate `ErrorResponse` type (B3)**

  Remove `middleware.ErrorResponse` struct from `error_mapper.go`. Import and
  use `model.ErrorResponse` instead. Remove the type alias from
  `handler/helpers.go`. Update `writeError` to use `model.ErrorResponse`.
  Update tests.

  ```bash
  cd backend && go test ./internal/middleware/... -v
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 9: Consolidate exported/unexported helper pairs (B4)**

  Decide on exported vs unexported. Remove the redundant set. Update all call
  sites.

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 10: Move nil-to-empty-slice normalization to service layer (B5)**

  Ensure all service `List` / `FindAll` / `GetBy*` methods that return slices
  normalize nil to empty slice. Remove the handler-side nil guards.

  ```bash
  cd backend && go test ./internal/service/... -v
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 11: Standardize error message casing (C1)**

  Audit all error messages in handlers and services. Lowercase all messages.
  Remove `ERROR:` prefixes. Update test assertions.

  ```bash
  cd backend && go test ./... -v
  ```

- [ ] **Step 12: Standardize handler response shapes (C4)**

  Document the response convention in `handler/helpers.go`. Ensure Create
  returns 201 with consistent ID shapes, state mutations return 200, and
  destructive operations return 204. Update tests.

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 13: Move `ArcherHandler.Routes` position (D3)**

  Move the `Routes` method to immediately after `NewArcherHandler` in
  `handler/archer.go`.

- [ ] **Step 14: Run full verification suite**

  ```bash
  cd backend && gofumpt -l .
  cd backend && golangci-lint run ./...
  cd backend && go vet ./...
  cd backend && go test ./... -v -count=1
  cd backend && go build ./...
  ```

- [ ] **Step 15: Commit**

  ```bash
  git add -A
  git commit -m "refactor: apply Effective Go best practices, eliminate DRY violations, and unify naming"
  ```

## Verification

- `cd backend && golangci-lint run ./...` — passes with zero findings.
- `cd backend && gofumpt -l .` — reports no unformatted files.
- `cd backend && go vet ./...` — clean.
- `cd backend && go test ./... -v -count=1` — all tests pass.
- `cd backend && go build ./...` — compiles.
- `grep -rn "func.*Delete.*uuid.UUID.*error" backend/internal/repository/*.go`
  — every `Delete` method is a one-liner delegating to `deleteByID`.
- `grep -rn "getURLParam" backend/internal/handler/archer.go` — no definition
  (it is in `helpers.go`).
- `grep -rn "ErrorResponse" backend/internal/middleware/error_mapper.go` — uses
  `model.ErrorResponse`, no local struct.
- `grep -rn "ERROR:" backend/internal/handler/` — no results.
