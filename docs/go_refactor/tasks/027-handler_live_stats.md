# Task 027: Build HTTP Handler — Live Stats + WebSocket Upgrade

## Git Branch

`refactor/027-handler-live-stats`

## Objective

Implement the live stats HTTP handler in `internal/handler/`, porting the Python
`live_stats_router.py`. This provides a REST endpoint for getting stats and a WebSocket endpoint
for streaming real-time shot notifications. The WebSocket endpoint upgrades the HTTP connection
and registers the client with the WebSocket hub.

## Dependencies

- Task 026 (WebSocket hub)
- Task 015 (shot service — for stats retrieval)
- Task 018 (middleware — auth)
- Task 019 (handler helpers)

## Acceptance Criteria

- [ ] `backend/internal/handler/live_stats.go` implements `LiveStatsHandler` with methods:
    - `GetStats(w, r)` — GET `/api/v0/stats/{slot_id}` — returns live statistics for a slot
    - `WebSocketStats(w, r)` — GET `/api/v0/stats/ws/{slot_id}` — upgrades to WebSocket,
    registers client with hub, streams NOTIFY payloads as JSON messages
- [ ] The WebSocket handler:
    - Upgrades the HTTP connection using `nhooyr.io/websocket`
    - Creates a `Client` and registers it with the hub
    - Starts `WritePump` and `ReadPump` goroutines
    - Unregisters the client when the connection closes
- [ ] The stats endpoint is authenticated (requires auth middleware).
- [ ] Unit tests verify:
    - GetStats returns 200 + stats JSON for valid slot
    - GetStats returns 404 for non-existent slot
- [ ] `go test ./internal/handler/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/handler/live_stats.go` |
| Create | `backend/internal/handler/live_stats_test.go` |

## Reference

- Python router: [live_stats_router.py](file:///home/juanpa/Projects/arch-stats/backend/src/routers/v0/live_stats_router.py)
- 2 endpoints: REST stats + WebSocket streaming

## Steps

- [ ] **Step 1: Write failing tests for GetStats**

  Create `backend/internal/handler/live_stats_test.go`:
    - Define mock `liveStatsService` interface
    - Test GetStats returns 200 with stats data
    - Test GetStats returns 404 for unknown slot
    - (WebSocket upgrade is hard to unit test — covered by integration tests in task 041)

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 3: Implement `live_stats.go`**

  Implement the REST handler and WebSocket upgrade handler:

  ```go
  func (h *LiveStatsHandler) WebSocketStats(w http.ResponseWriter, r *http.Request) {
      slotID := chi.URLParam(r, "slot_id")

      conn, err := websocket.Accept(w, r, nil)
      if err != nil {
          // Accept already wrote the error response
          return
      }

      client := ws.NewClient(conn)
      h.hub.Register(client)
      defer h.hub.Unregister(client)

      // Block until client disconnects
      client.ReadPump(r.Context())
  }
  ```

- [ ] **Step 4: Run tests to verify they pass**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 5: Run go vet and build**

  ```bash
  cd backend && go vet ./... && go build ./...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "feat: add live stats handler with REST and WebSocket endpoints"
  ```

## Verification

- `cd backend && go test ./internal/handler/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
