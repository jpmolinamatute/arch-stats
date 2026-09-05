package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
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

func TestArcherHandler_List(t *testing.T) {
	t.Run("returns 200 and JSON array when archers exist", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		svc := &mockArcherHandlerService{
			listFn: func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
				return []model.ArcherRead{
					{ArcherID: id1, FirstName: "Katniss", LastName: "Everdeen"},
					{ArcherID: id2, FirstName: "Robin", LastName: "Hood"},
				}, nil
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/", http.NoBody)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp []model.ArcherRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("expected 2 archers, got %d", len(resp))
		}
		if resp[0].ArcherID != id1 || resp[1].ArcherID != id2 {
			t.Fatalf("unexpected archers returned: %+v", resp)
		}
	})

	t.Run("returns 200 and empty JSON array when no archers exist", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			listFn: func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
				return []model.ArcherRead{}, nil
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/", http.NoBody)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		body := strings.TrimSpace(rec.Body.String())
		if body != "[]" {
			t.Fatalf("expected '[]', got %q", body)
		}
	})

	t.Run("returns 500 when service fails", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			listFn: func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
				return nil, errors.New("db query error")
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/", http.NoBody)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestArcherHandler_GetByID(t *testing.T) {
	targetID := uuid.New()

	t.Run("returns 200 and archer JSON when found", func(t *testing.T) {
		svc := &mockArcherHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				if id == targetID {
					return &model.ArcherRead{
						ArcherID:  targetID,
						FirstName: "Katniss",
						LastName:  "Everdeen",
						Email:     "katniss@district12.org",
					}, nil
				}
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/"+targetID.String(), http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", targetID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
		}

		var resp model.ArcherRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ArcherID != targetID || resp.FirstName != "Katniss" {
			t.Fatalf("unexpected archer: %+v", resp)
		}
	})

	t.Run("returns 404 when archer does not exist", func(t *testing.T) {
		nonExistentID := uuid.New()
		svc := &mockArcherHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/"+nonExistentID.String(), http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when id is invalid UUID", func(t *testing.T) {
		svc := &mockArcherHandlerService{}
		h := handler.NewArcherHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/invalid-uuid", http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetByID(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})
}
