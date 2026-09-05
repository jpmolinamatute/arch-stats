package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

//nolint:unused // getURLParam is used by GetByID and Delete
func getURLParam(r *http.Request, key string) string {
	if val := chi.URLParam(r, key); val != "" {
		return val
	}
	return r.PathValue(key)
}
