package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var authIdentityColumns = []string{
	"archer_id",
	"google_subject",
	"google_picture_url",
	"last_login_at",
	"created_at",
}

// AuthIdentityRepo manages database operations for OAuth authentication credentials.
type AuthIdentityRepo struct {
	db DBTX
}

// NewAuthIdentityRepo constructs an AuthIdentityRepo backed by DBTX.
func NewAuthIdentityRepo(db DBTX) *AuthIdentityRepo {
	return &AuthIdentityRepo{db: db}
}

// WithTx returns a new AuthIdentityRepo bound to the given transaction.
func (r *AuthIdentityRepo) WithTx(tx pgx.Tx) *AuthIdentityRepo {
	return &AuthIdentityRepo{db: tx}
}

func scanAuthIdentity(scanner interface{ Scan(dest ...any) error }) (model.AuthIdentityRead, error) {
	var a model.AuthIdentityRead
	err := scanner.Scan(
		&a.ArcherID,
		&a.GoogleSubject,
		&a.GooglePictureURL,
		&a.LastLoginAt,
		&a.CreatedAt,
	)
	if err != nil {
		return model.AuthIdentityRead{}, err
	}
	return a, nil
}

// FindByGoogleSubject retrieves an auth record by Google account subject claim.
func (r *AuthIdentityRepo) FindByGoogleSubject(ctx context.Context, sub string) (*model.AuthIdentityRead, error) {
	sql, args, err := StmtBuilder.Select(authIdentityColumns...).
		From("auth").
		Where(squirrel.Eq{"google_subject": sub}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find auth by google subject query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.AuthIdentityRead, error) {
		return scanAuthIdentity(r)
	})
}

// FindByID retrieves an auth identity by primary key identifier.
func (r *AuthIdentityRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.AuthIdentityRead, error) {
	return findByID(ctx, r.db, "auth", "archer_id", authIdentityColumns, id, func(row pgx.Row) (model.AuthIdentityRead, error) {
		return scanAuthIdentity(row)
	})
}

// Create inserts a new auth identity credential record returning the generated archer_id UUID.
func (r *AuthIdentityRepo) Create(ctx context.Context, subject string, picture *string) (uuid.UUID, error) {
	cols := []string{"google_subject"}
	vals := []any{subject}
	if picture != nil {
		cols = append(cols, "google_picture_url")
		vals = append(vals, *picture)
	}

	builder := StmtBuilder.Insert("auth").
		Columns(cols...).
		Values(vals...)

	return createReturningID(ctx, r.db, builder, "archer_id")
}

// UpdateLastLogin updates the timestamp of most recent sign-in and optionally refreshes the avatar URL.
func (r *AuthIdentityRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, login time.Time, pic *string) error {
	q := StmtBuilder.Update("auth").
		Set("last_login_at", login)
	if pic != nil {
		q = q.Set("google_picture_url", *pic)
	}
	q = q.Where(squirrel.Eq{"archer_id": id})
	return execUpdate(ctx, r.db, q)
}
