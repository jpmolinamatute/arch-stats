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

func sampleSlotRow(
	slotID, targetID, archerID, sessionID uuid.UUID,
	slotLetter model.SlotLetter,
	faceType model.FaceType,
	bowstyle model.Bowstyle,
	drawWeight float64,
	clubID *uuid.UUID,
	isShooting bool,
	shotPerRound *int,
	intervalSeconds int,
	createdAt *time.Time,
) []any {
	if createdAt == nil {
		t := time.Now().Truncate(time.Second).UTC()
		createdAt = &t
	}
	return []any{
		slotID,
		targetID,
		archerID,
		sessionID,
		slotLetter,
		faceType,
		bowstyle,
		drawWeight,
		clubID,
		isShooting,
		shotPerRound,
		intervalSeconds,
		createdAt,
	}
}

func TestSlotRepo_FindByID_Success(t *testing.T) {
	slotID := uuid.New()
	targetID := uuid.New()
	archerID := uuid.New()
	sessionID := uuid.New()
	clubID := uuid.New()
	spr := 6
	createdAt := time.Now().Truncate(time.Second).UTC()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleSlotRow(
						slotID, targetID, archerID, sessionID,
						model.SlotLetterA, model.FaceTypeWA40Full, model.BowstyleBarebow,
						32.5, &clubID, true, &spr, 20, &createdAt,
					)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewSlotRepo(mock)
	slot, err := repo.FindByID(context.Background(), slotID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slot == nil {
		t.Fatal("expected slot, got nil")
	}

	if !strings.HasPrefix(executedSQL, "SELECT slot_id, target_id, archer_id, session_id, slot_letter, face_type, bowstyle, draw_weight, club_id, is_shooting, shot_per_round, interval_seconds, created_at FROM slot") {
		t.Errorf("unexpected query SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE slot_id = $1") {
		t.Errorf("expected WHERE slot_id = $1, got: %s", executedSQL)
	}
	if len(executedArgs) != 1 || (executedArgs[0] != slotID && executedArgs[0] != slotID.String()) {
		t.Errorf("expected arg %v, got %v", slotID, executedArgs)
	}

	if slot.SlotID != slotID {
		t.Errorf("expected SlotID %v, got %v", slotID, slot.SlotID)
	}
	if slot.TargetID != targetID {
		t.Errorf("expected TargetID %v, got %v", targetID, slot.TargetID)
	}
	if slot.ArcherID != archerID {
		t.Errorf("expected ArcherID %v, got %v", archerID, slot.ArcherID)
	}
	if slot.SessionID != sessionID {
		t.Errorf("expected SessionID %v, got %v", sessionID, slot.SessionID)
	}
	if slot.SlotLetter != model.SlotLetterA {
		t.Errorf("expected SlotLetter %v, got %v", model.SlotLetterA, slot.SlotLetter)
	}
	if slot.FaceType != model.FaceTypeWA40Full {
		t.Errorf("expected FaceType %v, got %v", model.FaceTypeWA40Full, slot.FaceType)
	}
	if slot.Bowstyle != model.BowstyleBarebow {
		t.Errorf("expected Bowstyle %v, got %v", model.BowstyleBarebow, slot.Bowstyle)
	}
	if slot.DrawWeight != 32.5 {
		t.Errorf("expected DrawWeight 32.5, got %v", slot.DrawWeight)
	}
	if slot.ClubID == nil || *slot.ClubID != clubID {
		t.Errorf("expected ClubID %v, got %v", clubID, slot.ClubID)
	}
	if !slot.IsShooting {
		t.Errorf("expected IsShooting true, got %v", slot.IsShooting)
	}
	if slot.ShotPerRound == nil || *slot.ShotPerRound != spr {
		t.Errorf("expected ShotPerRound %v, got %v", spr, slot.ShotPerRound)
	}
	if slot.IntervalSeconds != 20 {
		t.Errorf("expected IntervalSeconds 20, got %v", slot.IntervalSeconds)
	}
	if slot.CreatedAt == nil || !slot.CreatedAt.Equal(createdAt) {
		t.Errorf("expected CreatedAt %v, got %v", createdAt, slot.CreatedAt)
	}
}

func TestSlotRepo_FindByID_NotFound(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewSlotRepo(mock)
	slot, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slot != nil {
		t.Fatalf("expected nil slot on not found, got %+v", slot)
	}
}

func TestSlotRepo_FindByID_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return dbErr
				},
			}
		},
	}

	repo := repository.NewSlotRepo(mock)
	slot, err := repo.FindByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if slot != nil {
		t.Fatalf("expected nil slot on error, got %+v", slot)
	}
}

func TestSlotRepo_FindBySessionID_Success(t *testing.T) {
	sessionID := uuid.New()
	slotID1 := uuid.New()
	slotID2 := uuid.New()
	targetID := uuid.New()
	archerID1 := uuid.New()
	archerID2 := uuid.New()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			records := [][]any{
				sampleSlotRow(slotID1, targetID, archerID1, sessionID, model.SlotLetterA, model.FaceTypeWA40Full, model.BowstyleBarebow, 30.0, nil, true, nil, 20, nil),
				sampleSlotRow(slotID2, targetID, archerID2, sessionID, model.SlotLetterB, model.FaceTypeWA40Full, model.BowstyleCompound, 50.0, nil, true, nil, 20, nil),
			}
			return &mockMultiRows{records: records}, nil
		},
	}

	repo := repository.NewSlotRepo(mock)
	slots, err := repo.FindBySessionID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(slots))
	}

	if !strings.HasPrefix(executedSQL, "SELECT slot_id, target_id, archer_id, session_id, slot_letter, face_type, bowstyle, draw_weight, club_id, is_shooting, shot_per_round, interval_seconds, created_at FROM slot") {
		t.Errorf("unexpected query SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE session_id = $1") {
		t.Errorf("expected WHERE session_id = $1, got: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "ORDER BY created_at ASC, slot_letter ASC") {
		t.Errorf("expected ORDER BY created_at ASC, slot_letter ASC, got: %s", executedSQL)
	}
	if len(executedArgs) != 1 || (executedArgs[0] != sessionID && executedArgs[0] != sessionID.String()) {
		t.Errorf("expected arg %v, got %v", sessionID, executedArgs)
	}

	if slots[0].SlotID != slotID1 || slots[0].SlotLetter != model.SlotLetterA {
		t.Errorf("unexpected slot 0: %+v", slots[0])
	}
	if slots[1].SlotID != slotID2 || slots[1].SlotLetter != model.SlotLetterB {
		t.Errorf("unexpected slot 1: %+v", slots[1])
	}
}

func TestSlotRepo_FindBySessionID_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return nil, dbErr
		},
	}

	repo := repository.NewSlotRepo(mock)
	slots, err := repo.FindBySessionID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if slots != nil {
		t.Fatalf("expected nil slots on error, got %+v", slots)
	}
}

func TestSlotRepo_FindAll_WithFilters(t *testing.T) {
	sessionID := uuid.New()
	targetID := uuid.New()
	archerID := uuid.New()
	slotID := uuid.New()
	slotLetter := model.SlotLetterC
	isShooting := true
	spr := 3
	createdAt := time.Now().Truncate(time.Second).UTC()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			records := [][]any{
				sampleSlotRow(slotID, targetID, archerID, sessionID, slotLetter, model.FaceTypeWA40Full, model.BowstyleBarebow, 30.0, nil, isShooting, &spr, 20, &createdAt),
			}
			return &mockMultiRows{records: records}, nil
		},
	}

	repo := repository.NewSlotRepo(mock)
	filter := model.SlotFilter{
		SlotID:       &slotID,
		TargetID:     &targetID,
		ArcherID:     &archerID,
		SessionID:    &sessionID,
		SlotLetter:   &slotLetter,
		IsShooting:   &isShooting,
		ShotPerRound: &spr,
		CreatedAt:    &createdAt,
	}

	slots, err := repo.FindAll(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(slots))
	}

	for _, col := range []string{"slot_id", "target_id", "archer_id", "session_id", "slot_letter", "is_shooting", "shot_per_round", "created_at"} {
		if !strings.Contains(executedSQL, col) {
			t.Errorf("expected SQL to contain filter for %s, got: %s", col, executedSQL)
		}
	}
	if len(executedArgs) != 8 {
		t.Errorf("expected 8 args, got %d", len(executedArgs))
	}
}

func TestSlotRepo_FindAll_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return nil, dbErr
		},
	}

	repo := repository.NewSlotRepo(mock)
	slots, err := repo.FindAll(context.Background(), model.SlotFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if slots != nil {
		t.Fatalf("expected nil slots on error, got %+v", slots)
	}
}

func TestSlotRepo_Create_Success(t *testing.T) {
	newID := uuid.New()
	targetID := uuid.New()
	archerID := uuid.New()
	sessionID := uuid.New()
	clubID := uuid.New()
	spr := 5

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					if len(dest) == 1 {
						if p, ok := dest[0].(*uuid.UUID); ok {
							*p = newID
							return nil
						}
					}
					return errors.New("invalid scan destination")
				},
			}
		},
	}

	repo := repository.NewSlotRepo(mock)
	data := model.SlotCreate{
		ArcherID:        archerID,
		SessionID:       sessionID,
		TargetID:        targetID,
		SlotLetter:      model.SlotLetterB,
		FaceType:        model.FaceTypeWA80Full,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      40.0,
		ClubID:          &clubID,
		IsShooting:      true,
		ShotPerRound:    &spr,
		IntervalSeconds: 25,
	}

	id, err := repo.Create(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != newID {
		t.Fatalf("expected id %v, got %v", newID, id)
	}

	if !strings.HasPrefix(executedSQL, "INSERT INTO slot (") {
		t.Errorf("expected INSERT INTO slot, got: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "RETURNING slot_id") {
		t.Errorf("expected RETURNING slot_id, got: %s", executedSQL)
	}
	if len(executedArgs) != 11 {
		t.Errorf("expected 11 args, got %d: %v", len(executedArgs), executedArgs)
	}
}

func TestSlotRepo_Create_DBError(t *testing.T) {
	dbErr := errors.New("insert failed")
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return dbErr
				},
			}
		},
	}

	repo := repository.NewSlotRepo(mock)
	id, err := repo.Create(context.Background(), model.SlotCreate{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if id != uuid.Nil {
		t.Fatalf("expected nil uuid on error, got %v", id)
	}
}

func TestSlotRepo_Update_Success(t *testing.T) {
	slotID := uuid.New()
	isShooting := false
	faceType := model.FaceTypeWA60Full
	slotLetter := model.SlotLetterD
	spr := 4
	intervalSeconds := 30

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := repository.NewSlotRepo(mock)
	data := model.SlotSet{
		IsShooting:      &isShooting,
		FaceType:        &faceType,
		SlotLetter:      &slotLetter,
		ShotPerRound:    &spr,
		IntervalSeconds: &intervalSeconds,
	}
	filter := model.SlotFilter{
		SlotID: &slotID,
	}

	err := repo.Update(context.Background(), data, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(executedSQL, "UPDATE slot SET") {
		t.Errorf("expected UPDATE slot SET, got: %s", executedSQL)
	}
	for _, col := range []string{"is_shooting", "face_type", "slot_letter", "shot_per_round", "interval_seconds"} {
		if !strings.Contains(executedSQL, col) {
			t.Errorf("expected SQL to set %s, got: %s", col, executedSQL)
		}
	}
	if !strings.Contains(executedSQL, "WHERE slot_id = $6") {
		t.Errorf("expected WHERE slot_id = $6, got: %s", executedSQL)
	}
	if len(executedArgs) != 6 {
		t.Errorf("expected 6 args, got %d: %v", len(executedArgs), executedArgs)
	}
}

func TestSlotRepo_Update_RowsAffectedZeroReturnsNotFound(t *testing.T) {
	slotID := uuid.New()
	isShooting := false

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}

	repo := repository.NewSlotRepo(mock)
	err := repo.Update(context.Background(), model.SlotSet{IsShooting: &isShooting}, model.SlotFilter{SlotID: &slotID})
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected apperror.ErrNotFound, got: %v", err)
	}
}

func TestSlotRepo_Update_EmptySetReturnsNil(t *testing.T) {
	slotID := uuid.New()
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			t.Fatal("exec should not be called on empty set")
			return pgconn.CommandTag{}, nil
		},
	}

	repo := repository.NewSlotRepo(mock)
	err := repo.Update(context.Background(), model.SlotSet{}, model.SlotFilter{SlotID: &slotID})
	if err != nil {
		t.Fatalf("expected nil on empty update set, got: %v", err)
	}
}

func TestSlotRepo_Update_MissingFilterReturnsError(t *testing.T) {
	isShooting := true
	mock := &mockDBTX{}

	repo := repository.NewSlotRepo(mock)
	err := repo.Update(context.Background(), model.SlotSet{IsShooting: &isShooting}, model.SlotFilter{})
	if err == nil {
		t.Fatal("expected error on missing filter condition, got nil")
	}
	if !strings.Contains(err.Error(), "at least one filter condition") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSlotRepo_Update_DBError(t *testing.T) {
	slotID := uuid.New()
	isShooting := false
	dbErr := errors.New("update exec error")

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, dbErr
		},
	}

	repo := repository.NewSlotRepo(mock)
	err := repo.Update(context.Background(), model.SlotSet{IsShooting: &isShooting}, model.SlotFilter{SlotID: &slotID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
}

func TestSlotRepo_Delete_Success(t *testing.T) {
	slotID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("DELETE 1"), nil
		},
	}

	repo := repository.NewSlotRepo(mock)
	err := repo.Delete(context.Background(), slotID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if executedSQL != "DELETE FROM slot WHERE slot_id = $1" {
		t.Errorf("unexpected delete SQL: %s", executedSQL)
	}
	if len(executedArgs) != 1 || (executedArgs[0] != slotID && executedArgs[0] != slotID.String()) {
		t.Errorf("expected arg %v, got %v", slotID, executedArgs)
	}
}

func TestSlotRepo_Delete_RowsAffectedZeroReturnsNotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("DELETE 0"), nil
		},
	}

	repo := repository.NewSlotRepo(mock)
	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected apperror.ErrNotFound, got: %v", err)
	}
}

func TestSlotRepo_Delete_DBError(t *testing.T) {
	dbErr := errors.New("delete failed")
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, dbErr
		},
	}

	repo := repository.NewSlotRepo(mock)
	err := repo.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
}

func TestSlotRepo_CountBySessionID_Success(t *testing.T) {
	sessionID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					if len(dest) == 1 {
						if p, ok := dest[0].(*int); ok {
							*p = 4
							return nil
						}
					}
					return errors.New("invalid scan destination")
				},
			}
		},
	}

	repo := repository.NewSlotRepo(mock)
	count, err := repo.CountBySessionID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected count 4, got %d", count)
	}

	if !strings.HasPrefix(executedSQL, "SELECT COUNT(*) FROM slot") {
		t.Errorf("unexpected query SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE session_id = $1") {
		t.Errorf("expected WHERE session_id = $1, got: %s", executedSQL)
	}
	if len(executedArgs) != 1 || (executedArgs[0] != sessionID && executedArgs[0] != sessionID.String()) {
		t.Errorf("expected arg %v, got %v", sessionID, executedArgs)
	}
}

func TestSlotRepo_CountBySessionID_DBError(t *testing.T) {
	dbErr := errors.New("count failed")
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return dbErr
				},
			}
		},
	}

	repo := repository.NewSlotRepo(mock)
	count, err := repo.CountBySessionID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0 on error, got %d", count)
	}
}

func TestSlotRepo_WithTx(t *testing.T) {
	mock := &mockDBTX{}
	repo := repository.NewSlotRepo(mock)

	tx := &mockTx{}
	repoWithTx := repo.WithTx(tx)
	if repoWithTx == nil {
		t.Fatal("expected non-nil repoWithTx")
	}
	if repoWithTx == repo {
		t.Errorf("expected new instance from WithTx")
	}
}
