package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
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
	sql, args, err := StmtBuilder.Select(targetColumns...).
		From("target").
		Where(squirrel.Eq{"target_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find target by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.TargetRead, error) {
		return scanTarget(r)
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
	sql, args, err := StmtBuilder.Insert("target").
		Columns("session_id", "distance", "lane").
		Values(data.SessionID, data.Distance, data.Lane).
		Suffix("RETURNING target_id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create target query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting target: %w", err)
	}

	return newID, nil
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
		return errors.New("at least one filter criterion is required for update")
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("building update target query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing update target: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// Delete removes a target configuration by primary key identifier.
// Returns apperror.ErrNotFound if no target existed with the given ID.
func (r *TargetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("target").
		Where(squirrel.Eq{"target_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete target query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete target: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
