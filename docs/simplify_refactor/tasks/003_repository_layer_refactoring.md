# Task 003: Repository Layer Refactoring and Obsolete Repositories Removal

## Git Branch

`feature/003-repository-layer-simplify`

## Objective

Refactor the database repository layer in `backend/internal/repository/` to support the canonical
single-archer session and shot schema. Update `session.go` with full configuration fields,
open session querying, closing with reflections, and soft deletion. Update `shot.go` to query
by `session_id` and fetch computed live stats. Delete obsolete repositories (`slot.go`, `target.go`,
`maintenance.go`, `reporting.go`).

## Dependencies

- Task 001 (Database migrations and canonical schema).
- Task 002 (Domain model structs and enum cleanup).

## Acceptance Criteria

- [ ] `backend/internal/repository/session.go` implements:
    - [ ] `Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)`  inserting all
          canonical session fields
    - [ ] `FindByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)` scanning all
          canonical columns and filtering `is_deleted = FALSE`
    - [ ] `FindOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)` finding
          session with `closed_at IS NULL AND is_deleted = FALSE`
    - [ ] `FindAll(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)`
          with dynamic Squirrel filtering
    - [ ] `Close(ctx context.Context, id uuid.UUID, data model.SessionClose) error` setting
          `closed_at = now()`, `status = data.Status`, `was_goal_achieved`, `did_well`, and
          `need_work`
    - [ ] `SoftDelete(ctx context.Context, id uuid.UUID) error` setting `is_deleted = TRUE`
    - [ ] Legacy `FindParticipating` method removed
- [ ] `backend/internal/repository/shot.go` implements:
    - [ ] `Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error)`
          inserting `session_id` directly
    - [ ] `CreateBatch(ctx context.Context, data []model.ShotCreate) ([]uuid.UUID, error)`
          batch inserting multiple shots in a single transaction
    - [ ] `FindByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error)`
    - [ ] `FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.ShotRead, error)`
          ordered by `created_at ASC`
    - [ ] `CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int, error)`
    - [ ] `Delete(ctx context.Context, id uuid.UUID) error`
    - [ ] `FindLiveStatsBySessionID(ctx context.Context, sid uuid.UUID)`
          `(*model.SessionLiveStatsRead, error)` querying `live_stat_by_session_id` view
    - [ ] Obsolete `FindBySlotID` and `CountBySlotID` methods removed
- [ ] Obsolete repository files deleted:
    - [ ] `backend/internal/repository/slot.go` and `slot_test.go`
    - [ ] `backend/internal/repository/target.go` and `target_test.go`
    - [ ] `backend/internal/repository/maintenance.go` and `maintenance_test.go`
    - [ ] `backend/internal/repository/reporting.go` and `reporting_test.go`
- [ ] Unit tests for `SessionRepo` and `ShotRepo` using mock `DBTX` verify Squirrel SQL generation.
- [ ] `cd backend && go test ./internal/repository/... -v` passes.

## Files to Create/Modify/Remove

| Action | Path |
| ------ | ---- |
| Modify | `backend/internal/repository/session.go` |
| Modify | `backend/internal/repository/session_test.go` |
| Modify | `backend/internal/repository/shot.go` |
| Modify | `backend/internal/repository/shot_test.go` |
| Delete | `backend/internal/repository/slot.go` |
| Delete | `backend/internal/repository/slot_test.go` |
| Delete | `backend/internal/repository/target.go` |
| Delete | `backend/internal/repository/target_test.go` |
| Delete | `backend/internal/repository/maintenance.go` |
| Delete | `backend/internal/repository/maintenance_test.go` |
| Delete | `backend/internal/repository/reporting.go` |
| Delete | `backend/internal/repository/reporting_test.go` |

## Reference

- [PRD.md](../PRD.md)
- [session.go](../../../backend/internal/repository/session.go)
- [shot.go](../../../backend/internal/repository/shot.go)

## Steps

- [ ] **Step 1: Write failing query building tests for refactored repositories**

  Update `backend/internal/repository/session_test.go` and `shot_test.go`:
    - Test `Create` session SQL includes `bow_id`, `distance`, `face_type`, `shots_per_end`
    - Test `FindOpen` SQL filters by `archer_id`, `closed_at IS NULL`, and `is_deleted = false`
    - Test `Close` session SQL updates `closed_at`, `status`, `was_goal_achieved`, `did_well`,
      `need_work`
    - Test `SoftDelete` SQL updates `is_deleted = true`
    - Test `ShotRepo.Create` and `CreateBatch` SQL target `session_id`
    - Test `FindLiveStatsBySessionID` queries `live_stat_by_session_id`

- [ ] **Step 2: Run tests to verify failure**

  ```bash
  cd backend && go test ./internal/repository/... -v
  ```

- [ ] **Step 3: Implement refactored `session.go` repository**

  Update `SessionRepo` with all methods using Squirrel and `DBTX`. Handle optional reflections
  and soft-delete filtering.

- [ ] **Step 4: Implement refactored `shot.go` repository**

  Update `ShotRepo` with `session_id` bindings, batch insertion helper, and query helper for
  `live_stat_by_session_id`.

- [ ] **Step 5: Delete obsolete repository files**

  ```bash
  cd backend/internal/repository
  rm -f slot.go slot_test.go target.go target_test.go
  rm -f maintenance.go maintenance_test.go reporting.go reporting_test.go
  ```

- [ ] **Step 6: Run tests to verify repository package passes**

  ```bash
  cd backend && go test ./internal/repository/... -v
  ```

- [ ] **Step 7: Commit changes**

  ```bash
  git add backend/internal/repository/
  git commit -m "feat(repo): refactor session and shot repos; delete slot, target, and maintenance"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` passes with 0 failures.
- No references to `slotRepo`, `targetRepo`, or `maintenanceRepo` in `internal/repository`.
