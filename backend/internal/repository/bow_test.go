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

func sampleBowRow(id, archerID uuid.UUID, name string, style model.Bowstyle, weight float64, isDeleted bool, createdAt time.Time) []any {
	return []any{
		id,
		archerID,
		name,
		style,
		weight,
		isDeleted,
		createdAt,
	}
}

func TestBowRepo_Create_Success(t *testing.T) {
	expectedID := uuid.New()
	archerID := uuid.New()
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

	repo := repository.NewBowRepo(mock)
	payload := model.BowCreate{
		ArcherID:   archerID,
		Name:       "Formula Xi",
		Bowstyle:   model.BowstyleRecurve,
		DrawWeight: 38.5,
	}

	id, err := repo.Create(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != expectedID {
		t.Errorf("expected id %v, got %v", expectedID, id)
	}
	if !strings.Contains(executedSQL, "INSERT INTO bow") {
		t.Errorf("expected INSERT INTO bow in query: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "RETURNING bow_id") {
		t.Errorf("expected RETURNING bow_id in query: %s", executedSQL)
	}
	if len(executedArgs) != 4 {
		t.Fatalf("expected 4 query args, got %d", len(executedArgs))
	}
	if executedArgs[0] != archerID {
		t.Errorf("expected arg[0] %v, got %v", archerID, executedArgs[0])
	}
	if executedArgs[1] != "Formula Xi" {
		t.Errorf("expected arg[1] Formula Xi, got %v", executedArgs[1])
	}
	if executedArgs[2] != model.BowstyleRecurve {
		t.Errorf("expected arg[2] %v, got %v", model.BowstyleRecurve, executedArgs[2])
	}
	if executedArgs[3] != 38.5 {
		t.Errorf("expected arg[3] 38.5, got %v", executedArgs[3])
	}
}

func TestBowRepo_FindByID_Success(t *testing.T) {
	bowID := uuid.New()
	archerID := uuid.New()
	createdAt := time.Now().Truncate(time.Second).UTC()

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleBowRow(bowID, archerID, "Hoyt Invicta", model.BowstyleCompound, 55.0, false, createdAt)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewBowRepo(mock)
	bow, err := repo.FindByID(context.Background(), bowID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bow == nil {
		t.Fatal("expected bow, got nil")
	}
	if bow.BowID != bowID {
		t.Errorf("expected bowID %v, got %v", bowID, bow.BowID)
	}
	if bow.ArcherID != archerID {
		t.Errorf("expected archerID %v, got %v", archerID, bow.ArcherID)
	}
	if bow.Name != "Hoyt Invicta" {
		t.Errorf("expected name Hoyt Invicta, got %s", bow.Name)
	}
	if bow.Bowstyle != model.BowstyleCompound {
		t.Errorf("expected bowstyle compound, got %v", bow.Bowstyle)
	}
	if bow.DrawWeight != 55.0 {
		t.Errorf("expected draw weight 55.0, got %v", bow.DrawWeight)
	}
	if bow.IsDeleted {
		t.Error("expected is_deleted false, got true")
	}
}

func TestBowRepo_FindByID_NotFound(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewBowRepo(mock)
	bow, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bow != nil {
		t.Errorf("expected nil bow on ErrNoRows, got %v", bow)
	}
}

func TestBowRepo_FindAllByArcherID_Success(t *testing.T) {
	archerID := uuid.New()
	bow1ID := uuid.New()
	bow2ID := uuid.New()
	createdAt := time.Now().Truncate(time.Second).UTC()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			return &mockMultiRows{
				records: [][]any{
					sampleBowRow(bow1ID, archerID, "Bow 1", model.BowstyleRecurve, 36.0, false, createdAt),
					sampleBowRow(bow2ID, archerID, "Bow 2", model.BowstyleBarebow, 32.0, false, createdAt.Add(time.Hour)),
				},
			}, nil
		},
	}

	repo := repository.NewBowRepo(mock)
	bows, err := repo.FindAllByArcherID(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bows) != 2 {
		t.Fatalf("expected 2 bows, got %d", len(bows))
	}
	if !strings.Contains(executedSQL, "ORDER BY created_at ASC") {
		t.Errorf("expected ORDER BY created_at ASC in query: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "is_deleted =") {
		t.Errorf("expected is_deleted filter in query: %s", executedSQL)
	}
	if len(executedArgs) != 2 || executedArgs[0] != archerID.String() || executedArgs[1] != false {
		t.Errorf("expected args [%s, false], got %v", archerID.String(), executedArgs)
	}
}

func TestBowRepo_CountByArcherID_Success(t *testing.T) {
	archerID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*dest[0].(*int) = 3
					return nil
				},
			}
		},
	}

	repo := repository.NewBowRepo(mock)
	count, err := repo.CountByArcherID(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
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

func TestBowRepo_WithTx(t *testing.T) {
	repo := repository.NewBowRepo(&mockDBTX{})
	txRepo := repo.WithTx(&mockTx{})
	if txRepo == nil {
		t.Fatal("expected non-nil txRepo")
	}
}
