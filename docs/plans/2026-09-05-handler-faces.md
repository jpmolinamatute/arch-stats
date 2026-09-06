# Task 024: Build HTTP Handler — Faces Endpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the public faces HTTP handler in `backend/internal/handler/face.go` and comprehensive test suite in `backend/internal/handler/face_test.go`, porting the 2 read-only endpoints from Python `faces_router.py` (`ListFaces` returning face summaries and `GetFace` returning full face definitions). Neither endpoint requires authentication. Upon completion, mark all tasks in `docs/go_refactor/tasks/024-handler_faces.md` and `docs/plans/task.md` as done.

**Architecture:**
- `backend/internal/handler/face.go`: Implements `FaceHandler` exposing `ListFaces`, `GetFace`, and `Routes(r chi.Router)`.
- `FaceService` interface: Defined in `package handler` specifying domain operations required by `FaceHandler`:
  - `ListAll(ctx context.Context) ([]model.FaceRead, error)`
  - `GetByID(ctx context.Context, id string) (*model.FaceRead, error)`
- Public Access: Both endpoints are public and do not pass through authentication middleware or query archer identity context.
- Error Translation: Domain errors (`apperror.ErrNotFound` -> 404, `apperror.ErrValidation` -> 422) mapped via `writeAppError(w, err)`. Unknown face type returns 404 with standard JSON error response.
- Serialization:
  - `ListFaces` converts `[]model.FaceRead` into `[]model.FaceMinimal` (containing `face_type` and `face_name`), ensuring an empty JSON array `[]` (not `null`) if no faces exist.
  - `GetFace` returns full `model.Face` (`model.FaceRead`) definition including spots, rings, viewBox, and renderCross.
- Acceptance & Task Tracking: Update `docs/go_refactor/tasks/024-handler_faces.md` and `docs/plans/task.md` marking all items completed.

**Tech Stack:** Go 1.24+, `github.com/go-chi/chi/v5`, `net/http`, `net/http/httptest`, `encoding/json`, internal packages (`apperror`, `middleware`, `model`).

**Spec:**
- [docs/go_refactor/tasks/024-handler_faces.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/024-handler_faces.md)
- [backend-old/src/routers/v0/faces_router.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/routers/v0/faces_router.py)
- [backend/internal/service/face.go](file:///home/juanpa/Projects/arch-stats/backend/internal/service/face.go)
- [backend/internal/model/face.go](file:///home/juanpa/Projects/arch-stats/backend/internal/model/face.go)

## Global Constraints

- Git branch: `refactor/024-handler-faces`
- Package: `package handler` (`github.com/jpmolinamatute/arch-stats/backend/internal/handler`)
- Endpoints:
  - `GET /api/v0/faces` -> `FaceHandler.ListFaces` (returns 200 + `[]model.FaceMinimal`)
  - `GET /api/v0/faces/{face_type}` -> `FaceHandler.GetFace` (returns 200 + `model.FaceRead` or 404 if not found)
- Public access: No authentication required; requests without tokens must succeed.
- Error handling: Translate domain errors using `writeAppError(w, err)`. Unknown `face_type` returns 404 Not Found.
- Linting & Formatting: Code must pass `./scripts/linting.bash --go` (`gofumpt` and `golangci-lint run ./...`).
- Verification: `go test -race -v ./internal/handler/...`, `go vet ./...`, and `go build ./...` must succeed.
- Task completion: Must mark all tasks in `docs/go_refactor/tasks/024-handler_faces.md` and `docs/plans/task.md` as done (`[x]`).

---

## File Structure

```
backend/
├── internal/
│   └── handler/
│       ├── face.go                # [NEW] FaceService interface, FaceHandler struct, ListFaces, GetFace, Routes
│       └── face_test.go           # [NEW] Comprehensive httptest suite verifying ListFaces, GetFace, Routes, 404 handling, and public access
docs/
├── plans/
│   ├── task.md                    # [MODIFY] Track Task 024 live checklist progress (table-only)
│   └── 2026-09-05-handler-faces.md # [NEW] Implementation plan document
└── go_refactor/
    └── tasks/
        └── 024-handler_faces.md   # [MODIFY] Mark all acceptance criteria and steps as completed
```

---

## Proposed Tasks

### Task 1: Git Branch Setup & Live Tracker Initialization

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Clean working tree on current branch.
- Produces: `refactor/024-handler-faces` branch and updated `docs/plans/task.md` table.

- [ ] **Step 1: Create and checkout git branch**

```bash
git checkout -b refactor/024-handler-faces
```

- [ ] **Step 2: Initialize `docs/plans/task.md` with Task 024 tracker**

Update `docs/plans/task.md` to:
```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | IN_PROGRESS | Switch to branch `refactor/024-handler-faces` and initialize tracker |
| Task 2: Write Failing Tests for FaceHandler Interface and Endpoints | PENDING | Scaffold `face_test.go` with mock service and tests for ListFaces, GetFace, 404, and Routes |
| Task 3: Implement FaceHandler (`face.go`) | PENDING | Implement `FaceService` interface, `FaceHandler`, `ListFaces`, `GetFace`, and `Routes` |
| Task 4: End-to-End Suite Verification, Formatting & Linting | PENDING | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 5: Mark Tasks as Completed in Task Spec and Live Tracker | PENDING | Mark `024-handler_faces.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Verify git branch and commit tracker initialization**

```bash
git status
git add docs/plans/task.md
git commit -m "docs: initialize task tracker for task 024 handler faces"
```

---

### Task 2: Write Failing Tests for FaceHandler Interface and Endpoints

**Files:**
- Create: `backend/internal/handler/face_test.go`

**Interfaces:**
- Consumes: `model.FaceRead`, `model.FaceMinimal`, `model.FaceType`, `github.com/go-chi/chi/v5`, `net/http/httptest`.
- Produces: `mockFaceHandlerService` and test functions:
  - `TestNewFaceHandler`
  - `TestFaceHandler_ListFaces` (faces exist, empty list, nil list, service error, public access verification)
  - `TestFaceHandler_GetFace` (found face, unknown face type -> 404, empty type -> 422, service error, public access verification)
  - `TestFaceHandler_Routes` (GET `/faces/` matches ListFaces, GET `/faces/{face_type}` matches GetFace)

- [ ] **Step 1: Write `backend/internal/handler/face_test.go`**

```go
package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockFaceHandlerService struct {
	listAllFn func(ctx context.Context) ([]model.FaceRead, error)
	getByIDFn func(ctx context.Context, id string) (*model.FaceRead, error)
}

func (m *mockFaceHandlerService) ListAll(ctx context.Context) ([]model.FaceRead, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockFaceHandlerService) GetByID(ctx context.Context, id string) (*model.FaceRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("unimplemented")
}

func TestNewFaceHandler(t *testing.T) {
	svc := &mockFaceHandlerService{}
	h := handler.NewFaceHandler(svc)
	if h == nil {
		t.Fatal("expected NewFaceHandler to return non-nil instance")
	}
}

func TestFaceHandler_ListFaces(t *testing.T) {
	t.Run("returns 200 and list of FaceMinimal summaries", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return []model.FaceRead{
					{
						FaceType:    model.FaceTypeWA40Full,
						FaceName:    "WA 40cm Full",
						ViewBox:     400,
						RenderCross: true,
						Spots:       []model.Spot{{Diameter: 400}},
						Rings:       []model.Ring{{DataScore: 10, Fill: "#FFD700"}},
					},
					{
						FaceType:    model.FaceTypeWA60Full,
						FaceName:    "WA 60cm Full",
						ViewBox:     600,
						RenderCross: false,
						Spots:       []model.Spot{{Diameter: 600}},
						Rings:       []model.Ring{{DataScore: 10, Fill: "#FFD700"}},
					},
				}, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp []model.FaceMinimal
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response JSON: %v", err)
		}

		if len(resp) != 2 {
			t.Fatalf("expected 2 face summaries, got %d", len(resp))
		}
		if resp[0].FaceType != model.FaceTypeWA40Full || resp[0].FaceName != "WA 40cm Full" {
			t.Errorf("unexpected first item: %+v", resp[0])
		}
		if resp[1].FaceType != model.FaceTypeWA60Full || resp[1].FaceName != "WA 60cm Full" {
			t.Errorf("unexpected second item: %+v", resp[1])
		}
	})

	t.Run("returns 200 and empty JSON array when no faces exist", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return []model.FaceRead{}, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if body != "[]\n" && body != "[]" {
			t.Fatalf("expected empty array '[]', got %q", body)
		}
	})

	t.Run("returns 200 and empty JSON array when service returns nil slice", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return nil, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if body != "[]\n" && body != "[]" {
			t.Fatalf("expected empty array '[]', got %q", body)
		}
	})

	t.Run("returns 500 when service returns internal error", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return nil, errors.New("database failure")
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})

	t.Run("does not require authentication (public)", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return []model.FaceRead{}, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		// Create request without any auth header or context
		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 for unauthenticated request, got %d", rec.Code)
		}
	})
}

func TestFaceHandler_GetFace(t *testing.T) {
	t.Run("returns 200 and full Face definition when found", func(t *testing.T) {
		expectedFace := &model.FaceRead{
			FaceType:    model.FaceTypeWA40Full,
			FaceName:    "WA 40cm Full",
			ViewBox:     400,
			RenderCross: true,
			Spots: []model.Spot{
				{XOffset: 0, YOffset: 0, Diameter: 400},
			},
			Rings: []model.Ring{
				{DataScore: 10, Fill: "#FFD700", R: 20, Stroke: "#000000", StrokeWidth: 1},
				{DataScore: 9, Fill: "#FFD700", R: 40, Stroke: "#000000", StrokeWidth: 1},
			},
		}

		svc := &mockFaceHandlerService{
			getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
				if id == string(model.FaceTypeWA40Full) {
					return expectedFace, nil
				}
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/"+string(model.FaceTypeWA40Full), http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", string(model.FaceTypeWA40Full))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp model.FaceRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response JSON: %v", err)
		}

		if resp.FaceType != expectedFace.FaceType {
			t.Errorf("expected face_type %s, got %s", expectedFace.FaceType, resp.FaceType)
		}
		if resp.FaceName != expectedFace.FaceName {
			t.Errorf("expected face_name %s, got %s", expectedFace.FaceName, resp.FaceName)
		}
		if len(resp.Spots) != 1 || len(resp.Rings) != 2 {
			t.Errorf("unexpected spots or rings length: %+v", resp)
		}
	})

	t.Run("returns 404 when face_type is unknown", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/nonexistent_face", http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", "nonexistent_face")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when face_type URL parameter is empty or whitespace", func(t *testing.T) {
		svc := &mockFaceHandlerService{}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/   ", http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", "   ")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 500 when service returns internal error", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
				return nil, errors.New("storage error")
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/"+string(model.FaceTypeWA40Full), http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", string(model.FaceTypeWA40Full))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})

	t.Run("does not require authentication (public)", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
				return &model.FaceRead{
					FaceType: model.FaceTypeWA40Full,
					FaceName: "WA 40cm Full",
				}, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/"+string(model.FaceTypeWA40Full), http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", string(model.FaceTypeWA40Full))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 for unauthenticated request, got %d", rec.Code)
		}
	})
}

func TestFaceHandler_Routes(t *testing.T) {
	svc := &mockFaceHandlerService{
		listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
			return []model.FaceRead{{FaceType: model.FaceTypeWA40Full, FaceName: "WA 40cm Full"}}, nil
		},
		getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
			if id == string(model.FaceTypeWA40Full) {
				return &model.FaceRead{FaceType: model.FaceTypeWA40Full, FaceName: "WA 40cm Full"}, nil
			}
			return nil, apperror.ErrNotFound
		},
	}
	h := handler.NewFaceHandler(svc)

	r := chi.NewRouter()
	r.Route("/faces", h.Routes)

	t.Run("GET /faces/ routes to ListFaces", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/faces/", http.NoBody)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("GET /faces/{face_type} routes to GetFace", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/faces/"+string(model.FaceTypeWA40Full), http.NoBody)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})
}
```

- [ ] **Step 2: Run tests to verify compilation failure**

```bash
cd backend && go test ./internal/handler/... -run TestFaceHandler
```
Expected: FAIL with `undefined: handler.NewFaceHandler` or `face.go` not found.

- [ ] **Step 3: Commit failing test suite**

```bash
git add backend/internal/handler/face_test.go
git commit -m "test: add failing tests for FaceHandler endpoints and routing"
```

---

### Task 3: Implement FaceHandler (`face.go`)

**Files:**
- Create: `backend/internal/handler/face.go`

**Interfaces:**
- Consumes: `FaceService` interface (`ListAll`, `GetByID`), `model.FaceRead`, `model.FaceMinimal`, `github.com/go-chi/chi/v5`, internal `writeJSON` and `writeAppError`.
- Produces: `FaceService` interface, `FaceHandler` struct, `NewFaceHandler`, `ListFaces`, `GetFace`, and `Routes`.

- [ ] **Step 1: Write `backend/internal/handler/face.go`**

```go
package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// FaceService defines the domain operations required by FaceHandler.
type FaceService interface {
	ListAll(ctx context.Context) ([]model.FaceRead, error)
	GetByID(ctx context.Context, id string) (*model.FaceRead, error)
}

// FaceHandler manages HTTP endpoints for target face catalog definitions.
type FaceHandler struct {
	faceSvc FaceService
}

// NewFaceHandler constructs a FaceHandler with service dependency injection.
func NewFaceHandler(faceSvc FaceService) *FaceHandler {
	return &FaceHandler{
		faceSvc: faceSvc,
	}
}

// Routes registers all face catalog endpoints on the provided chi Router.
func (h *FaceHandler) Routes(r chi.Router) {
	r.Get("/", h.ListFaces)
	r.Get("/{face_type}", h.GetFace)
}

// ListFaces handles GET /api/v0/faces.
// Returns a list of target face summaries (face_type and face_name).
// This endpoint is public and does not require authentication.
func (h *FaceHandler) ListFaces(w http.ResponseWriter, r *http.Request) {
	faces, err := h.faceSvc.ListAll(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	summaries := make([]model.FaceMinimal, len(faces))
	for i, f := range faces {
		summaries[i] = model.FaceMinimal{
			FaceType: f.FaceType,
			FaceName: f.FaceName,
		}
	}

	_ = writeJSON(w, http.StatusOK, summaries)
}

// GetFace handles GET /api/v0/faces/{face_type}.
// Returns the full geometry and scoring layout for the requested face type.
// If face_type is not found, returns HTTP 404.
// This endpoint is public and does not require authentication.
func (h *FaceHandler) GetFace(w http.ResponseWriter, r *http.Request) {
	faceType := getURLParam(r, "face_type")
	if faceType == "" {
		faceType = getURLParam(r, "id")
	}

	if strings.TrimSpace(faceType) == "" {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "face_type is required"))
		return
	}

	face, err := h.faceSvc.GetByID(r.Context(), faceType)
	if err != nil {
		writeAppError(w, err)
		return
	}
	if face == nil {
		writeAppError(w, apperror.ErrNotFound)
		return
	}

	_ = writeJSON(w, http.StatusOK, face)
}
```

- [ ] **Step 2: Run tests to verify they pass**

```bash
cd backend && go test -v ./internal/handler/... -run TestFaceHandler
```
Expected: PASS for all `TestFaceHandler_*` tests.

- [ ] **Step 3: Commit implementation**

```bash
git add backend/internal/handler/face.go
git commit -m "feat: add faces HTTP handler with list and get-by-type endpoints"
```

---

### Task 4: End-to-End Suite Verification, Formatting & Linting

**Files:**
- Modify: `backend/internal/handler/face.go` (if formatting adjustments needed)
- Modify: `backend/internal/handler/face_test.go` (if formatting adjustments needed)

**Interfaces:**
- Consumes: Complete backend package tree.
- Produces: Clean test run, zero lint issues, and successful compilation.

- [ ] **Step 1: Run full handler tests with race detector**

```bash
cd backend && go test -race -v ./internal/handler/...
```
Expected: PASS with 0 race conditions.

- [ ] **Step 2: Run go vet**

```bash
cd backend && go vet ./...
```
Expected: 0 errors/warnings.

- [ ] **Step 3: Format and lint backend code**

```bash
cd backend && gofumpt -w internal/handler/face.go internal/handler/face_test.go
golangci-lint run ./...
```
Or alternatively from workspace root:
```bash
./scripts/linting.bash --go
```
Expected: Clean pass with 0 lint violations.

- [ ] **Step 4: Build backend packages**

```bash
cd backend && go build ./...
```
Expected: Clean compilation with 0 errors.

- [ ] **Step 5: Commit any formatting or lint fixes (if any)**

```bash
git add -u
git commit -m "style: format and lint face handler code" || true
```

---

### Task 5: Mark Tasks as Completed in Task Spec and Live Tracker

**Files:**
- Modify: `docs/go_refactor/tasks/024-handler_faces.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified implementation of Task 024.
- Produces: Completed checkboxes in task specification and updated live tracker.

- [ ] **Step 1: Update `docs/go_refactor/tasks/024-handler_faces.md`**

Mark all acceptance criteria and steps as completed:
```markdown
## Acceptance Criteria

- [x] `backend/internal/handler/face.go` implements `FaceHandler` with methods:
    - `ListFaces(w, r)` — GET `/api/v0/faces` — returns list of face summaries (face_type + face_name)
    - `GetFace(w, r)` — GET `/api/v0/faces/{face_type}` — returns full face definition
- [x] These endpoints do NOT go through auth middleware (public).
- [x] GetFace with unknown face_type returns 404.
- [x] Unit tests using `httptest` with mock service verify:
    - ListFaces returns 200 + JSON array of face summaries
    - GetFace with valid type returns 200 + full face JSON
    - GetFace with unknown type returns 404
- [x] `go test ./internal/handler/...` passes.
- [x] `go vet ./...` reports no issues.

## Steps

- [x] **Step 1: Write failing tests**
- [x] **Step 2: Run tests to verify they fail**
- [x] **Step 3: Implement `face.go`**
- [x] **Step 4: Run tests to verify they pass**
- [x] **Step 5: Run go vet and build**
- [x] **Step 6: Commit**
```

- [ ] **Step 2: Update `docs/plans/task.md`**

Update status column in `docs/plans/task.md` to:
```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | DONE | Switch to branch `refactor/024-handler-faces` and initialize tracker |
| Task 2: Write Failing Tests for FaceHandler Interface and Endpoints | DONE | Scaffold `face_test.go` with mock service and tests for ListFaces, GetFace, 404, and Routes |
| Task 3: Implement FaceHandler (`face.go`) | DONE | Implement `FaceService` interface, `FaceHandler`, `ListFaces`, `GetFace`, and `Routes` |
| Task 4: End-to-End Suite Verification, Formatting & Linting | DONE | Run full tests with race detection, `go vet`, `golangci-lint`, and `go build` |
| Task 5: Mark Tasks as Completed in Task Spec and Live Tracker | DONE | Mark `024-handler_faces.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Commit completed documentation updates**

```bash
git add docs/go_refactor/tasks/024-handler_faces.md docs/plans/task.md
git commit -m "docs: mark task 024 handler faces as completed"
```
