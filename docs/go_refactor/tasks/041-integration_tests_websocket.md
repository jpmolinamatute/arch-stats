# Task 041: Integration Tests — WebSocket / Live Stats

## Git Branch

`refactor/041-integration-tests-websocket`

## Objective

Write integration tests for the WebSocket hub and live stats endpoint. These tests verify that
the PostgreSQL LISTEN/NOTIFY → WebSocket fan-out pipeline works end-to-end: a NOTIFY event
fired in PostgreSQL is received by connected WebSocket clients via the hub.

## Dependencies

- Task 037 (integration test infrastructure)
- Task 026 (WebSocket hub implementation)
- Task 027 (live stats handler with WebSocket upgrade)

## Acceptance Criteria

- [ ] `backend/tests/integration/websocket_test.go` tests the following scenarios:
    - **Single client connection**: Connect a WebSocket client → verify the connection is
    accepted and the client receives a welcome/ack message (if applicable)
    - **NOTIFY broadcast**: Connect a WebSocket client → fire a `pg_notify` on the configured
    channel → verify the client receives the notification payload within a timeout
    - **Multiple clients fan-out**: Connect 3 WebSocket clients → fire a single `pg_notify` →
    verify all 3 clients receive the same message
    - **Client disconnect cleanup**: Connect a client → disconnect it → fire a `pg_notify` →
    verify no panic or error in the hub (remaining clients still receive messages)
    - **Invalid upgrade request**: Send a regular HTTP GET to the WebSocket endpoint (without
    upgrade headers) → verify it returns an appropriate error (400 or 426)
- [ ] Tests use `nhooyr.io/websocket` (or `gorilla/websocket`) client for WebSocket connections.
- [ ] Tests use the testcontainers PostgreSQL instance to fire real `pg_notify` events.
- [ ] Each test truncates tables and stops the hub after completion.
- [ ] `go test ./tests/integration/... -v -count=1 -run WebSocket` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/tests/integration/websocket_test.go` |

## Reference

- Python live stats tests: [test_stats_endpoint.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_stats_endpoint.py)
- Python live stats manager: [live_stats_manager.py](file:///home/juanpa/Projects/arch-stats/backend/src/core/live_stats_manager.py)

## Steps

- [ ] **Step 1: Write single client connection test**

  ```go
  func TestWebSocket_SingleClientConnection(t *testing.T) {
      ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
      defer cancel()
      t.Cleanup(func() { truncateAll(context.Background(), testPool) })

      // Start test server with WebSocket handler
      srv := newTestServer()
      defer srv.Close()

      // Connect WebSocket client
      wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v0/live-stats"
      conn, _, err := websocket.Dial(ctx, wsURL, nil)
      if err != nil {
          t.Fatalf("WebSocket dial failed: %v", err)
      }
      defer conn.Close(websocket.StatusNormalClosure, "test done")

      // Verify connection is alive (e.g., ping or read initial message)
      conn.Ping(ctx)
  }
  ```

- [ ] **Step 2: Write NOTIFY broadcast test**

  ```go
  func TestWebSocket_NotifyBroadcast(t *testing.T) {
      ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
      defer cancel()
      t.Cleanup(func() { truncateAll(context.Background(), testPool) })

      srv := newTestServer()
      defer srv.Close()

      // Connect WebSocket client
      wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v0/live-stats"
      conn, _, err := websocket.Dial(ctx, wsURL, nil)
      if err != nil {
          t.Fatalf("dial: %v", err)
      }
      defer conn.Close(websocket.StatusNormalClosure, "done")

      // Fire pg_notify
      payload := `{"event":"shot_recorded","data":{"slot_id":"123"}}`
      _, err = testPool.Exec(ctx, "SELECT pg_notify($1, $2)", "archy", payload)
      if err != nil {
          t.Fatalf("pg_notify: %v", err)
      }

      // Read message from WebSocket
      _, msg, err := conn.Read(ctx)
      if err != nil {
          t.Fatalf("read: %v", err)
      }

      if !strings.Contains(string(msg), "shot_recorded") {
          t.Errorf("message = %q, want to contain 'shot_recorded'", msg)
      }
  }
  ```

- [ ] **Step 3: Write multiple clients fan-out test**

  Connect 3 clients, fire one NOTIFY, verify all 3 receive the message.

- [ ] **Step 4: Write client disconnect cleanup test**

  Connect 2 clients, disconnect one, fire NOTIFY, verify the remaining client still receives
  the message and no panic occurs.

- [ ] **Step 5: Write invalid upgrade request test**

  ```go
  func TestWebSocket_InvalidUpgrade(t *testing.T) {
      srv := newTestServer()
      defer srv.Close()

      // Regular HTTP GET without WebSocket upgrade headers
      resp, err := http.Get(srv.URL + "/api/v0/live-stats")
      if err != nil {
          t.Fatalf("GET: %v", err)
      }
      defer resp.Body.Close()

      if resp.StatusCode == 200 {
          t.Errorf("expected error status for non-WebSocket request, got %d", resp.StatusCode)
      }
  }
  ```

- [ ] **Step 6: Run WebSocket integration tests**

  ```bash
  cd backend
  go test ./tests/integration/... -v -count=1 -run WebSocket
  ```

- [ ] **Step 7: Run go vet**

  ```bash
  cd backend
  go vet ./...
  ```

- [ ] **Step 8: Commit**

  ```bash
  git add -A
  git commit -m "test: add WebSocket integration tests with pg_notify broadcast"
  ```

## Verification

- `cd backend && go test ./tests/integration/... -v -count=1 -run WebSocket` — all tests pass.
- `cd backend && go vet ./...` — clean.
- WebSocket tests exercise the real PostgreSQL NOTIFY mechanism (not mocked).
