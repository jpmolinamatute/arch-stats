package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
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
// Returns the open session ID owned by the archer, or null session_id if none exists.
// Enforces that only the authenticated archer can query their open session.
func (h *SessionHandler) GetOpenForArcher(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	archerIDStr := getURLParam(r, "archer_id")
	archerID, err := uuid.Parse(archerIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer_id is required"))
		return
	}

	if authArcherID != archerID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	session, err := h.sessionSvc.GetOpen(r.Context(), archerID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			_ = writeJSON(w, http.StatusOK, model.SessionID{SessionID: nil})
			return
		}
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, model.SessionID{SessionID: &session.SessionID})
}

// GetClosedForArcher handles GET /api/v0/session/archer/{archer_id}/close-session.
// Returns all closed sessions owned by the archer.
// Enforces that only the authenticated archer can query their closed sessions.
func (h *SessionHandler) GetClosedForArcher(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	archerIDStr := getURLParam(r, "archer_id")
	archerID, err := uuid.Parse(archerIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer_id is required"))
		return
	}

	if authArcherID != archerID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	isOpened := false
	filter := model.SessionFilter{
		OwnerArcherID: &archerID,
		IsOpened:      &isOpened,
	}

	sessions, err := h.sessionSvc.List(r.Context(), filter)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if sessions == nil {
		sessions = []model.SessionRead{}
	}

	_ = writeJSON(w, http.StatusOK, sessions)
}

// GetParticipating handles GET /api/v0/session/archer/{archer_id}/participating.
// Returns the open session ID the archer is currently participating in, or null if none.
// Enforces that only the authenticated archer can query their participation.
func (h *SessionHandler) GetParticipating(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	archerIDStr := getURLParam(r, "archer_id")
	archerID, err := uuid.Parse(archerIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer_id is required"))
		return
	}

	if authArcherID != archerID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	sessionID, err := h.sessionSvc.GetParticipating(r.Context(), archerID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, model.SessionID{SessionID: sessionID})
}

// ListAllOpen handles GET /api/v0/session/open.
// Returns all currently open sessions. Requires authentication.
func (h *SessionHandler) ListAllOpen(w http.ResponseWriter, r *http.Request) {
	if _, err := middleware.GetArcherID(r.Context()); err != nil {
		writeAppError(w, err)
		return
	}

	isOpened := true
	filter := model.SessionFilter{
		IsOpened: &isOpened,
	}

	sessions, err := h.sessionSvc.List(r.Context(), filter)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if sessions == nil {
		sessions = []model.SessionRead{}
	}

	_ = writeJSON(w, http.StatusOK, sessions)
}

// Create handles POST /api/v0/session.
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// GetByID handles GET /api/v0/session/{id}.
// Returns full session details. Open sessions are readable by any authenticated archer.
// Closed sessions are only readable by the session owner (403 Forbidden otherwise).
func (h *SessionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	idStr := getURLParam(r, "id")
	if idStr == "" {
		idStr = getURLParam(r, "session")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid session id is required"))
		return
	}

	session, err := h.sessionSvc.GetByID(r.Context(), id)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if !session.IsOpened && session.OwnerArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	_ = writeJSON(w, http.StatusOK, session)
}

// ReOpen handles PATCH /api/v0/session/re-open.
func (h *SessionHandler) ReOpen(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// Close handles PATCH /api/v0/session/close.
func (h *SessionHandler) Close(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
