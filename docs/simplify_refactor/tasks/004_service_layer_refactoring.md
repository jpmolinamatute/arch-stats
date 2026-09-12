# Task 004: Service Layer Refactoring and Business Rule Enforcement

## Git Branch

`feature/004-service-layer-simplify`

## Objective

Refactor business logic in `backend/internal/service/` to enforce single-archer session and shot
domain rules. Update `SessionService` to validate bow ownership, enforce a single open session per
archer, and check `shots_per_end` against registered arrow counts. Update `ShotService` to validate
open session state, handle scored vs volume sessions (`face_type = none`), enforce mandatory arrow
tagging for archers with registered arrows, and deliver HTTP live stats.
Delete obsolete service files (`slot.go`, `target.go`).

## Dependencies

- Task 002 (Domain model structs and enum cleanup).
- Task 003 (Repository layer refactoring).

## Acceptance Criteria

- [ ] `backend/internal/service/session.go` implements:
    - [ ] Injected dependencies: `SessionRepo`, `BowRepo`, and `ArrowRepo`
    - [ ] `Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)`:
        - Validates that `BowID` belongs to `data.ArcherID` via `BowRepo.FindByID`
        - Enforces that the archer does not already have an open session (`apperror.ErrConflict`)
        - Enforces `Distance` (1-100) and `IntervalSeconds` (1-100)
        - Checks `ShotsPerEnd`: queries `ArrowRepo.CountInUseByArcherID(ctx, data.ArcherID)`.
          If in-use count > 0, enforces `shots_per_end <= count` (returns `apperror.ErrValidation`);
          if count == 0, enforces `shots_per_end >= 3`
    - [ ] `Close(ctx context.Context, id, archerID uuid.UUID, data model.SessionClose) error`:
        - Verifies session exists, belongs to `archerID`, and is currently open
        - Enforces `data.Status` is provided and is one of `bad`, `neutral`, `good` (returns
          `apperror.ErrValidation` if omitted or `not_rated`)
        - Enforces `was_goal_achieved` is only permitted if `session.Goal != nil`
        - Enforces reflection strings length <= 2000 characters
        - Sets session read-only with `closed_at` and updates `status`
    - [ ] `Delete(ctx context.Context, id, archerID uuid.UUID) error`:
        - Verifies session exists and belongs to `archerID`, then performs soft delete
    - [ ] Obsolete methods `ReOpen` and `GetParticipating` removed
- [ ] `backend/internal/service/shot.go` implements:
    - [ ] Injected dependencies: `ShotRepo`, `SessionRepo`, and `ArrowRepo` (removes `slotRepo`)
    - [ ] `Create(ctx context.Context, data model.ShotCreate, archerID uuid.UUID)`
          `(uuid.UUID, error)` and `CreateBatch`:
        - Verifies session exists, belongs to `archerID`, and is open (`closed_at IS NULL`)
        - For scored sessions (`session.FaceType != FaceTypeNone`): coordinates and score
          must be provided; score between 0 and 10; `is_x` requires score 10
        - For volume sessions (`session.FaceType == FaceTypeNone`): coordinates and score
          must all be nil
        - Arrow tagging rule: queries `ArrowRepo.CountInUseByArcherID(ctx, archerID)`:
            - If archer has in-use registered arrows (count > 0): `data.ArrowID` is mandatory,
              must belong to `archerID`, and the arrow must have `status == ArrowStatusInUse`
            - If archer has no in-use registered arrows (count == 0): `data.ArrowID` must be nil
    - [ ] `GetLiveStats(ctx context.Context, sessionID, archerID uuid.UUID)`
          `(*model.SessionLiveStatsRead, error)`: verifies session ownership and returns stats
    - [ ] `Delete(ctx context.Context, id, archerID uuid.UUID) error`: verifies session
          ownership and deletes shot record
- [ ] Obsolete service files deleted:
    - [ ] `backend/internal/service/slot.go` and `slot_test.go`
    - [ ] `backend/internal/service/target.go` and `target_test.go`
- [ ] Comprehensive unit tests in `session_test.go` and `shot_test.go` verify all business logic,
      validation branches, and error cases using repository mocks.
- [ ] `cd backend && go test ./internal/service/... -v` passes.

## Files to Create/Modify/Remove

| Action | Path |
| ------ | ---- |
| Modify | `backend/internal/service/session.go` |
| Modify | `backend/internal/service/session_test.go` |
| Modify | `backend/internal/service/shot.go` |
| Modify | `backend/internal/service/shot_test.go` |
| Delete | `backend/internal/service/slot.go` |
| Delete | `backend/internal/service/slot_test.go` |
| Delete | `backend/internal/service/target.go` |
| Delete | `backend/internal/service/target_test.go` |

## Reference

- [PRD.md](../PRD.md)
- [session.go](../../../backend/internal/service/session.go)
- [shot.go](../../../backend/internal/service/shot.go)

## Steps

- [ ] **Step 1: Write failing unit tests for `SessionService`**

  Update `backend/internal/service/session_test.go`:
    - Test `Create` rejects bow belonging to another archer
    - Test `Create` rejects when archer already has an open session
    - Test `Create` rejects `shots_per_end > in_use_arrows_count` when in-use arrows exist
    - Test `Create` allows `shots_per_end >= 3` when no arrows exist
    - Test `Close` validates mandatory `status` (accepts bad, neutral, good; rejects not_rated or
      missing)
    - Test `Close` rejects reflection goal achievement when no goal was set
    - Test `Close` rejects closing an already closed session

- [ ] **Step 2: Write failing unit tests for `ShotService`**

  Update `backend/internal/service/shot_test.go`:
    - Test `Create` rejects recording into closed session
    - Test `Create` scored shot requires `ArrowID` when archer has registered arrows
    - Test `Create` scored shot rejects `ArrowID` when archer has no registered arrows
    - Test `Create` volume shot rejects non-nil coordinates or score
    - Test `Create` scored shot rejects `is_x = true` with score < 10

- [ ] **Step 3: Run service tests to verify failure**

  ```bash
  cd backend && go test ./internal/service/... -v
  ```

- [ ] **Step 4: Implement refactored `session.go` service**

  Update `SessionService` struct, constructor, and methods to enforce all business constraints.

- [ ] **Step 5: Implement refactored `shot.go` service**

  Update `ShotService` struct, constructor, and methods to enforce single-archer session rules
  and live stats retrieval.

- [ ] **Step 6: Delete obsolete service files**

  ```bash
  cd backend/internal/service
  rm -f slot.go slot_test.go target.go target_test.go
  ```

- [ ] **Step 7: Run service tests to verify all pass**

  ```bash
  cd backend && go test ./internal/service/... -v
  ```

- [ ] **Step 8: Commit changes**

  ```bash
  git add backend/internal/service/
  git commit -m "feat(service): refactor session and shot services; remove slots and targets"
  ```

## Verification

- `cd backend && go test ./internal/service/... -v` passes with 0 failures.
- No references to `SlotService` or `TargetService` in `backend/internal/service`.
