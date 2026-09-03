package repository

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var authSessionColumns = []string{
	"auth_id",
	"archer_id",
	"session_token_hash",
	"created_at",
	"expires_at",
	"revoked_at",
	"ua",
	"ip_inet",
}

// AuthSessionRepo manages database operations for authentication sessions.
type AuthSessionRepo struct {
	db DBTX
}

// NewAuthSessionRepo constructs an AuthSessionRepo backed by DBTX.
func NewAuthSessionRepo(db DBTX) *AuthSessionRepo {
	return &AuthSessionRepo{db: db}
}

// WithTx returns a new AuthSessionRepo bound to the given transaction.
func (r *AuthSessionRepo) WithTx(tx pgx.Tx) *AuthSessionRepo {
	return &AuthSessionRepo{db: tx}
}

func scanAuthSession(scanner interface{ Scan(dest ...any) error }) (model.AuthSessionRead, error) {
	var (
		s     model.AuthSessionRead
		ipRaw any
	)

	err := scanner.Scan(
		&s.AuthID,
		&s.ArcherID,
		&s.SessionTokenHash,
		&s.CreatedAt,
		&s.ExpiresAt,
		&s.RevokedAt,
		&s.UA,
		&ipRaw,
	)
	if err != nil {
		return model.AuthSessionRead{}, err
	}

	if ipRaw != nil {
		switch v := ipRaw.(type) {
		case string:
			s.IPInet = &v
		case *string:
			s.IPInet = v
		case netip.Prefix:
			str := v.Addr().String()
			s.IPInet = &str
		case *netip.Prefix:
			if v != nil {
				str := v.Addr().String()
				s.IPInet = &str
			}
		case netip.Addr:
			str := v.String()
			s.IPInet = &str
		case *netip.Addr:
			if v != nil {
				str := v.String()
				s.IPInet = &str
			}
		case fmt.Stringer:
			str := v.String()
			s.IPInet = &str
		default:
			str := fmt.Sprintf("%v", ipRaw)
			s.IPInet = &str
		}
	}

	return s, nil
}

// Create inserts a new auth session record.
//
//nolint:gocritic // hugeParam: data value parameter matches repository interface specification
func (r *AuthSessionRepo) Create(ctx context.Context, data model.AuthSessionCreate) error {
	createdAt := data.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	sql, args, err := StmtBuilder.Insert("auth").
		Columns(
			"archer_id",
			"session_token_hash",
			"created_at",
			"expires_at",
			"ua",
			"ip_inet",
		).
		Values(
			data.ArcherID,
			data.SessionTokenHash,
			createdAt,
			data.ExpiresAt,
			data.UA,
			data.IPInet,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("building create auth session query: %w", err)
	}

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("inserting auth session: %w", err)
	}

	return nil
}

// FindByTokenHash retrieves an auth session by its token hash.
// Returns nil, nil if no session exists for the given hash.
func (r *AuthSessionRepo) FindByTokenHash(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
	sql, args, err := StmtBuilder.Select(authSessionColumns...).
		From("auth").
		Where(squirrel.Eq{"session_token_hash": hash}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by token hash query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.AuthSessionRead, error) {
		return scanAuthSession(r)
	})
}

// DeleteByArcherID removes all auth sessions belonging to the specified archer (e.g. logout all sessions).
// Operation is idempotent and returns nil if no sessions exist.
func (r *AuthSessionRepo) DeleteByArcherID(ctx context.Context, archerID uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("auth").
		Where(squirrel.Eq{"archer_id": archerID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete by archer id query: %w", err)
	}

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("deleting auth sessions by archer id: %w", err)
	}

	return nil
}

// DeleteExpired removes all sessions whose expires_at is prior to the current time.
// Returns the count of deleted session records.
func (r *AuthSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	sql, args, err := StmtBuilder.Delete("auth").
		Where(squirrel.Lt{"expires_at": time.Now().UTC()}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building delete expired query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("executing delete expired auth sessions: %w", err)
	}

	return tag.RowsAffected(), nil
}

// RevokeByTokenHash marks an active session as revoked. Returns apperror.ErrNotFound
// if no matching unrevoked session exists.
func (r *AuthSessionRepo) RevokeByTokenHash(ctx context.Context, hash []byte, revokedAt time.Time) error {
	sql, args, err := StmtBuilder.Update("auth").
		Set("revoked_at", revokedAt).
		Where(squirrel.Eq{"session_token_hash": hash}).
		Where(squirrel.Eq{"revoked_at": nil}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building revoke by token hash query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing revoke by token hash: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// DeleteByTokenHash removes an auth session identified by its token hash.
// If no session was deleted, returns apperror.ErrNotFound.
func (r *AuthSessionRepo) DeleteByTokenHash(ctx context.Context, hash []byte) error {
	sql, args, err := StmtBuilder.Delete("auth").
		Where(squirrel.Eq{"session_token_hash": hash}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete by token hash query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete by token hash: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
