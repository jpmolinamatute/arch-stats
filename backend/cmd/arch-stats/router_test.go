package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/config"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// Mock implementations for router testing

type testAuthService struct {
	authenticateFn func(ctx context.Context, token string) (uuid.UUID, error)
	loginFn        func(ctx context.Context, cred string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error)
	registerFn     func(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error)
	revokeFn       func(ctx context.Context, token string) error
	decodeFn       func(token string) (*auth.Claims, error)
}

func (m *testAuthService) Authenticate(ctx context.Context, token string) (uuid.UUID, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, token)
	}
	if token == "valid-token" {
		return uuid.MustParse("00000000-0000-0000-0000-000000000001"), nil
	}
	return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid token")
}

func (m *testAuthService) LoginWithGoogle(ctx context.Context, cred string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, cred, now, meta...)
	}
	return &model.AuthAuthenticated{
		Status:      model.AuthStatusAuthenticated,
		AccessToken: "test-token",
		ExpiresAt:   now.Add(24 * time.Hour),
	}, nil, nil
}

//nolint:gocritic // hugeParam: payload matches AuthService interface
func (m *testAuthService) RegisterWithGoogle(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, payload, now, meta...)
	}
	return &model.AuthAuthenticated{
		Status:      model.AuthStatusAuthenticated,
		AccessToken: "test-token",
		ExpiresAt:   now.Add(24 * time.Hour),
	}, nil
}

func (m *testAuthService) RevokeToken(ctx context.Context, token string) error {
	if m.revokeFn != nil {
		return m.revokeFn(ctx, token)
	}
	return nil
}

func (m *testAuthService) DecodeToken(token string) (*auth.Claims, error) {
	if m.decodeFn != nil {
		return m.decodeFn(token)
	}
	return &auth.Claims{Sub: "00000000-0000-0000-0000-000000000001", Exp: time.Now().Add(time.Hour).Unix()}, nil
}

type testArcherService struct {
	listFn    func(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	createFn  func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	updateFn  func(ctx context.Context, id uuid.UUID, data model.ArcherSet) error
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

//nolint:gocritic // hugeParam: f matches ArcherService interface specification
func (m *testArcherService) List(ctx context.Context, f model.ArcherFilter) ([]model.ArcherRead, error) {
	if m.listFn != nil {
		return m.listFn(ctx, f)
	}
	return []model.ArcherRead{}, nil
}

func (m *testArcherService) GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &model.ArcherRead{ArcherID: id, FirstName: "Test", LastName: "Archer"}, nil
}

//nolint:gocritic // hugeParam: d matches domain model parameter specification
func (m *testArcherService) Create(ctx context.Context, d model.ArcherCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, d)
	}
	return uuid.New(), nil
}

func (m *testArcherService) Update(ctx context.Context, id uuid.UUID, d model.ArcherSet) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, d)
	}
	return nil
}

func (m *testArcherService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type testSessionService struct {
	getByIDFn          func(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	getOpenFn          func(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	listFn             func(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	createFn           func(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	closeFn            func(ctx context.Context, id uuid.UUID) error
	reOpenFn           func(ctx context.Context, id uuid.UUID) error
	getParticipatingFn func(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)
}

func (m *testSessionService) GetByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &model.SessionRead{SessionID: id}, nil
}

func (m *testSessionService) GetOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error) {
	if m.getOpenFn != nil {
		return m.getOpenFn(ctx, archerID)
	}
	return nil, nil
}

func (m *testSessionService) List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return []model.SessionRead{}, nil
}

func (m *testSessionService) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.New(), nil
}

func (m *testSessionService) Close(ctx context.Context, id uuid.UUID) error {
	if m.closeFn != nil {
		return m.closeFn(ctx, id)
	}
	return nil
}

func (m *testSessionService) ReOpen(ctx context.Context, id uuid.UUID) error {
	if m.reOpenFn != nil {
		return m.reOpenFn(ctx, id)
	}
	return nil
}

func (m *testSessionService) GetParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error) {
	if m.getParticipatingFn != nil {
		return m.getParticipatingFn(ctx, archerID)
	}
	return nil, nil
}

type testSlotService struct {
	getCurrentSlotFn func(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error)
	getSlotFn        func(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error)
	joinSessionFn    func(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error)
	reJoinSessionFn  func(ctx context.Context, slotID, archerID uuid.UUID) (*model.SlotJoinResponse, error)
	leaveSessionFn   func(ctx context.Context, slotID, archerID uuid.UUID) error
}

func (m *testSlotService) GetArcherCurrentSlot(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
	if m.getCurrentSlotFn != nil {
		return m.getCurrentSlotFn(ctx, archerID)
	}
	return &model.FullSlotInfo{SlotID: uuid.New(), ArcherID: archerID}, nil
}

func (m *testSlotService) GetSlot(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error) {
	if m.getSlotFn != nil {
		return m.getSlotFn(ctx, slotID)
	}
	return &model.FullSlotInfo{SlotID: slotID}, nil
}

//nolint:gocritic // hugeParam: req matches SlotService interface specification
func (m *testSlotService) JoinSession(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
	if m.joinSessionFn != nil {
		return m.joinSessionFn(ctx, req)
	}
	return &model.SlotJoinResponse{SlotID: uuid.New()}, nil
}

func (m *testSlotService) ReJoinSession(ctx context.Context, slotID, archerID uuid.UUID) (*model.SlotJoinResponse, error) {
	if m.reJoinSessionFn != nil {
		return m.reJoinSessionFn(ctx, slotID, archerID)
	}
	return &model.SlotJoinResponse{SlotID: slotID}, nil
}

func (m *testSlotService) LeaveSession(ctx context.Context, slotID, archerID uuid.UUID) error {
	if m.leaveSessionFn != nil {
		return m.leaveSessionFn(ctx, slotID, archerID)
	}
	return nil
}

type testShotService struct {
	createFn      func(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error)
	createBatchFn func(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error)
	getBySlotFn   func(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error)
	countBySlotFn func(ctx context.Context, slotID, archerID uuid.UUID) (int, error)
}

func (m *testShotService) Create(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, shot, archerID)
	}
	return uuid.New(), nil
}

func (m *testShotService) CreateBatch(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error) {
	if m.createBatchFn != nil {
		return m.createBatchFn(ctx, shots, archerID)
	}
	return []uuid.UUID{uuid.New()}, nil
}

func (m *testShotService) GetBySlot(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error) {
	if m.getBySlotFn != nil {
		return m.getBySlotFn(ctx, slotID, archerID)
	}
	return []model.ShotRead{}, nil
}

func (m *testShotService) CountBySlot(ctx context.Context, slotID, archerID uuid.UUID) (int, error) {
	if m.countBySlotFn != nil {
		return m.countBySlotFn(ctx, slotID, archerID)
	}
	return 0, nil
}

type testFaceService struct {
	listAllFn func(ctx context.Context) ([]model.FaceRead, error)
	getByIDFn func(ctx context.Context, id string) (*model.FaceRead, error)
}

func (m *testFaceService) ListAll(ctx context.Context) ([]model.FaceRead, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return []model.FaceRead{
		{FaceType: "WA_field", FaceName: "WA Field"},
	}, nil
}

func (m *testFaceService) GetByID(ctx context.Context, id string) (*model.FaceRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	if id == "unknown" {
		return nil, apperror.Wrap(apperror.ErrNotFound, "face not found")
	}
	return &model.FaceRead{FaceType: model.FaceType(id), FaceName: "Test Face"}, nil
}

type testMaintenanceService struct {
	schemaVer int64
	err       error
}

func (m *testMaintenanceService) GetSchemaVersion(ctx context.Context) (int64, error) {
	return m.schemaVer, m.err
}

func setupTestRouter() http.Handler {
	cfg := &config.Config{
		DevMode:       true,
		ServerPort:    8000,
		JWTTTLMinutes: 1440,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	authSvc := &testAuthService{}
	archerSvc := &testArcherService{}
	sessionSvc := &testSessionService{}
	slotSvc := &testSlotService{}
	shotSvc := &testShotService{}
	faceSvc := &testFaceService{}
	maintSvc := &testMaintenanceService{schemaVer: 10}

	mockSPA := handler.NewSPAHandler(fstest.MapFS{
		"index.html":     &fstest.MapFile{Data: []byte("<!DOCTYPE html><html><body>Router Test Index</body></html>")},
		"assets/test.js": &fstest.MapFile{Data: []byte("console.log('router test asset');")},
	}, false, "")

	deps := RouterDeps{
		Cfg:            cfg,
		Logger:         logger,
		AuthSvc:        authSvc,
		AuthHandler:    handler.NewAuthHandler(authSvc, archerSvc, handler.AuthHandlerConfig{DevMode: true, JWTTTLMinutes: 1440}),
		ArcherHandler:  handler.NewArcherHandler(archerSvc),
		SessionHandler: handler.NewSessionHandler(sessionSvc),
		SlotHandler:    handler.NewSlotHandler(slotSvc),
		ShotHandler:    handler.NewShotHandler(shotSvc),
		FaceHandler:    handler.NewFaceHandler(faceSvc),
		HealthHandler:  handler.NewHealthHandler(maintSvc),
		SPAHandler:     mockSPA,
	}

	return buildRouter(&deps)
}

func TestRouter_PublicRoutes(t *testing.T) {
	r := setupTestRouter()

	tests := []struct {
		name       string
		method     string
		url        string
		body       string
		wantStatus int
	}{
		{
			name:       "Health check returns 200",
			method:     http.MethodGet,
			url:        "/api/v0/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Faces list returns 200",
			method:     http.MethodGet,
			url:        "/api/v0/faces",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Faces get by type returns 200",
			method:     http.MethodGet,
			url:        "/api/v0/faces/WA_field",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Auth login returns 200 with valid body",
			method:     http.MethodPost,
			url:        "/api/v0/auth/login",
			body:       `{"credential": "google-test-token"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Auth google alias returns 200 with valid body",
			method:     http.MethodPost,
			url:        "/api/v0/auth/google",
			body:       `{"credential": "google-test-token"}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.url, body)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestRouter_ProtectedRoutes_RequireAuth(t *testing.T) {
	r := setupTestRouter()

	protectedURLs := []struct {
		method string
		url    string
	}{
		{http.MethodPost, "/api/v0/auth/logout"},
		{http.MethodGet, "/api/v0/auth/me"},
		{http.MethodGet, "/api/v0/archer"},
		{http.MethodGet, "/api/v0/session/open"},
		{http.MethodGet, "/api/v0/session/slot/00000000-0000-0000-0000-000000000001"},
		{http.MethodPost, "/api/v0/shot"},
	}

	for _, tt := range protectedURLs {
		t.Run(tt.method+" "+tt.url+" unauthenticated returns 401", func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, http.NoBody)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("got status %d, want 401 Unauthorized for %s %s; body: %s", rec.Code, tt.method, tt.url, rec.Body.String())
			}

			var errResp middleware.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}
			if errResp.Code != "UNAUTHORIZED" {
				t.Errorf("expected code UNAUTHORIZED, got %q", errResp.Code)
			}
		})
	}
}

func TestRouter_ProtectedRoutes_WithValidAuth(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 OK with valid auth; body: %s", rec.Code, rec.Body.String())
	}
}

func TestRouter_CORSPreflight(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodOptions, "/api/v0/faces", http.NoBody)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want 204 No Content", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("missing or incorrect Access-Control-Allow-Origin: %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestRouter_SPAFallbackAndAssets(t *testing.T) {
	r := setupTestRouter()

	tests := []struct {
		name           string
		method         string
		url            string
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:           "Root path serves SPA index",
			method:         http.MethodGet,
			url:            "/",
			wantStatus:     http.StatusOK,
			wantBodySubstr: "Router Test Index",
		},
		{
			name:           "Unknown non-API path falls back to SPA index",
			method:         http.MethodGet,
			url:            "/dashboard/live",
			wantStatus:     http.StatusOK,
			wantBodySubstr: "Router Test Index",
		},
		{
			name:           "Existing asset path serves asset",
			method:         http.MethodGet,
			url:            "/assets/test.js",
			wantStatus:     http.StatusOK,
			wantBodySubstr: "console.log('router test asset')",
		},
		{
			name:           "Unknown API endpoint returns 404",
			method:         http.MethodGet,
			url:            "/api/v0/unknown-endpoint",
			wantStatus:     http.StatusNotFound,
			wantBodySubstr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, http.NoBody)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBodySubstr != "" && !strings.Contains(rec.Body.String(), tt.wantBodySubstr) {
				t.Errorf("expected body to contain %q, got: %s", tt.wantBodySubstr, rec.Body.String())
			}
		})
	}
}

