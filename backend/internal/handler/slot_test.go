package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockSlotHandlerService struct {
	getArcherCurrentSlotFn func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error)
	getSlotFn              func(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error)
	joinSessionFn          func(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error)
	reJoinSessionFn        func(ctx context.Context, slotID, archerID uuid.UUID) (*model.SlotJoinResponse, error)
	leaveSessionFn         func(ctx context.Context, slotID, archerID uuid.UUID) error
}

func (m *mockSlotHandlerService) GetArcherCurrentSlot(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
	if m.getArcherCurrentSlotFn != nil {
		return m.getArcherCurrentSlotFn(ctx, archerID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSlotHandlerService) GetSlot(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error) {
	if m.getSlotFn != nil {
		return m.getSlotFn(ctx, slotID)
	}
	return nil, errors.New("unimplemented")
}

//nolint:gocritic // hugeParam: req matches SlotService interface specification
func (m *mockSlotHandlerService) JoinSession(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
	if m.joinSessionFn != nil {
		return m.joinSessionFn(ctx, req)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSlotHandlerService) ReJoinSession(ctx context.Context, slotID, archerID uuid.UUID) (*model.SlotJoinResponse, error) {
	if m.reJoinSessionFn != nil {
		return m.reJoinSessionFn(ctx, slotID, archerID)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockSlotHandlerService) LeaveSession(ctx context.Context, slotID, archerID uuid.UUID) error {
	if m.leaveSessionFn != nil {
		return m.leaveSessionFn(ctx, slotID, archerID)
	}
	return errors.New("unimplemented")
}

func newSlotTestRequest(method, url string, body io.Reader, authArcherID *uuid.UUID, paramKey, paramVal string) *http.Request {
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

func sampleFullSlotInfo(slotID, archerID, sessionID, targetID uuid.UUID) *model.FullSlotInfo {
	return &model.FullSlotInfo{
		SlotID:          slotID,
		TargetID:        targetID,
		ArcherID:        archerID,
		SessionID:       sessionID,
		SlotLetter:      model.SlotLetterA,
		Lane:            1,
		Distance:        18,
		Slot:            "1A",
		FaceType:        model.FaceTypeWA40Full,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      32.5,
		ClubID:          nil,
		IsShooting:      true,
		ShotPerRound:    nil,
		IntervalSeconds: 15,
		CreatedAt:       time.Now().UTC(),
	}
}

func TestSlotHandler_Routes(t *testing.T) {
	mockSvc := &mockSlotHandlerService{
		getArcherCurrentSlotFn: func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
			return sampleFullSlotInfo(uuid.New(), archerID, uuid.New(), uuid.New()), nil
		},
	}
	h := handler.NewSlotHandler(mockSvc)

	r := chi.NewRouter()
	authArcherID := uuid.New()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req = req.WithContext(middleware.WithArcherID(req.Context(), authArcherID))
			next.ServeHTTP(w, req)
		})
	})
	r.Route("/api/v0/session/slot", func(sub chi.Router) {
		h.Routes(sub)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v0/session/slot/archer/"+authArcherID.String(), http.NoBody)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSlotHandler_GetArcherCurrentSlot(t *testing.T) {
	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		req := newSlotTestRequest(http.MethodGet, "/archer/"+uuid.NewString(), nil, nil, "archer_id", uuid.NewString())
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_archer_id_is_invalid", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodGet, "/archer/not-a-uuid", nil, &authID, "archer_id", "not-a-uuid")
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_403_when_authenticated_archer_does_not_match_requested_archer", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		targetArcherID := uuid.New()
		req := newSlotTestRequest(http.MethodGet, "/archer/"+targetArcherID.String(), nil, &authID, "archer_id", targetArcherID.String())
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_404_when_no_active_slot_found", func(t *testing.T) {
		authID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			getArcherCurrentSlotFn: func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/archer/"+authID.String(), nil, &authID, "archer_id", authID.String())
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_with_full_slot_info_when_found", func(t *testing.T) {
		authID := uuid.New()
		expected := sampleFullSlotInfo(uuid.New(), authID, uuid.New(), uuid.New())
		mockSvc := &mockSlotHandlerService{
			getArcherCurrentSlotFn: func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
				return expected, nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/archer/"+authID.String(), nil, &authID, "archer_id", authID.String())
		rr := httptest.NewRecorder()

		h.GetArcherCurrentSlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.FullSlotInfo
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if resp.SlotID != expected.SlotID || resp.Slot != expected.Slot {
			t.Fatalf("mismatched slot info: expected %+v, got %+v", expected, resp)
		}
	})
}

func TestSlotHandler_GetSlot(t *testing.T) {
	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		req := newSlotTestRequest(http.MethodGet, "/"+uuid.NewString(), nil, nil, "slot_id", uuid.NewString())
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_slot_id_is_invalid", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodGet, "/invalid-uuid", nil, &authID, "slot_id", "invalid-uuid")
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_404_when_slot_not_found", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			getSlotFn: func(ctx context.Context, id uuid.UUID) (*model.FullSlotInfo, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_403_when_authenticated_archer_does_not_own_slot", func(t *testing.T) {
		authID := uuid.New()
		slotOwnerID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			getSlotFn: func(ctx context.Context, id uuid.UUID) (*model.FullSlotInfo, error) {
				return sampleFullSlotInfo(slotID, slotOwnerID, uuid.New(), uuid.New()), nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_with_full_slot_info_when_owner_requests", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		expected := sampleFullSlotInfo(slotID, authID, uuid.New(), uuid.New())
		mockSvc := &mockSlotHandlerService{
			getSlotFn: func(ctx context.Context, id uuid.UUID) (*model.FullSlotInfo, error) {
				return expected, nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodGet, "/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.GetSlot(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.FullSlotInfo
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SlotID != slotID || resp.ArcherID != authID {
			t.Fatalf("unexpected response payload: %+v", resp)
		}
	})
}

func TestSlotHandler_JoinSession(t *testing.T) {
	validJoinPayload := func(archerID, sessionID uuid.UUID) model.SlotJoinRequest {
		return model.SlotJoinRequest{
			ArcherID:        archerID,
			SessionID:       sessionID,
			Distance:        18,
			FaceType:        model.FaceTypeWA40Full,
			Bowstyle:        model.BowstyleRecurve,
			DrawWeight:      32.0,
			IsShooting:      true,
			IntervalSeconds: 15,
		}
	}

	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		payload := validJoinPayload(uuid.New(), uuid.New())
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), nil, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_body_is_invalid_JSON", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodPost, "/", strings.NewReader("{invalid"), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_required_ids_are_nil", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		payload := validJoinPayload(uuid.Nil, uuid.Nil)
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_403_when_authenticated_archer_does_not_match_payload_archer", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		otherID := uuid.New()
		payload := validJoinPayload(otherID, uuid.New())
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_409_when_service_returns_conflict", func(t *testing.T) {
		authID := uuid.New()
		sessionID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			joinSessionFn: func(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
				return nil, apperror.Wrap(apperror.ErrConflict, "archer already joined this session")
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		payload := validJoinPayload(authID, sessionID)
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_with_slot_join_response_when_valid", func(t *testing.T) {
		authID := uuid.New()
		sessionID := uuid.New()
		newSlotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			joinSessionFn: func(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
				return &model.SlotJoinResponse{
					SlotID: newSlotID,
					Slot:   "1A",
				}, nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		payload := validJoinPayload(authID, sessionID)
		bodyBytes, _ := json.Marshal(payload)
		req := newSlotTestRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes), &authID, "", "")
		rr := httptest.NewRecorder()

		h.JoinSession(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.SlotJoinResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SlotID != newSlotID || resp.Slot != "1A" {
			t.Fatalf("unexpected join response: %+v", resp)
		}
	})
}

func TestSlotHandler_ReJoinSession(t *testing.T) {
	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		req := newSlotTestRequest(http.MethodPatch, "/re-join/"+uuid.NewString(), nil, nil, "slot_id", uuid.NewString())
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_slot_id_is_invalid", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodPatch, "/re-join/not-uuid", nil, &authID, "slot_id", "not-uuid")
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_404_when_slot_not_found", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			reJoinSessionFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.SlotJoinResponse, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/re-join/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_403_when_user_forbidden_to_rejoin", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			reJoinSessionFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.SlotJoinResponse, error) {
				return nil, apperror.Wrap(apperror.ErrForbidden, "Forbidden")
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/re-join/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_with_slot_join_response_when_rejoined", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			reJoinSessionFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.SlotJoinResponse, error) {
				return &model.SlotJoinResponse{
					SlotID: slotID,
					Slot:   "2B",
				}, nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/re-join/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.ReJoinSession(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp model.SlotJoinResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SlotID != slotID || resp.Slot != "2B" {
			t.Fatalf("unexpected rejoin response: %+v", resp)
		}
	})
}

func TestSlotHandler_LeaveSession(t *testing.T) {
	t.Run("returns_401_when_unauthenticated", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		req := newSlotTestRequest(http.MethodPatch, "/leave/"+uuid.NewString(), nil, nil, "slot_id", uuid.NewString())
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_422_when_slot_id_is_invalid", func(t *testing.T) {
		h := handler.NewSlotHandler(&mockSlotHandlerService{})
		authID := uuid.New()
		req := newSlotTestRequest(http.MethodPatch, "/leave/not-uuid", nil, &authID, "slot_id", "not-uuid")
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_404_when_slot_not_found", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			leaveSessionFn: func(ctx context.Context, sID, aID uuid.UUID) error {
				return apperror.ErrNotFound
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/leave/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_409_when_archer_is_not_participating", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			leaveSessionFn: func(ctx context.Context, sID, aID uuid.UUID) error {
				return apperror.Wrap(apperror.ErrConflict, "archer is not participating in this session")
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/leave/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("returns_200_when_leave_succeeds", func(t *testing.T) {
		authID := uuid.New()
		slotID := uuid.New()
		mockSvc := &mockSlotHandlerService{
			leaveSessionFn: func(ctx context.Context, sID, aID uuid.UUID) error {
				return nil
			},
		}
		h := handler.NewSlotHandler(mockSvc)
		req := newSlotTestRequest(http.MethodPatch, "/leave/"+slotID.String(), nil, &authID, "slot_id", slotID.String())
		rr := httptest.NewRecorder()

		h.LeaveSession(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}
	})
}
