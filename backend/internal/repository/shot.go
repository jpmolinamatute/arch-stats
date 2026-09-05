package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var shotColumns = []string{
	"shot_id",
	"slot_id",
	"x",
	"y",
	"is_x",
	"score",
	"arrow_id",
	"created_at",
}

// ShotRepo manages database operations for arrow shots within a shooting slot.
type ShotRepo struct {
	db DBTX
}

// NewShotRepo constructs a ShotRepo backed by DBTX.
func NewShotRepo(db DBTX) *ShotRepo {
	return &ShotRepo{db: db}
}

// WithTx returns a new ShotRepo bound to the given transaction.
func (r *ShotRepo) WithTx(tx pgx.Tx) *ShotRepo {
	return &ShotRepo{db: tx}
}

func scanShot(scanner interface{ Scan(dest ...any) error }) (model.ShotRead, error) {
	var s model.ShotRead
	err := scanner.Scan(
		&s.ShotID,
		&s.SlotID,
		&s.X,
		&s.Y,
		&s.IsX,
		&s.Score,
		&s.ArrowID,
		&s.CreatedAt,
	)
	if err != nil {
		return model.ShotRead{}, err
	}
	return s, nil
}

// FindByID retrieves a shot record by its primary key identifier.
// Returns nil, nil if no shot exists with the given ID.
func (r *ShotRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error) {
	sql, args, err := StmtBuilder.Select(shotColumns...).
		From("shot").
		Where(squirrel.Eq{"shot_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find shot by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.ShotRead, error) {
		return scanShot(r)
	})
}

// FindBySlotID retrieves all shot records for a given slot, ordered chronologically.
func (r *ShotRepo) FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error) {
	sql, args, err := StmtBuilder.Select(shotColumns...).
		From("shot").
		Where(squirrel.Eq{"slot_id": slotID}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find shots by slot id query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying shots by slot id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ShotRead, error) {
		return scanShot(r)
	})
}

// FindAll queries all shots matching the optional criteria in filter, ordered chronologically.
func (r *ShotRepo) FindAll(ctx context.Context, filter model.ShotFilter) ([]model.ShotRead, error) {
	q := StmtBuilder.Select(shotColumns...).
		From("shot").
		OrderBy("created_at ASC")

	if filter.ShotID != nil {
		q = q.Where(squirrel.Eq{"shot_id": *filter.ShotID})
	}
	if filter.SlotID != nil {
		q = q.Where(squirrel.Eq{"slot_id": *filter.SlotID})
	}
	if filter.X != nil {
		q = q.Where(squirrel.Eq{"x": *filter.X})
	}
	if filter.Y != nil {
		q = q.Where(squirrel.Eq{"y": *filter.Y})
	}
	if filter.Score != nil {
		q = q.Where(squirrel.Eq{"score": *filter.Score})
	}
	if filter.ArrowID != nil {
		q = q.Where(squirrel.Eq{"arrow_id": *filter.ArrowID})
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all shots query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying shots: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ShotRead, error) {
		return scanShot(r)
	})
}
