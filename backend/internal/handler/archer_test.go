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

//nolint:gocritic // hugeParam: filter matches ArcherService interface specification
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
