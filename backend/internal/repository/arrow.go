package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var arrowColumns = []string{
	"arrow_id",
	"archer_id",
	"arrow_set",
	"arrow_number",
	"status",
	"spine",
	"length",
	"weight",
	"is_deleted",
	"created_at",
}

// ArrowRepo manages database operations for arrow equipment inventory.
type ArrowRepo struct {
	db DBTX
}

// NewArrowRepo constructs an ArrowRepo backed by DBTX.
func NewArrowRepo(db DBTX) *ArrowRepo {
	return &ArrowRepo{db: db}
}

// WithTx returns a new ArrowRepo bound to the given transaction.
func (r *ArrowRepo) WithTx(tx pgx.Tx) *ArrowRepo {
	return &ArrowRepo{db: tx}
}

func scanArrow(scanner interface{ Scan(dest ...any) error }) (model.ArrowRead, error) {
	var a model.ArrowRead
	err := scanner.Scan(
		&a.ArrowID,
		&a.ArcherID,
		&a.ArrowSet,
		&a.ArrowNumber,
		&a.Status,
		&a.Spine,
		&a.Length,
		&a.Weight,
		&a.IsDeleted,
		&a.CreatedAt,
	)
	if err != nil {
		return model.ArrowRead{}, err
	}
	return a, nil
}

// Create inserts a single arrow record, defaulting status to 'in_use' if omitted.
func (r *ArrowRepo) Create(ctx context.Context, data model.ArrowCreate) (uuid.UUID, error) {
	status := model.ArrowStatusInUse
	if data.Status != nil {
		status = *data.Status
	}

	cols := []string{"archer_id", "arrow_set", "arrow_number", "status", "spine", "length", "weight"}
	vals := []any{data.ArcherID, data.ArrowSet, data.ArrowNumber, status, data.Spine, data.Length, data.Weight}

	builder := StmtBuilder.Insert("arrow").
		Columns(cols...).
		Values(vals...)

	return createReturningID(ctx, r.db, builder, "arrow_id")
}

// CreateBatch inserts multiple arrow records in a single query, defaulting status to 'in_use'.
func (r *ArrowRepo) CreateBatch(ctx context.Context, data model.ArrowBatchCreate) ([]model.ArrowRead, error) {
	if data.Count <= 0 {
		return nil, nil
	}

	status := model.ArrowStatusInUse
	if data.Status != nil {
		status = *data.Status
	}

	cols := []string{"archer_id", "arrow_set", "arrow_number", "status", "spine", "length", "weight"}
	builder := StmtBuilder.Insert("arrow").Columns(cols...)

	for i := 1; i <= data.Count; i++ {
		builder = builder.Values(
			data.ArcherID,
			data.ArrowSet,
			int16(i),
			status,
			data.Spine,
			data.Length,
			data.Weight,
		)
	}

	builder = builder.Suffix("RETURNING " + strings.Join(arrowColumns, ", "))
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building create batch arrow query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("inserting arrow batch: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ArrowRead, error) {
		return scanArrow(r)
	})
}

// FindByID retrieves an arrow by primary key identifier.
func (r *ArrowRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ArrowRead, error) {
	return findByID(ctx, r.db, "arrow", "arrow_id", arrowColumns, id, func(row pgx.Row) (model.ArrowRead, error) {
		return scanArrow(row)
	})
}

// FindAllByArcherID retrieves all active arrows for an archer ordered by set and number.
func (r *ArrowRepo) FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.ArrowRead, error) {
	sql, args, err := StmtBuilder.Select(arrowColumns...).
		From("arrow").
		Where(squirrel.Eq{"archer_id": archerID}).
		Where(squirrel.Eq{"is_deleted": false}).
		OrderBy("arrow_set ASC, arrow_number ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all arrows query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying arrows by archer id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ArrowRead, error) {
		return scanArrow(r)
	})
}

// CountByArcherID counts all active arrows registered by an archer.
func (r *ArrowRepo) CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("arrow").
		Where(squirrel.Eq{"archer_id": archerID}).
		Where(squirrel.Eq{"is_deleted": false}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count arrows query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting arrows by archer id: %w", err)
	}

	return count, nil
}

// CountInUseByArcherID counts active arrows with status 'in_use' registered by an archer.
func (r *ArrowRepo) CountInUseByArcherID(ctx context.Context, archerID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("arrow").
		Where(squirrel.Eq{"archer_id": archerID}).
		Where(squirrel.Eq{"status": model.ArrowStatusInUse}).
		Where(squirrel.Eq{"is_deleted": false}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count in use arrows query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting in use arrows: %w", err)
	}

	return count, nil
}

// FindByArcherAndSet retrieves all active arrows in a specific set for an archer ordered by arrow number.
func (r *ArrowRepo) FindByArcherAndSet(ctx context.Context, archerID uuid.UUID, set int16) ([]model.ArrowRead, error) {
	sql, args, err := StmtBuilder.Select(arrowColumns...).
		From("arrow").
		Where(squirrel.Eq{"archer_id": archerID}).
		Where(squirrel.Eq{"arrow_set": set}).
		Where(squirrel.Eq{"is_deleted": false}).
		OrderBy("arrow_number ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find arrows by set query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying arrows by set: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ArrowRead, error) {
		return scanArrow(r)
	})
}
