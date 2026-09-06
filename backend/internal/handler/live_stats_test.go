package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	ws "github.com/jpmolinamatute/arch-stats/backend/internal/websocket"
)

type mockLiveStatsService struct {
	getStatsFn func(ctx context.Context, slotID, archerID uuid.UUID) (*model.LiveStat, error)
}

func (m *mockLiveStatsService) GetStats(ctx context.Context, slotID, archerID uuid.UUID) (*model.LiveStat, error) {
	if m.getStatsFn != nil {
		return m.getStatsFn(ctx, slotID, archerID)
	}
	return nil, errors.New("unimplemented")
}

type mockWebSocketHub struct {
	mu           sync.Mutex
	registered   []*ws.Client
	unregistered []*ws.Client
	onRegister   func(client *ws.Client)
	onUnregister func(client *ws.Client)
}

func (m *mockWebSocketHub) Register(client *ws.Client) {
	m.mu.Lock()
	m.registered = append(m.registered, client)
	m.mu.Unlock()
	if m.onRegister != nil {
		m.onRegister(client)
	}
}

func (m *mockWebSocketHub) Unregister(client *ws.Client) {
	m.mu.Lock()
	m.unregistered = append(m.unregistered, client)
	m.mu.Unlock()
	if m.onUnregister != nil {
		m.onUnregister(client)
	}
}

func newLiveStatsTestRequest(method, url string, body io.Reader, authArcherID *uuid.UUID, paramKey, paramVal string) *http.Request {
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

func sampleLiveStat(slotID uuid.UUID) *model.LiveStat {
	now := time.Now().UTC().Truncate(time.Second)
	return &model.LiveStat{
		Scores: []model.ShotScore{
			{
				ShotID:    uuid.New(),
				Score:     10,
				IsX:       true,
				CreatedAt: now,
			},
			{
				ShotID:    uuid.New(),
				Score:     9,
				IsX:       false,
				CreatedAt: now.Add(time.Minute),
			},
		},
		Stats: model.Stats{
			SlotID:        slotID,
			NumberOfShots: 2,
			TotalScore:    19,
			MaxScore:      10,
			Mean:          9.5,
		},
	}
}

func TestLiveStatsHandler_GetStats_Success(t *testing.T) {
	slotID := uuid.New()
	archerID := uuid.New()
	expected := sampleLiveStat(slotID)

	svc := &mockLiveStatsService{
		getStatsFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.LiveStat, error) {
			if sID != slotID {
				t.Errorf("expected slotID %v, got %v", slotID, sID)
			}
			if aID != archerID {
				t.Errorf("expected archerID %v, got %v", archerID, aID)
			}
			return expected, nil
		},
	}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/"+slotID.String(), nil, &archerID, "slot_id", slotID.String())
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var got model.LiveStat
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}

	if got.Stats.SlotID != slotID {
		t.Errorf("expected slot ID %v, got %v", slotID, got.Stats.SlotID)
	}
	if got.Stats.NumberOfShots != 2 || got.Stats.TotalScore != 19 || got.Stats.Mean != 9.5 {
		t.Errorf("unexpected stats values: %+v", got.Stats)
	}
	if len(got.Scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(got.Scores))
	}
	if got.Scores[0].Score != 10 || !got.Scores[0].IsX {
		t.Errorf("unexpected first score: %+v", got.Scores[0])
	}
}

func TestLiveStatsHandler_GetStats_NotFound(t *testing.T) {
	slotID := uuid.New()
	archerID := uuid.New()

	svc := &mockLiveStatsService{
		getStatsFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.LiveStat, error) {
			return nil, apperror.ErrNotFound
		},
	}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/"+slotID.String(), nil, &archerID, "slot_id", slotID.String())
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_GetStats_InvalidSlotID(t *testing.T) {
	archerID := uuid.New()
	svc := &mockLiveStatsService{}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/not-a-uuid", nil, &archerID, "slot_id", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422 Unprocessable Entity, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_GetStats_Unauthenticated(t *testing.T) {
	slotID := uuid.New()
	svc := &mockLiveStatsService{}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	// No archer ID in context
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/"+slotID.String(), nil, nil, "slot_id", slotID.String())
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_GetStats_InternalError(t *testing.T) {
	slotID := uuid.New()
	archerID := uuid.New()

	svc := &mockLiveStatsService{
		getStatsFn: func(ctx context.Context, sID, aID uuid.UUID) (*model.LiveStat, error) {
			return nil, errors.New("db query failed")
		},
	}

	h := handler.NewLiveStatsHandler(svc, &mockWebSocketHub{})
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/"+slotID.String(), nil, &archerID, "slot_id", slotID.String())
	rec := httptest.NewRecorder()

	h.GetStats(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 Internal Server Error, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_WebSocketStats_InvalidSlotID(t *testing.T) {
	svc := &mockLiveStatsService{}
	hub := &mockWebSocketHub{}

	h := handler.NewLiveStatsHandler(svc, hub)
	req := newLiveStatsTestRequest(http.MethodGet, "/api/v0/stats/ws/not-a-uuid", nil, nil, "slot_id", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.WebSocketStats(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422 Unprocessable Entity, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveStatsHandler_WebSocketStats_UpgradeAndLifecycle(t *testing.T) {
	svc := &mockLiveStatsService{}

	regChan := make(chan *ws.Client, 1)
	unregChan := make(chan *ws.Client, 1)

	hub := &mockWebSocketHub{
		onRegister: func(client *ws.Client) {
			regChan <- client
		},
		onUnregister: func(client *ws.Client) {
			unregChan <- client
		},
	}

	h := handler.NewLiveStatsHandler(svc, hub)

	r := chi.NewRouter()
	r.Route("/api/v0/stats", h.Routes)

	server := httptest.NewServer(r)
	defer server.Close()

	slotID := uuid.New()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v0/stats/ws/" + slotID.String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := coderws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	// Verify client is registered
	select {
	case client := <-regChan:
		if client == nil {
			t.Fatal("expected non-nil registered client")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client registration")
	}

	// Close client connection to trigger unregister
	_ = conn.Close(coderws.StatusNormalClosure, "test closing")

	select {
	case client := <-unregChan:
		if client == nil {
			t.Fatal("expected non-nil unregistered client")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client unregistration")
	}
}

func TestLiveStatsHandler_Routes_Registration(t *testing.T) {
	svc := &mockLiveStatsService{}
	hub := &mockWebSocketHub{}
	h := handler.NewLiveStatsHandler(svc, hub)

	r := chi.NewRouter()
	h.Routes(r)

	routes := r.Routes()
	if len(routes) == 0 {
		t.Fatal("expected routes to be registered")
	}

	hasStatsRoute := false
	hasWSRoute := false

	for _, route := range routes {
		if strings.Contains(route.Pattern, "{slot_id}") {
			hasStatsRoute = true
		}
		if strings.Contains(route.Pattern, "ws") {
			hasWSRoute = true
		}
	}

	if !hasStatsRoute {
		t.Error("expected {slot_id} route pattern to be registered")
	}
	if !hasWSRoute {
		t.Error("expected ws/{slot_id} route pattern to be registered")
	}
}
