# WebSocket / Live Stats Integration Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement end-to-end integration tests for the WebSocket hub and live stats endpoint (`backend/tests/integration/websocket_test.go`), verifying PostgreSQL LISTEN/NOTIFY fan-out, single client connection, multi-client broadcast, client disconnect cleanup, invalid upgrade rejection, and hub shutdown against a real PostgreSQL 17 container, and mark Task 041 as completed.

**Architecture:** Integration tests live in `backend/tests/integration/` under `package integration_test`. The tests use `testcontainers-go`'s PostgreSQL 17 instance with Goose migrations applied. A dedicated `ws.Hub` runs against the database via `testDSN` on a configured channel, with `handler.LiveStatsHandler` mounted on an `httptest.Server`. WebSocket clients use `github.com/coder/websocket` to connect, ping, and read broadcast events triggered via `testPool.Exec("SELECT pg_notify(...)")`. Every test registers cleanup to close servers, disconnect clients, cancel hub context, and truncate tables.

**Tech Stack:** Go 1.27+, `github.com/coder/websocket` (v1.8.15), `github.com/go-chi/chi/v5`, `github.com/jackc/pgx/v5` (`pgxpool`), `testcontainers-go` (v0.44+), PostgreSQL 17 container.

**Spec:** [docs/go_refactor/tasks/041-integration_tests_websocket.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/041-integration_tests_websocket.md)

## Global Constraints

- Target branch: `refactor/041-integration-tests-websocket`
- File to create: `backend/tests/integration/websocket_test.go`
- Infrastructure file to modify: `backend/tests/integration/testmain_test.go` (export `testDSN`)
- Package declaration: `package integration_test`
- Every test function must register table and resource cleanup via `t.Cleanup`
- Tests must execute against real PostgreSQL database in `testcontainers-go`, testing actual `LISTEN` / `pg_notify` SQL execution and fan-out
- All tests must pass with `go test ./tests/integration/... -v -count=1 -run WebSocket` and `go vet ./...` reporting zero issues
- At the end of implementation, mark all checklist items in [docs/go_refactor/tasks/041-integration_tests_websocket.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/041-integration_tests_websocket.md) as completed (`[x]`) and update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) marking all tasks as `DONE`

---

## File Structure

```
backend/
└── tests/
    └── integration/
        ├── testmain_test.go         # [MODIFY] Export testDSN package-level variable initialized from testcontainers
        └── websocket_test.go        # [NEW] Integration tests for WebSocket hub and live stats endpoint
docs/
├── plans/
│   ├── task.md                      # [MODIFY] Live checklist table tracking Task 041
│   └── 2026-09-06-integration-tests-websocket.md # [NEW] This implementation plan document
└── go_refactor/
    └── tasks/
        └── 041-integration_tests_websocket.md    # [MODIFY] Mark all acceptance criteria and steps as completed
```

---

### Task 1: Git Branch Setup, Test Infrastructure DSN Export & Plan Tracking

**Files:**
- Modify: `backend/tests/integration/testmain_test.go`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: `pgContainer.ConnectionString(ctx, "sslmode=disable")` in `TestMain`
- Produces:
  - Exported package-level variable `var testDSN string` accessible in `package integration_test`
  - Checked out git branch `refactor/041-integration-tests-websocket`
  - Live task tracking table in `docs/plans/task.md`

- [ ] **Step 1: Check out feature branch**

```bash
git checkout -b refactor/041-integration-tests-websocket
```

- [ ] **Step 2: Update `backend/tests/integration/testmain_test.go` to export `testDSN`**

In `backend/tests/integration/testmain_test.go`, update package variable declaration:

```go
var (
	testPool *pgxpool.Pool
	testDSN  string
)
```

And in `TestMain` after obtaining connection string:

```go
	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		os.Exit(1)
	}
	testDSN = dsn
```

- [ ] **Step 3: Initialize live task tracker in `docs/plans/task.md`**

Write table tracker in `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup, Test Infrastructure DSN Export & Plan Tracking | IN_PROGRESS | Switch to branch `refactor/041-integration-tests-websocket`, export `testDSN`, and initialize live task tracker |
| Task 2: WebSocket Test Environment & Single Client Connection Test | TODO | Implement test harness (`setupWSTestEnv`) and `TestWebSocket_SingleClientConnection` |
| Task 3: PostgreSQL NOTIFY Broadcast & Fan-Out Tests | TODO | Implement `TestWebSocket_NotifyBroadcast` and `TestWebSocket_MultipleClientsFanOut` |
| Task 4: Client Disconnect Cleanup & Invalid Upgrade Request Tests | TODO | Implement `TestWebSocket_ClientDisconnectCleanup`, `TestWebSocket_InvalidUpgrade`, and `TestWebSocket_GracefulShutdown` |
| Task 5: Full Test Suite Verification, Vet, Linting & Task Completion | TODO | Run test suite, verify vet/lint, mark Task 041 items done in spec and live tracker, and commit |
```

- [ ] **Step 4: Verify smoke test passes with updated `testmain_test.go`**

Run: `cd backend && go test ./tests/integration/... -v -run TestSmoke`
Expected: PASS

---

### Task 2: WebSocket Test Environment & Single Client Connection Test

**Files:**
- Create: `backend/tests/integration/websocket_test.go`

**Interfaces:**
- Consumes:
  - `testDSN` (from `testmain_test.go`)
  - `testPool` (from `testmain_test.go`)
  - `truncateAll(ctx, testPool)` (from `helpers_test.go`)
  - `ws.NewHub(dsn, channel, logger)` (from `internal/websocket`)
  - `handler.NewLiveStatsHandler(svc, hub)` (from `internal/handler`)
  - `coderws "github.com/coder/websocket"`
- Produces:
  - `waitForCondition(timeout time.Duration, cond func() bool) bool`
  - `setupWSTestEnv(t *testing.T, customChannel ...string) *wsTestEnv`
  - `TestWebSocket_SingleClientConnection(t *testing.T)`

- [ ] **Step 1: Write initial `websocket_test.go` with test harness and `TestWebSocket_SingleClientConnection`**

Create `backend/tests/integration/websocket_test.go`:

```go
package integration_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	ws "github.com/jpmolinamatute/arch-stats/backend/internal/websocket"
)

func waitForCondition(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cond()
}

type wsTestEnv struct {
	srv     *httptest.Server
	hub     *ws.Hub
	channel string
	wsURL   string
	slotID  uuid.UUID
	cancel  context.CancelFunc
}

func setupWSTestEnv(t *testing.T, customChannel ...string) *wsTestEnv {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	channel := "archy"
	if len(customChannel) > 0 && customChannel[0] != "" {
		channel = customChannel[0]
	}

	hub := ws.NewHub(testDSN, channel, slog.Default())
	go hub.Run(ctx)

	slotID := uuid.New()
	liveStatsHandler := handler.NewLiveStatsHandler(nil, hub)

	r := chi.NewRouter()
	r.Route("/api/v0", func(r chi.Router) {
		r.Route("/stats", liveStatsHandler.Routes)
		r.Get("/live-stats", func(w http.ResponseWriter, r *http.Request) {
			rctx := chi.RouteContext(r.Context())
			rctx.URLParams.Add("slot_id", slotID.String())
			liveStatsHandler.WebSocketStats(w, r)
		})
	})

	srv := httptest.NewServer(r)

	t.Cleanup(func() {
		srv.Close()
		cancel()
		_ = truncateAll(context.Background(), testPool)
	})

	// Allow brief window for LISTEN connection to establish on PostgreSQL
	time.Sleep(100 * time.Millisecond)

	wsURL := fmt.Sprintf("ws%s/api/v0/stats/ws/%s", strings.TrimPrefix(srv.URL, "http"), slotID)

	return &wsTestEnv{
		srv:     srv,
		hub:     hub,
		channel: channel,
		wsURL:   wsURL,
		slotID:  slotID,
		cancel:  cancel,
	}
}

func TestWebSocket_SingleClientConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	// Connect WebSocket client
	conn, resp, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	// Verify connection is alive via ping
	if err := conn.Ping(ctx); err != nil {
		t.Fatalf("ping failed: %v", err)
	}

	// Verify client is registered with the hub
	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 registered client, got %d", env.hub.ClientCount())
	}

	// Close connection cleanly and verify unregistration
	_ = conn.Close(websocket.StatusNormalClosure, "done")
	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 0 }) {
		t.Fatalf("expected 0 registered clients after close, got %d", env.hub.ClientCount())
	}
}
```

- [ ] **Step 2: Run `TestWebSocket_SingleClientConnection` to verify it passes**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestWebSocket_SingleClientConnection`
Expected: PASS

- [ ] **Step 3: Update `docs/plans/task.md`**

Mark Task 1 and Task 2 status in `docs/plans/task.md`.

---

### Task 3: PostgreSQL NOTIFY Broadcast & Fan-Out Tests

**Files:**
- Modify: `backend/tests/integration/websocket_test.go`

**Interfaces:**
- Consumes: `setupWSTestEnv`, `testPool.Exec("SELECT pg_notify($1, $2)", ...)`, `websocket.Conn.Read(ctx)`
- Produces:
  - `TestWebSocket_NotifyBroadcast(t *testing.T)`
  - `TestWebSocket_MultipleClientsFanOut(t *testing.T)`

- [ ] **Step 1: Append `TestWebSocket_NotifyBroadcast` and `TestWebSocket_MultipleClientsFanOut`**

Add to `backend/tests/integration/websocket_test.go`:

```go
func TestWebSocket_NotifyBroadcast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	conn, _, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 registered client, got %d", env.hub.ClientCount())
	}

	// Fire pg_notify on the configured channel
	payload := fmt.Sprintf(`{"event":"shot_recorded","data":{"slot_id":"%s","score":10}}`, env.slotID)
	if _, err := testPool.Exec(ctx, "SELECT pg_notify($1, $2)", env.channel, payload); err != nil {
		t.Fatalf("pg_notify failed: %v", err)
	}

	// Read message from WebSocket client
	readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
	defer readCancel()
	_, msg, err := conn.Read(readCtx)
	if err != nil {
		t.Fatalf("reading websocket message failed: %v", err)
	}

	if !strings.Contains(string(msg), "shot_recorded") {
		t.Errorf("message %q does not contain 'shot_recorded'", string(msg))
	}
	if !strings.Contains(string(msg), env.slotID.String()) {
		t.Errorf("message %q does not contain slot ID %s", string(msg), env.slotID)
	}
}

func TestWebSocket_MultipleClientsFanOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	const clientCount = 3
	conns := make([]*websocket.Conn, clientCount)
	for i := 0; i < clientCount; i++ {
		conn, _, err := websocket.Dial(ctx, env.wsURL, nil)
		if err != nil {
			t.Fatalf("client %d dial failed: %v", i, err)
		}
		defer conn.Close(websocket.StatusNormalClosure, "test done")
		conns[i] = conn
	}

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == clientCount }) {
		t.Fatalf("expected %d registered clients, got %d", clientCount, env.hub.ClientCount())
	}

	// Fire a single pg_notify event
	payload := `{"event":"fanout_test","data":{"status":"ok"}}`
	if _, err := testPool.Exec(ctx, "SELECT pg_notify($1, $2)", env.channel, payload); err != nil {
		t.Fatalf("pg_notify failed: %v", err)
	}

	// Verify all clients receive the exact same message
	for i, conn := range conns {
		readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
		_, msg, err := conn.Read(readCtx)
		readCancel()
		if err != nil {
			t.Fatalf("client %d failed to read broadcast: %v", i, err)
		}
		if string(msg) != payload {
			t.Errorf("client %d expected %q, got %q", i, payload, string(msg))
		}
	}
}
```

- [ ] **Step 2: Run broadcast and fan-out tests to verify they pass**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run "TestWebSocket_NotifyBroadcast|TestWebSocket_MultipleClientsFanOut"`
Expected: PASS

- [ ] **Step 3: Update `docs/plans/task.md`**

Mark Task 3 status as DONE in `docs/plans/task.md`.

---

### Task 4: Client Disconnect Cleanup, Invalid Upgrade & Graceful Shutdown Tests

**Files:**
- Modify: `backend/tests/integration/websocket_test.go`

**Interfaces:**
- Consumes: `setupWSTestEnv`, `testPool.Exec(...)`, `http.Get(...)`
- Produces:
  - `TestWebSocket_ClientDisconnectCleanup(t *testing.T)`
  - `TestWebSocket_InvalidUpgrade(t *testing.T)`
  - `TestWebSocket_GracefulShutdown(t *testing.T)`

- [ ] **Step 1: Append `TestWebSocket_ClientDisconnectCleanup`, `TestWebSocket_InvalidUpgrade`, and `TestWebSocket_GracefulShutdown`**

Add to `backend/tests/integration/websocket_test.go`:

```go
func TestWebSocket_ClientDisconnectCleanup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	conn1, _, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("client 1 dial failed: %v", err)
	}
	defer conn1.Close(websocket.StatusNormalClosure, "test done")

	conn2, _, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("client 2 dial failed: %v", err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "test done")

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 2 }) {
		t.Fatalf("expected 2 registered clients, got %d", env.hub.ClientCount())
	}

	// Disconnect client 1
	_ = conn1.Close(websocket.StatusNormalClosure, "client 1 disconnecting")

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 registered client after disconnect, got %d", env.hub.ClientCount())
	}

	// Fire pg_notify after client 1 disconnected
	payload := `{"event":"after_disconnect","remaining":1}`
	if _, err := testPool.Exec(ctx, "SELECT pg_notify($1, $2)", env.channel, payload); err != nil {
		t.Fatalf("pg_notify failed: %v", err)
	}

	// Verify remaining client (conn2) receives the notification with no panics
	readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
	defer readCancel()
	_, msg, err := conn2.Read(readCtx)
	if err != nil {
		t.Fatalf("client 2 failed to read notification: %v", err)
	}
	if string(msg) != payload {
		t.Errorf("client 2 expected %q, got %q", payload, string(msg))
	}
}

func TestWebSocket_InvalidUpgrade(t *testing.T) {
	env := setupWSTestEnv(t)

	t.Run("regular HTTP GET to valid slot without upgrade headers", func(t *testing.T) {
		resp, err := http.Get(env.srv.URL + "/api/v0/stats/ws/" + env.slotID.String())
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Errorf("expected error status for non-WebSocket request, got %d", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUpgradeRequired {
			t.Errorf("expected status 400 or 426, got %d", resp.StatusCode)
		}
	})

	t.Run("regular HTTP GET to /api/v0/live-stats without upgrade headers", func(t *testing.T) {
		resp, err := http.Get(env.srv.URL + "/api/v0/live-stats")
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Errorf("expected error status for non-WebSocket request, got %d", resp.StatusCode)
		}
	})

	t.Run("GET with invalid slot_id UUID returns 422", func(t *testing.T) {
		resp, err := http.Get(env.srv.URL + "/api/v0/stats/ws/not-a-valid-uuid")
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("expected status 422 for invalid UUID, got %d", resp.StatusCode)
		}
	})
}

func TestWebSocket_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	conn, _, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 client, got %d", env.hub.ClientCount())
	}

	// Cancel the hub context to simulate graceful shutdown
	env.cancel()

	// Reading from client should return a close error or EOF
	readCtx, readCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer readCancel()
	_, _, err = conn.Read(readCtx)
	if err == nil {
		t.Fatal("expected error reading after hub shutdown, got nil")
	}

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 0 }) {
		t.Fatalf("expected 0 clients after hub shutdown, got %d", env.hub.ClientCount())
	}
}
```

- [ ] **Step 2: Run all WebSocket integration tests to verify they pass**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run WebSocket`
Expected: PASS

- [ ] **Step 3: Update `docs/plans/task.md`**

Mark Task 4 status as DONE in `docs/plans/task.md`.

---

### Task 5: Full Test Suite Verification, Vet, Linting, Documentation Updates & Commit

**Files:**
- Modify: `docs/go_refactor/tasks/041-integration_tests_websocket.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Complete test suite and verification tooling
- Produces:
  - Clean `go test ./tests/integration/... -v -count=1 -run WebSocket` output
  - Clean `go vet ./...` output
  - Clean `golangci-lint run ./...`
  - Marked `041-integration_tests_websocket.md` with all checkboxes `[x]`
  - Marked `docs/plans/task.md` with all tasks `DONE`
  - Final git commit

- [ ] **Step 1: Run WebSocket integration tests**

Run:
```bash
cd backend && go test ./tests/integration/... -v -count=1 -run WebSocket
```
Expected: PASS (all tests pass)

- [ ] **Step 2: Run all integration tests to confirm zero regressions**

Run:
```bash
cd backend && go test ./tests/integration/... -v -count=1
```
Expected: PASS

- [ ] **Step 3: Run `go vet` across entire backend**

Run:
```bash
cd backend && go vet ./...
```
Expected: clean (zero issues)

- [ ] **Step 4: Run Go linting**

Run:
```bash
cd backend && golangci-lint run ./...
```
Expected: clean (zero issues)

- [ ] **Step 5: Mark all acceptance criteria and steps as completed in `docs/go_refactor/tasks/041-integration_tests_websocket.md`**

Update `docs/go_refactor/tasks/041-integration_tests_websocket.md`:
- Mark all acceptance criteria checkboxes from `[ ]` to `[x]`
- Mark all step checkboxes from `[ ]` to `[x]`

- [ ] **Step 6: Update `docs/plans/task.md` to mark all tasks `DONE`**

Update `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup, Test Infrastructure DSN Export & Plan Tracking | DONE | Switch to branch `refactor/041-integration-tests-websocket`, export `testDSN`, and initialize live task tracker |
| Task 2: WebSocket Test Environment & Single Client Connection Test | DONE | Implement test harness (`setupWSTestEnv`) and `TestWebSocket_SingleClientConnection` |
| Task 3: PostgreSQL NOTIFY Broadcast & Fan-Out Tests | DONE | Implement `TestWebSocket_NotifyBroadcast` and `TestWebSocket_MultipleClientsFanOut` |
| Task 4: Client Disconnect Cleanup & Invalid Upgrade Request Tests | DONE | Implement `TestWebSocket_ClientDisconnectCleanup`, `TestWebSocket_InvalidUpgrade`, and `TestWebSocket_GracefulShutdown` |
| Task 5: Full Test Suite Verification, Vet, Linting & Task Completion | DONE | Run test suite, verify vet/lint, mark Task 041 items done in spec and live tracker, and commit |
```

- [ ] **Step 7: Commit changes**

```bash
git add backend/tests/integration/testmain_test.go \
        backend/tests/integration/websocket_test.go \
        docs/go_refactor/tasks/041-integration_tests_websocket.md \
        docs/plans/task.md \
        docs/plans/2026-09-06-integration-tests-websocket.md
git commit -m "test: add WebSocket integration tests with pg_notify broadcast and mark task 041 done"
```
