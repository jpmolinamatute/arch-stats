package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockSessionHandlerService struct {
	getByIDFn          func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	getOpenFn          func(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	listFn             func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	createFn           func(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	closeFn            func(ctx context.Context, id uuid.UUID) error
	reOpenFn           func(ctx context.Context, id uuid.UUID) error
	getParticipatingFn func(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)
}

func (m *mockSessionHandlerService) GetByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSessionHandlerService) GetOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error) {
	if m.getOpenFn != nil {
		return m.getOpenFn(ctx, archerID)
	}
	return nil, errors.New("unimplemented")
}

//nolint:gocritic // hugeParam: filter matches interface specification
func (m *mockSessionHandlerService) List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, errors.New("unimplemented")
}

//nolint:gocritic // hugeParam: data matches interface specification
func (m *mockSessionHandlerService) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.Nil, errors.New("unimplemented")
}

func (m *mockSessionHandlerService) Close(ctx context.Context, id uuid.UUID) error {
	if m.closeFn != nil {
		return m.closeFn(ctx, id)
	}
	return errors.New("unimplemented")
}

func (m *mockSessionHandlerService) ReOpen(ctx context.Context, id uuid.UUID) error {
	if m.reOpenFn != nil {
		return m.reOpenFn(ctx, id)
	}
	return errors.New("unimplemented")
}

func (m *mockSessionHandlerService) GetParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error) {
	if m.getParticipatingFn != nil {
		return m.getParticipatingFn(ctx, archerID)
	}
	return nil, errors.New("unimplemented")
}

func TestNewSessionHandler(t *testing.T) {
	svc := &mockSessionHandlerService{}
	h := handler.NewSessionHandler(svc)
	if h == nil {
		t.Fatal("expected NewSessionHandler to return non-nil instance")
	}
}

func TestSessionHandler_RoutesRegistration(t *testing.T) {
	svc := &mockSessionHandlerService{}
	h := handler.NewSessionHandler(svc)
	r := chi.NewRouter()
	r.Route("/api/v0/session", h.Routes)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/session/open", http.NoBody)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Since method is stubbed with 501 Not Implemented, it should NOT be 404
	if rec.Code == http.StatusNotFound {
		t.Errorf("expected route /api/v0/session/open to be registered, got 404")
	}
}
