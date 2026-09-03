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

var archerColumns = []string{
	"archer_id",
	"first_name",
	"last_name",
	"email",
	"date_of_birth",
	"gender",
	"bowstyle",
	"draw_weight",
	"club_id",
	"google_picture_url",
	"google_subject",
	"last_login_at",
	"created_at",
}

// ArcherRepo manages database operations for archer profiles.
type ArcherRepo struct {
	db DBTX
}

// NewArcherRepo constructs an ArcherRepo backed by DBTX.
func NewArcherRepo(db DBTX) *ArcherRepo {
	return &ArcherRepo{db: db}
}

// WithTx returns a new ArcherRepo bound to the given transaction.
func (r *ArcherRepo) WithTx(tx pgx.Tx) *ArcherRepo {
	return &ArcherRepo{db: tx}
}

func scanArcher(scanner interface{ Scan(dest ...any) error }) (model.ArcherRead, error) {
	var (
		a      model.ArcherRead
		dobRaw any
	)

	err := scanner.Scan(
		&a.ArcherID,
		&a.FirstName,
		&a.LastName,
		&a.Email,
		&dobRaw,
		&a.Gender,
		&a.Bowstyle,
		&a.DrawWeight,
		&a.ClubID,
		&a.GooglePictureURL,
		&a.GoogleSubject,
		&a.LastLoginAt,
		&a.CreatedAt,
	)
	if err != nil {
		return model.ArcherRead{}, err
	}

	switch v := dobRaw.(type) {
	case time.Time:
		a.DateOfBirth = v.Format("2006-01-02")
	case string:
		a.DateOfBirth = v
	default:
		return model.ArcherRead{}, fmt.Errorf("unexpected type for date_of_birth: %T", dobRaw)
	}

	return a, nil
}

// FindByID retrieves an archer by primary key identifier.
func (r *ArcherRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	sql, args, err := StmtBuilder.Select(archerColumns...).
		From("archer").
		Where(squirrel.Eq{"archer_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.ArcherRead, error) {
		return scanArcher(r)
	})
}

// FindByEmail retrieves an archer by email address.
func (r *ArcherRepo) FindByEmail(ctx context.Context, email string) (*model.ArcherRead, error) {
	sql, args, err := StmtBuilder.Select(archerColumns...).
		From("archer").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by email query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.ArcherRead, error) {
		return scanArcher(r)
	})
}

// FindByGoogleSubject retrieves an archer by OAuth Google Subject identifier.
func (r *ArcherRepo) FindByGoogleSubject(ctx context.Context, sub string) (*model.ArcherRead, error) {
	sql, args, err := StmtBuilder.Select(archerColumns...).
		From("archer").
		Where(squirrel.Eq{"google_subject": sub}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by google subject query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.ArcherRead, error) {
		return scanArcher(r)
	})
}

// FindAll queries all archers matching the optional criteria in filter.
//
//nolint:gocritic // hugeParam: filter value parameter matches repository interface specification
func (r *ArcherRepo) FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
	q := StmtBuilder.Select(archerColumns...).
		From("archer").
		OrderBy("created_at DESC")

	if filter.ArcherID != nil {
		q = q.Where(squirrel.Eq{"archer_id": *filter.ArcherID})
	}
	if filter.FirstName != nil {
		q = q.Where(squirrel.Eq{"first_name": *filter.FirstName})
	}
	if filter.LastName != nil {
		q = q.Where(squirrel.Eq{"last_name": *filter.LastName})
	}
	if filter.Gender != nil {
		q = q.Where(squirrel.Eq{"gender": *filter.Gender})
	}
	if filter.Bowstyle != nil {
		q = q.Where(squirrel.Eq{"bowstyle": *filter.Bowstyle})
	}
	if filter.DrawWeight != nil {
		q = q.Where(squirrel.Eq{"draw_weight": *filter.DrawWeight})
	}
	if filter.ClubID != nil {
		q = q.Where(squirrel.Eq{"club_id": *filter.ClubID})
	}
	if filter.GoogleSubject != nil {
		q = q.Where(squirrel.Eq{"google_subject": *filter.GoogleSubject})
	}
	if filter.LastLoginAt != nil {
		q = q.Where(squirrel.Eq{"last_login_at": *filter.LastLoginAt})
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying archers: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ArcherRead, error) {
		return scanArcher(r)
	})
}

// Create inserts a new archer row, returning the generated UUID.
//
//nolint:gocritic // hugeParam: data value parameter matches repository interface specification
func (r *ArcherRepo) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	dob, err := time.Parse("2006-01-02", data.DateOfBirth)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing date_of_birth: %w", err)
	}

	sql, args, err := StmtBuilder.Insert("archer").
		Columns(
			"first_name",
			"last_name",
			"email",
			"date_of_birth",
			"gender",
			"bowstyle",
			"draw_weight",
			"club_id",
			"google_picture_url",
			"google_subject",
		).
		Values(
			data.FirstName,
			data.LastName,
			data.Email,
			dob,
			data.Gender,
			data.Bowstyle,
			data.DrawWeight,
			data.ClubID,
			data.GooglePictureURL,
			data.GoogleSubject,
		).
		Suffix("RETURNING archer_id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting archer: %w", err)
	}

	return newID, nil
}

// Update updates fields of an archer specified by data for rows matching filter.
//
//nolint:gocritic // hugeParam: filter value parameter matches repository interface specification
func (r *ArcherRepo) Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
	q := StmtBuilder.Update("archer")
	setCount := 0

	if data.FirstName != nil {
		q = q.Set("first_name", *data.FirstName)
		setCount++
	}
	if data.LastName != nil {
		q = q.Set("last_name", *data.LastName)
		setCount++
	}
	if data.Gender != nil {
		q = q.Set("gender", *data.Gender)
		setCount++
	}
	if data.Bowstyle != nil {
		q = q.Set("bowstyle", *data.Bowstyle)
		setCount++
	}
	if data.DrawWeight != nil {
		q = q.Set("draw_weight", *data.DrawWeight)
		setCount++
	}
	if data.ClubID != nil {
		q = q.Set("club_id", *data.ClubID)
		setCount++
	}
	if data.GooglePictureURL != nil {
		q = q.Set("google_picture_url", *data.GooglePictureURL)
		setCount++
	}
	if data.LastLoginAt != nil {
		q = q.Set("last_login_at", *data.LastLoginAt)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.ArcherID != nil {
		q = q.Where(squirrel.Eq{"archer_id": *filter.ArcherID})
		whereCount++
	}
	if filter.FirstName != nil {
		q = q.Where(squirrel.Eq{"first_name": *filter.FirstName})
		whereCount++
	}
	if filter.LastName != nil {
		q = q.Where(squirrel.Eq{"last_name": *filter.LastName})
		whereCount++
	}
	if filter.Gender != nil {
		q = q.Where(squirrel.Eq{"gender": *filter.Gender})
		whereCount++
	}
	if filter.Bowstyle != nil {
		q = q.Where(squirrel.Eq{"bowstyle": *filter.Bowstyle})
		whereCount++
	}
	if filter.DrawWeight != nil {
		q = q.Where(squirrel.Eq{"draw_weight": *filter.DrawWeight})
		whereCount++
	}
	if filter.ClubID != nil {
		q = q.Where(squirrel.Eq{"club_id": *filter.ClubID})
		whereCount++
	}
	if filter.GoogleSubject != nil {
		q = q.Where(squirrel.Eq{"google_subject": *filter.GoogleSubject})
		whereCount++
	}
	if filter.LastLoginAt != nil {
		q = q.Where(squirrel.Eq{"last_login_at": *filter.LastLoginAt})
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

// Delete removes an archer by primary key identifier.
func (r *ArcherRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("archer").
		Where(squirrel.Eq{"archer_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
