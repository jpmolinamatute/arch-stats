# Task 006: Backend Integration Tests for Simplified Architecture

## Git Branch

`feature/006-integration-tests-simplify`

## Objective

Refactor the backend integration test suite in `backend/tests/integration/` to comprehensively
verify the simplified single-archer architecture against a live PostgreSQL database. Update session
and shot endpoint tests, update repository integration tests, remove obsolete slot and WebSocket
test files (`endpoint_slot_test.go`, `repo_slot_test.go`, `websocket_test.go`), update legacy auth
integration tests, and verify that all dropped legacy routes (including `/auth/register`) return
404 Not Found.

## Dependencies

- Task 001 (Database migrations and canonical schema).
- Task 005 (HTTP handlers, route refactoring, and websocket teardown).

## Acceptance Criteria

- [ ] `backend/tests/integration/endpoint_session_test.go` updated:
    - [ ] Session creation with bow, location, indoor flag, distance, face_type, shots_per_end,
          interval_seconds, goal
    - [ ] `GET /api/v0/session/open` returns active session for owner
    - [ ] Rejection (409 Conflict) when attempting to create a second open session for the archer
    - [ ] Rejection (422 Unprocessable Entity) when bow belongs to a different archer
    - [ ] Rejection (422 Unprocessable Entity) when `shots_per_end` exceeds archer's `in_use`
          registered arrow count
    - [ ] `PATCH /api/v0/session/{id}/close` closes session with mandatory rating `status`
          (`bad`, `neutral`, `good`) and optional reflections (`was_goal_achieved`, `did_well`,
          `need_work`)
    - [ ] Rejection (400 Bad Request) when `PATCH /api/v0/session/{id}/close` omits `status` or
          sends `not_rated`
    - [ ] `DELETE /api/v0/session/{id}` marks session `is_deleted = TRUE` and omits it from list
    - [ ] Closed session cannot accept new shots
- [ ] `backend/tests/integration/endpoint_shot_test.go` updated:
    - [ ] Scored session shot creation with coordinates, score (0-10), and inner-10 `is_x`
    - [ ] Batch shot creation (3 to 10 shots) in a single request
    - [ ] Mandatory arrow enforcement: archer with registered arrows must provide valid `arrow_id`
    - [ ] Arrow exclusion: archer without registered arrows must provide `arrow_id = null`
    - [ ] Volume session (`face_type = none`): coordinates and score must be null
    - [ ] `GET /api/v0/session/{id}/live-stats` returns verified statistical aggregates
    - [ ] Shot deletion via `DELETE /api/v0/shot/{shot_id}`
- [ ] Route teardown verification:
    - [ ] `GET /api/v0/session/slot/*` returns 404
    - [ ] `GET /api/v0/stats/ws/*` returns 404
    - [ ] `POST /api/v0/slots/*` returns 404
    - [ ] `POST /api/v0/auth/register` returns 404
- [ ] Legacy auth integration tests updated:
    - [ ] `backend/tests/integration/endpoint_auth_test.go` and `auth_flow_test.go` updated to
          replace calls to `/auth/register` with calls to `POST /api/v0/archers` and
          `POST /api/v0/bows` or pruned
- [ ] Obsolete integration test files deleted:
    - [ ] `backend/tests/integration/endpoint_slot_test.go` deleted
    - [ ] `backend/tests/integration/repo_slot_test.go` deleted
    - [ ] `backend/tests/integration/websocket_test.go` deleted
- [ ] Test helpers updated:
    - [ ] `backend/tests/integration/helpers_test.go` cleaned of slot and target setup helpers;
          provides canonical session creation helper
- [ ] `cd backend && go test ./tests/integration/... -v -count=1` passes against test DB.

## Files to Create/Modify/Remove

| Action | Path |
| ------ | ---- |
| Modify | `backend/tests/integration/endpoint_session_test.go` |
| Modify | `backend/tests/integration/endpoint_shot_test.go` |
| Modify | `backend/tests/integration/endpoint_auth_test.go` |
| Modify | `backend/tests/integration/auth_flow_test.go` |
| Modify | `backend/tests/integration/repo_session_test.go` |
| Modify | `backend/tests/integration/repo_shot_test.go` |
| Modify | `backend/tests/integration/helpers_test.go` |
| Delete | `backend/tests/integration/endpoint_slot_test.go` |
| Delete | `backend/tests/integration/repo_slot_test.go` |
| Delete | `backend/tests/integration/websocket_test.go` |

## Reference

- [PRD.md](../PRD.md)
- [helpers_test.go](../../../backend/tests/integration/helpers_test.go)
- [endpoint_session_test.go](../../../backend/tests/integration/endpoint_session_test.go)
- [endpoint_auth_test.go](../../../backend/tests/integration/endpoint_auth_test.go)
- [auth_flow_test.go](../../../backend/tests/integration/auth_flow_test.go)

## Steps

- [ ] **Step 1: Update test helpers in `backend/tests/integration/helpers_test.go`**

  Remove slot creation helpers (`createTestSlot`). Add helper to register an archer, bow,
  optional arrows, and open session in a single call.

- [ ] **Step 2: Refactor `endpoint_session_test.go`**

  Write test cases for session creation with bow, single open session constraint, close reflection,
  and soft deletion.

- [ ] **Step 3: Refactor `endpoint_shot_test.go`**

  Write test cases for scored vs volume shots, arrow tagging rules, batch recording, and HTTP
  live stats calculation.

- [ ] **Step 4: Refactor repository integration tests**
  Update `repo_session_test.go` and `repo_shot_test.go` to test database queries against the
  canonical schema.

- [ ] **Step 5: Add route teardown tests**

  Verify that dropped routes (`/slot`, `/stats/ws`, `/auth/register`) return HTTP 404.

- [ ] **Step 6: Update legacy auth integration tests**

  In `backend/tests/integration/endpoint_auth_test.go` and `auth_flow_test.go`, replace calls to
  `/auth/register` with calls to `POST /api/v0/archers` and `POST /api/v0/bows`.

- [ ] **Step 7: Delete obsolete integration test files**

  ```bash
  cd backend/tests/integration
  rm -f endpoint_slot_test.go repo_slot_test.go websocket_test.go
  ```

- [ ] **Step 8: Execute integration test suite**

  ```bash
  cd backend && go test ./tests/integration/... -v -count=1
  ```

  Verify all tests pass with 0 failures.

- [ ] **Step 9: Commit changes**

  ```bash
  git add backend/tests/integration/
  git commit -m "test(integration): verify simplified session and shot flows; remove legacy tests"
  ```

## Verification

- `cd backend && go test ./tests/integration/... -v -count=1` passes with 0 failures.
- All obsolete slot, websocket, and `/auth/register` endpoints return 404.
