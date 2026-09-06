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
	if relPath == "" || relPath == "." || relPath == "index.html" {
		h.serveIndex(w, r)
		return
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
