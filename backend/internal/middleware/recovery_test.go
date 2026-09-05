package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestRecovery_CatchesPanicAndReturns500(t *testing.T) {
	handler := middleware.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went critically wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", http.NoBody)
	rec := httptest.NewRecorder()

	// Should not crash the test process
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp middleware.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Code != "INTERNAL_ERROR" {
		t.Errorf("response Code = %q, want INTERNAL_ERROR", resp.Code)
	}
	if resp.Detail != "internal server error" {
		t.Errorf("response Detail = %q, want 'internal server error'", resp.Detail)
	}
}

func TestRecovery_NormalHandlerUnaffected(t *testing.T) {
	handler := middleware.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthy"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthy", http.NoBody)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "healthy" {
		t.Errorf("body = %q, want 'healthy'", rec.Body.String())
	}
}
