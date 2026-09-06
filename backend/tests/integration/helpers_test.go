package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
	"github.com/jpmolinamatute/arch-stats/backend/internal/service"
)

// truncateAll truncates all application tables in foreign-key safe order.
func truncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	tables := []string{
		"shot",
		"arrow",
		"slot",
		"target",
		"session",
		"auth",
		"archer",
	}
	for _, table := range tables {
		if _, err := pool.Exec(ctx, "TRUNCATE "+table+" RESTART IDENTITY CASCADE"); err != nil {
			return fmt.Errorf("truncating table %s: %w", table, err)
		}
	}
	return nil
}

// ArcherOverride is a functional modifier to customize test archer creation.
type ArcherOverride func(*model.ArcherCreate)

// createTestArcher inserts a test archer into the database and returns the created record.
func createTestArcher(ctx context.Context, pool *pgxpool.Pool, overrides ...ArcherOverride) (*model.ArcherRead, error) {
	uniqueID := uuid.New().String()
	payload := model.ArcherCreate{
		FirstName:     "Robin",
		LastName:      "Hood",
		Email:         fmt.Sprintf("archer-%s@example.com", uniqueID[:8]),
		DateOfBirth:   "1990-05-15",
		Gender:        model.GenderMale,
		Bowstyle:      model.BowstyleRecurve,
		DrawWeight:    42.5,
		GoogleSubject: fmt.Sprintf("google-sub-%s", uniqueID),
	}

	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewArcherRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test archer: %w", err)
	}

	return repo.FindByID(ctx, id)
}

// SessionOverride is a functional modifier to customize test shooting session creation.
type SessionOverride func(*model.SessionCreate)

// createTestSession inserts a test shooting session into the database and returns the created record.
func createTestSession(ctx context.Context, pool *pgxpool.Pool, archerID uuid.UUID, overrides ...SessionOverride) (*model.SessionRead, error) {
	payload := model.SessionCreate{
		OwnerArcherID:   archerID,
		SessionLocation: "Outdoor Range",
		IsIndoor:        false,
		IsOpened:        true,
	}

	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewSessionRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test session: %w", err)
	}

	return repo.FindByID(ctx, id)
}

// jwtForArcher generates a valid signed HS256 JWT for the given archer UUID and secret.
func jwtForArcher(archerID uuid.UUID, secret string) string {
	now := time.Now().UTC()
	token, err := auth.BuildJWT(archerID, "test-sid", now, now.Add(time.Hour), secret, "HS256")
	if err != nil {
		panic(fmt.Sprintf("jwtForArcher: %v", err))
	}
	return token
}

// TargetOverride is a functional modifier to customize test target creation.
type TargetOverride func(*model.TargetCreate)

// createTestTarget inserts a test target configuration into the database.
//

func createTestTarget(ctx context.Context, pool *pgxpool.Pool, sessionID uuid.UUID, overrides ...TargetOverride) (*model.TargetRead, error) {
	payload := model.TargetCreate{
		SessionID: sessionID,
		Distance:  18,
		Lane:      1,
	}
	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewTargetRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test target: %w", err)
	}

	return repo.FindByID(ctx, id)
}

// SlotOverride is a functional modifier to customize test slot creation.
type SlotOverride func(*model.SlotCreate)

// createTestSlot inserts a test slot into the database.
//

func createTestSlot(ctx context.Context, pool *pgxpool.Pool, targetID, archerID, sessionID uuid.UUID, overrides ...SlotOverride) (*model.SlotRead, error) {
	shotPerRound := 3
	payload := model.SlotCreate{
		TargetID:        targetID,
		ArcherID:        archerID,
		SessionID:       sessionID,
		SlotLetter:      model.SlotLetterA,
		FaceType:        model.FaceTypeWA40Full,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      40.0,
		IsShooting:      true,
		ShotPerRound:    &shotPerRound,
		IntervalSeconds: 20,
	}
	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewSlotRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test slot: %w", err)
	}

	return repo.FindByID(ctx, id)
}

// ShotOverride is a functional modifier to customize test shot creation.
type ShotOverride func(*model.ShotCreate)

// createTestShot inserts a test shot into the database.
//

func createTestShot(ctx context.Context, pool *pgxpool.Pool, slotID uuid.UUID, overrides ...ShotOverride) (*model.ShotRead, error) {
	x := 0.5
	y := 0.5
	score := 10
	payload := model.ShotCreate{
		SlotID: slotID,
		X:      &x,
		Y:      &y,
		IsX:    true,
		Score:  &score,
	}
	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewShotRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test shot: %w", err)
	}

	return repo.FindByID(ctx, id)
}

const testJWTSecret = "integration-test-secret-key-32bytes!"

// testShotServiceAdapter adapts ShotService to handler.ShotService.
type testShotServiceAdapter struct {
	svc *service.ShotService
}

func (a *testShotServiceAdapter) Create(ctx context.Context, shot model.ShotCreate, _ uuid.UUID) (uuid.UUID, error) {
	return a.svc.Create(ctx, shot)
}

func (a *testShotServiceAdapter) CreateBatch(ctx context.Context, shots []model.ShotCreate, _ uuid.UUID) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(shots))
	for _, s := range shots {
		id, err := a.svc.Create(ctx, s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (a *testShotServiceAdapter) GetBySlot(ctx context.Context, slotID, _ uuid.UUID) ([]model.ShotRead, error) {
	return a.svc.ListBySlotID(ctx, slotID)
}

func (a *testShotServiceAdapter) CountBySlot(ctx context.Context, slotID, _ uuid.UUID) (int, error) {
	shots, err := a.svc.ListBySlotID(ctx, slotID)
	if err != nil {
		return 0, err
	}
	return len(shots), nil
}

// testSlotServiceAdapter adapts SlotService to handler.SlotService.
type testSlotServiceAdapter struct {
	svc         *service.SlotService
	slotRepo    *repository.SlotRepo
	sessionRepo *repository.SessionRepo
	targetRepo  *repository.TargetRepo
}

func (a *testSlotServiceAdapter) GetArcherCurrentSlot(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
	isShooting := true
	slots, err := a.slotRepo.FindAll(ctx, model.SlotFilter{
		ArcherID:   &archerID,
		IsShooting: &isShooting,
	})
	if err != nil {
		return nil, fmt.Errorf("finding current slot: %w", err)
	}
	if len(slots) == 0 {
		return nil, apperror.ErrNotFound
	}

	slot := slots[0]
	target, err := a.targetRepo.FindByID(ctx, slot.TargetID)
	if err != nil {
		return nil, fmt.Errorf("finding target: %w", err)
	}
	if target == nil {
		return nil, apperror.ErrNotFound
	}

	return toTestFullSlotInfo(&slot, target), nil
}

func (a *testSlotServiceAdapter) GetSlot(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error) {
	slot, err := a.slotRepo.FindByID(ctx, slotID)
	if err != nil {
		return nil, fmt.Errorf("finding slot: %w", err)
	}
	if slot == nil {
		return nil, apperror.ErrNotFound
	}

	target, err := a.targetRepo.FindByID(ctx, slot.TargetID)
	if err != nil {
		return nil, fmt.Errorf("finding target: %w", err)
	}
	if target == nil {
		return nil, apperror.ErrNotFound
	}

	return toTestFullSlotInfo(slot, target), nil
}

//nolint:gocritic // hugeParam: req matches SlotService interface specification
func (a *testSlotServiceAdapter) JoinSession(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
	session, err := a.sessionRepo.FindByID(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("finding session: %w", err)
	}
	if session == nil || session.ClosedAt != nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "session either does not exist or is closed")
	}

	targets, err := a.targetRepo.FindBySessionID(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("finding session targets: %w", err)
	}

	var targetID uuid.UUID
	var lane int
	for _, t := range targets {
		if t.Distance == req.Distance {
			targetID = t.TargetID
			lane = t.Lane
			break
		}
	}

	if targetID == uuid.Nil {
		lane = len(targets) + 1
		targetID, err = a.targetRepo.Create(ctx, model.TargetCreate{
			SessionID: req.SessionID,
			Distance:  req.Distance,
			Lane:      lane,
		})
		if err != nil {
			return nil, fmt.Errorf("creating target: %w", err)
		}
	}

	letter := model.SlotLetterA
	slotID, err := a.slotRepo.Create(ctx, model.SlotCreate{
		TargetID:        targetID,
		ArcherID:        req.ArcherID,
		SessionID:       req.SessionID,
		SlotLetter:      letter,
		FaceType:        req.FaceType,
		Bowstyle:        req.Bowstyle,
		DrawWeight:      req.DrawWeight,
		ClubID:          req.ClubID,
		ShotPerRound:    req.ShotPerRound,
		IntervalSeconds: req.IntervalSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("creating slot: %w", err)
	}

	return &model.SlotJoinResponse{
		SlotID: slotID,
		Slot:   fmt.Sprintf("%d%s", lane, letter),
	}, nil
}

func (a *testSlotServiceAdapter) ReJoinSession(ctx context.Context, slotID, archerID uuid.UUID) (*model.SlotJoinResponse, error) {
	slot, err := a.slotRepo.FindByID(ctx, slotID)
	if err != nil {
		return nil, fmt.Errorf("finding slot: %w", err)
	}
	if slot == nil {
		return nil, apperror.ErrNotFound
	}
	if slot.ArcherID != archerID {
		return nil, apperror.ErrForbidden
	}

	isShooting := true
	if err := a.slotRepo.Update(ctx, model.SlotSet{IsShooting: &isShooting}, model.SlotFilter{SlotID: &slotID}); err != nil {
		return nil, fmt.Errorf("rejoining slot: %w", err)
	}

	target, err := a.targetRepo.FindByID(ctx, slot.TargetID)
	if err != nil || target == nil {
		return nil, apperror.ErrNotFound
	}

	return &model.SlotJoinResponse{
		SlotID: slotID,
		Slot:   fmt.Sprintf("%d%s", target.Lane, slot.SlotLetter),
	}, nil
}

func (a *testSlotServiceAdapter) LeaveSession(ctx context.Context, slotID, archerID uuid.UUID) error {
	slot, err := a.slotRepo.FindByID(ctx, slotID)
	if err != nil {
		return fmt.Errorf("finding slot: %w", err)
	}
	if slot == nil {
		return apperror.ErrNotFound
	}
	if slot.ArcherID != archerID {
		return apperror.ErrForbidden
	}

	isShooting := false
	if err := a.slotRepo.Update(ctx, model.SlotSet{IsShooting: &isShooting}, model.SlotFilter{SlotID: &slotID}); err != nil {
		return fmt.Errorf("leaving slot: %w", err)
	}
	return nil
}

func toTestFullSlotInfo(slot *model.SlotRead, target *model.TargetRead) *model.FullSlotInfo {
	var createdAt time.Time
	if slot.CreatedAt != nil {
		createdAt = *slot.CreatedAt
	}
	return &model.FullSlotInfo{
		SlotID:          slot.SlotID,
		TargetID:        slot.TargetID,
		ArcherID:        slot.ArcherID,
		SessionID:       slot.SessionID,
		SlotLetter:      slot.SlotLetter,
		Lane:            target.Lane,
		Distance:        target.Distance,
		Slot:            fmt.Sprintf("%d%s", target.Lane, slot.SlotLetter),
		FaceType:        slot.FaceType,
		Bowstyle:        slot.Bowstyle,
		DrawWeight:      slot.DrawWeight,
		ClubID:          slot.ClubID,
		IsShooting:      slot.IsShooting,
		ShotPerRound:    slot.ShotPerRound,
		IntervalSeconds: slot.IntervalSeconds,
		CreatedAt:       createdAt,
	}
}

// buildTestRouter constructs a full chi.Router wired to PostgreSQL testPool.
func buildTestRouter(pool *pgxpool.Pool, customVerifier auth.GooglePayloadVerifier) (chi.Router, *auth.Service) {
	archerRepo := repository.NewArcherRepo(pool)
	authSessionRepo := repository.NewAuthSessionRepo(pool)
	sessionRepo := repository.NewSessionRepo(pool)
	slotRepo := repository.NewSlotRepo(pool)
	shotRepo := repository.NewShotRepo(pool)
	faceRepo := repository.NewFaceRepo(pool)
	targetRepo := repository.NewTargetRepo(pool)
	maintenanceRepo := repository.NewMaintenanceRepo(pool)

	archerSvc := service.NewArcherService(archerRepo)
	sessionSvc := service.NewSessionService(sessionRepo)
	slotSvc := service.NewSlotService(slotRepo, sessionRepo)
	shotSvc := service.NewShotService(shotRepo, slotRepo)
	faceSvc := service.NewFaceService(faceRepo)

	authCfg := auth.Config{
		JWTSecret:           testJWTSecret,
		JWTAlgorithm:        "HS256",
		JWTTTLMinutes:       1440,
		SessionTokenBytes:   32,
		GoogleOAuthClientID: "test-google-client-id",
		GoogleVerifier:      customVerifier,
	}
	authSvc := auth.NewService(archerRepo, authSessionRepo, authCfg)

	authHandlerCfg := handler.AuthHandlerConfig{
		JWTTTLMinutes: 1440,
		DevMode:       true,
	}
	authHandler := handler.NewAuthHandler(authSvc, archerSvc, authHandlerCfg)
	archerHandler := handler.NewArcherHandler(archerSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	slotHandler := handler.NewSlotHandler(&testSlotServiceAdapter{
		svc:         slotSvc,
		slotRepo:    slotRepo,
		sessionRepo: sessionRepo,
		targetRepo:  targetRepo,
	})
	shotHandler := handler.NewShotHandler(&testShotServiceAdapter{svc: shotSvc})
	faceHandler := handler.NewFaceHandler(faceSvc)
	healthHandler := handler.NewHealthHandler(maintenanceRepo)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(slog.Default()))
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS(true))

	r.Route("/api/v0", func(r chi.Router) {
		r.Get("/health", healthHandler.Health)
		r.Route("/faces", faceHandler.Routes)

		r.Route("/auth", func(r chi.Router) {
			r.Use(middleware.ErrorMapper)
			r.Post("/login", authHandler.Login)
			r.Post("/google", authHandler.Login)
			r.Post("/register", authHandler.Register)

			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(authSvc))
				r.Post("/logout", authHandler.Logout)
				r.Get("/me", authHandler.Me)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc))
			r.Use(middleware.ErrorMapper)

			r.Route("/archer", archerHandler.Routes)
			r.Route("/session", func(r chi.Router) {
				r.Route("/slot", slotHandler.Routes)
				sessionHandler.Routes(r)
			})
			r.Route("/shot", shotHandler.Routes)
		})
	})

	return r, authSvc
}

// newTestServer boots an httptest.Server and registers cleanup.
func newTestServer(t *testing.T, customVerifier ...auth.GooglePayloadVerifier) (*httptest.Server, *auth.Service) {
	t.Helper()
	var verifier auth.GooglePayloadVerifier
	if len(customVerifier) > 0 {
		verifier = customVerifier[0]
	}
	router, authSvc := buildTestRouter(testPool, verifier)
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	return ts, authSvc
}

// createAuthenticatedArcher creates an archer, active auth session, and returns valid JWT.
func createAuthenticatedArcher(ctx context.Context, pool *pgxpool.Pool, secret string, overrides ...ArcherOverride) (*model.ArcherRead, string, error) {
	archer, err := createTestArcher(ctx, pool, overrides...)
	if err != nil {
		return nil, "", err
	}

	rawSession, err := auth.GenerateSessionToken(32)
	if err != nil {
		return nil, "", err
	}
	tokenHash := auth.HashSessionToken(rawSession)

	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	authRepo := repository.NewAuthSessionRepo(pool)
	if err := authRepo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: tokenHash,
		CreatedAt:        now,
		ExpiresAt:        expiresAt,
	}); err != nil {
		return nil, "", err
	}

	sid := auth.EncodeSessionID(rawSession)
	jwtToken, err := auth.BuildJWT(archer.ArcherID, sid, now, expiresAt, secret, "HS256")
	if err != nil {
		return nil, "", err
	}

	return archer, jwtToken, nil
}

// authRequest constructs an HTTP request with authentication cookie and header.
func authRequest(method, url, token string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(fmt.Sprintf("authRequest: %v", err))
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.AddCookie(&http.Cookie{Name: middleware.AuthCookieName, Value: token})
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

// doJSONRequest executes an HTTP request, decodes response JSON if target != nil, and returns raw response bytes.
func doJSONRequest(t *testing.T, client *http.Client, req *http.Request, target any) (resp *http.Response, bodyBytes []byte) {
	t.Helper()
	if client == nil {
		client = http.DefaultClient
	}
	var err error
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("client.Do failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response body: %v", err)
	}

	if target != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, target); err != nil {
			t.Fatalf("unmarshaling json response: %v\nraw: %s", err, string(bodyBytes))
		}
	}
	return resp, bodyBytes
}
