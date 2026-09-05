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
