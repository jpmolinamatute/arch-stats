# Task 016: Build Service Layer — Face and Target Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the service layer for target face catalog definitions (`FaceService`) and shooting lane target configurations (`TargetService`) with repository interface dependency injection, input validation, error mapping, and complete unit test suites with mock repositories.

**Architecture:**
- `FaceService` encapsulates access to target face geometry and layout definitions. It accepts a `FaceRepository` interface via constructor injection, validates queries (`GetByID`, `ListByType`), and translates misses to `apperror.ErrNotFound`.
- `TargetService` manages shooting lane target configurations (`session_id`, `distance`, `lane`). It accepts a `TargetRepository` interface via constructor injection, validates incoming creation and update payloads (`TargetCreate`, `TargetSet`), checks target existence on update, and translates persistence operations to domain entities and sentinel errors. Note on domain model clarification: As verified in `model.TargetCreate`, `repository.TargetRepo`, and DB migration `004_2025-09-26_shooting_sessions_table.sql`, the `target` entity represents a lane target in a session (`session_id`, `distance`, `lane`). Target faces (`face_type`) are assigned to individual `slot` entities, not `target`. `TargetService.Create` therefore validates `session_id`, `distance` (1–100), and `lane` (1–100).
- Both services consume repository interfaces defined directly in the service package, enabling mock-based unit tests without database dependencies.

**Tech Stack:** Go 1.27+, `github.com/google/uuid`, standard library (`context`, `errors`, `fmt`, `strings`), internal packages (`model`, `apperror`, `repository`).

**Spec:** [docs/go_refactor/tasks/016-service_layer_face_and_target.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/016-service_layer_face_and_target.md)

## Global Constraints

- Git branch: `refactor/016-service-layer-face-and-target`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/service`
- Error handling: Wrap internal errors with `%w` using contextual descriptive messages (`fmt.Errorf("...: %w", err)`). Return sentinel `apperror.ErrNotFound` and `apperror.Wrap(apperror.ErrValidation, ...)` as appropriate.
- Dependency injection: Services must accept repository interfaces in their constructors (`NewFaceService(repo FaceRepository)`, `NewTargetService(repo TargetRepository)`).
- Mock testing: Service unit tests must use mock repositories implementing the repository interfaces without database or network dependencies.
- Formatting must adhere to `gofumpt` and linting must pass `golangci-lint run ./...`.
- `go test -race ./internal/service/... -v` must pass.
- `go vet ./...` must report no issues.

---

## File Structure

```
backend/
└── internal/
    └── service/
        ├── face.go               # [NEW] FaceRepository interface, FaceService struct, constructor, query & validation methods
        ├── face_test.go          # [NEW] mockFaceRepo, interface assertions, and unit test suite for FaceService
        ├── target.go             # [NEW] TargetRepository interface, TargetService struct, constructor, CRUD & validation methods
        └── target_test.go        # [NEW] mockTargetRepo, interface assertions, and unit test suite for TargetService
```

---

### Task 1: Git Branch Setup

**Files:**
- Modify: git branch switch

**Interfaces:**
- Consumes: `main` branch
- Produces: `refactor/016-service-layer-face-and-target` branch

- [x] **Step 1: Check out git branch**

```bash
git switch -c refactor/016-service-layer-face-and-target
```

- [x] **Step 2: Verify git status is clean**

```bash
git status
```
Expected: On branch `refactor/016-service-layer-face-and-target`, working tree clean.

---

### Task 2: Face Service Tests & Implementation (`face_test.go` & `face.go`)

**Files:**
- Create: `backend/internal/service/face_test.go`
- Create: `backend/internal/service/face.go`

**Interfaces:**
- Consumes:
  - `model.FaceRead`, `model.FaceType` from `backend/internal/model/face.go` and `backend/internal/model/enums.go`
  - `isValidFaceType` helper (declared in `backend/internal/service/slot.go` within package `service`)
  - `apperror.ErrNotFound`, `apperror.ErrValidation`, `apperror.Wrap` from `backend/internal/apperror/errors.go`
- Produces:
  - `FaceRepository` interface:
    - `FindByID(ctx context.Context, id string) (*model.FaceRead, error)`
    - `FindAll(ctx context.Context) ([]model.FaceRead, error)`
    - `FindByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error)`
  - `FaceService` struct and constructor `NewFaceService(repo FaceRepository) *FaceService`
  - Methods:
    - `GetByID(ctx context.Context, id string) (*model.FaceRead, error)`
    - `ListAll(ctx context.Context) ([]model.FaceRead, error)`
    - `ListByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error)`

- [x] **Step 1: Write failing tests in `backend/internal/service/face_test.go`**

Create `backend/internal/service/face_test.go`:

```go
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
	"github.com/jpmolinamatute/arch-stats/backend/internal/service"
)

var (
	_ service.FaceRepository = (*mockFaceRepo)(nil)
	_ service.FaceRepository = (*repository.FaceRepo)(nil)
)

type mockFaceRepo struct {
	findByIDFn   func(ctx context.Context, id string) (*model.FaceRead, error)
	findAllFn    func(ctx context.Context) ([]model.FaceRead, error)
	findByTypeFn func(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error)
}

func (m *mockFaceRepo) FindByID(ctx context.Context, id string) (*model.FaceRead, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockFaceRepo) FindAll(ctx context.Context) ([]model.FaceRead, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return nil, nil
}

func (m *mockFaceRepo) FindByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
	if m.findByTypeFn != nil {
		return m.findByTypeFn(ctx, faceType)
	}
	return nil, nil
}

func sampleFaceRead(faceType model.FaceType, faceName string) model.FaceRead {
	return model.FaceRead{
		FaceType:    faceType,
		FaceName:    faceName,
		Spots:       []model.Spot{{XOffset: 0, YOffset: 0, Diameter: 40}},
		Rings:       []model.Ring{{DataScore: 10, Fill: "#FFD700", R: 2, Stroke: "#000000", StrokeWidth: 0.1}},
		ViewBox:     400,
		RenderCross: true,
	}
}

func TestFaceService_GetByID_Success(t *testing.T) {
	expected := sampleFaceRead(model.FaceTypeWA40Full, "WA 40cm Full")
	mock := &mockFaceRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
			if id == "wa_40cm_full" {
				return &expected, nil
			}
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	face, err := svc.GetByID(context.Background(), "wa_40cm_full")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if face == nil || face.FaceType != model.FaceTypeWA40Full {
		t.Fatalf("unexpected face: %+v", face)
	}
}

func TestFaceService_GetByID_EmptyIDReturnsValidationError(t *testing.T) {
	mock := &mockFaceRepo{}
	svc := service.NewFaceService(mock)

	_, err := svc.GetByID(context.Background(), "  ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestFaceService_GetByID_NotFound(t *testing.T) {
	mock := &mockFaceRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	_, err := svc.GetByID(context.Background(), "unknown_face")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFaceService_GetByID_RepoError(t *testing.T) {
	mock := &mockFaceRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
			return nil, errors.New("database failure")
		},
	}

	svc := service.NewFaceService(mock)
	_, err := svc.GetByID(context.Background(), "wa_40cm_full")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFaceService_ListAll_Success(t *testing.T) {
	expected := []model.FaceRead{
		sampleFaceRead(model.FaceTypeWA40Full, "WA 40cm Full"),
		sampleFaceRead(model.FaceTypeWA60Full, "WA 60cm Full"),
	}
	mock := &mockFaceRepo{
		findAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
			return expected, nil
		},
	}

	svc := service.NewFaceService(mock)
	faces, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != 2 {
		t.Fatalf("expected 2 faces, got %d", len(faces))
	}
}

func TestFaceService_ListAll_NilReturnsEmptySlice(t *testing.T) {
	mock := &mockFaceRepo{
		findAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	faces, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if faces == nil || len(faces) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v", faces)
	}
}

func TestFaceService_ListAll_RepoError(t *testing.T) {
	mock := &mockFaceRepo{
		findAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
			return nil, errors.New("query failure")
		},
	}

	svc := service.NewFaceService(mock)
	_, err := svc.ListAll(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFaceService_ListByType_Success(t *testing.T) {
	expected := []model.FaceRead{sampleFaceRead(model.FaceTypeWA40Full, "WA 40cm Full")}
	mock := &mockFaceRepo{
		findByTypeFn: func(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
			if faceType == model.FaceTypeWA40Full {
				return expected, nil
			}
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	faces, err := svc.ListByType(context.Background(), model.FaceTypeWA40Full)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != 1 {
		t.Fatalf("expected 1 face, got %d", len(faces))
	}
}

func TestFaceService_ListByType_InvalidTypeReturnsValidationError(t *testing.T) {
	mock := &mockFaceRepo{}
	svc := service.NewFaceService(mock)

	_, err := svc.ListByType(context.Background(), model.FaceType("invalid_type"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestFaceService_ListByType_NilReturnsEmptySlice(t *testing.T) {
	mock := &mockFaceRepo{
		findByTypeFn: func(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	faces, err := svc.ListByType(context.Background(), model.FaceTypeWA80Full)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if faces == nil || len(faces) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v", faces)
	}
}

func TestFaceService_ListByType_RepoError(t *testing.T) {
	mock := &mockFaceRepo{
		findByTypeFn: func(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
			return nil, errors.New("lookup failure")
		},
	}

	svc := service.NewFaceService(mock)
	_, err := svc.ListByType(context.Background(), model.FaceTypeWA40Full)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
```

- [x] **Step 2: Run test to verify it fails compilation**

```bash
cd backend && go test ./internal/service/... -v -run TestFaceService
```
Expected: FAIL due to undefined `service.FaceRepository` and `service.FaceService`.

- [x] **Step 3: Write minimal implementation in `backend/internal/service/face.go`**

Create `backend/internal/service/face.go`:

```go
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// FaceRepository defines persistence operations required by FaceService.
type FaceRepository interface {
	FindByID(ctx context.Context, id string) (*model.FaceRead, error)
	FindAll(ctx context.Context) ([]model.FaceRead, error)
	FindByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error)
}

// FaceService encapsulates business logic and catalog queries for target face definitions.
type FaceService struct {
	repo FaceRepository
}

// NewFaceService constructs a FaceService with repository dependency injection.
func NewFaceService(repo FaceRepository) *FaceService {
	return &FaceService{repo: repo}
}

// GetByID retrieves a target face definition by its string identifier (face_type).
// Returns apperror.ErrNotFound if no matching face exists.
func (s *FaceService) GetByID(ctx context.Context, id string) (*model.FaceRead, error) {
	if strings.TrimSpace(id) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "face id is required")
	}

	face, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching face: %w", err)
	}
	if face == nil {
		return nil, apperror.ErrNotFound
	}
	return face, nil
}

// ListAll returns all available target face definitions in the catalog.
func (s *FaceService) ListAll(ctx context.Context) ([]model.FaceRead, error) {
	faces, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing faces: %w", err)
	}
	if faces == nil {
		return []model.FaceRead{}, nil
	}
	return faces, nil
}

// ListByType retrieves target face definitions matching the provided face type.
// Returns apperror.ErrValidation if faceType is invalid.
func (s *FaceService) ListByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
	if !isValidFaceType(faceType) {
		return nil, apperror.Wrap(apperror.ErrValidation, "invalid face type")
	}

	faces, err := s.repo.FindByType(ctx, faceType)
	if err != nil {
		return nil, fmt.Errorf("listing faces by type: %w", err)
	}
	if faces == nil {
		return []model.FaceRead{}, nil
	}
	return faces, nil
}
```

- [x] **Step 4: Run test to verify it passes**

```bash
cd backend && go test -race ./internal/service/... -v -run TestFaceService
```
Expected: PASS for all `TestFaceService_*` tests.

- [x] **Step 5: Commit**

```bash
git add backend/internal/service/face.go backend/internal/service/face_test.go
git commit -m "feat(service): add face service with catalog query methods and unit tests"
```

---

### Task 3: Target Service Tests & Implementation (`target_test.go` & `target.go`)

**Files:**
- Create: `backend/internal/service/target_test.go`
- Create: `backend/internal/service/target.go`

**Interfaces:**
- Consumes:
  - `model.TargetRead`, `model.TargetCreate`, `model.TargetSet`, `model.TargetFilter` from `backend/internal/model/target.go`
  - `apperror.ErrNotFound`, `apperror.ErrValidation`, `apperror.Wrap` from `backend/internal/apperror/errors.go`
- Produces:
  - `TargetRepository` interface:
    - `FindByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error)`
    - `FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error)`
    - `FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error)`
    - `Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`
  - `TargetService` struct and constructor `NewTargetService(repo TargetRepository) *TargetService`
  - Methods:
    - `GetByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error)`
    - `ListBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error)`
    - `ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error)`
    - `Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, id uuid.UUID, data model.TargetSet) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`

- [x] **Step 1: Write failing tests in `backend/internal/service/target_test.go`**

Create `backend/internal/service/target_test.go`:

```go
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
	"github.com/jpmolinamatute/arch-stats/backend/internal/service"
)

var (
	_ service.TargetRepository = (*mockTargetRepo)(nil)
	_ service.TargetRepository = (*repository.TargetRepo)(nil)
)

type mockTargetRepo struct {
	findByIDFn        func(ctx context.Context, id uuid.UUID) (*model.TargetRead, error)
	findBySlotIDFn    func(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error)
	findBySessionIDFn func(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error)
	createFn          func(ctx context.Context, data model.TargetCreate) (uuid.UUID, error)
	updateFn          func(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error
	deleteFn          func(ctx context.Context, id uuid.UUID) error
}

func (m *mockTargetRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockTargetRepo) FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error) {
	if m.findBySlotIDFn != nil {
		return m.findBySlotIDFn(ctx, slotID)
	}
	return nil, nil
}

func (m *mockTargetRepo) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error) {
	if m.findBySessionIDFn != nil {
		return m.findBySessionIDFn(ctx, sessionID)
	}
	return nil, nil
}

func (m *mockTargetRepo) Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, nil
}

func (m *mockTargetRepo) Update(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, data, filter)
	}
	return nil
}

func (m *mockTargetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func sampleTargetRead(id, sessionID uuid.UUID, distance, lane int) model.TargetRead {
	return model.TargetRead{
		TargetID:  id,
		SessionID: sessionID,
		Distance:  distance,
		Lane:      lane,
		CreatedAt: time.Now().UTC(),
	}
}

func validTargetCreate(sessionID uuid.UUID) model.TargetCreate {
	return model.TargetCreate{
		SessionID: sessionID,
		Distance:  18,
		Lane:      1,
	}
}

func TestTargetService_GetByID_Success(t *testing.T) {
	targetID := uuid.New()
	sessionID := uuid.New()
	expected := sampleTargetRead(targetID, sessionID, 18, 1)

	mock := &mockTargetRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
			if id == targetID {
				return &expected, nil
			}
			return nil, nil
		},
	}

	svc := service.NewTargetService(mock)
	target, err := svc.GetByID(context.Background(), targetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == nil || target.TargetID != targetID {
		t.Fatalf("unexpected target: %+v", target)
	}
}

func TestTargetService_GetByID_NilIDReturnsValidationError(t *testing.T) {
	mock := &mockTargetRepo{}
	svc := service.NewTargetService(mock)

	_, err := svc.GetByID(context.Background(), uuid.Nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestTargetService_GetByID_NotFound(t *testing.T) {
	mock := &mockTargetRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
			return nil, nil
		},
	}

	svc := service.NewTargetService(mock)
	_, err := svc.GetByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTargetService_GetByID_RepoError(t *testing.T) {
	mock := &mockTargetRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
			return nil, errors.New("db error")
		},
	}

	svc := service.NewTargetService(mock)
	_, err := svc.GetByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetService_ListBySlotID_Success(t *testing.T) {
	slotID := uuid.New()
	expected := []model.TargetRead{sampleTargetRead(uuid.New(), uuid.New(), 70, 2)}

	mock := &mockTargetRepo{
		findBySlotIDFn: func(ctx context.Context, id uuid.UUID) ([]model.TargetRead, error) {
			if id == slotID {
				return expected, nil
			}
			return nil, nil
		},
	}

	svc := service.NewTargetService(mock)
	targets, err := svc.ListBySlotID(context.Background(), slotID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
}

func TestTargetService_ListBySlotID_NilIDReturnsValidationError(t *testing.T) {
	mock := &mockTargetRepo{}
	svc := service.NewTargetService(mock)

	_, err := svc.ListBySlotID(context.Background(), uuid.Nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestTargetService_ListBySlotID_NilReturnsEmptySlice(t *testing.T) {
	mock := &mockTargetRepo{
		findBySlotIDFn: func(ctx context.Context, id uuid.UUID) ([]model.TargetRead, error) {
			return nil, nil
		},
	}

	svc := service.NewTargetService(mock)
	targets, err := svc.ListBySlotID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targets == nil || len(targets) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v", targets)
	}
}

func TestTargetService_ListBySlotID_RepoError(t *testing.T) {
	mock := &mockTargetRepo{
		findBySlotIDFn: func(ctx context.Context, id uuid.UUID) ([]model.TargetRead, error) {
			return nil, errors.New("lookup failure")
		},
	}

	svc := service.NewTargetService(mock)
	_, err := svc.ListBySlotID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetService_ListBySessionID_Success(t *testing.T) {
	sessionID := uuid.New()
	expected := []model.TargetRead{
		sampleTargetRead(uuid.New(), sessionID, 18, 1),
		sampleTargetRead(uuid.New(), sessionID, 18, 2),
	}

	mock := &mockTargetRepo{
		findBySessionIDFn: func(ctx context.Context, id uuid.UUID) ([]model.TargetRead, error) {
			if id == sessionID {
				return expected, nil
			}
			return nil, nil
		},
	}

	svc := service.NewTargetService(mock)
	targets, err := svc.ListBySessionID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestTargetService_ListBySessionID_NilIDReturnsValidationError(t *testing.T) {
	mock := &mockTargetRepo{}
	svc := service.NewTargetService(mock)

	_, err := svc.ListBySessionID(context.Background(), uuid.Nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestTargetService_ListBySessionID_NilReturnsEmptySlice(t *testing.T) {
	mock := &mockTargetRepo{
		findBySessionIDFn: func(ctx context.Context, id uuid.UUID) ([]model.TargetRead, error) {
			return nil, nil
		},
	}

	svc := service.NewTargetService(mock)
	targets, err := svc.ListBySessionID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targets == nil || len(targets) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v", targets)
	}
}

func TestTargetService_ListBySessionID_RepoError(t *testing.T) {
	mock := &mockTargetRepo{
		findBySessionIDFn: func(ctx context.Context, id uuid.UUID) ([]model.TargetRead, error) {
			return nil, errors.New("lookup failure")
		},
	}

	svc := service.NewTargetService(mock)
	_, err := svc.ListBySessionID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetService_Create_Success(t *testing.T) {
	sessionID := uuid.New()
	newID := uuid.New()
	mock := &mockTargetRepo{
		createFn: func(ctx context.Context, data model.TargetCreate) (uuid.UUID, error) {
			if data.SessionID == sessionID && data.Distance == 18 && data.Lane == 1 {
				return newID, nil
			}
			return uuid.Nil, errors.New("unexpected payload")
		},
	}

	svc := service.NewTargetService(mock)
	id, err := svc.Create(context.Background(), validTargetCreate(sessionID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != newID {
		t.Fatalf("expected id %s, got %s", newID, id)
	}
}

func TestTargetService_Create_ValidationFailures(t *testing.T) {
	mock := &mockTargetRepo{}
	svc := service.NewTargetService(mock)

	tests := []struct {
		name string
		data model.TargetCreate
	}{
		{
			name: "nil session_id",
			data: model.TargetCreate{SessionID: uuid.Nil, Distance: 18, Lane: 1},
		},
		{
			name: "distance too low",
			data: model.TargetCreate{SessionID: uuid.New(), Distance: 0, Lane: 1},
		},
		{
			name: "distance too high",
			data: model.TargetCreate{SessionID: uuid.New(), Distance: 101, Lane: 1},
		},
		{
			name: "lane too low",
			data: model.TargetCreate{SessionID: uuid.New(), Distance: 18, Lane: 0},
		},
		{
			name: "lane too high",
			data: model.TargetCreate{SessionID: uuid.New(), Distance: 18, Lane: 101},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.data)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !errors.Is(err, apperror.ErrValidation) {
				t.Fatalf("expected ErrValidation, got %v", err)
			}
		})
	}
}

func TestTargetService_Create_RepoError(t *testing.T) {
	mock := &mockTargetRepo{
		createFn: func(ctx context.Context, data model.TargetCreate) (uuid.UUID, error) {
			return uuid.Nil, errors.New("insert failed")
		},
	}

	svc := service.NewTargetService(mock)
	_, err := svc.Create(context.Background(), validTargetCreate(uuid.New()))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetService_Update_Success(t *testing.T) {
	targetID := uuid.New()
	sessionID := uuid.New()
	existing := sampleTargetRead(targetID, sessionID, 18, 1)
	newDistance := 25
	newLane := 2

	mock := &mockTargetRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
			if id == targetID {
				return &existing, nil
			}
			return nil, nil
		},
		updateFn: func(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error {
			if filter.TargetID != nil && *filter.TargetID == targetID && *data.Distance == 25 && *data.Lane == 2 {
				return nil
			}
			return errors.New("unexpected update arguments")
		},
	}

	svc := service.NewTargetService(mock)
	err := svc.Update(context.Background(), targetID, model.TargetSet{
		Distance: &newDistance,
		Lane:     &newLane,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTargetService_Update_NilIDReturnsValidationError(t *testing.T) {
	mock := &mockTargetRepo{}
	svc := service.NewTargetService(mock)

	dist := 18
	err := svc.Update(context.Background(), uuid.Nil, model.TargetSet{Distance: &dist})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestTargetService_Update_ValidationFailures(t *testing.T) {
	mock := &mockTargetRepo{}
	svc := service.NewTargetService(mock)

	invalidDistLow := 0
	invalidDistHigh := 101
	invalidLaneLow := 0
	invalidLaneHigh := 101

	tests := []struct {
		name string
		data model.TargetSet
	}{
		{name: "distance too low", data: model.TargetSet{Distance: &invalidDistLow}},
		{name: "distance too high", data: model.TargetSet{Distance: &invalidDistHigh}},
		{name: "lane too low", data: model.TargetSet{Lane: &invalidLaneLow}},
		{name: "lane too high", data: model.TargetSet{Lane: &invalidLaneHigh}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.Update(context.Background(), uuid.New(), tc.data)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, apperror.ErrValidation) {
				t.Fatalf("expected ErrValidation, got %v", err)
			}
		})
	}
}

func TestTargetService_Update_NotFound(t *testing.T) {
	mock := &mockTargetRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
			return nil, nil
		},
	}

	svc := service.NewTargetService(mock)
	dist := 25
	err := svc.Update(context.Background(), uuid.New(), model.TargetSet{Distance: &dist})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTargetService_Update_EmptyDataIsNoOp(t *testing.T) {
	targetID := uuid.New()
	existing := sampleTargetRead(targetID, uuid.New(), 18, 1)

	mock := &mockTargetRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
			return &existing, nil
		},
		updateFn: func(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error {
			t.Fatal("Update should not be called when data is empty")
			return nil
		},
	}

	svc := service.NewTargetService(mock)
	err := svc.Update(context.Background(), targetID, model.TargetSet{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTargetService_Update_FindByIDError(t *testing.T) {
	mock := &mockTargetRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
			return nil, errors.New("read error")
		},
	}

	svc := service.NewTargetService(mock)
	dist := 20
	err := svc.Update(context.Background(), uuid.New(), model.TargetSet{Distance: &dist})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetService_Update_RepoError(t *testing.T) {
	targetID := uuid.New()
	existing := sampleTargetRead(targetID, uuid.New(), 18, 1)

	mock := &mockTargetRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
			return &existing, nil
		},
		updateFn: func(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error {
			return errors.New("update failed")
		},
	}

	svc := service.NewTargetService(mock)
	dist := 20
	err := svc.Update(context.Background(), targetID, model.TargetSet{Distance: &dist})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetService_Delete_Success(t *testing.T) {
	targetID := uuid.New()
	mock := &mockTargetRepo{
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			if id == targetID {
				return nil
			}
			return errors.New("unexpected id")
		},
	}

	svc := service.NewTargetService(mock)
	err := svc.Delete(context.Background(), targetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTargetService_Delete_NilIDReturnsValidationError(t *testing.T) {
	mock := &mockTargetRepo{}
	svc := service.NewTargetService(mock)

	err := svc.Delete(context.Background(), uuid.Nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestTargetService_Delete_NotFound(t *testing.T) {
	mock := &mockTargetRepo{
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return apperror.ErrNotFound
		},
	}

	svc := service.NewTargetService(mock)
	err := svc.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTargetService_Delete_RepoError(t *testing.T) {
	mock := &mockTargetRepo{
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("delete failed")
		},
	}

	svc := service.NewTargetService(mock)
	err := svc.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
```

- [x] **Step 2: Run test to verify it fails compilation**

```bash
cd backend && go test ./internal/service/... -v -run TestTargetService
```
Expected: FAIL due to undefined `service.TargetRepository` and `service.TargetService`.

- [x] **Step 3: Write minimal implementation in `backend/internal/service/target.go`**

Create `backend/internal/service/target.go`:

```go
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// TargetRepository defines persistence operations required by TargetService.
type TargetRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error)
	FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error)
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error)
	Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TargetService encapsulates business logic, validation, and operations for lane targets.
type TargetService struct {
	repo TargetRepository
}

// NewTargetService constructs a TargetService with repository dependency injection.
func NewTargetService(repo TargetRepository) *TargetService {
	return &TargetService{repo: repo}
}

// GetByID retrieves a lane target configuration by its primary key identifier.
// Returns apperror.ErrNotFound if the target does not exist.
func (s *TargetService) GetByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "target id is required")
	}

	target, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching target: %w", err)
	}
	if target == nil {
		return nil, apperror.ErrNotFound
	}
	return target, nil
}

// ListBySlotID retrieves target configurations associated with a specific slot.
func (s *TargetService) ListBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error) {
	if slotID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	targets, err := s.repo.FindBySlotID(ctx, slotID)
	if err != nil {
		return nil, fmt.Errorf("listing targets by slot: %w", err)
	}
	if targets == nil {
		return []model.TargetRead{}, nil
	}
	return targets, nil
}

// ListBySessionID retrieves target configurations associated with a specific session.
func (s *TargetService) ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error) {
	if sessionID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	targets, err := s.repo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("listing targets by session: %w", err)
	}
	if targets == nil {
		return []model.TargetRead{}, nil
	}
	return targets, nil
}

// Create validates incoming target fields and persists the record.
func (s *TargetService) Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error) {
	if err := validateTargetCreate(data); err != nil {
		return uuid.Nil, err
	}

	id, err := s.repo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating target: %w", err)
	}
	return id, nil
}

// Update validates mutation fields, verifies target existence, and updates the target record.
// Returns apperror.ErrNotFound if the target does not exist.
func (s *TargetService) Update(ctx context.Context, id uuid.UUID, data model.TargetSet) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "target id is required")
	}

	if err := validateTargetSet(data); err != nil {
		return err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying target existence: %w", err)
	}
	if existing == nil {
		return apperror.ErrNotFound
	}

	if data.Distance == nil && data.Lane == nil {
		return nil
	}

	filter := model.TargetFilter{TargetID: &id}
	if err := s.repo.Update(ctx, data, filter); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("updating target: %w", err)
	}
	return nil
}

// Delete removes a target configuration by its primary key identifier.
// Returns apperror.ErrNotFound if the target does not exist.
func (s *TargetService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "target id is required")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("deleting target: %w", err)
	}
	return nil
}

func validateTargetCreate(data model.TargetCreate) error {
	if data.SessionID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session_id is required")
	}
	if data.Distance < 1 || data.Distance > 100 {
		return apperror.Wrap(apperror.ErrValidation, "distance must be between 1 and 100")
	}
	if data.Lane < 1 || data.Lane > 100 {
		return apperror.Wrap(apperror.ErrValidation, "lane must be between 1 and 100")
	}
	return nil
}

func validateTargetSet(data model.TargetSet) error {
	if data.Distance != nil && (*data.Distance < 1 || *data.Distance > 100) {
		return apperror.Wrap(apperror.ErrValidation, "distance must be between 1 and 100")
	}
	if data.Lane != nil && (*data.Lane < 1 || *data.Lane > 100) {
		return apperror.Wrap(apperror.ErrValidation, "lane must be between 1 and 100")
	}
	return nil
}
```

- [x] **Step 4: Run test to verify it passes**

```bash
cd backend && go test -race ./internal/service/... -v -run TestTargetService
```
Expected: PASS for all `TestTargetService_*` tests.

- [x] **Step 5: Commit**

```bash
git add backend/internal/service/target.go backend/internal/service/target_test.go
git commit -m "feat(service): add target service with lane validation and unit tests"
```

---

### Task 4: Full Suite Verification, Linting, & Docs Update

**Files:**
- Modify: `docs/go_refactor/tasks/016-service_layer_face_and_target.md`

**Interfaces:**
- Consumes: All packages and files in `backend/internal/...`
- Produces: Clean lint, clean vet, passing test suite, updated task status

- [x] **Step 1: Run the full test suite with race detector**

```bash
cd backend && go test -race ./... -v
```
Expected: PASS across all packages (`apperror`, `config`, `model`, `repository`, `service`).

- [x] **Step 2: Run `go vet` and compile `backend/cmd/arch-stats`**

```bash
cd backend && go vet ./... && go build ./...
```
Expected: No vet warnings and clean build.

- [x] **Step 3: Run Go linting**

```bash
./scripts/linting.bash --go
```
Expected: `golangci-lint` passes without errors.

- [x] **Step 4: Update `docs/go_refactor/tasks/016-service_layer_face_and_target.md`**

Mark all checklist items in `docs/go_refactor/tasks/016-service_layer_face_and_target.md` as completed (`[x]`).

- [x] **Step 5: Commit**

```bash
git add docs/go_refactor/tasks/016-service_layer_face_and_target.md
git commit -m "docs: mark task 016 face and target service layers as complete"
```

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-09-05-service-layer-face-and-target.md`. Two execution options:

1. **Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. **Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
