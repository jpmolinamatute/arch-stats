package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockMaintenanceService struct {
	getSchemaVersionFn func(ctx context.Context) (int64, error)
}

func (m *mockMaintenanceService) GetSchemaVersion(ctx context.Context) (int64, error) {
	if m.getSchemaVersionFn != nil {
		return m.getSchemaVersionFn(ctx)
	}
	return 0, nil
}

func TestHealthHandler_Health_Success(t *testing.T) {
	mockSvc := &mockMaintenanceService{
		getSchemaVersionFn: func(ctx context.Context) (int64, error) {
			return 42, nil
		},
	}
	h := NewHealthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/health", http.NoBody)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.SchemaVersion != 42 {
		t.Errorf("expected schema_version 42, got %d", resp.SchemaVersion)
	}
}

func TestHealthHandler_Health_DatabaseError(t *testing.T) {
	mockSvc := &mockMaintenanceService{
		getSchemaVersionFn: func(ctx context.Context) (int64, error) {
			return 0, errors.New("database connection refused")
		},
	}
	h := NewHealthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/health", http.NoBody)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "unhealthy" {
		t.Errorf("expected status 'unhealthy', got %q", resp.Status)
	}
	if resp.Error != "database connection refused" {
		t.Errorf("expected error message 'database connection refused', got %q", resp.Error)
	}
}
