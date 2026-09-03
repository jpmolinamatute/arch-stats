package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

// mockTx implements pgx.Tx for testing WithTx.
type mockTx struct {
	committed  bool
	rolledBack bool
	commitErr  error
	rbErr      error
}

func (m *mockTx) Begin(ctx context.Context) (pgx.Tx, error) { return m, nil }
func (m *mockTx) Commit(ctx context.Context) error {
	m.committed = true
	return m.commitErr
}

func (m *mockTx) Rollback(ctx context.Context) error {
	m.rolledBack = true
	return m.rbErr
}

func (m *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (m *mockTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }

func (m *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (m *mockTx) Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error) {
	return pgconn.CommandTag{}, nil
}

func (m *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}

func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return nil
}
func (m *mockTx) Conn() *pgx.Conn { return nil }

type mockTransactor struct {
	tx       *mockTx
	beginErr error
}

func (m *mockTransactor) Begin(ctx context.Context) (pgx.Tx, error) {
	if m.beginErr != nil {
		return nil, m.beginErr
	}
	return m.tx, nil
}

// mockSingleRow implements pgx.Row.
type mockSingleRow struct {
	scanFn func(dest ...any) error
}

func (r *mockSingleRow) Scan(dest ...any) error {
	return r.scanFn(dest...)
}

func TestWithTx_SuccessCommits(t *testing.T) {
	tx := &mockTx{}
	transactor := &mockTransactor{tx: tx}

	err := repository.WithTx(context.Background(), transactor, func(tx pgx.Tx) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !tx.committed {
		t.Error("expected transaction to be committed")
	}
}

func TestWithTx_CallbackErrorRollsBack(t *testing.T) {
	tx := &mockTx{}
	transactor := &mockTransactor{tx: tx}
	callbackErr := errors.New("something went wrong")

	err := repository.WithTx(context.Background(), transactor, func(tx pgx.Tx) error {
		return callbackErr
	})
	if !errors.Is(err, callbackErr) {
		t.Fatalf("expected callbackErr, got: %v", err)
	}
	if tx.committed {
		t.Error("expected transaction not to be committed")
	}
	if !tx.rolledBack {
		t.Error("expected transaction to be rolled back")
	}
}

func TestWithTx_BeginError(t *testing.T) {
	beginErr := errors.New("begin failed")
	transactor := &mockTransactor{beginErr: beginErr}

	err := repository.WithTx(context.Background(), transactor, func(tx pgx.Tx) error {
		return nil
	})
	if !errors.Is(err, beginErr) {
		t.Fatalf("expected beginErr, got: %v", err)
	}
}

func TestScanOne_Success(t *testing.T) {
	row := &mockSingleRow{
		scanFn: func(dest ...any) error {
			val := dest[0].(*string)
			*val = "archer-1"
			return nil
		},
	}

	result, err := repository.ScanOne(row, func(r pgx.Row) (string, error) {
		var s string
		err := r.Scan(&s)
		return s, err
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if result == nil || *result != "archer-1" {
		t.Fatalf("expected 'archer-1', got: %v", result)
	}
}

func TestScanOne_NoRowsReturnsNil(t *testing.T) {
	row := &mockSingleRow{
		scanFn: func(dest ...any) error {
			return pgx.ErrNoRows
		},
	}

	result, err := repository.ScanOne(row, func(r pgx.Row) (string, error) {
		var s string
		err := r.Scan(&s)
		return s, err
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got: %v", result)
	}
}

func TestScanOne_ScanError(t *testing.T) {
	scanErr := errors.New("scan failed")
	row := &mockSingleRow{
		scanFn: func(dest ...any) error {
			return scanErr
		},
	}

	_, err := repository.ScanOne(row, func(r pgx.Row) (string, error) {
		var s string
		err := r.Scan(&s)
		return s, err
	})
	if !errors.Is(err, scanErr) {
		t.Fatalf("expected scanErr, got: %v", err)
	}
}
