# Task 015: Build Service Layer — Slot and Shot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the business logic service layer for shooting slot assignments (`SlotService`) and arrow shot records (`ShotService`) with repository interface dependency injection, input validation, state validation (ensuring slots belong to open sessions and shots belong to valid slots), score constraints, and comprehensive unit tests.

**Architecture:** 
- `SlotService` sits between HTTP handlers and `SlotRepository` / `SessionRepository`. It validates incoming data payloads (`model.SlotCreate`, `model.SlotSet`), ensures that the target session exists and is currently open (`session.IsOpened == true`), verifies slot existence on mutation, and delegates persistence to `SlotRepository`.
- `ShotService` sits between HTTP handlers and `ShotRepository` / `SlotRepository`. It validates incoming shot records (`model.ShotCreate`, `model.ShotSet`), enforces coordinate/score consistency (all-or-none), score bounds (0–10), `is_x` constraints (`is_x == true` requires `score == 10`), ensures the target slot exists, and delegates persistence to `ShotRepository`.
- Both services consume repository interfaces defined directly in the service package, enabling unit testing with mock repositories without requiring database connections or pools.

**Tech Stack:** Go 1.27+, `github.com/google/uuid`, standard library (`context`, `errors`, `fmt`), internal packages (`model`, `apperror`, `repository`).

**Spec:** [docs/go_refactor/tasks/015-service_layer_slot_and_shot.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/015-service_layer_slot_and_shot.md)

## Global Constraints

- Git branch: `refactor/015-service-layer-slot-and-shot`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/service`
- Error handling: Wrap internal errors with `%w` using contextual descriptive messages (`fmt.Errorf("...: %w", err)`). Return sentinel `apperror.ErrNotFound` and `apperror.Wrap(apperror.ErrValidation, ...)` as appropriate.
- Dependency injection: Services must accept repository interfaces in their constructors (`NewSlotService(slotRepo SlotRepository, sessionRepo SessionRepository)`, `NewShotService(shotRepo ShotRepository, slotRepo SlotRepository)`).
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
        ├── slot.go               # [NEW] SlotRepository interface, SlotService struct, constructor, CRUD & validation methods
        ├── slot_test.go          # [NEW] mockSlotRepo, mock SessionRepository usage, and unit test suite for SlotService
        ├── shot.go               # [NEW] ShotRepository interface, ShotService struct, constructor, CRUD & validation methods
        └── shot_test.go          # [NEW] mockShotRepo, mock SlotRepository usage, and unit test suite for ShotService
```

---

### Task 1: Git Branch Setup

**Files:**
- Modify: git branch switch

**Interfaces:**
- Consumes: `main` branch
- Produces: `refactor/015-service-layer-slot-and-shot` branch

- [ ] **Step 1: Check out git branch**

```bash
git switch -c refactor/015-service-layer-slot-and-shot
```

- [ ] **Step 2: Verify git status is clean**

```bash
git status
```
Expected: On branch `refactor/015-service-layer-slot-and-shot`, working tree clean.

---

### Task 2: Slot Service Tests & Implementation (`slot_test.go` & `slot.go`)

**Files:**
- Create: `backend/internal/service/slot_test.go`
- Create: `backend/internal/service/slot.go`

**Interfaces:**
- Consumes:
  - `model.SlotRead`, `model.SlotCreate`, `model.SlotSet`, `model.SlotFilter` from `backend/internal/model/slot.go`
  - `model.SlotLetter`, `model.FaceType`, `model.Bowstyle` from `backend/internal/model/enums.go`
  - `model.SessionRead` from `backend/internal/model/session.go`
  - `service.SessionRepository` from `backend/internal/service/session.go`
  - `apperror.ErrNotFound`, `apperror.ErrValidation`, `apperror.Wrap` from `backend/internal/apperror/errors.go`
- Produces:
  - `SlotRepository` interface:
    - `FindByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error)`
    - `FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error)`
    - `FindAll(ctx context.Context, filter model.SlotFilter) ([]model.SlotRead, error)`
    - `Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, data model.SlotSet, filter model.SlotFilter) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`
  - `SlotService` struct and constructor `NewSlotService(slotRepo SlotRepository, sessionRepo SessionRepository) *SlotService`
  - Methods:
    - `GetByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error)`
    - `ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error)`
    - `Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, id uuid.UUID, data model.SlotSet) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`

- [ ] **Step 1: Write failing tests in `backend/internal/service/slot_test.go`**

Create `backend/internal/service/slot_test.go`:

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
	_ service.SlotRepository = (*mockSlotRepo)(nil)
	_ service.SlotRepository = (*repository.SlotRepo)(nil)
)

type mockSlotRepo struct {
	findByIDFn        func(ctx context.Context, id uuid.UUID) (*model.SlotRead, error)
	findBySessionIDFn func(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error)
	findAllFn         func(ctx context.Context, filter model.SlotFilter) ([]model.SlotRead, error)
	createFn          func(ctx context.Context, data model.SlotCreate) (uuid.UUID, error)
	updateFn          func(ctx context.Context, data model.SlotSet, filter model.SlotFilter) error
	deleteFn          func(ctx context.Context, id uuid.UUID) error
}

func (m *mockSlotRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSlotRepo) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error) {
	if m.findBySessionIDFn != nil {
		return m.findBySessionIDFn(ctx, sessionID)
	}
	return nil, nil
}

func (m *mockSlotRepo) FindAll(ctx context.Context, filter model.SlotFilter) ([]model.SlotRead, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockSlotRepo) Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, nil
}

func (m *mockSlotRepo) Update(ctx context.Context, data model.SlotSet, filter model.SlotFilter) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, data, filter)
	}
	return nil
}

func (m *mockSlotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func sampleSlotRead(id, sessionID, archerID, targetID uuid.UUID) model.SlotRead {
	now := time.Now().UTC()
	slotCode := "1A"
	shots := 6
	return model.SlotRead{
		SlotID:          id,
		TargetID:        targetID,
		ArcherID:        archerID,
		SessionID:       sessionID,
		SlotLetter:      model.SlotLetterA,
		Slot:            &slotCode,
		FaceType:        model.FaceTypeWA40Full,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      40.0,
		IsShooting:      true,
		ShotPerRound:    &shots,
		IntervalSeconds: 20,
		CreatedAt:       &now,
	}
}

func validSlotCreate(sessionID, archerID, targetID uuid.UUID) model.SlotCreate {
	shots := 6
	return model.SlotCreate{
		ArcherID:        archerID,
		SessionID:       sessionID,
		TargetID:        targetID,
		SlotLetter:      model.SlotLetterA,
		FaceType:        model.FaceTypeWA40Full,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      40.0,
		IsShooting:      true,
		ShotPerRound:    &shots,
		IntervalSeconds: 20,
	}
}

func TestSlotService_GetByID(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	sessionID := uuid.New()
	archerID := uuid.New()
	targetID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := sampleSlotRead(testID, sessionID, archerID, targetID)
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SlotRead, error) {
				if id == testID {
					return &expected, nil
				}
				return nil, nil
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		got, err := svc.GetByID(ctx, testID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SlotID != testID || got.SlotLetter != expected.SlotLetter {
			t.Errorf("got %+v, expected %+v", got, expected)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SlotRead, error) {
				return nil, nil
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		got, err := svc.GetByID(ctx, testID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil result, got %+v", got)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		svc := service.NewSlotService(&mockSlotRepo{}, &mockSessionRepo{})

		_, err := svc.GetByID(ctx, uuid.Nil)
		if err == nil {
			t.Fatal("expected error for nil UUID, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("db connection down")
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SlotRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		_, err := svc.GetByID(ctx, testID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestSlotService_ListBySessionID(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()

	t.Run("success with items", func(t *testing.T) {
		expected := []model.SlotRead{
			sampleSlotRead(uuid.New(), sessionID, uuid.New(), uuid.New()),
			sampleSlotRead(uuid.New(), sessionID, uuid.New(), uuid.New()),
		}
		slotRepo := &mockSlotRepo{
			findBySessionIDFn: func(_ context.Context, sID uuid.UUID) ([]model.SlotRead, error) {
				if sID == sessionID {
					return expected, nil
				}
				return nil, nil
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		got, err := svc.ListBySessionID(ctx, sessionID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(expected) {
			t.Errorf("got %d slots, expected %d", len(got), len(expected))
		}
	})

	t.Run("returns empty slice when repo returns nil", func(t *testing.T) {
		slotRepo := &mockSlotRepo{
			findBySessionIDFn: func(_ context.Context, _ uuid.UUID) ([]model.SlotRead, error) {
				return nil, nil
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		got, err := svc.ListBySessionID(ctx, sessionID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("expected empty slice, got %+v", got)
		}
	})

	t.Run("validates empty session UUID", func(t *testing.T) {
		svc := service.NewSlotService(&mockSlotRepo{}, &mockSessionRepo{})

		_, err := svc.ListBySessionID(ctx, uuid.Nil)
		if err == nil {
			t.Fatal("expected error for nil UUID, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("db query failed")
		slotRepo := &mockSlotRepo{
			findBySessionIDFn: func(_ context.Context, _ uuid.UUID) ([]model.SlotRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		_, err := svc.ListBySessionID(ctx, sessionID)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestSlotService_Create(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	archerID := uuid.New()
	targetID := uuid.New()
	newSlotID := uuid.New()

	openSession := sampleSessionRead(sessionID, archerID, true)
	closedSession := sampleSessionRead(sessionID, archerID, false)

	t.Run("success when session is open", func(t *testing.T) {
		sessionRepo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
				if id == sessionID {
					return &openSession, nil
				}
				return nil, nil
			},
		}
		slotRepo := &mockSlotRepo{
			createFn: func(_ context.Context, d model.SlotCreate) (uuid.UUID, error) {
				if d.SessionID == sessionID && d.ArcherID == archerID {
					return newSlotID, nil
				}
				return uuid.Nil, errors.New("unexpected payload")
			},
		}
		svc := service.NewSlotService(slotRepo, sessionRepo)

		id, err := svc.Create(ctx, validSlotCreate(sessionID, archerID, targetID))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != newSlotID {
			t.Errorf("got id %v, expected %v", id, newSlotID)
		}
	})

	t.Run("returns validation error when session is not open", func(t *testing.T) {
		sessionRepo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
				if id == sessionID {
					return &closedSession, nil
				}
				return nil, nil
			},
		}
		slotRepo := &mockSlotRepo{}
		svc := service.NewSlotService(slotRepo, sessionRepo)

		_, err := svc.Create(ctx, validSlotCreate(sessionID, archerID, targetID))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("returns not found when session does not exist", func(t *testing.T) {
		sessionRepo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, nil
			},
		}
		slotRepo := &mockSlotRepo{}
		svc := service.NewSlotService(slotRepo, sessionRepo)

		_, err := svc.Create(ctx, validSlotCreate(sessionID, archerID, targetID))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		invalidShotsLow := 2
		invalidShotsHigh := 11

		tests := []struct {
			name   string
			mutate func(d *model.SlotCreate)
		}{
			{"empty archer id", func(d *model.SlotCreate) { d.ArcherID = uuid.Nil }},
			{"empty session id", func(d *model.SlotCreate) { d.SessionID = uuid.Nil }},
			{"empty target id", func(d *model.SlotCreate) { d.TargetID = uuid.Nil }},
			{"invalid slot letter", func(d *model.SlotCreate) { d.SlotLetter = model.SlotLetter("E") }},
			{"invalid face type", func(d *model.SlotCreate) { d.FaceType = model.FaceType("unknown_face") }},
			{"invalid bowstyle", func(d *model.SlotCreate) { d.Bowstyle = model.Bowstyle("slingshot") }},
			{"draw weight zero", func(d *model.SlotCreate) { d.DrawWeight = 0 }},
			{"draw weight negative", func(d *model.SlotCreate) { d.DrawWeight = -10 }},
			{"draw weight excessive", func(d *model.SlotCreate) { d.DrawWeight = 250 }},
			{"interval seconds zero", func(d *model.SlotCreate) { d.IntervalSeconds = 0 }},
			{"interval seconds excessive", func(d *model.SlotCreate) { d.IntervalSeconds = 150 }},
			{"shot per round low", func(d *model.SlotCreate) { d.ShotPerRound = &invalidShotsLow }},
			{"shot per round high", func(d *model.SlotCreate) { d.ShotPerRound = &invalidShotsHigh }},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				data := validSlotCreate(sessionID, archerID, targetID)
				tc.mutate(&data)

				svc := service.NewSlotService(&mockSlotRepo{}, &mockSessionRepo{})

				_, err := svc.Create(ctx, data)
				if err == nil {
					t.Fatalf("expected validation error for %s, got nil", tc.name)
				}
				if !errors.Is(err, apperror.ErrValidation) {
					t.Errorf("expected ErrValidation, got: %v", err)
				}
			})
		}
	})

	t.Run("propagates session check error", func(t *testing.T) {
		sessionErr := errors.New("db session error")
		sessionRepo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, sessionErr
			},
		}
		svc := service.NewSlotService(&mockSlotRepo{}, sessionRepo)

		_, err := svc.Create(ctx, validSlotCreate(sessionID, archerID, targetID))
		if !errors.Is(err, sessionErr) {
			t.Errorf("expected wrapped sessionErr, got: %v", err)
		}
	})

	t.Run("propagates slot create error", func(t *testing.T) {
		createErr := errors.New("db slot insert error")
		sessionRepo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return &openSession, nil
			},
		}
		slotRepo := &mockSlotRepo{
			createFn: func(_ context.Context, _ model.SlotCreate) (uuid.UUID, error) {
				return uuid.Nil, createErr
			},
		}
		svc := service.NewSlotService(slotRepo, sessionRepo)

		_, err := svc.Create(ctx, validSlotCreate(sessionID, archerID, targetID))
		if !errors.Is(err, createErr) {
			t.Errorf("expected wrapped createErr, got: %v", err)
		}
	})
}

func TestSlotService_Update(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	sessionID := uuid.New()
	archerID := uuid.New()
	targetID := uuid.New()

	existing := sampleSlotRead(testID, sessionID, archerID, targetID)
	newShooting := false

	t.Run("success", func(t *testing.T) {
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SlotRead, error) {
				if id == testID {
					return &existing, nil
				}
				return nil, nil
			},
			updateFn: func(_ context.Context, data model.SlotSet, filter model.SlotFilter) error {
				if *filter.SlotID != testID {
					return errors.New("id mismatch")
				}
				if *data.IsShooting != newShooting {
					return errors.New("data mismatch")
				}
				return nil
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		err := svc.Update(ctx, testID, model.SlotSet{IsShooting: &newShooting})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SlotRead, error) {
				return nil, nil
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		err := svc.Update(ctx, testID, model.SlotSet{IsShooting: &newShooting})
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		svc := service.NewSlotService(&mockSlotRepo{}, &mockSessionRepo{})

		err := svc.Update(ctx, uuid.Nil, model.SlotSet{IsShooting: &newShooting})
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("validates invalid fields", func(t *testing.T) {
		invalidFace := model.FaceType("laser_face")
		invalidLetter := model.SlotLetter("Z")
		invalidShots := 15
		invalidInterval := 200

		tests := []struct {
			name string
			data model.SlotSet
		}{
			{"invalid face type", model.SlotSet{FaceType: &invalidFace}},
			{"invalid slot letter", model.SlotSet{SlotLetter: &invalidLetter}},
			{"invalid shots per round", model.SlotSet{ShotPerRound: &invalidShots}},
			{"invalid interval seconds", model.SlotSet{IntervalSeconds: &invalidInterval}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				svc := service.NewSlotService(&mockSlotRepo{}, &mockSessionRepo{})

				err := svc.Update(ctx, testID, tc.data)
				if !errors.Is(err, apperror.ErrValidation) {
					t.Errorf("expected ErrValidation for %s, got: %v", tc.name, err)
				}
			})
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("update exec failed")
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SlotRead, error) {
				return &existing, nil
			},
			updateFn: func(_ context.Context, _ model.SlotSet, _ model.SlotFilter) error {
				return repoErr
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		err := svc.Update(ctx, testID, model.SlotSet{IsShooting: &newShooting})
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestSlotService_Delete(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()

	t.Run("success", func(t *testing.T) {
		slotRepo := &mockSlotRepo{
			deleteFn: func(_ context.Context, id uuid.UUID) error {
				if id == testID {
					return nil
				}
				return apperror.ErrNotFound
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		err := svc.Delete(ctx, testID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		slotRepo := &mockSlotRepo{
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return apperror.ErrNotFound
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		err := svc.Delete(ctx, testID)
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		svc := service.NewSlotService(&mockSlotRepo{}, &mockSessionRepo{})

		err := svc.Delete(ctx, uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("delete failed")
		slotRepo := &mockSlotRepo{
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return repoErr
			},
		}
		svc := service.NewSlotService(slotRepo, &mockSessionRepo{})

		err := svc.Delete(ctx, testID)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/service/... -v -run TestSlotService
```
Expected: FAIL with compilation error (`service.NewSlotService` and `service.SlotRepository` undefined).

- [ ] **Step 3: Implement `backend/internal/service/slot.go`**

Create `backend/internal/service/slot.go`:

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

// SlotRepository defines persistence operations required by SlotService.
type SlotRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error)
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error)
	FindAll(ctx context.Context, filter model.SlotFilter) ([]model.SlotRead, error)
	Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.SlotSet, filter model.SlotFilter) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SlotService encapsulates business logic, state rules, and validation for slot assignments.
type SlotService struct {
	slotRepo    SlotRepository
	sessionRepo SessionRepository
}

// NewSlotService constructs a SlotService with repository dependency injection.
func NewSlotService(slotRepo SlotRepository, sessionRepo SessionRepository) *SlotService {
	return &SlotService{
		slotRepo:    slotRepo,
		sessionRepo: sessionRepo,
	}
}

// GetByID retrieves a slot assignment by primary key identifier.
// Returns apperror.ErrNotFound if the slot does not exist.
func (s *SlotService) GetByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	slot, err := s.slotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching slot: %w", err)
	}
	if slot == nil {
		return nil, apperror.ErrNotFound
	}
	return slot, nil
}

// ListBySessionID retrieves all slot assignments belonging to a given session.
func (s *SlotService) ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error) {
	if sessionID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	slots, err := s.slotRepo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("listing slots by session: %w", err)
	}
	if slots == nil {
		return []model.SlotRead{}, nil
	}
	return slots, nil
}

// Create validates incoming slot assignment data, verifies that the target session exists and is open,
// and persists the slot record.
// Returns apperror.ErrNotFound if the session does not exist, or apperror.ErrValidation if the session is closed.
func (s *SlotService) Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error) {
	if err := validateSlotCreate(data); err != nil {
		return uuid.Nil, err
	}

	session, err := s.sessionRepo.FindByID(ctx, data.SessionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("verifying session: %w", err)
	}
	if session == nil {
		return uuid.Nil, apperror.ErrNotFound
	}
	if !session.IsOpened {
		return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "session is not open")
	}

	id, err := s.slotRepo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating slot: %w", err)
	}
	return id, nil
}

// Update validates mutation fields and updates a slot assignment.
// Returns apperror.ErrNotFound if the slot does not exist.
func (s *SlotService) Update(ctx context.Context, id uuid.UUID, data model.SlotSet) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	if err := validateSlotSet(data); err != nil {
		return err
	}

	existing, err := s.slotRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying slot: %w", err)
	}
	if existing == nil {
		return apperror.ErrNotFound
	}

	filter := model.SlotFilter{SlotID: &id}
	if err := s.slotRepo.Update(ctx, data, filter); err != nil {
		return fmt.Errorf("updating slot: %w", err)
	}
	return nil
}

// Delete removes a slot assignment by primary key identifier.
// Returns apperror.ErrNotFound if the slot does not exist.
func (s *SlotService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	if err := s.slotRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("deleting slot: %w", err)
	}
	return nil
}

func validateSlotCreate(data model.SlotCreate) error {
	if data.ArcherID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "archer_id is required")
	}
	if data.SessionID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session_id is required")
	}
	if data.TargetID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "target_id is required")
	}
	if !isValidSlotLetter(data.SlotLetter) {
		return apperror.Wrap(apperror.ErrValidation, "invalid slot_letter")
	}
	if !isValidFaceType(data.FaceType) {
		return apperror.Wrap(apperror.ErrValidation, "invalid face_type")
	}
	if !isValidBowstyle(data.Bowstyle) {
		return apperror.Wrap(apperror.ErrValidation, "invalid bowstyle")
	}
	if data.DrawWeight <= 0 || data.DrawWeight > 200 {
		return apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}
	if data.IntervalSeconds < 1 || data.IntervalSeconds > 100 {
		return apperror.Wrap(apperror.ErrValidation, "interval_seconds must be between 1 and 100")
	}
	if data.ShotPerRound != nil && (*data.ShotPerRound < 3 || *data.ShotPerRound > 10) {
		return apperror.Wrap(apperror.ErrValidation, "shot_per_round must be between 3 and 10")
	}
	return nil
}

func validateSlotSet(data model.SlotSet) error {
	if data.FaceType != nil && !isValidFaceType(*data.FaceType) {
		return apperror.Wrap(apperror.ErrValidation, "invalid face_type")
	}
	if data.SlotLetter != nil && !isValidSlotLetter(*data.SlotLetter) {
		return apperror.Wrap(apperror.ErrValidation, "invalid slot_letter")
	}
	if data.ShotPerRound != nil && (*data.ShotPerRound < 3 || *data.ShotPerRound > 10) {
		return apperror.Wrap(apperror.ErrValidation, "shot_per_round must be between 3 and 10")
	}
	if data.IntervalSeconds != nil && (*data.IntervalSeconds < 1 || *data.IntervalSeconds > 100) {
		return apperror.Wrap(apperror.ErrValidation, "interval_seconds must be between 1 and 100")
	}
	return nil
}

func isValidSlotLetter(l model.SlotLetter) bool {
	switch l {
	case model.SlotLetterA, model.SlotLetterB, model.SlotLetterC, model.SlotLetterD:
		return true
	default:
		return false
	}
}

func isValidFaceType(f model.FaceType) bool {
	switch f {
	case model.FaceTypeWA40Full,
		model.FaceTypeWA60Full,
		model.FaceTypeWA80Full,
		model.FaceTypeWA122Full,
		model.FaceTypeWA406Rings,
		model.FaceTypeWA606Rings,
		model.FaceTypeWA806Rings,
		model.FaceTypeWA1226Rings,
		model.FaceTypeWA40TripleVertical,
		model.FaceTypeWA60TripleTriangular,
		model.FaceTypeNone:
		return true
	default:
		return false
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/service/... -v -run TestSlotService
```
Expected: PASS for all `TestSlotService_*` tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/slot.go backend/internal/service/slot_test.go
git commit -m "feat(service): add slot service with session state validation and repository injection"
```

---

### Task 3: Shot Service Tests & Implementation (`shot_test.go` & `shot.go`)

**Files:**
- Create: `backend/internal/service/shot_test.go`
- Create: `backend/internal/service/shot.go`

**Interfaces:**
- Consumes:
  - `model.ShotRead`, `model.ShotCreate`, `model.ShotSet`, `model.ShotFilter` from `backend/internal/model/shot.go`
  - `model.SlotRead` from `backend/internal/model/slot.go`
  - `service.SlotRepository` from `backend/internal/service/slot.go`
  - `apperror.ErrNotFound`, `apperror.ErrValidation`, `apperror.Wrap` from `backend/internal/apperror/errors.go`
- Produces:
  - `ShotRepository` interface:
    - `FindByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error)`
    - `FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error)`
    - `FindAll(ctx context.Context, filter model.ShotFilter) ([]model.ShotRead, error)`
    - `Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, data model.ShotSet, filter model.ShotFilter) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`
  - `ShotService` struct and constructor `NewShotService(shotRepo ShotRepository, slotRepo SlotRepository) *ShotService`
  - Methods:
    - `GetByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error)`
    - `ListBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error)`
    - `Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, id uuid.UUID, data model.ShotSet) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`

- [ ] **Step 1: Write failing tests in `backend/internal/service/shot_test.go`**

Create `backend/internal/service/shot_test.go`:

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
	_ service.ShotRepository = (*mockShotRepo)(nil)
	_ service.ShotRepository = (*repository.ShotRepo)(nil)
)

type mockShotRepo struct {
	findByIDFn     func(ctx context.Context, id uuid.UUID) (*model.ShotRead, error)
	findBySlotIDFn func(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error)
	findAllFn      func(ctx context.Context, filter model.ShotFilter) ([]model.ShotRead, error)
	createFn       func(ctx context.Context, data model.ShotCreate) (uuid.UUID, error)
	updateFn       func(ctx context.Context, data model.ShotSet, filter model.ShotFilter) error
	deleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockShotRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockShotRepo) FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error) {
	if m.findBySlotIDFn != nil {
		return m.findBySlotIDFn(ctx, slotID)
	}
	return nil, nil
}

func (m *mockShotRepo) FindAll(ctx context.Context, filter model.ShotFilter) ([]model.ShotRead, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockShotRepo) Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, nil
}

func (m *mockShotRepo) Update(ctx context.Context, data model.ShotSet, filter model.ShotFilter) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, data, filter)
	}
	return nil
}

func (m *mockShotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func sampleShotRead(id, slotID uuid.UUID, score int, isX bool) model.ShotRead {
	x := 10.5
	y := -4.2
	arrowID := uuid.New()
	return model.ShotRead{
		ShotID:    id,
		SlotID:    slotID,
		X:         &x,
		Y:         &y,
		IsX:       isX,
		Score:     &score,
		ArrowID:   &arrowID,
		CreatedAt: time.Now().UTC(),
	}
}

func validShotCreate(slotID uuid.UUID) model.ShotCreate {
	x := 5.0
	y := 8.0
	score := 9
	arrowID := uuid.New()
	return model.ShotCreate{
		SlotID:  slotID,
		X:       &x,
		Y:       &y,
		IsX:     false,
		Score:   &score,
		ArrowID: &arrowID,
	}
}

func TestShotService_GetByID(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	slotID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := sampleShotRead(testID, slotID, 10, true)
		shotRepo := &mockShotRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.ShotRead, error) {
				if id == testID {
					return &expected, nil
				}
				return nil, nil
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		got, err := svc.GetByID(ctx, testID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ShotID != testID || got.IsX != expected.IsX {
			t.Errorf("got %+v, expected %+v", got, expected)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		shotRepo := &mockShotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ShotRead, error) {
				return nil, nil
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		got, err := svc.GetByID(ctx, testID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil result, got %+v", got)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		svc := service.NewShotService(&mockShotRepo{}, &mockSlotRepo{})

		_, err := svc.GetByID(ctx, uuid.Nil)
		if err == nil {
			t.Fatal("expected error for nil UUID, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("db connection down")
		shotRepo := &mockShotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ShotRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		_, err := svc.GetByID(ctx, testID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestShotService_ListBySlotID(t *testing.T) {
	ctx := context.Background()
	slotID := uuid.New()

	t.Run("success with items", func(t *testing.T) {
		expected := []model.ShotRead{
			sampleShotRead(uuid.New(), slotID, 10, true),
			sampleShotRead(uuid.New(), slotID, 8, false),
		}
		shotRepo := &mockShotRepo{
			findBySlotIDFn: func(_ context.Context, sID uuid.UUID) ([]model.ShotRead, error) {
				if sID == slotID {
					return expected, nil
				}
				return nil, nil
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		got, err := svc.ListBySlotID(ctx, slotID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(expected) {
			t.Errorf("got %d shots, expected %d", len(got), len(expected))
		}
	})

	t.Run("returns empty slice when repo returns nil", func(t *testing.T) {
		shotRepo := &mockShotRepo{
			findBySlotIDFn: func(_ context.Context, _ uuid.UUID) ([]model.ShotRead, error) {
				return nil, nil
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		got, err := svc.ListBySlotID(ctx, slotID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("expected empty slice, got %+v", got)
		}
	})

	t.Run("validates empty slot UUID", func(t *testing.T) {
		svc := service.NewShotService(&mockShotRepo{}, &mockSlotRepo{})

		_, err := svc.ListBySlotID(ctx, uuid.Nil)
		if err == nil {
			t.Fatal("expected error for nil UUID, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("db query failed")
		shotRepo := &mockShotRepo{
			findBySlotIDFn: func(_ context.Context, _ uuid.UUID) ([]model.ShotRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		_, err := svc.ListBySlotID(ctx, slotID)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestShotService_Create(t *testing.T) {
	ctx := context.Background()
	slotID := uuid.New()
	newShotID := uuid.New()
	slotRecord := sampleSlotRead(slotID, uuid.New(), uuid.New(), uuid.New())

	t.Run("success with valid slot and complete coordinates", func(t *testing.T) {
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SlotRead, error) {
				if id == slotID {
					return &slotRecord, nil
				}
				return nil, nil
			},
		}
		shotRepo := &mockShotRepo{
			createFn: func(_ context.Context, d model.ShotCreate) (uuid.UUID, error) {
				if d.SlotID == slotID && *d.Score == 9 {
					return newShotID, nil
				}
				return uuid.Nil, errors.New("unexpected payload")
			},
		}
		svc := service.NewShotService(shotRepo, slotRepo)

		id, err := svc.Create(ctx, validShotCreate(slotID))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != newShotID {
			t.Errorf("got id %v, expected %v", id, newShotID)
		}
	})

	t.Run("success with valid slot and all nil coordinates and score", func(t *testing.T) {
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SlotRead, error) {
				return &slotRecord, nil
			},
		}
		shotRepo := &mockShotRepo{
			createFn: func(_ context.Context, d model.ShotCreate) (uuid.UUID, error) {
				if d.X == nil && d.Y == nil && d.Score == nil {
					return newShotID, nil
				}
				return uuid.Nil, errors.New("expected nil fields")
			},
		}
		svc := service.NewShotService(shotRepo, slotRepo)

		data := model.ShotCreate{SlotID: slotID}
		id, err := svc.Create(ctx, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != newShotID {
			t.Errorf("got id %v, expected %v", id, newShotID)
		}
	})

	t.Run("returns not found when slot does not exist", func(t *testing.T) {
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SlotRead, error) {
				return nil, nil
			},
		}
		shotRepo := &mockShotRepo{}
		svc := service.NewShotService(shotRepo, slotRepo)

		_, err := svc.Create(ctx, validShotCreate(slotID))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		validX := 1.0
		validY := 2.0
		scoreNegative := -1
		scoreExcessive := 11
		scoreNine := 9
		scoreTen := 10

		tests := []struct {
			name   string
			mutate func(d *model.ShotCreate)
		}{
			{"empty slot id", func(d *model.ShotCreate) { d.SlotID = uuid.Nil }},
			{"score negative", func(d *model.ShotCreate) { d.Score = &scoreNegative }},
			{"score excessive", func(d *model.ShotCreate) { d.Score = &scoreExcessive }},
			{"is_x true with score not 10", func(d *model.ShotCreate) {
				d.IsX = true
				d.Score = &scoreNine
			}},
			{"is_x true with score nil", func(d *model.ShotCreate) {
				d.IsX = true
				d.Score = nil
				d.X = nil
				d.Y = nil
			}},
			{"x present but y and score nil", func(d *model.ShotCreate) {
				d.X = &validX
				d.Y = nil
				d.Score = nil
			}},
			{"x and y present but score nil", func(d *model.ShotCreate) {
				d.X = &validX
				d.Y = &validY
				d.Score = nil
			}},
			{"score present but x and y nil", func(d *model.ShotCreate) {
				d.X = nil
				d.Y = nil
				d.Score = &scoreTen
			}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				data := validShotCreate(slotID)
				tc.mutate(&data)

				svc := service.NewShotService(&mockShotRepo{}, &mockSlotRepo{})

				_, err := svc.Create(ctx, data)
				if err == nil {
					t.Fatalf("expected validation error for %s, got nil", tc.name)
				}
				if !errors.Is(err, apperror.ErrValidation) {
					t.Errorf("expected ErrValidation, got: %v", err)
				}
			})
		}
	})

	t.Run("propagates slot verify error", func(t *testing.T) {
		slotErr := errors.New("db slot error")
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SlotRead, error) {
				return nil, slotErr
			},
		}
		svc := service.NewShotService(&mockShotRepo{}, slotRepo)

		_, err := svc.Create(ctx, validShotCreate(slotID))
		if !errors.Is(err, slotErr) {
			t.Errorf("expected wrapped slotErr, got: %v", err)
		}
	})

	t.Run("propagates shot create error", func(t *testing.T) {
		createErr := errors.New("db shot insert error")
		slotRepo := &mockSlotRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SlotRead, error) {
				return &slotRecord, nil
			},
		}
		shotRepo := &mockShotRepo{
			createFn: func(_ context.Context, _ model.ShotCreate) (uuid.UUID, error) {
				return uuid.Nil, createErr
			},
		}
		svc := service.NewShotService(shotRepo, slotRepo)

		_, err := svc.Create(ctx, validShotCreate(slotID))
		if !errors.Is(err, createErr) {
			t.Errorf("expected wrapped createErr, got: %v", err)
		}
	})
}

func TestShotService_Update(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	slotID := uuid.New()

	existingTen := sampleShotRead(testID, slotID, 10, false)
	existingEight := sampleShotRead(testID, slotID, 8, false)
	newScore := 10
	trueVal := true

	t.Run("success", func(t *testing.T) {
		shotRepo := &mockShotRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.ShotRead, error) {
				if id == testID {
					return &existingTen, nil
				}
				return nil, nil
			},
			updateFn: func(_ context.Context, data model.ShotSet, filter model.ShotFilter) error {
				if *filter.ShotID != testID {
					return errors.New("id mismatch")
				}
				if *data.Score != newScore {
					return errors.New("data mismatch")
				}
				return nil
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		err := svc.Update(ctx, testID, model.ShotSet{Score: &newScore})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		shotRepo := &mockShotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ShotRead, error) {
				return nil, nil
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		err := svc.Update(ctx, testID, model.ShotSet{Score: &newScore})
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		svc := service.NewShotService(&mockShotRepo{}, &mockSlotRepo{})

		err := svc.Update(ctx, uuid.Nil, model.ShotSet{Score: &newScore})
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("validates invalid fields", func(t *testing.T) {
		invalidScoreLow := -1
		invalidScoreHigh := 12
		scoreNine := 9

		tests := []struct {
			name string
			data model.ShotSet
		}{
			{"score negative", model.ShotSet{Score: &invalidScoreLow}},
			{"score excessive", model.ShotSet{Score: &invalidScoreHigh}},
			{"is_x true with score not 10", model.ShotSet{IsX: &trueVal, Score: &scoreNine}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				svc := service.NewShotService(&mockShotRepo{}, &mockSlotRepo{})

				err := svc.Update(ctx, testID, tc.data)
				if !errors.Is(err, apperror.ErrValidation) {
					t.Errorf("expected ErrValidation for %s, got: %v", tc.name, err)
				}
			})
		}
	})

	t.Run("validates is_x true when existing score is not 10 and no new score provided", func(t *testing.T) {
		shotRepo := &mockShotRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.ShotRead, error) {
				return &existingEight, nil
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		err := svc.Update(ctx, testID, model.ShotSet{IsX: &trueVal})
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("update exec failed")
		shotRepo := &mockShotRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ShotRead, error) {
				return &existingTen, nil
			},
			updateFn: func(_ context.Context, _ model.ShotSet, _ model.ShotFilter) error {
				return repoErr
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		err := svc.Update(ctx, testID, model.ShotSet{Score: &newScore})
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestShotService_Delete(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()

	t.Run("success", func(t *testing.T) {
		shotRepo := &mockShotRepo{
			deleteFn: func(_ context.Context, id uuid.UUID) error {
				if id == testID {
					return nil
				}
				return apperror.ErrNotFound
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		err := svc.Delete(ctx, testID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		shotRepo := &mockShotRepo{
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return apperror.ErrNotFound
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		err := svc.Delete(ctx, testID)
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		svc := service.NewShotService(&mockShotRepo{}, &mockSlotRepo{})

		err := svc.Delete(ctx, uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("delete failed")
		shotRepo := &mockShotRepo{
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return repoErr
			},
		}
		svc := service.NewShotService(shotRepo, &mockSlotRepo{})

		err := svc.Delete(ctx, testID)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/service/... -v -run TestShotService
```
Expected: FAIL with compilation error (`service.NewShotService` and `service.ShotRepository` undefined).

- [ ] **Step 3: Implement `backend/internal/service/shot.go`**

Create `backend/internal/service/shot.go`:

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

// ShotRepository defines persistence operations required by ShotService.
type ShotRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error)
	FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error)
	FindAll(ctx context.Context, filter model.ShotFilter) ([]model.ShotRead, error)
	Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.ShotSet, filter model.ShotFilter) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ShotService encapsulates business logic, score constraints, and validation for arrow shots.
type ShotService struct {
	shotRepo ShotRepository
	slotRepo SlotRepository
}

// NewShotService constructs a ShotService with repository dependency injection.
func NewShotService(shotRepo ShotRepository, slotRepo SlotRepository) *ShotService {
	return &ShotService{
		shotRepo: shotRepo,
		slotRepo: slotRepo,
	}
}

// GetByID retrieves a shot record by its primary key identifier.
// Returns apperror.ErrNotFound if the shot does not exist.
func (s *ShotService) GetByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "shot id is required")
	}

	shot, err := s.shotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching shot: %w", err)
	}
	if shot == nil {
		return nil, apperror.ErrNotFound
	}
	return shot, nil
}

// ListBySlotID retrieves all shots recorded for a given slot.
func (s *ShotService) ListBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error) {
	if slotID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	shots, err := s.shotRepo.FindBySlotID(ctx, slotID)
	if err != nil {
		return nil, fmt.Errorf("listing shots by slot: %w", err)
	}
	if shots == nil {
		return []model.ShotRead{}, nil
	}
	return shots, nil
}

// Create validates incoming shot data, ensures score rules and coordinate consistency are met,
// verifies that the target slot exists, and persists the shot record.
// Returns apperror.ErrNotFound if the slot does not exist.
func (s *ShotService) Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error) {
	if err := validateShotCreate(data); err != nil {
		return uuid.Nil, err
	}

	slot, err := s.slotRepo.FindByID(ctx, data.SlotID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("verifying slot: %w", err)
	}
	if slot == nil {
		return uuid.Nil, apperror.ErrNotFound
	}

	id, err := s.shotRepo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating shot: %w", err)
	}
	return id, nil
}

// Update validates mutation fields, checks score and is_x constraints, verifies shot existence,
// and updates the shot record.
// Returns apperror.ErrNotFound if the shot does not exist.
func (s *ShotService) Update(ctx context.Context, id uuid.UUID, data model.ShotSet) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "shot id is required")
	}

	if err := validateShotSet(data); err != nil {
		return err
	}

	existing, err := s.shotRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying shot: %w", err)
	}
	if existing == nil {
		return apperror.ErrNotFound
	}

	if data.IsX != nil && *data.IsX && data.Score == nil {
		if existing.Score == nil || *existing.Score != 10 {
			return apperror.Wrap(apperror.ErrValidation, "is_x requires score of 10")
		}
	}

	filter := model.ShotFilter{ShotID: &id}
	if err := s.shotRepo.Update(ctx, data, filter); err != nil {
		return fmt.Errorf("updating shot: %w", err)
	}
	return nil
}

// Delete removes a shot record by its primary key identifier.
// Returns apperror.ErrNotFound if the shot does not exist.
func (s *ShotService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "shot id is required")
	}

	if err := s.shotRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("deleting shot: %w", err)
	}
	return nil
}

func validateShotCreate(data model.ShotCreate) error {
	if data.SlotID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "slot_id is required")
	}

	if data.Score != nil && (*data.Score < 0 || *data.Score > 10) {
		return apperror.Wrap(apperror.ErrValidation, "score must be between 0 and 10")
	}

	if data.IsX && (data.Score == nil || *data.Score != 10) {
		return apperror.Wrap(apperror.ErrValidation, "is_x requires score of 10")
	}

	allNil := data.X == nil && data.Y == nil && data.Score == nil
	allPresent := data.X != nil && data.Y != nil && data.Score != nil
	if !allNil && !allPresent {
		return apperror.Wrap(apperror.ErrValidation, "coordinates x, y, and score must either all be present or all be nil")
	}

	return nil
}

func validateShotSet(data model.ShotSet) error {
	if data.Score != nil && (*data.Score < 0 || *data.Score > 10) {
		return apperror.Wrap(apperror.ErrValidation, "score must be between 0 and 10")
	}

	if data.IsX != nil && *data.IsX && data.Score != nil && *data.Score != 10 {
		return apperror.Wrap(apperror.ErrValidation, "is_x requires score of 10")
	}

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/service/... -v -run TestShotService
```
Expected: PASS for all `TestShotService_*` tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/shot.go backend/internal/service/shot_test.go
git commit -m "feat(service): add shot service with score constraints and slot validation"
```

---

### Task 4: Full Verification and Quality Gate

**Files:**
- None (verification across backend codebase)

**Interfaces:**
- Consumes: all files in `backend/internal/service/`

- [ ] **Step 1: Run full service test suite with race detector**

```bash
cd backend && go test -race ./internal/service/... -v
```
Expected: PASS for all tests across `archer_test.go`, `session_test.go`, `slot_test.go`, and `shot_test.go`.

- [ ] **Step 2: Run go vet and build**

```bash
cd backend && go vet ./... && go build ./...
```
Expected: Clean exit code 0, no vet warnings, build succeeds.

- [ ] **Step 3: Run Go linting and formatting check**

```bash
./scripts/linting.bash --go
```
Expected: `gofumpt` reports no formatting differences, `golangci-lint` passes with zero issues.

- [ ] **Step 4: Verify working tree clean**

```bash
git status
```
Expected: Working tree clean, all commits recorded on branch `refactor/015-service-layer-slot-and-shot`.
