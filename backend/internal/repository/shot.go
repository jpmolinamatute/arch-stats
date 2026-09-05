package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
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

// Create inserts a new shot record and returns the generated UUID identifier.
func (r *ShotRepo) Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error) {
	insertBuilder := StmtBuilder.Insert("shot")
	if data.CreatedAt != nil {
		insertBuilder = insertBuilder.
			Columns("slot_id", "x", "y", "is_x", "score", "arrow_id", "created_at").
			Values(data.SlotID, data.X, data.Y, data.IsX, data.Score, data.ArrowID, *data.CreatedAt)
	} else {
		insertBuilder = insertBuilder.
			Columns("slot_id", "x", "y", "is_x", "score", "arrow_id").
			Values(data.SlotID, data.X, data.Y, data.IsX, data.Score, data.ArrowID)
	}

	sql, args, err := insertBuilder.Suffix("RETURNING shot_id").ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create shot query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting shot: %w", err)
	}

	return newID, nil
}

// CreateBatch inserts a batch of shot records in a single query and returns their generated UUIDs.
func (r *ShotRepo) CreateBatch(ctx context.Context, data []model.ShotCreate) ([]uuid.UUID, error) {
	if len(data) == 0 {
		return nil, nil
	}

	b := StmtBuilder.Insert("shot").
		Columns("slot_id", "x", "y", "is_x", "score", "arrow_id", "created_at")

	for _, s := range data {
		var createdAt any
		if s.CreatedAt != nil {
			createdAt = *s.CreatedAt
		} else {
			createdAt = squirrel.Expr("now()")
		}
		b = b.Values(s.SlotID, s.X, s.Y, s.IsX, s.Score, s.ArrowID, createdAt)
	}

	sql, args, err := b.Suffix("RETURNING shot_id").ToSql()
	if err != nil {
		return nil, fmt.Errorf("building batch create shot query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("executing batch create shot: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (uuid.UUID, error) {
		var id uuid.UUID
		err := r.Scan(&id)
		return id, err
	})
}

// Update mutates shot fields specified in data for rows matching filter.
// Requires at least one filter criterion to prevent unrestricted updates.
// Returns apperror.ErrNotFound if no shot matched the filter.
func (r *ShotRepo) Update(ctx context.Context, data model.ShotSet, filter model.ShotFilter) error {
	q := StmtBuilder.Update("shot")
	setCount := 0

	if data.X != nil {
		q = q.Set("x", *data.X)
		setCount++
	}
	if data.Y != nil {
		q = q.Set("y", *data.Y)
		setCount++
	}
	if data.IsX != nil {
		q = q.Set("is_x", *data.IsX)
		setCount++
	}
	if data.Score != nil {
		q = q.Set("score", *data.Score)
		setCount++
	}
	if data.ArrowID != nil {
		q = q.Set("arrow_id", *data.ArrowID)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.ShotID != nil {
		q = q.Where(squirrel.Eq{"shot_id": *filter.ShotID})
		whereCount++
	}
	if filter.SlotID != nil {
		q = q.Where(squirrel.Eq{"slot_id": *filter.SlotID})
		whereCount++
	}
	if filter.X != nil {
		q = q.Where(squirrel.Eq{"x": *filter.X})
		whereCount++
	}
	if filter.Y != nil {
		q = q.Where(squirrel.Eq{"y": *filter.Y})
		whereCount++
	}
	if filter.Score != nil {
		q = q.Where(squirrel.Eq{"score": *filter.Score})
		whereCount++
	}
	if filter.ArrowID != nil {
		q = q.Where(squirrel.Eq{"arrow_id": *filter.ArrowID})
		whereCount++
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
		whereCount++
	}

	if whereCount == 0 {
		return errors.New("update requires at least one filter condition to prevent unrestricted updates")
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("building update shot query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing update shot: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// Delete removes a shot record by its primary key identifier.
// Returns apperror.ErrNotFound if no shot existed with the given ID.
func (r *ShotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("shot").
		Where(squirrel.Eq{"shot_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete shot query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete shot: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// CountBySlotID counts the total number of shots recorded for a given slot.
func (r *ShotRepo) CountBySlotID(ctx context.Context, slotID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("shot").
		Where(squirrel.Eq{"slot_id": slotID}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count shots by slot id query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting shots by slot id: %w", err)
	}

	return count, nil
}

// GetLatestShotTime retrieves the timestamp of the most recently recorded shot for a given slot.
// Returns nil, nil if no shots exist for the slot.
func (r *ShotRepo) GetLatestShotTime(ctx context.Context, slotID uuid.UUID) (*time.Time, error) {
	sql, args, err := StmtBuilder.Select("created_at").
		From("shot").
		Where(squirrel.Eq{"slot_id": slotID}).
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building get latest shot time query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (time.Time, error) {
		var t time.Time
		err := r.Scan(&t)
		return t, err
	})
}
