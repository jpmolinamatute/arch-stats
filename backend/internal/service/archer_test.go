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

//nolint:gocritic // hugeParam: filter value parameter matches repository interface specification
func (m *mockArcherRepo) FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, filter)
	}
	return nil, nil
}

//nolint:gocritic // hugeParam: data value parameter matches repository interface specification
func (m *mockArcherRepo) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, nil
}

//nolint:gocritic // hugeParam: filter value parameter matches repository interface specification
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
