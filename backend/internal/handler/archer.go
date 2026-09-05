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

func getURLParam(r *http.Request, key string) string {
	if val := chi.URLParam(r, key); val != "" {
		return val
	}
	return r.PathValue(key)
}
