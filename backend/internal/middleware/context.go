package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

type contextKey string

const archerIDContextKey contextKey = "arch_stats_archer_id"

// WithArcherID injects the authenticated archer UUID into the request context.
func WithArcherID(ctx context.Context, archerID uuid.UUID) context.Context {
	return context.WithValue(ctx, archerIDContextKey, archerID)
}

// GetArcherID extracts the authenticated archer UUID from the request context.
// Returns apperror.ErrUnauthorized if the ID is missing or is uuid.Nil.
func GetArcherID(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(archerIDContextKey)
	if val == nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "archer id not found in context")
	}

	id, ok := val.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid archer id in context")
	}

	return id, nil
}
