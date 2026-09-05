package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestCORS_DevModePreflight(t *testing.T) {
	corsMw := middleware.CORS(true)
	handler := corsMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("preflight OPTIONS request should not reach downstream handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v0/session", http.NoBody)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:5173" {
		t.Errorf("Allow-Origin = %q, want http://localhost:5173", origin)
	}
	if creds := rec.Header().Get("Access-Control-Allow-Credentials"); creds != "true" {
		t.Errorf("Allow-Credentials = %q, want 'true'", creds)
	}
	if methods := rec.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("Allow-Methods is empty")
	}
	if headers := rec.Header().Get("Access-Control-Allow-Headers"); headers == "" {
		t.Error("Allow-Headers is empty")
	}
}

func TestCORS_DevModeActualRequest(t *testing.T) {
	corsMw := middleware.CORS(true)
	handlerReached := false
	handler := corsMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerReached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerReached {
		t.Error("actual GET request did not reach downstream handler")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:5173" {
		t.Errorf("Allow-Origin = %q, want http://localhost:5173", origin)
	}
	if creds := rec.Header().Get("Access-Control-Allow-Credentials"); creds != "true" {
		t.Errorf("Allow-Credentials = %q, want 'true'", creds)
	}
}

func TestCORS_ProdModeNoDevOrigins(t *testing.T) {
	corsMw := middleware.CORS(false)
	handlerReached := false
	handler := corsMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerReached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
	req.Header.Set("Origin", "http://malicious.example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerReached {
		t.Error("expected GET request to reach downstream handler in prod")
	}
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin == "http://malicious.example.com" {
		t.Errorf("Allow-Origin reflected untrusted origin in prod: %q", origin)
	}
}
