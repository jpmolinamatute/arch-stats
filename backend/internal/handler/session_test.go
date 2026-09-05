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

func (m *mockSessionHandlerService) List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, errors.New("unimplemented")
}

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

func newSessionTestRequest(method, url string, body io.Reader, authArcherID *uuid.UUID, paramKey, paramVal string) *http.Request {
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

func sampleSessionReadData(id, archerID uuid.UUID, isOpen bool) *model.SessionRead {
	now := time.Now().UTC()
	var closedAt *time.Time
	if !isOpen {
		c := now
		closedAt = &c
	}
	return &model.SessionRead{
		SessionID:       id,
		OwnerArcherID:   archerID,
		SessionLocation: "Main Range",
		IsIndoor:        false,
		IsOpened:        isOpen,
		CreatedAt:       now,
		ClosedAt:        closedAt,
	}
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

func TestSessionHandler_GetOpenForArcher(t *testing.T) {
	archerID := uuid.New()
	sessionID := uuid.New()

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/open-session", http.NoBody, nil, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetOpenForArcher(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when archer_id is invalid", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/not-a-uuid/open-session", http.NoBody, &archerID, "archer_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.GetOpenForArcher(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 403 when authenticated archer does not match requested archer", func(t *testing.T) {
		otherID := uuid.New()
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/open-session", http.NoBody, &otherID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetOpenForArcher(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("returns 200 with null session_id when no open session exists", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getOpenFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/open-session", http.NoBody, &archerID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetOpenForArcher(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp model.SessionID
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SessionID != nil {
			t.Fatalf("expected nil session_id, got %v", *resp.SessionID)
		}
	})

	t.Run("returns 200 with session_id when open session exists", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getOpenFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, archerID, true), nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/open-session", http.NoBody, &archerID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetOpenForArcher(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp model.SessionID
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SessionID == nil || *resp.SessionID != sessionID {
			t.Fatalf("expected session_id %v, got %v", sessionID, resp.SessionID)
		}
	})

	t.Run("returns 500 when service fails unexpectedly", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getOpenFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return nil, errors.New("db error")
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/open-session", http.NoBody, &archerID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetOpenForArcher(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestSessionHandler_GetClosedForArcher(t *testing.T) {
	archerID := uuid.New()
	s1 := uuid.New()
	s2 := uuid.New()

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/close-session", http.NoBody, nil, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetClosedForArcher(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when archer_id is invalid", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/invalid/close-session", http.NoBody, &archerID, "archer_id", "invalid")
		rec := httptest.NewRecorder()

		h.GetClosedForArcher(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 403 when authenticated archer does not match requested archer", func(t *testing.T) {
		otherID := uuid.New()
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/close-session", http.NoBody, &otherID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetClosedForArcher(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("returns 200 with empty array when no closed sessions exist", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			listFn: func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
				return []model.SessionRead{}, nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/close-session", http.NoBody, &archerID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetClosedForArcher(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		body := strings.TrimSpace(rec.Body.String())
		if body != "[]" {
			t.Fatalf("expected '[]', got %q", body)
		}
	})

	t.Run("returns 200 with closed sessions array when found", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			listFn: func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
				if filter.OwnerArcherID == nil || *filter.OwnerArcherID != archerID {
					t.Errorf("expected filter.OwnerArcherID = %v, got %v", archerID, filter.OwnerArcherID)
				}
				if filter.IsOpened == nil || *filter.IsOpened != false {
					t.Errorf("expected filter.IsOpened = false, got %v", filter.IsOpened)
				}
				return []model.SessionRead{
					*sampleSessionReadData(s1, archerID, false),
					*sampleSessionReadData(s2, archerID, false),
				}, nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/close-session", http.NoBody, &archerID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetClosedForArcher(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp []model.SessionRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp) != 2 || resp[0].SessionID != s1 || resp[1].SessionID != s2 {
			t.Fatalf("unexpected sessions list: %+v", resp)
		}
	})
}

func TestSessionHandler_GetParticipating(t *testing.T) {
	archerID := uuid.New()
	sessionID := uuid.New()

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/participating", http.NoBody, nil, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetParticipating(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when archer_id is invalid", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/bad/participating", http.NoBody, &archerID, "archer_id", "bad")
		rec := httptest.NewRecorder()

		h.GetParticipating(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 403 when authenticated archer does not match requested archer", func(t *testing.T) {
		otherID := uuid.New()
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/participating", http.NoBody, &otherID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetParticipating(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("returns 200 with null session_id when not participating", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getParticipatingFn: func(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
				return nil, nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/participating", http.NoBody, &archerID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetParticipating(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp model.SessionID
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SessionID != nil {
			t.Fatalf("expected nil session_id, got %v", *resp.SessionID)
		}
	})

	t.Run("returns 200 with session_id when participating", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getParticipatingFn: func(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
				return &sessionID, nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/archer/"+archerID.String()+"/participating", http.NoBody, &archerID, "archer_id", archerID.String())
		rec := httptest.NewRecorder()

		h.GetParticipating(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp model.SessionID
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SessionID == nil || *resp.SessionID != sessionID {
			t.Fatalf("expected session_id %v, got %v", sessionID, resp.SessionID)
		}
	})
}

func TestSessionHandler_ListAllOpen(t *testing.T) {
	authID := uuid.New()
	s1 := uuid.New()

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/open", http.NoBody, nil, "", "")
		rec := httptest.NewRecorder()

		h.ListAllOpen(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("returns 200 with empty array when no open sessions exist", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			listFn: func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
				return nil, nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/open", http.NoBody, &authID, "", "")
		rec := httptest.NewRecorder()

		h.ListAllOpen(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		body := strings.TrimSpace(rec.Body.String())
		if body != "[]" {
			t.Fatalf("expected '[]', got %q", body)
		}
	})

	t.Run("returns 200 with array when open sessions exist", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			listFn: func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
				if filter.IsOpened == nil || *filter.IsOpened != true {
					t.Errorf("expected filter.IsOpened = true, got %v", filter.IsOpened)
				}
				return []model.SessionRead{
					*sampleSessionReadData(s1, authID, true),
				}, nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/open", http.NoBody, &authID, "", "")
		rec := httptest.NewRecorder()

		h.ListAllOpen(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp []model.SessionRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp) != 1 || resp[0].SessionID != s1 {
			t.Fatalf("unexpected sessions returned: %+v", resp)
		}
	})
}

func TestSessionHandler_GetByID(t *testing.T) {
	authID := uuid.New()
	sessionID := uuid.New()

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/"+sessionID.String(), http.NoBody, nil, "id", sessionID.String())
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when id is not a valid UUID", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/invalid", http.NoBody, &authID, "id", "invalid")
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 404 when session not found", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/"+sessionID.String(), http.NoBody, &authID, "id", sessionID.String())
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("returns 403 when session is closed and user is not owner", func(t *testing.T) {
		ownerID := uuid.New()
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, ownerID, false), nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/"+sessionID.String(), http.NoBody, &authID, "id", sessionID.String())
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("returns 200 when session is closed and user IS owner", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, authID, false), nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/"+sessionID.String(), http.NoBody, &authID, "id", sessionID.String())
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var resp model.SessionRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SessionID != sessionID {
			t.Fatalf("unexpected session ID: %v", resp.SessionID)
		}
	})

	t.Run("returns 200 when session is open even if user is not owner", func(t *testing.T) {
		ownerID := uuid.New()
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, ownerID, true), nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodGet, "/api/v0/session/"+sessionID.String(), http.NoBody, &authID, "id", sessionID.String())
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var resp model.SessionRead
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SessionID != sessionID {
			t.Fatalf("unexpected session ID: %v", resp.SessionID)
		}
	})
}

func TestSessionHandler_Create(t *testing.T) {
	authID := uuid.New()
	sessionID := uuid.New()

	validPayload := func(owner uuid.UUID) []byte {
		data, _ := json.Marshal(model.SessionCreate{
			OwnerArcherID:   owner,
			SessionLocation: "Main Range",
			IsIndoor:        false,
			IsOpened:        true,
		})
		return data
	}

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodPost, "/api/v0/session", bytes.NewReader(validPayload(authID)), nil, "", "")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when body is invalid JSON", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodPost, "/api/v0/session", strings.NewReader("{invalid json"), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 403 when creating session for another archer", func(t *testing.T) {
		otherID := uuid.New()
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodPost, "/api/v0/session", bytes.NewReader(validPayload(otherID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
		var errResp middleware.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if !strings.Contains(errResp.Detail, "user not allowed to open a session for another archer") {
			t.Fatalf("unexpected error detail: %s", errResp.Detail)
		}
	})

	t.Run("returns 409 when archer already has an open session", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			createFn: func(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
				return uuid.Nil, apperror.Wrap(apperror.ErrConflict, "archer already has an open session")
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPost, "/api/v0/session", bytes.NewReader(validPayload(authID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when service returns validation error", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			createFn: func(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
				return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "session_location is required")
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPost, "/api/v0/session", bytes.NewReader(validPayload(authID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 201 with session_id when successfully created", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			createFn: func(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
				if data.OwnerArcherID != authID || data.SessionLocation != "Main Range" {
					t.Errorf("unexpected create data: %+v", data)
				}
				return sessionID, nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPost, "/api/v0/session", bytes.NewReader(validPayload(authID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rec.Code)
		}
		var resp model.SessionID
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SessionID == nil || *resp.SessionID != sessionID {
			t.Fatalf("expected session_id %v, got %v", sessionID, resp.SessionID)
		}
	})
}

func TestSessionHandler_ReOpen(t *testing.T) {
	authID := uuid.New()
	sessionID := uuid.New()

	payload := func(sid *uuid.UUID) []byte {
		data, _ := json.Marshal(model.SessionID{SessionID: sid})
		return data
	}

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/re-open", bytes.NewReader(payload(&sessionID)), nil, "", "")
		rec := httptest.NewRecorder()

		h.ReOpen(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when session_id is missing or nil", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		nilUUID := uuid.Nil
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/re-open", bytes.NewReader(payload(&nilUUID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.ReOpen(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 404 when session does not exist", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/re-open", bytes.NewReader(payload(&sessionID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.ReOpen(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("returns 403 when authenticated archer is not the session owner", func(t *testing.T) {
		ownerID := uuid.New()
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, ownerID, false), nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/re-open", bytes.NewReader(payload(&sessionID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.ReOpen(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when session is already open", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, authID, true), nil
			},
			reOpenFn: func(ctx context.Context, id uuid.UUID) error {
				return apperror.Wrap(apperror.ErrValidation, "session is already open")
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/re-open", bytes.NewReader(payload(&sessionID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.ReOpen(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 409 when owner has another conflicting open session", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, authID, false), nil
			},
			reOpenFn: func(ctx context.Context, id uuid.UUID) error {
				return apperror.Wrap(apperror.ErrConflict, "archer already has an open session")
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/re-open", bytes.NewReader(payload(&sessionID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.ReOpen(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", rec.Code)
		}
	})

	t.Run("returns 200 with session_id when successfully reopened", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, authID, false), nil
			},
			reOpenFn: func(ctx context.Context, id uuid.UUID) error {
				if id != sessionID {
					t.Errorf("expected id %v, got %v", sessionID, id)
				}
				return nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/re-open", bytes.NewReader(payload(&sessionID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.ReOpen(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var resp model.SessionID
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.SessionID == nil || *resp.SessionID != sessionID {
			t.Fatalf("expected session_id %v, got %v", sessionID, resp.SessionID)
		}
	})
}

func TestSessionHandler_Close(t *testing.T) {
	authID := uuid.New()
	sessionID := uuid.New()

	payload := func(sid *uuid.UUID) []byte {
		data, _ := json.Marshal(model.SessionID{SessionID: sid})
		return data
	}

	t.Run("returns 401 when unauthenticated", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/close", bytes.NewReader(payload(&sessionID)), nil, "", "")
		rec := httptest.NewRecorder()

		h.Close(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("returns 400 when session_id is missing", func(t *testing.T) {
		h := handler.NewSessionHandler(&mockSessionHandlerService{})
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/close", strings.NewReader("{}"), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Close(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
		var errResp middleware.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Detail != "ERROR: session_id wasn't provided" {
			t.Fatalf("unexpected detail message: %q", errResp.Detail)
		}
	})

	t.Run("returns 404 when session does not exist", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return nil, apperror.ErrNotFound
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/close", bytes.NewReader(payload(&sessionID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Close(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("returns 403 when authenticated archer is not owner", func(t *testing.T) {
		ownerID := uuid.New()
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, ownerID, true), nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/close", bytes.NewReader(payload(&sessionID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Close(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("returns 422 when session is already closed", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, authID, false), nil
			},
			closeFn: func(ctx context.Context, id uuid.UUID) error {
				return apperror.Wrap(apperror.ErrValidation, "session is already closed")
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/close", bytes.NewReader(payload(&sessionID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Close(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})

	t.Run("returns 200 with status closed when successful", func(t *testing.T) {
		svc := &mockSessionHandlerService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
				return sampleSessionReadData(sessionID, authID, true), nil
			},
			closeFn: func(ctx context.Context, id uuid.UUID) error {
				if id != sessionID {
					t.Errorf("expected session id %v, got %v", sessionID, id)
				}
				return nil
			},
		}
		h := handler.NewSessionHandler(svc)
		req := newSessionTestRequest(http.MethodPatch, "/api/v0/session/close", bytes.NewReader(payload(&sessionID)), &authID, "", "")
		rec := httptest.NewRecorder()

		h.Close(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp["status"] != "closed" {
			t.Fatalf("expected status 'closed', got %q", resp["status"])
		}
	})
}
