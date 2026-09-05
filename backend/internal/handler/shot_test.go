package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockShotHandlerService struct {
	createFn      func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error)
	createBatchFn func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error)
	getBySlotFn   func(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error)
	countBySlotFn func(ctx context.Context, slotID, archerID uuid.UUID) (int, error)
}

func (m *mockShotHandlerService) Create(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, shot, archerID)
	}
	return uuid.Nil, errors.New("unimplemented")
}

func (m *mockShotHandlerService) CreateBatch(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
	if m.createBatchFn != nil {
		return m.createBatchFn(ctx, shots, archerID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockShotHandlerService) GetBySlot(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error) {
	if m.getBySlotFn != nil {
		return m.getBySlotFn(ctx, slotID, archerID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockShotHandlerService) CountBySlot(ctx context.Context, slotID, archerID uuid.UUID) (int, error) {
	if m.countBySlotFn != nil {
		return m.countBySlotFn(ctx, slotID, archerID)
	}
	return 0, errors.New("unimplemented")
}

func newShotTestRequest(method, url string, body io.Reader, authArcherID *uuid.UUID, paramKey, paramVal string) *http.Request {
	req := httptest.NewRequest(method, url, body)
	if paramKey != "" {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add(paramKey, paramVal)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}
	if authArcherID != nil {
		req = req.WithContext(middleware.WithArcherID(req.Context(), *authArcherID))
	}
	return req
}

func sampleShotRead(shotID, slotID uuid.UUID, score int) model.ShotRead {
	x := 5.0
	y := 5.0
	return model.ShotRead{
		ShotID:    shotID,
		SlotID:    slotID,
		X:         &x,
		Y:         &y,
		IsX:       score == 10,
		Score:     &score,
		ArrowID:   nil,
		CreatedAt: time.Now().UTC(),
	}
}

func TestShotHandler_Routes(t *testing.T) {
	svc := &mockShotHandlerService{}
	h := handler.NewShotHandler(svc)

	r := chi.NewRouter()
	r.Route("/api/v0/shot", func(sub chi.Router) {
		h.Routes(sub)
	})

	slotID := uuid.New()
	routes := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v0/shot"},
		{method: http.MethodGet, path: "/api/v0/shot/by-slot/" + slotID.String()},
		{method: http.MethodGet, path: "/api/v0/shot/count-by-slot/" + slotID.String()},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			req := httptest.NewRequest(rt.method, rt.path, http.NoBody)
			rctx := chi.NewRouteContext()
			if !r.Match(rctx, req.Method, req.URL.Path) {
				t.Fatalf("expected route %s %s to match router", rt.method, rt.path)
			}
		})
	}
}

func TestShotHandler_Create(t *testing.T) {
	authArcherID := uuid.New()
	slotID := uuid.New()
	score := 9
	xVal := 2.5
	yVal := 3.5

	singlePayload := model.ShotCreate{
		SlotID: slotID,
		X:      &xVal,
		Y:      &yVal,
		Score:  &score,
	}

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), nil, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("returns 422 when body is empty", func(t *testing.T) {
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader([]byte("")), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("returns 422 when body is invalid JSON", func(t *testing.T) {
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader([]byte("{invalid-json")), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("single: returns 201 with ShotID on success", func(t *testing.T) {
		newShotID := uuid.New()
		svc := &mockShotHandlerService{
			createFn: func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
				if archerID != authArcherID {
					t.Errorf("expected archerID %s, got %s", authArcherID, archerID)
				}
				if shot.SlotID != slotID {
					t.Errorf("expected slotID %s, got %s", slotID, shot.SlotID)
				}
				return newShotID, nil
			},
		}

		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.ShotID
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ShotID != newShotID {
			t.Fatalf("expected shotID %s, got %s", newShotID, resp.ShotID)
		}
	})

	t.Run("single: returns 404 when slot does not exist", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createFn: func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
				return uuid.Nil, apperror.ErrNotFound
			},
		}

		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("single: returns 403 when archer does not own slot", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createFn: func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
				return uuid.Nil, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}

		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("single: returns 422 when validation fails", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createFn: func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
				return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "score must be between 0 and 10")
			},
		}

		body, _ := json.Marshal(singlePayload)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("batch: returns 400 when empty array", func(t *testing.T) {
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader([]byte("[]")), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
		var errResp middleware.ErrorResponse
		_ = json.NewDecoder(rr.Body).Decode(&errResp)
		if errResp.Detail != "Invalid input" {
			t.Fatalf("expected detail 'Invalid input', got %q", errResp.Detail)
		}
	})

	t.Run("batch: returns 400 when fewer than 3 shots", func(t *testing.T) {
		twoShots := []model.ShotCreate{singlePayload, singlePayload}
		body, _ := json.Marshal(twoShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
		var errResp middleware.ErrorResponse
		_ = json.NewDecoder(rr.Body).Decode(&errResp)
		if errResp.Detail != "Invalid input" {
			t.Fatalf("expected detail 'Invalid input', got %q", errResp.Detail)
		}
	})

	t.Run("batch: returns 400 when more than 10 shots", func(t *testing.T) {
		elevenShots := make([]model.ShotCreate, 11)
		for i := range elevenShots {
			elevenShots[i] = singlePayload
		}
		body, _ := json.Marshal(elevenShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
		var errResp middleware.ErrorResponse
		_ = json.NewDecoder(rr.Body).Decode(&errResp)
		if errResp.Detail != "Invalid input" {
			t.Fatalf("expected detail 'Invalid input', got %q", errResp.Detail)
		}
	})

	t.Run("batch: returns 400 when shots belong to different slots", func(t *testing.T) {
		otherSlotID := uuid.New()
		diffSlotShots := []model.ShotCreate{
			singlePayload,
			singlePayload,
			{SlotID: otherSlotID, X: &xVal, Y: &yVal, Score: &score},
		}
		body, _ := json.Marshal(diffSlotShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
		var errResp middleware.ErrorResponse
		_ = json.NewDecoder(rr.Body).Decode(&errResp)
		if errResp.Detail != "All shots must belong to the same slot" {
			t.Fatalf("expected detail 'All shots must belong to the same slot', got %q", errResp.Detail)
		}
	})

	t.Run("batch: returns 201 with array of ShotID on success", func(t *testing.T) {
		id1, id2, id3 := uuid.New(), uuid.New(), uuid.New()
		svc := &mockShotHandlerService{
			createBatchFn: func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
				if archerID != authArcherID {
					t.Errorf("expected archerID %s, got %s", authArcherID, archerID)
				}
				if len(shots) != 3 {
					t.Errorf("expected 3 shots, got %d", len(shots))
				}
				return []uuid.UUID{id1, id2, id3}, nil
			},
		}

		threeShots := []model.ShotCreate{singlePayload, singlePayload, singlePayload}
		body, _ := json.Marshal(threeShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp []model.ShotID
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp) != 3 {
			t.Fatalf("expected 3 items in response, got %d", len(resp))
		}
		if resp[0].ShotID != id1 || resp[1].ShotID != id2 || resp[2].ShotID != id3 {
			t.Fatalf("unexpected shot IDs returned")
		}
	})

	t.Run("batch: returns 403 when forbidden", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createBatchFn: func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
				return nil, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}

		threeShots := []model.ShotCreate{singlePayload, singlePayload, singlePayload}
		body, _ := json.Marshal(threeShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("batch: returns 404 when slot not found", func(t *testing.T) {
		svc := &mockShotHandlerService{
			createBatchFn: func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
				return nil, apperror.ErrNotFound
			},
		}

		threeShots := []model.ShotCreate{singlePayload, singlePayload, singlePayload}
		body, _ := json.Marshal(threeShots)
		req := newShotTestRequest(http.MethodPost, "/api/v0/shot", bytes.NewReader(body), &authArcherID, "", "")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.Create(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})
}

func TestShotHandler_GetBySlot(t *testing.T) {
	authArcherID := uuid.New()
	slotID := uuid.New()

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, nil, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("returns 422 when slot_id is invalid", func(t *testing.T) {
		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/invalid-uuid", nil, &authArcherID, "slot_id", "invalid-uuid")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("returns 404 when slot not found", func(t *testing.T) {
		svc := &mockShotHandlerService{
			getBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) ([]model.ShotRead, error) {
				return nil, apperror.ErrNotFound
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("returns 403 when archer does not own slot", func(t *testing.T) {
		svc := &mockShotHandlerService{
			getBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) ([]model.ShotRead, error) {
				return nil, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("returns 200 with empty array when no shots recorded", func(t *testing.T) {
		svc := &mockShotHandlerService{
			getBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) ([]model.ShotRead, error) {
				return nil, nil
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp []model.ShotRead
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}
		if resp == nil || len(resp) != 0 {
			t.Fatalf("expected non-nil empty slice, got %v", resp)
		}
	})

	t.Run("returns 200 with shots array on success", func(t *testing.T) {
		s1 := sampleShotRead(uuid.New(), slotID, 10)
		s2 := sampleShotRead(uuid.New(), slotID, 9)
		svc := &mockShotHandlerService{
			getBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) ([]model.ShotRead, error) {
				return []model.ShotRead{s1, s2}, nil
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.GetBySlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp []model.ShotRead
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("expected 2 shots, got %d", len(resp))
		}
		if resp[0].ShotID != s1.ShotID || resp[1].ShotID != s2.ShotID {
			t.Fatalf("unexpected shots returned")
		}
	})
}

func TestShotHandler_CountBySlot(t *testing.T) {
	authArcherID := uuid.New()
	slotID := uuid.New()

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, nil, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("returns 422 when slot_id is invalid", func(t *testing.T) {
		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/invalid-uuid", nil, &authArcherID, "slot_id", "invalid-uuid")
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(&mockShotHandlerService{})
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rr.Code)
		}
	})

	t.Run("returns 404 when slot not found", func(t *testing.T) {
		svc := &mockShotHandlerService{
			countBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) (int, error) {
				return 0, apperror.ErrNotFound
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("returns 403 when archer does not own slot", func(t *testing.T) {
		svc := &mockShotHandlerService{
			countBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) (int, error) {
				return 0, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("returns 200 with 0 count when slot has no shots", func(t *testing.T) {
		svc := &mockShotHandlerService{
			countBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) (int, error) {
				return 0, nil
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var count int
		if err := json.NewDecoder(rr.Body).Decode(&count); err != nil {
			t.Fatalf("failed to decode integer response: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected count 0, got %d", count)
		}
	})

	t.Run("returns 200 with count when slot has shots", func(t *testing.T) {
		svc := &mockShotHandlerService{
			countBySlotFn: func(ctx context.Context, sID, aID uuid.UUID) (int, error) {
				return 6, nil
			},
		}

		req := newShotTestRequest(http.MethodGet, "/api/v0/shot/count-by-slot/"+slotID.String(), nil, &authArcherID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h := handler.NewShotHandler(svc)
		h.CountBySlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var count int
		if err := json.NewDecoder(rr.Body).Decode(&count); err != nil {
			t.Fatalf("failed to decode integer response: %v", err)
		}
		if count != 6 {
			t.Fatalf("expected count 6, got %d", count)
		}
	})
}
