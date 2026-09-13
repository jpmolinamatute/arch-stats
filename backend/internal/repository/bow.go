package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var bowColumns = []string{
	"bow_id",
	"archer_id",
	"name",
	"bowstyle",
	"draw_weight",
	"is_deleted",
	"created_at",
}

// BowRepo manages database operations for bow equipment inventory.
type BowRepo struct {
	db DBTX
}

// NewBowRepo constructs a BowRepo backed by DBTX.
func NewBowRepo(db DBTX) *BowRepo {
	return &BowRepo{db: db}
}

// WithTx returns a new BowRepo bound to the given transaction.
func (r *BowRepo) WithTx(tx pgx.Tx) *BowRepo {
	return &BowRepo{db: tx}
}

func scanBow(scanner interface{ Scan(dest ...any) error }) (model.BowRead, error) {
	var b model.BowRead
	err := scanner.Scan(
		&b.BowID,
		&b.ArcherID,
		&b.Name,
		&b.Bowstyle,
		&b.DrawWeight,
		&b.IsDeleted,
		&b.CreatedAt,
	)
	if err != nil {
		return model.BowRead{}, err
	}
	return b, nil
}

// Create inserts a new bow equipment record, returning the generated UUID.
func (r *BowRepo) Create(ctx context.Context, data model.BowCreate) (uuid.UUID, error) {
	cols := []string{"archer_id", "name", "bowstyle", "draw_weight"}
	vals := []any{data.ArcherID, data.Name, data.Bowstyle, data.DrawWeight}

	builder := StmtBuilder.Insert("bow").
		Columns(cols...).
		Values(vals...)

	return createReturningID(ctx, r.db, builder, "bow_id")
}

// FindByID retrieves a bow by primary key identifier.
func (r *BowRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.BowRead, error) {
	return findByID(ctx, r.db, "bow", "bow_id", bowColumns, id, func(row pgx.Row) (model.BowRead, error) {
		return scanBow(row)
	})
}

// FindAllByArcherID retrieves all active bows belonging to the specified archer ordered by creation time.
func (r *BowRepo) FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.BowRead, error) {
	sql, args, err := StmtBuilder.Select(bowColumns...).
		From("bow").
		Where(squirrel.Eq{"archer_id": archerID}).
		Where(squirrel.Eq{"is_deleted": false}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all bows query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying bows: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.BowRead, error) {
		return scanBow(r)
	})
}

// CountByArcherID counts active (non-deleted) bows owned by an archer.
func (r *BowRepo) CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("bow").
		Where(squirrel.Eq{"archer_id": archerID}).
		Where(squirrel.Eq{"is_deleted": false}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count bows query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting bows by archer id: %w", err)
	}

	return count, nil
}
