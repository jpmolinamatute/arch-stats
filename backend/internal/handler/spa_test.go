package handler_test

import (
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

	for _, p := range testPaths {
		t.Run(p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p, http.NoBody)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200 for SPA route %s, got %d", p, rec.Code)
			}

			body := rec.Body.String()
			if !strings.Contains(body, "Root Index") {
				t.Errorf("expected fallback to index.html for %s, got: %s", p, body)
			}

			cacheControl := rec.Header().Get("Cache-Control")
			if !strings.Contains(cacheControl, "no-cache") {
				t.Errorf("expected Cache-Control no-cache for fallback route %s, got: %s", p, cacheControl)
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
