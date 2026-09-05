# Task 021: Build HTTP Handler — Sessions Endpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the sessions HTTP handler in `backend/internal/handler/session.go` and its comprehensive test suite in `backend/internal/handler/session_test.go`, porting all 8 endpoints from Python `session_router.py`. All endpoints enforce authentication context extraction, archer ownership validation, error translation via `writeAppError`, and chi router mounting. Mark all acceptance criteria and steps in `docs/go_refactor/tasks/021-handler_sessions.md` and `docs/plans/task.md` as completed upon finishing.

**Architecture:**
- `session.go`: `SessionHandler` exposing `GetOpenForArcher`, `GetClosedForArcher`, `GetParticipating`, `ListAllOpen`, `Create`, `GetByID`, `ReOpen`, `Close`, and `Routes` methods.
- `SessionService` interface: Defined in `package handler` specifying the domain contract needed by `SessionHandler`: `GetByID`, `GetOpen`, `List`, `Create`, `Close`, `ReOpen`, and `GetParticipating`.
- `helpers.go`: Shared HTTP helpers (`readJSON`, `writeJSON`, `writeError`, `writeAppError`) for JSON serialization and envelope management.
- `context.go`: Authenticated archer identity is retrieved via `middleware.GetArcherID(r.Context())`. Unauthorized requests without context return 401. Mismatched archer access returns 403.
- `error_mapper`: Translates domain sentinel errors (`apperror.ErrNotFound` -> 404, `apperror.ErrForbidden` -> 403, `apperror.ErrConflict` -> 409, `apperror.ErrValidation` -> 422).
- `service/session.go` & `repository/session.go`: Add `GetParticipating` / `FindParticipating` method so `*service.SessionService` satisfies `handler.SessionService`.

**Tech Stack:** Go 1.27+, `github.com/go-chi/chi/v5`, `net/http`, `net/http/httptest`, `encoding/json`, `github.com/google/uuid`, internal packages (`apperror`, `middleware`, `model`, `service`, `repository`).

**Spec:**
- [docs/go_refactor/tasks/021-handler_sessions.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/021-handler_sessions.md)
- [backend-old/src/routers/v0/session_router.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/routers/v0/session_router.py)
- [backend-old/src/core/session_manager.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/core/session_manager.py)
- [backend-old/tests/endpoints/test_session_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend-old/tests/endpoints/test_session_endpoints.py)

## Global Constraints

- Git branch: `refactor/021-handler-sessions`
- Package paths:
  - `github.com/jpmolinamatute/arch-stats/backend/internal/handler`
  - `github.com/jpmolinamatute/arch-stats/backend/internal/service`
  - `github.com/jpmolinamatute/arch-stats/backend/internal/repository`
- Error handling: Wrap internal errors with `%w` or `apperror.Wrap`. Domain errors must map to HTTP 400, 401, 403, 404, 409, 422 via `writeAppError(w, err)` and `writeError(w, status, msg)`.
- Authentication: All endpoints extract authenticated archer ID via `middleware.GetArcherID(r.Context())`.
- Authorization:
  - `GetOpenForArcher`: authenticated archer ID must match URL `archer_id` (403 if mismatch).
  - `GetClosedForArcher`: authenticated archer ID must match URL `archer_id` (403 if mismatch).
  - `GetParticipating`: authenticated archer ID must match URL `archer_id` (403 if mismatch).
  - `Create`: authenticated archer ID must match payload `owner_archer_id` (403 if mismatch: `"ERROR: user not allowed to open a session for another archer"`).
  - `GetByID`: if session is closed (`!session.IsOpened`), only the owner can view (403 if mismatch). If open, any authenticated archer can view.
  - `ReOpen`: only session owner can re-open (403 if mismatch: `"Archer is not allowed to re-open this session"`).
  - `Close`: only session owner can close (403 if mismatch). If `session_id` missing in body, 400 Bad Request (`"ERROR: session_id wasn't provided"`).
- Formatting & Linting: Code must pass `gofumpt` and `golangci-lint run ./...` (`./scripts/linting.bash --go`).
- Tests: `go test -race -v ./internal/handler/...` and `go test -race -v ./internal/service/...` must pass.
- Verification: `go vet ./...` and `go build ./...` must succeed.

---

## File Structure

```
backend/
├── internal/
│   ├── repository/
│   │   ├── session.go             # [MODIFY] Add FindParticipating(ctx, archerID) (*uuid.UUID, error)
│   │   └── session_test.go        # [MODIFY] Add unit test for FindParticipating
│   ├── service/
│   │   ├── session.go             # [MODIFY] Add FindParticipating to SessionRepository, add GetParticipating
│   │   └── session_test.go        # [MODIFY] Add unit test for GetParticipating
│   └── handler/
│       ├── session.go             # [NEW] SessionService interface, SessionHandler struct, 8 endpoints, Routes
│       └── session_test.go        # [NEW] Comprehensive httptest suite verifying all 8 endpoints, auth, and error states
docs/
├── plans/
│   ├── task.md                    # [MODIFY] Track Task 021 live checklist progress (table-only)
│   └── 2026-09-05-handler-sessions.md # [NEW] Implementation plan document
└── go_refactor/
    └── tasks/
        └── 021-handler_sessions.md # [MODIFY] Mark all acceptance criteria and steps as completed
```

---

## Proposed Tasks

### Task 1: Git Branch Setup & Live Tracker Initialization

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Current repository state on main.
- Produces: `refactor/021-handler-sessions` branch and updated `docs/plans/task.md` table.

- [ ] **Step 1: Switch to new feature branch**

```bash
git switch -c refactor/021-handler-sessions
```

- [ ] **Step 2: Update live tracker table in `docs/plans/task.md`**

Replace `docs/plans/task.md` with:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker | IN_PROGRESS | Switch to `refactor/021-handler-sessions` and initialize task tracker |
| Task 2: Service & Repo GetParticipating Extension | PENDING | Add `FindParticipating` to SessionRepo and `GetParticipating` to SessionService |
| Task 3: SessionService Interface & Handler Scaffolding | PENDING | Define SessionService interface and SessionHandler struct with chi Routes |
| Task 4: Implement Read Endpoints (TDD) | PENDING | Implement GetOpenForArcher, GetClosedForArcher, GetParticipating, ListAllOpen, GetByID |
| Task 5: Implement Mutation Endpoints (TDD) | PENDING | Implement Create, ReOpen, Close with validation and error mappings |
| Task 6: Full Verification Suite & Linting | PENDING | Run race tests, go vet, golangci-lint, gofumpt, and go build |
| Task 7: Mark Tasks as Completed | PENDING | Check all boxes in 021-handler_sessions.md and task.md |
```

- [ ] **Step 3: Commit tracker initialization**

```bash
git add docs/plans/task.md
git commit -m "docs: initialize task tracker for refactor/021-handler-sessions"
```

---

### Task 2: Service & Repository `GetParticipating` Extension (TDD)

**Files:**
- Modify: `backend/internal/repository/session.go`
- Modify: `backend/internal/repository/session_test.go`
- Modify: `backend/internal/service/session.go`
- Modify: `backend/internal/service/session_test.go`

**Interfaces:**
- Consumes: `slot` and `session` tables in database.
- Produces:
  - `repository.SessionRepo.FindParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)`
  - `service.SessionService.GetParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)`

- [ ] **Step 1: Write failing unit test for `SessionService.GetParticipating`**

Add in `backend/internal/service/session_test.go`:

```go
func TestSessionService_GetParticipating(t *testing.T) {
	t.Run("returns session id when archer is participating", func(t *testing.T) {
		archerID := uuid.New()
		sessionID := uuid.New()
		repo := &mockSessionRepo{
			findParticipatingFn: func(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
				if id == archerID {
					return &sessionID, nil
				}
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		got, err := svc.GetParticipating(context.Background(), archerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || *got != sessionID {
			t.Fatalf("expected session ID %v, got %v", sessionID, got)
		}
	})

	t.Run("returns nil when archer is not participating", func(t *testing.T) {
		archerID := uuid.New()
		repo := &mockSessionRepo{
			findParticipatingFn: func(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		got, err := svc.GetParticipating(context.Background(), archerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil session ID, got %v", got)
		}
	})

	t.Run("returns validation error for nil archer ID", func(t *testing.T) {
		repo := &mockSessionRepo{}
		svc := service.NewSessionService(repo)

		_, err := svc.GetParticipating(context.Background(), uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected apperror.ErrValidation, got %v", err)
		}
	})
}
```

- [ ] **Step 2: Update `mockSessionRepo` in `session_test.go`**

Add `findParticipatingFn func(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)` and method:

```go
func (m *mockSessionRepo) FindParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error) {
	if m.findParticipatingFn != nil {
		return m.findParticipatingFn(ctx, archerID)
	}
	return nil, nil
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd backend && go test ./internal/service/... -run TestSessionService_GetParticipating
```
Expected: FAIL (method `GetParticipating` undefined).

- [ ] **Step 4: Implement `FindParticipating` in `SessionRepository` and `SessionService`**

In `backend/internal/service/session.go`:
Add to `SessionRepository` interface:
```go
FindParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)
```

Add method to `SessionService`:
```go
// GetParticipating retrieves the active open session ID that the archer is currently shooting in, if any.
// Returns nil, nil if the archer is not participating in any active session.
func (s *SessionService) GetParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error) {
	if archerID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "archer id is required")
	}

	sessionID, err := s.repo.FindParticipating(ctx, archerID)
	if err != nil {
		return nil, fmt.Errorf("checking active session participation: %w", err)
	}
	return sessionID, nil
}
```

In `backend/internal/repository/session.go`:
Add method `FindParticipating`:
```go
// FindParticipating queries whether the specified archer is assigned to an active shooting slot in an open session.
// Returns the session ID pointer if participating, or nil, nil if not found.
func (r *SessionRepo) FindParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error) {
	sql, args, err := StmtBuilder.Select("s.session_id").
		From("slot s").
		Join("session ses ON s.session_id = ses.session_id").
		Where(squirrel.Eq{"s.archer_id": archerID}).
		Where(squirrel.Eq{"s.is_shooting": true}).
		Where(squirrel.Eq{"ses.is_opened": true}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find participating query: %w", err)
	}

	var sessionID uuid.UUID
	err = r.db.QueryRow(ctx, sql, args...).Scan(&sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying participating session: %w", err)
	}

	return &sessionID, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd backend && go test -race -v ./internal/service/... ./internal/repository/...
```
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/repository/session.go backend/internal/repository/session_test.go backend/internal/service/session.go backend/internal/service/session_test.go
git commit -m "feat: add FindParticipating to session repository and GetParticipating to session service"
```

---

### Task 3: `SessionService` Interface, `SessionHandler` Scaffolding & Routing

**Files:**
- Create: `backend/internal/handler/session.go`
- Create: `backend/internal/handler/session_test.go`

**Interfaces:**
- Consumes: `github.com/go-chi/chi/v5`, `model.SessionRead`, `model.SessionCreate`, `model.SessionFilter`, `model.SessionID`.
- Produces:
  - `handler.SessionService` interface
  - `handler.SessionHandler` struct
  - `handler.NewSessionHandler(sessionSvc SessionService) *SessionHandler`
  - `Routes(r chi.Router)` method mounting all 8 endpoints

- [ ] **Step 1: Write failing test verifying constructor and route registration**

In `backend/internal/handler/session_test.go`:

```go
package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockSessionHandlerService struct {
	getByIDFn          func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	getOpenFn          func(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	listFn             func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	createFn           func(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	closeFn            func(ctx context.Context, id uuid.UUID) error
	reOpenFn           func(ctx context.Context, id uuid.UUID) error
	getParticipatingFn func(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)
}

func (m *mockSessionHandlerService) GetByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSessionHandlerService) GetOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error) {
	if m.getOpenFn != nil {
		return m.getOpenFn(ctx, archerID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSessionHandlerService) List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSessionHandlerService) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, errors.New("unimplemented")
}

func (m *mockSessionHandlerService) Close(ctx context.Context, id uuid.UUID) error {
	if m.closeFn != nil {
		return m.closeFn(ctx, id)
	}
	return errors.New("unimplemented")
}

func (m *mockSessionHandlerService) ReOpen(ctx context.Context, id uuid.UUID) error {
	if m.reOpenFn != nil {
		return m.reOpenFn(ctx, id)
	}
	return errors.New("unimplemented")
}

func (m *mockSessionHandlerService) GetParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error) {
	if m.getParticipatingFn != nil {
		return m.getParticipatingFn(ctx, archerID)
	}
	return nil, errors.New("unimplemented")
}

func TestNewSessionHandler(t *testing.T) {
	svc := &mockSessionHandlerService{}
	h := handler.NewSessionHandler(svc)
	if h == nil {
		t.Fatal("expected NewSessionHandler to return non-nil instance")
	}
}

func TestSessionHandler_RoutesRegistration(t *testing.T) {
	svc := &mockSessionHandlerService{}
	h := handler.NewSessionHandler(svc)
	r := chi.NewRouter()
	r.Route("/api/v0/session", h.Routes)

	// Verify route registration by making an options or not-allowed request
	req := httptest.NewRequest(http.MethodGet, "/api/v0/session/open", http.NoBody)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should not be 404 (not found)
	if rec.Code == http.StatusNotFound {
		t.Errorf("expected route /api/v0/session/open to be registered, got 404")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/handler/... -run TestNewSessionHandler
```
Expected: FAIL (handler.NewSessionHandler undefined).

- [ ] **Step 3: Implement `backend/internal/handler/session.go` scaffolding**

```go
package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// SessionService defines domain operations required by SessionHandler.
type SessionService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	GetOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	Close(ctx context.Context, id uuid.UUID) error
	ReOpen(ctx context.Context, id uuid.UUID) error
	GetParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)
}

// SessionHandler manages HTTP endpoints for session lifecycle and querying.
type SessionHandler struct {
	sessionSvc SessionService
}

// NewSessionHandler constructs a SessionHandler with service dependency injection.
func NewSessionHandler(sessionSvc SessionService) *SessionHandler {
	return &SessionHandler{
		sessionSvc: sessionSvc,
	}
}

// Routes registers all session lifecycle and query endpoints on the provided chi Router.
func (h *SessionHandler) Routes(r chi.Router) {
	r.Get("/archer/{archer_id}/open-session", h.GetOpenForArcher)
	r.Get("/archer/{archer_id}/close-session", h.GetClosedForArcher)
	r.Get("/archer/{archer_id}/participating", h.GetParticipating)
	r.Get("/open", h.ListAllOpen)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Patch("/re-open", h.ReOpen)
	r.Patch("/close", h.Close)
}
```

Add stub handler methods returning 501 / empty to satisfy compilation.

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/handler/... -run 'TestNewSessionHandler|TestSessionHandler_RoutesRegistration'
```
Expected: PASS.

- [ ] **Step 5: Commit scaffolding**

```bash
git add backend/internal/handler/session.go backend/internal/handler/session_test.go
git commit -m "feat: scaffold SessionHandler struct, SessionService interface, and route registration"
```

---

### Task 4: Implement Read Endpoints (TDD)

**Files:**
- Modify: `backend/internal/handler/session.go`
- Modify: `backend/internal/handler/session_test.go`

**Endpoints:**
1. `GetOpenForArcher(w, r)` — GET `/archer/{archer_id}/open-session`
2. `GetClosedForArcher(w, r)` — GET `/archer/{archer_id}/close-session`
3. `GetParticipating(w, r)` — GET `/archer/{archer_id}/participating`
4. `ListAllOpen(w, r)` — GET `/open`
5. `GetByID(w, r)` — GET `/{id}`

- [ ] **Step 1: Write failing tests for read endpoints in `session_test.go`**

Test scenarios:
- `GetOpenForArcher`:
  - 401 Unauthorized when unauthenticated
  - 422 Unprocessable Entity when invalid archer UUID
  - 403 Forbidden when authenticated archer != target archer
  - 200 OK with `{"session_id": null}` when no open session exists (service returns `apperror.ErrNotFound`)
  - 200 OK with `{"session_id": "<uuid>"}` when open session exists
- `GetClosedForArcher`:
  - 401 Unauthorized when unauthenticated
  - 403 Forbidden when authenticated archer != target archer
  - 200 OK with empty array `[]` when no closed sessions exist
  - 200 OK with array of `model.SessionRead` when closed sessions exist
- `GetParticipating`:
  - 401 Unauthorized when unauthenticated
  - 403 Forbidden when authenticated archer != target archer
  - 200 OK with `{"session_id": null}` when not participating
  - 200 OK with `{"session_id": "<uuid>"}` when participating
- `ListAllOpen`:
  - 401 Unauthorized when unauthenticated
  - 200 OK with empty array `[]` when no open sessions exist
  - 200 OK with list of open sessions
- `GetByID`:
  - 401 Unauthorized when unauthenticated
  - 422 Unprocessable Entity when invalid session UUID
  - 404 Not Found when session does not exist
  - 403 Forbidden when session is closed and authenticated archer is not owner
  - 200 OK when session is open (even if authenticated archer is not owner)
  - 200 OK when session is closed and authenticated archer IS owner

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./internal/handler/... -run 'TestSessionHandler_Get'
```
Expected: FAIL.

- [ ] **Step 3: Implement read methods in `backend/internal/handler/session.go`**

```go
// GetOpenForArcher handles GET /api/v0/session/archer/{archer_id}/open-session.
// Returns the open session ID owned by the archer, or null session_id if none exists.
// Enforces that only the authenticated archer can query their open session.
func (h *SessionHandler) GetOpenForArcher(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	archerIDStr := getURLParam(r, "archer_id")
	archerID, err := uuid.Parse(archerIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer_id is required"))
		return
	}

	if authArcherID != archerID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	session, err := h.sessionSvc.GetOpen(r.Context(), archerID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			_ = writeJSON(w, http.StatusOK, model.SessionID{SessionID: nil})
			return
		}
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, model.SessionID{SessionID: &session.SessionID})
}

// GetClosedForArcher handles GET /api/v0/session/archer/{archer_id}/close-session.
// Returns all closed sessions owned by the archer.
// Enforces that only the authenticated archer can query their closed sessions.
func (h *SessionHandler) GetClosedForArcher(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	archerIDStr := getURLParam(r, "archer_id")
	archerID, err := uuid.Parse(archerIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer_id is required"))
		return
	}

	if authArcherID != archerID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	isOpened := false
	filter := model.SessionFilter{
		OwnerArcherID: &archerID,
		IsOpened:      &isOpened,
	}

	sessions, err := h.sessionSvc.List(r.Context(), filter)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if sessions == nil {
		sessions = []model.SessionRead{}
	}

	_ = writeJSON(w, http.StatusOK, sessions)
}

// GetParticipating handles GET /api/v0/session/archer/{archer_id}/participating.
// Returns the open session ID the archer is currently participating in, or null if none.
// Enforces that only the authenticated archer can query their participation.
func (h *SessionHandler) GetParticipating(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	archerIDStr := getURLParam(r, "archer_id")
	archerID, err := uuid.Parse(archerIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer_id is required"))
		return
	}

	if authArcherID != archerID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	sessionID, err := h.sessionSvc.GetParticipating(r.Context(), archerID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, model.SessionID{SessionID: sessionID})
}

// ListAllOpen handles GET /api/v0/session/open.
// Returns all currently open sessions. Requires authentication.
func (h *SessionHandler) ListAllOpen(w http.ResponseWriter, r *http.Request) {
	if _, err := middleware.GetArcherID(r.Context()); err != nil {
		writeAppError(w, err)
		return
	}

	isOpened := true
	filter := model.SessionFilter{
		IsOpened: &isOpened,
	}

	sessions, err := h.sessionSvc.List(r.Context(), filter)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if sessions == nil {
		sessions = []model.SessionRead{}
	}

	_ = writeJSON(w, http.StatusOK, sessions)
}

// GetByID handles GET /api/v0/session/{id}.
// Returns full session details. Open sessions are readable by any authenticated archer.
// Closed sessions are only readable by the session owner (403 Forbidden otherwise).
func (h *SessionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	idStr := getURLParam(r, "id")
	if idStr == "" {
		idStr = getURLParam(r, "session")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid session id is required"))
		return
	}

	session, err := h.sessionSvc.GetByID(r.Context(), id)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if !session.IsOpened && session.OwnerArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	_ = writeJSON(w, http.StatusOK, session)
}
```

- [ ] **Step 4: Run tests to verify read endpoints pass**

```bash
cd backend && go test -race -v ./internal/handler/... -run 'TestSessionHandler_Get|TestSessionHandler_List'
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/session.go backend/internal/handler/session_test.go
git commit -m "feat: implement session read endpoints (GetOpenForArcher, GetClosedForArcher, GetParticipating, ListAllOpen, GetByID)"
```

---

### Task 5: Implement Mutation Endpoints (TDD)

**Files:**
- Modify: `backend/internal/handler/session.go`
- Modify: `backend/internal/handler/session_test.go`

**Endpoints:**
1. `Create(w, r)` — POST `/api/v0/session` — returns 201 Created
2. `ReOpen(w, r)` — PATCH `/api/v0/session/re-open` — returns 200 OK
3. `Close(w, r)` — PATCH `/api/v0/session/close` — returns 200 OK

- [ ] **Step 1: Write failing tests for mutation endpoints in `session_test.go`**

Test scenarios:
- `Create`:
  - 401 Unauthorized when unauthenticated
  - 422 Unprocessable Entity when request body is empty or invalid JSON
  - 403 Forbidden with `"ERROR: user not allowed to open a session for another archer"` when `req.OwnerArcherID != authArcherID`
  - 409 Conflict when archer already has an open session (service returns `apperror.ErrConflict`)
  - 422 Unprocessable Entity when validation fails (service returns `apperror.ErrValidation`)
  - 201 Created with `{"session_id": "<uuid>"}` on success
- `ReOpen`:
  - 401 Unauthorized when unauthenticated
  - 422 Unprocessable Entity when body missing `session_id` or invalid
  - 404 Not Found when session does not exist (service returns `apperror.ErrNotFound`)
  - 403 Forbidden with `"Archer is not allowed to re-open this session"` when authenticated archer is not the session owner
  - 409 Conflict when owner already has another open session (service returns `apperror.ErrConflict`)
  - 422 Unprocessable Entity when session is already open (service returns `apperror.ErrValidation`)
  - 200 OK with `{"session_id": "<uuid>"}` on success
- `Close`:
  - 401 Unauthorized when unauthenticated
  - 400 Bad Request with `"ERROR: session_id wasn't provided"` when body missing `session_id`
  - 404 Not Found when session does not exist (service returns `apperror.ErrNotFound`)
  - 403 Forbidden when authenticated archer is not the session owner
  - 422 Unprocessable Entity when session is already closed (service returns `apperror.ErrValidation`)
  - 200 OK with `{"status": "closed"}` on success

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./internal/handler/... -run 'TestSessionHandler_(Create|ReOpen|Close)'
```
Expected: FAIL.

- [ ] **Step 3: Implement mutation methods in `backend/internal/handler/session.go`**

```go
// Create handles POST /api/v0/session.
// Creates a new shooting session and returns 201 Created with the session ID.
// Enforces that an archer can only create a session for themselves.
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	var req model.SessionCreate
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.OwnerArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "ERROR: user not allowed to open a session for another archer"))
		return
	}

	id, err := h.sessionSvc.Create(r.Context(), req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusCreated, model.SessionID{SessionID: &id})
}

// ReOpen handles PATCH /api/v0/session/re-open.
// Re-opens a closed shooting session after verifying owner identity and absence of conflicts.
// Returns 200 OK with the re-opened session ID.
func (h *SessionHandler) ReOpen(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	var req model.SessionID
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.SessionID == nil || *req.SessionID == uuid.Nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "session_id is required"))
		return
	}

	session, err := h.sessionSvc.GetByID(r.Context(), *req.SessionID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if session.OwnerArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Archer is not allowed to re-open this session"))
		return
	}

	if err := h.sessionSvc.ReOpen(r.Context(), *req.SessionID); err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, model.SessionID{SessionID: req.SessionID})
}

// Close handles PATCH /api/v0/session/close.
// Marks an active shooting session as closed. Returns 200 OK with {"status": "closed"}.
// Validates session presence, owner identity, and active open status.
func (h *SessionHandler) Close(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	var req model.SessionID
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.SessionID == nil || *req.SessionID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "ERROR: session_id wasn't provided")
		return
	}

	session, err := h.sessionSvc.GetByID(r.Context(), *req.SessionID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if session.OwnerArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	if err := h.sessionSvc.Close(r.Context(), *req.SessionID); err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}
```

- [ ] **Step 4: Run all handler tests to verify they pass**

```bash
cd backend && go test -race -v ./internal/handler/...
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/session.go backend/internal/handler/session_test.go
git commit -m "feat: implement session mutation endpoints (Create, ReOpen, Close)"
```

---

### Task 6: Full Verification Suite, Linting & Build

**Files:**
- Test across backend codebase.

**Interfaces:**
- Consumes: All updated and created Go source files.
- Produces: 0 lint errors, 0 vet issues, clean build artifact, 100% passing unit tests.

- [ ] **Step 1: Run comprehensive backend test suite with race detector**

```bash
cd backend && go test -race -v ./...
```
Expected: All tests pass across `apperror`, `auth`, `config`, `handler`, `middleware`, `model`, `repository`, `service`.

- [ ] **Step 2: Run go vet**

```bash
cd backend && go vet ./...
```
Expected: 0 issues.

- [ ] **Step 3: Run backend build**

```bash
cd backend && go build ./...
```
Expected: Success.

- [ ] **Step 4: Run full Go linting and formatting verification**

```bash
./scripts/linting.bash --go
```
Expected: PASS (gofumpt, golangci-lint).

---

### Task 7: Mark Tasks as Completed & Update Trackers

**Files:**
- Modify: `docs/go_refactor/tasks/021-handler_sessions.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified test and build results from Task 6.
- Produces: Updated markdown checklists marking all acceptance criteria and steps as completed.

- [ ] **Step 1: Mark all checkboxes as completed in `docs/go_refactor/tasks/021-handler_sessions.md`**

Update `docs/go_refactor/tasks/021-handler_sessions.md`:
Replace all `- [ ]` with `- [x]` in:
- `## Acceptance Criteria` (lines 22, 31, 33, 34, 35, 40, 41)
- `## Steps` (lines 57, 64, 70, 75, 81, 87)

- [ ] **Step 2: Update `docs/plans/task.md` to show all tasks DONE**

Update `docs/plans/task.md`:
```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker | DONE | Switch to `refactor/021-handler-sessions` and initialize task tracker |
| Task 2: Service & Repo GetParticipating Extension | DONE | Add `FindParticipating` to SessionRepo and `GetParticipating` to SessionService |
| Task 3: SessionService Interface & Handler Scaffolding | DONE | Define SessionService interface and SessionHandler struct with chi Routes |
| Task 4: Implement Read Endpoints (TDD) | DONE | Implement GetOpenForArcher, GetClosedForArcher, GetParticipating, ListAllOpen, GetByID |
| Task 5: Implement Mutation Endpoints (TDD) | DONE | Implement Create, ReOpen, Close with validation and error mappings |
| Task 6: Full Verification Suite & Linting | DONE | Run race tests, go vet, golangci-lint, gofumpt, and go build |
| Task 7: Mark Tasks as Completed | DONE | Check all boxes in 021-handler_sessions.md and task.md |
```

- [ ] **Step 3: Commit completion updates**

```bash
git add docs/go_refactor/tasks/021-handler_sessions.md docs/plans/task.md
git commit -m "docs: mark task 021 handler sessions and checklist as completed"
```
