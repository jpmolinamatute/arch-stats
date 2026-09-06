# Task 027: Build HTTP Handler — Live Stats + WebSocket Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `LiveStatsHandler` in `backend/internal/handler/live_stats.go` providing an authenticated REST endpoint (`GET /api/v0/stats/{slot_id}`) for retrieving live slot statistics and a WebSocket upgrade endpoint (`GET /api/v0/stats/ws/{slot_id}`) for streaming real-time notifications with full unit tests, linting, and task tracking.

**Architecture:**
- `backend/internal/handler/live_stats.go`: Defines `LiveStatsHandler`, constructor `NewLiveStatsHandler(svc LiveStatsService, hub WebSocketHub)`, and route registration `Routes(r chi.Router)`.
  - `GetStats(w, r)`: Validates authenticated archer from context via `middleware.GetArcherID(r.Context())`, extracts and validates `slot_id` UUID, queries `LiveStatsService.GetStats(ctx, slotID, archerID)`, and writes JSON response (or maps `apperror.ErrNotFound` to 404).
  - `WebSocketStats(w, r)`: Validates `slot_id` UUID, upgrades HTTP connection using `github.com/coder/websocket`, instantiates `websocket.NewClient(conn)`, registers client with `WebSocketHub`, starts `WritePump` and `ReadPump`, and unregisters client on disconnect.
- `backend/internal/handler/live_stats_test.go`: Comprehensive unit tests verifying GetStats (200 with stats, 404 not found, 422 invalid UUID, 401 unauthenticated, 500 error), WebSocketStats (422 invalid UUID, full connection upgrade, client registration and unregistration lifecycle), and Chi route mounting.
- Task tracking & completion: Mark all checklist items in `docs/go_refactor/tasks/027-handler_live_stats.md` and `docs/plans/task.md` as completed.

**Tech Stack:** Go 1.24+, `github.com/go-chi/chi/v5`, `github.com/coder/websocket`, `github.com/google/uuid`, `net/http/httptest`.

**Spec:**
- [docs/go_refactor/tasks/027-handler_live_stats.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/027-handler_live_stats.md)
- [backend-old/src/routers/v0/live_stats_router.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/routers/v0/live_stats_router.py)
- [backend-old/src/core/live_stats_manager.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/core/live_stats_manager.py)
- [backend/internal/model/live_stats.go](file:///home/juanpa/Projects/arch-stats/backend/internal/model/live_stats.go)
- [backend/internal/websocket/hub.go](file:///home/juanpa/Projects/arch-stats/backend/internal/websocket/hub.go)
- [backend/internal/websocket/client.go](file:///home/juanpa/Projects/arch-stats/backend/internal/websocket/client.go)

## Global Constraints

- Git branch: `refactor/027-handler-live-stats`
- Error handling: Use `apperror` sentinel errors (`apperror.ErrNotFound`, `apperror.ErrValidation`, `apperror.ErrUnauthorized`) and serialize via `writeAppError(w, err)` and `writeJSON(w, status, data)`.
- WebSocket library: Use `github.com/coder/websocket` (as established in Task 026 and `backend/go.mod`).
- Concurrency & race safety: Strictly zero data races (`go test -race ./...`).
- Single-flow execution: Exactly one active task tracked in `docs/plans/task.md`.
- Coding standards: Follow Effective Go conventions, `gofumpt` formatting, and zero `golangci-lint` warnings.
- Task completion: Mark all acceptance criteria and steps in `docs/go_refactor/tasks/027-handler_live_stats.md` as done (`[x]`) and update `docs/plans/task.md` with all tasks `DONE`.

---

## File Structure

```
backend/
└── internal/
    └── handler/
        ├── live_stats.go                 # [NEW] LiveStatsHandler, LiveStatsService interface, WebSocketHub interface, Routes, GetStats, WebSocketStats
        └── live_stats_test.go            # [NEW] Comprehensive unit test suite with mock service, mock hub, and httptest
docs/
├── plans/
│   ├── task.md                           # [MODIFY] Track Task 027 live checklist progress (table-only)
│   └── 2026-09-05-handler-live-stats.md  # [NEW] This implementation plan document
└── go_refactor/
    └── tasks/
        └── 027-handler_live_stats.md     # [MODIFY] Mark all acceptance criteria and steps as completed
```

---

### Task 1: Git Branch Setup & Live Tracker Initialization

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Clean working tree on `main`.
- Produces: Checked out branch `refactor/027-handler-live-stats` and initialized `docs/plans/task.md`.

- [ ] **Step 1: Create and switch to git branch**

```bash
git checkout -b refactor/027-handler-live-stats
```

- [ ] **Step 2: Initialize `docs/plans/task.md` with Task 027 checklist table**

Update `docs/plans/task.md` to:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | IN_PROGRESS | Switch to branch `refactor/027-handler-live-stats` and initialize live progress tracker |
| Task 2: Live Stats Unit Tests Suite (`live_stats_test.go`) | PENDING | Write failing unit tests for GetStats, WebSocket upgrade, error mapping, and routes |
| Task 3: Live Stats Handler Implementation (`live_stats.go`) | PENDING | Implement `LiveStatsHandler`, interfaces, `GetStats`, `WebSocketStats`, and `Routes` |
| Task 4: Test Suite Verification, Race Detection, and Linting | PENDING | Run full test suite with `-race`, `go vet`, `golangci-lint`, and `go build` |
| Task 5: Mark Tasks as Completed in Task Spec and Live Tracker | PENDING | Mark `027-handler_live_stats.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Commit initial tracker setup**

```bash
git add docs/plans/task.md docs/plans/2026-09-05-handler-live-stats.md
git commit -m "docs: initialize task tracker for task 027 handler live stats"
```

---

### Task 2: Live Stats Unit Tests Suite (`live_stats_test.go`)

**Files:**
- Create: `backend/internal/handler/live_stats_test.go`

**Interfaces:**
- Consumes: `model.LiveStat`, `model.Stats`, `model.ShotScore`, `apperror`, `middleware.WithArcherID`, `coder/websocket`.
- Produces: Failing test suite covering all REST and WebSocket handler behaviors.

- [ ] **Step 1: Write unit tests in `backend/internal/handler/live_stats_test.go`**

```go
package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	ws "github.com/jpmolinamatute/arch-stats/backend/internal/websocket"
)

type mockLiveStatsService struct {
	getStatsFn func(ctx context.Context, slotID, archerID uuid.UUID) (*model.LiveStat, error)
}

func (m *mockLiveStatsService) GetStats(ctx context.Context, slotID, archerID uuid.UUID) (*model.LiveStat, error) {
	if m.getStatsFn != nil {
		return m.getStatsFn(ctx, slotID, archerID)
	}
	return nil, errors.New("unimplemented")
}

type mockWebSocketHub struct {
	mu           sync.Mutex
	registered   []*ws.Client
	unregistered []*ws.Client
	onRegister   func(client *ws.Client)
	onUnregister func(client *ws.Client)
}

func (m *mockWebSocketHub) Register(client *ws.Client) {
	m.mu.Lock()
	m.registered = append(m.registered, client)
	m.mu.Unlock()
	if m.onRegister != nil {
		m.onRegister(client)
	}
}

func (m *mockWebSocketHub) Unregister(client *ws.Client) {
	m.mu.Lock()
	m.unregistered = append(m.unregistered, client)
	m.mu.Unlock()
	if m.onUnregister != nil {
		m.onUnregister(client)
	}
}

func newLiveStatsTestRequest(method, url string, body io.Reader, authArcherID *uuid.UUID, paramKey, paramVal string) *http.Request {
	req := httptest.NewRequest(method, url, body)
	if paramKey != "" {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add(paramKey, paramVal)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}
	if authArcherID != nil {
		req = req.WithContext(middleware.WithArcherID(req.Context(), *authArcherID))
	}
	return req
}

func sampleLiveStat(slotID uuid.UUID) *model.LiveStat {
	now := time.Now().UTC().Truncate(time.Second)
	return &model.LiveStat{
		Scores: []model.ShotScore{
			{
				ShotID:    uuid.New(),
				Score:     10,
				IsX:       true,
				CreatedAt: now,
			},
			{
				ShotID:    uuid.New(),
				Score:     9,
				IsX:       false,
				CreatedAt: now.Add(time.Minute),
			},
		},
		Stats: model.Stats{
			SlotID:        slotID,
			NumberOfShots: 2,
			TotalScore:    19,
			MaxScore:      10,
			Mean:          9.5,
		},
	}
}

func TestLiveStatsHandler_GetStats_Success(t *testing.T) {
	slotID := uuid.New()
	archerID := uuid.New()
	expected := sampleLiveStat(slotID)

	svc := &mockLiveStatsService{
		getStatsFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.LiveStat, error) {
			if sID != slotID {
				t.Errorf("expected slotID %v, got %v", slotID, sID)
			}
			if aID != archerID {
				t.Errorf("expected archerID %v, got %v", archerID, aID)
			}
			return expected, nil
		},
	}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/"+slotID.String(), nil, &archerID, "slot_id", slotID.String())
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var got model.LiveStat
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}

	if got.Stats.SlotID != slotID {
		t.Errorf("expected slot ID %v, got %v", slotID, got.Stats.SlotID)
	}
	if got.Stats.NumberOfShots != 2 || got.Stats.TotalScore != 19 || got.Stats.Mean != 9.5 {
		t.Errorf("unexpected stats values: %+v", got.Stats)
	}
	if len(got.Scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(got.Scores))
	}
	if got.Scores[0].Score != 10 || !got.Scores[0].IsX {
		t.Errorf("unexpected first score: %+v", got.Scores[0])
	}
}

func TestLiveStatsHandler_GetStats_NotFound(t *testing.T) {
	slotID := uuid.New()
	archerID := uuid.New()

	svc := &mockLiveStatsService{
		getStatsFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.LiveStat, error) {
			return nil, apperror.ErrNotFound
		},
	}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/"+slotID.String(), nil, &archerID, "slot_id", slotID.String())
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_GetStats_InvalidSlotID(t *testing.T) {
	archerID := uuid.New()
	svc := &mockLiveStatsService{}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/not-a-uuid", nil, &archerID, "slot_id", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422 Unprocessable Entity, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_GetStats_Unauthenticated(t *testing.T) {
	slotID := uuid.New()
	svc := &mockLiveStatsService{}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	// No archer ID in context
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/"+slotID.String(), nil, nil, "slot_id", slotID.String())
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_GetStats_InternalError(t *testing.T) {
	slotID := uuid.New()
	archerID := uuid.New()

	svc := &mockLiveStatsService{
		getStatsFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.LiveStat, error) {
			return nil, errors.New("db query failed")
		},
	}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/"+slotID.String(), nil, &archerID, "slot_id", slotID.String())
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 Internal Server Error, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_WebSocketStats_InvalidSlotID(t *testing.T) {
	svc := &mockLiveStatsService{}
	hub := &mockWebSocketHub{}

	h := handler.NewLiveStatsHandler(svc, hub)
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/ws/not-a-uuid", nil, nil, "slot_id", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.WebSocketStats(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422 Unprocessable Entity, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_WebSocketStats_UpgradeAndLifecycle(t *testing.T) {
	svc := &mockLiveStatsService{}

	regChan := make(chan *ws.Client, 1)
	unregChan := make(chan *ws.Client, 1)

	hub := &mockWebSocketHub{
		onRegister: func(client *ws.Client) {
			regChan <- client
		},
		onUnregister: func(client *ws.Client) {
			unregChan <- client
		},
	}

	h := handler.NewLiveStatsHandler(svc, hub)

	r := chi.NewRouter()
	r.Route("/api/v0/stats", h.Routes)

	server := httptest.NewServer(r)
	defer server.Close()

	slotID := uuid.New()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v0/stats/ws/" + slotID.String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := coderws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	// Verify client is registered
	select {
	case client := <-regChan:
		if client == nil {
			t.Fatal("expected non-nil registered client")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client registration")
	}

	// Close client connection to trigger unregister
	_ = conn.Close(coderws.StatusNormalClosure, "test closing")

	select {
	case client := <-unregChan:
		if client == nil {
			t.Fatal("expected non-nil unregistered client")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client unregistration")
	}
}

func TestLiveStatsHandler_Routes_Registration(t *testing.T) {
	svc := &mockLiveStatsService{}
	hub := &mockWebSocketHub{}
	h := handler.NewLiveStatsHandler(svc, hub)

	r := chi.NewRouter()
	h.Routes(r)

	routes := r.Routes()
	if len(routes) == 0 {
		t.Fatal("expected routes to be registered")
	}

	hasStatsRoute := false
	hasWSRoute := false

	for _, route := range routes {
		if strings.Contains(route.Pattern, "{slot_id}") {
			hasStatsRoute = true
		}
		if strings.Contains(route.Pattern, "ws") {
			hasWSRoute = true
		}
	}

	if !hasStatsRoute {
		t.Error("expected {slot_id} route pattern to be registered")
	}
	if !hasWSRoute {
		t.Error("expected ws/{slot_id} route pattern to be registered")
	}
}
```

- [ ] **Step 2: Run tests to verify compilation failure**

```bash
cd backend && go test ./internal/handler/... -v -run TestLiveStatsHandler
```
Expected: Compilation failure because `handler.NewLiveStatsHandler` and `handler.LiveStatsHandler` are not yet defined.

---

### Task 3: Live Stats Handler Implementation (`live_stats.go`)

**Files:**
- Create: `backend/internal/handler/live_stats.go`

**Interfaces:**
- Consumes: `model.LiveStat`, `apperror`, `middleware.GetArcherID`, `middleware.WriteError`, `coder/websocket`, `websocket.Client`.
- Produces: `LiveStatsHandler`, `LiveStatsService`, `WebSocketHub`, `Routes(r chi.Router)`, `GetStats(w, r)`, `WebSocketStats(w, r)`.

- [ ] **Step 1: Implement `backend/internal/handler/live_stats.go`**

```go
package handler

import (
	"context"
	"net/http"

	coderws "github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	ws "github.com/jpmolinamatute/arch-stats/backend/internal/websocket"
)

// LiveStatsService defines the business operations required by LiveStatsHandler.
type LiveStatsService interface {
	GetStats(ctx context.Context, slotID, archerID uuid.UUID) (*model.LiveStat, error)
}

// WebSocketHub defines the hub operations required by LiveStatsHandler.
type WebSocketHub interface {
	Register(client *ws.Client)
	Unregister(client *ws.Client)
}

// LiveStatsHandler manages HTTP and WebSocket endpoints for real-time and aggregate slot statistics.
type LiveStatsHandler struct {
	svc LiveStatsService
	hub WebSocketHub
}

// NewLiveStatsHandler constructs a LiveStatsHandler with service and hub dependency injection.
func NewLiveStatsHandler(svc LiveStatsService, hub WebSocketHub) *LiveStatsHandler {
	return &LiveStatsHandler{
		svc: svc,
		hub: hub,
	}
}

// Routes registers live stats endpoints on the provided chi Router.
// The literal `/ws/{slot_id}` route is registered before `/{slot_id}` to ensure unambiguous route matching.
func (h *LiveStatsHandler) Routes(r chi.Router) {
	r.Get("/ws/{slot_id}", h.WebSocketStats)
	r.Get("/{slot_id}", h.GetStats)
}

// GetStats handles GET /api/v0/stats/{slot_id}.
// Returns HTTP 200 OK with model.LiveStat JSON containing current stats and shot scores.
// Requires authentication and returns 401 if missing auth context.
func (h *LiveStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "invalid slot_id UUID"))
		return
	}

	stat, err := h.svc.GetStats(r.Context(), slotID, authArcherID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, stat); err != nil {
		writeAppError(w, err)
		return
	}
}

// WebSocketStats handles GET /api/v0/stats/ws/{slot_id}.
// Upgrades the HTTP request to a WebSocket connection, registers the client with the hub,
// starts WritePump and ReadPump goroutines, and unregisters the client when the connection closes.
func (h *LiveStatsHandler) WebSocketStats(w http.ResponseWriter, r *http.Request) {
	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	if _, err := uuid.Parse(slotIDStr); err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "invalid slot_id UUID"))
		return
	}

	conn, err := coderws.Accept(w, r, &coderws.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		// Accept already writes the HTTP error response if handshake fails
		return
	}

	client := ws.NewClient(conn)
	if h.hub != nil {
		h.hub.Register(client)
		defer h.hub.Unregister(client)
	}

	ctx := r.Context()
	go client.WritePump(ctx)
	client.ReadPump(ctx)
}
```

- [ ] **Step 2: Run unit tests to verify they pass**

```bash
cd backend && go test ./internal/handler/... -v -run TestLiveStatsHandler
```
Expected: All tests pass.

- [ ] **Step 3: Update `docs/plans/task.md` with Task 2 and Task 3 complete**

Update `docs/plans/task.md` to reflect progress.

- [ ] **Step 4: Commit handler and tests**

```bash
git add backend/internal/handler/live_stats.go backend/internal/handler/live_stats_test.go docs/plans/task.md
git commit -m "feat(handler): implement LiveStatsHandler with REST stats and WebSocket streaming"
```

---

### Task 4: Test Suite Verification, Race Detection, and Linting

**Files:**
- Test: All backend tests

**Interfaces:**
- Consumes: Complete backend package suite.
- Produces: Clean test outputs, zero race conditions, zero linting violations, clean binary compilation.

- [ ] **Step 1: Run handler package tests with race detector**

```bash
cd backend && go test -race ./internal/handler/... -v
```
Expected: All handler tests pass with 0 data races.

- [ ] **Step 2: Run entire backend test suite with race detector**

```bash
cd backend && go test -race ./...
```
Expected: All packages pass with cached or fresh test runs.

- [ ] **Step 3: Run `go vet`**

```bash
cd backend && go vet ./...
```
Expected: Exit code 0, no warnings.

- [ ] **Step 4: Run `golangci-lint`**

```bash
cd backend && golangci-lint run ./...
```
Expected: 0 issues reported.

- [ ] **Step 5: Verify binary compiles**

```bash
cd backend && go build ./...
```
Expected: Successful compilation of all backend packages and cmd entrypoint.

- [ ] **Step 6: Update `docs/plans/task.md` with Task 4 complete**

Update `docs/plans/task.md` to reflect Task 4 DONE.

---

### Task 5: Mark Tasks as Completed in Task Spec and Live Tracker

**Files:**
- Modify: `docs/go_refactor/tasks/027-handler_live_stats.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified implementation from Task 4.
- Produces: Updated markdown documentation with all checkboxes checked and all tasks marked DONE.

- [ ] **Step 1: Mark all acceptance criteria and steps in `docs/go_refactor/tasks/027-handler_live_stats.md` as completed (`[x]`)**

Update `docs/go_refactor/tasks/027-handler_live_stats.md`:

```markdown
## Acceptance Criteria

- [x] `backend/internal/handler/live_stats.go` implements `LiveStatsHandler` with methods:
    - `GetStats(w, r)` — GET `/api/v0/stats/{slot_id}` — returns live statistics for a slot
    - `WebSocketStats(w, r)` — GET `/api/v0/stats/ws/{slot_id}` — upgrades to WebSocket,
    registers client with hub, streams NOTIFY payloads as JSON messages
- [x] The WebSocket handler:
    - Upgrades the HTTP connection using `nhooyr.io/websocket`
    - Creates a `Client` and registers it with the hub
    - Starts `WritePump` and `ReadPump` goroutines
    - Unregisters the client when the connection closes
- [x] The stats endpoint is authenticated (requires auth middleware).
- [x] Unit tests verify:
    - GetStats returns 200 + stats JSON for valid slot
    - GetStats returns 404 for non-existent slot
- [x] `go test ./internal/handler/...` passes.
- [x] `go vet ./...` reports no issues.
```

And in `Steps`:
```markdown
- [x] **Step 1: Write failing tests for GetStats**
- [x] **Step 2: Run tests to verify they fail**
- [x] **Step 3: Implement `live_stats.go`**
- [x] **Step 4: Run tests to verify they pass**
- [x] **Step 5: Run go vet and build**
- [x] **Step 6: Commit**
```

- [ ] **Step 2: Update `docs/plans/task.md` with all tasks marked `DONE`**

Update `docs/plans/task.md` to:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | DONE | Switch to branch `refactor/027-handler-live-stats` and initialize live progress tracker |
| Task 2: Live Stats Unit Tests Suite (`live_stats_test.go`) | DONE | Write failing unit tests for GetStats, WebSocket upgrade, error mapping, and routes |
| Task 3: Live Stats Handler Implementation (`live_stats.go`) | DONE | Implement `LiveStatsHandler`, interfaces, `GetStats`, `WebSocketStats`, and `Routes` |
| Task 4: Test Suite Verification, Race Detection, and Linting | DONE | Run full test suite with `-race`, `go vet`, `golangci-lint`, and `go build` |
| Task 5: Mark Tasks as Completed in Task Spec and Live Tracker | DONE | Mark `027-handler_live_stats.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Commit final documentation and task completion changes**

```bash
git add docs/go_refactor/tasks/027-handler_live_stats.md docs/plans/task.md
git commit -m "docs: mark task 027 live stats handler as completed"
```
