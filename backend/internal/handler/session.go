package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// SessionService defines domain operations required by SessionHandler.
type SessionService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	GetOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	Close(ctx context.Context, id uuid.UUID) error
	ReOpen(ctx context.Context, id uuid.UUID) error
	GetParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error)
}

// SessionHandler manages HTTP endpoints for session lifecycle and querying.
type SessionHandler struct {
	sessionSvc SessionService
}

// NewSessionHandler constructs a SessionHandler with service dependency injection.
func NewSessionHandler(sessionSvc SessionService) *SessionHandler {
	return &SessionHandler{
		sessionSvc: sessionSvc,
	}
}

// Routes registers all session lifecycle and query endpoints on the provided chi Router.
func (h *SessionHandler) Routes(r chi.Router) {
	r.Get("/archer/{archer_id}/open-session", h.GetOpenForArcher)
	r.Get("/archer/{archer_id}/close-session", h.GetClosedForArcher)
	r.Get("/archer/{archer_id}/participating", h.GetParticipating)
	r.Get("/open", h.ListAllOpen)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Patch("/re-open", h.ReOpen)
	r.Patch("/close", h.Close)
}

// GetOpenForArcher handles GET /api/v0/session/archer/{archer_id}/open-session.
func (h *SessionHandler) GetOpenForArcher(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// GetClosedForArcher handles GET /api/v0/session/archer/{archer_id}/close-session.
func (h *SessionHandler) GetClosedForArcher(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// GetParticipating handles GET /api/v0/session/archer/{archer_id}/participating.
func (h *SessionHandler) GetParticipating(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// ListAllOpen handles GET /api/v0/session/open.
func (h *SessionHandler) ListAllOpen(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// Create handles POST /api/v0/session.
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// GetByID handles GET /api/v0/session/{id}.
func (h *SessionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// ReOpen handles PATCH /api/v0/session/re-open.
func (h *SessionHandler) ReOpen(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// Close handles PATCH /api/v0/session/close.
func (h *SessionHandler) Close(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
