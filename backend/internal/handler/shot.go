package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ShotService defines the domain operations required by ShotHandler.
type ShotService interface {
	Create(ctx context.Context, shot model.ShotCreate, archerID uuid.UUID) (uuid.UUID, error)
	CreateBatch(ctx context.Context, shots []model.ShotCreate, archerID uuid.UUID) ([]uuid.UUID, error)
	GetBySlot(ctx context.Context, slotID, archerID uuid.UUID) ([]model.ShotRead, error)
	CountBySlot(ctx context.Context, slotID, archerID uuid.UUID) (int, error)
}

// ShotHandler manages HTTP endpoints for shot records.
type ShotHandler struct {
	shotSvc ShotService
}

// NewShotHandler constructs a ShotHandler with service dependency injection.
func NewShotHandler(shotSvc ShotService) *ShotHandler {
	return &ShotHandler{
		shotSvc: shotSvc,
	}
}

// Routes registers all shot management endpoints on the provided chi Router.
func (h *ShotHandler) Routes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/by-slot/{slot_id}", h.GetBySlot)
	r.Get("/count-by-slot/{slot_id}", h.CountBySlot)
}

// Create godoc
// @Summary     Create Shot
// @Description Record a single shot or a batch array of shots (3 to 10 shots) for the authenticated archer
// @Tags        Shots
// @Accept      json
// @Produce     json
// @Param       request body model.ShotCreate true "Shot creation payload (object or array)"
// @Success     201 {object} model.ShotId
// @Failure     400 {object} model.ErrorResponse
// @Failure     401 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /shot [post]
func (h *ShotHandler) Create(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		WriteAppError(w, err)
		return
	}

	if r.Body == nil {
		WriteAppError(w, apperror.Wrap(apperror.ErrValidation, "request body is empty"))
		return
	}
	defer r.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		WriteAppError(w, apperror.Wrap(apperror.ErrValidation, "failed to read request body"))
		return
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		WriteAppError(w, apperror.Wrap(apperror.ErrValidation, "request body is empty"))
		return
	}

	if trimmed[0] == '[' {
		h.handleBatchCreate(w, r.Context(), trimmed, authArcherID)
		return
	}

	h.handleSingleCreate(w, r.Context(), trimmed, authArcherID)
}

func (h *ShotHandler) handleSingleCreate(w http.ResponseWriter, ctx context.Context, raw []byte, authArcherID uuid.UUID) {
	var shot model.ShotCreate
	if err := json.Unmarshal(raw, &shot); err != nil {
		WriteAppError(w, apperror.Wrap(apperror.ErrValidation, "invalid request body: "+err.Error()))
		return
	}

	id, err := h.shotSvc.Create(ctx, shot, authArcherID)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusCreated, model.ShotID{ShotID: id})
}

func (h *ShotHandler) handleBatchCreate(w http.ResponseWriter, ctx context.Context, raw []byte, authArcherID uuid.UUID) {
	var shots []model.ShotCreate
	if err := json.Unmarshal(raw, &shots); err != nil {
		WriteAppError(w, apperror.Wrap(apperror.ErrValidation, "invalid request body: "+err.Error()))
		return
	}

	if len(shots) < 3 || len(shots) > 10 {
		WriteError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	slotID := shots[0].SlotID
	for _, s := range shots[1:] {
		if s.SlotID != slotID {
			WriteError(w, http.StatusBadRequest, "All shots must belong to the same slot")
			return
		}
	}

	ids, err := h.shotSvc.CreateBatch(ctx, shots, authArcherID)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	resp := make([]model.ShotID, len(ids))
	for i, id := range ids {
		resp[i] = model.ShotID{ShotID: id}
	}

	_ = WriteJSON(w, http.StatusCreated, resp)
}

// GetBySlot godoc
// @Summary     Get Shots By Slot
// @Description Returns list of shots for the given slot owned by the authenticated archer
// @Tags        Shots
// @Produce     json
// @Param       slot_id path string true "Slot UUID"
// @Success     200 {array} model.ShotRead
// @Failure     401 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /shot/by-slot/{slot_id} [get]
func (h *ShotHandler) GetBySlot(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		WriteAppError(w, err)
		return
	}

	slotID, ok := parseUUIDParam(w, r, "slot_id", "slot", "id")
	if !ok {
		return
	}

	shots, err := h.shotSvc.GetBySlot(r.Context(), slotID, authArcherID)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	if shots == nil {
		shots = []model.ShotRead{}
	}

	_ = WriteJSON(w, http.StatusOK, shots)
}

// CountBySlot godoc
// @Summary     Get Shots Count By Slot
// @Description Returns total shot count for the given slot owned by the authenticated archer
// @Tags        Shots
// @Produce     json
// @Param       slot_id path string true "Slot UUID"
// @Success     200 {integer} int
// @Failure     401 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /shot/count-by-slot/{slot_id} [get]
func (h *ShotHandler) CountBySlot(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		WriteAppError(w, err)
		return
	}

	slotID, ok := parseUUIDParam(w, r, "slot_id", "slot", "id")
	if !ok {
		return
	}

	count, err := h.shotSvc.CountBySlot(r.Context(), slotID, authArcherID)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, count)
}
