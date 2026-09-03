package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestMaintenanceRepo_RefreshOpenParticipants_Success(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			expectedSQL := "REFRESH MATERIALIZED VIEW CONCURRENTLY open_participants"
			if sql != expectedSQL {
				t.Errorf("expected SQL %q, got %q", expectedSQL, sql)
			}
			return pgconn.NewCommandTag("REFRESH MATERIALIZED VIEW"), nil
		},
	}

	repo := repository.NewMaintenanceRepo(mock)
	err := repo.RefreshOpenParticipants(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMaintenanceRepo_RefreshOpenParticipants_Error(t *testing.T) {
	dbErr := errors.New("database failure")
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, dbErr
		},
	}

	repo := repository.NewMaintenanceRepo(mock)
	err := repo.RefreshOpenParticipants(context.Background())
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

func TestMaintenanceRepo_GetSchemaVersion_Success(t *testing.T) {
	expectedVersion := int64(6)
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					v := dest[0].(*int64)
					*v = expectedVersion
					return nil
				},
			}
		},
	}

	repo := repository.NewMaintenanceRepo(mock)
	version, err := repo.GetSchemaVersion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != expectedVersion {
		t.Errorf("expected version %d, got %d", expectedVersion, version)
	}
}

func TestMaintenanceRepo_GetSchemaVersion_Error(t *testing.T) {
	dbErr := errors.New("relation does not exist")
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return dbErr
				},
			}
		},
	}

	repo := repository.NewMaintenanceRepo(mock)
	_, err := repo.GetSchemaVersion(context.Background())
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}
