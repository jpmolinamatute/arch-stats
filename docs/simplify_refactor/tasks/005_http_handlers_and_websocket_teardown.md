# Task 005: HTTP Handlers, Route Refactoring, and WebSocket Teardown

## Git Branch

`feature/005-handlers-routes-websocket-teardown`

## Objective

Refactor the HTTP transport layer in `backend/internal/handler/` and `backend/cmd/arch-stats/`.
Update session endpoints to handle full configuration on create, reflection fields on close, and
soft deletion. Update shot endpoints to bind directly to `session_id`. Introduce a standard HTTP
GET endpoint for session live stats. Delete obsolete slot and WebSocket handlers, remove the
`internal/websocket` package entirely, remove legacy registration endpoint (`POST /auth/register`)
and supporting auth service methods, and clean up Chi router wiring and adapters.

## Dependencies

- Task 004 (Service layer refactoring and business rules).

## Acceptance Criteria

- [ ] `backend/internal/handler/session.go` implements:
    - [ ] `POST /api/v0/session`: creates session with bow, distance, face_type, shots_per_end,
          interval_seconds, goal. Returns 201 + `SessionId`.
    - [ ] `GET /api/v0/session/open`: returns authenticated archer's open session or 200 with null
          session_id.
    - [ ] `GET /api/v0/session/{id}`: returns `SessionRead` if owned by authenticated archer.
    - [ ] `GET /api/v0/session`: returns array of `SessionRead` for archer (excludes soft-deleted).
    - [ ] `PATCH /api/v0/session/{id}/close`: accepts `SessionClose` payload containing mandatory
          `status` (`bad`, `neutral`, `good`) and optional reflections; returns 200 OK (returns 400
          if `status` is omitted or invalid).
    - [ ] `DELETE /api/v0/session/{id}`: soft-deletes session; returns 204 No Content.
    - [ ] `GET /api/v0/session/{id}/live-stats`: returns `SessionLiveStatsRead` via HTTP GET.
    - [ ] Obsolete routes removed: `/re-open`, `/participating`, `/open`,
          `/archer/{archer_id}/open-session`, `/archer/{archer_id}/close-session`.
- [ ] `backend/internal/handler/shot.go` implements:
    - [ ] `POST /api/v0/shot`: accepts single `ShotCreate` or batch array targeting `session_id`.
          Returns 201.
    - [ ] `GET /api/v0/shot/by-session/{session_id}`: lists shots for session; returns 200.
    - [ ] `GET /api/v0/shot/count-by-session/{session_id}`: returns integer count; returns 200.
    - [ ] `DELETE /api/v0/shot/{shot_id}`: deletes shot; returns 204 No Content.
    - [ ] Obsolete routes `/by-slot/{slot_id}` and `/count-by-slot/{slot_id}` removed.
- [ ] Obsolete handler and websocket packages removed:
    - [ ] `backend/internal/handler/slot.go` and `slot_test.go` deleted
    - [ ] `backend/internal/handler/live_stats.go` and `live_stats_test.go` (WebSocket) deleted
    - [ ] `backend/internal/websocket/` directory deleted entirely
- [ ] Obsolete legacy registration endpoint and service logic removed:
    - [ ] `POST /api/v0/auth/register` removed from `backend/internal/handler/auth.go`
    - [ ] `AuthService` interface in `backend/internal/handler/auth.go` cleaned of
          `RegisterWithGoogle`
    - [ ] `RegisterWithGoogle` and `Register` methods removed from
          `backend/internal/auth/service.go`
    - [ ] Legacy unit tests testing `/auth/register` removed from
          `backend/internal/handler/auth_test.go` and `backend/internal/auth/service_test.go`
- [ ] Application wiring updated:
    - [ ] `backend/cmd/arch-stats/router.go` cleaned of all slot and websocket route mounts
    - [ ] `backend/cmd/arch-stats/main.go` cleaned of slot, websocket, and `/auth/register`
          routes
    - [ ] `backend/cmd/arch-stats/adapters.go` cleaned of `slotServiceAdapter`
- [ ] Handler unit tests in `session_test.go` and `shot_test.go` verify request parsing,
      context archer authentication, response status codes, and error mapping.
- [ ] `cd backend && go test ./internal/handler/... ./cmd/arch-stats/... ./internal/auth/... -v`
      passes.

## Files to Create/Modify/Remove

| Action | Path |
| ------ | ---- |
| Modify | `backend/internal/handler/session.go` |
| Modify | `backend/internal/handler/session_test.go` |
| Modify | `backend/internal/handler/shot.go` |
| Modify | `backend/internal/handler/shot_test.go` |
| Modify | `backend/internal/handler/auth.go` |
| Modify | `backend/internal/handler/auth_test.go` |
| Modify | `backend/internal/auth/service.go` |
| Modify | `backend/internal/auth/service_test.go` |
| Modify | `backend/cmd/arch-stats/router.go` |
| Modify | `backend/cmd/arch-stats/main.go` |
| Modify | `backend/cmd/arch-stats/adapters.go` |
| Delete | `backend/internal/handler/slot.go` |
| Delete | `backend/internal/handler/slot_test.go` |
| Delete | `backend/internal/handler/live_stats.go` |
| Delete | `backend/internal/handler/live_stats_test.go` |
| Delete | `backend/internal/websocket/hub.go` |
| Delete | `backend/internal/websocket/hub_test.go` |
| Delete | `backend/internal/websocket/client.go` |
| Delete | `backend/internal/websocket/client_test.go` |

## Reference

- [PRD.md](../PRD.md)
- [router.go](../../../backend/cmd/arch-stats/router.go)
- [session.go](../../../backend/internal/handler/session.go)
- [auth.go](../../../backend/internal/handler/auth.go)
- [service.go](../../../backend/internal/auth/service.go)

## Steps

- [ ] **Step 1: Write failing handler unit tests**

  Update `backend/internal/handler/session_test.go` and `shot_test.go`:
    - Test `POST /api/v0/session` parses all new fields and returns 201
    - Test `GET /api/v0/session/open` returns open session or null
    - Test `PATCH /api/v0/session/{id}/close` parses reflections
    - Test `GET /api/v0/session/{id}/live-stats` returns JSON stats
    - Test `POST /api/v0/shot` creates shot with `session_id`

- [ ] **Step 2: Run handler tests to verify failure**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 3: Update `handler/session.go`**

  Implement updated route methods for session lifecycle, close reflection, soft deletion, and
  HTTP live stats retrieval.

- [ ] **Step 4: Update `handler/shot.go`**

  Implement updated routes for shot creation (single and batch), session queries, and deletion.

- [ ] **Step 5: Remove legacy `/auth/register` endpoint and service methods**

  In `backend/internal/handler/auth.go`, remove `Register` handler method and annotations,
  and remove `RegisterWithGoogle` from `AuthService` interface.

  In `backend/internal/auth/service.go`, remove `RegisterWithGoogle` and `Register` methods.

  In `backend/internal/handler/auth_test.go` and `backend/internal/auth/service_test.go`,
  remove unit tests verifying `/auth/register`.

- [ ] **Step 6: Delete obsolete handler and websocket files**

  ```bash
  cd backend
  rm -f internal/handler/slot.go internal/handler/slot_test.go
  rm -f internal/handler/live_stats.go internal/handler/live_stats_test.go
  rm -rf internal/websocket/
  ```

- [ ] **Step 7: Update router and main server wiring**

  Update `cmd/arch-stats/router.go`, `cmd/arch-stats/main.go`, and `cmd/arch-stats/adapters.go`.
  Remove `SlotHandler`, `/slot` route mounts, and `/auth/register` route registration. Wire the
  updated session and shot handlers.

- [ ] **Step 8: Run all backend tests to verify compilation and passing**

  ```bash
  cd backend && go test ./... -v
  ```

- [ ] **Step 9: Commit changes**

  ```bash
  git add backend/
  git commit -m "feat(api): refactor endpoints; tear down websockets, slots, and legacy auth"
  ```

## Verification

- `cd backend && go test ./... -v` passes with 0 failures.
- Obsolete routes `/api/v0/session/slot/*`, `/api/v0/stats/ws/*`, and `/api/v0/auth/register`
  are completely removed.
