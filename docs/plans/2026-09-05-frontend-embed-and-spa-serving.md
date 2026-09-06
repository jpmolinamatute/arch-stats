# Task 028: Embed Frontend Assets + SPA Serving with Dev Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Vue 3 SPA frontend asset embedding using Go 1.16+ `//go:embed` and SPA-aware file serving with client-side route fallback, cache headers, dev mode filesystem serving via `--dev` flag or `ARCH_STATS_DEV_MODE=true`, chi router catch-all wiring, full unit tests, and completion marking.

**Architecture:**
- `backend/frontend/index.html`: Scaffolds a lightweight placeholder HTML document ensuring `//go:embed` compiles out of the box in all environments without requiring a prior frontend build.
- `backend/embed.go`: Declares `//go:embed all:frontend` on `var Frontend embed.FS` and provides a helper `FS() (fs.FS, error)` returning an `fs.FS` rooted at `frontend/`.
- `backend/internal/handler/spa.go`: Implements `NewSPAHandler(embeddedFS fs.FS, devMode bool, frontendDir string) http.Handler` with `spaHandler`. It handles static file serving, applies Cache-Control headers (`immutable` for hashed `/assets/*`, `no-cache` for `index.html` and dev mode), falls back to `index.html` for client-side SPA routing (while isolating `/api/*` paths to 404), and serves live files from disk in dev mode.
- `backend/cmd/arch-stats/router.go` & `main.go`: Adds `SPAHandler` to `RouterDeps`, mounts `r.Handle("/*", deps.SPAHandler)` as catch-all after `/api/v0` route groups, parses `--dev` CLI flag to override `cfg.DevMode`, initializes `spaHandler`, and wires into the HTTP server.
- `docs/go_refactor/tasks/028-frontend_embed_and_spa_serving.md` & `docs/plans/task.md`: Updates all acceptance criteria and steps to marked as completed upon full verification.

**Tech Stack:** Go 1.27, `//go:embed`, `io/fs`, `testing/fstest`, `net/http`, `net/http/httptest`, `github.com/go-chi/chi/v5`.

**Spec:**
- [docs/go_refactor/tasks/028-frontend_embed_and_spa_serving.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/028-frontend_embed_and_spa_serving.md)
- [docs/go_refactor/high_level_refactoring_plan.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/high_level_refactoring_plan.md) §3
- [backend/cmd/arch-stats/main.go](file:///home/juanpa/Projects/arch-stats/backend/cmd/arch-stats/main.go)
- [backend/cmd/arch-stats/router.go](file:///home/juanpa/Projects/arch-stats/backend/cmd/arch-stats/router.go)

## Global Constraints

- Git branch: `refactor/028-frontend-embed-and-spa-serving`
- Clean compilation: `go build ./cmd/arch-stats` must compile even without a prior frontend build.
- Routing isolation: Any request under `/api/` (e.g. `/api/v0/invalid`) must return 404, never falling back to `index.html`.
- Caching policy:
  - Hashed assets (`/assets/*`): `Cache-Control: public, max-age=31536000, immutable`.
  - HTML & SPA fallback: `Cache-Control: no-cache, no-store, must-revalidate`.
  - Dev mode: `Cache-Control: no-cache, no-store, must-revalidate` for all requests.
  - Other assets: `Cache-Control: public, max-age=3600`.
- Method restrictions: Non-GET and non-HEAD requests to static routes return 405 Method Not Allowed.
- Concurrency & race safety: Zero data races with `go test -race ./...`.
- Code quality & formatting: Clean `go vet ./...` and `golangci-lint run ./...`.
- Single-flow execution: Exactly one active task tracked in `docs/plans/task.md`.
- Task completion: Mark all acceptance criteria and steps in `docs/go_refactor/tasks/028-frontend_embed_and_spa_serving.md` as done (`[x]`) and update `docs/plans/task.md` with all tasks `DONE`.

---

## File Structure

```
backend/
├── embed.go                             # [NEW] Package backend embed directive for all:frontend and FS() helper
├── frontend/
│   └── index.html                       # [NEW] Placeholder SPA index.html for compilation before FE build
├── internal/
│   └── handler/
│       ├── spa.go                       # [NEW] NewSPAHandler, spaHandler, caching & SPA fallback logic
│       └── spa_test.go                  # [NEW] Comprehensive unit test suite with fstest.MapFS and disk temp dirs
└── cmd/
    └── arch-stats/
        ├── router.go                    # [MODIFY] Add SPAHandler to RouterDeps and mount r.Handle("/*", deps.SPAHandler)
        ├── router_test.go               # [MODIFY] Add router-level SPA fallback and static route tests
        └── main.go                      # [MODIFY] Parse --dev flag, construct SPAHandler, and wire into buildRouter
docs/
├── plans/
│   ├── task.md                          # [MODIFY] Track Task 028 live checklist progress (table-only)
│   └── 2026-09-05-frontend-embed-and-spa-serving.md # [NEW] This implementation plan document
└── go_refactor/
    └── tasks/
        └── 028-frontend_embed_and_spa_serving.md # [MODIFY] Mark all acceptance criteria and steps as completed
```

---

### Task 1: Git Branch Setup & Live Tracker Initialization

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Clean working tree on `main`.
- Produces: Checked out branch `refactor/028-frontend-embed-and-spa-serving` and initialized `docs/plans/task.md`.

- [ ] **Step 1: Create and switch to git branch**

```bash
git checkout -b refactor/028-frontend-embed-and-spa-serving
```

- [ ] **Step 2: Initialize `docs/plans/task.md` with Task 028 checklist table**

Update `docs/plans/task.md` to:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | IN_PROGRESS | Switch to branch `refactor/028-frontend-embed-and-spa-serving` and initialize live progress tracker |
| Task 2: Embedded Frontend Assets Scaffold (`embed.go` & placeholder `index.html`) | PENDING | Create placeholder `backend/frontend/index.html` and `backend/embed.go` with embed directive |
| Task 3: SPA Handler Unit Tests Suite (`spa_test.go`) | PENDING | Write comprehensive unit tests for root, fallback, asset serving, cache headers, dev mode, and error handling |
| Task 4: SPA Handler Implementation (`spa.go`) | PENDING | Implement `NewSPAHandler`, `spaHandler`, cache headers, file serving, and SPA route fallback |
| Task 5: Chi Router Integration & Dev Mode Wiring (`router.go`, `router_test.go`, `main.go`) | PENDING | Wire `SPAHandler` into router catch-all, parse `--dev` CLI flag in `main.go`, and add router integration tests |
| Task 6: Full Verification, Linting, and Clean Build | PENDING | Run full test suite with `-race`, `go vet`, `golangci-lint`, and `go build ./cmd/arch-stats` |
| Task 7: Mark Tasks as Completed in Task Spec and Live Tracker | PENDING | Mark `028-frontend_embed_and_spa_serving.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Commit initial tracker setup**

```bash
git add docs/plans/task.md docs/plans/2026-09-05-frontend-embed-and-spa-serving.md
git commit -m "docs: initialize task tracker for task 028 frontend embed and spa serving"
```

---

### Task 2: Embedded Frontend Assets Scaffold (`embed.go` & placeholder `index.html`)

**Files:**
- Create: `backend/frontend/index.html`
- Create: `backend/embed.go`

**Interfaces:**
- Consumes: Go 1.16+ `embed` package.
- Produces: `backend.Frontend` (`embed.FS`) and `backend.FS() (fs.FS, error)`.

- [ ] **Step 1: Create placeholder `backend/frontend/index.html`**

Create `backend/frontend/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Arch Stats</title>
  </head>
  <body>
    <div id="app">Arch Stats SPA Placeholder</div>
  </body>
</html>
```

- [ ] **Step 2: Create `backend/embed.go`**

Create `backend/embed.go`:

```go
package backend

import (
	"embed"
	"fmt"
	"io/fs"
)

// Frontend holds the built Vue 3 SPA assets.
// In development, the embed may be a placeholder; use filesystem serving instead.
//
//go:embed all:frontend
var Frontend embed.FS

// FS returns an fs.FS sub-tree rooted at the embedded "frontend" directory.
func FS() (fs.FS, error) {
	sub, err := fs.Sub(Frontend, "frontend")
	if err != nil {
		return nil, fmt.Errorf("opening embedded frontend subtree: %w", err)
	}
	return sub, nil
}
```

- [ ] **Step 3: Verify package compiles and passes vet**

Run:
```bash
cd backend && go vet ./...
```
Expected: clean exit status 0.

- [ ] **Step 4: Update task tracker to mark Task 2 completed**

Update `docs/plans/task.md`: Task 2 marked as `DONE`, Task 3 marked as `IN_PROGRESS`.

- [ ] **Step 5: Commit scaffold**

```bash
git add backend/frontend/index.html backend/embed.go docs/plans/task.md
git commit -m "feat(backend): add embed directive and placeholder frontend assets"
```

---

### Task 3: SPA Handler Unit Tests Suite (`spa_test.go`)

**Files:**
- Create: `backend/internal/handler/spa_test.go`

**Interfaces:**
- Consumes: `testing/fstest`, `net/http/httptest`, `io/fs`, `handler.NewSPAHandler`.
- Produces: Test suite validating all acceptance criteria.

- [ ] **Step 1: Write failing unit tests in `backend/internal/handler/spa_test.go`**

Create `backend/internal/handler/spa_test.go`:

```go
package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
)

func newTestMockFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>Root Index</body></html>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('app bundle');"),
		},
		"assets/style.css": &fstest.MapFile{
			Data: []byte("body { margin: 0; }"),
		},
		"favicon.ico": &fstest.MapFile{
			Data: []byte("favicon-content"),
		},
	}
}

func TestSPAHandler_RootPath(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, false, "")

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Root Index") {
		t.Errorf("expected body to contain 'Root Index', got: %s", body)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got: %s", contentType)
	}

	cacheControl := rec.Header().Get("Cache-Control")
	if !strings.Contains(cacheControl, "no-cache") {
		t.Errorf("expected Cache-Control to contain no-cache, got: %s", cacheControl)
	}
}

func TestSPAHandler_DirectIndexHtml(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, false, "")

	req := httptest.NewRequest(http.MethodGet, "/index.html", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Root Index") {
		t.Errorf("expected body to contain 'Root Index', got: %s", body)
	}
}

func TestSPAHandler_SPAFallback_UnknownRoute(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, false, "")

	testPaths := []string{
		"/some/route",
		"/sessions",
		"/sessions/00000000-0000-0000-0000-000000000001",
		"/archers/42/details",
	}

	for _, path := range testPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200 for SPA route %s, got %d", path, rec.Code)
			}

			body := rec.Body.String()
			if !strings.Contains(body, "Root Index") {
				t.Errorf("expected fallback to index.html for %s, got: %s", path, body)
			}

			cacheControl := rec.Header().Get("Cache-Control")
			if !strings.Contains(cacheControl, "no-cache") {
				t.Errorf("expected Cache-Control no-cache for fallback route %s, got: %s", path, cacheControl)
			}
		})
	}
}

func TestSPAHandler_ServesExistingAsset(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, false, "")

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "console.log('app bundle')") {
		t.Errorf("expected actual asset body, got: %s", body)
	}

	cacheControl := rec.Header().Get("Cache-Control")
	expectedCache := "public, max-age=31536000, immutable"
	if cacheControl != expectedCache {
		t.Errorf("expected Cache-Control %q, got %q", expectedCache, cacheControl)
	}
}

func TestSPAHandler_ServesFavicon(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, false, "")

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	cacheControl := rec.Header().Get("Cache-Control")
	if !strings.Contains(cacheControl, "public, max-age=3600") {
		t.Errorf("expected max-age=3600 for non-hashed asset, got %q", cacheControl)
	}
}

func TestSPAHandler_ApiRoute_ReturnsNotFound(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, false, "")

	apiRoutes := []string{
		"/api",
		"/api/",
		"/api/v0",
		"/api/v0/nonexistent",
		"/api/v0/faces/unknown-endpoint",
	}

	for _, route := range apiRoutes {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route, http.NoBody)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Errorf("expected status 404 for API route %s, got %d", route, rec.Code)
			}
			if strings.Contains(rec.Body.String(), "Root Index") {
				t.Errorf("API route %s must not fall back to index.html", route)
			}
		})
	}
}

func TestSPAHandler_MissingAsset_ReturnsNotFound(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, false, "")

	req := httptest.NewRequest(http.MethodGet, "/assets/nonexistent.js", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for missing asset, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "Root Index") {
		t.Error("missing asset in /assets/ must not fall back to index.html")
	}
}

func TestSPAHandler_MethodNotAllowed(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, false, "")

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/some/route", http.NoBody)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405 for %s, got %d", method, rec.Code)
			}
		})
	}
}

func TestSPAHandler_HeadRequest(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, false, "")

	req := httptest.NewRequest(http.MethodHead, "/", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 for HEAD, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for HEAD request, got %d bytes", rec.Body.Len())
	}
}

func TestSPAHandler_DevMode_DiskServing(t *testing.T) {
	tmpDir := t.TempDir()

	indexContent := "<!DOCTYPE html><html><body>Dev Disk Index</body></html>"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(indexContent), 0o600); err != nil {
		t.Fatalf("failed to write dev index.html: %v", err)
	}

	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "dev.js"), []byte("console.log('dev');"), 0o600); err != nil {
		t.Fatalf("failed to write dev asset: %v", err)
	}

	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, true, tmpDir)

	// 1. Root path serves dev disk index
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Dev Disk Index") {
		t.Fatalf("expected dev disk index, got %d body: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Cache-Control"), "no-cache") {
		t.Errorf("expected dev mode cache-control no-cache, got: %s", rec.Header().Get("Cache-Control"))
	}

	// 2. Dev asset serves live file and has no-cache in dev mode
	req = httptest.NewRequest(http.MethodGet, "/assets/dev.js", http.NoBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "console.log('dev')") {
		t.Fatalf("expected dev asset content, got %d body: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Cache-Control"), "no-cache") {
		t.Errorf("expected dev mode asset to have no-cache, got: %s", rec.Header().Get("Cache-Control"))
	}

	// 3. SPA fallback on unknown path in dev mode
	req = httptest.NewRequest(http.MethodGet, "/sessions/live", http.NoBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Dev Disk Index") {
		t.Fatalf("expected SPA fallback in dev mode, got %d body: %s", rec.Code, rec.Body.String())
	}
}

func TestSPAHandler_DevMode_FallbackToEmbeddedWhenDirMissing(t *testing.T) {
	mockFS := newTestMockFS()
	h := handler.NewSPAHandler(mockFS, true, "/nonexistent/dev/dir/does/not/exist")

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Root Index") {
		t.Errorf("expected graceful fallback to embedded FS, got: %s", rec.Body.String())
	}
}

func TestSPAHandler_MissingIndexHtml_ReturnsNotFound(t *testing.T) {
	emptyFS := fstest.MapFS{}
	h := handler.NewSPAHandler(emptyFS, false, "")

	req := httptest.NewRequest(http.MethodGet, "/unknown", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 when index.html is missing, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run test suite to verify it fails compilation / execution**

Run:
```bash
cd backend && go test ./internal/handler/spa_test.go ./internal/handler/spa.go
```
Expected: FAIL (handler.NewSPAHandler undefined).

- [ ] **Step 3: Update task tracker**

Update `docs/plans/task.md`: Task 3 marked as `IN_PROGRESS`.

---

### Task 4: SPA Handler Implementation (`spa.go`)

**Files:**
- Create: `backend/internal/handler/spa.go`

**Interfaces:**
- Consumes: `io/fs`, `net/http`, `path`, `strings`.
- Produces: `handler.NewSPAHandler(embeddedFS fs.FS, devMode bool, frontendDir string) http.Handler`.

- [ ] **Step 1: Write `backend/internal/handler/spa.go`**

Create `backend/internal/handler/spa.go`:

```go
package handler

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

// spaHandler implements SPA-aware static asset serving with client-side route fallback.
type spaHandler struct {
	fs      fs.FS
	devMode bool
}

// NewSPAHandler constructs an http.Handler that serves static files and falls back to index.html.
// In devMode, if frontendDir exists on the filesystem, it serves live files from disk.
// Otherwise, it serves from the provided embedded filesystem.
func NewSPAHandler(embeddedFS fs.FS, devMode bool, frontendDir string) http.Handler {
	targetFS := embeddedFS
	if devMode && frontendDir != "" {
		if info, err := os.Stat(frontendDir); err == nil && info.IsDir() {
			targetFS = os.DirFS(frontendDir)
		}
	}
	return &spaHandler{
		fs:      targetFS,
		devMode: devMode,
	}
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	urlPath := path.Clean(r.URL.Path)

	// Never serve SPA fallback for API routes
	if strings.HasPrefix(urlPath, "/api/") || urlPath == "/api" {
		http.NotFound(w, r)
		return
	}

	relPath := strings.TrimPrefix(urlPath, "/")
	if relPath == "" || relPath == "." {
		relPath = "index.html"
	}

	// Check if the requested file exists in the filesystem
	if f, err := h.fs.Open(relPath); err == nil {
		stat, statErr := f.Stat()
		_ = f.Close()
		if statErr == nil && !stat.IsDir() {
			h.setCacheHeaders(w, relPath)
			http.FileServerFS(h.fs).ServeHTTP(w, r)
			return
		}
	}

	// Missing assets in /assets/ must return 404, not fall back to index.html
	if strings.HasPrefix(urlPath, "/assets/") {
		http.NotFound(w, r)
		return
	}

	// SPA route fallback: serve index.html
	h.serveIndex(w, r)
}

func (h *spaHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	f, err := h.fs.Open("index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.setIndexCacheHeaders(w)

	if seeker, ok := f.(io.ReadSeeker); ok {
		http.ServeContent(w, r, "index.html", stat.ModTime(), seeker)
		return
	}

	data, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "failed to read index.html", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *spaHandler) setCacheHeaders(w http.ResponseWriter, relPath string) {
	if h.devMode {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		return
	}

	if relPath == "index.html" {
		h.setIndexCacheHeaders(w)
		return
	}

	// Hashed Vite assets in /assets/
	if strings.HasPrefix(relPath, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}

	// Other static assets (favicon.ico, manifest.json, etc.)
	w.Header().Set("Cache-Control", "public, max-age=3600")
}

func (h *spaHandler) setIndexCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
}
```

- [ ] **Step 2: Run handler tests to verify they pass**

Run:
```bash
cd backend && go test -v ./internal/handler/... -run TestSPAHandler
```
Expected: PASS for all 10 unit test cases.

- [ ] **Step 3: Update task tracker to mark Task 3 and Task 4 completed**

Update `docs/plans/task.md`: Task 3 and Task 4 marked as `DONE`, Task 5 marked as `IN_PROGRESS`.

- [ ] **Step 4: Commit SPA handler implementation**

```bash
git add backend/internal/handler/spa.go backend/internal/handler/spa_test.go docs/plans/task.md
git commit -m "feat(backend): implement SPA-aware file serving with caching and fallback"
```

---

### Task 5: Chi Router Integration & Dev Mode Wiring (`router.go`, `router_test.go`, `main.go`)

**Files:**
- Modify: `backend/cmd/arch-stats/router.go`
- Modify: `backend/cmd/arch-stats/router_test.go`
- Modify: `backend/cmd/arch-stats/main.go`

**Interfaces:**
- Consumes: `handler.NewSPAHandler`, `backend.FS()`, `RouterDeps.SPAHandler`.
- Produces: Complete Chi router catch-all route and `--dev` flag CLI parsing.

- [ ] **Step 1: Update `RouterDeps` and `buildRouter` in `backend/cmd/arch-stats/router.go`**

Add `SPAHandler http.Handler` to `RouterDeps`:

```go
type RouterDeps struct {
	Cfg            *config.Config
	Logger         *slog.Logger
	AuthSvc        middleware.TokenAuthenticator
	AuthHandler    *handler.AuthHandler
	ArcherHandler  *handler.ArcherHandler
	SessionHandler *handler.SessionHandler
	SlotHandler    *handler.SlotHandler
	ShotHandler    *handler.ShotHandler
	FaceHandler    *handler.FaceHandler
	HealthHandler  *handler.HealthHandler
	SPAHandler     http.Handler
}
```

And at the end of `buildRouter(deps *RouterDeps) chi.Router`, before `return r`:

```go
	// Catch-all SPA handler mounted after all API routes
	if deps.SPAHandler != nil {
		r.Handle("/*", deps.SPAHandler)
	}

	return r
```

- [ ] **Step 2: Add SPA router tests to `backend/cmd/arch-stats/router_test.go`**

In `setupTestRouter()` in `router_test.go`, provide an SPAHandler in `deps`:

```go
	mockSPA := handler.NewSPAHandler(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!DOCTYPE html><html><body>Router Test Index</body></html>")},
		"assets/test.js": &fstest.MapFile{Data: []byte("console.log('router test asset');")},
	}, false, "")
```
And pass `SPAHandler: mockSPA` into `deps`.

Add tests `TestRouter_SPAFallbackAndAssets(t *testing.T)`:

```go
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
```

- [ ] **Step 3: Update `backend/cmd/arch-stats/main.go`**

1. Import `github.com/jpmolinamatute/arch-stats/backend`.
2. Check `os.Args` for `--dev` or `-dev` immediately after `config.Load()`:

```go
	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return err
	}

	for _, arg := range os.Args[1:] {
		if arg == "--dev" || arg == "-dev" {
			cfg.DevMode = true
			break
		}
	}
```

3. In Section 8 (Handlers), initialize the SPA handler:

```go
	// 8b. SPA Handler
	frontendFS, err := backend.FS()
	if err != nil {
		slog.Error("failed to load embedded frontend filesystem", "error", err)
		return err
	}

	frontendDir := os.Getenv("ARCH_STATS_FRONTEND_DIR")
	if frontendDir == "" {
		for _, candidate := range []string{"frontend", "backend/frontend", "../frontend/dist", "dist"} {
			if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
				frontendDir = candidate
				break
			}
		}
	}

	spaHandler := handler.NewSPAHandler(frontendFS, cfg.DevMode, frontendDir)
```

4. Pass `SPAHandler: spaHandler` in `routerDeps`:

```go
	routerDeps := RouterDeps{
		Cfg:            cfg,
		Logger:         logger,
		AuthSvc:        authSvc,
		AuthHandler:    authHandler,
		ArcherHandler:  archerHandler,
		SessionHandler: sessionHandler,
		SlotHandler:    slotHandler,
		ShotHandler:    shotHandler,
		FaceHandler:    faceHandler,
		HealthHandler:  healthHandler,
		SPAHandler:     spaHandler,
	}
	router := buildRouter(&routerDeps)
```

- [ ] **Step 4: Run all cmd/arch-stats tests to verify router wiring**

Run:
```bash
cd backend && go test -v ./cmd/arch-stats/...
```
Expected: PASS for all tests including the new SPA router integration tests.

- [ ] **Step 5: Update task tracker to mark Task 5 completed**

Update `docs/plans/task.md`: Task 5 marked as `DONE`, Task 6 marked as `IN_PROGRESS`.

- [ ] **Step 6: Commit router and main wiring**

```bash
git add backend/cmd/arch-stats/router.go backend/cmd/arch-stats/router_test.go backend/cmd/arch-stats/main.go docs/plans/task.md
git commit -m "feat(cmd): wire SPA handler into chi router with --dev flag support"
```

---

### Task 6: Full Verification, Linting, and Clean Build

**Files:**
- None (verification commands across whole backend)

**Interfaces:**
- Consumes: Go compiler, race detector, `golangci-lint`, `go vet`.
- Produces: Clean build output binary `backend/arch-stats`.

- [ ] **Step 1: Run full unit test suite with race detector**

Run:
```bash
cd backend && go test -race -count=1 ./... -v
```
Expected: All tests pass with 0 race condition warnings.

- [ ] **Step 2: Run `go vet`**

Run:
```bash
cd backend && go vet ./...
```
Expected: 0 warnings.

- [ ] **Step 3: Run `golangci-lint`**

Run:
```bash
cd backend && golangci-lint run ./...
```
Expected: 0 lint errors.

- [ ] **Step 4: Build backend binary**

Run:
```bash
cd backend && go build -o arch-stats ./cmd/arch-stats
```
Expected: Clean compilation, binary created at `backend/arch-stats`.

- [ ] **Step 5: Verify binary help / execution**

Run:
```bash
cd backend && ./arch-stats --dev -h 2>&1 || true
```
Clean up compiled binary:
```bash
rm -f backend/arch-stats
```

- [ ] **Step 6: Update task tracker to mark Task 6 completed**

Update `docs/plans/task.md`: Task 6 marked as `DONE`, Task 7 marked as `IN_PROGRESS`.

---

### Task 7: Mark Tasks as Completed in Task Spec and Live Tracker

**Files:**
- Modify: `docs/go_refactor/tasks/028-frontend_embed_and_spa_serving.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified test and build results.
- Produces: All checkboxes marked `[x]` in Task 028 specification and all statuses `DONE` in `task.md`.

- [ ] **Step 1: Mark all Acceptance Criteria and Steps in `docs/go_refactor/tasks/028-frontend_embed_and_spa_serving.md`**

Check off all items:
- Acceptance Criteria:
  - `[x] backend/embed.go uses //go:embed to embed the frontend/ directory (built SPA assets).`
  - `[x] backend/internal/handler/spa.go implements SPA-aware file serving...`
  - `[x] main.go adds the SPA handler as a catch-all route after API routes.`
  - `[x] A --dev flag or ARCH_STATS_DEV_MODE=true env var switches to filesystem serving.`
  - `[x] Unit tests verify: SPA handler serves index.html for root path, unknown paths, actual files...`
  - `[x] go build ./cmd/arch-stats compiles...`
  - `[x] go vet ./... reports no issues.`
- Steps:
  - `[x] Step 1: Create the embed directive`
  - `[x] Step 2: Write failing tests for SPA handler`
  - `[x] Step 3: Run tests to verify they fail`
  - `[x] Step 4: Implement spa.go`
  - `[x] Step 5: Wire into main.go`
  - `[x] Step 6: Run tests to verify they pass`
  - `[x] Step 7: Run go vet and build`
  - `[x] Step 8: Commit`

- [ ] **Step 2: Update `docs/plans/task.md` with all tasks `DONE`**

Update `docs/plans/task.md`:
```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | DONE | Switch to branch `refactor/028-frontend-embed-and-spa-serving` and initialize live progress tracker |
| Task 2: Embedded Frontend Assets Scaffold (`embed.go` & placeholder `index.html`) | DONE | Create placeholder `backend/frontend/index.html` and `backend/embed.go` with embed directive |
| Task 3: SPA Handler Unit Tests Suite (`spa_test.go`) | DONE | Write comprehensive unit tests for root, fallback, asset serving, cache headers, dev mode, and error handling |
| Task 4: SPA Handler Implementation (`spa.go`) | DONE | Implement `NewSPAHandler`, `spaHandler`, cache headers, file serving, and SPA route fallback |
| Task 5: Chi Router Integration & Dev Mode Wiring (`router.go`, `router_test.go`, `main.go`) | DONE | Wire `SPAHandler` into router catch-all, parse `--dev` CLI flag in `main.go`, and add router integration tests |
| Task 6: Full Verification, Linting, and Clean Build | DONE | Run full test suite with `-race`, `go vet`, `golangci-lint`, and `go build ./cmd/arch-stats` |
| Task 7: Mark Tasks as Completed in Task Spec and Live Tracker | DONE | Mark `028-frontend_embed_and_spa_serving.md` checklist and `docs/plans/task.md` as DONE |
```

- [ ] **Step 3: Commit task completion updates**

```bash
git add docs/go_refactor/tasks/028-frontend_embed_and_spa_serving.md docs/plans/task.md
git commit -m "docs: mark task 028 as completed in spec and tracker"
```
