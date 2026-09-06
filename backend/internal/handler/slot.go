package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// SlotService defines the business operations required by SlotHandler.
type SlotService interface {
	GetArcherCurrentSlot(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error)
	GetSlot(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error)
	JoinSession(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error)
	ReJoinSession(ctx context.Context, slotID, archerID uuid.UUID) (*model.SlotJoinResponse, error)
	LeaveSession(ctx context.Context, slotID, archerID uuid.UUID) error
}

// SlotHandler manages HTTP endpoints for session slot assignments.
type SlotHandler struct {
	slotSvc SlotService
}

// NewSlotHandler constructs a SlotHandler with service dependency injection.
func NewSlotHandler(slotSvc SlotService) *SlotHandler {
	return &SlotHandler{
		slotSvc: slotSvc,
	}
}

// Routes registers all slot management endpoints on the provided chi Router.
func (h *SlotHandler) Routes(r chi.Router) {
	r.Get("/archer/{archer_id}", h.GetArcherCurrentSlot)
	r.Get("/{slot_id}", h.GetSlot)
	r.Post("/", h.JoinSession)
	r.Patch("/re-join/{slot_id}", h.ReJoinSession)
	r.Patch("/leave/{slot_id}", h.LeaveSession)
}

// GetArcherCurrentSlot godoc
// @Summary     Get Archer Current Slot
// @Description Returns active slot assignment for the given archer (open session and is_shooting = true)
// @Tags        Slots
// @Produce     json
// @Param       archer_id path string true "Archer UUID"
// @Success     200 {object} model.FullSlotInfo
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/slot/archer/{archer_id} [get]
func (h *SlotHandler) GetArcherCurrentSlot(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	archerIDStr := getURLParam(r, "archer_id")
	if archerIDStr == "" {
		archerIDStr = getURLParam(r, "id")
	}

	archerID, err := uuid.Parse(archerIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer_id is required"))
		return
	}

	if authArcherID != archerID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	info, err := h.slotSvc.GetArcherCurrentSlot(r.Context(), archerID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, info)
}

// GetSlot godoc
// @Summary     Get Slot
// @Description Returns active slot assignment details by slot UUID
// @Tags        Slots
// @Produce     json
// @Param       slot_id path string true "Slot UUID"
// @Success     200 {object} model.FullSlotInfo
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/slot/{slot_id} [get]
func (h *SlotHandler) GetSlot(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "slot")
	}
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid slot_id is required"))
		return
	}

	info, err := h.slotSvc.GetSlot(r.Context(), slotID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if info.ArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	_ = writeJSON(w, http.StatusOK, info)
}

// JoinSession godoc
// @Summary     Join Session
// @Description Assigns an archer to a target slot within an open session
// @Tags        Slots
// @Accept      json
// @Produce     json
// @Param       request body model.SlotJoinRequest true "Slot join parameters"
// @Success     200 {object} model.SlotJoinResponse
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/slot [post]
func (h *SlotHandler) JoinSession(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	var req model.SlotJoinRequest
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.ArcherID == uuid.Nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "archer_id is required"))
		return
	}
	if req.SessionID == uuid.Nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "session_id is required"))
		return
	}

	if req.ArcherID != authArcherID {
		writeAppError(w, apperror.Wrap(apperror.ErrForbidden, "Forbidden"))
		return
	}

	resp, err := h.slotSvc.JoinSession(r.Context(), req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, resp)
}

// ReJoinSession godoc
// @Summary     Re Join Session
// @Description Re-activates a previously inactive slot assignment
// @Tags        Slots
// @Produce     json
// @Param       slot_id path string true "Slot UUID"
// @Success     200 {object} model.SlotJoinResponse
// @Failure     401 {object} model.ErrorResponse
// @Failure     403 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/slot/re-join/{slot_id} [patch]
func (h *SlotHandler) ReJoinSession(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "slot")
	}
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid slot_id is required"))
		return
	}

	resp, err := h.slotSvc.ReJoinSession(r.Context(), slotID, authArcherID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, resp)
}

// LeaveSession godoc
// @Summary     Leave Session
// @Description Deactivates an active slot assignment (stop shooting in the session)
// @Tags        Slots
// @Produce     json
// @Param       slot_id path string true "Slot UUID"
// @Success     200
// @Failure     401 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /session/slot/leave/{slot_id} [patch]
func (h *SlotHandler) LeaveSession(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "slot")
	}
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid slot_id is required"))
		return
	}

	if err := h.slotSvc.LeaveSession(r.Context(), slotID, authArcherID); err != nil {
		writeAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
