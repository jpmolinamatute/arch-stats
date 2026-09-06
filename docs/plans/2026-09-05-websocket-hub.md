# Task 026: Implement WebSocket Hub with pg LISTEN/NOTIFY Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a thread-safe, goroutine-based WebSocket hub in `backend/internal/websocket/` that receives PostgreSQL `LISTEN/NOTIFY` events on a dedicated `pgx` connection and fans out notifications to connected WebSocket clients, with complete unit test coverage, clean shutdown, and task completion tracking.

**Architecture:**
- `backend/internal/websocket/client.go`: Defines `Client` representing an active WebSocket connection with a buffered `Send` channel (`chan []byte`), `WritePump(ctx)` streaming messages from the channel to the client connection using `nhooyr.io/websocket`, and `ReadPump(ctx)` reading from the connection to detect client disconnects.
- `backend/internal/websocket/hub.go`: Defines `Hub` maintaining connected clients (`map[*Client]bool`), synchronization (`sync.RWMutex`), and channels (`register`, `unregister`, `broadcast`). Runs an event loop in `Run(ctx)` that registers/unregisters clients, distributes payloads, gracefully evicts slow clients, and shuts down cleanly on context cancellation. In production, spawns a dedicated goroutine `listenNotify(ctx)` using `pgx.Connect(ctx, dsn)` to `LISTEN` to the configured channel and pump payloads into the broadcast channel.
- `backend/internal/websocket/hub_test.go` & `client_test.go`: Comprehensive unit tests verifying client count tracking, broadcast delivery to registered clients, isolation from unregistered clients, slow client eviction, graceful shutdown, pump lifecycles, and context cancellation.
- Acceptance & Task Tracking: Mark all checklist items in `docs/go_refactor/tasks/026-websocket_hub.md` and `docs/plans/task.md` as completed.

**Tech Stack:** Go 1.24+, `github.com/jackc/pgx/v5`, `nhooyr.io/websocket`, `net/http/httptest`, `log/slog`, `sync`.

**Spec:**
- [docs/go_refactor/tasks/026-websocket_hub.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/026-websocket_hub.md)
- [backend-old/src/core/live_stats_manager.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/core/live_stats_manager.py)
- [backend-old/src/models/live_stats_model.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/models/live_stats_model.py)
- [docs/go_refactor/tasks/027-handler_live_stats.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/027-handler_live_stats.md)
- [docs/go_refactor/tasks/041-integration_tests_websocket.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/041-integration_tests_websocket.md)

## Global Constraints

- Git branch: `refactor/026-websocket-hub`
- Concurrency model: Goroutines + channels + mutex synchronization; strictly zero data races (`go test -race ./...`).
- Database LISTEN connection: Use a dedicated `pgx.Connect` connection (never acquire from `pgxpool.Pool` because `LISTEN` holds the connection open indefinitely).
- Client message buffering: `Send` channel buffered (default 256 messages) to prevent slow network writes from blocking the hub broadcast loop.
- Eviction policy: If a client's send buffer is full, the hub logs a warning, unregisters the client, and closes its send channel to protect the system.
- Graceful shutdown: When `ctx` is cancelled, close all client connections, close send channels, stop the `LISTEN` goroutine, and release the dedicated database connection.
- Clean compilation & linting: `go vet ./...` and `golangci-lint run ./...` report zero issues, `gofumpt` applied.
- Task completion: Mark all acceptance criteria and steps in `docs/go_refactor/tasks/026-websocket_hub.md` as done (`[x]`) and update `docs/plans/task.md` with all tasks `DONE`.

---

## File Structure

```
backend/
├── go.mod                                # [MODIFY] Add nhooyr.io/websocket dependency
├── go.sum                                # [MODIFY] Update checksums
└── internal/
    └── websocket/
        ├── .gitkeep                      # [DELETE] Remove placeholder
        ├── client.go                     # [NEW] Client struct, NewClient, WritePump, ReadPump
        ├── client_test.go                # [NEW] Unit tests for Client WritePump and ReadPump
        ├── hub.go                        # [NEW] Hub struct, NewHub, Run, Register, Unregister, Broadcast, ClientCount, listenNotify
        └── hub_test.go                   # [NEW] Unit tests for Hub registration, fan-out, eviction, shutdown
docs/
├── plans/
│   ├── task.md                           # [MODIFY] Track Task 026 live checklist progress (table-only)
│   └── 2026-09-05-websocket-hub.md       # [NEW] This implementation plan document
└── go_refactor/
    └── tasks/
        └── 026-websocket_hub.md          # [MODIFY] Mark all acceptance criteria and steps as completed
```

---

### Task 1: Git Branch Setup, Live Tracker Initialization & Dependency Management

**Files:**
- Modify: `docs/plans/task.md`
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`

**Interfaces:**
- Consumes: Clean working tree on `main`.
- Produces: Checked out branch `refactor/026-websocket-hub`, initialized `docs/plans/task.md`, and `nhooyr.io/websocket` added to `backend/go.mod`.

- [ ] **Step 1: Create and switch to git branch**

```bash
git checkout -b refactor/026-websocket-hub
```

- [ ] **Step 2: Initialize `docs/plans/task.md` with Task 026 checklist table**

Update `docs/plans/task.md` to:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup, Live Tracker & Dependency Management | IN_PROGRESS | Switch to branch `refactor/026-websocket-hub`, initialize tracker, and add `nhooyr.io/websocket` |
| Task 2: Client Implementation (`client.go` & `client_test.go`) | PENDING | Implement `Client`, `NewClient`, `WritePump`, and `ReadPump` with unit tests |
| Task 3: Hub Implementation (`hub.go` & `hub_test.go`) | PENDING | Implement `Hub` with registration, broadcast fan-out, pg LISTEN, and unit tests |
| Task 4: End-to-End Suite Verification, Formatting & Linting | PENDING | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 5: Mark Tasks as Completed in Task Spec and Live Tracker | PENDING | Mark `026-websocket_hub.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Add `nhooyr.io/websocket` dependency**

```bash
cd backend
go get nhooyr.io/websocket
go mod tidy
```

- [ ] **Step 4: Verify dependencies compile**

```bash
cd backend && go build ./...
```

- [ ] **Step 5: Update Task 1 status to DONE**

Update `docs/plans/task.md`:

```markdown
| Task 1: Git Branch Setup, Live Tracker & Dependency Management | DONE | Switch to branch `refactor/026-websocket-hub`, initialize tracker, and add `nhooyr.io/websocket` |
```

- [ ] **Step 6: Commit**

```bash
git add backend/go.mod backend/go.sum docs/plans/task.md
git commit -m "chore: add nhooyr.io/websocket dependency and setup task tracker"
```

---

### Task 2: Client Implementation (`client.go` & `client_test.go`)

**Files:**
- Create: `backend/internal/websocket/client.go`
- Create: `backend/internal/websocket/client_test.go`

**Interfaces:**
- Consumes: `nhooyr.io/websocket`.
- Produces: `Client` struct, `NewClient(conn *websocket.Conn) *Client`, `WritePump(ctx context.Context)`, `ReadPump(ctx context.Context)`, `Send chan []byte`, `Conn *websocket.Conn`.

- [ ] **Step 1: Write failing tests for Client**

Create `backend/internal/websocket/client_test.go`:

```go
package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestClient_WritePump_DeliversMessages(t *testing.T) {
	serverReceived := make(chan string, 1)

	// Create test server accepting websocket connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept failed: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "done")

		client := NewClient(conn)
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		go client.WritePump(ctx)

		// Send message through client.Send
		client.Send <- []byte("test message payload")

		// Keep connection open briefly for client read
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	// Client dials the test server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	_, msg, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if string(msg) != "test message payload" {
		t.Fatalf("expected 'test message payload', got %q", string(msg))
	}
}

func TestClient_ReadPump_DetectsDisconnect(t *testing.T) {
	readPumpFinished := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept failed: %v", err)
			return
		}

		client := NewClient(conn)
		go func() {
			client.ReadPump(r.Context())
			close(readPumpFinished)
		}()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	// Close dial connection to trigger ReadPump exit on server
	_ = conn.Close(websocket.StatusNormalClosure, "client disconnect")

	select {
	case <-readPumpFinished:
		// Succeeded
	case <-time.After(3 * time.Second):
		t.Fatal("ReadPump did not detect disconnect within timeout")
	}
}

func TestClient_NilConn_SafeOperations(t *testing.T) {
	client := NewClient(nil)
	if client.Send == nil {
		t.Fatal("expected non-nil Send channel")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Both pumps should exit gracefully on context cancellation when Conn is nil
	done := make(chan struct{})
	go func() {
		client.WritePump(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("WritePump did not exit on cancelled context with nil conn")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/websocket/... -v
```

Expected: FAIL with `undefined: NewClient`.

- [ ] **Step 3: Implement `client.go`**

Create `backend/internal/websocket/client.go`:

```go
// Package websocket provides a real-time WebSocket hub and client connections.
package websocket

import (
	"context"
	"time"

	"nhooyr.io/websocket"
)

const (
	// defaultSendBufferSize is the buffer size for a client's outbound message channel.
	defaultSendBufferSize = 256

	// writeTimeout is the maximum duration to wait for a write to complete.
	writeTimeout = 5 * time.Second
)

// Client represents a single connected WebSocket client.
type Client struct {
	// hub is the central hub managing this client.
	hub *Hub

	// Conn is the underlying WebSocket connection.
	Conn *websocket.Conn

	// Send is the buffered channel of outbound messages to be delivered to the client.
	Send chan []byte
}

// NewClient creates a new Client with a buffered send channel.
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
		Send: make(chan []byte, defaultSendBufferSize),
	}
}

// WritePump pumps messages from the client's Send channel to the WebSocket connection.
// It terminates when the context is cancelled, the Send channel is closed, or a write error occurs.
func (c *Client) WritePump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.Send:
			if !ok {
				// Hub closed the channel
				if c.Conn != nil {
					_ = c.Conn.Close(websocket.StatusNormalClosure, "hub closed channel")
				}
				return
			}

			if c.Conn == nil {
				continue
			}

			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.Conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				if c.hub != nil {
					c.hub.Unregister(c)
				}
				return
			}
		}
	}
}

// ReadPump pumps messages from the WebSocket connection to detect client disconnect.
// It blocks until the connection is closed or an error occurs, then cleans up the client.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		if c.hub != nil {
			c.hub.Unregister(c)
		}
		if c.Conn != nil {
			_ = c.Conn.Close(websocket.StatusNormalClosure, "read pump closed")
		}
	}()

	if c.Conn == nil {
		<-ctx.Done()
		return
	}

	for {
		_, _, err := c.Conn.Read(ctx)
		if err != nil {
			return
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/websocket/... -v -run TestClient_
```

Expected: PASS.

- [ ] **Step 5: Update Task 2 status to DONE**

Update `docs/plans/task.md`:

```markdown
| Task 2: Client Implementation (`client.go` & `client_test.go`) | DONE | Implement `Client`, `NewClient`, `WritePump`, and `ReadPump` with unit tests |
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/websocket/client.go backend/internal/websocket/client_test.go docs/plans/task.md
git commit -m "feat(websocket): implement Client struct with WritePump and ReadPump"
```

---

### Task 3: Hub Implementation (`hub.go` & `hub_test.go`)

**Files:**
- Create: `backend/internal/websocket/hub.go`
- Create: `backend/internal/websocket/hub_test.go`
- Delete: `backend/internal/websocket/.gitkeep`

**Interfaces:**
- Consumes: `Client`, `github.com/jackc/pgx/v5`.
- Produces: `Hub` struct, `NewHub(dsn string, channel string, logger *slog.Logger) *Hub`, `Run(ctx context.Context)`, `Register(client *Client)`, `Unregister(client *Client)`, `Broadcast(payload []byte)`, `ClientCount() int`, `HasClient(client *Client) bool`.

- [ ] **Step 1: Write failing tests for Hub**

Create `backend/internal/websocket/hub_test.go`:

```go
package websocket

import (
	"context"
	"testing"
	"time"
)

// waitForCondition polls until condition is true or timeout expires.
func waitForCondition(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

func TestHub_RegisterAndUnregister(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub("", "", nil)
	go hub.Run(ctx)

	client1 := NewClient(nil)
	client2 := NewClient(nil)

	// Initially 0 clients
	if count := hub.ClientCount(); count != 0 {
		t.Fatalf("expected 0 clients, got %d", count)
	}

	// Register client1
	hub.Register(client1)
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 client after registering client1, got %d", hub.ClientCount())
	}
	if !hub.HasClient(client1) {
		t.Fatal("expected hub to have client1")
	}

	// Register client2
	hub.Register(client2)
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 }) {
		t.Fatalf("expected 2 clients after registering client2, got %d", hub.ClientCount())
	}
	if !hub.HasClient(client2) {
		t.Fatal("expected hub to have client2")
	}

	// Unregister client1
	hub.Unregister(client1)
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 client after unregistering client1, got %d", hub.ClientCount())
	}
	if hub.HasClient(client1) {
		t.Fatal("expected hub to not have client1")
	}

	// Unregister client2
	hub.Unregister(client2)
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 0 }) {
		t.Fatalf("expected 0 clients after unregistering client2, got %d", hub.ClientCount())
	}
	if hub.HasClient(client2) {
		t.Fatal("expected hub to not have client2")
	}
}

func TestHub_Broadcast_ReachesAllRegisteredClients(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub("", "", nil)
	go hub.Run(ctx)

	client1 := NewClient(nil)
	client2 := NewClient(nil)
	client3Unregistered := NewClient(nil)

	hub.Register(client1)
	hub.Register(client2)

	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 }) {
		t.Fatalf("expected 2 registered clients, got %d", hub.ClientCount())
	}

	payload := []byte(`{"event":"shot_insert","score":10}`)
	hub.Broadcast(payload)

	// Verify client1 received message
	select {
	case msg := <-client1.Send:
		if string(msg) != string(payload) {
			t.Fatalf("client1 expected %s, got %s", payload, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("client1 timed out waiting for broadcast")
	}

	// Verify client2 received message
	select {
	case msg := <-client2.Send:
		if string(msg) != string(payload) {
			t.Fatalf("client2 expected %s, got %s", payload, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("client2 timed out waiting for broadcast")
	}

	// Verify unregistered client3 did NOT receive message
	select {
	case msg := <-client3Unregistered.Send:
		t.Fatalf("unregistered client received unexpected message: %s", msg)
	default:
		// Succeeded, nothing received
	}
}

func TestHub_Broadcast_SkipsUnregisteredClients(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub("", "", nil)
	go hub.Run(ctx)

	client1 := NewClient(nil)
	client2 := NewClient(nil)

	hub.Register(client1)
	hub.Register(client2)
	waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 })

	// Unregister client1
	hub.Unregister(client1)
	waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 1 })

	// client1.Send should be closed
	_, ok := <-client1.Send
	if ok {
		t.Fatal("expected client1.Send to be closed upon unregister")
	}

	payload := []byte(`{"event":"update"}`)
	hub.Broadcast(payload)

	// Verify client2 receives message
	select {
	case msg := <-client2.Send:
		if string(msg) != string(payload) {
			t.Fatalf("client2 expected %s, got %s", payload, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("client2 timed out waiting for broadcast")
	}
}

func TestHub_SlowClientEviction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub("", "", nil)
	go hub.Run(ctx)

	// Create client with tiny buffer of 1 message
	slowClient := &Client{
		Send: make(chan []byte, 1),
	}
	normalClient := NewClient(nil)

	hub.Register(slowClient)
	hub.Register(normalClient)
	waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 })

	// Fill slowClient buffer to capacity
	slowClient.Send <- []byte("already full")

	// Broadcast should evict slowClient because its send buffer is full
	hub.Broadcast([]byte("broadcast payload"))

	// Normal client receives broadcast
	select {
	case msg := <-normalClient.Send:
		if string(msg) != "broadcast payload" {
			t.Fatalf("normalClient expected broadcast payload, got %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("normalClient timed out")
	}

	// slowClient should be evicted from hub
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 1 }) {
		t.Fatalf("expected slow client to be evicted, client count: %d", hub.ClientCount())
	}
	if hub.HasClient(slowClient) {
		t.Fatal("expected slowClient to be removed from hub")
	}
}

func TestHub_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	hub := NewHub("", "", nil)
	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(done)
	}()

	client1 := NewClient(nil)
	client2 := NewClient(nil)

	hub.Register(client1)
	hub.Register(client2)
	waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 })

	// Cancel hub context
	cancel()

	select {
	case <-done:
		// Hub exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("hub did not exit cleanly on context cancellation")
	}

	if count := hub.ClientCount(); count != 0 {
		t.Fatalf("expected 0 clients after shutdown, got %d", count)
	}

	// Verify both client channels are closed
	if _, ok := <-client1.Send; ok {
		t.Fatal("expected client1.Send to be closed on shutdown")
	}
	if _, ok := <-client2.Send; ok {
		t.Fatal("expected client2.Send to be closed on shutdown")
	}
}

func TestHub_ListenNotify_ContextCancellation(t *testing.T) {
	// Tests that listenNotify respects context cancellation with an invalid/unreachable DSN
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	hub := NewHub("postgres://invalid_user:invalid_pass@127.0.0.1:54329/invalid_db", "test_channel", nil)
	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Successfully exited on context timeout
	case <-time.After(2 * time.Second):
		t.Fatal("hub.Run did not terminate on context cancellation during listenNotify")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/websocket/... -v
```

Expected: FAIL with `undefined: NewHub`.

- [ ] **Step 3: Implement `hub.go`**

Create `backend/internal/websocket/hub.go`:

```go
package websocket

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"nhooyr.io/websocket"
)

const (
	// defaultChannelBufferSize is the buffer size for hub management channels.
	defaultChannelBufferSize = 256

	// reconnectBackoff is the wait time between LISTEN connection retry attempts.
	reconnectBackoff = 1 * time.Second
)

// Hub maintains the set of active WebSocket clients and broadcasts PostgreSQL
// NOTIFY events to them.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client

	dsn     string
	channel string
	logger  *slog.Logger

	mu sync.RWMutex
}

// NewHub creates a new Hub instance.
// If dsn is non-empty, Run(ctx) will start a dedicated goroutine listening for PostgreSQL NOTIFY events.
// If logger is nil, slog.Default() is used.
func NewHub(dsn string, channel string, logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.Default()
	}

	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, defaultChannelBufferSize),
		register:   make(chan *Client, defaultChannelBufferSize),
		unregister: make(chan *Client, defaultChannelBufferSize),
		dsn:        dsn,
		channel:    channel,
		logger:     logger,
	}
}

// Run starts the hub's main event loop and PostgreSQL notification listener.
// It blocks until the provided context is cancelled, at which point it gracefully
// shuts down and disconnects all clients.
func (h *Hub) Run(ctx context.Context) {
	if h.dsn != "" && h.channel != "" {
		go h.listenNotify(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					h.logger.Warn("client send channel full, evicting slow client")
					delete(h.clients, client)
					close(client.Send)
					if client.Conn != nil {
						_ = client.Conn.Close(websocket.StatusPolicyViolation, "send buffer full")
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// Register registers a new client for receiving broadcasts.
func (h *Hub) Register(client *Client) {
	client.hub = h
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	select {
	case h.unregister <- client:
	default:
		// Fallback for non-blocking unregister if channel is saturated or hub shutting down
		h.mu.Lock()
		if _, ok := h.clients[client]; ok {
			delete(h.clients, client)
			close(client.Send)
		}
		h.mu.Unlock()
	}
}

// Broadcast distributes a payload to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn("hub broadcast channel full, dropping message")
	}
}

// ClientCount returns the number of active clients registered with the hub.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// HasClient checks whether a client is currently registered with the hub.
func (h *Hub) HasClient(client *Client) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[client]
}

// shutdown disconnects all clients and closes their send channels.
func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		delete(h.clients, client)
		close(client.Send)
		if client.Conn != nil {
			_ = client.Conn.Close(websocket.StatusNormalClosure, "hub shutdown")
		}
	}
	h.logger.Info("websocket hub shut down cleanly")
}

// listenNotify runs on a dedicated goroutine, holding an open dedicated pgx connection
// (outside the connection pool) to execute LISTEN and receive NOTIFY events.
func (h *Hub) listenNotify(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := pgx.Connect(ctx, h.dsn)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			h.logger.Error("failed to connect dedicated pgx connection for LISTEN", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(reconnectBackoff):
				continue
			}
		}

		listenErr := h.listenLoop(ctx, conn)
		_ = conn.Close(context.Background())

		if ctx.Err() != nil {
			return
		}

		if listenErr != nil {
			h.logger.Warn("LISTEN connection lost, reconnecting...", "error", listenErr)
			select {
			case <-ctx.Done():
				return
			case <-time.After(reconnectBackoff):
			}
		}
	}
}

// listenLoop executes the LISTEN statement and waits for notifications.
func (h *Hub) listenLoop(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", pgx.Identifier{h.channel}.Sanitize()))
	if err != nil {
		return fmt.Errorf("executing LISTEN statement: %w", err)
	}

	h.logger.Info("listening for PostgreSQL notifications", "channel", h.channel)

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("waiting for notification: %w", err)
		}

		h.Broadcast([]byte(notification.Payload))
	}
}
```

- [ ] **Step 4: Remove `.gitkeep` placeholder**

```bash
rm -f backend/internal/websocket/.gitkeep
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd backend && go test -race ./internal/websocket/... -v
```

Expected: All unit tests pass with zero race conditions.

- [ ] **Step 6: Update Task 3 status to DONE**

Update `docs/plans/task.md`:

```markdown
| Task 3: Hub Implementation (`hub.go` & `hub_test.go`) | DONE | Implement `Hub` with registration, broadcast fan-out, pg LISTEN, and unit tests |
```

- [ ] **Step 7: Commit**

```bash
rm -f backend/internal/websocket/.gitkeep
git add backend/internal/websocket/hub.go backend/internal/websocket/hub_test.go docs/plans/task.md
git commit -m "feat(websocket): implement Hub struct with fan-out and pg LISTEN/NOTIFY"
```

---

### Task 4: End-to-End Suite Verification, Formatting & Linting

**Files:**
- Modify: `backend/internal/websocket/client.go` (if formatting changes needed)
- Modify: `backend/internal/websocket/hub.go` (if formatting changes needed)

**Interfaces:**
- Consumes: Complete `backend/internal/websocket` package.
- Produces: Clean `go test -race`, `go vet`, `golangci-lint`, and `go build`.

- [ ] **Step 1: Run WebSocket package tests with race detector**

```bash
cd backend && go test -race ./internal/websocket/... -v
```

Expected: PASS with 0 race warnings.

- [ ] **Step 2: Run all backend tests to ensure zero regressions**

```bash
cd backend && go test -race ./...
```

Expected: All existing tests in `apperror`, `auth`, `config`, `handler`, `middleware`, `repository`, `service`, `websocket` pass.

- [ ] **Step 3: Run `go vet` and `go build`**

```bash
cd backend && go vet ./... && go build ./...
```

Expected: No errors reported, all packages compile cleanly.

- [ ] **Step 4: Run Go formatting and linting script**

```bash
./scripts/linting.bash --go
```

Expected: `gofumpt` and `golangci-lint` pass with zero issues.

- [ ] **Step 5: Update Task 4 status to DONE**

Update `docs/plans/task.md`:

```markdown
| Task 4: End-to-End Suite Verification, Formatting & Linting | DONE | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
```

- [ ] **Step 6: Commit (if any formatting changes made)**

```bash
git add -A
git commit -m "chore(websocket): format and verify linting compliance" || true
```

---

### Task 5: Mark Tasks as Completed in Task Spec and Live Tracker

**Files:**
- Modify: `docs/go_refactor/tasks/026-websocket_hub.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified, tested, and linted WebSocket implementation.
- Produces: Marked completion status across all tasks and acceptance criteria.

- [ ] **Step 1: Update `docs/go_refactor/tasks/026-websocket_hub.md`**

Mark all acceptance criteria checkboxes as checked `[x]`:
- `[x]` `backend/internal/websocket/hub.go` implements a `Hub` struct...
- `[x]` `backend/internal/websocket/client.go` implements a `Client` struct...
- `[x]` The hub uses a dedicated `pgx` connection...
- `[x]` Hub gracefully shuts down when context is cancelled...
- `[x]` Unit tests verify registering, unregistering, broadcasting, and skipping unregistered clients...
- `[x]` `go test ./internal/websocket/...` passes.
- `[x]` `go vet ./...` reports no issues.

Mark all steps checkboxes as checked `[x]`:
- `[x]` Step 1: Add WebSocket dependency
- `[x]` Step 2: Write failing tests for the hub
- `[x]` Step 3: Run tests to verify they fail
- `[x]` Step 4: Implement `client.go`
- `[x]` Step 5: Implement `hub.go`
- `[x]` Step 6: Run tests to verify they pass
- `[x]` Step 7: Run go vet and build
- `[x]` Step 8: Commit

- [ ] **Step 2: Update `docs/plans/task.md` to reflect all tasks DONE**

Update `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup, Live Tracker & Dependency Management | DONE | Switch to branch `refactor/026-websocket-hub`, initialize tracker, and add `nhooyr.io/websocket` |
| Task 2: Client Implementation (`client.go` & `client_test.go`) | DONE | Implement `Client`, `NewClient`, `WritePump`, and `ReadPump` with unit tests |
| Task 3: Hub Implementation (`hub.go` & `hub_test.go`) | DONE | Implement `Hub` with registration, broadcast fan-out, pg LISTEN, and unit tests |
| Task 4: End-to-End Suite Verification, Formatting & Linting | DONE | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 5: Mark Tasks as Completed in Task Spec and Live Tracker | DONE | Mark `026-websocket_hub.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Commit and finalize**

```bash
git add docs/go_refactor/tasks/026-websocket_hub.md docs/plans/task.md
git commit -m "docs: mark task 026 websocket hub as completed"
```
