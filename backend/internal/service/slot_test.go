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

//nolint:gocritic // hugeParam: data value parameter matches repository interface specification
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
