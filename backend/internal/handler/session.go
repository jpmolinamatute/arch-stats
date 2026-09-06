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

// GetOpenForArcher godoc
// @Summary     Get Open Session For Archer
// @Description Returns the open session ID owned by the archer, or null session_id if none exists
// @Tags        Sessions
// @Produce     json
// @Param       archer_id path string true "Archer UUID"
// @Success     200 {object} model.SessionId
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/archer/{archer_id}/open-session [get]
func (h *SessionHandler) GetOpenForArcher(w http.ResponseWriter, r *http.Request) {
	archerID, ok := requireOwnership(w, r)
	if !ok {
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

// GetClosedForArcher godoc
// @Summary     Get Closed Session For Archer
// @Description Returns all closed sessions owned by the archer
// @Tags        Sessions
// @Produce     json
// @Param       archer_id path string true "Archer UUID"
// @Success     200 {array} model.SessionRead
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/archer/{archer_id}/close-session [get]
func (h *SessionHandler) GetClosedForArcher(w http.ResponseWriter, r *http.Request) {
	archerID, ok := requireOwnership(w, r)
	if !ok {
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

// GetParticipating godoc
// @Summary     Get Participating Session For Archer
// @Description Returns the open session ID the archer is currently participating in, or null if none
// @Tags        Sessions
// @Produce     json
// @Param       archer_id path string true "Archer UUID"
// @Success     200 {object} model.SessionId
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/archer/{archer_id}/participating [get]
func (h *SessionHandler) GetParticipating(w http.ResponseWriter, r *http.Request) {
	archerID, ok := requireOwnership(w, r)
	if !ok {
		return
	}

	sessionID, err := h.sessionSvc.GetParticipating(r.Context(), archerID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, model.SessionID{SessionID: sessionID})
}

// ListAllOpen godoc
// @Summary     Get All Open Sessions
// @Description Returns all currently open sessions in the system
// @Tags        Sessions
// @Produce     json
// @Success     200 {array} model.SessionRead
// @Failure     401 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/open [get]
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

// Create godoc
// @Summary     Create Session
// @Description Creates a new shooting session and returns the assigned session ID
// @Tags        Sessions
// @Accept      json
// @Produce     json
// @Param       request body model.SessionCreate true "Session creation payload"
// @Success     201 {object} model.SessionId
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session [post]
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	var req model.SessionCreate
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.OwnerArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "ERROR: user not allowed to open a session for another archer"))
		return
	}

	id, err := h.sessionSvc.Create(r.Context(), req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusCreated, model.SessionID{SessionID: &id})
}

// GetByID godoc
// @Summary     Get Session
// @Description Returns full session details for the given session UUID
// @Tags        Sessions
// @Produce     json
// @Param       id path string true "Session UUID"
// @Success     200 {object} model.SessionRead
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/{id} [get]
func (h *SessionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	id, ok := parseUUIDParam(w, r, "id", "session")
	if !ok {
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

// ReOpen godoc
// @Summary     Re Open Session
// @Description Re-opens a closed shooting session after verifying owner identity and absence of conflicts
// @Tags        Sessions
// @Accept      json
// @Produce     json
// @Param       request body model.SessionId true "Session identifier payload"
// @Success     200 {object} model.SessionId
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/re-open [patch]
func (h *SessionHandler) ReOpen(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	var req model.SessionID
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.SessionID == nil || *req.SessionID == uuid.Nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "session_id is required"))
		return
	}

	session, err := h.sessionSvc.GetByID(r.Context(), *req.SessionID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if session.OwnerArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Archer is not allowed to re-open this session"))
		return
	}

	if err := h.sessionSvc.ReOpen(r.Context(), *req.SessionID); err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, model.SessionID{SessionID: req.SessionID})
}

// Close godoc
// @Summary     Close Session
// @Description Marks an active shooting session as closed
// @Tags        Sessions
// @Accept      json
// @Produce     json
// @Param       request body model.SessionId true "Session identifier payload"
// @Success     200 {object} map[string]string
// @Failure     400 {object} model.ErrorResponse
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/close [patch]
func (h *SessionHandler) Close(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	var req model.SessionID
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.SessionID == nil || *req.SessionID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "ERROR: session_id wasn't provided")
		return
	}

	session, err := h.sessionSvc.GetByID(r.Context(), *req.SessionID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if session.OwnerArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	if err := h.sessionSvc.Close(r.Context(), *req.SessionID); err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}
