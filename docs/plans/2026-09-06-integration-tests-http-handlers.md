# HTTP Handlers Integration Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port existing Python endpoint tests from `backend-old/tests/endpoints/` to comprehensive Go integration tests (`backend/tests/integration/endpoint_*_test.go`), verifying the full HTTP stack (handler → service → repository → PostgreSQL) with `httptest.Server`, chi router, and testcontainers PostgreSQL, and mark Task 040 as completed in task tracking documents.

**Architecture:** Integration tests reside in `backend/tests/integration/` under `package integration_test`, executing against the shared PostgreSQL 17 testcontainers instance (`testPool`) initialized in `TestMain`. Tests instantiate a complete `chi.Router` via `newTestServer(t)` with real repositories, services, and handlers. Helper utilities in `helpers_test.go` provide authenticated request construction (`authRequest`), archer session generation (`createAuthenticatedArcher`), and mock Google OIDC verifiers for auth endpoint testing. Every test registers table truncation in `t.Cleanup` to ensure isolation.

**Tech Stack:** Go 1.27+, `chi/v5`, `pgx/v5` (`pgxpool`), `testcontainers-go` (v0.44+), PostgreSQL 17 container, `golang-jwt/jwt/v5`, `google/uuid`, `net/http/httptest`.

**Spec:** [docs/go_refactor/tasks/040-integration_tests_http_handlers.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/040-integration_tests_http_handlers.md)

## Global Constraints

- Target branch: `refactor/040-integration-tests-http-handlers`
- Package declaration for all tests: `package integration_test`
- Files to create:
  - `backend/tests/integration/endpoint_archer_test.go`
  - `backend/tests/integration/endpoint_auth_test.go`
  - `backend/tests/integration/endpoint_session_test.go`
  - `backend/tests/integration/endpoint_slot_test.go`
  - `backend/tests/integration/endpoint_shot_test.go`
  - `backend/tests/integration/endpoint_faces_test.go`
  - `backend/tests/integration/endpoint_security_test.go`
- File to modify: `backend/tests/integration/helpers_test.go`
- Every test function must register table cleanup via `t.Cleanup(func() { _ = truncateAll(ctx, testPool) })`
- Test server must be cleaned up via `t.Cleanup(ts.Close)`
- All tests must pass with `go test ./tests/integration/... -v -count=1` and `go test -race ./tests/integration/... -v -count=1`
- `go vet ./...` and `./scripts/linting.bash --go` must report zero issues
- At the end of implementation, mark all checklist items in [docs/go_refactor/tasks/040-integration_tests_http_handlers.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/040-integration_tests_http_handlers.md) as completed (`[x]`) and update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) marking all tasks as `DONE`

---

### Task 1: Git Branch Setup, Task Tracker Initialization & HTTP Test Helpers (`helpers_test.go`)

**Files:**
- Modify: `docs/plans/task.md`
- Modify: `backend/tests/integration/helpers_test.go`

**Interfaces:**
- Consumes: `testPool *pgxpool.Pool`, `model.*`, `repository.*`, `service.*`, `handler.*`, `middleware.*`, `auth.*`
- Produces:
  - `buildTestRouter(pool *pgxpool.Pool, customVerifier auth.GooglePayloadVerifier) (chi.Router, *auth.Service)`
  - `newTestServer(t *testing.T, customVerifier ...auth.GooglePayloadVerifier) (*httptest.Server, *auth.Service)`
  - `createAuthenticatedArcher(ctx context.Context, pool *pgxpool.Pool, secret string, overrides ...ArcherOverride) (*model.ArcherRead, string, error)`
  - `authRequest(method, url, token string, body io.Reader) *http.Request`
  - `doJSONRequest(t *testing.T, client *http.Client, req *http.Request, target any) (*http.Response, []byte)`
  - `testSlotServiceAdapter` struct implementing `handler.SlotService`
  - `testShotServiceAdapter` struct implementing `handler.ShotService`

- [ ] **Step 1: Check out feature branch**

```bash
git checkout -b refactor/040-integration-tests-http-handlers
```

- [ ] **Step 2: Initialize live task tracker in `docs/plans/task.md`**

Write table tracker in `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & HTTP Test Helpers | IN_PROGRESS | Branch checkout and HTTP test server / client helpers in `helpers_test.go` |
| Task 2: Archer Endpoints Integration Test | TODO | Port `test_archer_endpoints.py` to `endpoint_archer_test.go` |
| Task 3: Auth Endpoints Integration Test | TODO | Port `test_auth_endpoints.py` to `endpoint_auth_test.go` |
| Task 4: Session Endpoints Integration Test | TODO | Port `test_session_endpoints.py` to `endpoint_session_test.go` |
| Task 5: Slot Endpoints Integration Test | TODO | Port `test_slot_endpoints.py` to `endpoint_slot_test.go` |
| Task 6: Shot Endpoints Integration Test | TODO | Port `test_shot_endpoints.py` to `endpoint_shot_test.go` |
| Task 7: Faces Endpoints Integration Test | TODO | Port `test_faces_endpoints.py` to `endpoint_faces_test.go` |
| Task 8: Security Edge Cases Integration Test | TODO | Port `test_security_edge_cases.py` to `endpoint_security_test.go` |
| Task 9: Full Verification, Documentation Updates & Commit | TODO | Run test suite, race detector, linters, mark Task 040 done, and commit |
```

- [ ] **Step 3: Add HTTP test helpers and adapters to `backend/tests/integration/helpers_test.go`**

Append test router construction, `httptest.Server` lifecycle helpers, `authRequest`, `doJSONRequest`, `createAuthenticatedArcher`, and service adapters (`testSlotServiceAdapter`, `testShotServiceAdapter`) to `backend/tests/integration/helpers_test.go`:

```go
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
func doJSONRequest(t *testing.T, client *http.Client, req *http.Request, target any) (*http.Response, []byte) {
	t.Helper()
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
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
```

- [ ] **Step 4: Verify compilation and tests**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestSmoke`
Expected: PASS

- [ ] **Step 5: Commit Task 1 changes**

```bash
git add docs/plans/task.md backend/tests/integration/helpers_test.go
git commit -m "test: add HTTP test helpers and adapters for integration testing"
```

---

### Task 2: Archer Endpoints Integration Test (`endpoint_archer_test.go`)

**Files:**
- Create: `backend/tests/integration/endpoint_archer_test.go`

**Interfaces:**
- Consumes: `newTestServer`, `createAuthenticatedArcher`, `authRequest`, `doJSONRequest`, `testPool`, `truncateAll`
- Produces:
  - `TestEndpointArcher_CreateSuccess(t *testing.T)`
  - `TestEndpointArcher_ListArchers(t *testing.T)`
  - `TestEndpointArcher_GetArcherSuccess(t *testing.T)`
  - `TestEndpointArcher_GetArcherNotFound(t *testing.T)`
  - `TestEndpointArcher_UpdateArcherSuccess(t *testing.T)`
  - `TestEndpointArcher_DeleteArcherSuccess(t *testing.T)`
  - `TestEndpointArcher_Unauthenticated(t *testing.T)`

- [ ] **Step 1: Write `endpoint_archer_test.go` with full endpoint test suite**

Port tests from `test_archer_endpoints.py`:
- `POST /api/v0/archer/` creates archer (201 Created), returns `{"archer_id": ...}`
- `GET /api/v0/archer/` returns list of archers (200 OK)
- `GET /api/v0/archer/{id}` returns archer details (200 OK)
- `GET /api/v0/archer/{random_uuid}` returns 404 Not Found
- `PATCH /api/v0/archer/` updates archer fields (200 OK)
- `DELETE /api/v0/archer/{id}` deletes archer (204 No Content)
- Unauthenticated requests return 401 Unauthorized

```go
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

func TestEndpointArcher_CreateSuccess(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	_, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	payload := model.ArcherCreate{
		FirstName:     "Test",
		LastName:      "Archer",
		Email:         "test.archer@example.com",
		DateOfBirth:   "1990-01-01",
		Gender:        model.GenderUnspecified,
		Bowstyle:      model.BowstyleRecurve,
		DrawWeight:    40.0,
		GoogleSubject: "test_subject_123",
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/archer/", token, bytes.NewReader(body))
	var respData map[string]any
	resp, _ := doJSONRequest(t, nil, req, &respData)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if _, ok := respData["archer_id"]; !ok {
		t.Fatalf("missing archer_id in response: %+v", respData)
	}
}

func TestEndpointArcher_ListArchers(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/archer/", token, nil)
	var archers []model.ArcherRead
	resp, _ := doJSONRequest(t, nil, req, &archers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(archers) == 0 {
		t.Fatal("expected at least 1 archer")
	}
	found := false
	for _, a := range archers {
		if a.ArcherID == archer.ArcherID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created archer %v not found in list", archer.ArcherID)
	}
}

func TestEndpointArcher_GetArcherSuccess(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, archer.ArcherID), token, nil)
	var res model.ArcherRead
	resp, _ := doJSONRequest(t, nil, req, &res)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if res.ArcherID != archer.ArcherID {
		t.Errorf("ArcherID = %v, want %v", res.ArcherID, archer.ArcherID)
	}
	if res.Email != archer.Email {
		t.Errorf("Email = %v, want %v", res.Email, archer.Email)
	}
}

func TestEndpointArcher_GetArcherNotFound(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	_, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	randomID := uuid.New()
	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, randomID), token, nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestEndpointArcher_UpdateArcherSuccess(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	updatedName := "UpdatedRobin"
	updatePayload := model.ArcherUpdatePayload{
		Where: model.ArcherFilter{ArcherID: &archer.ArcherID},
		Data:  model.ArcherSet{FirstName: &updatedName},
	}
	body, _ := json.Marshal(updatePayload)

	req := authRequest(http.MethodPatch, ts.URL+"/api/v0/archer/", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("patch status = %d, want 200", resp.StatusCode)
	}

	// Verify update with GET
	getReq := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, archer.ArcherID), token, nil)
	var res model.ArcherRead
	getResp, _ := doJSONRequest(t, nil, getReq, &res)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want 200", getResp.StatusCode)
	}
	if res.FirstName != updatedName {
		t.Errorf("FirstName = %v, want %v", res.FirstName, updatedName)
	}
}

func TestEndpointArcher_DeleteArcherSuccess(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	// Create another archer to be deleted
	archerToDelete, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	_, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	delReq := authRequest(http.MethodDelete, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, archerToDelete.ArcherID), token, nil)
	delResp, _ := doJSONRequest(t, nil, delReq, nil)
	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", delResp.StatusCode)
	}

	getReq := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, archerToDelete.ArcherID), token, nil)
	getResp, _ := doJSONRequest(t, nil, getReq, nil)
	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, want 404", getResp.StatusCode)
	}
}

func TestEndpointArcher_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodGet, ts.URL + "/api/v0/archer/"},
		{http.MethodPost, ts.URL + "/api/v0/archer/"},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, uuid.New())},
		{http.MethodPatch, ts.URL + "/api/v0/archer/"},
		{http.MethodDelete, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, uuid.New())},
	}

	for _, ep := range endpoints {
		req := authRequest(ep.method, ep.url, "", nil)
		resp, _ := doJSONRequest(t, nil, req, nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", ep.method, ep.url, resp.StatusCode)
		}
	}
}
```

- [ ] **Step 2: Run archer endpoint tests**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestEndpointArcher`
Expected: PASS

- [ ] **Step 3: Commit Task 2**

```bash
git add backend/tests/integration/endpoint_archer_test.go
git commit -m "test: add archer HTTP endpoint integration tests"
```

---

### Task 3: Auth Endpoints Integration Test (`endpoint_auth_test.go`)

**Files:**
- Create: `backend/tests/integration/endpoint_auth_test.go`

**Interfaces:**
- Consumes: `newTestServer`, `createAuthenticatedArcher`, `createTestArcher`, `authRequest`, `doJSONRequest`, `testPool`, `truncateAll`
- Produces:
  - `TestEndpointAuth_Login_ValidGoogleToken_ExistingArcher(t *testing.T)`
  - `TestEndpointAuth_Login_ValidGoogleToken_NewArcherNeedsRegistration(t *testing.T)`
  - `TestEndpointAuth_Login_InvalidCredential(t *testing.T)`
  - `TestEndpointAuth_Register_Success(t *testing.T)`
  - `TestEndpointAuth_Register_InvalidCredential(t *testing.T)`
  - `TestEndpointAuth_Register_MissingFields(t *testing.T)`
  - `TestEndpointAuth_Logout_ClearsCookieAndRevokes(t *testing.T)`
  - `TestEndpointAuth_Me_Authenticated(t *testing.T)`
  - `TestEndpointAuth_Me_Unauthenticated(t *testing.T)`

- [ ] **Step 1: Write `endpoint_auth_test.go` with mock Google verifier and full auth flow**

Port tests from `test_auth_endpoints.py`:
- POST `/api/v0/auth/login` (and `/google`) with valid Google token for existing archer sets cookie and returns 200 OK
- POST `/api/v0/auth/login` for new archer returns 200 OK with `needs_registration`
- POST `/api/v0/auth/login` with invalid credential returns 401 Unauthorized
- POST `/api/v0/auth/register` with valid credential creates archer, sets cookie, and returns 201 Created
- POST `/api/v0/auth/register` with missing demographic fields returns 422 Unprocessable Entity
- POST `/api/v0/auth/register` with invalid credential returns 401 Unauthorized
- POST `/api/v0/auth/logout` clears session cookie and revokes session in DB
- GET `/api/v0/auth/me` with cookie returns authenticated archer (200 OK)
- GET `/api/v0/auth/me` without cookie returns 401 Unauthorized

```go
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"google.golang.org/api/idtoken"
)

func newMockGoogleVerifier() func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
	return func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
		switch idToken {
		case "mock-google-existing":
			return &idtoken.Payload{
				Subject: "google-sub-existing",
				Claims: map[string]any{
					"email":       "existing@example.com",
					"given_name":  "Existing",
					"family_name": "User",
				},
			}, nil
		case "mock-google-new":
			return &idtoken.Payload{
				Subject: "google-sub-new",
				Claims: map[string]any{
					"email":       "newarcher@example.com",
					"given_name":  "New",
					"family_name": "Archer",
				},
			}, nil
		default:
			return nil, errors.New("invalid google id token credential")
		}
	}
}

func TestEndpointAuth_Login_ValidGoogleToken_ExistingArcher(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	// Create existing archer with matching google subject
	_, err := createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Email = "existing@example.com"
		a.GoogleSubject = "google-sub-existing"
	})
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	payload := model.GoogleOneTapRequest{Credential: "mock-google-existing"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var authResp model.AuthAuthenticated
	resp, _ := doJSONRequest(t, nil, req, &authResp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if authResp.Status != model.AuthStatusAuthenticated {
		t.Fatalf("status = %v, want authenticated", authResp.Status)
	}
	if authResp.AccessToken == "" {
		t.Fatal("expected non-empty access_token")
	}

	// Verify cookie is set
	var authCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == middleware.AuthCookieName {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatalf("expected cookie %q in response", middleware.AuthCookieName)
	}
}

func TestEndpointAuth_Login_ValidGoogleToken_NewArcherNeedsRegistration(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	payload := model.GoogleOneTapRequest{Credential: "mock-google-new"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var needsReg model.AuthNeedsRegistration
	resp, _ := doJSONRequest(t, nil, req, &needsReg)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if needsReg.Status != model.AuthStatusNeedsRegistration {
		t.Fatalf("status = %v, want needs_registration", needsReg.Status)
	}
	if needsReg.GoogleEmail != "newarcher@example.com" {
		t.Fatalf("email = %v, want newarcher@example.com", needsReg.GoogleEmail)
	}
}

func TestEndpointAuth_Login_InvalidCredential(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	payload := model.GoogleOneTapRequest{Credential: "invalid_jwt_token"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := doJSONRequest(t, nil, req, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestEndpointAuth_Register_Success(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	first := "New"
	last := "Archer"
	regPayload := model.AuthRegistrationRequest{
		Credential:  "mock-google-new",
		FirstName:   &first,
		LastName:    &last,
		DateOfBirth: "1995-06-20",
		Gender:      model.GenderUnspecified,
		Bowstyle:    model.BowstyleRecurve,
		DrawWeight:  35.0,
	}
	body, _ := json.Marshal(regPayload)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var authResp model.AuthAuthenticated
	resp, _ := doJSONRequest(t, nil, req, &authResp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if authResp.Archer.Email != "newarcher@example.com" {
		t.Errorf("archer email = %v, want newarcher@example.com", authResp.Archer.Email)
	}
}

func TestEndpointAuth_Register_MissingFields(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	regPayload := model.AuthRegistrationRequest{
		Credential:  "mock-google-new",
		DateOfBirth: "not-a-valid-date",
		DrawWeight:  -10.0,
	}
	body, _ := json.Marshal(regPayload)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := doJSONRequest(t, nil, req, nil)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEndpointAuth_Register_InvalidCredential(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	first := "Invalid"
	last := "User"
	regPayload := model.AuthRegistrationRequest{
		Credential:  "invalid_jwt_token",
		FirstName:   &first,
		LastName:    &last,
		DateOfBirth: "1990-01-01",
		Gender:      model.GenderUnspecified,
		Bowstyle:    model.BowstyleRecurve,
		DrawWeight:  30.0,
	}
	body, _ := json.Marshal(regPayload)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := doJSONRequest(t, nil, req, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestEndpointAuth_Logout_ClearsCookieAndRevokes(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	_, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/auth/logout", token, nil)
	var logoutResp model.LogoutResponse
	resp, _ := doJSONRequest(t, nil, req, &logoutResp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if !logoutResp.Success {
		t.Fatalf("success = false, want true")
	}

	// Verify cookie is cleared (MaxAge < 0 or empty)
	var clearedCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == middleware.AuthCookieName {
			clearedCookie = c
			break
		}
	}
	if clearedCookie == nil || clearedCookie.MaxAge > 0 {
		t.Fatalf("expected cleared cookie with MaxAge <= 0, got: %+v", clearedCookie)
	}

	// Subsequent /me request with same revoked token must fail with 401
	meReq := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", token, nil)
	meResp, _ := doJSONRequest(t, nil, meReq, nil)
	if meResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("me request after logout status = %d, want 401", meResp.StatusCode)
	}
}

func TestEndpointAuth_Me_Authenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", token, nil)
	var meResp model.AuthAuthenticated
	resp, _ := doJSONRequest(t, nil, req, &meResp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if meResp.Archer.ArcherID != archer.ArcherID {
		t.Errorf("ArcherID = %v, want %v", meResp.Archer.ArcherID, archer.ArcherID)
	}
}

func TestEndpointAuth_Me_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", "", nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run auth endpoint tests**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestEndpointAuth`
Expected: PASS

- [ ] **Step 3: Commit Task 3**

```bash
git add backend/tests/integration/endpoint_auth_test.go
git commit -m "test: add authentication HTTP endpoint integration tests"
```

---

### Task 4: Session Endpoints Integration Test (`endpoint_session_test.go`)

**Files:**
- Create: `backend/tests/integration/endpoint_session_test.go`

**Interfaces:**
- Consumes: `newTestServer`, `createAuthenticatedArcher`, `createTestSession`, `authRequest`, `doJSONRequest`, `testPool`, `truncateAll`
- Produces:
  - `TestEndpointSession_CreateSession(t *testing.T)`
  - `TestEndpointSession_CreateSession_AlreadyOpen_Conflict(t *testing.T)`
  - `TestEndpointSession_GetByID(t *testing.T)`
  - `TestEndpointSession_GetOpenForArcher(t *testing.T)`
  - `TestEndpointSession_ListAllOpen(t *testing.T)`
  - `TestEndpointSession_CloseSession(t *testing.T)`
  - `TestEndpointSession_ReOpenSession(t *testing.T)`
  - `TestEndpointSession_ReOpen_BlockedIfAlreadyOpen(t *testing.T)`
  - `TestEndpointSession_GetParticipating(t *testing.T)`
  - `TestEndpointSession_Unauthenticated(t *testing.T)`

- [ ] **Step 1: Write `endpoint_session_test.go`**

Port tests from `test_session_endpoints.py`:
- POST `/api/v0/session` creates session (201 Created)
- POST `/api/v0/session` when archer already has an open session returns 409 Conflict
- GET `/api/v0/session/{id}` returns session details
- GET `/api/v0/session/archer/{archer_id}/open-session` returns open session ID
- GET `/api/v0/session/open` returns all open sessions
- PATCH `/api/v0/session/close` closes the session
- PATCH `/api/v0/session/re-open` re-opens a closed session
- PATCH `/api/v0/session/re-open` when another session is open returns 409 Conflict
- GET `/api/v0/session/archer/{archer_id}/participating` returns session ID
- Unauthenticated requests return 401 Unauthorized

```go
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

func TestEndpointSession_CreateSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	payload := model.SessionCreate{
		OwnerArcherID:   archer.ArcherID,
		SessionLocation: "Main Range",
		IsIndoor:        false,
		IsOpened:        true,
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/session", token, bytes.NewReader(body))
	var respData model.SessionID
	resp, _ := doJSONRequest(t, nil, req, &respData)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if respData.SessionID == nil || *respData.SessionID == uuid.Nil {
		t.Fatalf("missing session_id in response: %+v", respData)
	}
}

func TestEndpointSession_CreateSession_AlreadyOpen_Conflict(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	// 1. Create first open session
	_, err = createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	// 2. Attempt to create second open session
	payload := model.SessionCreate{
		OwnerArcherID:   archer.ArcherID,
		SessionLocation: "Second Range",
		IsIndoor:        false,
		IsOpened:        true,
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/session", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409 Conflict", resp.StatusCode)
	}
}

func TestEndpointSession_GetByID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/%s", ts.URL, sess.SessionID), token, nil)
	var readSess model.SessionRead
	resp, _ := doJSONRequest(t, nil, req, &readSess)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if readSess.SessionID != sess.SessionID {
		t.Errorf("SessionID = %v, want %v", readSess.SessionID, sess.SessionID)
	}
}

func TestEndpointSession_GetOpenForArcher(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/archer/%s/open-session", ts.URL, archer.ArcherID), token, nil)
	var respData model.SessionID
	resp, _ := doJSONRequest(t, nil, req, &respData)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if respData.SessionID == nil || *respData.SessionID != sess.SessionID {
		t.Errorf("session_id = %v, want %v", respData.SessionID, sess.SessionID)
	}
}

func TestEndpointSession_ListAllOpen(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/session/open", token, nil)
	var sessions []model.SessionRead
	resp, _ := doJSONRequest(t, nil, req, &sessions)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	found := false
	for _, s := range sessions {
		if s.SessionID == sess.SessionID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("open session %v not found in list", sess.SessionID)
	}
}

func TestEndpointSession_CloseSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	closePayload := model.SessionID{SessionID: &sess.SessionID}
	body, _ := json.Marshal(closePayload)

	req := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/close", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 200 or 204", resp.StatusCode)
	}

	// Verify session is no longer open
	getReq := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/%s", ts.URL, sess.SessionID), token, nil)
	var readSess model.SessionRead
	_, _ = doJSONRequest(t, nil, getReq, &readSess)
	if readSess.IsOpened {
		t.Fatal("expected session to be closed, but is_opened is true")
	}
}

func TestEndpointSession_ReOpenSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	// Close first
	closePayload := model.SessionID{SessionID: &sess.SessionID}
	body, _ := json.Marshal(closePayload)
	closeReq := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/close", token, bytes.NewReader(body))
	_, _ = doJSONRequest(t, nil, closeReq, nil)

	// Now re-open
	reopenReq := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/re-open", token, bytes.NewReader(body))
	reopenResp, _ := doJSONRequest(t, nil, reopenReq, nil)
	if reopenResp.StatusCode != http.StatusOK {
		t.Fatalf("re-open status = %d, want 200", reopenResp.StatusCode)
	}
}

func TestEndpointSession_ReOpen_BlockedIfAlreadyOpen(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	// 1. Create Session A and close it
	sessA, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession A failed: %v", err)
	}
	bodyA, _ := json.Marshal(model.SessionID{SessionID: &sessA.SessionID})
	closeReq := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/close", token, bytes.NewReader(bodyA))
	_, _ = doJSONRequest(t, nil, closeReq, nil)

	// 2. Create Session B (now open)
	_, err = createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession B failed: %v", err)
	}

	// 3. Attempt to re-open Session A -> conflict
	reopenReq := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/re-open", token, bytes.NewReader(bodyA))
	resp, _ := doJSONRequest(t, nil, reopenReq, nil)
	if resp.StatusCode != http.StatusConflict && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 409 or 422", resp.StatusCode)
	}
}

func TestEndpointSession_GetParticipating(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	_, err = createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/archer/%s/participating", ts.URL, archer.ArcherID), token, nil)
	var respData map[string]*uuid.UUID
	resp, _ := doJSONRequest(t, nil, req, &respData)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if respData["session_id"] == nil || *respData["session_id"] != sess.SessionID {
		t.Errorf("session_id = %v, want %v", respData["session_id"], sess.SessionID)
	}
}

func TestEndpointSession_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodGet, ts.URL + "/api/v0/session/open"},
		{http.MethodPost, ts.URL + "/api/v0/session"},
		{http.MethodPatch, ts.URL + "/api/v0/session/close"},
		{http.MethodPatch, ts.URL + "/api/v0/session/re-open"},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/%s", ts.URL, uuid.New())},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/archer/%s/open-session", ts.URL, uuid.New())},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/archer/%s/participating", ts.URL, uuid.New())},
	}

	for _, ep := range endpoints {
		req := authRequest(ep.method, ep.url, "", nil)
		resp, _ := doJSONRequest(t, nil, req, nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", ep.method, ep.url, resp.StatusCode)
		}
	}
}
```

- [ ] **Step 2: Run session endpoint tests**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestEndpointSession`
Expected: PASS

- [ ] **Step 3: Commit Task 4**

```bash
git add backend/tests/integration/endpoint_session_test.go
git commit -m "test: add session HTTP endpoint integration tests"
```

---

### Task 5: Slot Endpoints Integration Test (`endpoint_slot_test.go`)

**Files:**
- Create: `backend/tests/integration/endpoint_slot_test.go`

**Interfaces:**
- Consumes: `newTestServer`, `createAuthenticatedArcher`, `createTestSession`, `createTestTarget`, `createTestSlot`, `authRequest`, `doJSONRequest`, `testPool`, `truncateAll`
- Produces:
  - `TestEndpointSlot_JoinSession_AssignsSlot(t *testing.T)`
  - `TestEndpointSlot_GetSlot(t *testing.T)`
  - `TestEndpointSlot_GetArcherCurrentSlot(t *testing.T)`
  - `TestEndpointSlot_LeaveSession(t *testing.T)`
  - `TestEndpointSlot_ReJoinSession(t *testing.T)`
  - `TestEndpointSlot_JoinClosedSession_Unprocessable(t *testing.T)`
  - `TestEndpointSlot_Unauthenticated(t *testing.T)`

- [ ] **Step 1: Write `endpoint_slot_test.go`**

Port tests from `test_slot_endpoints.py`:
- POST `/api/v0/session/slot` assigns target and slot, marks participating
- GET `/api/v0/session/slot/{slot_id}` returns full slot info
- GET `/api/v0/session/slot/archer/{archer_id}` returns archer's current active slot
- PATCH `/api/v0/session/slot/leave/{slot_id}` marks `is_shooting = false`
- PATCH `/api/v0/session/slot/re-join/{slot_id}` marks `is_shooting = true`
- POST `/api/v0/session/slot` in closed session returns 422 Unprocessable Entity
- Unauthenticated requests return 401 Unauthorized

```go
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestEndpointSlot_JoinSession_AssignsSlot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	owner, ownerToken, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher owner failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, owner.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	shotPerRound := 3
	joinPayload := model.SlotJoinRequest{
		SessionID:       sess.SessionID,
		ArcherID:        owner.ArcherID,
		Distance:        18,
		FaceType:        model.FaceTypeWA40Full,
		IsShooting:      true,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      40.0,
		ShotPerRound:    &shotPerRound,
		IntervalSeconds: 20,
	}
	body, _ := json.Marshal(joinPayload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/session/slot", ownerToken, bytes.NewReader(body))
	var joinResp model.SlotJoinResponse
	resp, _ := doJSONRequest(t, nil, req, &joinResp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("join status = %d, want 200 or 201", resp.StatusCode)
	}
	if joinResp.SlotID == uuid.Nil {
		t.Fatal("expected valid slot_id")
	}
	if joinResp.Slot == "" {
		t.Fatal("expected non-empty slot string (e.g. 1A)")
	}
}

func TestEndpointSlot_GetSlot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/slot/%s", ts.URL, slot.SlotID), token, nil)
	var fullInfo model.FullSlotInfo
	resp, _ := doJSONRequest(t, nil, req, &fullInfo)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get slot status = %d, want 200", resp.StatusCode)
	}
	if fullInfo.SlotID != slot.SlotID {
		t.Errorf("SlotID = %v, want %v", fullInfo.SlotID, slot.SlotID)
	}
	if fullInfo.ArcherID != archer.ArcherID {
		t.Errorf("ArcherID = %v, want %v", fullInfo.ArcherID, archer.ArcherID)
	}
}

func TestEndpointSlot_GetArcherCurrentSlot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/slot/archer/%s", ts.URL, archer.ArcherID), token, nil)
	var fullInfo model.FullSlotInfo
	resp, _ := doJSONRequest(t, nil, req, &fullInfo)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get archer slot status = %d, want 200", resp.StatusCode)
	}
	if fullInfo.SlotID != slot.SlotID {
		t.Errorf("SlotID = %v, want %v", fullInfo.SlotID, slot.SlotID)
	}
}

func TestEndpointSlot_LeaveSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	req := authRequest(http.MethodPatch, fmt.Sprintf("%s/api/v0/session/slot/leave/%s", ts.URL, slot.SlotID), token, nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave status = %d, want 200 or 204", resp.StatusCode)
	}

	// Verify slot is_shooting is false
	slotRepo := repository.NewSlotRepo(testPool)
	updatedSlot, err := slotRepo.FindByID(ctx, slot.SlotID)
	if err != nil || updatedSlot == nil {
		t.Fatalf("finding slot after leave: %v", err)
	}
	if updatedSlot.IsShooting {
		t.Fatal("expected is_shooting to be false after leave")
	}
}

func TestEndpointSlot_ReJoinSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID, func(s *model.SlotCreate) {
		s.IsShooting = false
	})
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	req := authRequest(http.MethodPatch, fmt.Sprintf("%s/api/v0/session/slot/re-join/%s", ts.URL, slot.SlotID), token, nil)
	var rejoinResp model.SlotJoinResponse
	resp, _ := doJSONRequest(t, nil, req, &rejoinResp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("re-join status = %d, want 200", resp.StatusCode)
	}
	if rejoinResp.SlotID != slot.SlotID {
		t.Errorf("SlotID = %v, want %v", rejoinResp.SlotID, slot.SlotID)
	}
}

func TestEndpointSlot_JoinClosedSession_Unprocessable(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	// Create and close session
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	sessionRepo := repository.NewSessionRepo(testPool)
	_ = sessionRepo.Close(ctx, sess.SessionID)

	shotPerRound := 3
	joinPayload := model.SlotJoinRequest{
		SessionID:       sess.SessionID,
		ArcherID:        archer.ArcherID,
		Distance:        18,
		FaceType:        model.FaceTypeWA40Full,
		IsShooting:      true,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      40.0,
		ShotPerRound:    &shotPerRound,
		IntervalSeconds: 20,
	}
	body, _ := json.Marshal(joinPayload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/session/slot", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 Unprocessable Entity", resp.StatusCode)
	}
}

func TestEndpointSlot_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodPost, ts.URL + "/api/v0/session/slot"},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/slot/%s", ts.URL, uuid.New())},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/slot/archer/%s", ts.URL, uuid.New())},
		{http.MethodPatch, fmt.Sprintf("%s/api/v0/session/slot/leave/%s", ts.URL, uuid.New())},
		{http.MethodPatch, fmt.Sprintf("%s/api/v0/session/slot/re-join/%s", ts.URL, uuid.New())},
	}

	for _, ep := range endpoints {
		req := authRequest(ep.method, ep.url, "", nil)
		resp, _ := doJSONRequest(t, nil, req, nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", ep.method, ep.url, resp.StatusCode)
		}
	}
}
```

- [ ] **Step 2: Run slot endpoint tests**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestEndpointSlot`
Expected: PASS

- [ ] **Step 3: Commit Task 5**

```bash
git add backend/tests/integration/endpoint_slot_test.go
git commit -m "test: add slot HTTP endpoint integration tests"
```

---

### Task 6: Shot Endpoints Integration Test (`endpoint_shot_test.go`)

**Files:**
- Create: `backend/tests/integration/endpoint_shot_test.go`

**Interfaces:**
- Consumes: `newTestServer`, `createAuthenticatedArcher`, `createTestSession`, `createTestTarget`, `createTestSlot`, `authRequest`, `doJSONRequest`, `testPool`, `truncateAll`
- Produces:
  - `TestEndpointShot_CreateSingleShot(t *testing.T)`
  - `TestEndpointShot_CreateBatchShots(t *testing.T)`
  - `TestEndpointShot_GetBySlot(t *testing.T)`
  - `TestEndpointShot_CountBySlot(t *testing.T)`
  - `TestEndpointShot_InvalidScore(t *testing.T)`
  - `TestEndpointShot_CreateShotInClosedSession(t *testing.T)`
  - `TestEndpointShot_Unauthenticated(t *testing.T)`

- [ ] **Step 1: Write `endpoint_shot_test.go`**

Port tests from `test_shot_endpoints.py`:
- POST `/api/v0/shot` creates single shot (201 Created), returns `{"shot_id": ...}`
- POST `/api/v0/shot` creates batch of shots (201 Created)
- GET `/api/v0/shot/by-slot/{slot_id}` returns list of shots for slot
- GET `/api/v0/shot/count-by-slot/{slot_id}` returns count
- POST `/api/v0/shot` with invalid score (<0 or >10) returns 422 Unprocessable Entity
- POST `/api/v0/shot` in closed session returns 422 Unprocessable Entity
- Unauthenticated requests return 401 Unauthorized

```go
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestEndpointShot_CreateSingleShot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	x := 10.5
	y := 20.5
	score := 9
	payload := model.ShotCreate{
		SlotID: slot.SlotID,
		X:      &x,
		Y:      &y,
		Score:  &score,
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/shot", token, bytes.NewReader(body))
	var shotResp model.ShotID
	resp, _ := doJSONRequest(t, nil, req, &shotResp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if shotResp.ShotID == uuid.Nil {
		t.Fatal("expected valid shot_id")
	}
}

func TestEndpointShot_CreateBatchShots(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	x := 5.0
	y := 5.0
	score := 10
	shots := []model.ShotCreate{
		{SlotID: slot.SlotID, X: &x, Y: &y, Score: &score},
		{SlotID: slot.SlotID, X: &x, Y: &y, Score: &score},
		{SlotID: slot.SlotID, X: &x, Y: &y, Score: &score},
	}
	body, _ := json.Marshal(shots)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/shot", token, bytes.NewReader(body))
	var batchResp model.ShotBatchResponse
	resp, _ := doJSONRequest(t, nil, req, &batchResp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if len(batchResp.ShotIDs) != 3 {
		t.Fatalf("got %d shot_ids, want 3", len(batchResp.ShotIDs))
	}
}

func TestEndpointShot_GetBySlot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	shot1, _ := createTestShot(ctx, testPool, slot.SlotID)
	shot2, _ := createTestShot(ctx, testPool, slot.SlotID)

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/shot/by-slot/%s", ts.URL, slot.SlotID), token, nil)
	var shots []model.ShotRead
	resp, _ := doJSONRequest(t, nil, req, &shots)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(shots) != 2 {
		t.Fatalf("got %d shots, want 2", len(shots))
	}
	gotIDs := map[uuid.UUID]bool{shots[0].ShotID: true, shots[1].ShotID: true}
	if !gotIDs[shot1.ShotID] || !gotIDs[shot2.ShotID] {
		t.Errorf("missing created shots in response: %+v", shots)
	}
}

func TestEndpointShot_CountBySlot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	_, _ = createTestShot(ctx, testPool, slot.SlotID)
	_, _ = createTestShot(ctx, testPool, slot.SlotID)
	_, _ = createTestShot(ctx, testPool, slot.SlotID)

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/shot/count-by-slot/%s", ts.URL, slot.SlotID), token, nil)
	var countResp model.ShotCountResponse
	resp, _ := doJSONRequest(t, nil, req, &countResp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if countResp.Count != 3 {
		t.Fatalf("count = %d, want 3", countResp.Count)
	}
}

func TestEndpointShot_InvalidScore(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	invalidScores := []int{-1, 11, 99}
	for _, sc := range invalidScores {
		score := sc
		payload := model.ShotCreate{
			SlotID: slot.SlotID,
			Score:  &score,
		}
		body, _ := json.Marshal(payload)

		req := authRequest(http.MethodPost, ts.URL+"/api/v0/shot", token, bytes.NewReader(body))
		resp, _ := doJSONRequest(t, nil, req, nil)

		if resp.StatusCode != http.StatusUnprocessableEntity && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("score %d status = %d, want 422 or 400", sc, resp.StatusCode)
		}
	}
}

func TestEndpointShot_CreateShotInClosedSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	// Close the session
	sessionRepo := repository.NewSessionRepo(testPool)
	_ = sessionRepo.Close(ctx, sess.SessionID)

	score := 10
	payload := model.ShotCreate{
		SlotID: slot.SlotID,
		Score:  &score,
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/shot", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 Unprocessable Entity", resp.StatusCode)
	}
}

func TestEndpointShot_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodPost, ts.URL + "/api/v0/shot"},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/shot/by-slot/%s", ts.URL, uuid.New())},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/shot/count-by-slot/%s", ts.URL, uuid.New())},
	}

	for _, ep := range endpoints {
		req := authRequest(ep.method, ep.url, "", nil)
		resp, _ := doJSONRequest(t, nil, req, nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", ep.method, ep.url, resp.StatusCode)
		}
	}
}
```

- [ ] **Step 2: Run shot endpoint tests**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestEndpointShot`
Expected: PASS

- [ ] **Step 3: Commit Task 6**

```bash
git add backend/tests/integration/endpoint_shot_test.go
git commit -m "test: add shot HTTP endpoint integration tests"
```

---

### Task 7: Target Faces Endpoints Integration Test (`endpoint_faces_test.go`)

**Files:**
- Create: `backend/tests/integration/endpoint_faces_test.go`

**Interfaces:**
- Consumes: `newTestServer`, `doJSONRequest`
- Produces:
  - `TestEndpointFaces_ListFaces(t *testing.T)`
  - `TestEndpointFaces_GetFace_Success(t *testing.T)`
  - `TestEndpointFaces_GetFace_NotFound(t *testing.T)`

- [ ] **Step 1: Write `endpoint_faces_test.go`**

Port tests from `test_faces_endpoints.py`:
- GET `/api/v0/faces` returns all available face summaries with `face_type` and `face_name` (public endpoint, 200 OK)
- GET `/api/v0/faces/{face_type}` returns geometry, rings, and scoring layout (200 OK)
- GET `/api/v0/faces/non_existent_face` returns 404 Not Found

```go
package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

func TestEndpointFaces_ListFaces(t *testing.T) {
	ts, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v0/faces", nil)
	var summaries []model.FaceMinimal
	resp, _ := doJSONRequest(t, nil, req, &summaries)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(summaries) == 0 {
		t.Fatal("expected at least 1 face definition")
	}
	for _, f := range summaries {
		if f.FaceType == "" || f.FaceName == "" {
			t.Errorf("incomplete face summary: %+v", f)
		}
	}
}

func TestEndpointFaces_GetFace_Success(t *testing.T) {
	ts, _ := newTestServer(t)

	// Fetch catalog first to get a real face_type
	listReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v0/faces", nil)
	var summaries []model.FaceMinimal
	_, _ = doJSONRequest(t, nil, listReq, &summaries)

	if len(summaries) == 0 {
		t.Fatal("no faces in catalog")
	}
	firstType := summaries[0].FaceType

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/faces/%s", ts.URL, firstType), nil)
	var face model.FaceRead
	resp, _ := doJSONRequest(t, nil, req, &face)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if face.FaceType != firstType {
		t.Errorf("FaceType = %v, want %v", face.FaceType, firstType)
	}
	if len(face.Rings) == 0 {
		t.Errorf("expected rings in face %v", firstType)
	}
}

func TestEndpointFaces_GetFace_NotFound(t *testing.T) {
	ts, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v0/faces/non_existent_face_type_xyz", nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 404 or 422", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run faces endpoint tests**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestEndpointFaces`
Expected: PASS

- [ ] **Step 3: Commit Task 7**

```bash
git add backend/tests/integration/endpoint_faces_test.go
git commit -m "test: add target faces HTTP endpoint integration tests"
```

---

### Task 8: Security Edge Cases Integration Test (`endpoint_security_test.go`)

**Files:**
- Create: `backend/tests/integration/endpoint_security_test.go`

**Interfaces:**
- Consumes: `newTestServer`, `createAuthenticatedArcher`, `createTestSession`, `createTestTarget`, `createTestSlot`, `authRequest`, `testPool`, `truncateAll`
- Produces:
  - `TestEndpointSecurity_CloseSessionForbiddenForNonOwner(t *testing.T)`
  - `TestEndpointSecurity_ExpiredJWT(t *testing.T)`
  - `TestEndpointSecurity_MalformedJWT(t *testing.T)`
  - `TestEndpointSecurity_MissingAuthCookie(t *testing.T)`
  - `TestEndpointSecurity_RevokedSession(t *testing.T)`

- [ ] **Step 1: Write `endpoint_security_test.go`**

Port tests from `test_security_edge_cases.py`:
- Cross-user session close attempts return 403 Forbidden
- Expired JWT token returns 401 Unauthorized
- Malformed JWT token string returns 401 Unauthorized
- Missing auth cookie returns 401 Unauthorized
- Revoked session in database returns 401 Unauthorized

```go
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestEndpointSecurity_CloseSessionForbiddenForNonOwner(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	owner, _, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("create owner failed: %v", err)
	}

	_, strangerToken, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("create stranger failed: %v", err)
	}

	// Owner creates a session
	sess, err := createTestSession(ctx, testPool, owner.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	// Stranger attempts to close owner's session
	payload := model.SessionID{SessionID: &sess.SessionID}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/close", strangerToken, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 Forbidden", resp.StatusCode)
	}
}

func TestEndpointSecurity_ExpiredJWT(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	// Generate expired JWT
	expiredToken, err := auth.BuildJWT(
		archer.ArcherID,
		"expired-sid",
		time.Now().UTC().Add(-2*time.Hour),
		time.Now().UTC().Add(-1*time.Hour),
		testJWTSecret,
		"HS256",
	)
	if err != nil {
		t.Fatalf("BuildJWT failed: %v", err)
	}

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", expiredToken, nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 Unauthorized for expired JWT", resp.StatusCode)
	}
}

func TestEndpointSecurity_MalformedJWT(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	malformedTokens := []string{
		"not-a-jwt",
		"header.payload.signature.extra",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.badpayload.badsig",
	}

	for _, tok := range malformedTokens {
		req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", tok, nil)
		resp, _ := doJSONRequest(t, nil, req, nil)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("token %q status = %d, want 401 Unauthorized", tok, resp.StatusCode)
		}
	}
}

func TestEndpointSecurity_MissingAuthCookie(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 Unauthorized for missing auth cookie", resp.StatusCode)
	}
}

func TestEndpointSecurity_RevokedSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	// Revoke session in database
	authSessionRepo := repository.NewAuthSessionRepo(testPool)
	if err := authSessionRepo.DeleteByArcherID(ctx, archer.ArcherID); err != nil {
		t.Fatalf("DeleteByArcherID failed: %v", err)
	}

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", token, nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 Unauthorized for revoked DB session", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run security edge cases tests**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestEndpointSecurity`
Expected: PASS

- [ ] **Step 3: Commit Task 8**

```bash
git add backend/tests/integration/endpoint_security_test.go
git commit -m "test: add security edge cases HTTP integration tests"
```

---

### Task 9: Verification, Linting, Documentation Completion & Commit

**Files:**
- Modify: `docs/go_refactor/tasks/040-integration_tests_http_handlers.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: full integration test suite, `golangci-lint`, `go vet`
- Produces: All acceptance criteria and steps marked completed (`[x]`), task tracker marked `DONE`, clean commit

- [ ] **Step 1: Run complete integration test suite**

Run:
```bash
cd backend && go test ./tests/integration/... -v -count=1
```
Expected: All tests pass cleanly.

- [ ] **Step 2: Run integration tests with race detector**

Run:
```bash
cd backend && go test -race ./tests/integration/... -v -count=1
```
Expected: All integration tests pass cleanly with 0 data races.

- [ ] **Step 3: Run `go vet` and full Go linter**

Run:
```bash
cd backend && go vet ./...
./scripts/linting.bash --go
```
Expected: 0 issues reported.

- [ ] **Step 4: Mark Task 040 as completed in `docs/go_refactor/tasks/040-integration_tests_http_handlers.md`**

Update `docs/go_refactor/tasks/040-integration_tests_http_handlers.md` marking all acceptance criteria checkboxes and steps as `[x]`:

```markdown
## Acceptance Criteria

- [x] `backend/tests/integration/endpoint_archer_test.go` ports tests from
  `test_archer_endpoints.py`:
    - GET `/api/v0/archers` — list archers (authenticated)
    - GET `/api/v0/archers/:id` — get single archer
    - Unauthenticated requests → 401
- [x] `backend/tests/integration/endpoint_auth_test.go` ports tests from
  `test_auth_endpoints.py`:
    - POST `/api/v0/auth/login` — with valid/invalid credentials
    - POST `/api/v0/auth/register` — with valid/missing fields
    - POST `/api/v0/auth/logout` — clears session
    - GET `/api/v0/auth/me` — returns authenticated archer
- [x] `backend/tests/integration/endpoint_session_test.go` ports tests from
  `test_session_endpoints.py`:
    - POST `/api/v0/sessions` — create session
    - GET `/api/v0/sessions` — list sessions
    - GET `/api/v0/sessions/open` — get open session
    - PUT `/api/v0/sessions/:id/close` — close session
    - Cannot create session when one is already open → 409
- [x] `backend/tests/integration/endpoint_slot_test.go` ports tests from
  `test_slot_endpoints.py`:
    - POST `/api/v0/slots` — create slot in open session
    - GET `/api/v0/slots?session_id=X` — list slots by session
    - PUT `/api/v0/slots/:id` — update slot
    - DELETE `/api/v0/slots/:id` — delete slot
    - Create slot in closed session → 422
- [x] `backend/tests/integration/endpoint_shot_test.go` ports tests from
  `test_shot_endpoints.py`:
    - POST `/api/v0/shots` — create shot in slot
    - GET `/api/v0/shots?slot_id=X` — list shots by slot
    - PUT `/api/v0/shots/:id` — update shot
    - DELETE `/api/v0/shots/:id` — delete shot
- [x] `backend/tests/integration/endpoint_faces_test.go` ports tests from
  `test_faces_endpoints.py`:
    - GET `/api/v0/faces` — list available faces
- [x] `backend/tests/integration/endpoint_security_test.go` ports tests from
  `test_security_edge_cases.py`:
    - Cross-user access attempts → 403
    - Expired JWT → 401
    - Malformed JWT → 401
    - Missing auth cookie → 401
- [x] All tests verify JSON response bodies match the expected API contract (field names,
  types, structure).
- [x] Each test truncates tables after completion.
- [x] `go test ./tests/integration/... -v -count=1` passes.
- [x] `go vet ./...` reports no issues.

...

## Steps

- [x] **Step 1: Add HTTP test helpers to `helpers_test.go`**
- [x] **Step 2: Write `endpoint_auth_test.go`**
- [x] **Step 3: Write `endpoint_archer_test.go`**
- [x] **Step 4: Write `endpoint_session_test.go`**
- [x] **Step 5: Write `endpoint_slot_test.go`**
- [x] **Step 6: Write `endpoint_shot_test.go`**
- [x] **Step 7: Write `endpoint_faces_test.go`**
- [x] **Step 8: Write `endpoint_security_test.go`**
- [x] **Step 9: Run all integration tests**
- [x] **Step 10: Run go vet**
- [x] **Step 11: Commit**
```

- [ ] **Step 5: Mark all tasks as `DONE` in `docs/plans/task.md`**

Update `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & HTTP Test Helpers | DONE | Branch checkout and HTTP test server / client helpers in `helpers_test.go` |
| Task 2: Archer Endpoints Integration Test | DONE | Port `test_archer_endpoints.py` to `endpoint_archer_test.go` |
| Task 3: Auth Endpoints Integration Test | DONE | Port `test_auth_endpoints.py` to `endpoint_auth_test.go` |
| Task 4: Session Endpoints Integration Test | DONE | Port `test_session_endpoints.py` to `endpoint_session_test.go` |
| Task 5: Slot Endpoints Integration Test | DONE | Port `test_slot_endpoints.py` to `endpoint_slot_test.go` |
| Task 6: Shot Endpoints Integration Test | DONE | Port `test_shot_endpoints.py` to `endpoint_shot_test.go` |
| Task 7: Faces Endpoints Integration Test | DONE | Port `test_faces_endpoints.py` to `endpoint_faces_test.go` |
| Task 8: Security Edge Cases Integration Test | DONE | Port `test_security_edge_cases.py` to `endpoint_security_test.go` |
| Task 9: Full Verification, Documentation Updates & Commit | DONE | Run test suite, race detector, linters, mark Task 040 done, and commit |
```

- [ ] **Step 6: Commit all remaining changes to git**

Run:
```bash
git add backend/tests/integration/ docs/go_refactor/tasks/040-integration_tests_http_handlers.md docs/plans/task.md docs/plans/2026-09-06-integration-tests-http-handlers.md
git commit -m "test: port Python endpoint tests to Go HTTP integration tests"
```
Expected: Clean commit on branch `refactor/040-integration-tests-http-handlers`.
