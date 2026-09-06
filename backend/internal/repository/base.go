package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

// DBTX abstracts database operations across connection pools and transactions.
// Both *pgxpool.Pool and pgx.Tx satisfy this interface.
type DBTX interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Transactor abstracts beginning a database transaction.
// *pgxpool.Pool satisfies this interface.
type Transactor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// StmtBuilder is the configured Squirrel statement builder using PostgreSQL dollar placeholders ($1, $2).
var StmtBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// WithTx wraps a function in a database transaction. The callback receives a pgx.Tx
// which satisfies the DBTX interface, so any repository method can participate in the transaction.
func WithTx(ctx context.Context, db Transactor, fn func(tx pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx) // no-op if already committed
	}()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ScanRows iterates over pgx.Rows, applying scanFn for each row, and returns a slice of results.
// It ensures rows are closed and checks rows.Err().
func ScanRows[T any](rows pgx.Rows, scanFn func(pgx.Rows) (T, error)) ([]T, error) {
	defer rows.Close()

	items := make([]T, 0)
	for rows.Next() {
		item, err := scanFn(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return items, nil
}

// ScanOne scans a single row using scanFn. If no row is returned (pgx.ErrNoRows),
// it returns nil, nil to signify an absent entity.
func ScanOne[T any](row pgx.Row, scanFn func(pgx.Row) (T, error)) (*T, error) {
	item, err := scanFn(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning row: %w", err)
	}
	return &item, nil
}

// errUpdateRequiresFilter is returned when an update query has no WHERE clauses.
var errUpdateRequiresFilter = errors.New("update requires at least one filter condition")

// deleteByID builds and executes a DELETE WHERE pk = id statement.
// Returns apperror.ErrNotFound if no row was deleted.
func deleteByID(ctx context.Context, db DBTX, table, pkColumn string, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete(table).
		Where(squirrel.Eq{pkColumn: id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("deleting %s: %w", table, err)
	}

	tag, err := db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("deleting %s: %w", table, err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// execUpdate executes a squirrel update builder and returns apperror.ErrNotFound
// if no rows were affected.
func execUpdate(ctx context.Context, db DBTX, q squirrel.UpdateBuilder) error {
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("building update query: %w", err)
	}

	tag, err := db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing update: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// findByID builds a SELECT WHERE pk = id query and scans a single result.
func findByID[T any](
	ctx context.Context,
	db DBTX,
	table string,
	pkColumn string,
	columns []string,
	id uuid.UUID,
	scanFn func(pgx.Row) (T, error),
) (*T, error) {
	sql, args, err := StmtBuilder.Select(columns...).
		From(table).
		Where(squirrel.Eq{pkColumn: id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("finding %s by id: %w", table, err)
	}

	row := db.QueryRow(ctx, sql, args...)
	return ScanOne(row, scanFn)
}

// createReturningID builds an INSERT ... RETURNING pk query and scans the result.
func createReturningID(
	ctx context.Context,
	db DBTX,
	builder squirrel.InsertBuilder,
	pkColumn string,
) (uuid.UUID, error) {
	sql, args, err := builder.Suffix("RETURNING " + pkColumn).ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create query: %w", err)
	}

	var newID uuid.UUID
	if err := db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting record: %w", err)
	}

	return newID, nil
}
