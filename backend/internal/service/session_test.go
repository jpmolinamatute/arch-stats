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
	_ service.SessionRepository = (*mockSessionRepo)(nil)
	_ service.SessionRepository = (*repository.SessionRepo)(nil)
)

type mockSessionRepo struct {
	findByIDFn          func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	findOpenFn          func(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	findAllFn           func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	createFn            func(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	updateFn            func(ctx context.Context, data model.SessionSet, filter model.SessionFilter) error
	closeFn             func(ctx context.Context, id uuid.UUID) error
	deleteFn            func(ctx context.Context, id uuid.UUID) error
	findParticipatingFn func(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)
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

func (m *mockSessionRepo) FindParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error) {
	if m.findParticipatingFn != nil {
		return m.findParticipatingFn(ctx, archerID)
	}
	return nil, nil
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
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.SessionRead, error) {
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

func TestSessionService_GetParticipating(t *testing.T) {
	t.Run("returns session id when archer is participating", func(t *testing.T) {
		archerID := uuid.New()
		sessionID := uuid.New()
		repo := &mockSessionRepo{
			findParticipatingFn: func(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
				if id == archerID {
					return &sessionID, nil
				}
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		got, err := svc.GetParticipating(context.Background(), archerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || *got != sessionID {
			t.Fatalf("expected session ID %v, got %v", sessionID, got)
		}
	})

	t.Run("returns nil when archer is not participating", func(t *testing.T) {
		archerID := uuid.New()
		repo := &mockSessionRepo{
			findParticipatingFn: func(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
				return nil, nil
			},
		}
		svc := service.NewSessionService(repo)

		got, err := svc.GetParticipating(context.Background(), archerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil session ID, got %v", got)
		}
	})

	t.Run("validates empty UUID", func(t *testing.T) {
		repo := &mockSessionRepo{}
		svc := service.NewSessionService(repo)

		_, err := svc.GetParticipating(context.Background(), uuid.Nil)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected apperror.ErrValidation, got %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repoErr := errors.New("query failed")
		archerID := uuid.New()
		repo := &mockSessionRepo{
			findParticipatingFn: func(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
				return nil, repoErr
			},
		}
		svc := service.NewSessionService(repo)

		_, err := svc.GetParticipating(context.Background(), archerID)
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected wrapped repoErr, got %v", err)
		}
	})
}
