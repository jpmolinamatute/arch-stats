# Task 026: Implement WebSocket Hub with pg LISTEN/NOTIFY

## Git Branch

`refactor/026-websocket-hub`

## Objective

Implement the WebSocket hub in `internal/websocket/` using Go's goroutines and channels. The hub
listens for PostgreSQL NOTIFY events on a dedicated goroutine and broadcasts them to all connected
WebSocket clients via a fan-out pattern. This replaces the Python async approach with Go's native
concurrency model.

## Dependencies

- Task 004 (database pool — pgx connection for LISTEN)
- Task 003 (config — `WSChannel` field)
- Task 006 (logging)

## Acceptance Criteria

- [ ] `backend/internal/websocket/hub.go` implements a `Hub` struct with:
    - `Run(ctx)` — starts the hub goroutine, listens for NOTIFY events on the configured channel
    - `Register(client)` — registers a new WebSocket client for broadcasts
    - `Unregister(client)` — removes a disconnected client
    - Internal broadcast channel distributing NOTIFY payloads to all registered clients
- [ ] `backend/internal/websocket/client.go` implements a `Client` struct with:
    - `WritePump(ctx)` — goroutine writing messages from the hub to the WebSocket connection
    - `ReadPump(ctx)` — goroutine reading (to detect client disconnect)
    - `Send` channel for receiving messages from the hub
- [ ] The hub uses a dedicated `pgx` connection (not from the pool) for `LISTEN` since it must
  hold the connection open indefinitely.
- [ ] Hub gracefully shuts down when context is cancelled: closes all client connections, stops
  the LISTEN goroutine.
- [ ] Unit tests verify:
    - Registering and unregistering clients changes the client count
    - Broadcasting a message reaches all registered clients
    - Unregistered clients do not receive messages
- [ ] `go test ./internal/websocket/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/websocket/hub.go` |
| Create | `backend/internal/websocket/hub_test.go` |
| Create | `backend/internal/websocket/client.go` |
| Delete | `backend/internal/websocket/.gitkeep` |
| Modify | `backend/go.mod` (add websocket dependency) |

## Reference

- Python live stats: [live_stats_manager.py](file:///home/juanpa/Projects/arch-stats/backend/src/core/live_stats_manager.py)
- High-level plan §9: goroutines + channels, central hub, fan-out

## Steps

- [ ] **Step 1: Add WebSocket dependency**

  ```bash
  cd backend
  go get nhooyr.io/websocket
  ```

- [ ] **Step 2: Write failing tests for the hub**

  Create `backend/internal/websocket/hub_test.go`:
    - Test register adds client to the hub
    - Test unregister removes client
    - Test broadcast sends message to all registered clients
    - Test broadcast skips unregistered clients
    - Use Go channels to simulate client Send channels (no real WebSocket needed)

- [ ] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend && go test ./internal/websocket/... -v
  ```

- [ ] **Step 4: Implement `client.go`**

  Define the `Client` struct with a `Send` channel and `Conn` field.

- [ ] **Step 5: Implement `hub.go`**

  Implement the hub with:
    - `register` channel (chan *Client)
    - `unregister` channel (chan *Client)
    - `broadcast` channel (chan []byte)
    - `clients` map[*Client]bool
    - `Run()` goroutine selecting on all channels
    - `ListenNotify()` goroutine using pgx `conn.WaitForNotification()`

- [ ] **Step 6: Run tests to verify they pass**

  ```bash
  cd backend && go test ./internal/websocket/... -v
  ```

- [ ] **Step 7: Run go vet and build**

  ```bash
  cd backend && go vet ./... && go build ./...
  ```

- [ ] **Step 8: Commit**

  ```bash
  rm -f backend/internal/websocket/.gitkeep
  git add -A
  git commit -m "feat: add WebSocket hub with goroutine fan-out and pg LISTEN/NOTIFY"
  ```

## Verification

- `cd backend && go test ./internal/websocket/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
