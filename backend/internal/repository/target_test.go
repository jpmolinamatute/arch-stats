package repository_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func sampleTargetRow(id, sessionID uuid.UUID, distance, lane int) []any {
	now := time.Now().Truncate(time.Second)
	return []any{
		id,
		sessionID,
		distance,
		lane,
		now,
	}
}

func TestTargetRepo_FindByID_Success(t *testing.T) {
	targetID := uuid.New()
	sessionID := uuid.New()
	expectedRow := sampleTargetRow(targetID, sessionID, 18, 1)

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if !strings.Contains(sql, "FROM target") || !strings.Contains(sql, "WHERE target_id = $1") {
				t.Fatalf("unexpected SQL: %s", sql)
			}
			if len(args) != 1 || (args[0] != targetID && args[0] != targetID.String()) {
				t.Fatalf("unexpected args: %v", args)
			}
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					multi := &mockMultiRows{records: [][]any{expectedRow}, idx: 1}
					return multi.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewTargetRepo(mock)
	target, err := repo.FindByID(context.Background(), targetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == nil {
		t.Fatal("expected target, got nil")
	}
	if target.TargetID != targetID || target.SessionID != sessionID || target.Distance != 18 || target.Lane != 1 {
		t.Fatalf("mismatched target fields: %+v", target)
	}
}

func TestTargetRepo_FindByID_NotFoundReturnsNil(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewTargetRepo(mock)
	target, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target != nil {
		t.Fatalf("expected nil target, got %+v", target)
	}
}

func TestTargetRepo_FindByID_QueryError(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return errors.New("db error")
				},
			}
		},
	}

	repo := repository.NewTargetRepo(mock)
	_, err := repo.FindByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetRepo_FindBySlotID_Success(t *testing.T) {
	slotID := uuid.New()
	targetID := uuid.New()
	sessionID := uuid.New()
	expectedRows := [][]any{sampleTargetRow(targetID, sessionID, 70, 3)}

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			if !strings.Contains(sql, "FROM target") || !strings.Contains(sql, "JOIN slot") || !strings.Contains(sql, "slot.slot_id = $1") {
				t.Fatalf("unexpected SQL: %s", sql)
			}
			if len(args) != 1 || (args[0] != slotID && args[0] != slotID.String()) {
				t.Fatalf("unexpected args: %v", args)
			}
			return &mockMultiRows{records: expectedRows}, nil
		},
	}

	repo := repository.NewTargetRepo(mock)
	targets, err := repo.FindBySlotID(context.Background(), slotID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].TargetID != targetID || targets[0].Lane != 3 {
		t.Fatalf("unexpected target: %+v", targets[0])
	}
}

func TestTargetRepo_FindBySlotID_QueryError(t *testing.T) {
	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return nil, errors.New("query failed")
		},
	}

	repo := repository.NewTargetRepo(mock)
	_, err := repo.FindBySlotID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetRepo_FindBySessionID_Success(t *testing.T) {
	sessionID := uuid.New()
	expectedRows := [][]any{
		sampleTargetRow(uuid.New(), sessionID, 18, 1),
		sampleTargetRow(uuid.New(), sessionID, 18, 2),
	}

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			if !strings.Contains(sql, "WHERE session_id = $1") || !strings.Contains(sql, "ORDER BY lane ASC") {
				t.Fatalf("unexpected SQL: %s", sql)
			}
			if len(args) != 1 || (args[0] != sessionID && args[0] != sessionID.String()) {
				t.Fatalf("unexpected args: %v", args)
			}
			return &mockMultiRows{records: expectedRows}, nil
		},
	}

	repo := repository.NewTargetRepo(mock)
	targets, err := repo.FindBySessionID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestTargetRepo_Create_Success(t *testing.T) {
	sessionID := uuid.New()
	newID := uuid.New()

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if !strings.Contains(sql, "INSERT INTO target") || !strings.Contains(sql, "RETURNING target_id") {
				t.Fatalf("unexpected SQL: %s", sql)
			}
			if len(args) != 3 || args[0] != sessionID || args[1] != 50 || args[2] != 4 {
				t.Fatalf("unexpected args: %v", args)
			}
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					d := dest[0].(*uuid.UUID)
					*d = newID
					return nil
				},
			}
		},
	}

	repo := repository.NewTargetRepo(mock)
	id, err := repo.Create(context.Background(), model.TargetCreate{
		SessionID: sessionID,
		Distance:  50,
		Lane:      4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != newID {
		t.Fatalf("expected ID %s, got %s", newID, id)
	}
}

func TestTargetRepo_Create_DBError(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return errors.New("insert error")
				},
			}
		},
	}

	repo := repository.NewTargetRepo(mock)
	_, err := repo.Create(context.Background(), model.TargetCreate{
		SessionID: uuid.New(),
		Distance:  18,
		Lane:      1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetRepo_Update_Success(t *testing.T) {
	targetID := uuid.New()
	newDistance := 70
	newLane := 5

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			if !strings.Contains(sql, "UPDATE target") || !strings.Contains(sql, "distance = $1") || !strings.Contains(sql, "lane = $2") {
				t.Fatalf("unexpected SQL: %s", sql)
			}
			if !strings.Contains(sql, "target_id = $3") {
				t.Fatalf("unexpected WHERE clause: %s", sql)
			}
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := repository.NewTargetRepo(mock)
	err := repo.Update(context.Background(), model.TargetSet{
		Distance: &newDistance,
		Lane:     &newLane,
	}, model.TargetFilter{
		TargetID: &targetID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTargetRepo_Update_EmptySetIsNoOp(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			t.Fatal("Exec should not be called on empty set")
			return pgconn.CommandTag{}, nil
		},
	}

	targetID := uuid.New()
	repo := repository.NewTargetRepo(mock)
	err := repo.Update(context.Background(), model.TargetSet{}, model.TargetFilter{
		TargetID: &targetID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTargetRepo_Update_MissingFilterReturnsError(t *testing.T) {
	mock := &mockDBTX{}
	repo := repository.NewTargetRepo(mock)
	dist := 18
	err := repo.Update(context.Background(), model.TargetSet{Distance: &dist}, model.TargetFilter{})
	if err == nil {
		t.Fatal("expected error when filter is empty, got nil")
	}
}

func TestTargetRepo_Update_NotFound(t *testing.T) {
	targetID := uuid.New()
	dist := 25
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}

	repo := repository.NewTargetRepo(mock)
	err := repo.Update(context.Background(), model.TargetSet{Distance: &dist}, model.TargetFilter{
		TargetID: &targetID,
	})
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTargetRepo_Delete_Success(t *testing.T) {
	targetID := uuid.New()
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			if !strings.Contains(sql, "DELETE FROM target") || !strings.Contains(sql, "target_id = $1") {
				t.Fatalf("unexpected SQL: %s", sql)
			}
			return pgconn.NewCommandTag("DELETE 1"), nil
		},
	}

	repo := repository.NewTargetRepo(mock)
	if err := repo.Delete(context.Background(), targetID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTargetRepo_Delete_NotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("DELETE 0"), nil
		},
	}

	repo := repository.NewTargetRepo(mock)
	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTargetRepo_WithTx(t *testing.T) {
	mock := &mockDBTX{}
	repo := repository.NewTargetRepo(mock)
	txRepo := repo.WithTx(nil)
	if txRepo == nil {
		t.Fatal("expected non-nil repo from WithTx")
	}
}
