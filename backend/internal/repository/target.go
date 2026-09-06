package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var targetColumns = []string{
	"target_id",
	"session_id",
	"distance",
	"lane",
	"created_at",
}

// TargetRepo manages database operations for lane target configurations.
type TargetRepo struct {
	db DBTX
}

// NewTargetRepo constructs a TargetRepo backed by DBTX.
func NewTargetRepo(db DBTX) *TargetRepo {
	return &TargetRepo{db: db}
}

// WithTx returns a new TargetRepo bound to the given transaction.
func (r *TargetRepo) WithTx(tx pgx.Tx) *TargetRepo {
	return &TargetRepo{db: tx}
}

func scanTarget(scanner interface{ Scan(dest ...any) error }) (model.TargetRead, error) {
	var t model.TargetRead
	err := scanner.Scan(
		&t.TargetID,
		&t.SessionID,
		&t.Distance,
		&t.Lane,
		&t.CreatedAt,
	)
	if err != nil {
		return model.TargetRead{}, err
	}
	return t, nil
}

// FindByID retrieves a target configuration by primary key identifier.
// Returns nil, nil if no target exists with the given ID.
func (r *TargetRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
	return findByID(ctx, r.db, "target", "target_id", targetColumns, id, func(row pgx.Row) (model.TargetRead, error) {
		return scanTarget(row)
	})
}

// FindBySlotID retrieves target configurations associated with a specific slot.
func (r *TargetRepo) FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error) {
	cols := make([]string, len(targetColumns))
	for i, c := range targetColumns {
		cols[i] = "target." + c
	}

	sql, args, err := StmtBuilder.Select(cols...).
		From("target").
		Join("slot ON target.target_id = slot.target_id").
		Where(squirrel.Eq{"slot.slot_id": slotID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find targets by slot id query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying targets by slot id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.TargetRead, error) {
		return scanTarget(r)
	})
}

// FindBySessionID retrieves all target configurations for a given session, ordered by lane ascending.
func (r *TargetRepo) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error) {
	sql, args, err := StmtBuilder.Select(targetColumns...).
		From("target").
		Where(squirrel.Eq{"session_id": sessionID}).
		OrderBy("lane ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find targets by session id query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying targets by session id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.TargetRead, error) {
		return scanTarget(r)
	})
}

// Create inserts a new target configuration record and returns the generated UUID identifier.
func (r *TargetRepo) Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error) {
	builder := StmtBuilder.Insert("target").
		Columns("session_id", "distance", "lane").
		Values(data.SessionID, data.Distance, data.Lane)

	return createReturningID(ctx, r.db, builder, "target_id")
}

// Update mutates target fields specified in data for rows matching filter.
// Requires at least one filter criterion to prevent unrestricted updates.
// Returns apperror.ErrNotFound if no target matched the filter.
func (r *TargetRepo) Update(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error {
	q := StmtBuilder.Update("target")
	setCount := 0

	if data.Distance != nil {
		q = q.Set("distance", *data.Distance)
		setCount++
	}
	if data.Lane != nil {
		q = q.Set("lane", *data.Lane)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.TargetID != nil {
		q = q.Where(squirrel.Eq{"target_id": *filter.TargetID})
		whereCount++
	}
	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
		whereCount++
	}
	if filter.Distance != nil {
		q = q.Where(squirrel.Eq{"distance": *filter.Distance})
		whereCount++
	}
	if filter.Lane != nil {
		q = q.Where(squirrel.Eq{"lane": *filter.Lane})
		whereCount++
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
		whereCount++
	}

	if whereCount == 0 {
		return errUpdateRequiresFilter
	}

	return execUpdate(ctx, r.db, q)
}

// Delete removes a target configuration by primary key identifier.
// Returns apperror.ErrNotFound if no target existed with the given ID.
func (r *TargetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return deleteByID(ctx, r.db, "target", "target_id", id)
}
