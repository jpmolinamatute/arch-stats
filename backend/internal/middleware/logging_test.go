package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestRequestLogger_LogsFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	loggingMw := middleware.RequestLogger(logger)

	handler := loggingMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v0/shot", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("User-Agent", "ArchStatsClient/1.0")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("log output is not valid JSON: %v, raw: %s", err, buf.String())
	}

	if logEntry["msg"] != "http request" {
		t.Errorf("msg = %v, want 'http request'", logEntry["msg"])
	}
	if logEntry["method"] != "POST" {
		t.Errorf("method = %v, want 'POST'", logEntry["method"])
	}
	if logEntry["path"] != "/api/v0/shot" {
		t.Errorf("path = %v, want '/api/v0/shot'", logEntry["path"])
	}
	if int(logEntry["status"].(float64)) != http.StatusCreated {
		t.Errorf("status = %v, want %d", logEntry["status"], http.StatusCreated)
	}
	if _, ok := logEntry["duration_ms"]; !ok {
		t.Errorf("missing duration_ms in log entry")
	}
}

func TestRequestLogger_DefaultStatus200(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	loggingMw := middleware.RequestLogger(logger)

	handler := loggingMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok without explicit WriteHeader"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/faces", http.NoBody)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("log output is not JSON: %v", err)
	}
	if int(logEntry["status"].(float64)) != http.StatusOK {
		t.Errorf("status = %v, want %d", logEntry["status"], http.StatusOK)
	}
}
