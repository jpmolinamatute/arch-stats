# Task 022: Build HTTP Handler — Slots Endpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the slots HTTP handler in `backend/internal/handler/slot.go` and its comprehensive test suite in `backend/internal/handler/slot_test.go`, porting all 5 endpoints from Python `slot_router.py` (`GetArcherCurrentSlot`, `GetSlot`, `JoinSession`, `ReJoinSession`, `LeaveSession`). All endpoints enforce authentication context extraction, archer ownership/identity validation, error translation via `writeAppError`, and chi router mounting. Mark all acceptance criteria and steps in `docs/go_refactor/tasks/022-handler_slots.md` and `docs/plans/task.md` as completed upon finishing.

**Architecture:**
- `backend/internal/handler/slot.go`: Implements `SlotHandler` struct exposing `GetArcherCurrentSlot`, `GetSlot`, `JoinSession`, `ReJoinSession`, `LeaveSession`, and `Routes(r chi.Router)` methods.
- `SlotService` interface: Defined in `package handler` specifying the domain contract needed by `SlotHandler`: `GetArcherCurrentSlot`, `GetSlot`, `JoinSession`, `ReJoinSession`, and `LeaveSession`.
- `helpers.go`: Utilizes shared HTTP helpers (`readJSON`, `writeJSON`, `writeError`, `writeAppError`, `getURLParam`) for JSON serialization and envelope management.
- `context.go`: Authenticated archer identity is retrieved via `middleware.GetArcherID(r.Context())`. Missing auth context returns 401 Unauthorized; mismatched archer identity returns 403 Forbidden.
- Error translation: Translates domain sentinel errors (`apperror.ErrNotFound` -> 404, `apperror.ErrForbidden` -> 403, `apperror.ErrConflict` -> 409, `apperror.ErrValidation` -> 422, bad requests -> 400).
- Acceptance & task tracking: Ensure all acceptance checkboxes in `docs/go_refactor/tasks/022-handler_slots.md` and rows in `docs/plans/task.md` are marked completed.

**Tech Stack:** Go 1.27+, `github.com/go-chi/chi/v5`, `net/http`, `net/http/httptest`, `encoding/json`, `github.com/google/uuid`, internal packages (`apperror`, `middleware`, `model`).

**Spec:**
- [docs/go_refactor/tasks/022-handler_slots.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/022-handler_slots.md)
- [backend-old/src/routers/v0/slot_router.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/routers/v0/slot_router.py)
- [backend-old/src/core/slot_manager.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/core/slot_manager.py)
- [backend-old/tests/endpoints/test_slot_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend-old/tests/endpoints/test_slot_endpoints.py)

## Global Constraints

- Git branch: `refactor/022-handler-slots`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/handler`
- Error handling: Wrap internal errors with `%w` or `apperror.Wrap`. Domain errors map to HTTP 400, 401, 403, 404, 409, 422 via `writeAppError(w, err)` and `writeError(w, status, msg)`.
- Authentication: All 5 endpoints extract authenticated archer ID via `middleware.GetArcherID(r.Context())`.
- Authorization:
  - `GetArcherCurrentSlot`: Authenticated archer ID must match URL `archer_id` (403 if mismatch).
  - `GetSlot`: Authenticated archer ID must match slot owner `info.ArcherID` (403 if mismatch).
  - `JoinSession`: Authenticated archer ID must match payload `archer_id` (403 if mismatch).
  - `ReJoinSession`: Authenticated archer ID is passed to `slotSvc.ReJoinSession` (service verifies ownership; 403 if mismatch).
  - `LeaveSession`: Authenticated archer ID is passed to `slotSvc.LeaveSession` (service verifies ownership; 403 if mismatch).
- Formatting & Linting: Code must pass `gofumpt` and `golangci-lint run ./...` (`./scripts/linting.bash --go`).
- Tests: `go test -race -v ./internal/handler/...` must pass.
- Verification: `go vet ./...` and `go build ./...` must succeed.
- Task completion: Must mark all tasks in `docs/go_refactor/tasks/022-handler_slots.md` as done (`[x]`).

---

## File Structure

```
backend/
├── internal/
│   └── handler/
│       ├── slot.go                # [NEW] SlotService interface, SlotHandler struct, 5 endpoints, Routes
│       └── slot_test.go           # [NEW] Comprehensive httptest suite verifying all 5 endpoints, auth, and error states
docs/
├── plans/
│   ├── task.md                    # [MODIFY] Track Task 022 live checklist progress (table-only)
│   └── 2026-09-05-handler-slots.md # [NEW] Implementation plan document
└── go_refactor/
    └── tasks/
        └── 022-handler_slots.md   # [MODIFY] Mark all acceptance criteria and steps as completed
```

---

## Proposed Tasks

### Task 1: Git Branch Setup & Live Tracker Initialization

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Current repository state on main.
- Produces: `refactor/022-handler-slots` branch and updated `docs/plans/task.md` table.

- [ ] **Step 1: Create and checkout git branch**

```bash
git checkout -b refactor/022-handler-slots
```

- [ ] **Step 2: Initialize `docs/plans/task.md` with Task 022 tracker**

Update `docs/plans/task.md` to:
```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | IN_PROGRESS | Switch to branch `refactor/022-handler-slots` and initialize tracker |
| Task 2: Scaffold SlotService Interface, SlotHandler Struct, and Route Mounting | PENDING | Define `SlotService`, `SlotHandler`, `NewSlotHandler`, and `Routes(chi.Router)` |
| Task 3: Implement Slot Read Endpoints (`GetArcherCurrentSlot` and `GetSlot`) | PENDING | Implement `GetArcherCurrentSlot` and `GetSlot` with auth and ownership verification |
| Task 4: Implement Slot Mutation Endpoints (`JoinSession`, `ReJoinSession`, `LeaveSession`) | PENDING | Implement `JoinSession`, `ReJoinSession`, and `LeaveSession` with auth and error mapping |
| Task 5: End-to-End Suite Verification, Formatting & Linting | PENDING | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 6: Mark Tasks as Completed in Task Spec and Live Tracker | PENDING | Mark `022-handler_slots.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Verify git branch and commit tracker initialization**

```bash
git status
git add docs/plans/task.md
git commit -m "docs: initialize task tracker for task 022 handler slots"
```

---

### Task 2: Scaffold SlotService Interface, SlotHandler Struct, and Route Mounting

**Files:**
- Create: `backend/internal/handler/slot.go`
- Create: `backend/internal/handler/slot_test.go`

**Interfaces:**
- Consumes: `model.SlotJoinRequest`, `model.SlotJoinResponse`, `model.FullSlotInfo`, `github.com/google/uuid`, `github.com/go-chi/chi/v5`.
- Produces:
  - `handler.SlotService` interface:
    - `GetArcherCurrentSlot(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error)`
    - `GetSlot(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error)`
    - `JoinSession(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error)`
    - `ReJoinSession(ctx context.Context, slotID uuid.UUID, archerID uuid.UUID) (*model.SlotJoinResponse, error)`
    - `LeaveSession(ctx context.Context, slotID uuid.UUID, archerID uuid.UUID) error`
  - `handler.SlotHandler` struct with `NewSlotHandler(slotSvc SlotService) *SlotHandler`
  - `SlotHandler.Routes(r chi.Router)` mounting:
    - `r.Get("/archer/{archer_id}", h.GetArcherCurrentSlot)`
    - `r.Get("/{slot_id}", h.GetSlot)`
    - `r.Post("/", h.JoinSession)`
    - `r.Patch("/re-join/{slot_id}", h.ReJoinSession)`
    - `r.Patch("/leave/{slot_id}", h.LeaveSession)`

- [ ] **Step 1: Write failing test for router wiring and stubs in `slot_test.go`**

Create `backend/internal/handler/slot_test.go`:
```go
package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockSlotHandlerService struct {
	getArcherCurrentSlotFn func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error)
	getSlotFn              func(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error)
	joinSessionFn          func(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error)
	reJoinSessionFn        func(ctx context.Context, slotID uuid.UUID, archerID uuid.UUID) (*model.SlotJoinResponse, error)
	leaveSessionFn         func(ctx context.Context, slotID uuid.UUID, archerID uuid.UUID) error
}

func (m *mockSlotHandlerService) GetArcherCurrentSlot(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
	if m.getArcherCurrentSlotFn != nil {
		return m.getArcherCurrentSlotFn(ctx, archerID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSlotHandlerService) GetSlot(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error) {
	if m.getSlotFn != nil {
		return m.getSlotFn(ctx, slotID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSlotHandlerService) JoinSession(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
	if m.joinSessionFn != nil {
		return m.joinSessionFn(ctx, req)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSlotHandlerService) ReJoinSession(ctx context.Context, slotID uuid.UUID, archerID uuid.UUID) (*model.SlotJoinResponse, error) {
	if m.reJoinSessionFn != nil {
		return m.reJoinSessionFn(ctx, slotID, archerID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSlotHandlerService) LeaveSession(ctx context.Context, slotID uuid.UUID, archerID uuid.UUID) error {
	if m.leaveSessionFn != nil {
		return m.leaveSessionFn(ctx, slotID, archerID)
	}
	return errors.New("unimplemented")
}

func newSlotTestRequest(method, url string, body io.Reader, authArcherID *uuid.UUID, paramKey, paramVal string) *http.Request {
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

func sampleFullSlotInfo(slotID, archerID, sessionID, targetID uuid.UUID) *model.FullSlotInfo {
	return &model.FullSlotInfo{
		SlotID:          slotID,
		TargetID:        targetID,
		ArcherID:        archerID,
		SessionID:       sessionID,
		SlotLetter:      model.SlotLetterA,
		Lane:            1,
		Distance:        18,
		Slot:            "1A",
		FaceType:        model.FaceTypeWA40Full,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      32.5,
		ClubID:          nil,
		IsShooting:      true,
		ShotPerRound:    nil,
		IntervalSeconds: 15,
		CreatedAt:       time.Now().UTC(),
	}
}

func TestSlotHandler_Routes(t *testing.T) {
	mockSvc := &mockSlotHandlerService{
		getArcherCurrentSlotFn: func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
			return sampleFullSlotInfo(uuid.New(), archerID, uuid.New(), uuid.New()), nil
		},
	}
	h := handler.NewSlotHandler(mockSvc)

	r := chi.NewRouter()
	authArcherID := uuid.New()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req = req.WithContext(middleware.WithArcherID(req.Context(), authArcherID))
			next.ServeHTTP(w, req)
		})
	})
	r.Route("/api/v0/session/slot", func(sub chi.Router) {
		h.Routes(sub)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v0/session/slot/archer/"+authArcherID.String(), nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/handler/... -run TestSlotHandler_Routes -v
```
Expected: FAIL with undefined `handler.NewSlotHandler`.

- [ ] **Step 3: Create `backend/internal/handler/slot.go` with interface, struct, and method stubs**

Create `backend/internal/handler/slot.go`:
```go
package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// SlotService defines the business operations required by SlotHandler.
type SlotService interface {
	GetArcherCurrentSlot(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error)
	GetSlot(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error)
	JoinSession(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error)
	ReJoinSession(ctx context.Context, slotID uuid.UUID, archerID uuid.UUID) (*model.SlotJoinResponse, error)
	LeaveSession(ctx context.Context, slotID uuid.UUID, archerID uuid.UUID) error
}

// SlotHandler manages HTTP endpoints for session slot assignments.
type SlotHandler struct {
	slotSvc SlotService
}

// NewSlotHandler constructs a SlotHandler with service dependency injection.
func NewSlotHandler(slotSvc SlotService) *SlotHandler {
	return &SlotHandler{
		slotSvc: slotSvc,
	}
}

// Routes registers all slot management endpoints on the provided chi Router.
func (h *SlotHandler) Routes(r chi.Router) {
	r.Get("/archer/{archer_id}", h.GetArcherCurrentSlot)
	r.Get("/{slot_id}", h.GetSlot)
	r.Post("/", h.JoinSession)
	r.Patch("/re-join/{slot_id}", h.ReJoinSession)
	r.Patch("/leave/{slot_id}", h.LeaveSession)
}

// GetArcherCurrentSlot handles GET /api/v0/session/slot/archer/{archer_id}.
func (h *SlotHandler) GetArcherCurrentSlot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GetSlot handles GET /api/v0/session/slot/{slot_id}.
func (h *SlotHandler) GetSlot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// JoinSession handles POST /api/v0/session/slot.
func (h *SlotHandler) JoinSession(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// ReJoinSession handles PATCH /api/v0/session/slot/re-join/{slot_id}.
func (h *SlotHandler) ReJoinSession(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// LeaveSession handles PATCH /api/v0/session/slot/leave/{slot_id}.
func (h *SlotHandler) LeaveSession(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/handler/... -run TestSlotHandler_Routes -v
```
Expected: PASS (or 501 when calling stub; update stub in step 3 temporarily or implement read endpoints in Task 3).
To make `TestSlotHandler_Routes` pass in Step 3, `GetArcherCurrentSlot` can call `h.slotSvc.GetArcherCurrentSlot`:
```go
func (h *SlotHandler) GetArcherCurrentSlot(w http.ResponseWriter, r *http.Request) {
	archerIDStr := getURLParam(r, "archer_id")
	archerID, err := uuid.Parse(archerIDStr)
	if err != nil {
		writeAppError(w, err)
		return
	}
	info, err := h.slotSvc.GetArcherCurrentSlot(r.Context(), archerID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, info)
}
```
Run `cd backend && go test ./internal/handler/... -run TestSlotHandler_Routes -v` -> PASS.

- [ ] **Step 5: Commit scaffolded slot handler**

```bash
git add backend/internal/handler/slot.go backend/internal/handler/slot_test.go
git commit -m "feat(handler): scaffold SlotService interface, SlotHandler struct, and route mounting"
```

---

### Task 3: Implement Slot Read Endpoints (`GetArcherCurrentSlot` and `GetSlot`)

**Files:**
- Modify: `backend/internal/handler/slot.go`
- Modify: `backend/internal/handler/slot_test.go`

**Interfaces:**
- Consumes: `middleware.GetArcherID(r.Context())`, `slotSvc.GetArcherCurrentSlot`, `slotSvc.GetSlot`, `apperror.ErrNotFound`, `apperror.ErrForbidden`, `apperror.ErrValidation`.
- Produces:
  - `GetArcherCurrentSlot(w http.ResponseWriter, r *http.Request)`:
    - Returns 401 if unauthorized.
    - Returns 422 if `archer_id` is invalid UUID.
    - Returns 403 if `authArcherID != archerID`.
    - Returns 404 if `apperror.ErrNotFound`.
    - Returns 200 + `FullSlotInfo` if found.
  - `GetSlot(w http.ResponseWriter, r *http.Request)`:
    - Returns 401 if unauthorized.
    - Returns 422 if `slot_id` is invalid UUID.
    - Returns 404 if `apperror.ErrNotFound`.
    - Returns 403 if `info.ArcherID != authArcherID`.
    - Returns 200 + `FullSlotInfo` if found.

- [ ] **Step 1: Write failing tests for read endpoints in `slot_test.go`**

Add unit tests to `backend/internal/handler/slot_test.go`:
```go
func TestSlotHandler_GetArcherCurrentSlot(t *testing.T) {
	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		req := newSlotTestRequest(http.MethodGet, "/archer/"+uuid.NewString(), nil, nil, "archer_id", uuid.NewString())
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_archer_id_is_invalid", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodGet, "/archer/not-a-uuid", nil, &authID, "archer_id", "not-a-uuid")
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_403_when_authenticated_archer_does_not_match_requested_archer", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		targetArcherID := uuid.New()
		req := newSlotTestRequest(http.MethodGet, "/archer/"+targetArcherID.String(), nil, &authID, "archer_id", targetArcherID.String())
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_404_when_no_active_slot_found", func(t *testing.T) {
		authID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			getArcherCurrentSlotFn: func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/archer/"+authID.String(), nil, &authID, "archer_id", authID.String())
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_with_full_slot_info_when_found", func(t *testing.T) {
		authID := uuid.New()
		expected := sampleFullSlotInfo(uuid.New(), authID, uuid.New(), uuid.New())
		mockSvc := &mockSlotHandlerService{
			getArcherCurrentSlotFn: func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
				return expected, nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/archer/"+authID.String(), nil, &authID, "archer_id", authID.String())
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.FullSlotInfo
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if resp.SlotID != expected.SlotID || resp.Slot != expected.Slot {
			t.Fatalf("mismatched slot info: expected %+v, got %+v", expected, resp)
		}
	})
}

func TestSlotHandler_GetSlot(t *testing.T) {
	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		req := newSlotTestRequest(http.MethodGet, "/"+uuid.NewString(), nil, nil, "slot_id", uuid.NewString())
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_slot_id_is_invalid", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodGet, "/invalid-uuid", nil, &authID, "slot_id", "invalid-uuid")
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_404_when_slot_not_found", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			getSlotFn: func(ctx context.Context, id uuid.UUID) (*model.FullSlotInfo, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_403_when_authenticated_archer_does_not_own_slot", func(t *testing.T) {
		authID := uuid.New()
		slotOwnerID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			getSlotFn: func(ctx context.Context, id uuid.UUID) (*model.FullSlotInfo, error) {
				return sampleFullSlotInfo(slotID, slotOwnerID, uuid.New(), uuid.New()), nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_with_full_slot_info_when_owner_requests", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		expected := sampleFullSlotInfo(slotID, authID, uuid.New(), uuid.New())
		mockSvc := &mockSlotHandlerService{
			getSlotFn: func(ctx context.Context, id uuid.UUID) (*model.FullSlotInfo, error) {
				return expected, nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.FullSlotInfo
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SlotID != slotID || resp.ArcherID != authID {
			t.Fatalf("unexpected response payload: %+v", resp)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/handler/... -run "TestSlotHandler_Get" -v
```
Expected: FAIL on `TestSlotHandler_GetSlot` (501 returned by stub).

- [ ] **Step 3: Implement `GetArcherCurrentSlot` and `GetSlot` in `backend/internal/handler/slot.go`**

Update `backend/internal/handler/slot.go`:
```go
// GetArcherCurrentSlot handles GET /api/v0/session/slot/archer/{archer_id}.
// Returns active slot assignment (open session and is_shooting = true).
// Enforces that only the authenticated archer can query their current slot.
func (h *SlotHandler) GetArcherCurrentSlot(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	archerIDStr := getURLParam(r, "archer_id")
	if archerIDStr == "" {
		archerIDStr = getURLParam(r, "id")
	}

	archerID, err := uuid.Parse(archerIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer_id is required"))
		return
	}

	if authArcherID != archerID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	info, err := h.slotSvc.GetArcherCurrentSlot(r.Context(), archerID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, info)
}

// GetSlot handles GET /api/v0/session/slot/{slot_id}.
// Returns active slot assignment details.
// Enforces that only the authenticated archer owning the slot can retrieve it.
func (h *SlotHandler) GetSlot(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "slot")
	}
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid slot_id is required"))
		return
	}

	info, err := h.slotSvc.GetSlot(r.Context(), slotID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if info.ArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	_ = writeJSON(w, http.StatusOK, info)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/handler/... -run "TestSlotHandler_Get" -v
```
Expected: PASS for all `TestSlotHandler_GetArcherCurrentSlot` and `TestSlotHandler_GetSlot` cases.

- [ ] **Step 5: Commit read endpoints**

```bash
git add backend/internal/handler/slot.go backend/internal/handler/slot_test.go
git commit -m "feat(handler): implement slot read endpoints GetArcherCurrentSlot and GetSlot"
```

---

### Task 4: Implement Slot Mutation Endpoints (`JoinSession`, `ReJoinSession`, `LeaveSession`)

**Files:**
- Modify: `backend/internal/handler/slot.go`
- Modify: `backend/internal/handler/slot_test.go`

**Interfaces:**
- Consumes: `model.SlotJoinRequest`, `model.SlotJoinResponse`, `middleware.GetArcherID`, `slotSvc.JoinSession`, `slotSvc.ReJoinSession`, `slotSvc.LeaveSession`.
- Produces:
  - `JoinSession(w http.ResponseWriter, r *http.Request)`:
    - 401 if unauthenticated.
    - 422 if body invalid JSON or missing `archer_id`/`session_id`.
    - 403 if `req.ArcherID != authArcherID`.
    - Translates domain errors from service (400 target full, 409 conflict, 404/422 session closed).
    - 200 OK + `SlotJoinResponse` on success.
  - `ReJoinSession(w http.ResponseWriter, r *http.Request)`:
    - 401 if unauthenticated.
    - 422 if `slot_id` is invalid UUID.
    - Translates domain errors from service (403 forbidden, 404 not found, 422 already in session, 400 bad request).
    - 200 OK + `SlotJoinResponse` on success.
  - `LeaveSession(w http.ResponseWriter, r *http.Request)`:
    - 401 if unauthenticated.
    - 422 if `slot_id` is invalid UUID.
    - Translates domain errors from service (403 forbidden, 404 not found, 409 not participating, 400 bad request).
    - 200 OK on success.

- [ ] **Step 1: Write failing tests for mutation endpoints in `slot_test.go`**

Add unit tests to `backend/internal/handler/slot_test.go`:
```go
func TestSlotHandler_JoinSession(t *testing.T) {
	validJoinPayload := func(archerID, sessionID uuid.UUID) model.SlotJoinRequest {
		return model.SlotJoinRequest{
			ArcherID:        archerID,
			SessionID:       sessionID,
			Distance:        18,
			FaceType:        model.FaceTypeWA40Full,
			Bowstyle:        model.BowstyleRecurve,
			DrawWeight:      32.0,
			IsShooting:      true,
			IntervalSeconds: 15,
		}
	}

	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		payload := validJoinPayload(uuid.New(), uuid.New())
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), nil, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_body_is_invalid_JSON", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodPost, "/", strings.NewReader("{invalid"), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_required_ids_are_nil", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		payload := validJoinPayload(uuid.Nil, uuid.Nil)
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_403_when_authenticated_archer_does_not_match_payload_archer", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		otherID := uuid.New()
		payload := validJoinPayload(otherID, uuid.New())
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_409_when_service_returns_conflict", func(t *testing.T) {
		authID := uuid.New()
		sessionID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			joinSessionFn: func(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
				return nil, apperror.Wrap(apperror.ErrConflict, "archer already joined this session")
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		payload := validJoinPayload(authID, sessionID)
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_with_slot_join_response_when_valid", func(t *testing.T) {
		authID := uuid.New()
		sessionID := uuid.New()
		newSlotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			joinSessionFn: func(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
				return &model.SlotJoinResponse{
					SlotID: newSlotID,
					Slot:   "1A",
				}, nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		payload := validJoinPayload(authID, sessionID)
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.SlotJoinResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SlotID != newSlotID || resp.Slot != "1A" {
			t.Fatalf("unexpected join response: %+v", resp)
		}
	})
}

func TestSlotHandler_ReJoinSession(t *testing.T) {
	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		req := newSlotTestRequest(http.MethodPatch, "/re-join/"+uuid.NewString(), nil, nil, "slot_id", uuid.NewString())
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_slot_id_is_invalid", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodPatch, "/re-join/not-uuid", nil, &authID, "slot_id", "not-uuid")
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_404_when_slot_not_found", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			reJoinSessionFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.SlotJoinResponse, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/re-join/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_403_when_user_forbidden_to_rejoin", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			reJoinSessionFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.SlotJoinResponse, error) {
				return nil, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/re-join/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_with_slot_join_response_when_rejoined", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			reJoinSessionFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.SlotJoinResponse, error) {
				return &model.SlotJoinResponse{
					SlotID: slotID,
					Slot:   "2B",
				}, nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/re-join/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.SlotJoinResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SlotID != slotID || resp.Slot != "2B" {
			t.Fatalf("unexpected rejoin response: %+v", resp)
		}
	})
}

func TestSlotHandler_LeaveSession(t *testing.T) {
	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		req := newSlotTestRequest(http.MethodPatch, "/leave/"+uuid.NewString(), nil, nil, "slot_id", uuid.NewString())
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_slot_id_is_invalid", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodPatch, "/leave/not-uuid", nil, &authID, "slot_id", "not-uuid")
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_404_when_slot_not_found", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			leaveSessionFn: func(ctx context.Context, sID, aID uuid.UUID) error {
				return apperror.ErrNotFound
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/leave/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_409_when_archer_is_not_participating", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			leaveSessionFn: func(ctx context.Context, sID, aID uuid.UUID) error {
				return apperror.Wrap(apperror.ErrConflict, "archer is not participating in this session")
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/leave/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_when_leave_succeeds", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			leaveSessionFn: func(ctx context.Context, sID, aID uuid.UUID) error {
				return nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/leave/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/handler/... -run "TestSlotHandler_Join|TestSlotHandler_ReJoin|TestSlotHandler_Leave" -v
```
Expected: FAIL (501 returned by stubs).

- [ ] **Step 3: Implement `JoinSession`, `ReJoinSession`, and `LeaveSession` in `slot.go`**

Update `backend/internal/handler/slot.go`:
```go
// JoinSession handles POST /api/v0/session/slot.
// Assigns an archer to a target slot within an open session.
// Enforces that only the authenticated archer can join for themselves.
func (h *SlotHandler) JoinSession(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	var req model.SlotJoinRequest
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.ArcherID == uuid.Nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "archer_id is required"))
		return
	}
	if req.SessionID == uuid.Nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "session_id is required"))
		return
	}

	if req.ArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	resp, err := h.slotSvc.JoinSession(r.Context(), req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, resp)
}

// ReJoinSession handles PATCH /api/v0/session/slot/re-join/{slot_id}.
// Re-activates a previously inactive slot assignment.
func (h *SlotHandler) ReJoinSession(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "slot")
	}
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid slot_id is required"))
		return
	}

	resp, err := h.slotSvc.ReJoinSession(r.Context(), slotID, authArcherID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, resp)
}

// LeaveSession handles PATCH /api/v0/session/slot/leave/{slot_id}.
// Deactivates an active slot assignment (leave the session).
func (h *SlotHandler) LeaveSession(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "slot")
	}
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid slot_id is required"))
		return
	}

	if err := h.slotSvc.LeaveSession(r.Context(), slotID, authArcherID); err != nil {
		writeAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/handler/... -run "TestSlotHandler" -v
```
Expected: PASS for all tests.

- [ ] **Step 5: Commit mutation endpoints**

```bash
git add backend/internal/handler/slot.go backend/internal/handler/slot_test.go
git commit -m "feat(handler): implement slot mutation endpoints JoinSession, ReJoinSession, and LeaveSession"
```

---

### Task 5: End-to-End Suite Verification, Formatting & Linting

**Files:**
- Modify: `backend/internal/handler/slot.go` (if formatting needed)
- Modify: `backend/internal/handler/slot_test.go` (if formatting needed)

**Interfaces:**
- Consumes: All tests in `./internal/handler/...` and full Go backend suite.
- Produces: 100% passing tests with race detector, clean `go vet`, clean `go build`, and clean `golangci-lint` / `gofumpt`.

- [ ] **Step 1: Run handler unit tests with race detection**

```bash
cd backend && go test -race -v ./internal/handler/...
```
Expected: PASS across all tests in `internal/handler`.

- [ ] **Step 2: Run compiler verification and vet checks**

```bash
cd backend && go vet ./... && go build ./...
```
Expected: Clean output, exit code 0.

- [ ] **Step 3: Run repository Go linter and formatter**

```bash
./scripts/linting.bash --go
```
Expected: Format with `gofumpt` and pass `golangci-lint` with 0 warnings/errors.

- [ ] **Step 4: Commit any formatting or lint fixes**

```bash
git add -A
git commit -m "style: apply gofumpt and lint checks to slot handler" || true
```

---

### Task 6: Mark Tasks as Completed in Task Spec and Live Tracker

**Files:**
- Modify: `docs/go_refactor/tasks/022-handler_slots.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified test execution and build artifacts.
- Produces: Marked checkboxes in `docs/go_refactor/tasks/022-handler_slots.md` and `DONE` states in `docs/plans/task.md`.

- [ ] **Step 1: Check off all items in `docs/go_refactor/tasks/022-handler_slots.md`**

Mark all acceptance criteria and steps in `docs/go_refactor/tasks/022-handler_slots.md` with `[x]`:
```markdown
## Acceptance Criteria

- [x] `backend/internal/handler/slot.go` implements `SlotHandler` with methods:
    - `GetArcherCurrentSlot(w, r)` — GET `/api/v0/session/slot/archer/{archer_id}`
    - `GetSlot(w, r)` — GET `/api/v0/session/slot/{slot_id}`
    - `JoinSession(w, r)` — POST `/api/v0/session/slot`
    - `ReJoinSession(w, r)` — PATCH `/api/v0/session/slot/re-join/{slot_id}`
    - `LeaveSession(w, r)` — PATCH `/api/v0/session/slot/leave/{slot_id}`
- [x] All endpoints extract the authenticated archer ID from request context.
- [x] Handler delegates business logic to `SlotService`.
- [x] Error responses: 400 (bad request), 403 (forbidden), 404 (not found), 422 (validation).
- [x] Unit tests using `httptest` with mock service verify:
    - GetArcherCurrentSlot returns 200 + full slot info
    - JoinSession with valid payload returns 200 + slot join response
    - LeaveSession returns 200
    - GetSlot with non-existent ID returns 404
- [x] `go test ./internal/handler/...` passes.
- [x] `go vet ./...` reports no issues.

## Steps

- [x] **Step 1: Write failing tests**
- [x] **Step 2: Run tests to verify they fail**
- [x] **Step 3: Implement `slot.go`**
- [x] **Step 4: Run tests to verify they pass**
- [x] **Step 5: Run go vet and build**
- [x] **Step 6: Commit**
```

- [ ] **Step 2: Update `docs/plans/task.md` to indicate all tasks are DONE**

Update `docs/plans/task.md`:
```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | DONE | Switch to branch `refactor/022-handler-slots` and initialize tracker |
| Task 2: Scaffold SlotService Interface, SlotHandler Struct, and Route Mounting | DONE | Define `SlotService`, `SlotHandler`, `NewSlotHandler`, and `Routes(chi.Router)` |
| Task 3: Implement Slot Read Endpoints (`GetArcherCurrentSlot` and `GetSlot`) | DONE | Implement `GetArcherCurrentSlot` and `GetSlot` with auth and ownership verification |
| Task 4: Implement Slot Mutation Endpoints (`JoinSession`, `ReJoinSession`, `LeaveSession`) | DONE | Implement `JoinSession`, `ReJoinSession`, and `LeaveSession` with auth and error mapping |
| Task 5: End-to-End Suite Verification, Formatting & Linting | DONE | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 6: Mark Tasks as Completed in Task Spec and Live Tracker | DONE | Mark `022-handler_slots.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Commit documentation checklist updates**

```bash
git add docs/go_refactor/tasks/022-handler_slots.md docs/plans/task.md
git commit -m "docs: mark task 022 handler slots checklist as completed"
```

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-09-05-handler-slots.md`. Two execution options:

1. **Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. **Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
