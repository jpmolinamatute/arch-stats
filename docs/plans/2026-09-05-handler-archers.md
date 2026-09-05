# Task 020: Build HTTP Handler — Archers Endpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the archers HTTP handler in `backend/internal/handler/archer.go` and its comprehensive test suite in `backend/internal/handler/archer_test.go`, porting the Python `archer_router.py`. This provides CRUD HTTP endpoints for archer management (`List`, `GetByID`, `Create`, `Update`, `Delete`), chi router integration, and domain error mapping. In addition, review and finalize Task 019 (`019-handler_auth.md`) by verifying its implementation against acceptance criteria and marking all its steps as completed.

**Architecture:**
- `archer.go`: `ArcherHandler` exposing `List`, `GetByID`, `Create`, `Update`, `Delete`, and `Routes` methods adhering to standard `http.HandlerFunc` / `chi.Router` conventions. Delegates business logic and validation to `ArcherService` interface. Extracts URL parameters (`id` / `archer_id`) using `chi.URLParam` with `r.PathValue` fallback.
- `helpers.go`: Shared HTTP handler utilities for JSON decoding (`readJSON`), JSON encoding (`writeJSON`), and error writing (`writeAppError`) mapped to standard frontend-compatible JSON envelopes (`{"detail": "..."}`).
- `interfaces`: Consolidate `ArcherService` interface in `package handler` to provide the full contract needed for both `AuthHandler` and `ArcherHandler` (`List`, `GetByID`, `Create`, `Update`, `Delete`).
- `error_mapper`: Errors are mapped via `middleware.WriteError` through `writeAppError(w, err)` (`apperror.ErrNotFound` -> 404, `apperror.ErrValidation` -> 422).

**Tech Stack:** Go 1.27+, `github.com/go-chi/chi/v5`, `net/http`, `net/http/httptest`, `encoding/json`, `github.com/google/uuid`, internal packages (`apperror`, `middleware`, `model`, `service`).

**Spec:**
- [docs/go_refactor/tasks/020-handler_archers.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/020-handler_archers.md)
- [docs/go_refactor/tasks/019-handler_auth.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/019-handler_auth.md)

## Global Constraints

- Git branch: `refactor/020-handler-archers`
- Package paths:
  - `github.com/jpmolinamatute/arch-stats/backend/internal/handler`
- Error handling: Wrap internal errors with `%w` or `apperror.Wrap`. Domain errors (`apperror.ErrNotFound`, `apperror.ErrValidation`) must map to HTTP 404 and 422 respectively via `writeAppError(w, err)`.
- JSON formatting: JSON field tags match existing contracts (snake_case, e.g. `archer_id`, `first_name`, `last_name`, `date_of_birth`).
- Formatting & Linting: Code must pass `gofumpt` and `golangci-lint run ./...` (`./scripts/linting.bash --go`).
- Tests: `go test -race -v ./internal/handler/...` must pass.
- Verification: `go vet ./...` and `go build ./...` must succeed.

---

## File Structure

```
backend/
├── go.mod                       # [MODIFY] Add github.com/go-chi/chi/v5 dependency
├── go.sum                       # [MODIFY] Checksum for chi dependency
└── internal/
    └── handler/
        ├── auth.go              # [MODIFY] Remove duplicate ArcherService declaration; reference consolidated interface
        ├── auth_test.go         # [MODIFY] Expand mockArcherService to implement full ArcherService interface
        ├── archer.go            # [NEW] ArcherService interface, ArcherHandler struct, List, GetByID, Create, Update, Delete, Routes
        └── archer_test.go       # [NEW] Comprehensive httptest unit tests for archer endpoints
docs/
├── plans/
│   ├── task.md                  # [MODIFY] Track Task 020 live checklist progress
│   └── 2026-09-05-handler-archers.md # [NEW] Implementation plan
└── go_refactor/
    └── tasks/
        ├── 019-handler_auth.md  # [MODIFY] Mark all steps and acceptance criteria as completed
        └── 020-handler_archers.md # [MODIFY] Mark all steps and acceptance criteria as completed after implementation
```

---

## Proposed Tasks

### Task 1: Review and Finalize Task 019 (Auth Handler)

**Files:**
- Modify: `docs/go_refactor/tasks/019-handler_auth.md`

**Interfaces:**
- Consumes: Task 019 acceptance criteria and existing `backend/internal/handler/auth.go`, `auth_test.go`, `helpers.go`, `helpers_test.go`.
- Produces: Fully checked and verified `019-handler_auth.md`.

- [ ] **Step 1: Verify all Task 019 acceptance criteria against codebase**

Review checklist:
1. `backend/internal/handler/auth.go` implements `AuthHandler` with `Login`, `Register`, `Logout`, `Me`: Verified.
2. Handler sets HTTP-only `arch_stats_auth` cookies with appropriate flags: Verified.
3. Handler returns JSON responses matching API contracts: Verified.
4. Unit tests in `backend/internal/handler/auth_test.go` and `helpers_test.go` verify all cases: Verified (12 tests pass).
5. `go test ./internal/handler/...` passes with race detection: Verified.
6. `go vet ./...` and `golangci-lint` report no issues: Verified (0 issues).
7. Commit `0319e40 refactor/019 handler auth (#206)` exists on `main`: Verified.

- [ ] **Step 2: Update `docs/go_refactor/tasks/019-handler_auth.md` with `- [x]` for all steps and acceptance criteria**

Update lines 20-37 and 52-97 in `docs/go_refactor/tasks/019-handler_auth.md`:
Replace all `- [ ]` with `- [x]`.

- [ ] **Step 3: Commit the marked task file**

```bash
git add docs/go_refactor/tasks/019-handler_auth.md
git commit -m "docs: mark task 019 handler auth as completed"
```

---

### Task 2: Git Branch Setup & Chi Dependency

**Files:**
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: `github.com/go-chi/chi/v5`
- Produces: Chi router dependency installed in `backend/go.mod`

- [ ] **Step 1: Create git branch `refactor/020-handler-archers`**

```bash
git switch -c refactor/020-handler-archers
```

- [ ] **Step 2: Update live tracker in `docs/plans/task.md`**

Replace `docs/plans/task.md` with the live tracker table for Task 020:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Review & Finalize Task 019 | DONE | Verify auth handler implementation and mark 019-handler_auth.md complete |
| Task 2: Git Branch & Chi Dependency | IN_PROGRESS | Switch to `refactor/020-handler-archers` and add `github.com/go-chi/chi/v5` |
| Task 3: ArcherService Interface & Handler Scaffolding | PENDING | Define consolidated ArcherService and ArcherHandler constructor |
| Task 4: Implement List & GetByID Endpoints (TDD) | PENDING | Implement List and GetByID with unit test coverage |
| Task 5: Implement Create, Update, & Delete Endpoints (TDD) | PENDING | Implement Create, Update, and Delete with unit test coverage |
| Task 6: Chi Routes Mounting & Integration | PENDING | Expose Routes(r chi.Router) and test router integration |
| Task 7: Verification, Linting, & Mark Task 020 Complete | PENDING | Run tests with race detection, linter, and mark task complete |
```

- [ ] **Step 3: Install `github.com/go-chi/chi/v5` dependency**

```bash
cd backend && go get github.com/go-chi/chi/v5 && go mod tidy
```

- [ ] **Step 4: Verify build and existing tests**

```bash
cd backend && go test ./internal/handler/... && go vet ./...
```
Expected: PASS

- [ ] **Step 5: Commit dependency changes**

```bash
git add backend/go.mod backend/go.sum docs/plans/task.md
git commit -m "build(backend): add go-chi/chi dependency for handler routing"
```

---

### Task 3: Interface Consolidation & ArcherHandler Scaffolding

**Files:**
- Modify: `backend/internal/handler/auth.go:26-30`
- Modify: `backend/internal/handler/auth_test.go:66-76`
- Create: `backend/internal/handler/archer.go`
- Create: `backend/internal/handler/archer_test.go`

**Interfaces:**
- Consumes: `model.ArcherRead`, `model.ArcherFilter`, `model.ArcherCreate`, `model.ArcherSet`, `uuid.UUID`
- Produces:
  ```go
  type ArcherService interface {
      List(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
      GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
      Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
      Update(ctx context.Context, id uuid.UUID, data model.ArcherSet) error
      Delete(ctx context.Context, id uuid.UUID) error
  }
  type ArcherHandler struct { ... }
  func NewArcherHandler(archerSvc ArcherService) *ArcherHandler
  ```

- [ ] **Step 1: Write failing scaffolding test in `backend/internal/handler/archer_test.go`**

Create `backend/internal/handler/archer_test.go`:

```go
package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockArcherHandlerService struct {
	listFn    func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	createFn  func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	updateFn  func(ctx context.Context, id uuid.UUID, data model.ArcherSet) error
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockArcherHandlerService) List(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockArcherHandlerService) GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("unimplemented")
}

//nolint:gocritic // hugeParam: data matches domain model parameter specification
func (m *mockArcherHandlerService) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, errors.New("unimplemented")
}

//nolint:gocritic // hugeParam: data matches domain model parameter specification
func (m *mockArcherHandlerService) Update(ctx context.Context, id uuid.UUID, data model.ArcherSet) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, data)
	}
	return errors.New("unimplemented")
}

func (m *mockArcherHandlerService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return errors.New("unimplemented")
}

func TestNewArcherHandler(t *testing.T) {
	svc := &mockArcherHandlerService{}
	h := handler.NewArcherHandler(svc)
	if h == nil {
		t.Fatal("expected NewArcherHandler to return non-nil instance")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/handler/... -run TestNewArcherHandler
```
Expected: Compilation failure (`handler.NewArcherHandler undefined`).

- [ ] **Step 3: Consolidate `ArcherService` and implement `archer.go` scaffolding**

1. In `backend/internal/handler/auth.go`, remove the local `ArcherService` interface definition (lines 26-29) since it will now be defined as the full service interface in `backend/internal/handler/archer.go`.
2. In `backend/internal/handler/auth_test.go`, expand `mockArcherService` to satisfy the full interface:
```go
type mockArcherService struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	listFn    func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
	createFn  func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	updateFn  func(ctx context.Context, id uuid.UUID, data model.ArcherSet) error
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockArcherService) GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockArcherService) List(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, errors.New("unimplemented")
}

//nolint:gocritic // hugeParam: data matches domain model parameter specification
func (m *mockArcherService) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, errors.New("unimplemented")
}

//nolint:gocritic // hugeParam: data matches domain model parameter specification
func (m *mockArcherService) Update(ctx context.Context, id uuid.UUID, data model.ArcherSet) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, data)
	}
	return errors.New("unimplemented")
}

func (m *mockArcherService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return errors.New("unimplemented")
}
```

3. Create `backend/internal/handler/archer.go`:
```go
package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ArcherService defines the persistence and business operations required by the archer and auth handlers.
type ArcherService interface {
	List(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	Update(ctx context.Context, id uuid.UUID, data model.ArcherSet) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ArcherHandler manages HTTP endpoints for archer CRUD operations.
type ArcherHandler struct {
	archerSvc ArcherService
}

// NewArcherHandler constructs an ArcherHandler with service dependency injection.
func NewArcherHandler(archerSvc ArcherService) *ArcherHandler {
	return &ArcherHandler{
		archerSvc: archerSvc,
	}
}

func getURLParam(r *http.Request, key string) string {
	if val := chi.URLParam(r, key); val != "" {
		return val
	}
	return r.PathValue(key)
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/handler/... -run TestNewArcherHandler -v
```
Expected: PASS

- [ ] **Step 5: Verify existing auth tests still pass**

```bash
cd backend && go test ./internal/handler/... -v
```
Expected: All tests PASS.

- [ ] **Step 6: Commit scaffolding**

```bash
git add backend/internal/handler/
git commit -m "feat(handler): define ArcherService interface and ArcherHandler constructor"
```

---

### Task 4: Implement `List` & `GetByID` Endpoints (TDD)

**Files:**
- Modify: `backend/internal/handler/archer.go`
- Modify: `backend/internal/handler/archer_test.go`

**Interfaces:**
- `List(w http.ResponseWriter, r *http.Request)` — GET `/api/v0/archer/`
  - Calls `h.archerSvc.List(ctx, model.ArcherFilter{})`
  - Returns 200 OK + `[]model.ArcherRead` (or empty array `[]`)
  - Handles errors with `writeAppError(w, err)`
- `GetByID(w http.ResponseWriter, r *http.Request)` — GET `/api/v0/archer/{id}`
  - Extracts ID from route param (`"id"` or `"archer_id"`)
  - Parses UUID; returns 422 (`apperror.ErrValidation`) on invalid/empty UUID
  - Calls `h.archerSvc.GetByID(ctx, id)`
  - Returns 200 OK + `model.ArcherRead` on success
  - Returns 404 (`apperror.ErrNotFound`) if archer does not exist

- [ ] **Step 1: Write failing tests for `List` and `GetByID` in `backend/internal/handler/archer_test.go`**

Append to `backend/internal/handler/archer_test.go`:

```go
func TestArcherHandler_List(t *testing.T) {
	t.Run("returns 200 and JSON array when archers exist", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		svc := &mockArcherHandlerService{
			listFn: func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
				return []model.ArcherRead{
					{ArcherID: id1, FirstName: "Katniss", LastName: "Everdeen"},
					{ArcherID: id2, FirstName: "Robin", LastName: "Hood"},
				}, nil
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp []model.ArcherRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("expected 2 archers, got %d", len(resp))
		}
		if resp[0].ArcherID != id1 || resp[1].ArcherID != id2 {
			t.Fatalf("unexpected archers returned: %+v", resp)
		}
	})

	t.Run("returns 200 and empty JSON array when no archers exist", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			listFn: func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
				return []model.ArcherRead{}, nil
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		body := strings.TrimSpace(rec.Body.String())
		if body != "[]" {
			t.Fatalf("expected '[]', got %q", body)
		}
	})

	t.Run("returns 500 when service fails", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			listFn: func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
				return nil, errors.New("db query error")
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestArcherHandler_GetByID(t *testing.T) {
	targetID := uuid.New()

	t.Run("returns 200 and archer JSON when found", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				if id == targetID {
					return &model.ArcherRead{
						ArcherID:  targetID,
						FirstName: "Katniss",
						LastName:  "Everdeen",
						Email:     "katniss@district12.org",
					}, nil
				}
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/"+targetID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", targetID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
		}

		var resp model.ArcherRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ArcherID != targetID || resp.FirstName != "Katniss" {
			t.Fatalf("unexpected archer: %+v", resp)
		}
	})

	t.Run("returns 404 when archer does not exist", func(t *testing.T) {
		nonExistentID := uuid.New()
		svc := &mockArcherHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/"+nonExistentID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when id is invalid UUID", func(t *testing.T) {
		svc := &mockArcherHandlerService{}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/invalid-uuid", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetByID(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./internal/handler/... -run "TestArcherHandler_List|TestArcherHandler_GetByID"
```
Expected: Compilation failure (`h.List` and `h.GetByID` undefined).

- [ ] **Step 3: Implement `List` and `GetByID` in `backend/internal/handler/archer.go`**

Add `List` and `GetByID` to `backend/internal/handler/archer.go`:

```go
// List handles GET /api/v0/archer/.
// It queries archers matching default filter criteria and returns a JSON list.
func (h *ArcherHandler) List(w http.ResponseWriter, r *http.Request) {
	archers, err := h.archerSvc.List(r.Context(), model.ArcherFilter{})
	if err != nil {
		writeAppError(w, err)
		return
	}

	if archers == nil {
		archers = []model.ArcherRead{}
	}

	_ = writeJSON(w, http.StatusOK, archers)
}

// GetByID handles GET /api/v0/archer/{id}.
// It retrieves a single archer profile by primary key identifier.
func (h *ArcherHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := getURLParam(r, "id")
	if idStr == "" {
		idStr = getURLParam(r, "archer_id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer id is required"))
		return
	}

	archer, err := h.archerSvc.GetByID(r.Context(), id)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, archer)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/handler/... -run "TestArcherHandler_List|TestArcherHandler_GetByID" -v
```
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add backend/internal/handler/archer.go backend/internal/handler/archer_test.go
git commit -m "feat(handler): implement List and GetByID archer endpoints"
```

---

### Task 5: Implement `Create`, `Update`, & `Delete` Endpoints (TDD)

**Files:**
- Modify: `backend/internal/handler/archer.go`
- Modify: `backend/internal/handler/archer_test.go`

**Interfaces:**
- `Create(w http.ResponseWriter, r *http.Request)` — POST `/api/v0/archer/`
  - Decodes `model.ArcherCreate` payload via `readJSON`
  - Calls `h.archerSvc.Create(ctx, req)`
  - Returns 201 Created + `model.ArcherID{ArcherID: id}` (`{"archer_id": "..."}`)
  - Maps validation errors to 422
- `Update(w http.ResponseWriter, r *http.Request)` — PATCH `/api/v0/archer/`
  - Decodes `model.ArcherUpdate` payload via `readJSON`
  - Validates `where.archer_id` is present and valid UUID (422 if missing)
  - Calls `h.archerSvc.Update(ctx, *req.Where.ArcherID, req.Data)`
  - Returns 200 OK
  - Maps `ErrNotFound` to 404, `ErrValidation` to 422
- `Delete(w http.ResponseWriter, r *http.Request)` — DELETE `/api/v0/archer/{id}`
  - Extracts ID parameter; validates UUID (422 if invalid)
  - Calls `h.archerSvc.Delete(ctx, id)`
  - Returns 204 No Content
  - Maps `ErrNotFound` to 404

- [ ] **Step 1: Write failing tests for `Create`, `Update`, and `Delete` in `backend/internal/handler/archer_test.go`**

Append to `backend/internal/handler/archer_test.go`:

```go
func TestArcherHandler_Create(t *testing.T) {
	createdID := uuid.New()

	t.Run("create with valid payload returns 201 and archer_id JSON", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			createFn: func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
				if data.FirstName == "Katniss" && data.LastName == "Everdeen" {
					return createdID, nil
				}
				return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "invalid data")
			},
		}
		h := handler.NewArcherHandler(svc)

		payload := model.ArcherCreate{
			FirstName:     "Katniss",
			LastName:      "Everdeen",
			Email:         "katniss@district12.org",
			DateOfBirth:   "2000-05-08",
			Gender:        model.GenderFemale,
			Bowstyle:      model.BowstyleBarebow,
			DrawWeight:    40.0,
			GoogleSubject: "google-sub-12345",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v0/archer/", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
		}

		var resp model.ArcherID
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ArcherID != createdID {
			t.Fatalf("expected created id %s, got %s", createdID, resp.ArcherID)
		}
	})

	t.Run("create with invalid json payload returns 422", func(t *testing.T) {
		svc := &mockArcherHandlerService{}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/archer/", strings.NewReader("{invalid-json"))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("create when service returns validation error returns 422", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			createFn: func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
				return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "first_name is required")
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/archer/", strings.NewReader(`{"first_name":""}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})
}

func TestArcherHandler_Update(t *testing.T) {
	targetID := uuid.New()
	newFirstName := "Mockingjay"

	t.Run("update with valid payload returns 200", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			updateFn: func(ctx context.Context, id uuid.UUID, data model.ArcherSet) error {
				if id == targetID && data.FirstName != nil && *data.FirstName == "Mockingjay" {
					return nil
				}
				return apperror.Wrap(apperror.ErrValidation, "unexpected update parameters")
			},
		}
		h := handler.NewArcherHandler(svc)

		updatePayload := model.ArcherUpdate{
			Where: model.ArcherFilter{ArcherID: &targetID},
			Data:  model.ArcherSet{FirstName: &newFirstName},
		}
		body, _ := json.Marshal(updatePayload)
		req := httptest.NewRequest(http.MethodPatch, "/api/v0/archer/", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("update with missing where.archer_id returns 422", func(t *testing.T) {
		svc := &mockArcherHandlerService{}
		h := handler.NewArcherHandler(svc)

		updatePayload := model.ArcherUpdate{
			Where: model.ArcherFilter{},
			Data:  model.ArcherSet{FirstName: &newFirstName},
		}
		body, _ := json.Marshal(updatePayload)
		req := httptest.NewRequest(http.MethodPatch, "/api/v0/archer/", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("update returns 404 when archer does not exist", func(t *testing.T) {
		nonExistentID := uuid.New()
		svc := &mockArcherHandlerService{
			updateFn: func(ctx context.Context, id uuid.UUID, data model.ArcherSet) error {
				return apperror.ErrNotFound
			},
		}
		h := handler.NewArcherHandler(svc)

		updatePayload := model.ArcherUpdate{
			Where: model.ArcherFilter{ArcherID: &nonExistentID},
			Data:  model.ArcherSet{FirstName: &newFirstName},
		}
		body, _ := json.Marshal(updatePayload)
		req := httptest.NewRequest(http.MethodPatch, "/api/v0/archer/", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestArcherHandler_Delete(t *testing.T) {
	targetID := uuid.New()

	t.Run("delete with valid id returns 204", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				if id == targetID {
					return nil
				}
				return apperror.ErrNotFound
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodDelete, "/api/v0/archer/"+targetID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", targetID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.Delete(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("delete returns 404 when archer does not exist", func(t *testing.T) {
		nonExistentID := uuid.New()
		svc := &mockArcherHandlerService{
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				return apperror.ErrNotFound
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodDelete, "/api/v0/archer/"+nonExistentID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("delete returns 422 when id is invalid UUID", func(t *testing.T) {
		svc := &mockArcherHandlerService{}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodDelete, "/api/v0/archer/invalid-uuid", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.Delete(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./internal/handler/... -run "TestArcherHandler_Create|TestArcherHandler_Update|TestArcherHandler_Delete"
```
Expected: Compilation failure (`h.Create`, `h.Update`, `h.Delete` undefined).

- [ ] **Step 3: Implement `Create`, `Update`, and `Delete` in `backend/internal/handler/archer.go`**

Add methods to `backend/internal/handler/archer.go`:

```go
// Create handles POST /api/v0/archer/.
// It parses the archer creation payload, validates and persists the new archer, and returns 201 Created.
func (h *ArcherHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.ArcherCreate
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	id, err := h.archerSvc.Create(r.Context(), req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusCreated, model.ArcherID{ArcherID: id})
}

// Update handles PATCH /api/v0/archer/.
// It updates archer fields matching the specified where filter.
func (h *ArcherHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.ArcherUpdate
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.Where.ArcherID == nil || *req.Where.ArcherID == uuid.Nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "where.archer_id is required"))
		return
	}

	if err := h.archerSvc.Update(r.Context(), *req.Where.ArcherID, req.Data); err != nil {
		writeAppError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// Delete handles DELETE /api/v0/archer/{id}.
// It removes an archer profile by primary key identifier and returns 204 No Content.
func (h *ArcherHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := getURLParam(r, "id")
	if idStr == "" {
		idStr = getURLParam(r, "archer_id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer id is required"))
		return
	}

	if err := h.archerSvc.Delete(r.Context(), id); err != nil {
		writeAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/handler/... -run "TestArcherHandler_Create|TestArcherHandler_Update|TestArcherHandler_Delete" -v
```
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add backend/internal/handler/archer.go backend/internal/handler/archer_test.go
git commit -m "feat(handler): implement Create, Update, and Delete archer endpoints"
```

---

### Task 6: Chi Route Mounting & Integration

**Files:**
- Modify: `backend/internal/handler/archer.go`
- Modify: `backend/internal/handler/archer_test.go`

**Interfaces:**
- Produces: `func (h *ArcherHandler) Routes(r chi.Router)`

- [ ] **Step 1: Write failing test for `Routes` mounting in `backend/internal/handler/archer_test.go`**

Append to `backend/internal/handler/archer_test.go`:

```go
func TestArcherHandler_Routes(t *testing.T) {
	targetID := uuid.New()
	svc := &mockArcherHandlerService{
		listFn: func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
			return []model.ArcherRead{{ArcherID: targetID, FirstName: "Katniss"}}, nil
		},
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
			if id == targetID {
				return &model.ArcherRead{ArcherID: targetID, FirstName: "Katniss"}, nil
			}
			return nil, apperror.ErrNotFound
		},
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			if id == targetID {
				return nil
			}
			return apperror.ErrNotFound
		},
	}
	h := handler.NewArcherHandler(svc)

	r := chi.NewRouter()
	r.Route("/archer", h.Routes)

	t.Run("GET /archer/ matches List", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/archer/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GET /archer/{id} matches GetByID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/archer/"+targetID.String(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("DELETE /archer/{id} matches Delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/archer/"+targetID.String(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/handler/... -run TestArcherHandler_Routes
```
Expected: Compilation failure (`h.Routes undefined`).

- [ ] **Step 3: Implement `Routes` method in `backend/internal/handler/archer.go`**

Add `Routes` to `backend/internal/handler/archer.go`:

```go
// Routes registers all archer CRUD endpoints on the provided chi Router.
func (h *ArcherHandler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Post("/", h.Create)
	r.Patch("/", h.Update)
	r.Delete("/{id}", h.Delete)
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/handler/... -run TestArcherHandler_Routes -v
```
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add backend/internal/handler/archer.go backend/internal/handler/archer_test.go
git commit -m "feat(handler): expose chi Routes mounting for archer handler"
```

---

### Task 7: Full Verification, Linting, & Mark Task 020 Complete

**Files:**
- Modify: `docs/go_refactor/tasks/020-handler_archers.md`
- Modify: `docs/plans/task.md`

- [ ] **Step 1: Run full unit test suite with race detector**

```bash
cd backend && go test -race -v ./internal/handler/...
```
Expected: All tests pass.

- [ ] **Step 2: Run Go linter and formatter**

```bash
./scripts/linting.bash --go
```
Expected: 0 issues, all tests pass.

- [ ] **Step 3: Run `go vet` and `go build`**

```bash
cd backend && go vet ./... && go build ./...
```
Expected: Clean exit status 0.

- [ ] **Step 4: Mark Task 020 acceptance criteria and steps as completed**

In `docs/go_refactor/tasks/020-handler_archers.md`:
Replace all `- [ ]` with `- [x]`.

- [ ] **Step 5: Update `docs/plans/task.md` tracker to reflect completion**

Update all rows in `docs/plans/task.md` to `DONE`.

- [ ] **Step 6: Commit documentation and tracker updates**

```bash
git add docs/go_refactor/tasks/020-handler_archers.md docs/plans/task.md
git commit -m "docs: mark task 020 handler archers as completed"
```

---

## Verification Plan

### Automated Tests
1. **Handler unit test suite:**
   ```bash
   cd backend && go test -race -v ./internal/handler/...
   ```
   Validates:
   - `List`: empty array, populated array, error propagation
   - `GetByID`: 200 OK + JSON, 404 on missing UUID, 422 on invalid UUID format
   - `Create`: 201 Created + `{"archer_id": "..."}`, 422 on malformed JSON, 422 on service validation error
   - `Update`: 200 OK, 404 on missing record, 422 on missing `where.archer_id`, 422 on invalid data
   - `Delete`: 204 No Content, 404 on missing record, 422 on invalid UUID format
   - `Routes`: Chi router mounting and routing to correct handler methods
   - Existing `AuthHandler` and `helpers` tests continue to pass without regressions.

2. **Linting and code style:**
   ```bash
   ./scripts/linting.bash --go
   ```
   Validates formatting (`gofumpt`), linter passes with 0 issues (`golangci-lint`), and all backend package tests pass.

3. **Compilation & Vet:**
   ```bash
   cd backend && go vet ./... && go build ./...
   ```
   Ensures clean compilation across the entire module.

### Manual Verification
1. Inspect `docs/go_refactor/tasks/019-handler_auth.md` and confirm all checkboxes are marked `- [x]`.
2. Inspect `docs/go_refactor/tasks/020-handler_archers.md` and confirm all checkboxes are marked `- [x]`.
3. Check `git status` and `git log` to ensure all commits are structured cleanly.
