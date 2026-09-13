package repository_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func sampleArrowRow(
	id, archerID uuid.UUID,
	set, number int16,
	status model.ArrowStatus,
	spine, length, weight *float64,
	isDeleted bool,
	createdAt time.Time,
) []any {
	return []any{
		id,
		archerID,
		set,
		number,
		status,
		spine,
		length,
		weight,
		isDeleted,
		createdAt,
	}
}

func TestArrowRepo_Create_Success(t *testing.T) {
	expectedID := uuid.New()
	archerID := uuid.New()
	spine := 500.0
	length := 28.5
	weight := 320.0
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

	repo := repository.NewArrowRepo(mock)
	payload := model.ArrowCreate{
		ArcherID:    archerID,
		ArrowSet:    1,
		ArrowNumber: 1,
		Status:      nil, // should default to 'in_use'
		Spine:       &spine,
		Length:      &length,
		Weight:      &weight,
	}

	id, err := repo.Create(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != expectedID {
		t.Errorf("expected id %v, got %v", expectedID, id)
	}
	if !strings.Contains(executedSQL, "INSERT INTO arrow") {
		t.Errorf("expected INSERT INTO arrow in query: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "RETURNING arrow_id") {
		t.Errorf("expected RETURNING arrow_id in query: %s", executedSQL)
	}
	if len(executedArgs) != 7 {
		t.Fatalf("expected 7 query args, got %d", len(executedArgs))
	}
	if executedArgs[0] != archerID {
		t.Errorf("expected arg[0] %v, got %v", archerID, executedArgs[0])
	}
	if executedArgs[1] != int16(1) {
		t.Errorf("expected arg[1] 1, got %v", executedArgs[1])
	}
	if executedArgs[2] != int16(1) {
		t.Errorf("expected arg[2] 1, got %v", executedArgs[2])
	}
	if executedArgs[3] != model.ArrowStatusInUse {
		t.Errorf("expected arg[3] default in_use, got %v", executedArgs[3])
	}
}

func TestArrowRepo_Create_CustomStatus(t *testing.T) {
	expectedID := uuid.New()
	status := model.ArrowStatusDamaged
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = expectedID
					return nil
				},
			}
		},
	}

	repo := repository.NewArrowRepo(mock)
	payload := model.ArrowCreate{
		ArcherID:    uuid.New(),
		ArrowSet:    1,
		ArrowNumber: 2,
		Status:      &status,
	}

	_, err := repo.Create(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if executedArgs[3] != model.ArrowStatusDamaged {
		t.Errorf("expected arg[3] damaged, got %v", executedArgs[3])
	}
}

func TestArrowRepo_CreateBatch_Success(t *testing.T) {
	archerID := uuid.New()
	spine := 600.0
	length := 29.0
	weight := 310.0
	createdAt := time.Now().Truncate(time.Second).UTC()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			return &mockMultiRows{
				records: [][]any{
					sampleArrowRow(uuid.New(), archerID, 1, 1, model.ArrowStatusInUse, &spine, &length, &weight, false, createdAt),
					sampleArrowRow(uuid.New(), archerID, 1, 2, model.ArrowStatusInUse, &spine, &length, &weight, false, createdAt),
					sampleArrowRow(uuid.New(), archerID, 1, 3, model.ArrowStatusInUse, &spine, &length, &weight, false, createdAt),
				},
			}, nil
		},
	}

	repo := repository.NewArrowRepo(mock)
	batch := model.ArrowBatchCreate{
		ArcherID: archerID,
		ArrowSet: 1,
		Count:    3,
		Status:   nil, // default to in_use
		Spine:    &spine,
		Length:   &length,
		Weight:   &weight,
	}

	arrows, err := repo.CreateBatch(context.Background(), batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arrows) != 3 {
		t.Fatalf("expected 3 arrows, got %d", len(arrows))
	}
	if !strings.Contains(executedSQL, "INSERT INTO arrow") {
		t.Errorf("expected INSERT INTO arrow in query: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "RETURNING") {
		t.Errorf("expected RETURNING in query: %s", executedSQL)
	}
	if len(executedArgs) != 21 { // 3 rows * 7 columns
		t.Fatalf("expected 21 query args for batch of 3, got %d", len(executedArgs))
	}
}

func TestArrowRepo_CreateBatch_EmptyReturnsNil(t *testing.T) {
	repo := repository.NewArrowRepo(&mockDBTX{})
	arrows, err := repo.CreateBatch(context.Background(), model.ArrowBatchCreate{Count: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if arrows != nil {
		t.Errorf("expected nil for 0 count, got %v", arrows)
	}
}

func TestArrowRepo_FindByID_Success(t *testing.T) {
	arrowID := uuid.New()
	archerID := uuid.New()
	spine := 400.0
	length := 29.5
	weight := 350.0
	createdAt := time.Now().Truncate(time.Second).UTC()

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleArrowRow(arrowID, archerID, 1, 4, model.ArrowStatusInUse, &spine, &length, &weight, false, createdAt)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewArrowRepo(mock)
	arrow, err := repo.FindByID(context.Background(), arrowID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if arrow == nil {
		t.Fatal("expected arrow, got nil")
	}
	if arrow.ArrowID != arrowID {
		t.Errorf("expected arrowID %v, got %v", arrowID, arrow.ArrowID)
	}
	if arrow.ArrowSet != 1 || arrow.ArrowNumber != 4 {
		t.Errorf("expected set 1, number 4, got set %d, number %d", arrow.ArrowSet, arrow.ArrowNumber)
	}
	if arrow.Status != model.ArrowStatusInUse {
		t.Errorf("expected status in_use, got %v", arrow.Status)
	}
	if arrow.Spine == nil || *arrow.Spine != 400.0 {
		t.Errorf("expected spine 400.0, got %v", arrow.Spine)
	}
}

func TestArrowRepo_FindByID_NotFound(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewArrowRepo(mock)
	arrow, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if arrow != nil {
		t.Errorf("expected nil arrow on ErrNoRows, got %v", arrow)
	}
}

func TestArrowRepo_FindAllByArcherID_Success(t *testing.T) {
	archerID := uuid.New()
	createdAt := time.Now().Truncate(time.Second).UTC()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			return &mockMultiRows{
				records: [][]any{
					sampleArrowRow(uuid.New(), archerID, 1, 1, model.ArrowStatusInUse, nil, nil, nil, false, createdAt),
					sampleArrowRow(uuid.New(), archerID, 1, 2, model.ArrowStatusInUse, nil, nil, nil, false, createdAt),
				},
			}, nil
		},
	}

	repo := repository.NewArrowRepo(mock)
	arrows, err := repo.FindAllByArcherID(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arrows) != 2 {
		t.Fatalf("expected 2 arrows, got %d", len(arrows))
	}
	if !strings.Contains(executedSQL, "ORDER BY arrow_set ASC, arrow_number ASC") {
		t.Errorf("expected ORDER BY arrow_set ASC, arrow_number ASC in query: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "is_deleted =") {
		t.Errorf("expected is_deleted filter in query: %s", executedSQL)
	}
	if len(executedArgs) != 2 || executedArgs[0] != archerID.String() || executedArgs[1] != false {
		t.Errorf("expected args [%s, false], got %v", archerID.String(), executedArgs)
	}
}

func TestArrowRepo_CountByArcherID_Success(t *testing.T) {
	archerID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*dest[0].(*int) = 6
					return nil
				},
			}
		},
	}

	repo := repository.NewArrowRepo(mock)
	count, err := repo.CountByArcherID(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 6 {
		t.Errorf("expected count 6, got %d", count)
	}
	if !strings.Contains(executedSQL, "COUNT(*)") {
		t.Errorf("expected COUNT(*) in query: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "is_deleted =") {
		t.Errorf("expected is_deleted filter in query: %s", executedSQL)
	}
	if len(executedArgs) != 2 || executedArgs[0] != archerID.String() || executedArgs[1] != false {
		t.Errorf("expected args [%s, false], got %v", archerID.String(), executedArgs)
	}
}

func TestArrowRepo_CountInUseByArcherID_Success(t *testing.T) {
	archerID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*dest[0].(*int) = 5
					return nil
				},
			}
		},
	}

	repo := repository.NewArrowRepo(mock)
	count, err := repo.CountInUseByArcherID(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
	if !strings.Contains(executedSQL, "COUNT(*)") {
		t.Errorf("expected COUNT(*) in query: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "status =") {
		t.Errorf("expected status filter in query: %s", executedSQL)
	}
	if len(executedArgs) != 3 || executedArgs[0] != archerID.String() || executedArgs[1] != model.ArrowStatusInUse || executedArgs[2] != false {
		t.Errorf("expected args [%s, in_use, false], got %v", archerID.String(), executedArgs)
	}
}

func TestArrowRepo_FindByArcherAndSet_Success(t *testing.T) {
	archerID := uuid.New()
	createdAt := time.Now().Truncate(time.Second).UTC()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			return &mockMultiRows{
				records: [][]any{
					sampleArrowRow(uuid.New(), archerID, 2, 1, model.ArrowStatusInUse, nil, nil, nil, false, createdAt),
					sampleArrowRow(uuid.New(), archerID, 2, 2, model.ArrowStatusInUse, nil, nil, nil, false, createdAt),
					sampleArrowRow(uuid.New(), archerID, 2, 3, model.ArrowStatusInUse, nil, nil, nil, false, createdAt),
				},
			}, nil
		},
	}

	repo := repository.NewArrowRepo(mock)
	arrows, err := repo.FindByArcherAndSet(context.Background(), archerID, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arrows) != 3 {
		t.Fatalf("expected 3 arrows, got %d", len(arrows))
	}
	if !strings.Contains(executedSQL, "ORDER BY arrow_number ASC") {
		t.Errorf("expected ORDER BY arrow_number ASC in query: %s", executedSQL)
	}
	if len(executedArgs) != 3 || executedArgs[0] != archerID.String() || executedArgs[1] != int16(2) || executedArgs[2] != false {
		t.Errorf("expected args [%s, 2, false], got %v", archerID.String(), executedArgs)
	}
}

func TestArrowRepo_WithTx(t *testing.T) {
	repo := repository.NewArrowRepo(&mockDBTX{})
	txRepo := repo.WithTx(&mockTx{})
	if txRepo == nil {
		t.Fatal("expected non-nil txRepo")
	}
}
