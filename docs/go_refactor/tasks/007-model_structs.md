# Task 007: Define Domain Model Structs in `internal/model/`

## Git Branch

`refactor/007-model-structs`

## Objective

Define all domain model structs in `internal/model/` with JSON tags, matching the existing
Pydantic schemas from `backend-old/src/schema/`. This includes request/response types for all
domains: archer, auth, session, slot, shot, face, target, and shared enums.

## Dependencies

- Task 001 (Go module scaffold exists)

## Acceptance Criteria

- [ ] The following files exist in `backend/internal/model/`:
    - `enums.go` — shared enums (gender, bowstyle, session status, slot type, etc.)
    - `archer.go` — Archer domain types (create, read, update, filter)
    - `auth.go` — Auth types (session create, authenticated response, registration)
    - `session.go` — Shooting session types (create, read, update, filter)
    - `slot.go` — Slot types (create, read, update, filter)
    - `shot.go` — Shot types (create, read, update, filter)
    - `face.go` — Face types (create, read)
    - `target.go` — Target types (create, read)
    - `live_stats.go` — Live stats / WebSocket message types
    - `base.go` — Shared base types (pagination, list response wrapper)
- [ ] All structs have appropriate JSON tags matching the current API contract.
- [ ] All structs have validation tags where appropriate.
- [ ] Enums are implemented as typed string constants.
- [ ] Unit tests verify JSON marshaling/unmarshaling for representative types.
- [ ] `go test ./internal/model/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/model/enums.go` |
| Create | `backend/internal/model/base.go` |
| Create | `backend/internal/model/archer.go` |
| Create | `backend/internal/model/auth.go` |
| Create | `backend/internal/model/session.go` |
| Create | `backend/internal/model/slot.go` |
| Create | `backend/internal/model/shot.go` |
| Create | `backend/internal/model/face.go` |
| Create | `backend/internal/model/target.go` |
| Create | `backend/internal/model/live_stats.go` |
| Create | `backend/internal/model/model_test.go` |
| Delete | `backend/internal/model/.gitkeep` |

## Reference

The Python schemas to port are in:

- [enums.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/enums.py)
- [archer_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/archer_schema.py)
- [auth_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/auth_schema.py)
- [session_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/session_schema.py)
- [slot_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/slot_schema.py)
- [shot_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/shot_schema.py)
- [face_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/face_schema.py)
- [target_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/target_schema.py)
- [base.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/base.py)

## Steps

- [ ] **Step 1: Write failing tests for JSON round-tripping**

  Create `backend/internal/model/model_test.go` with tests that:
    - Marshal an `ArcherRead` struct to JSON and verify field names match the API contract
    - Unmarshal a JSON object into an `ArcherRead` struct
    - Verify enum values marshal as expected strings
    - Test `SessionRead` and `SlotRead` similarly

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/model/... -v
  ```

- [ ] **Step 3: Implement `enums.go`**

  Define typed string constants for all enums matching `backend-old/src/schema/enums.py`:
    - `Gender` (male, female, other, prefer_not_to_say)
    - `Bowstyle` (recurve, compound, barebow, longbow, traditional)
    - `SessionStatus` (open, closed)
    - Any other enums used in the schemas

- [ ] **Step 4: Implement `base.go`**

  Define shared types:
    - `ListResponse[T any]` — generic wrapper with `data []T` and pagination fields
    - `ErrorResponse` — standard error response body

- [ ] **Step 5: Implement domain model files**

  For each domain (archer, auth, session, slot, shot, face, target, live_stats):
    - Define `XxxCreate`, `XxxRead`, `XxxUpdate`, `XxxFilter` structs
    - Use `json:"field_name"` tags matching current snake_case API
    - Use `*T` (pointer) for optional fields
    - Use `time.Time` for datetime fields
    - Use `uuid.UUID` for ID fields (from `github.com/google/uuid`)

- [ ] **Step 6: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/model/... -v
  ```

- [ ] **Step 7: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [ ] **Step 8: Commit**

  ```bash
  rm -f backend/internal/model/.gitkeep
  git add -A
  git commit -m "feat: define all domain model structs with JSON tags and enums"
  ```

## Verification

- `cd backend && go test ./internal/model/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
- JSON tags match the existing API contract (snake_case field names).
