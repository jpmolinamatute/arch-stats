# Task 002: Domain Model Structs and Enum Cleanup

## Git Branch

`feature/002-domain-models-enums-simplify`

## Objective

Refactor the Go domain model structs and enums in `backend/internal/model/` to reflect the
simplified single-archer session and shot architecture. Absorb slot fields into `session.go`,
re-point `shot.go` from `slot_id` to `session_id`, update `live_stats.go` for HTTP delivery,
remove obsolete enums (`SlotLetter`, `WSContentType`), refactor `SessionStatus` to represent
session ratings (`not_rated`, `bad`, `neutral`, `good`), delete obsolete model files
(`slot.go`, `target.go`), and delete obsolete registration struct `AuthRegistrationRequest`.

## Dependencies

- Task 001 (Database migrations and canonical schema).

## Acceptance Criteria

- [ ] Enums cleaned and refactored in `backend/internal/model/enums.go`:
    - [ ] `SlotLetter` type and constants (`SlotLetterA`..`SlotLetterD`) removed
    - [ ] `WSContentType` type and constants removed
    - [ ] `SessionStatus` refactored to session rating enum: `not_rated`, `bad`, `neutral`, `good`
    - [ ] `ArrowStatus` (`in_use`, `damaged`, `lost`) supported
    - [ ] `Gender`, `Bowstyle`, `FaceType`, and `AuthStatus` preserved
- [ ] Refactored `backend/internal/model/session.go`:
    - [ ] `SessionCreate` struct:
        - `ArcherID uuid.UUID` (`json:"archer_id" validate:"required"`)
        - `BowID uuid.UUID` (`json:"bow_id" validate:"required"`)
        - `SessionLocation string` (`json:"session_location" validate:"required,min=1,max=255"`)
        - `IsIndoor bool` (`json:"is_indoor"`)
        - `Distance int16` (`json:"distance" validate:"required,gte=1,lte=100"`)
        - `FaceType FaceType` (`json:"face_type" validate:"required"`)
        - `ShotsPerEnd int16` (`json:"shots_per_end" validate:"required,gte=3"`)
        - `IntervalSeconds int16` (`json:"interval_seconds" validate:"required,gte=1,lte=100"`)
        - `Goal *string` (`json:"goal,omitempty" validate:"omitempty,max=2000"`)
        - `Status *SessionStatus` (`json:"status,omitempty"`; defaults to `not_rated`)
    - [ ] `SessionClose` struct:
        - `Status SessionStatus` (`json:"status" validate:"required,oneof=bad neutral good"`)
          (mandatory on close)
        - `WasGoalAchieved *bool` (`json:"was_goal_achieved,omitempty"`)
        - `DidWell *string` (`json:"did_well,omitempty" validate:"omitempty,max=2000"`)
        - `NeedWork *string` (`json:"need_work,omitempty" validate:"omitempty,max=2000"`)
    - [ ] `SessionSet` struct:
        - Updated with optional pointers for location, indoor, distance, face_type,
          shots_per_end, interval_seconds, goal, status, reflections, is_deleted, closed_at
    - [ ] `SessionRead` struct:
        - Contains `SessionID`, `ArcherID`, `BowID`, `SessionLocation`, `IsIndoor`, `Distance`,
          `FaceType`, `ShotsPerEnd`, `IntervalSeconds`, `Goal`, `Status`, `WasGoalAchieved`,
          `DidWell`, `NeedWork`, `IsDeleted`, `CreatedAt`, `ClosedAt`
    - [ ] `SessionFilter` struct:
        - Filter by `SessionID`, `ArcherID`, `BowID`, `Status`, `IsIndoor`, `IsDeleted`, `ClosedAt`
    - [ ] `SessionID` and `SessionId` alias preserved
- [ ] Refactored `backend/internal/model/shot.go`:
    - [ ] `ShotCreate` references `SessionID uuid.UUID` instead of `SlotID`
    - [ ] `ShotRead` references `SessionID uuid.UUID` instead of `SlotID`
    - [ ] `ShotFilter` references `SessionID *uuid.UUID` instead of `SlotID`
    - [ ] `ShotSet` and `ShotID` structs updated accordingly
- [ ] Refactored `backend/internal/model/live_stats.go`:
    - [ ] `Stats` struct references `SessionID uuid.UUID` instead of `SlotID`
    - [ ] `SessionLiveStatsRead` aggregates `SessionID`, `Stats`, and `Scores []ShotScore`
    - [ ] `WebSocketMessage` struct removed
- [ ] Obsolete model files and structs removed:
    - [ ] `backend/internal/model/slot.go` deleted
    - [ ] `backend/internal/model/target.go` deleted
    - [ ] `model.AuthRegistrationRequest` deleted from `backend/internal/model/auth.go`
- [ ] Unit tests in `backend/internal/model/` verify serialization and validation rules.
- [ ] `cd backend && go test ./internal/model/... -v` passes.

## Files to Create/Modify/Remove

| Action | Path |
| ------ | ---- |
| Modify | `backend/internal/model/enums.go` |
| Modify | `backend/internal/model/session.go` |
| Modify | `backend/internal/model/shot.go` |
| Modify | `backend/internal/model/live_stats.go` |
| Modify | `backend/internal/model/auth.go` |
| Create | `backend/internal/model/session_test.go` |
| Create | `backend/internal/model/shot_test.go` |
| Delete | `backend/internal/model/slot.go` |
| Delete | `backend/internal/model/target.go` |

## Reference

- [PRD.md](../PRD.md)
- [enums.go](../../../backend/internal/model/enums.go)
- [session.go](../../../backend/internal/model/session.go)
- [auth.go](../../../backend/internal/model/auth.go)

## Steps

- [ ] **Step 1: Write model validation unit tests**

  Create `backend/internal/model/session_test.go` and `shot_test.go`:
    - Test `SessionCreate` validation for distance (1-100), shots_per_end (>= 3), interval_seconds
    - Test `SessionClose` reflection field length constraints (<= 2000)
    - Test `ShotCreate` validation with `SessionID` and coordinate/score consistency

- [ ] **Step 2: Clean up and refactor enums in `backend/internal/model/enums.go`**

  Remove `SlotLetter` and `WSContentType`. Refactor `SessionStatus` to represent session ratings
  (`not_rated`, `bad`, `neutral`, `good`). Ensure remaining enums compile.

- [ ] **Step 3: Update `backend/internal/model/session.go`**

  Define `SessionCreate`, `SessionClose`, `SessionSet`, `SessionFilter`, and `SessionRead` with
  all canonical fields matching the database schema.

- [ ] **Step 4: Update `backend/internal/model/shot.go`**

  Replace `SlotID` with `SessionID` across all shot structs.

- [ ] **Step 5: Update `backend/internal/model/live_stats.go`**

  Replace `slot_id` with `session_id` in `Stats`. Define `SessionLiveStatsRead`.
  Remove `WebSocketMessage`.

- [ ] **Step 6: Remove `AuthRegistrationRequest` from `auth.go`**

  In `backend/internal/model/auth.go`, delete `AuthRegistrationRequest` struct.

- [ ] **Step 7: Delete obsolete model files**

  ```bash
  rm backend/internal/model/slot.go backend/internal/model/target.go
  ```

- [ ] **Step 8: Run tests to verify model package passes**

  ```bash
  cd backend && go test ./internal/model/... -v
  ```

- [ ] **Step 9: Commit changes**

  ```bash
  git add backend/internal/model/
  git commit -m "feat(model): update models and enums; remove slots and legacy auth models"
  ```

## Verification

- `cd backend && go test ./internal/model/... -v` passes with 0 failures.
- No references to `SlotLetter` or `WSContentType` in `backend/internal/model/`.
- `SessionStatus` defines ratings (`not_rated`, `bad`, `neutral`, `good`).
- No references to `AuthRegistrationRequest` in `backend/internal/model/`.
