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

var sessionColumns = []string{
	"session_id",
	"owner_archer_id",
	"session_location",
	"is_indoor",
	"is_opened",
	"created_at",
	"closed_at",
}

// SessionRepo manages database operations for archery shooting sessions.
type SessionRepo struct {
	db DBTX
}

// NewSessionRepo constructs a SessionRepo backed by DBTX.
func NewSessionRepo(db DBTX) *SessionRepo {
	return &SessionRepo{db: db}
}

// WithTx returns a new SessionRepo bound to the given transaction.
func (r *SessionRepo) WithTx(tx pgx.Tx) *SessionRepo {
	return &SessionRepo{db: tx}
}

func scanSession(scanner interface{ Scan(dest ...any) error }) (model.SessionRead, error) {
	var s model.SessionRead
	err := scanner.Scan(
		&s.SessionID,
		&s.OwnerArcherID,
		&s.SessionLocation,
		&s.IsIndoor,
		&s.IsOpened,
		&s.CreatedAt,
		&s.ClosedAt,
	)
	if err != nil {
		return model.SessionRead{}, err
	}
	return s, nil
}

// FindByID retrieves a shooting session by its primary key identifier.
// Returns nil, nil if no session exists with the given ID.
func (r *SessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
	sql, args, err := StmtBuilder.Select(sessionColumns...).
		From("session").
		Where(squirrel.Eq{"session_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.SessionRead, error) {
		return scanSession(r)
	})
}

// FindOpen finds the currently open session owned by the specified archer.
// Returns nil, nil if the archer has no currently open session.
func (r *SessionRepo) FindOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error) {
	sql, args, err := StmtBuilder.Select(sessionColumns...).
		From("session").
		Where(squirrel.Eq{"owner_archer_id": archerID}).
		Where(squirrel.Eq{"is_opened": true}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find open query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.SessionRead, error) {
		return scanSession(r)
	})
}

// FindAll queries all shooting sessions matching the optional criteria in filter.
//

func (r *SessionRepo) FindAll(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	q := StmtBuilder.Select(sessionColumns...).
		From("session").
		OrderBy("created_at DESC")

	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
	}
	if filter.OwnerArcherID != nil {
		q = q.Where(squirrel.Eq{"owner_archer_id": *filter.OwnerArcherID})
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
	}
	if filter.ClosedAt != nil {
		q = q.Where(squirrel.Eq{"closed_at": *filter.ClosedAt})
	}
	if filter.SessionLocation != nil {
		q = q.Where(squirrel.Eq{"session_location": *filter.SessionLocation})
	}
	if filter.IsOpened != nil {
		q = q.Where(squirrel.Eq{"is_opened": *filter.IsOpened})
	}
	if filter.IsIndoor != nil {
		q = q.Where(squirrel.Eq{"is_indoor": *filter.IsIndoor})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying sessions: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.SessionRead, error) {
		return scanSession(r)
	})
}

// Create inserts a new shooting session record and returns the generated UUID identifier.
//

func (r *SessionRepo) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
	sql, args, err := StmtBuilder.Insert("session").
		Columns(
			"owner_archer_id",
			"session_location",
			"is_indoor",
			"is_opened",
		).
		Values(
			data.OwnerArcherID,
			data.SessionLocation,
			data.IsIndoor,
			data.IsOpened,
		).
		Suffix("RETURNING session_id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create session query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting session: %w", err)
	}

	return newID, nil
}

// Update mutates session fields specified in data for rows matching filter.
// Requires at least one filter criterion to prevent unbounded updates.
// Returns apperror.ErrNotFound if no session matched the filter.
//

func (r *SessionRepo) Update(ctx context.Context, data model.SessionSet, filter model.SessionFilter) error {
	q := StmtBuilder.Update("session")
	setCount := 0

	if data.SessionLocation != nil {
		q = q.Set("session_location", *data.SessionLocation)
		setCount++
	}
	if data.IsIndoor != nil {
		q = q.Set("is_indoor", *data.IsIndoor)
		setCount++
	}
	if data.IsOpened != nil {
		q = q.Set("is_opened", *data.IsOpened)
		setCount++
	}
	if data.ClosedAt != nil {
		q = q.Set("closed_at", *data.ClosedAt)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
		whereCount++
	}
	if filter.OwnerArcherID != nil {
		q = q.Where(squirrel.Eq{"owner_archer_id": *filter.OwnerArcherID})
		whereCount++
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
		whereCount++
	}
	if filter.ClosedAt != nil {
		q = q.Where(squirrel.Eq{"closed_at": *filter.ClosedAt})
		whereCount++
	}
	if filter.SessionLocation != nil {
		q = q.Where(squirrel.Eq{"session_location": *filter.SessionLocation})
		whereCount++
	}
	if filter.IsOpened != nil {
		q = q.Where(squirrel.Eq{"is_opened": *filter.IsOpened})
		whereCount++
	}
	if filter.IsIndoor != nil {
		q = q.Where(squirrel.Eq{"is_indoor": *filter.IsIndoor})
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

// Close marks an active shooting session as closed, setting is_opened = false and closed_at = current UTC timestamp.
// If the session does not exist or is already closed, it returns apperror.ErrNotFound.
func (r *SessionRepo) Close(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	sql, args, err := StmtBuilder.Update("session").
		Set("is_opened", false).
		Set("closed_at", now).
		Where(squirrel.Eq{"session_id": id}).
		Where(squirrel.Eq{"is_opened": true}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building close query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing close session: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// Delete removes a shooting session by its primary key identifier.
// Returns apperror.ErrNotFound if no session existed with the given ID.
func (r *SessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("session").
		Where(squirrel.Eq{"session_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete session query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete session: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// FindParticipating queries whether the specified archer is assigned to an active shooting slot in an open session.
// Returns the session ID pointer if participating, or nil, nil if not found.
func (r *SessionRepo) FindParticipating(ctx context.Context, archerID uuid.UUID) (*uuid.UUID, error) {
	sql, args, err := StmtBuilder.Select("s.session_id").
		From("slot s").
		Join("session ses ON s.session_id = ses.session_id").
		Where(squirrel.Eq{"s.archer_id": archerID}).
		Where(squirrel.Eq{"s.is_shooting": true}).
		Where(squirrel.Eq{"ses.is_opened": true}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find participating query: %w", err)
	}

	var sessionID uuid.UUID
	err = r.db.QueryRow(ctx, sql, args...).Scan(&sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying participating session: %w", err)
	}

	return &sessionID, nil
}
