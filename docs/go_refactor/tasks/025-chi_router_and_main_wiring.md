# Task 025: Wire chi Router with All Handlers + DI in `main.go`

## Git Branch

`refactor/025-chi-router-and-main-wiring`

## Objective

Wire the chi HTTP router with all handler implementations, apply the middleware stack, configure
route groups under `/api/v0/`, and complete the dependency injection wiring in `main.go`. After
this task, the Go binary is a fully runnable HTTP server serving all API endpoints.

## Dependencies

- Task 003 (config)
- Task 004 (database pool)
- Task 005 (migrations)
- Task 006 (logging)
- Task 008–013 (all repositories)
- Task 014–016 (all services)
- Task 017 (auth service)
- Task 018 (middleware stack)
- Task 019–024 (all handlers)

## Acceptance Criteria

- [ ] `backend/cmd/arch-stats/main.go` performs full dependency wiring:
  1. Load config
  2. Create logger
  3. Connect database pool
  4. Run migrations (if configured)
  5. Log current schema version via `MaintenanceRepo.GetSchemaVersion()`
  6. Create all repositories (passing pool), including `MaintenanceRepo` and `ReportingRepo`
  7. Create all services (passing repositories)
  8. Create auth service (passing repos + config)
  9. Create all handlers (passing services)
  10. Build chi router with route groups
  11. Apply middleware stack
  12. Start HTTP server with graceful shutdown
- [ ] Route groups match the Python API structure:
    - `/api/v0/auth/` — auth handler (public: login, register; protected: logout, me)
    - `/api/v0/archer/` — archer handler (protected)
    - `/api/v0/session/` — session handler (protected)
    - `/api/v0/session/slot/` — slot handler (protected)
    - `/api/v0/shot/` — shot handler (protected)
    - `/api/v0/faces/` — face handler (public)
    - `/api/v0/stats/` — live stats handler (protected)
- [ ] Middleware stack applied in correct order: logging → recovery → CORS → (per-group auth)
  → error mapper.
- [ ] A `GET /api/v0/health` endpoint returns JSON with at minimum the current database schema
  version (from `MaintenanceRepo.GetSchemaVersion()`). This is a public, unauthenticated endpoint.
- [ ] `go build ./cmd/arch-stats` compiles cleanly.
- [ ] `go vet ./...` reports no issues.
- [ ] Running the binary with valid DB config starts the HTTP server and logs "listening on :PORT".

## Files to Modify

| Action | Path |
| ------ | ---- |
| Modify | `backend/cmd/arch-stats/main.go` |
| Modify | `backend/go.mod` (add chi dependency) |

## Steps

- [ ] **Step 1: Add chi dependency**

  ```bash
  cd backend
  go get github.com/go-chi/chi/v5
  go get github.com/go-chi/cors
  ```

- [ ] **Step 2: Refactor `main.go` into a structured wiring function**

  Organize `main.go` into clear sections:

  ```go
  func main() {
      // 1. Config
      cfg, err := config.Load()
      // ...

      // 2. Logger
      logger := config.NewLogger(cfg.DevMode)
      slog.SetDefault(logger)

      // 3. Database
      ctx := context.Background()
      pool, err := repository.NewPool(ctx, dsn, cfg.PostgresPoolMinSize, cfg.PostgresPoolMaxSize)
      defer pool.Close()

      // 4. Migrations
      if cfg.ApplyMigrationsOnStart {
          repository.RunMigrations(ctx, pool, "migrations")
      }

      // 5. Repositories
      archerRepo := repository.NewArcherRepo(pool)
      authSessionRepo := repository.NewAuthSessionRepo(pool)
      sessionRepo := repository.NewSessionRepo(pool)
      slotRepo := repository.NewSlotRepo(pool)
      shotRepo := repository.NewShotRepo(pool)
      faceRepo := repository.NewFaceRepo(pool)
      targetRepo := repository.NewTargetRepo(pool)
      maintenanceRepo := repository.NewMaintenanceRepo(pool)
      reportingRepo := repository.NewReportingRepo(pool)

      // 5b. Log schema version
      if ver, err := maintenanceRepo.GetSchemaVersion(ctx); err == nil {
          slog.Info("database schema version", "version", ver)
      }

      // 6. Services
      archerSvc := service.NewArcherService(archerRepo)
      sessionSvc := service.NewSessionService(sessionRepo)
      slotSvc := service.NewSlotService(slotRepo, sessionRepo)
      shotSvc := service.NewShotService(shotRepo, slotRepo)
      faceSvc := service.NewFaceService(faceRepo)
      targetSvc := service.NewTargetService(targetRepo, faceRepo)

      // 7. Auth
      authSvc := auth.NewService(cfg, archerRepo, authSessionRepo)

      // 8. Handlers
      authHandler := handler.NewAuthHandler(authSvc, archerSvc)
      archerHandler := handler.NewArcherHandler(archerSvc)
      sessionHandler := handler.NewSessionHandler(sessionSvc)
      slotHandler := handler.NewSlotHandler(slotSvc)
      shotHandler := handler.NewShotHandler(shotSvc)
      faceHandler := handler.NewFaceHandler(faceSvc)

      // 9. Router
      r := buildRouter(cfg, authSvc, authHandler, archerHandler, ...)

      // 10. Server
      srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.ServerPort), Handler: r}
      // graceful shutdown...
  }
  ```

- [ ] **Step 3: Build the chi router with route groups**

  ```go
  func buildRouter(cfg *config.Config, authSvc *auth.Service, ...) chi.Router {
      r := chi.NewRouter()

      // Global middleware
      r.Use(middleware.RequestLogger(logger))
      r.Use(middleware.Recovery)
      r.Use(cors.Handler(middleware.CORSOptions(cfg.DevMode)))

      r.Route("/api/v0", func(r chi.Router) {
          // Public routes
          r.Route("/auth", func(r chi.Router) {
              r.Post("/login", authHandler.Login)
              r.Post("/register", authHandler.Register)
              r.Group(func(r chi.Router) {
                  r.Use(middleware.Auth(authSvc))
                  r.Post("/logout", authHandler.Logout)
                  r.Get("/me", authHandler.Me)
              })
          })

          r.Route("/faces", func(r chi.Router) {
              r.Get("/", faceHandler.ListFaces)
              r.Get("/{face_type}", faceHandler.GetFace)
          })

          // Protected routes
          r.Group(func(r chi.Router) {
              r.Use(middleware.Auth(authSvc))
              r.Route("/archer", archerHandler.Routes)
              r.Route("/session", sessionHandler.Routes)
              r.Route("/shot", shotHandler.Routes)
              r.Route("/stats", liveStatsHandler.Routes)
          })
      })

      return r
  }
  ```

- [ ] **Step 4: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./cmd/arch-stats
  ```

- [ ] **Step 5: Manual verification**

  Start the server with a running PostgreSQL instance:

  ```bash
  cd backend
  ./arch-stats
  ```

  Expected: logs "arch-stats listening on :8000" (or configured port).

  Test with curl:

  ```bash
  curl -s http://localhost:8000/api/v0/faces | jq .
  ```

  Expected: JSON array of face data (public endpoint, no auth needed).

- [ ] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "feat: wire chi router with all handlers, middleware, and DI in main.go"
  ```

## Verification

- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./cmd/arch-stats` — compiles.
- `cd backend && go test ./... -count=1` — all existing tests still pass.
- Manual: start server, `curl /api/v0/faces` returns 200 with JSON.
