# Task 025: Wire chi Router with All Handlers + DI in `main.go` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete dependency injection wiring and server lifecycle in `backend/cmd/arch-stats/main.go`, build the chi router with all domain handlers and middleware stack under `/api/v0/`, implement the public `/api/v0/health` endpoint returning schema version, and verify with comprehensive automated tests. Upon completion, mark all tasks in `docs/go_refactor/tasks/025-chi_router_and_main_wiring.md` and `docs/plans/task.md` as done.

**Architecture:**
- `backend/internal/handler/health.go`: Implements `HealthHandler` exposing `GET /api/v0/health`, querying database schema version via `MaintenanceService` (`GetSchemaVersion(ctx context.Context) (int64, error)`). Public endpoint returning 200 OK with `{"status": "ok", "schema_version": <version>}` or 503 on error.
- `backend/cmd/arch-stats/router.go`: Implements `buildRouter(deps RouterDeps) chi.Router` configuring middleware stack and route groups matching the Python API structure under `/api/v0`:
  - Global middleware: `middleware.RequestLogger(deps.Logger)`, `middleware.Recovery`, `middleware.CORS(deps.Cfg.DevMode)`.
  - Public routes:
    - `GET /health` -> `deps.HealthHandler.Health`
    - `/faces` -> `deps.FaceHandler.Routes` (`GET /`, `GET /{face_type}`)
    - `/auth`: `POST /login`, `POST /google`, `POST /register`
  - Protected routes (wrapped with `middleware.Auth(deps.AuthSvc)` and `middleware.ErrorMapper`):
    - Subgroup under `/auth`: `POST /logout`, `GET /me`
    - `/archer` -> `deps.ArcherHandler.Routes`
    - `/session`: `/slot` -> `deps.SlotHandler.Routes`, session lifecycle -> `deps.SessionHandler.Routes`
    - `/shot` -> `deps.ShotHandler.Routes`
- `backend/cmd/arch-stats/main.go`: Refactors `run()` into structured dependency injection:
  1. Config loading & validation (`config.Load()`, `cfg.Validate()`).
  2. Logger initialization (`config.NewLogger(cfg.DevMode)`, `slog.SetDefault(logger)`).
  3. Database pool connection (`repository.NewPool(...)`).
  4. Migrations CLI subcommand (`migrate`) and startup auto-migrations (`cfg.ApplyMigrationsOnStart`).
  5. Repositories: `archerRepo`, `authSessionRepo`, `sessionRepo`, `slotRepo`, `shotRepo`, `faceRepo`, `targetRepo`, `maintenanceRepo`, `reportingRepo`.
  6. Schema version logging via `maintenanceRepo.GetSchemaVersion(ctx)`.
  7. Services: `archerSvc`, `sessionSvc`, `slotSvc`, `shotSvc`, `faceSvc`, `targetSvc`.
  8. Auth service: `auth.NewService(...)`.
  9. Handlers: `authHandler`, `archerHandler`, `sessionHandler`, `slotHandler`, `shotHandler`, `faceHandler`, `healthHandler`.
  10. Router: `buildRouter(deps)`.
  11. HTTP server with timeouts and graceful shutdown via `signal.NotifyContext` and `srv.Shutdown(shutdownCtx)`.
- Acceptance & Task Tracking: Update `docs/go_refactor/tasks/025-chi_router_and_main_wiring.md` and `docs/plans/task.md` marking all items completed.

**Tech Stack:** Go 1.24+, `github.com/go-chi/chi/v5`, `net/http`, `net/http/httptest`, `log/slog`, `github.com/jackc/pgx/v5/pgxpool`, internal packages (`config`, `repository`, `service`, `auth`, `handler`, `middleware`, `model`, `apperror`).

**Spec:**
- [docs/go_refactor/tasks/025-chi_router_and_main_wiring.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/025-chi_router_and_main_wiring.md)
- [backend-old/app.py](file:///home/juanpa/Projects/arch-stats/backend-old/app.py)
- [backend-old/src/routers/v0/__init__.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/routers/v0/__init__.py)
- [backend/cmd/arch-stats/main.go](file:///home/juanpa/Projects/arch-stats/backend/cmd/arch-stats/main.go)
- [backend/internal/middleware/](file:///home/juanpa/Projects/arch-stats/backend/internal/middleware)

## Global Constraints

- Git branch: `refactor/025-chi-router-and-main-wiring`
- Route structure matching Python API under `/api/v0/`:
  - `/api/v0/health` — public health check returning database schema version
  - `/api/v0/faces/` — public face catalog
  - `/api/v0/auth/` — public login/register, protected logout/me
  - `/api/v0/archer/` — protected archer management
  - `/api/v0/session/` — protected session lifecycle
  - `/api/v0/session/slot/` — protected slot management
  - `/api/v0/shot/` — protected shot management
  - (Note: `/api/v0/stats/` is scheduled for Task 027 after WebSocket Hub in Task 026)
- Middleware ordering: logging -> recovery -> CORS -> (per-group auth) -> error mapper
- Graceful shutdown: `http.Server` handles `SIGINT`/`SIGTERM`, stops accepting new connections, drains existing requests with timeout, closes database pool on exit
- Clean compilation: `go vet ./...` reports no issues, `go build ./cmd/arch-stats` compiles cleanly
- Linting: Passes `./scripts/linting.bash --go` (`gofumpt` and `golangci-lint run ./...`)
- Task completion: Mark all acceptance criteria and steps in `docs/go_refactor/tasks/025-chi_router_and_main_wiring.md` and `docs/plans/task.md` as done (`[x]` / `DONE`)

---

## File Structure

```
backend/
├── cmd/
│   └── arch-stats/
│       ├── main.go               # [MODIFY] Full DI wiring, migrations, graceful HTTP server lifecycle
│       ├── router.go             # [NEW] RouterDeps struct, buildRouter() configuring chi route groups & middleware
│       └── router_test.go        # [NEW] Test suite verifying route registration, public/protected boundaries, CORS, recovery
├── internal/
│   └── handler/
│       ├── health.go             # [NEW] HealthHandler, MaintenanceService interface, HealthResponse
│       └── health_test.go        # [NEW] Unit tests for HealthHandler (200 OK with schema version, 503 on failure)
docs/
├── plans/
│   ├── task.md                   # [MODIFY] Track Task 025 live checklist progress (table-only)
│   └── 2026-09-05-chi-router-and-main-wiring.md # [NEW] Implementation plan document
└── go_refactor/
    └── tasks/
        └── 025-chi_router_and_main_wiring.md    # [MODIFY] Mark all acceptance criteria and steps as completed
```

---

### Task 1: Git Branch Setup & Live Tracker Initialization

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Clean working tree on `main`.
- Produces: Checked out branch `refactor/025-chi-router-and-main-wiring` and initialized `docs/plans/task.md`.

- [ ] **Step 1: Create and switch to git branch**

```bash
git checkout -b refactor/025-chi-router-and-main-wiring
```

- [ ] **Step 2: Initialize `docs/plans/task.md` with Task 025 checklist table**

Update `docs/plans/task.md` to:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | IN_PROGRESS | Switch to branch `refactor/025-chi-router-and-main-wiring` and initialize tracker |
| Task 2: Health Handler Implementation | PENDING | Implement `HealthHandler` in `internal/handler/health.go` and unit tests in `health_test.go` |
| Task 3: Chi Router Construction, Middleware Stack & Routing Tests | PENDING | Implement `buildRouter` in `cmd/arch-stats/router.go` and comprehensive route tests in `router_test.go` |
| Task 4: Main Application Wiring & Graceful Server Lifecycle | PENDING | Refactor `cmd/arch-stats/main.go` with full dependency injection and graceful shutdown |
| Task 5: End-to-End Suite Verification, Formatting & Linting | PENDING | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 6: Mark Tasks as Completed in Task Spec and Live Tracker | PENDING | Mark `025-chi_router_and_main_wiring.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Update Task 1 status to DONE**

Update `docs/plans/task.md`:

```markdown
| Task 1: Git Branch Setup & Live Tracker Initialization | DONE | Switch to branch `refactor/025-chi-router-and-main-wiring` and initialize tracker |
```

---

### Task 2: Health Handler Implementation

**Files:**
- Create: `backend/internal/handler/health.go`
- Create: `backend/internal/handler/health_test.go`

**Interfaces:**
- Consumes: `repository.MaintenanceRepo.GetSchemaVersion(ctx context.Context) (int64, error)`.
- Produces: `HealthHandler` with `Health(w http.ResponseWriter, r *http.Request)`.

- [ ] **Step 1: Write failing tests for HealthHandler**

Create `backend/internal/handler/health_test.go`:

```go
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockMaintenanceService struct {
	getSchemaVersionFn func(ctx context.Context) (int64, error)
}

func (m *mockMaintenanceService) GetSchemaVersion(ctx context.Context) (int64, error) {
	if m.getSchemaVersionFn != nil {
		return m.getSchemaVersionFn(ctx)
	}
	return 0, nil
}

func TestHealthHandler_Health_Success(t *testing.T) {
	mockSvc := &mockMaintenanceService{
		getSchemaVersionFn: func(ctx context.Context) (int64, error) {
			return 42, nil
		},
	}
	h := NewHealthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.SchemaVersion != 42 {
		t.Errorf("expected schema_version 42, got %d", resp.SchemaVersion)
	}
}

func TestHealthHandler_Health_DatabaseError(t *testing.T) {
	mockSvc := &mockMaintenanceService{
		getSchemaVersionFn: func(ctx context.Context) (int64, error) {
			return 0, errors.New("database connection refused")
		},
	}
	h := NewHealthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "unhealthy" {
		t.Errorf("expected status 'unhealthy', got %q", resp.Status)
	}
	if resp.Error != "database connection refused" {
		t.Errorf("expected error message 'database connection refused', got %q", resp.Error)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/handler -run TestHealthHandler -v
```

Expected: compilation failure (`NewHealthHandler` undefined).

- [ ] **Step 3: Write `HealthHandler` implementation in `backend/internal/handler/health.go`**

Create `backend/internal/handler/health.go`:

```go
package handler

import (
	"context"
	"net/http"
)

// MaintenanceService defines schema and maintenance operations required by HealthHandler.
type MaintenanceService interface {
	GetSchemaVersion(ctx context.Context) (int64, error)
}

// HealthResponse represents the payload returned by GET /api/v0/health.
type HealthResponse struct {
	Status        string `json:"status"`
	SchemaVersion int64  `json:"schema_version"`
	Error         string `json:"error,omitempty"`
}

// HealthHandler manages health check endpoints.
type HealthHandler struct {
	maintenance MaintenanceService
}

// NewHealthHandler constructs a HealthHandler with maintenance service dependency injection.
func NewHealthHandler(maintenance MaintenanceService) *HealthHandler {
	return &HealthHandler{maintenance: maintenance}
}

// Health handles GET /api/v0/health.
// It returns the current database migration schema version.
// This is an unauthenticated, public endpoint.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ver, err := h.maintenance.GetSchemaVersion(r.Context())
	if err != nil {
		_ = writeJSON(w, http.StatusServiceUnavailable, HealthResponse{
			Status: "unhealthy",
			Error:  err.Error(),
		})
		return
	}

	_ = writeJSON(w, http.StatusOK, HealthResponse{
		Status:        "ok",
		SchemaVersion: ver,
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/handler -run TestHealthHandler -v
```

Expected: PASS for `TestHealthHandler_Health_Success` and `TestHealthHandler_Health_DatabaseError`.

- [ ] **Step 5: Update `docs/plans/task.md`**

```markdown
| Task 2: Health Handler Implementation | DONE | Implement `HealthHandler` in `internal/handler/health.go` and unit tests in `health_test.go` |
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/health.go backend/internal/handler/health_test.go docs/plans/task.md
git commit -m "feat(handler): add HealthHandler with database schema version check"
```

---

### Task 3: Chi Router Construction, Middleware Stack & Routing Tests

**Files:**
- Create: `backend/cmd/arch-stats/router.go`
- Create: `backend/cmd/arch-stats/router_test.go`

**Interfaces:**
- Consumes:
  - `RouterDeps` struct containing `*config.Config`, `*slog.Logger`, `middleware.TokenAuthenticator`, and handlers (`*handler.AuthHandler`, `*handler.ArcherHandler`, `*handler.SessionHandler`, `*handler.SlotHandler`, `*handler.ShotHandler`, `*handler.FaceHandler`, `*handler.HealthHandler`).
- Produces:
  - `buildRouter(deps RouterDeps) chi.Router`

- [ ] **Step 1: Write comprehensive router tests in `backend/cmd/arch-stats/router_test.go`**

Create `backend/cmd/arch-stats/router_test.go`:

```go
package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/config"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// Mock implementations for router testing

type testAuthService struct {
	authenticateFn func(ctx context.Context, token string) (uuid.UUID, error)
	loginFn        func(ctx context.Context, cred string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error)
	registerFn     func(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error)
	revokeFn       func(ctx context.Context, token string) error
	decodeFn       func(token string) (*auth.Claims, error)
}

func (m *testAuthService) Authenticate(ctx context.Context, token string) (uuid.UUID, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, token)
	}
	if token == "valid-token" {
		return uuid.MustParse("00000000-0000-0000-0000-000000000001"), nil
	}
	return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid token")
}

func (m *testAuthService) LoginWithGoogle(ctx context.Context, cred string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, cred, now, meta...)
	}
	return &model.AuthAuthenticated{
		Status:      model.AuthStatusAuthenticated,
		AccessToken: "test-token",
		ExpiresAt:   now.Add(24 * time.Hour),
	}, nil, nil
}

func (m *testAuthService) RegisterWithGoogle(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, payload, now, meta...)
	}
	return &model.AuthAuthenticated{
		Status:      model.AuthStatusAuthenticated,
		AccessToken: "test-token",
		ExpiresAt:   now.Add(24 * time.Hour),
	}, nil
}

func (m *testAuthService) RevokeToken(ctx context.Context, token string) error {
	if m.revokeFn != nil {
		return m.revokeFn(ctx, token)
	}
	return nil
}

func (m *testAuthService) DecodeToken(token string) (*auth.Claims, error) {
	if m.decodeFn != nil {
		return m.decodeFn(token)
	}
	return &auth.Claims{Sub: "00000000-0000-0000-0000-000000000001", Exp: time.Now().Add(time.Hour).Unix()}, nil
}

type testArcherService struct {
	listFn    func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	createFn  func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	updateFn  func(ctx context.Context, id uuid.UUID, data model.ArcherSet) error
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *testArcherService) List(ctx context.Context, f model.ArcherFilter) ([]model.ArcherRead, error) {
	if m.listFn != nil {
		return m.listFn(ctx, f)
	}
	return []model.ArcherRead{}, nil
}

func (m *testArcherService) GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &model.ArcherRead{ArcherID: id, ArcherName: "Test Archer"}, nil
}

func (m *testArcherService) Create(ctx context.Context, d model.ArcherCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, d)
	}
	return uuid.New(), nil
}

func (m *testArcherService) Update(ctx context.Context, id uuid.UUID, d model.ArcherSet) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, d)
	}
	return nil
}

func (m *testArcherService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type testSessionService struct {
	getByIDFn         func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	getOpenFn         func(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	listFn            func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	createFn          func(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	closeFn           func(ctx context.Context, id uuid.UUID) error
	reOpenFn          func(ctx context.Context, id uuid.UUID) error
	getParticipatingFn func(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)
}

func (m *testSessionService) GetByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &model.SessionRead{SessionID: id}, nil
}

func (m *testSessionService) GetOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error) {
	if m.getOpenFn != nil {
		return m.getOpenFn(ctx, archerID)
	}
	return nil, nil
}

func (m *testSessionService) List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return []model.SessionRead{}, nil
}

func (m *testSessionService) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.New(), nil
}

func (m *testSessionService) Close(ctx context.Context, id uuid.UUID) error {
	if m.closeFn != nil {
		return m.closeFn(ctx, id)
	}
	return nil
}

func (m *testSessionService) ReOpen(ctx context.Context, id uuid.UUID) error {
	if m.reOpenFn != nil {
		return m.reOpenFn(ctx, id)
	}
	return nil
}

func (m *testSessionService) GetParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error) {
	if m.getParticipatingFn != nil {
		return m.getParticipatingFn(ctx, archerID)
	}
	return nil, nil
}

type testSlotService struct {
	getCurrentSlotFn func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error)
	getSlotFn        func(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error)
	joinSessionFn    func(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error)
	reJoinSessionFn  func(ctx context.Context, slotID, archerID uuid.UUID) (*model.SlotJoinResponse, error)
	leaveSessionFn   func(ctx context.Context, slotID, archerID uuid.UUID) error
}

func (m *testSlotService) GetArcherCurrentSlot(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
	if m.getCurrentSlotFn != nil {
		return m.getCurrentSlotFn(ctx, archerID)
	}
	return &model.FullSlotInfo{SlotID: uuid.New(), ArcherID: archerID}, nil
}

func (m *testSlotService) GetSlot(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error) {
	if m.getSlotFn != nil {
		return m.getSlotFn(ctx, slotID)
	}
	return &model.FullSlotInfo{SlotID: slotID}, nil
}

func (m *testSlotService) JoinSession(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
	if m.joinSessionFn != nil {
		return m.joinSessionFn(ctx, req)
	}
	return &model.SlotJoinResponse{SlotID: uuid.New()}, nil
}

func (m *testSlotService) ReJoinSession(ctx context.Context, slotID, archerID uuid.UUID) (*model.SlotJoinResponse, error) {
	if m.reJoinSessionFn != nil {
		return m.reJoinSessionFn(ctx, slotID, archerID)
	}
	return &model.SlotJoinResponse{SlotID: slotID}, nil
}

func (m *testSlotService) LeaveSession(ctx context.Context, slotID, archerID uuid.UUID) error {
	if m.leaveSessionFn != nil {
		return m.leaveSessionFn(ctx, slotID, archerID)
	}
	return nil
}

type testShotService struct {
	createFn      func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error)
	createBatchFn func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error)
	getBySlotFn   func(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error)
	countBySlotFn func(ctx context.Context, slotID, archerID uuid.UUID) (int, error)
}

func (m *testShotService) Create(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, shot, archerID)
	}
	return uuid.New(), nil
}

func (m *testShotService) CreateBatch(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
	if m.createBatchFn != nil {
		return m.createBatchFn(ctx, shots, archerID)
	}
	return []uuid.UUID{uuid.New()}, nil
}

func (m *testShotService) GetBySlot(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error) {
	if m.getBySlotFn != nil {
		return m.getBySlotFn(ctx, slotID, archerID)
	}
	return []model.ShotRead{}, nil
}

func (m *testShotService) CountBySlot(ctx context.Context, slotID, archerID uuid.UUID) (int, error) {
	if m.countBySlotFn != nil {
		return m.countBySlotFn(ctx, slotID, archerID)
	}
	return 0, nil
}

type testFaceService struct {
	listAllFn func(ctx context.Context) ([]model.FaceRead, error)
	getByIDFn func(ctx context.Context, id string) (*model.FaceRead, error)
}

func (m *testFaceService) ListAll(ctx context.Context) ([]model.FaceRead, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return []model.FaceRead{
		{FaceType: "WA_field", FaceName: "WA Field"},
	}, nil
}

func (m *testFaceService) GetByID(ctx context.Context, id string) (*model.FaceRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	if id == "unknown" {
		return nil, apperror.Wrap(apperror.ErrNotFound, "face not found")
	}
	return &model.FaceRead{FaceType: model.FaceType(id), FaceName: "Test Face"}, nil
}

type testMaintenanceService struct {
	schemaVer int64
	err       error
}

func (m *testMaintenanceService) GetSchemaVersion(ctx context.Context) (int64, error) {
	return m.schemaVer, m.err
}

func setupTestRouter() http.Handler {
	cfg := &config.Config{
		DevMode:       true,
		ServerPort:    8000,
		JWTTTLMinutes: 1440,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	authSvc := &testAuthService{}
	archerSvc := &testArcherService{}
	sessionSvc := &testSessionService{}
	slotSvc := &testSlotService{}
	shotSvc := &testShotService{}
	faceSvc := &testFaceService{}
	maintSvc := &testMaintenanceService{schemaVer: 10}

	deps := RouterDeps{
		Cfg:            cfg,
		Logger:         logger,
		AuthSvc:        authSvc,
		AuthHandler:    handler.NewAuthHandler(authSvc, archerSvc, handler.AuthHandlerConfig{DevMode: true, JWTTTLMinutes: 1440}),
		ArcherHandler:  handler.NewArcherHandler(archerSvc),
		SessionHandler: handler.NewSessionHandler(sessionSvc),
		SlotHandler:    handler.NewSlotHandler(slotSvc),
		ShotHandler:    handler.NewShotHandler(shotSvc),
		FaceHandler:    handler.NewFaceHandler(faceSvc),
		HealthHandler:  handler.NewHealthHandler(maintSvc),
	}

	return buildRouter(deps)
}

func TestRouter_PublicRoutes(t *testing.T) {
	r := setupTestRouter()

	tests := []struct {
		name       string
		method     string
		url        string
		body       string
		wantStatus int
	}{
		{
			name:       "Health check returns 200",
			method:     http.MethodGet,
			url:        "/api/v0/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Faces list returns 200",
			method:     http.MethodGet,
			url:        "/api/v0/faces",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Faces get by type returns 200",
			method:     http.MethodGet,
			url:        "/api/v0/faces/WA_field",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Auth login returns 200 with valid body",
			method:     http.MethodPost,
			url:        "/api/v0/auth/login",
			body:       `{"credential": "google-test-token"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Auth google alias returns 200 with valid body",
			method:     http.MethodPost,
			url:        "/api/v0/auth/google",
			body:       `{"credential": "google-test-token"}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.url, body)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestRouter_ProtectedRoutes_RequireAuth(t *testing.T) {
	r := setupTestRouter()

	protectedURLs := []struct {
		method string
		url    string
	}{
		{http.MethodPost, "/api/v0/auth/logout"},
		{http.MethodGet, "/api/v0/auth/me"},
		{http.MethodGet, "/api/v0/archer"},
		{http.MethodGet, "/api/v0/session/open"},
		{http.MethodGet, "/api/v0/session/slot/00000000-0000-0000-0000-000000000001"},
		{http.MethodPost, "/api/v0/shot"},
	}

	for _, tt := range protectedURLs {
		t.Run(tt.method+" "+tt.url+" unauthenticated returns 401", func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("got status %d, want 401 Unauthorized for %s %s; body: %s", rec.Code, tt.method, tt.url, rec.Body.String())
			}

			var errResp middleware.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}
			if errResp.Code != "UNAUTHORIZED" {
				t.Errorf("expected code UNAUTHORIZED, got %q", errResp.Code)
			}
		})
	}
}

func TestRouter_ProtectedRoutes_WithValidAuth(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 OK with valid auth; body: %s", rec.Code, rec.Body.String())
	}
}

func TestRouter_CORSPreflight(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodOptions, "/api/v0/faces", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want 204 No Content", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("missing or incorrect Access-Control-Allow-Origin: %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./cmd/arch-stats -v
```

Expected: compilation failure (`RouterDeps` and `buildRouter` undefined).

- [ ] **Step 3: Implement `backend/cmd/arch-stats/router.go`**

Create `backend/cmd/arch-stats/router.go`:

```go
package main

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/config"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

// RouterDeps encapsulates all router dependencies.
type RouterDeps struct {
	Cfg            *config.Config
	Logger         *slog.Logger
	AuthSvc        middleware.TokenAuthenticator
	AuthHandler    *handler.AuthHandler
	ArcherHandler  *handler.ArcherHandler
	SessionHandler *handler.SessionHandler
	SlotHandler    *handler.SlotHandler
	ShotHandler    *handler.ShotHandler
	FaceHandler    *handler.FaceHandler
	HealthHandler  *handler.HealthHandler
}

// buildRouter constructs the chi HTTP router with global middleware and nested route groups.
func buildRouter(deps RouterDeps) chi.Router {
	r := chi.NewRouter()

	// Global middleware stack: logging -> recovery -> CORS
	r.Use(middleware.RequestLogger(deps.Logger))
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS(deps.Cfg.DevMode))

	r.Route("/api/v0", func(r chi.Router) {
		// Public health check endpoint
		r.Get("/health", deps.HealthHandler.Health)

		// Public face catalog endpoints
		r.Route("/faces", deps.FaceHandler.Routes)

		// Auth route group: public login/google/register, protected logout/me
		r.Route("/auth", func(r chi.Router) {
			r.Use(middleware.ErrorMapper)

			r.Post("/login", deps.AuthHandler.Login)
			r.Post("/google", deps.AuthHandler.Login)
			r.Post("/register", deps.AuthHandler.Register)

			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(deps.AuthSvc))
				r.Post("/logout", deps.AuthHandler.Logout)
				r.Get("/me", deps.AuthHandler.Me)
			})
		})

		// Protected route groups requiring valid authentication token
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(deps.AuthSvc))
			r.Use(middleware.ErrorMapper)

			r.Route("/archer", deps.ArcherHandler.Routes)
			r.Route("/session", func(r chi.Router) {
				r.Route("/slot", deps.SlotHandler.Routes)
				deps.SessionHandler.Routes(r)
			})
			r.Route("/shot", deps.ShotHandler.Routes)
		})
	})

	return r
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./cmd/arch-stats -v
```

Expected: PASS for all tests in `router_test.go`.

- [ ] **Step 5: Update `docs/plans/task.md`**

```markdown
| Task 3: Chi Router Construction, Middleware Stack & Routing Tests | DONE | Implement `buildRouter` in `cmd/arch-stats/router.go` and comprehensive route tests in `router_test.go` |
```

- [ ] **Step 6: Commit**

```bash
git add backend/cmd/arch-stats/router.go backend/cmd/arch-stats/router_test.go docs/plans/task.md
git commit -m "feat(cmd): build chi router with middleware stack and route groups"
```

---

### Task 4: Main Application Wiring & Graceful Server Lifecycle

**Files:**
- Modify: `backend/cmd/arch-stats/main.go`

**Interfaces:**
- Consumes: All configuration, repositories, services, handlers, and `buildRouter`.
- Produces: Runnable server entrypoint with graceful shutdown on interrupt signals.

- [ ] **Step 1: Update `backend/cmd/arch-stats/main.go`**

Refactor `backend/cmd/arch-stats/main.go` with full dependency injection:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/config"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
	"github.com/jpmolinamatute/arch-stats/backend/internal/service"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return err
	}

	// 2. Logger
	logger := config.NewLogger(cfg.DevMode)
	slog.SetDefault(logger)
	logger.Info("arch-stats starting", "dev_mode", cfg.DevMode)

	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		return err
	}

	dsn, err := cfg.DatabaseURL()
	if err != nil {
		slog.Error("failed to construct database URL", "error", err)
		return err
	}

	// 3. Database Connection Pool
	slog.Info("connecting to database...")
	pool, err := repository.NewPool(ctx, dsn, int32(cfg.PostgresPoolMinSize), int32(cfg.PostgresPoolMaxSize))
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return err
	}
	defer pool.Close()

	slog.Info("database connection pool initialized",
		"min_conns", cfg.PostgresPoolMinSize,
		"max_conns", cfg.PostgresPoolMaxSize,
	)

	// Standalone migration CLI subcommand
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		slog.Info("running database migrations...")
		if err := repository.RunMigrations(ctx, pool, "migrations"); err != nil {
			slog.Error("migration failed", "error", err)
			return err
		}
		slog.Info("migrations applied successfully")
		return nil
	}

	// 4. Migrations (Startup auto-migration if configured)
	if cfg.ApplyMigrationsOnStart {
		slog.Info("applying database migrations on startup...")
		if err := repository.RunMigrations(ctx, pool, "migrations"); err != nil {
			slog.Error("startup migration failed", "error", err)
			return err
		}
		slog.Info("startup migrations applied successfully")
	}

	// 5. Repositories
	archerRepo := repository.NewArcherRepo(pool)
	authSessionRepo := repository.NewAuthSessionRepo(pool)
	sessionRepo := repository.NewSessionRepo(pool)
	slotRepo := repository.NewSlotRepo(pool)
	shotRepo := repository.NewShotRepo(pool)
	faceRepo := repository.NewFaceRepo(pool)
	targetRepo := repository.NewTargetRepo(pool)
	maintenanceRepo := repository.NewMaintenanceRepo(pool)
	reportingRepo := repository.NewReportingRepo(pool)
	_ = reportingRepo // Reserved for reporting queries

	// 5b. Log Current Database Schema Version
	if ver, err := maintenanceRepo.GetSchemaVersion(ctx); err == nil {
		slog.Info("database schema version", "version", ver)
	} else {
		slog.Warn("could not read schema version", "error", err)
	}

	// 6. Services
	archerSvc := service.NewArcherService(archerRepo)
	sessionSvc := service.NewSessionService(sessionRepo)
	slotSvc := service.NewSlotService(slotRepo, sessionRepo)
	shotSvc := service.NewShotService(shotRepo, slotRepo)
	faceSvc := service.NewFaceService(faceRepo)
	_ = service.NewTargetService(targetRepo, faceRepo) // Reserved for target operations

	// 7. Auth Service
	authCfg := auth.Config{
		JWTSecret:           cfg.JWTSecret,
		JWTAlgorithm:        cfg.JWTAlgorithm,
		JWTTTLMinutes:       cfg.JWTTTLMinutes,
		SessionTokenBytes:   cfg.SessionTokenBytes,
		GoogleOAuthClientID: cfg.GoogleOAuthClientID,
	}
	authSvc := auth.NewService(archerRepo, authSessionRepo, authCfg)

	// 8. Handlers
	authHandlerCfg := handler.AuthHandlerConfig{
		JWTTTLMinutes: cfg.JWTTTLMinutes,
		DevMode:       cfg.DevMode,
	}
	authHandler := handler.NewAuthHandler(authSvc, archerSvc, authHandlerCfg)
	archerHandler := handler.NewArcherHandler(archerSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	slotHandler := handler.NewSlotHandler(slotSvc)
	shotHandler := handler.NewShotHandler(shotSvc)
	faceHandler := handler.NewFaceHandler(faceSvc)
	healthHandler := handler.NewHealthHandler(maintenanceRepo)

	// 9. Build Chi Router
	routerDeps := RouterDeps{
		Cfg:            cfg,
		Logger:         logger,
		AuthSvc:        authSvc,
		AuthHandler:    authHandler,
		ArcherHandler:  archerHandler,
		SessionHandler: sessionHandler,
		SlotHandler:    slotHandler,
		ShotHandler:    shotHandler,
		FaceHandler:    faceHandler,
		HealthHandler:  healthHandler,
	}
	router := buildRouter(routerDeps)

	// 10. HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info(fmt.Sprintf("arch-stats listening on :%d", cfg.ServerPort), "port", cfg.ServerPort, "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		slog.Info("shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		slog.Info("server stopped gracefully")
	}

	return nil
}
```

- [ ] **Step 2: Verify `main.go` builds cleanly**

```bash
cd backend && go build ./cmd/arch-stats
```

Expected: zero errors, binary `arch-stats` produced. Clean it up with `rm -f arch-stats`.

- [ ] **Step 3: Update `docs/plans/task.md`**

```markdown
| Task 4: Main Application Wiring & Graceful Server Lifecycle | DONE | Refactor `cmd/arch-stats/main.go` with full dependency injection and graceful shutdown |
```

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/arch-stats/main.go docs/plans/task.md
git commit -m "feat(cmd): complete dependency injection wiring and graceful server lifecycle in main.go"
```

---

### Task 5: End-to-End Suite Verification, Formatting & Linting

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: All handler, router, and main code.
- Produces: Clean lint, formatted code, passing tests, and verified build.

- [ ] **Step 1: Run all unit tests across backend**

```bash
cd backend && go test -race ./... -v
```

Expected: PASS for all packages without race condition warnings.

- [ ] **Step 2: Run `go vet`**

```bash
cd backend && go vet ./...
```

Expected: zero issues reported.

- [ ] **Step 3: Run Go linting and formatting verification**

```bash
./scripts/linting.bash --go
```

Expected: zero linting errors from `golangci-lint` and `gofumpt`.

- [ ] **Step 4: Test build of backend binary**

```bash
cd backend && go build -o /dev/null ./cmd/arch-stats
```

Expected: clean compilation.

- [ ] **Step 5: Update `docs/plans/task.md`**

```markdown
| Task 5: End-to-End Suite Verification, Formatting & Linting | DONE | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
```

- [ ] **Step 6: Commit**

```bash
git add docs/plans/task.md
git commit -m "chore: verify tests, linting, and build pass cleanly"
```

---

### Task 6: Mark Tasks as Completed in Task Spec and Live Tracker

**Files:**
- Modify: `docs/go_refactor/tasks/025-chi_router_and_main_wiring.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified implementation of Task 025.
- Produces: Completed checkboxes in task specification and completed live tracker table.

- [ ] **Step 1: Update `docs/go_refactor/tasks/025-chi_router_and_main_wiring.md`**

Mark all acceptance criteria and steps as completed:

```markdown
## Acceptance Criteria

- [x] `backend/cmd/arch-stats/main.go` performs full dependency wiring:
  1. Load config
  2. Create logger
  3. Connect database pool
  4. Run migrations (if configured)
  5. Log current schema version via `MaintenanceRepo.GetSchemaVersion()`
  6. Create all repositories (passing pool), including `MaintenanceRepo` and `ReportingRepo`
  7. Create all services (passing repositories)
  8. Create auth service (passing repos + config)
  9. Create all handlers (passing services)
  10. Build chi router with route groups
  11. Apply middleware stack
  12. Start HTTP server with graceful shutdown
- [x] Route groups match the Python API structure:
    - `/api/v0/auth/` — auth handler (public: login, register; protected: logout, me)
    - `/api/v0/archer/` — archer handler (protected)
    - `/api/v0/session/` — session handler (protected)
    - `/api/v0/session/slot/` — slot handler (protected)
    - `/api/v0/shot/` — shot handler (protected)
    - `/api/v0/faces/` — face handler (public)
    - `/api/v0/stats/` — live stats handler (protected) (Note: stats handler scheduled for Task 027)
- [x] Middleware stack applied in correct order: logging → recovery → CORS → (per-group auth)
  → error mapper.
- [x] A `GET /api/v0/health` endpoint returns JSON with at minimum the current database schema
  version (from `MaintenanceRepo.GetSchemaVersion()`). This is a public, unauthenticated endpoint.
- [x] `go build ./cmd/arch-stats` compiles cleanly.
- [x] `go vet ./...` reports no issues.
- [x] Running the binary with valid DB config starts the HTTP server and logs "listening on :PORT".

## Steps

- [x] **Step 1: Add chi dependency**
- [x] **Step 2: Refactor `main.go` into a structured wiring function**
- [x] **Step 3: Build the chi router with route groups**
- [x] **Step 4: Run go vet and build**
- [x] **Step 5: Manual verification**
- [x] **Step 6: Commit**
```

- [ ] **Step 2: Update `docs/plans/task.md`**

Ensure all rows in `docs/plans/task.md` are marked `DONE`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | DONE | Switch to branch `refactor/025-chi-router-and-main-wiring` and initialize tracker |
| Task 2: Health Handler Implementation | DONE | Implement `HealthHandler` in `internal/handler/health.go` and unit tests in `health_test.go` |
| Task 3: Chi Router Construction, Middleware Stack & Routing Tests | DONE | Implement `buildRouter` in `cmd/arch-stats/router.go` and comprehensive route tests in `router_test.go` |
| Task 4: Main Application Wiring & Graceful Server Lifecycle | DONE | Refactor `cmd/arch-stats/main.go` with full dependency injection and graceful shutdown |
| Task 5: End-to-End Suite Verification, Formatting & Linting | DONE | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 6: Mark Tasks as Completed in Task Spec and Live Tracker | DONE | Mark `025-chi_router_and_main_wiring.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Commit completed documentation updates**

```bash
git add docs/go_refactor/tasks/025-chi_router_and_main_wiring.md docs/plans/task.md
git commit -m "docs: mark task 025 chi router and main wiring as completed"
```
