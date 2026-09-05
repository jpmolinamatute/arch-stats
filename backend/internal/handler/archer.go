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

// List handles GET /api/v0/archer/.
// It queries archers matching default filter criteria and returns a JSON list.
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

// GetByID handles GET /api/v0/archer/{id}.
// It retrieves a single archer profile by primary key identifier.
func (h *ArcherHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := getURLParam(r, "id")
	if idStr == "" {
		idStr = getURLParam(r, "archer_id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer id is required"))
		return
	}

	archer, err := h.archerSvc.GetByID(r.Context(), id)
	if err != nil {
		writeAppError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, archer)
}

// Create handles POST /api/v0/archer/.
// It parses the archer creation payload, validates and persists the new archer, and returns 201 Created.
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

// Update handles PATCH /api/v0/archer/.
// It updates archer fields matching the specified where filter.
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

// Delete handles DELETE /api/v0/archer/{id}.
// It removes an archer profile by primary key identifier and returns 204 No Content.
func (h *ArcherHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := getURLParam(r, "id")
	if idStr == "" {
		idStr = getURLParam(r, "archer_id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "valid archer id is required"))
		return
	}

	if err := h.archerSvc.Delete(r.Context(), id); err != nil {
		writeAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getURLParam(r *http.Request, key string) string {
	if val := chi.URLParam(r, key); val != "" {
		return val
	}
	return r.PathValue(key)
}
