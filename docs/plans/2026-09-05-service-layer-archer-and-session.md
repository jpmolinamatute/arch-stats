# Task 014: Build Service Layer — Archer and Shooting Session Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the business logic service layer for archer profile management (`ArcherService`) and shooting session lifecycle management (`SessionService`) with repository interface dependency injection, input validation, conflict checks, and comprehensive unit tests.

**Architecture:** 
- `ArcherService` sits between HTTP handlers and `ArcherRepo`. It validates incoming data payloads (`model.ArcherCreate`, `model.ArcherSet`), verifies entity existence, and delegates persistence to `ArcherRepository`.
- `SessionService` sits between HTTP handlers and `SessionRepo`. It enforces domain invariants such as the single active open session constraint per archer, session state validation on close/re-open, and delegates persistence to `SessionRepository`.
- Both services consume repository interfaces defined directly in the service package, enabling pure unit testing with mock repositories without requiring database transactions or pools.

**Tech Stack:** Go 1.27+, `github.com/google/uuid`, standard library (`context`, `errors`, `fmt`, `net/mail`, `strings`, `time`), internal packages (`model`, `apperror`).

**Spec:** [docs/go_refactor/tasks/014-service_layer_archer_and_session.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/014-service_layer_archer_and_session.md)

## Global Constraints

- Git branch: `refactor/014-service-layer-archer-and-session`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/service`
- Error handling: Wrap internal errors with `%w` using contextual descriptive messages (`fmt.Errorf("...: %w", err)`). Return sentinel `apperror.ErrNotFound`, `apperror.ErrConflict`, and `apperror.Wrap(apperror.ErrValidation, ...)` as appropriate.
- Dependency injection: Services must accept repository interfaces in their constructors (`NewArcherService(repo ArcherRepository)`, `NewSessionService(repo SessionRepository)`).
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
        ├── .gitkeep              # [DELETE] Remove placeholder once files are added
        ├── archer.go             # [NEW] ArcherRepository interface, ArcherService struct, constructor, CRUD & validation methods
        ├── archer_test.go        # [NEW] mockArcherRepo and unit test suite for ArcherService
        ├── session.go            # [NEW] SessionRepository interface, SessionService struct, constructor, lifecycle & validation methods
        └── session_test.go       # [NEW] mockSessionRepo and unit test suite for SessionService
```

---

### Task 1: Git Branch Setup & Scaffold

**Files:**
- Modify: git branch checkout
- Delete: `backend/internal/service/.gitkeep`

**Interfaces:**
- Consumes: `main` branch
- Produces: `refactor/014-service-layer-archer-and-session` branch

- [ ] **Step 1: Check out git branch**

```bash
git switch -c refactor/014-service-layer-archer-and-session
```

- [ ] **Step 2: Verify git status is clean**

```bash
git status
```
Expected: On branch `refactor/014-service-layer-archer-and-session`, working tree clean.

---

### Task 2: Archer Service Tests & Implementation (`archer_test.go` & `archer.go`)

**Files:**
- Create: `backend/internal/service/archer_test.go`
- Create: `backend/internal/service/archer.go`

**Interfaces:**
- Consumes:
  - `model.ArcherRead`, `model.ArcherCreate`, `model.ArcherSet`, `model.ArcherFilter` from `backend/internal/model/archer.go`
  - `model.Gender`, `model.Bowstyle` from `backend/internal/model/enums.go`
  - `apperror.ErrNotFound`, `apperror.ErrValidation`, `apperror.Wrap` from `backend/internal/apperror/errors.go`
- Produces:
  - `ArcherRepository` interface:
    - `FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)`
    - `FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)`
    - `Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`
  - `ArcherService` struct and constructor `NewArcherService(repo ArcherRepository) *ArcherService`
  - Methods:
    - `GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)`
    - `List(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)`
    - `Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, id uuid.UUID, data model.ArcherSet) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`

- [ ] **Step 1: Write failing tests in `backend/internal/service/archer_test.go`**

Create `backend/internal/service/archer_test.go`:

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
	"github.com/jpmolinamatute/arch-stats/backend/internal/service"
)

type mockArcherRepo struct {
	findByIDFn func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	findAllFn  func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
	createFn   func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	updateFn   func(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error
	deleteFn   func(ctx context.Context, id uuid.UUID) error
}

func (m *mockArcherRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockArcherRepo) FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockArcherRepo) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, nil
}

func (m *mockArcherRepo) Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, data, filter)
	}
	return nil
}

func (m *mockArcherRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func validArcherCreate() model.ArcherCreate {
	return model.ArcherCreate{
		FirstName:     "Robin",
		LastName:      "Hood",
		Email:         "robin@sherwood.org",
		DateOfBirth:   "1990-05-15",
		Gender:        model.GenderMale,
		Bowstyle:      model.BowstyleRecurve,
		DrawWeight:    42.5,
		GoogleSubject: "google-sub-12345",
	}
}

func sampleArcherRead(id uuid.UUID) model.ArcherRead {
	return model.ArcherRead{
		ArcherID:      id,
		FirstName:     "Robin",
		LastName:      "Hood",
		Email:         "robin@sherwood.org",
		DateOfBirth:   "1990-05-15",
		Gender:        model.GenderMale,
		Bowstyle:      model.BowstyleRecurve,
		DrawWeight:    42.5,
		GoogleSubject: "google-sub-12345",
		CreatedAt:     time.Now().UTC(),
		LastLoginAt:   time.Now().UTC(),
	}
}

func TestArcherService_GetByID(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := sampleArcherRead(testID)
		repo := &mockArcherRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				if id == testID {
					return &expected, nil
				}
				return nil, nil
			},
		}
		svc := service.NewArcherService(repo)

		got, err := svc.GetByID(ctx, testID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ArcherID != testID || got.FirstName != expected.FirstName {
			t.Errorf("got %+v, expected %+v", got, expected)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &mockArcherRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ArcherRead, error) {
				return nil, nil
			},
		}
		svc := service.NewArcherService(repo)

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
		repo := &mockArcherRepo{}
		svc := service.NewArcherService(repo)

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
		repo := &mockArcherRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ArcherRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewArcherService(repo)

		_, err := svc.GetByID(ctx, testID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestArcherService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("success with items", func(t *testing.T) {
		expected := []model.ArcherRead{sampleArcherRead(uuid.New()), sampleArcherRead(uuid.New())}
		repo := &mockArcherRepo{
			findAllFn: func(_ context.Context, _ model.ArcherFilter) ([]model.ArcherRead, error) {
				return expected, nil
			},
		}
		svc := service.NewArcherService(repo)

		got, err := svc.List(ctx, model.ArcherFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(expected) {
			t.Errorf("got %d archers, expected %d", len(got), len(expected))
		}
	})

	t.Run("returns empty slice when repo returns nil", func(t *testing.T) {
		repo := &mockArcherRepo{
			findAllFn: func(_ context.Context, _ model.ArcherFilter) ([]model.ArcherRead, error) {
				return nil, nil
			},
		}
		svc := service.NewArcherService(repo)

		got, err := svc.List(ctx, model.ArcherFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("expected empty slice, got %+v", got)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("db query failed")
		repo := &mockArcherRepo{
			findAllFn: func(_ context.Context, _ model.ArcherFilter) ([]model.ArcherRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewArcherService(repo)

		_, err := svc.List(ctx, model.ArcherFilter{})
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestArcherService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		newID := uuid.New()
		data := validArcherCreate()
		repo := &mockArcherRepo{
			createFn: func(_ context.Context, d model.ArcherCreate) (uuid.UUID, error) {
				if d.Email != data.Email {
					return uuid.Nil, errors.New("email mismatch")
				}
				return newID, nil
			},
		}
		svc := service.NewArcherService(repo)

		id, err := svc.Create(ctx, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != newID {
			t.Errorf("got id %v, expected %v", id, newID)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name   string
			mutate func(d *model.ArcherCreate)
		}{
			{"empty first name", func(d *model.ArcherCreate) { d.FirstName = "  " }},
			{"empty last name", func(d *model.ArcherCreate) { d.LastName = "" }},
			{"empty email", func(d *model.ArcherCreate) { d.Email = "" }},
			{"invalid email format", func(d *model.ArcherCreate) { d.Email = "invalid-email" }},
			{"empty date of birth", func(d *model.ArcherCreate) { d.DateOfBirth = "" }},
			{"invalid date of birth format", func(d *model.ArcherCreate) { d.DateOfBirth = "15/05/1990" }},
			{"invalid gender", func(d *model.ArcherCreate) { d.Gender = "alien" }},
			{"invalid bowstyle", func(d *model.ArcherCreate) { d.Bowstyle = "crossbow" }},
			{"draw weight zero", func(d *model.ArcherCreate) { d.DrawWeight = 0 }},
			{"draw weight negative", func(d *model.ArcherCreate) { d.DrawWeight = -5 }},
			{"draw weight excessive", func(d *model.ArcherCreate) { d.DrawWeight = 250 }},
			{"empty google subject", func(d *model.ArcherCreate) { d.GoogleSubject = "" }},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				data := validArcherCreate()
				tc.mutate(&data)

				repo := &mockArcherRepo{}
				svc := service.NewArcherService(repo)

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

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("insert failed")
		repo := &mockArcherRepo{
			createFn: func(_ context.Context, _ model.ArcherCreate) (uuid.UUID, error) {
				return uuid.Nil, repoErr
			},
		}
		svc := service.NewArcherService(repo)

		_, err := svc.Create(ctx, validArcherCreate())
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestArcherService_Update(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	newName := "Marion"

	t.Run("success", func(t *testing.T) {
		existing := sampleArcherRead(testID)
		repo := &mockArcherRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				if id == testID {
					return &existing, nil
				}
				return nil, nil
			},
			updateFn: func(_ context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
				if *filter.ArcherID != testID {
					return errors.New("id mismatch")
				}
				if *data.FirstName != newName {
					return errors.New("data mismatch")
				}
				return nil
			},
		}
		svc := service.NewArcherService(repo)

		err := svc.Update(ctx, testID, model.ArcherSet{FirstName: &newName})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &mockArcherRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ArcherRead, error) {
				return nil, nil
			},
		}
		svc := service.NewArcherService(repo)

		err := svc.Update(ctx, testID, model.ArcherSet{FirstName: &newName})
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		repo := &mockArcherRepo{}
		svc := service.NewArcherService(repo)

		err := svc.Update(ctx, uuid.Nil, model.ArcherSet{FirstName: &newName})
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("validates invalid fields", func(t *testing.T) {
		invalidWeight := -10.0
		invalidGender := model.Gender("robot")
		invalidBow := model.Bowstyle("laser")
		emptyFirst := "   "

		tests := []struct {
			name string
			data model.ArcherSet
		}{
			{"empty first name", model.ArcherSet{FirstName: &emptyFirst}},
			{"invalid draw weight", model.ArcherSet{DrawWeight: &invalidWeight}},
			{"invalid gender", model.ArcherSet{Gender: &invalidGender}},
			{"invalid bowstyle", model.ArcherSet{Bowstyle: &invalidBow}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				repo := &mockArcherRepo{}
				svc := service.NewArcherService(repo)

				err := svc.Update(ctx, testID, tc.data)
				if !errors.Is(err, apperror.ErrValidation) {
					t.Errorf("expected ErrValidation for %s, got: %v", tc.name, err)
				}
			})
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		existing := sampleArcherRead(testID)
		repoErr := errors.New("update exec failed")
		repo := &mockArcherRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ArcherRead, error) {
				return &existing, nil
			},
			updateFn: func(_ context.Context, _ model.ArcherSet, _ model.ArcherFilter) error {
				return repoErr
			},
		}
		svc := service.NewArcherService(repo)

		err := svc.Update(ctx, testID, model.ArcherSet{FirstName: &newName})
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestArcherService_Delete(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := &mockArcherRepo{
			deleteFn: func(_ context.Context, id uuid.UUID) error {
				if id == testID {
					return nil
				}
				return apperror.ErrNotFound
			},
		}
		svc := service.NewArcherService(repo)

		err := svc.Delete(ctx, testID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &mockArcherRepo{
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return apperror.ErrNotFound
			},
		}
		svc := service.NewArcherService(repo)

		err := svc.Delete(ctx, testID)
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		repo := &mockArcherRepo{}
		svc := service.NewArcherService(repo)

		err := svc.Delete(ctx, uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("delete failed")
		repo := &mockArcherRepo{
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return repoErr
			},
		}
		svc := service.NewArcherService(repo)

		err := svc.Delete(ctx, testID)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/service/... -v
```
Expected: FAIL with compilation error (undefined `service.NewArcherService`, etc.).

- [ ] **Step 3: Implement `backend/internal/service/archer.go`**

Create `backend/internal/service/archer.go`:

```go
package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ArcherRepository defines persistence operations required by ArcherService.
type ArcherRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
	Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ArcherService encapsulates business logic and validation for archer profiles.
type ArcherService struct {
	repo ArcherRepository
}

// NewArcherService constructs an ArcherService with repository dependency injection.
func NewArcherService(repo ArcherRepository) *ArcherService {
	return &ArcherService{repo: repo}
}

// GetByID retrieves an archer profile by primary key identifier.
// Returns apperror.ErrNotFound if the archer does not exist.
func (s *ArcherService) GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "archer id is required")
	}

	archer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching archer: %w", err)
	}
	if archer == nil {
		return nil, apperror.ErrNotFound
	}
	return archer, nil
}

// List queries archer profiles matching the provided filter criteria.
func (s *ArcherService) List(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
	archers, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing archers: %w", err)
	}
	if archers == nil {
		return []model.ArcherRead{}, nil
	}
	return archers, nil
}

// Create validates incoming archer profile fields and persists the record.
func (s *ArcherService) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	if err := validateArcherCreate(data); err != nil {
		return uuid.Nil, err
	}

	id, err := s.repo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating archer: %w", err)
	}
	return id, nil
}

// Update validates mutation fields and updates an archer profile.
// Returns apperror.ErrNotFound if the archer does not exist.
func (s *ArcherService) Update(ctx context.Context, id uuid.UUID, data model.ArcherSet) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "archer id is required")
	}

	if err := validateArcherSet(data); err != nil {
		return err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying archer existence: %w", err)
	}
	if existing == nil {
		return apperror.ErrNotFound
	}

	filter := model.ArcherFilter{ArcherID: &id}
	if err := s.repo.Update(ctx, data, filter); err != nil {
		return fmt.Errorf("updating archer: %w", err)
	}
	return nil
}

// Delete removes an archer profile by primary key identifier.
// Returns apperror.ErrNotFound if the archer does not exist.
func (s *ArcherService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "archer id is required")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting archer: %w", err)
	}
	return nil
}

func validateArcherCreate(data model.ArcherCreate) error {
	if strings.TrimSpace(data.FirstName) == "" {
		return apperror.Wrap(apperror.ErrValidation, "first_name is required")
	}
	if strings.TrimSpace(data.LastName) == "" {
		return apperror.Wrap(apperror.ErrValidation, "last_name is required")
	}
	if strings.TrimSpace(data.Email) == "" {
		return apperror.Wrap(apperror.ErrValidation, "email is required")
	}
	if _, err := mail.ParseAddress(data.Email); err != nil {
		return apperror.Wrap(apperror.ErrValidation, "invalid email address")
	}
	if strings.TrimSpace(data.DateOfBirth) == "" {
		return apperror.Wrap(apperror.ErrValidation, "date_of_birth is required")
	}
	if _, err := time.Parse("2006-01-02", data.DateOfBirth); err != nil {
		return apperror.Wrap(apperror.ErrValidation, "date_of_birth must be formatted as YYYY-MM-DD")
	}
	if !isValidGender(data.Gender) {
		return apperror.Wrap(apperror.ErrValidation, "invalid gender")
	}
	if !isValidBowstyle(data.Bowstyle) {
		return apperror.Wrap(apperror.ErrValidation, "invalid bowstyle")
	}
	if data.DrawWeight <= 0 || data.DrawWeight > 200 {
		return apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}
	if strings.TrimSpace(data.GoogleSubject) == "" {
		return apperror.Wrap(apperror.ErrValidation, "google_subject is required")
	}
	return nil
}

func validateArcherSet(data model.ArcherSet) error {
	if data.FirstName != nil && strings.TrimSpace(*data.FirstName) == "" {
		return apperror.Wrap(apperror.ErrValidation, "first_name cannot be empty")
	}
	if data.LastName != nil && strings.TrimSpace(*data.LastName) == "" {
		return apperror.Wrap(apperror.ErrValidation, "last_name cannot be empty")
	}
	if data.Gender != nil && !isValidGender(*data.Gender) {
		return apperror.Wrap(apperror.ErrValidation, "invalid gender")
	}
	if data.Bowstyle != nil && !isValidBowstyle(*data.Bowstyle) {
		return apperror.Wrap(apperror.ErrValidation, "invalid bowstyle")
	}
	if data.DrawWeight != nil && (*data.DrawWeight <= 0 || *data.DrawWeight > 200) {
		return apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}
	return nil
}

func isValidGender(g model.Gender) bool {
	switch g {
	case model.GenderMale, model.GenderFemale, model.GenderNonBinary, model.GenderOther, model.GenderUnspecified:
		return true
	default:
		return false
	}
}

func isValidBowstyle(b model.Bowstyle) bool {
	switch b {
	case model.BowstyleRecurve, model.BowstyleCompound, model.BowstyleBarebow, model.BowstyleLongbow:
		return true
	default:
		return false
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/service/... -v -run TestArcherService
```
Expected: PASS for all `TestArcherService_*` tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/archer.go backend/internal/service/archer_test.go
git commit -m "feat(service): add archer service with validation and repository injection"
```

---

### Task 3: Session Service Tests & Implementation (`session_test.go` & `session.go`)

**Files:**
- Create: `backend/internal/service/session_test.go`
- Create: `backend/internal/service/session.go`

**Interfaces:**
- Consumes:
  - `model.SessionRead`, `model.SessionCreate`, `model.SessionSet`, `model.SessionFilter` from `backend/internal/model/session.go`
  - `apperror.ErrNotFound`, `apperror.ErrConflict`, `apperror.ErrValidation`, `apperror.Wrap` from `backend/internal/apperror/errors.go`
- Produces:
  - `SessionRepository` interface:
    - `FindByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)`
    - `FindOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)`
    - `FindAll(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)`
    - `Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, data model.SessionSet, filter model.SessionFilter) error`
    - `Close(ctx context.Context, id uuid.UUID) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`
  - `SessionService` struct and constructor `NewSessionService(repo SessionRepository) *SessionService`
  - Methods:
    - `GetByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)`
    - `GetOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)`
    - `List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)`
    - `Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)`
    - `Close(ctx context.Context, id uuid.UUID) error`
    - `ReOpen(ctx context.Context, id uuid.UUID) error`
    - `Delete(ctx context.Context, id uuid.UUID) error`

- [ ] **Step 1: Write failing tests in `backend/internal/service/session_test.go`**

Create `backend/internal/service/session_test.go`:

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
	"github.com/jpmolinamatute/arch-stats/backend/internal/service"
)

type mockSessionRepo struct {
	findByIDFn func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	findOpenFn func(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	findAllFn  func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	createFn   func(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	updateFn   func(ctx context.Context, data model.SessionSet, filter model.SessionFilter) error
	closeFn    func(ctx context.Context, id uuid.UUID) error
	deleteFn   func(ctx context.Context, id uuid.UUID) error
}

func (m *mockSessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSessionRepo) FindOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error) {
	if m.findOpenFn != nil {
		return m.findOpenFn(ctx, archerID)
	}
	return nil, nil
}

func (m *mockSessionRepo) FindAll(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockSessionRepo) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, nil
}

func (m *mockSessionRepo) Update(ctx context.Context, data model.SessionSet, filter model.SessionFilter) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, data, filter)
	}
	return nil
}

func (m *mockSessionRepo) Close(ctx context.Context, id uuid.UUID) error {
	if m.closeFn != nil {
		return m.closeFn(ctx, id)
	}
	return nil
}

func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func sampleSessionRead(id, archerID uuid.UUID, isOpened bool) model.SessionRead {
	var closedAt *time.Time
	if !isOpened {
		t := time.Now().UTC()
		closedAt = &t
	}
	return model.SessionRead{
		SessionID:       id,
		OwnerArcherID:   archerID,
		SessionLocation: "Sherwood Outdoor Range",
		IsIndoor:        false,
		IsOpened:        isOpened,
		CreatedAt:       time.Now().UTC().Add(-1 * time.Hour),
		ClosedAt:        closedAt,
	}
}

func TestSessionService_GetByID(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	archerID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := sampleSessionRead(testID, archerID, true)
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
				if id == testID {
					return &expected, nil
				}
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		got, err := svc.GetByID(ctx, testID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SessionID != testID || got.SessionLocation != expected.SessionLocation {
			t.Errorf("got %+v, expected %+v", got, expected)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

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
		repo := &mockSessionRepo{}
		svc := service.NewSessionService(repo)

		_, err := svc.GetByID(ctx, uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("db find failed")
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewSessionService(repo)

		_, err := svc.GetByID(ctx, testID)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestSessionService_GetOpen(t *testing.T) {
	ctx := context.Background()
	archerID := uuid.New()
	sessionID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := sampleSessionRead(sessionID, archerID, true)
		repo := &mockSessionRepo{
			findOpenFn: func(_ context.Context, aid uuid.UUID) (*model.SessionRead, error) {
				if aid == archerID {
					return &expected, nil
				}
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		got, err := svc.GetOpen(ctx, archerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SessionID != sessionID || !got.IsOpened {
			t.Errorf("got %+v, expected open session %+v", got, expected)
		}
	})

	t.Run("returns not found when no open session exists", func(t *testing.T) {
		repo := &mockSessionRepo{
			findOpenFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		got, err := svc.GetOpen(ctx, archerID)
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil result, got %+v", got)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		repo := &mockSessionRepo{}
		svc := service.NewSessionService(repo)

		_, err := svc.GetOpen(ctx, uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("db find open failed")
		repo := &mockSessionRepo{
			findOpenFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewSessionService(repo)

		_, err := svc.GetOpen(ctx, archerID)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestSessionService_List(t *testing.T) {
	ctx := context.Background()
	archerID := uuid.New()

	t.Run("success with items", func(t *testing.T) {
		expected := []model.SessionRead{
			sampleSessionRead(uuid.New(), archerID, true),
			sampleSessionRead(uuid.New(), archerID, false),
		}
		repo := &mockSessionRepo{
			findAllFn: func(_ context.Context, _ model.SessionFilter) ([]model.SessionRead, error) {
				return expected, nil
			},
		}
		svc := service.NewSessionService(repo)

		got, err := svc.List(ctx, model.SessionFilter{OwnerArcherID: &archerID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(expected) {
			t.Errorf("got %d sessions, expected %d", len(got), len(expected))
		}
	})

	t.Run("returns empty slice when repo returns nil", func(t *testing.T) {
		repo := &mockSessionRepo{
			findAllFn: func(_ context.Context, _ model.SessionFilter) ([]model.SessionRead, error) {
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		got, err := svc.List(ctx, model.SessionFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("expected empty slice, got %+v", got)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("db find all failed")
		repo := &mockSessionRepo{
			findAllFn: func(_ context.Context, _ model.SessionFilter) ([]model.SessionRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewSessionService(repo)

		_, err := svc.List(ctx, model.SessionFilter{})
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestSessionService_Create(t *testing.T) {
	ctx := context.Background()
	archerID := uuid.New()
	newSessionID := uuid.New()

	t.Run("success when no open session exists", func(t *testing.T) {
		createPayload := model.SessionCreate{
			OwnerArcherID:   archerID,
			SessionLocation: "Sherwood Range",
			IsIndoor:        false,
		}
		repo := &mockSessionRepo{
			findOpenFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return nil, nil // No active open session
			},
			createFn: func(_ context.Context, data model.SessionCreate) (uuid.UUID, error) {
				if !data.IsOpened {
					return uuid.Nil, errors.New("new session must be open")
				}
				return newSessionID, nil
			},
		}
		svc := service.NewSessionService(repo)

		id, err := svc.Create(ctx, createPayload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != newSessionID {
			t.Errorf("got id %v, expected %v", id, newSessionID)
		}
	})

	t.Run("conflict when archer already has an open session", func(t *testing.T) {
		activeSession := sampleSessionRead(uuid.New(), archerID, true)
		createPayload := model.SessionCreate{
			OwnerArcherID:   archerID,
			SessionLocation: "Sherwood Range",
			IsIndoor:        false,
		}
		repo := &mockSessionRepo{
			findOpenFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return &activeSession, nil // Existing open session
			},
		}
		svc := service.NewSessionService(repo)

		_, err := svc.Create(ctx, createPayload)
		if err == nil {
			t.Fatal("expected conflict error, got nil")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got: %v", err)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			payload model.SessionCreate
		}{
			{"empty owner archer id", model.SessionCreate{OwnerArcherID: uuid.Nil, SessionLocation: "Range"}},
			{"empty session location", model.SessionCreate{OwnerArcherID: archerID, SessionLocation: "   "}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				repo := &mockSessionRepo{}
				svc := service.NewSessionService(repo)

				_, err := svc.Create(ctx, tc.payload)
				if !errors.Is(err, apperror.ErrValidation) {
					t.Errorf("expected ErrValidation for %s, got: %v", tc.name, err)
				}
			})
		}
	})

	t.Run("propagates check error", func(t *testing.T) {
		repoErr := errors.New("find open check failed")
		repo := &mockSessionRepo{
			findOpenFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, repoErr
			},
		}
		svc := service.NewSessionService(repo)

		_, err := svc.Create(ctx, model.SessionCreate{OwnerArcherID: archerID, SessionLocation: "Range"})
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})

	t.Run("propagates create error", func(t *testing.T) {
		repoErr := errors.New("create session failed")
		repo := &mockSessionRepo{
			findOpenFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, nil
			},
			createFn: func(_ context.Context, _ model.SessionCreate) (uuid.UUID, error) {
				return uuid.Nil, repoErr
			},
		}
		svc := service.NewSessionService(repo)

		_, err := svc.Create(ctx, model.SessionCreate{OwnerArcherID: archerID, SessionLocation: "Range"})
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestSessionService_Close(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	archerID := uuid.New()

	t.Run("success closing open session", func(t *testing.T) {
		openSession := sampleSessionRead(sessionID, archerID, true)
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
				if id == sessionID {
					return &openSession, nil
				}
				return nil, nil
			},
			closeFn: func(_ context.Context, id uuid.UUID) error {
				if id == sessionID {
					return nil
				}
				return apperror.ErrNotFound
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.Close(ctx, sessionID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when session does not exist", func(t *testing.T) {
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.Close(ctx, sessionID)
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("returns validation error when session is already closed", func(t *testing.T) {
		closedSession := sampleSessionRead(sessionID, archerID, false)
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return &closedSession, nil
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.Close(ctx, sessionID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		repo := &mockSessionRepo{}
		svc := service.NewSessionService(repo)

		err := svc.Close(ctx, uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository close error", func(t *testing.T) {
		openSession := sampleSessionRead(sessionID, archerID, true)
		repoErr := errors.New("db close failed")
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return &openSession, nil
			},
			closeFn: func(_ context.Context, _ uuid.UUID) error {
				return repoErr
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.Close(ctx, sessionID)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestSessionService_ReOpen(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	archerID := uuid.New()

	t.Run("success reopening closed session", func(t *testing.T) {
		closedSession := sampleSessionRead(sessionID, archerID, false)
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return &closedSession, nil
			},
			findOpenFn: func(_ context.Context, aid uuid.UUID) (*model.SessionRead, error) {
				return nil, nil // No current open session
			},
			updateFn: func(_ context.Context, data model.SessionSet, filter model.SessionFilter) error {
				if *filter.SessionID != sessionID {
					return errors.New("id mismatch")
				}
				if data.IsOpened == nil || !*data.IsOpened {
					return errors.New("is_opened must be true")
				}
				return nil
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.ReOpen(ctx, sessionID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when session does not exist", func(t *testing.T) {
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.ReOpen(ctx, sessionID)
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("returns validation error when session is already open", func(t *testing.T) {
		openSession := sampleSessionRead(sessionID, archerID, true)
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return &openSession, nil
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.ReOpen(ctx, sessionID)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("returns conflict when archer already has another open session", func(t *testing.T) {
		closedSession := sampleSessionRead(sessionID, archerID, false)
		otherOpenSession := sampleSessionRead(uuid.New(), archerID, true)
		repo := &mockSessionRepo{
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return &closedSession, nil
			},
			findOpenFn: func(_ context.Context, _ uuid.UUID) (*model.SessionRead, error) {
				return &otherOpenSession, nil
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.ReOpen(ctx, sessionID)
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got: %v", err)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		repo := &mockSessionRepo{}
		svc := service.NewSessionService(repo)

		err := svc.ReOpen(ctx, uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})
}

func TestSessionService_Delete(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := &mockSessionRepo{
			deleteFn: func(_ context.Context, id uuid.UUID) error {
				if id == sessionID {
					return nil
				}
				return apperror.ErrNotFound
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.Delete(ctx, sessionID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &mockSessionRepo{
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return apperror.ErrNotFound
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.Delete(ctx, sessionID)
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		repo := &mockSessionRepo{}
		svc := service.NewSessionService(repo)

		err := svc.Delete(ctx, uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation, got: %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("delete failed")
		repo := &mockSessionRepo{
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return repoErr
			},
		}
		svc := service.NewSessionService(repo)

		err := svc.Delete(ctx, sessionID)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repoErr, got: %v", err)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/service/... -v -run TestSessionService
```
Expected: FAIL with compilation error (undefined `service.NewSessionService`, etc.).

- [ ] **Step 3: Implement `backend/internal/service/session.go`**

Create `backend/internal/service/session.go`:

```go
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// SessionRepository defines persistence operations required by SessionService.
type SessionRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	FindOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	FindAll(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.SessionSet, filter model.SessionFilter) error
	Close(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SessionService encapsulates business logic and lifecycle rules for shooting sessions.
type SessionService struct {
	repo SessionRepository
}

// NewSessionService constructs a SessionService with repository dependency injection.
func NewSessionService(repo SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

// GetByID retrieves a shooting session by primary key identifier.
// Returns apperror.ErrNotFound if the session does not exist.
func (s *SessionService) GetByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	session, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching session: %w", err)
	}
	if session == nil {
		return nil, apperror.ErrNotFound
	}
	return session, nil
}

// GetOpen retrieves the active open session owned by the archer.
// Returns apperror.ErrNotFound if no open session exists.
func (s *SessionService) GetOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error) {
	if archerID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "archer id is required")
	}

	session, err := s.repo.FindOpen(ctx, archerID)
	if err != nil {
		return nil, fmt.Errorf("fetching open session: %w", err)
	}
	if session == nil {
		return nil, apperror.ErrNotFound
	}
	return session, nil
}

// List retrieves all shooting sessions matching the specified filter criteria.
func (s *SessionService) List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	sessions, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	if sessions == nil {
		return []model.SessionRead{}, nil
	}
	return sessions, nil
}

// Create validates session data and ensures the owner does not already have an open session before creating.
// Returns apperror.ErrConflict if an open session already exists for the archer.
func (s *SessionService) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
	if data.OwnerArcherID == uuid.Nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "owner_archer_id is required")
	}
	if strings.TrimSpace(data.SessionLocation) == "" {
		return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "session_location is required")
	}

	openSession, err := s.repo.FindOpen(ctx, data.OwnerArcherID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("checking for active open session: %w", err)
	}
	if openSession != nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrConflict, "archer already has an open session")
	}

	data.IsOpened = true
	id, err := s.repo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating session: %w", err)
	}
	return id, nil
}

// Close closes an active shooting session after validating that it exists and is currently open.
// Returns apperror.ErrNotFound if missing, or apperror.ErrValidation if already closed.
func (s *SessionService) Close(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	session, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying session: %w", err)
	}
	if session == nil {
		return apperror.ErrNotFound
	}
	if !session.IsOpened {
		return apperror.Wrap(apperror.ErrValidation, "session is already closed")
	}

	if err := s.repo.Close(ctx, id); err != nil {
		return fmt.Errorf("closing session: %w", err)
	}
	return nil
}

// ReOpen re-opens a closed shooting session after verifying it exists, is closed, and
// the archer does not currently have another open session.
// Returns apperror.ErrNotFound if missing, apperror.ErrValidation if already open,
// or apperror.ErrConflict if another open session exists for the archer.
func (s *SessionService) ReOpen(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	session, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying session: %w", err)
	}
	if session == nil {
		return apperror.ErrNotFound
	}
	if session.IsOpened {
		return apperror.Wrap(apperror.ErrValidation, "session is already open")
	}

	openSession, err := s.repo.FindOpen(ctx, session.OwnerArcherID)
	if err != nil {
		return fmt.Errorf("checking for conflicting open session: %w", err)
	}
	if openSession != nil {
		return apperror.Wrap(apperror.ErrConflict, "archer already has an open session")
	}

	isOpened := true
	updateData := model.SessionSet{
		IsOpened: &isOpened,
	}
	filter := model.SessionFilter{
		SessionID: &id,
	}
	if err := s.repo.Update(ctx, updateData, filter); err != nil {
		return fmt.Errorf("reopening session: %w", err)
	}
	return nil
}

// Delete removes a session by primary key identifier.
// Returns apperror.ErrNotFound if the session does not exist.
func (s *SessionService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/service/... -v -run TestSessionService
```
Expected: PASS for all `TestSessionService_*` tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/session.go backend/internal/service/session_test.go
git commit -m "feat(service): add session service with open session check and lifecycle transitions"
```

---

### Task 4: Cleanup, Full Verification, and Verification Script

**Files:**
- Delete: `backend/internal/service/.gitkeep`

- [ ] **Step 1: Delete `backend/internal/service/.gitkeep`**

```bash
rm -f backend/internal/service/.gitkeep
```

- [ ] **Step 2: Run all service tests with race detector**

```bash
cd backend && go test -race ./internal/service/... -v
```
Expected: All tests pass with zero race conditions.

- [ ] **Step 3: Run `go vet` on the entire backend**

```bash
cd backend && go vet ./...
```
Expected: Clean output with 0 errors.

- [ ] **Step 4: Run full project Go verification script**

```bash
./scripts/linting.bash --go
```
Expected:
```
INFO: Running Go formatter (gofumpt)...
INFO: Running Go linter (golangci-lint)...
0 issues.
INFO: Running Go tests...
...
Summary: ALL tests passed
```

- [ ] **Step 5: Commit cleanup**

```bash
git add -A
git commit -m "refactor(service): remove .gitkeep and verify service package passes all checks"
```
