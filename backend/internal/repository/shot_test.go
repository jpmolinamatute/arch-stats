package repository_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func sampleShotRow(
	shotID, slotID uuid.UUID,
	x, y *float64,
	isX bool,
	score *int,
	arrowID *uuid.UUID,
	createdAt time.Time,
) []any {
	return []any{
		shotID,
		slotID,
		x,
		y,
		isX,
		score,
		arrowID,
		createdAt,
	}
}

func TestShotRepo_FindByID_Success(t *testing.T) {
	shotID := uuid.New()
	slotID := uuid.New()
	arrowID := uuid.New()
	x := 15.2
	y := -8.4
	score := 10
	createdAt := time.Now().Truncate(time.Second).UTC()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleShotRow(shotID, slotID, &x, &y, true, &score, &arrowID, createdAt)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewShotRepo(mock)
	shot, err := repo.FindByID(context.Background(), shotID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shot == nil {
		t.Fatal("expected shot, got nil")
	}

	expectedPrefix := "SELECT shot_id, slot_id, x, y, is_x, score, arrow_id, created_at FROM shot"
	if !strings.HasPrefix(executedSQL, expectedPrefix) {
		t.Errorf("unexpected query SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE shot_id = $1") {
		t.Errorf("expected WHERE shot_id = $1, got: %s", executedSQL)
	}
	if len(executedArgs) != 1 || (executedArgs[0] != shotID && executedArgs[0] != shotID.String()) {
		t.Errorf("expected arg %v, got %v", shotID, executedArgs)
	}

	if shot.ShotID != shotID {
		t.Errorf("expected ShotID %v, got %v", shotID, shot.ShotID)
	}
	if shot.SlotID != slotID {
		t.Errorf("expected SlotID %v, got %v", slotID, shot.SlotID)
	}
	if shot.X == nil || *shot.X != x {
		t.Errorf("expected X %v, got %v", x, shot.X)
	}
	if shot.Y == nil || *shot.Y != y {
		t.Errorf("expected Y %v, got %v", y, shot.Y)
	}
	if !shot.IsX {
		t.Errorf("expected IsX true, got false")
	}
	if shot.Score == nil || *shot.Score != score {
		t.Errorf("expected Score %v, got %v", score, shot.Score)
	}
	if shot.ArrowID == nil || *shot.ArrowID != arrowID {
		t.Errorf("expected ArrowID %v, got %v", arrowID, shot.ArrowID)
	}
	if !shot.CreatedAt.Equal(createdAt) {
		t.Errorf("expected CreatedAt %v, got %v", createdAt, shot.CreatedAt)
	}
}

func TestShotRepo_FindByID_NotFoundReturnsNil(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewShotRepo(mock)
	shot, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shot != nil {
		t.Fatalf("expected nil shot, got %+v", shot)
	}
}

func TestShotRepo_FindByID_QueryError(t *testing.T) {
	expectedErr := errors.New("connection failed")
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return expectedErr
				},
			}
		},
	}

	repo := repository.NewShotRepo(mock)
	_, err := repo.FindByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestShotRepo_FindBySlotID_Success(t *testing.T) {
	slotID := uuid.New()
	shot1ID := uuid.New()
	shot2ID := uuid.New()
	x1, y1, s1 := 10.0, 10.0, 9
	x2, y2, s2 := 0.0, 0.0, 10
	t1 := time.Now().Add(-10 * time.Second).Truncate(time.Second).UTC()
	t2 := time.Now().Truncate(time.Second).UTC()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			return &mockMultiRows{
				records: [][]any{
					sampleShotRow(shot1ID, slotID, &x1, &y1, false, &s1, nil, t1),
					sampleShotRow(shot2ID, slotID, &x2, &y2, true, &s2, nil, t2),
				},
			}, nil
		},
	}

	repo := repository.NewShotRepo(mock)
	shots, err := repo.FindBySlotID(context.Background(), slotID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shots) != 2 {
		t.Fatalf("expected 2 shots, got %d", len(shots))
	}

	if !strings.Contains(executedSQL, "WHERE slot_id = $1") {
		t.Errorf("expected WHERE slot_id = $1, got: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "ORDER BY created_at ASC") {
		t.Errorf("expected ORDER BY created_at ASC, got: %s", executedSQL)
	}
	if len(executedArgs) != 1 || (executedArgs[0] != slotID && executedArgs[0] != slotID.String()) {
		t.Errorf("expected arg %v, got %v", slotID, executedArgs)
	}
	if shots[0].ShotID != shot1ID || shots[1].ShotID != shot2ID {
		t.Errorf("unexpected shot IDs in returned list: %+v", shots)
	}
}

func TestShotRepo_FindAll_WithFilter(t *testing.T) {
	shotID := uuid.New()
	slotID := uuid.New()
	arrowID := uuid.New()
	x := 5.5
	y := -2.1
	score := 8
	createdAt := time.Now().Truncate(time.Second).UTC()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			return &mockMultiRows{records: [][]any{}}, nil
		},
	}

	repo := repository.NewShotRepo(mock)
	filter := model.ShotFilter{
		ShotID:    &shotID,
		SlotID:    &slotID,
		X:         &x,
		Y:         &y,
		Score:     &score,
		ArrowID:   &arrowID,
		CreatedAt: &createdAt,
	}

	_, err := repo.FindAll(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, col := range []string{"shot_id", "slot_id", "x", "y", "score", "arrow_id", "created_at"} {
		if !strings.Contains(executedSQL, col) {
			t.Errorf("expected filter condition for %s in SQL: %s", col, executedSQL)
		}
	}
	if len(executedArgs) != 7 {
		t.Errorf("expected 7 args, got %d: %v", len(executedArgs), executedArgs)
	}
}
