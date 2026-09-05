# Task 023: Build HTTP Handler — Shots Endpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the shots HTTP handler in `backend/internal/handler/shot.go` and comprehensive test suite in `backend/internal/handler/shot_test.go`, porting all 3 endpoints from Python `shot_router.py` (`Create` supporting single/batch union payload, `GetBySlot`, and `CountBySlot`). All endpoints enforce authentication context extraction, batch validation, error mapping, and chi route registration. Upon completion, mark all tasks in `docs/go_refactor/tasks/023-handler_shots.md` and `docs/plans/task.md` as done.

**Architecture:**
- `backend/internal/handler/shot.go`: Implements `ShotHandler` exposing `Create`, `GetBySlot`, `CountBySlot`, and `Routes(r chi.Router)`.
- `ShotService` interface: Defined in `package handler` specifying domain operations required by `ShotHandler`:
  - `Create(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error)`
  - `CreateBatch(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error)`
  - `GetBySlot(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error)`
  - `CountBySlot(ctx context.Context, slotID, archerID uuid.UUID) (int, error)`
- Payload Union Handling: In `Create`, peeks at first non-whitespace byte in request body:
  - If `[`, unmarshals as `[]model.ShotCreate`, validates batch length (3 <= len <= 10) and uniform slot IDs, invokes `CreateBatch`, and returns 201 with `[]model.ShotID`.
  - Otherwise, unmarshals as `model.ShotCreate`, invokes `Create`, and returns 201 with `model.ShotID`.
- Authentication & Authorization: All endpoints extract authenticated archer ID via `middleware.GetArcherID(r.Context())`. Returns 401 if missing.
- Error Translation: Domain errors (`apperror.ErrNotFound` -> 404, `apperror.ErrForbidden` -> 403, `apperror.ErrValidation` -> 422) mapped via `writeAppError(w, err)`. Bad input (batch size <3 or >10, mismatched slot IDs) mapped via `writeError(w, http.StatusBadRequest, msg)`.
- Acceptance & Task Tracking: Update `docs/go_refactor/tasks/023-handler_shots.md` and `docs/plans/task.md` marking all items completed.

**Tech Stack:** Go 1.24+, `github.com/go-chi/chi/v5`, `net/http`, `net/http/httptest`, `encoding/json`, `github.com/google/uuid`, internal packages (`apperror`, `middleware`, `model`).

**Spec:**
- [docs/go_refactor/tasks/023-handler_shots.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/023-handler_shots.md)
- [backend-old/src/routers/v0/shot_router.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/routers/v0/shot_router.py)
- [backend-old/src/core/shot_manager.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/core/shot_manager.py)
- [backend-old/tests/endpoints/test_shot_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend-old/tests/endpoints/test_shot_endpoints.py)

## Global Constraints

- Git branch: `refactor/023-handler-shots`
- Package: `package handler` (`github.com/jpmolinamatute/arch-stats/backend/internal/handler`)
- Endpoints:
  - `POST /api/v0/shot` -> `ShotHandler.Create` (returns 201 + `ShotID` or `[]ShotID`)
  - `GET /api/v0/shot/by-slot/{slot_id}` -> `ShotHandler.GetBySlot` (returns 200 + `[]ShotRead`)
  - `GET /api/v0/shot/count-by-slot/{slot_id}` -> `ShotHandler.CountBySlot` (returns 200 + integer count)
- Batch constraints: Batch size between 3 and 10; all shots in batch must have identical `slot_id`. Violation returns HTTP 400 Bad Request.
- Error handling: Wrap internal errors with `%w` or `apperror.Wrap`. Translate domain errors using `writeAppError(w, err)` and HTTP error envelopes using `writeError(w, status, msg)`.
- Linting & Formatting: Code must pass `./scripts/linting.bash --go` (`gofumpt` and `golangci-lint run ./...`).
- Verification: `go test -race -v ./internal/handler/...`, `go vet ./...`, and `go build ./...` must succeed.
- Task completion: Must mark all tasks in `docs/go_refactor/tasks/023-handler_shots.md` as done (`[x]`).

---

## File Structure

```
backend/
├── internal/
│   └── handler/
│       ├── shot.go                # [NEW] ShotService interface, ShotHandler struct, 3 endpoints, Routes
│       └── shot_test.go           # [NEW] Comprehensive httptest suite verifying single/batch create, get, count, auth, and errors
docs/
├── plans/
│   ├── task.md                    # [MODIFY] Track Task 023 live checklist progress (table-only)
│   └── 2026-09-05-handler-shots.md # [NEW] Implementation plan document
└── go_refactor/
    └── tasks/
        └── 023-handler_shots.md   # [MODIFY] Mark all acceptance criteria and steps as completed
```

---

## Proposed Tasks

### Task 1: Git Branch Setup & Live Tracker Initialization

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Clean working tree on `main`.
- Produces: `refactor/023-handler-shots` branch and updated `docs/plans/task.md` table.

- [ ] **Step 1: Create and checkout git branch**

```bash
git checkout -b refactor/023-handler-shots
```

- [ ] **Step 2: Initialize `docs/plans/task.md` with Task 023 tracker**

Update `docs/plans/task.md` to:
```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | IN_PROGRESS | Switch to branch `refactor/023-handler-shots` and initialize tracker |
| Task 2: Write Failing Tests for ShotHandler Interface and Endpoints | PENDING | Scaffold `shot_test.go` with mock service and tests for single/batch create, list, and count |
| Task 3: Implement ShotHandler (`shot.go`) | PENDING | Implement `ShotService` interface, `ShotHandler`, single/batch `Create`, `GetBySlot`, `CountBySlot`, and `Routes` |
| Task 4: End-to-End Suite Verification, Formatting & Linting | PENDING | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 5: Mark Tasks as Completed in Task Spec and Live Tracker | PENDING | Mark `023-handler_shots.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Verify git branch and commit tracker initialization**

```bash
git status
git add docs/plans/task.md
git commit -m "docs: initialize task tracker for task 023 handler shots"
```

---

### Task 2: Write Failing Tests for ShotHandler Interface and Endpoints

**Files:**
- Create: `backend/internal/handler/shot_test.go`

**Interfaces:**
- Consumes: `model.ShotCreate`, `model.ShotRead`, `model.ShotID`, `github.com/google/uuid`, `github.com/go-chi/chi/v5`.
- Produces: `mockShotHandlerService` and test functions:
  - `TestShotHandler_Routes`
  - `TestShotHandler_Create` (unauthenticated, empty body, invalid JSON, single create success/errors, batch create success/errors)
  - `TestShotHandler_GetBySlot` (unauthenticated, invalid UUID, not found, forbidden, empty list, populated list, internal error)
  - `TestShotHandler_CountBySlot` (unauthenticated, invalid UUID, not found, forbidden, 0 count, N count, internal error)

- [ ] **Step 1: Write `backend/internal/handler/shot_test.go`**

```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockShotHandlerService struct {
	createFn      func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error)
	createBatchFn func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error)
	getBySlotFn   func(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error)
	countBySlotFn func(ctx context.Context, slotID, archerID uuid.UUID) (int, error)
}

func (m *mockShotHandlerService) Create(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, shot, archerID)
	}
	return uuid.Nil, errors.New("unimplemented")
}

func (m *mockShotHandlerService) CreateBatch(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
	if m.createBatchFn != nil {
		return m.createBatchFn(ctx, shots, archerID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockShotHandlerService) GetBySlot(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error) {
	if m.getBySlotFn != nil {
		return m.getBySlotFn(ctx, slotID, archerID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockShotHandlerService) CountBySlot(ctx context.Context, slotID, archerID uuid.UUID) (int, error) {
	if m.countBySlotFn != nil {
		return m.countBySlotFn(ctx, slotID, archerID)
	}
	return 0, errors.New("unimplemented")
}

func newShotTestRequest(method, url string, body io.Reader, authArcherID *uuid.UUID, paramKey, paramVal string) *http.Request {
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

func sampleShotRead(shotID, slotID uuid.UUID, score int) model.ShotRead {
	x := 5.0
	y := 5.0
	return model.ShotRead{
		ShotID:    shotID,
		SlotID:    slotID,
		X:         &x,
		Y:         &y,
		IsX:       score == 10,
		Score:     &score,
		ArrowID:   nil,
		CreatedAt: time.Now().UTC(),
	}
}

func TestShotHandler_Routes(t *testing.T) {
	svc := &mockShotHandlerService{}
	h := handler.NewShotHandler(svc)

	r := chi.NewRouter()
	r.Route("/api/v0/shot", func(sub chi.Router) {
		h.Routes(sub)
	})

	slotID := uuid.New()
	routes := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v0/shot"},
		{method: http.MethodGet, path: "/api/v0/shot/by-slot/" + slotID.String()},
		{method: http.MethodGet, path: "/api/v0/shot/count-by-slot/" + slotID.String()},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			req := httptest.NewRequest(rt.method, rt.path, nil)
			rctx := chi.NewRouteContext()
			if !r.Match(rctx, rt.method, rt.path) {
				t.Fatalf("expected route %s %s to match router", rt.method, rt.path)
			}
		})
	}
}

func TestShotHandler_Create(t *testing.T) {
	authArcherID := uuid.New()
	slotID := uuid.New()
	score := 9
	xVal := 2.5
	yVal := 3.5

	singlePayload := model.ShotCreate{
		SlotID: slotID,
		X:      &xVal,
		Y:      &yVal,
		Score:  &score,
	}

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), nil, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("returns 422 when body is empty", func(t *testing.T) {
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader([]byte("")), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("returns 422 when body is invalid JSON", func(t *testing.T) {
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader([]byte("{invalid-json")), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("single: returns 201 with ShotID on success", func(t *testing.T) {
		newShotID := uuid.New()
		svc := &mockShotHandlerService{
			createFn: func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
				if archerID != authArcherID {
					t.Errorf("expected archerID %s, got %s", authArcherID, archerID)
				}
				if shot.SlotID != slotID {
					t.Errorf("expected slotID %s, got %s", slotID, shot.SlotID)
				}
				return newShotID, nil
			},
		}

		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.ShotID
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ShotID != newShotID {
			t.Fatalf("expected shotID %s, got %s", newShotID, resp.ShotID)
		}
	})

	t.Run("single: returns 404 when slot does not exist", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createFn: func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
				return uuid.Nil, apperror.ErrNotFound
			},
		}

		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("single: returns 403 when archer does not own slot", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createFn: func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
				return uuid.Nil, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}

		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("single: returns 422 when validation fails", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createFn: func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
				return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "score must be between 0 and 10")
			},
		}

		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("batch: returns 400 when empty array", func(t *testing.T) {
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader([]byte("[]")), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
		var errResp middleware.ErrorResponse
		_ = json.NewDecoder(rr.Body).Decode(&errResp)
		if errResp.Detail != "Invalid input" {
			t.Fatalf("expected detail 'Invalid input', got %q", errResp.Detail)
		}
	})

	t.Run("batch: returns 400 when fewer than 3 shots", func(t *testing.T) {
		twoShots := []model.ShotCreate{singlePayload, singlePayload}
		body, _ := json.Marshal(twoShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
		var errResp middleware.ErrorResponse
		_ = json.NewDecoder(rr.Body).Decode(&errResp)
		if errResp.Detail != "Invalid input" {
			t.Fatalf("expected detail 'Invalid input', got %q", errResp.Detail)
		}
	})

	t.Run("batch: returns 400 when more than 10 shots", func(t *testing.T) {
		elevenShots := make([]model.ShotCreate, 11)
		for i := range elevenShots {
			elevenShots[i] = singlePayload
		}
		body, _ := json.Marshal(elevenShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
		var errResp middleware.ErrorResponse
		_ = json.NewDecoder(rr.Body).Decode(&errResp)
		if errResp.Detail != "Invalid input" {
			t.Fatalf("expected detail 'Invalid input', got %q", errResp.Detail)
		}
	})

	t.Run("batch: returns 400 when shots belong to different slots", func(t *testing.T) {
		otherSlotID := uuid.New()
		diffSlotShots := []model.ShotCreate{
			singlePayload,
			singlePayload,
			{SlotID: otherSlotID, X: &xVal, Y: &yVal, Score: &score},
		}
		body, _ := json.Marshal(diffSlotShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
		var errResp middleware.ErrorResponse
		_ = json.NewDecoder(rr.Body).Decode(&errResp)
		if errResp.Detail != "All shots must belong to the same slot" {
			t.Fatalf("expected detail 'All shots must belong to the same slot', got %q", errResp.Detail)
		}
	})

	t.Run("batch: returns 201 with array of ShotID on success", func(t *testing.T) {
		id1, id2, id3 := uuid.New(), uuid.New(), uuid.New()
		svc := &mockShotHandlerService{
			createBatchFn: func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
				if archerID != authArcherID {
					t.Errorf("expected archerID %s, got %s", authArcherID, archerID)
				}
				if len(shots) != 3 {
					t.Errorf("expected 3 shots, got %d", len(shots))
				}
				return []uuid.UUID{id1, id2, id3}, nil
			},
		}

		threeShots := []model.ShotCreate{singlePayload, singlePayload, singlePayload}
		body, _ := json.Marshal(threeShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp []model.ShotID
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp) != 3 {
			t.Fatalf("expected 3 items in response, got %d", len(resp))
		}
		if resp[0].ShotID != id1 || resp[1].ShotID != id2 || resp[2].ShotID != id3 {
			t.Fatalf("unexpected shot IDs returned")
		}
	})

	t.Run("batch: returns 403 when forbidden", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createBatchFn: func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
				return nil, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}

		threeShots := []model.ShotCreate{singlePayload, singlePayload, singlePayload}
		body, _ := json.Marshal(threeShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("batch: returns 404 when slot not found", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createBatchFn: func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
				return nil, apperror.ErrNotFound
			},
		}

		threeShots := []model.ShotCreate{singlePayload, singlePayload, singlePayload}
		body, _ := json.Marshal(threeShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})
}

func TestShotHandler_GetBySlot(t *testing.T) {
	authArcherID := uuid.New()
	slotID := uuid.New()

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, nil, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("returns 422 when slot_id is invalid", func(t *testing.T) {
		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/invalid-uuid", nil, &authArcherID, "slot_id", "invalid-uuid")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("returns 404 when slot not found", func(t *testing.T) {
		svc := &mockShotHandlerService{
			getBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) ([]model.ShotRead, error) {
				return nil, apperror.ErrNotFound
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("returns 403 when archer does not own slot", func(t *testing.T) {
		svc := &mockShotHandlerService{
			getBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) ([]model.ShotRead, error) {
				return nil, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("returns 200 with empty array when no shots recorded", func(t *testing.T) {
		svc := &mockShotHandlerService{
			getBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) ([]model.ShotRead, error) {
				return nil, nil
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp []model.ShotRead
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}
		if resp == nil || len(resp) != 0 {
			t.Fatalf("expected non-nil empty slice, got %v", resp)
		}
	})

	t.Run("returns 200 with shots array on success", func(t *testing.T) {
		s1 := sampleShotRead(uuid.New(), slotID, 10)
		s2 := sampleShotRead(uuid.New(), slotID, 9)
		svc := &mockShotHandlerService{
			getBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) ([]model.ShotRead, error) {
				return []model.ShotRead{s1, s2}, nil
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp []model.ShotRead
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("expected 2 shots, got %d", len(resp))
		}
		if resp[0].ShotID != s1.ShotID || resp[1].ShotID != s2.ShotID {
			t.Fatalf("unexpected shots returned")
		}
	})
}

func TestShotHandler_CountBySlot(t *testing.T) {
	authArcherID := uuid.New()
	slotID := uuid.New()

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, nil, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("returns 422 when slot_id is invalid", func(t *testing.T) {
		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/invalid-uuid", nil, &authArcherID, "slot_id", "invalid-uuid")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("returns 404 when slot not found", func(t *testing.T) {
		svc := &mockShotHandlerService{
			countBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) (int, error) {
				return 0, apperror.ErrNotFound
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("returns 403 when archer does not own slot", func(t *testing.T) {
		svc := &mockShotHandlerService{
			countBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) (int, error) {
				return 0, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("returns 200 with 0 count when slot has no shots", func(t *testing.T) {
		svc := &mockShotHandlerService{
			countBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) (int, error) {
				return 0, nil
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var count int
		if err := json.NewDecoder(rr.Body).Decode(&count); err != nil {
			t.Fatalf("failed to decode integer response: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected count 0, got %d", count)
		}
	})

	t.Run("returns 200 with count when slot has shots", func(t *testing.T) {
		svc := &mockShotHandlerService{
			countBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) (int, error) {
				return 6, nil
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var count int
		if err := json.NewDecoder(rr.Body).Decode(&count); err != nil {
			t.Fatalf("failed to decode integer response: %v", err)
		}
		if count != 6 {
			t.Fatalf("expected count 6, got %d", count)
		}
	})
}
```

- [ ] **Step 2: Run tests to verify compilation failure before implementation**

```bash
cd backend && go test ./internal/handler/... -v
```
Expected output: compilation failure due to undefined `handler.NewShotHandler`.

- [ ] **Step 3: Commit failing test suite**

```bash
git add backend/internal/handler/shot_test.go
git commit -m "test(handler): add failing tests for shot endpoints"
```

---

### Task 3: Implement ShotHandler (`shot.go`)

**Files:**
- Create: `backend/internal/handler/shot.go`

**Interfaces:**
- Consumes: `model.ShotCreate`, `model.ShotRead`, `model.ShotID`, `apperror`, `middleware.GetArcherID`, `chi.Router`.
- Produces:
  - `ShotService` interface:
    - `Create(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error)`
    - `CreateBatch(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error)`
    - `GetBySlot(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error)`
    - `CountBySlot(ctx context.Context, slotID, archerID uuid.UUID) (int, error)`
  - `ShotHandler` struct with `NewShotHandler(shotSvc ShotService) *ShotHandler`
  - `ShotHandler.Routes(r chi.Router)`:
    - `r.Post("/", h.Create)`
    - `r.Get("/by-slot/{slot_id}", h.GetBySlot)`
    - `r.Get("/count-by-slot/{slot_id}", h.CountBySlot)`
  - `Create(w http.ResponseWriter, r *http.Request)`
  - `GetBySlot(w http.ResponseWriter, r *http.Request)`
  - `CountBySlot(w http.ResponseWriter, r *http.Request)`

- [ ] **Step 1: Write `backend/internal/handler/shot.go`**

```go
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ShotService defines the domain operations required by ShotHandler.
type ShotService interface {
	Create(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error)
	CreateBatch(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error)
	GetBySlot(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error)
	CountBySlot(ctx context.Context, slotID, archerID uuid.UUID) (int, error)
}

// ShotHandler manages HTTP endpoints for shot records.
type ShotHandler struct {
	shotSvc ShotService
}

// NewShotHandler constructs a ShotHandler with service dependency injection.
func NewShotHandler(shotSvc ShotService) *ShotHandler {
	return &ShotHandler{
		shotSvc: shotSvc,
	}
}

// Routes registers all shot management endpoints on the provided chi Router.
func (h *ShotHandler) Routes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/by-slot/{slot_id}", h.GetBySlot)
	r.Get("/count-by-slot/{slot_id}", h.CountBySlot)
}

// Create handles POST /api/v0/shot.
// Supports both a single ShotCreate object and a batch []ShotCreate array.
// Returns HTTP 201 Created with model.ShotID or []model.ShotID.
func (h *ShotHandler) Create(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	if r.Body == nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "request body is empty"))
		return
	}
	defer r.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "failed to read request body"))
		return
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "request body is empty"))
		return
	}

	if trimmed[0] == '[' {
		h.handleBatchCreate(w, r.Context(), trimmed, authArcherID)
		return
	}

	h.handleSingleCreate(w, r.Context(), trimmed, authArcherID)
}

func (h *ShotHandler) handleSingleCreate(w http.ResponseWriter, ctx context.Context, raw []byte, authArcherID uuid.UUID) {
	var shot model.ShotCreate
	if err := json.Unmarshal(raw, &shot); err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "invalid request body: "+err.Error()))
		return
	}

	id, err := h.shotSvc.Create(ctx, shot, authArcherID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusCreated, model.ShotID{ShotID: id})
}

func (h *ShotHandler) handleBatchCreate(w http.ResponseWriter, ctx context.Context, raw []byte, authArcherID uuid.UUID) {
	var shots []model.ShotCreate
	if err := json.Unmarshal(raw, &shots); err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "invalid request body: "+err.Error()))
		return
	}

	if len(shots) < 3 || len(shots) > 10 {
		writeError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	slotID := shots[0].SlotID
	for _, s := range shots[1:] {
		if s.SlotID != slotID {
			writeError(w, http.StatusBadRequest, "All shots must belong to the same slot")
			return
		}
	}

	ids, err := h.shotSvc.CreateBatch(ctx, shots, authArcherID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	resp := make([]model.ShotID, len(ids))
	for i, id := range ids {
		resp[i] = model.ShotID{ShotID: id}
	}

	_ = writeJSON(w, http.StatusCreated, resp)
}

// GetBySlot handles GET /api/v0/shot/by-slot/{slot_id}.
// Returns list of shots for the given slot owned by the authenticated archer.
func (h *ShotHandler) GetBySlot(w http.ResponseWriter, r *http.Request) {
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

	shots, err := h.shotSvc.GetBySlot(r.Context(), slotID, authArcherID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if shots == nil {
		shots = []model.ShotRead{}
	}

	_ = writeJSON(w, http.StatusOK, shots)
}

// CountBySlot handles GET /api/v0/shot/count-by-slot/{slot_id}.
// Returns total shot count for the given slot owned by the authenticated archer.
func (h *ShotHandler) CountBySlot(w http.ResponseWriter, r *http.Request) {
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

	count, err := h.shotSvc.CountBySlot(r.Context(), slotID, authArcherID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, count)
}
```

- [ ] **Step 2: Run tests to verify all tests pass**

```bash
cd backend && go test ./internal/handler/... -v
```
Expected output: `PASS` across all handler tests.

- [ ] **Step 3: Commit implementation**

```bash
git add backend/internal/handler/shot.go
git commit -m "feat(handler): implement shot endpoints with single/batch create, get, and count"
```

---

### Task 4: End-to-End Suite Verification, Formatting & Linting

**Files:**
- None (verification on all touched and existing files)

**Interfaces:**
- Consumes: `backend/internal/handler/...` and entire backend module.
- Produces: Clean test execution with race detector, clean lint report, clean compilation.

- [ ] **Step 1: Run handler package tests with race detector**

```bash
cd backend && go test -race -v ./internal/handler/...
```
Expected: `PASS`.

- [ ] **Step 2: Run `go vet ./...`**

```bash
cd backend && go vet ./...
```
Expected: clean exit, no errors.

- [ ] **Step 3: Run full linting and formatting suite**

```bash
./scripts/linting.bash --go
```
Expected: `gofumpt` clean, `golangci-lint` 0 issues, all unit tests pass.

- [ ] **Step 4: Run `go build ./...`**

```bash
cd backend && go build ./...
```
Expected: successful build.

---

### Task 5: Mark Tasks as Completed in Task Spec and Live Tracker

**Files:**
- Modify: `docs/go_refactor/tasks/023-handler_shots.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verification results from Task 4.
- Produces: Marked checkboxes in `023-handler_shots.md` and updated `task.md` table marked `DONE`.

- [ ] **Step 1: Mark all acceptance criteria and steps in `docs/go_refactor/tasks/023-handler_shots.md`**

Update `docs/go_refactor/tasks/023-handler_shots.md` checkboxes from `[ ]` to `[x]`:
```markdown
## Acceptance Criteria

- [x] `backend/internal/handler/shot.go` implements `ShotHandler` with methods:
    - `Create(w, r)` — POST `/api/v0/shot` — accepts single shot or array of shots, returns 201
    - `GetBySlot(w, r)` — GET `/api/v0/shot/by-slot/{slot_id}` — list shots for a slot
    - `CountBySlot(w, r)` — GET `/api/v0/shot/count-by-slot/{slot_id}` — count shots in a slot
- [x] The `Create` endpoint handles both `ShotCreate` and `[]ShotCreate` payloads (matching
  the Python union type `ShotCreate | list[ShotCreate]`).
- [x] All endpoints extract authenticated archer ID from request context.
- [x] Unit tests using `httptest` with mock service verify:
    - Create single shot returns 201 + shot ID
    - Create batch returns 201 + array of shot IDs
    - GetBySlot returns 200 + array of shots
    - CountBySlot returns 200 + integer count
- [x] `go test ./internal/handler/...` passes.
- [x] `go vet ./...` reports no issues.
```
and steps:
```markdown
## Steps

- [x] **Step 1: Write failing tests**
- [x] **Step 2: Run tests to verify they fail**
- [x] **Step 3: Implement `shot.go`**
- [x] **Step 4: Run tests to verify they pass**
- [x] **Step 5: Run go vet and build**
- [x] **Step 6: Commit**
```

- [ ] **Step 2: Update `docs/plans/task.md` to reflect full completion**

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | DONE | Switch to branch `refactor/023-handler-shots` and initialize tracker |
| Task 2: Write Failing Tests for ShotHandler Interface and Endpoints | DONE | Scaffold `shot_test.go` with mock service and tests for single/batch create, list, and count |
| Task 3: Implement ShotHandler (`shot.go`) | DONE | Implement `ShotService` interface, `ShotHandler`, single/batch `Create`, `GetBySlot`, `CountBySlot`, and `Routes` |
| Task 4: End-to-End Suite Verification, Formatting & Linting | DONE | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 5: Mark Tasks as Completed in Task Spec and Live Tracker | DONE | Mark `023-handler_shots.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Commit documentation updates**

```bash
git add docs/go_refactor/tasks/023-handler_shots.md docs/plans/task.md
git commit -m "docs: mark task 023 handler shots checklist as completed"
```
