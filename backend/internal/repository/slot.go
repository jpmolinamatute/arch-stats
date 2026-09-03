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

var slotColumns = []string{
	"slot_id",
	"target_id",
	"archer_id",
	"session_id",
	"slot_letter",
	"face_type",
	"bowstyle",
	"draw_weight",
	"club_id",
	"is_shooting",
	"shot_per_round",
	"interval_seconds",
	"created_at",
}

// SlotRepo manages database operations for shooting slot assignments within a session.
type SlotRepo struct {
	db DBTX
}

// NewSlotRepo constructs a SlotRepo backed by DBTX.
func NewSlotRepo(db DBTX) *SlotRepo {
	return &SlotRepo{db: db}
}

// WithTx returns a new SlotRepo bound to the given transaction.
func (r *SlotRepo) WithTx(tx pgx.Tx) *SlotRepo {
	return &SlotRepo{db: tx}
}

func scanSlot(scanner interface{ Scan(dest ...any) error }) (model.SlotRead, error) {
	var s model.SlotRead
	err := scanner.Scan(
		&s.SlotID,
		&s.TargetID,
		&s.ArcherID,
		&s.SessionID,
		&s.SlotLetter,
		&s.FaceType,
		&s.Bowstyle,
		&s.DrawWeight,
		&s.ClubID,
		&s.IsShooting,
		&s.ShotPerRound,
		&s.IntervalSeconds,
		&s.CreatedAt,
	)
	if err != nil {
		return model.SlotRead{}, err
	}
	return s, nil
}

// FindByID retrieves a slot assignment by primary key identifier.
// Returns nil, nil if no slot exists with the given ID.
func (r *SlotRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error) {
	sql, args, err := StmtBuilder.Select(slotColumns...).
		From("slot").
		Where(squirrel.Eq{"slot_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find slot by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.SlotRead, error) {
		return scanSlot(r)
	})
}

// FindBySessionID retrieves all slot assignments for a given session, ordered chronologically then by slot letter.
func (r *SlotRepo) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error) {
	sql, args, err := StmtBuilder.Select(slotColumns...).
		From("slot").
		Where(squirrel.Eq{"session_id": sessionID}).
		OrderBy("created_at ASC, slot_letter ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find slots by session id query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying slots by session id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.SlotRead, error) {
		return scanSlot(r)
	})
}

// FindAll queries all shooting slots matching the optional criteria in filter.
func (r *SlotRepo) FindAll(ctx context.Context, filter model.SlotFilter) ([]model.SlotRead, error) {
	q := StmtBuilder.Select(slotColumns...).
		From("slot").
		OrderBy("created_at ASC, slot_letter ASC")

	if filter.SlotID != nil {
		q = q.Where(squirrel.Eq{"slot_id": *filter.SlotID})
	}
	if filter.TargetID != nil {
		q = q.Where(squirrel.Eq{"target_id": *filter.TargetID})
	}
	if filter.ArcherID != nil {
		q = q.Where(squirrel.Eq{"archer_id": *filter.ArcherID})
	}
	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
	}
	if filter.SlotLetter != nil {
		q = q.Where(squirrel.Eq{"slot_letter": *filter.SlotLetter})
	}
	if filter.IsShooting != nil {
		q = q.Where(squirrel.Eq{"is_shooting": *filter.IsShooting})
	}
	if filter.ShotPerRound != nil {
		q = q.Where(squirrel.Eq{"shot_per_round": *filter.ShotPerRound})
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all slots query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying slots: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.SlotRead, error) {
		return scanSlot(r)
	})
}

// Create inserts a new slot assignment record and returns the generated UUID identifier.
//
//nolint:gocritic // hugeParam: data value parameter matches repository interface specification
func (r *SlotRepo) Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error) {
	sql, args, err := StmtBuilder.Insert("slot").
		Columns(
			"target_id",
			"archer_id",
			"session_id",
			"slot_letter",
			"face_type",
			"bowstyle",
			"draw_weight",
			"club_id",
			"is_shooting",
			"shot_per_round",
			"interval_seconds",
		).
		Values(
			data.TargetID,
			data.ArcherID,
			data.SessionID,
			data.SlotLetter,
			data.FaceType,
			data.Bowstyle,
			data.DrawWeight,
			data.ClubID,
			data.IsShooting,
			data.ShotPerRound,
			data.IntervalSeconds,
		).
		Suffix("RETURNING slot_id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create slot query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting slot: %w", err)
	}

	return newID, nil
}

// Update mutates slot assignment fields specified in data for rows matching filter.
// Requires at least one filter criterion to prevent unrestricted updates.
// Returns apperror.ErrNotFound if no slot assignment matched the filter.
func (r *SlotRepo) Update(ctx context.Context, data model.SlotSet, filter model.SlotFilter) error {
	q := StmtBuilder.Update("slot")
	setCount := 0

	if data.IsShooting != nil {
		q = q.Set("is_shooting", *data.IsShooting)
		setCount++
	}
	if data.FaceType != nil {
		q = q.Set("face_type", *data.FaceType)
		setCount++
	}
	if data.SlotLetter != nil {
		q = q.Set("slot_letter", *data.SlotLetter)
		setCount++
	}
	if data.ShotPerRound != nil {
		q = q.Set("shot_per_round", *data.ShotPerRound)
		setCount++
	}
	if data.IntervalSeconds != nil {
		q = q.Set("interval_seconds", *data.IntervalSeconds)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.SlotID != nil {
		q = q.Where(squirrel.Eq{"slot_id": *filter.SlotID})
		whereCount++
	}
	if filter.TargetID != nil {
		q = q.Where(squirrel.Eq{"target_id": *filter.TargetID})
		whereCount++
	}
	if filter.ArcherID != nil {
		q = q.Where(squirrel.Eq{"archer_id": *filter.ArcherID})
		whereCount++
	}
	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
		whereCount++
	}
	if filter.SlotLetter != nil {
		q = q.Where(squirrel.Eq{"slot_letter": *filter.SlotLetter})
		whereCount++
	}
	if filter.IsShooting != nil {
		q = q.Where(squirrel.Eq{"is_shooting": *filter.IsShooting})
		whereCount++
	}
	if filter.ShotPerRound != nil {
		q = q.Where(squirrel.Eq{"shot_per_round": *filter.ShotPerRound})
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
		return fmt.Errorf("building update query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing update: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// Delete removes a slot assignment by primary key identifier.
// Returns apperror.ErrNotFound if no slot assignment existed with the given ID.
func (r *SlotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("slot").
		Where(squirrel.Eq{"slot_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete slot query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete slot: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// CountBySessionID counts the total number of slot assignments for a given session.
func (r *SlotRepo) CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("slot").
		Where(squirrel.Eq{"session_id": sessionID}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count by session id query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting slots by session id: %w", err)
	}

	return count, nil
}
