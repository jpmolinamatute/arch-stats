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

var archerColumns = []string{
	"archer_id",
	"email",
	"first_name",
	"last_name",
	"date_of_birth",
	"gender",
	"is_deleted",
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
		&a.Email,
		&a.FirstName,
		&a.LastName,
		&dobRaw,
		&a.Gender,
		&a.IsDeleted,
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
	return findByID(ctx, r.db, "archer", "archer_id", archerColumns, id, func(row pgx.Row) (model.ArcherRead, error) {
		return scanArcher(row)
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
	cols := make([]string, len(archerColumns))
	for i, c := range archerColumns {
		cols[i] = "archer." + c
	}
	sql, args, err := StmtBuilder.Select(cols...).
		From("archer").
		Join("auth ON auth.archer_id = archer.archer_id").
		Where(squirrel.Eq{"auth.google_subject": sub}).
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
func (r *ArcherRepo) FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
	q := StmtBuilder.Select(archerColumns...).
		From("archer").
		OrderBy("archer_id ASC")

	if filter.ArcherID != nil {
		q = q.Where(squirrel.Eq{"archer_id": *filter.ArcherID})
	}
	if filter.Email != nil {
		q = q.Where(squirrel.Eq{"email": *filter.Email})
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
	if filter.IsDeleted != nil {
		q = q.Where(squirrel.Eq{"is_deleted": *filter.IsDeleted})
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

	var archerID uuid.UUID
	if data.ArcherID != nil && *data.ArcherID != uuid.Nil {
		archerID = *data.ArcherID
	} else {
		sub := "google-sub-" + uuid.New().String()
		err := r.db.QueryRow(ctx, "INSERT INTO auth (google_subject) VALUES ($1) RETURNING archer_id", sub).Scan(&archerID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("creating auth identity: %w", err)
		}
	}

	cols := []string{"archer_id", "first_name", "last_name", "email", "date_of_birth", "gender"}
	vals := []any{archerID, data.FirstName, data.LastName, data.Email, dob, data.Gender}

	builder := StmtBuilder.Insert("archer").
		Columns(cols...).
		Values(vals...)

	return createReturningID(ctx, r.db, builder, "archer_id")
}

// Update updates fields of an archer specified by data for rows matching filter.
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
	if data.Email != nil {
		q = q.Set("email", *data.Email)
		setCount++
	}
	if data.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *data.DateOfBirth)
		if err != nil {
			return fmt.Errorf("parsing date_of_birth: %w", err)
		}
		q = q.Set("date_of_birth", dob)
		setCount++
	}
	if data.Gender != nil {
		q = q.Set("gender", *data.Gender)
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
	if filter.Email != nil {
		q = q.Where(squirrel.Eq{"email": *filter.Email})
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
	if filter.IsDeleted != nil {
		q = q.Where(squirrel.Eq{"is_deleted": *filter.IsDeleted})
		whereCount++
	}

	if whereCount == 0 {
		return errUpdateRequiresFilter
	}

	return execUpdate(ctx, r.db, q)
}

// Delete removes an archer by primary key identifier.
func (r *ArcherRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return deleteByID(ctx, r.db, "archer", "archer_id", id)
}
