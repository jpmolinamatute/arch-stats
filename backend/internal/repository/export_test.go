package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrUpdateRequiresFilter exposes errUpdateRequiresFilter for testing.
var ErrUpdateRequiresFilter = errUpdateRequiresFilter

// DeleteByID exposes deleteByID for testing.
func DeleteByID(ctx context.Context, db DBTX, table, pkColumn string, id uuid.UUID) error {
	return deleteByID(ctx, db, table, pkColumn, id)
}

// ExecUpdate exposes execUpdate for testing.
func ExecUpdate(ctx context.Context, db DBTX, q squirrel.UpdateBuilder) error {
	return execUpdate(ctx, db, q)
}

// FindByID exposes findByID for testing.
func FindByID[T any](
	ctx context.Context,
	db DBTX,
	table string,
	pkColumn string,
	columns []string,
	id uuid.UUID,
	scanFn func(pgx.Row) (T, error),
) (*T, error) {
	return findByID(ctx, db, table, pkColumn, columns, id, scanFn)
}

// CreateReturningID exposes createReturningID for testing.
func CreateReturningID(
	ctx context.Context,
	db DBTX,
	builder squirrel.InsertBuilder,
	pkColumn string,
) (uuid.UUID, error) {
	return createReturningID(ctx, db, builder, pkColumn)
}
