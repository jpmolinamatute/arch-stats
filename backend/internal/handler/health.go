package handler

import (
	"context"
	"net/http"
)

// MaintenanceService defines schema and maintenance operations required by HealthHandler.
type MaintenanceService interface {
	GetSchemaVersion(ctx context.Context) (int64, error)
}

// HealthResponse represents the payload returned by GET /api/v0/health.
type HealthResponse struct {
	Status        string `json:"status"`
	SchemaVersion int64  `json:"schema_version"`
	Error         string `json:"error,omitempty"`
}

// HealthHandler manages health check endpoints.
type HealthHandler struct {
	maintenance MaintenanceService
}

// NewHealthHandler constructs a HealthHandler with maintenance service dependency injection.
func NewHealthHandler(maintenance MaintenanceService) *HealthHandler {
	return &HealthHandler{maintenance: maintenance}
}

// Health godoc
// @Summary     Health Check
// @Description Returns current service health and database migration schema version. Public endpoint.
// @Tags        System
// @Produce     json
// @Success     200 {object} HealthResponse
// @Failure     503 {object} HealthResponse
// @Router      /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ver, err := h.maintenance.GetSchemaVersion(r.Context())
	if err != nil {
		_ = WriteJSON(w, http.StatusServiceUnavailable, HealthResponse{
			Status: "unhealthy",
			Error:  err.Error(),
		})
		return
	}

	_ = WriteJSON(w, http.StatusOK, HealthResponse{
		Status:        "ok",
		SchemaVersion: ver,
	})
}
