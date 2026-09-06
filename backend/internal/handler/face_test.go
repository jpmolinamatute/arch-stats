package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockFaceHandlerService struct {
	listAllFn func(ctx context.Context) ([]model.FaceRead, error)
	getByIDFn func(ctx context.Context, id string) (*model.FaceRead, error)
}

func (m *mockFaceHandlerService) ListAll(ctx context.Context) ([]model.FaceRead, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockFaceHandlerService) GetByID(ctx context.Context, id string) (*model.FaceRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("unimplemented")
}

func TestNewFaceHandler(t *testing.T) {
	svc := &mockFaceHandlerService{}
	h := handler.NewFaceHandler(svc)
	if h == nil {
		t.Fatal("expected NewFaceHandler to return non-nil instance")
	}
}

func TestFaceHandler_ListFaces(t *testing.T) {
	t.Run("returns 200 and list of FaceMinimal summaries", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return []model.FaceRead{
					{
						FaceType:    model.FaceTypeWA40Full,
						FaceName:    "WA 40cm Full",
						ViewBox:     400,
						RenderCross: true,
						Spots:       []model.Spot{{Diameter: 400}},
						Rings:       []model.Ring{{DataScore: 10, Fill: "#FFD700"}},
					},
					{
						FaceType:    model.FaceTypeWA60Full,
						FaceName:    "WA 60cm Full",
						ViewBox:     600,
						RenderCross: false,
						Spots:       []model.Spot{{Diameter: 600}},
						Rings:       []model.Ring{{DataScore: 10, Fill: "#FFD700"}},
					},
				}, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp []model.FaceMinimal
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response JSON: %v", err)
		}

		if len(resp) != 2 {
			t.Fatalf("expected 2 face summaries, got %d", len(resp))
		}
		if resp[0].FaceType != model.FaceTypeWA40Full || resp[0].FaceName != "WA 40cm Full" {
			t.Errorf("unexpected first item: %+v", resp[0])
		}
		if resp[1].FaceType != model.FaceTypeWA60Full || resp[1].FaceName != "WA 60cm Full" {
			t.Errorf("unexpected second item: %+v", resp[1])
		}
	})

	t.Run("returns 200 and empty JSON array when no faces exist", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return []model.FaceRead{}, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if body != "[]\n" && body != "[]" {
			t.Fatalf("expected empty array '[]', got %q", body)
		}
	})

	t.Run("returns 200 and empty JSON array when service returns nil slice", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return nil, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if body != "[]\n" && body != "[]" {
			t.Fatalf("expected empty array '[]', got %q", body)
		}
	})

	t.Run("returns 500 when service returns internal error", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return nil, errors.New("database failure")
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})

	t.Run("does not require authentication (public)", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
				return []model.FaceRead{}, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
		rec := httptest.NewRecorder()

		h.ListFaces(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 for unauthenticated request, got %d", rec.Code)
		}
	})
}

func TestFaceHandler_GetFace(t *testing.T) {
	t.Run("returns 200 and full Face definition when found", func(t *testing.T) {
		expectedFace := &model.FaceRead{
			FaceType:    model.FaceTypeWA40Full,
			FaceName:    "WA 40cm Full",
			ViewBox:     400,
			RenderCross: true,
			Spots: []model.Spot{
				{XOffset: 0, YOffset: 0, Diameter: 400},
			},
			Rings: []model.Ring{
				{DataScore: 10, Fill: "#FFD700", R: 20, Stroke: "#000000", StrokeWidth: 1},
				{DataScore: 9, Fill: "#FFD700", R: 40, Stroke: "#000000", StrokeWidth: 1},
			},
		}

		svc := &mockFaceHandlerService{
			getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
				if id == string(model.FaceTypeWA40Full) {
					return expectedFace, nil
				}
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/"+string(model.FaceTypeWA40Full), http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", string(model.FaceTypeWA40Full))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp model.FaceRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response JSON: %v", err)
		}

		if resp.FaceType != expectedFace.FaceType {
			t.Errorf("expected face_type %s, got %s", expectedFace.FaceType, resp.FaceType)
		}
		if resp.FaceName != expectedFace.FaceName {
			t.Errorf("expected face_name %s, got %s", expectedFace.FaceName, resp.FaceName)
		}
		if len(resp.Spots) != 1 || len(resp.Rings) != 2 {
			t.Errorf("unexpected spots or rings length: %+v", resp)
		}
	})

	t.Run("returns 404 when face_type is unknown", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/nonexistent_face", http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", "nonexistent_face")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when face_type URL parameter is empty or whitespace", func(t *testing.T) {
		svc := &mockFaceHandlerService{}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/%20%20%20", http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", "   ")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 500 when service returns internal error", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
				return nil, errors.New("storage error")
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/"+string(model.FaceTypeWA40Full), http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", string(model.FaceTypeWA40Full))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})

	t.Run("does not require authentication (public)", func(t *testing.T) {
		svc := &mockFaceHandlerService{
			getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
				return &model.FaceRead{
					FaceType: model.FaceTypeWA40Full,
					FaceName: "WA 40cm Full",
				}, nil
			},
		}
		h := handler.NewFaceHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/faces/"+string(model.FaceTypeWA40Full), http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("face_type", string(model.FaceTypeWA40Full))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetFace(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 for unauthenticated request, got %d", rec.Code)
		}
	})
}

func TestFaceHandler_Routes(t *testing.T) {
	svc := &mockFaceHandlerService{
		listAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
			return []model.FaceRead{{FaceType: model.FaceTypeWA40Full, FaceName: "WA 40cm Full"}}, nil
		},
		getByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
			if id == string(model.FaceTypeWA40Full) {
				return &model.FaceRead{FaceType: model.FaceTypeWA40Full, FaceName: "WA 40cm Full"}, nil
			}
			return nil, apperror.ErrNotFound
		},
	}
	h := handler.NewFaceHandler(svc)

	r := chi.NewRouter()
	r.Route("/faces", h.Routes)

	t.Run("GET /faces/ routes to ListFaces", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/faces/", http.NoBody)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("GET /faces/{face_type} routes to GetFace", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/faces/"+string(model.FaceTypeWA40Full), http.NoBody)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})
}
