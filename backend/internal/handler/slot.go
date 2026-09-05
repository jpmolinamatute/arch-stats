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

// GetArcherCurrentSlot handles GET /api/v0/session/slot/archer/{archer_id}.
// Returns active slot assignment (open session and is_shooting = true).
// Enforces that only the authenticated archer can query their current slot.
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

// GetSlot handles GET /api/v0/session/slot/{slot_id}.
// Returns active slot assignment details.
// Enforces that only the authenticated archer owning the slot can retrieve it.
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

// JoinSession handles POST /api/v0/session/slot.
func (h *SlotHandler) JoinSession(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// ReJoinSession handles PATCH /api/v0/session/slot/re-join/{slot_id}.
func (h *SlotHandler) ReJoinSession(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// LeaveSession handles PATCH /api/v0/session/slot/leave/{slot_id}.
func (h *SlotHandler) LeaveSession(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
