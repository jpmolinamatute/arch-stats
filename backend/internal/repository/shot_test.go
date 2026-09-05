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

func TestShotRepo_Create_WithCreatedAt(t *testing.T) {
	slotID := uuid.New()
	arrowID := uuid.New()
	expectedID := uuid.New()
	x := 12.0
	y := -4.0
	score := 9
	createdAt := time.Now().Truncate(time.Second).UTC()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = expectedID
					return nil
				},
			}
		},
	}

	repo := repository.NewShotRepo(mock)
	id, err := repo.Create(context.Background(), model.ShotCreate{
		SlotID:    slotID,
		X:         &x,
		Y:         &y,
		IsX:       false,
		Score:     &score,
		ArrowID:   &arrowID,
		CreatedAt: &createdAt,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != expectedID {
		t.Fatalf("expected ID %v, got %v", expectedID, id)
	}

	if !strings.HasPrefix(executedSQL, "INSERT INTO shot (slot_id,x,y,is_x,score,arrow_id,created_at)") {
		t.Errorf("unexpected INSERT SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "RETURNING shot_id") {
		t.Errorf("expected RETURNING shot_id in SQL: %s", executedSQL)
	}
	if len(executedArgs) != 7 {
		t.Errorf("expected 7 args, got %d: %v", len(executedArgs), executedArgs)
	}
}

func TestShotRepo_Create_WithoutCreatedAt(t *testing.T) {
	slotID := uuid.New()
	expectedID := uuid.New()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = expectedID
					return nil
				},
			}
		},
	}

	repo := repository.NewShotRepo(mock)
	id, err := repo.Create(context.Background(), model.ShotCreate{
		SlotID: slotID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != expectedID {
		t.Fatalf("expected ID %v, got %v", expectedID, id)
	}

	if !strings.HasPrefix(executedSQL, "INSERT INTO shot (slot_id,x,y,is_x,score,arrow_id)") {
		t.Errorf("unexpected INSERT SQL without created_at: %s", executedSQL)
	}
	if strings.Contains(executedSQL, "created_at") {
		t.Errorf("expected created_at omitted from INSERT, got: %s", executedSQL)
	}
	if len(executedArgs) != 6 {
		t.Errorf("expected 6 args, got %d: %v", len(executedArgs), executedArgs)
	}
}

func TestShotRepo_CreateBatch_Success(t *testing.T) {
	slotID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()
	now := time.Now().Truncate(time.Second).UTC()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			return &mockMultiRows{
				records: [][]any{
					{id1},
					{id2},
				},
			}, nil
		},
	}

	repo := repository.NewShotRepo(mock)
	shots := []model.ShotCreate{
		{SlotID: slotID, CreatedAt: &now},
		{SlotID: slotID, CreatedAt: &now},
	}

	ids, err := repo.CreateBatch(context.Background(), shots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != id1 || ids[1] != id2 {
		t.Fatalf("unexpected IDs returned: %v", ids)
	}

	if !strings.HasPrefix(executedSQL, "INSERT INTO shot (slot_id,x,y,is_x,score,arrow_id,created_at) VALUES") {
		t.Errorf("unexpected batch INSERT SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "RETURNING shot_id") {
		t.Errorf("expected RETURNING shot_id in SQL: %s", executedSQL)
	}
	if len(executedArgs) != 14 {
		t.Errorf("expected 14 args, got %d: %v", len(executedArgs), executedArgs)
	}
}

func TestShotRepo_CreateBatch_EmptyReturnsNil(t *testing.T) {
	mock := &mockDBTX{}
	repo := repository.NewShotRepo(mock)
	ids, err := repo.CreateBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ids != nil {
		t.Fatalf("expected nil IDs for empty batch, got %v", ids)
	}
}

func TestShotRepo_Update_Success(t *testing.T) {
	shotID := uuid.New()
	arrowID := uuid.New()
	x := 1.0
	y := 2.0
	score := 10
	isX := true

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := repository.NewShotRepo(mock)
	err := repo.Update(
		context.Background(),
		model.ShotSet{
			X:       &x,
			Y:       &y,
			Score:   &score,
			IsX:     &isX,
			ArrowID: &arrowID,
		},
		model.ShotFilter{ShotID: &shotID},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(executedSQL, "UPDATE shot SET") {
		t.Errorf("unexpected UPDATE SQL: %s", executedSQL)
	}
	for _, col := range []string{"x", "y", "is_x", "score", "arrow_id"} {
		if !strings.Contains(executedSQL, col+" =") {
			t.Errorf("expected SET clause for %s in SQL: %s", col, executedSQL)
		}
	}
	if !strings.Contains(executedSQL, "WHERE shot_id = $6") {
		t.Errorf("expected WHERE shot_id = $6, got: %s", executedSQL)
	}
	if len(executedArgs) != 6 {
		t.Errorf("expected 6 args, got %d: %v", len(executedArgs), executedArgs)
	}
}

func TestShotRepo_Update_EmptySetIsNoOp(t *testing.T) {
	mock := &mockDBTX{}
	repo := repository.NewShotRepo(mock)
	shotID := uuid.New()

	err := repo.Update(context.Background(), model.ShotSet{}, model.ShotFilter{ShotID: &shotID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShotRepo_Update_MissingFilterReturnsError(t *testing.T) {
	score := 10
	repo := repository.NewShotRepo(&mockDBTX{})
	err := repo.Update(context.Background(), model.ShotSet{Score: &score}, model.ShotFilter{})
	if err == nil {
		t.Fatal("expected error for empty filter, got nil")
	}
}

func TestShotRepo_Update_NotFound(t *testing.T) {
	shotID := uuid.New()
	score := 10
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}

	repo := repository.NewShotRepo(mock)
	err := repo.Update(context.Background(), model.ShotSet{Score: &score}, model.ShotFilter{ShotID: &shotID})
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestShotRepo_Delete_Success(t *testing.T) {
	shotID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("DELETE 1"), nil
		},
	}

	repo := repository.NewShotRepo(mock)
	err := repo.Delete(context.Background(), shotID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(executedSQL, "DELETE FROM shot WHERE shot_id = $1") {
		t.Errorf("unexpected DELETE SQL: %s", executedSQL)
	}
	if len(executedArgs) != 1 || (executedArgs[0] != shotID && executedArgs[0] != shotID.String()) {
		t.Errorf("expected arg %v, got %v", shotID, executedArgs)
	}
}

func TestShotRepo_Delete_NotFound(t *testing.T) {
	shotID := uuid.New()
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("DELETE 0"), nil
		},
	}

	repo := repository.NewShotRepo(mock)
	err := repo.Delete(context.Background(), shotID)
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestShotRepo_CountBySlotID_Success(t *testing.T) {
	slotID := uuid.New()
	expectedCount := 18

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*(dest[0].(*int)) = expectedCount
					return nil
				},
			}
		},
	}

	repo := repository.NewShotRepo(mock)
	count, err := repo.CountBySlotID(context.Background(), slotID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != expectedCount {
		t.Fatalf("expected count %d, got %d", expectedCount, count)
	}

	if !strings.HasPrefix(executedSQL, "SELECT COUNT(*) FROM shot WHERE slot_id = $1") {
		t.Errorf("unexpected COUNT query SQL: %s", executedSQL)
	}
	if len(executedArgs) != 1 || (executedArgs[0] != slotID && executedArgs[0] != slotID.String()) {
		t.Errorf("expected arg %v, got %v", slotID, executedArgs)
	}
}

func TestShotRepo_GetLatestShotTime_Success(t *testing.T) {
	slotID := uuid.New()
	now := time.Now().Truncate(time.Second).UTC()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*(dest[0].(*time.Time)) = now
					return nil
				},
			}
		},
	}

	repo := repository.NewShotRepo(mock)
	ts, err := repo.GetLatestShotTime(context.Background(), slotID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts == nil || !ts.Equal(now) {
		t.Fatalf("expected timestamp %v, got %v", now, ts)
	}

	if !strings.HasPrefix(executedSQL, "SELECT created_at FROM shot WHERE slot_id = $1 ORDER BY created_at DESC LIMIT 1") {
		t.Errorf("unexpected SQL for latest shot time: %s", executedSQL)
	}
	if len(executedArgs) != 1 || (executedArgs[0] != slotID && executedArgs[0] != slotID.String()) {
		t.Errorf("expected arg %v, got %v", slotID, executedArgs)
	}
}

func TestShotRepo_GetLatestShotTime_NoneReturnsNil(t *testing.T) {
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
	ts, err := repo.GetLatestShotTime(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts != nil {
		t.Fatalf("expected nil timestamp, got %v", ts)
	}
}
