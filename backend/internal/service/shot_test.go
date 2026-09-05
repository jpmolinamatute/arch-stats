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
