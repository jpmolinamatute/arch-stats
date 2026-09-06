package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ArcherService defines the persistence and business operations required by the archer and auth handlers.
type ArcherService interface {
	List(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	Update(ctx context.Context, id uuid.UUID, data model.ArcherSet) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ArcherHandler manages HTTP endpoints for archer CRUD operations.
type ArcherHandler struct {
	archerSvc ArcherService
}

// NewArcherHandler constructs an ArcherHandler with service dependency injection.
func NewArcherHandler(archerSvc ArcherService) *ArcherHandler {
	return &ArcherHandler{
		archerSvc: archerSvc,
	}
}

// List godoc
// @Summary     List Archers
// @Description Query archers matching default filter criteria and return a JSON list
// @Tags        Archers
// @Produce     json
// @Success     200 {array} model.ArcherRead
// @Failure     401 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /archer/ [get]
func (h *ArcherHandler) List(w http.ResponseWriter, r *http.Request) {
	archers, err := h.archerSvc.List(r.Context(), model.ArcherFilter{})
	if err != nil {
		writeAppError(w, err)
		return
	}

	if archers == nil {
		archers = []model.ArcherRead{}
	}

	_ = writeJSON(w, http.StatusOK, archers)
}

// GetByID godoc
// @Summary     Get Archer
// @Description Retrieve a single archer profile by UUID primary key identifier
// @Tags        Archers
// @Produce     json
// @Param       id path string true "Archer UUID"
// @Success     200 {object} model.ArcherRead
// @Failure     401 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /archer/{id} [get]
func (h *ArcherHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUIDParam(w, r, "id", "archer_id")
	if !ok {
		return
	}

	archer, err := h.archerSvc.GetByID(r.Context(), id)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, archer)
}

// Create godoc
// @Summary     Create Archer
// @Description Parse archer creation payload, persist new profile, and return created identifier
// @Tags        Archers
// @Accept      json
// @Produce     json
// @Param       request body model.ArcherCreate true "Archer creation payload"
// @Success     201 {object} model.ArcherID
// @Failure     401 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /archer/ [post]
func (h *ArcherHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.ArcherCreate
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	id, err := h.archerSvc.Create(r.Context(), req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusCreated, model.ArcherID{ArcherID: id})
}

// Update godoc
// @Summary     Update Archer
// @Description Update archer fields matching the specified where filter
// @Tags        Archers
// @Accept      json
// @Produce     json
// @Param       request body model.ArcherUpdate true "Archer update filter and data"
// @Success     200
// @Failure     401 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /archer/ [patch]
func (h *ArcherHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.ArcherUpdate
	if err := readJSON(r, &req); err != nil {
		writeAppError(w, err)
		return
	}

	if req.Where.ArcherID == nil || *req.Where.ArcherID == uuid.Nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "where.archer_id is required"))
		return
	}

	if err := h.archerSvc.Update(r.Context(), *req.Where.ArcherID, req.Data); err != nil {
		writeAppError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// Delete godoc
// @Summary     Delete Archer
// @Description Remove an archer profile by UUID primary key identifier
// @Tags        Archers
// @Produce     json
// @Param       id path string true "Archer UUID"
// @Success     204
// @Failure     401 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /archer/{id} [delete]
func (h *ArcherHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUIDParam(w, r, "id", "archer_id")
	if !ok {
		return
	}

	if err := h.archerSvc.Delete(r.Context(), id); err != nil {
		writeAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Routes registers all archer CRUD endpoints on the provided chi Router.
func (h *ArcherHandler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Post("/", h.Create)
	r.Patch("/", h.Update)
	r.Delete("/{id}", h.Delete)
}
