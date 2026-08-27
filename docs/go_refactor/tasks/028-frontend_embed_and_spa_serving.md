# Task 028: Embed Frontend Assets + SPA Serving with Dev Mode

## Git Branch

`refactor/028-frontend-embed-and-spa-serving`

## Objective

Implement frontend asset embedding using `//go:embed` and SPA-aware file serving. In production,
the Go binary serves the pre-built Vue 3 SPA from embedded assets. In dev mode (`--dev` flag),
it serves from the filesystem so the Vite dev server handles HMR.

## Dependencies

- Task 025 (chi router wiring — server is running)
- Task 003 (config — `DevMode` flag)

## Acceptance Criteria

- [ ] `backend/embed.go` uses `//go:embed` to embed the `frontend/` directory (built SPA assets).
- [ ] `backend/internal/handler/spa.go` implements SPA-aware file serving:
    - Serves static files from the embedded filesystem (or disk in dev mode)
    - Falls back to `index.html` for client-side routing (any non-API, non-file path)
    - Sets appropriate cache headers (Cache-Control for hashed assets)
- [ ] `main.go` adds the SPA handler as a catch-all route after API routes.
- [ ] A `--dev` flag or `ARCH_STATS_DEV_MODE=true` env var switches to filesystem serving.
- [ ] Unit tests verify:
    - SPA handler serves `index.html` for root path
    - SPA handler serves `index.html` for unknown paths (SPA fallback)
    - SPA handler serves actual files when they exist (e.g., `/assets/app.js`)
- [ ] `go build ./cmd/arch-stats` compiles (even without a real frontend build — embed handles
    missing dir gracefully or uses a placeholder).
- [ ] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/embed.go` |
| Create | `backend/internal/handler/spa.go` |
| Create | `backend/internal/handler/spa_test.go` |
| Modify | `backend/cmd/arch-stats/main.go` |

## Reference

- High-level plan §3: `//go:embed`, `--dev` flag, Vite HMR in dev mode

## Steps

- [ ] **Step 1: Create the embed directive**

  Create `backend/embed.go`:

  ```go
  package backend

  import "embed"

  // Frontend holds the built Vue 3 SPA assets.
  // In development, the embed may be empty; use filesystem serving instead.
  //go:embed all:frontend
  var Frontend embed.FS
  ```

  Note: The `frontend/` directory here refers to the built output that gets copied into
  `backend/frontend/` during the CI build step. Create a placeholder `backend/frontend/index.html`
  for compilation.

- [ ] **Step 2: Write failing tests for SPA handler**

  Create `backend/internal/handler/spa_test.go`:
    - Test root path `/` serves index.html content
    - Test unknown path `/some/route` serves index.html (SPA fallback)
    - Test existing file path serves the actual file
    - Use `embed.FS` or `os.DirFS` for test filesystem

- [ ] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 4: Implement `spa.go`**

  ```go
  func NewSPAHandler(fs fs.FS, devMode bool, frontendDir string) http.Handler {
      if devMode {
          return http.FileServer(http.Dir(frontendDir))
      }
      return &spaHandler{fs: fs}
  }

  type spaHandler struct {
      fs fs.FS
  }

  func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
      path := r.URL.Path
      // Try to serve the file; if not found, serve index.html
      f, err := h.fs.Open(path)
      if err != nil {
          // SPA fallback: serve index.html for client-side routing
          serveIndex(w, r, h.fs)
          return
      }
      f.Close()
      http.FileServer(http.FS(h.fs)).ServeHTTP(w, r)
  }
  ```

- [ ] **Step 5: Wire into `main.go`**

  Add the SPA handler as a catch-all after all API routes:

  ```go
  // After all /api/v0 routes
  r.Handle("/*", spaHandler)
  ```

- [ ] **Step 6: Run tests to verify they pass**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 7: Run go vet and build**

  ```bash
  cd backend && go vet ./... && go build ./cmd/arch-stats
  ```

- [ ] **Step 8: Commit**

  ```bash
  git add -A
  git commit -m "feat: add frontend asset embedding with SPA serving and dev mode"
  ```

## Verification

- `cd backend && go test ./internal/handler/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./cmd/arch-stats` — compiles.
- Manual: start server, open `http://localhost:8000/` — serves index.html.
