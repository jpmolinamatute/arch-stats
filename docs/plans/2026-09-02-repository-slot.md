# Task 011: Build Repository — Slot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the slot repository (`SlotRepo`) for managing shooting slot assignments within a session (creation, retrieval by ID or session ID, filtering, dynamic updates, deletion, and session slot counting) using Squirrel SQL query building and `pgx/v5` against the PostgreSQL `slot` table.

**Architecture:** The `SlotRepo` operates on the `DBTX` interface (`*pgxpool.Pool` or `pgx.Tx`), building PostgreSQL parameterized queries with Squirrel targeting the `slot` table. It exposes CRUD operations (`Create`, `FindByID`, `FindBySessionID`, `FindAll`, `Update`, `Delete`) plus aggregate queries (`CountBySessionID`). Unit tests utilize mocked `DBTX` implementations to verify exact SQL generation, parameter binding, error wrapping, and defensive row scanning without requiring a live database.

**Tech Stack:** Go 1.27+, `github.com/Masterminds/squirrel`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`, standard library (`context`, `errors`, `fmt`, `time`).

**Spec:** [docs/go_refactor/tasks/011-repository_slot.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/011-repository_slot.md)

## Global Constraints

- Git branch: `refactor/011-repository-slot`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/repository`
- All SQL queries must use PostgreSQL dollar placeholder format (`$1`, `$2`) via `StmtBuilder` (`squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)`).
- No ORMs; use `pgx/v5` only via `DBTX`.
- Error handling: Wrap errors with `%w` using contextual descriptive messages (`fmt.Errorf("...: %w", err)`). Return sentinel `apperror.ErrNotFound` on 0 rows affected for mutations where an entity was expected.
- Schema alignment note: PostgreSQL migration `004_2025-09-26_shooting_sessions_table.sql` and domain models in `backend/internal/model/slot.go` define the `slot` table with columns: `slot_id`, `target_id`, `archer_id`, `session_id`, `slot_letter`, `face_type`, `bowstyle`, `draw_weight`, `club_id`, `is_shooting`, `shot_per_round`, `interval_seconds`, `created_at`.
  - In `011-repository_slot.md`, informal references to `slot_number` and `distance` correspond to the actual schema where `slot_letter` ('A', 'B', 'C', 'D') resides on `slot`, while `lane` and `distance` reside on the `target` table (`target_id`).
  - `FindBySessionID` queries the `slot` table filtered by `session_id = $1` ordered by `created_at ASC, slot_letter ASC`.
- Formatting must adhere to `gofumpt` and linting must pass `golangci-lint run ./...`.
- `go test -race ./internal/repository/... -v` must pass.
- `go vet ./...` must report no issues.

---

## File Structure

```
backend/
└── internal/
    └── repository/
        ├── base.go                    # [EXISTING] DBTX interface, StmtBuilder, ScanOne, ScanRows
        ├── archer.go                  # [EXISTING] Reference repository implementation
        ├── archer_test.go             # [MODIFY] Extend mockMultiRows to support SlotLetter, FaceType, and int scan targets
        ├── slot.go                    # [NEW] SlotRepo struct, NewSlotRepo, WithTx, FindByID, FindBySessionID, FindAll, Create, Update, Delete, CountBySessionID
        └── slot_test.go               # [NEW] Unit test suite covering all SlotRepo methods using mockDBTX
```

---

### Task 1: Git Branch Setup & Mock Scanner Update

**Files:**
- Modify: `backend/internal/repository/archer_test.go:145-155`

**Interfaces:**
- Consumes: `mockMultiRows` from `backend/internal/repository/archer_test.go`
- Produces: Enhanced `mockMultiRows.Scan` supporting destination types `*model.SlotLetter`, `*model.FaceType`, `*int`, `**int`, and `*int64`

- [ ] **Step 1: Check out git branch**

```bash
git switch -c refactor/011-repository-slot
```

- [ ] **Step 2: Add `SlotLetter`, `FaceType`, and `int` scanning to `mockMultiRows.Scan` in `archer_test.go`**

In `backend/internal/repository/archer_test.go`, update `mockMultiRows.Scan` around line 147 (before `case *any:`) to include:

```go
		case *model.SlotLetter:
			switch val := v.(type) {
			case model.SlotLetter:
				*d = val
			case string:
				*d = model.SlotLetter(val)
			}
		case *model.FaceType:
			switch val := v.(type) {
			case model.FaceType:
				*d = val
			case string:
				*d = model.FaceType(val)
			}
		case *int:
			switch val := v.(type) {
			case int:
				*d = val
			case int64:
				*d = int(val)
			}
		case **int:
			switch val := v.(type) {
			case nil:
				*d = nil
			case *int:
				*d = val
			case int:
				*d = &val
			}
		case *int64:
			switch val := v.(type) {
			case int64:
				*d = val
			case int:
				*d = int64(val)
			}
```

- [ ] **Step 3: Run existing repository tests to verify no regression**

Run: `cd backend && go test ./internal/repository/... -v`
Expected: PASS

- [ ] **Step 4: Commit mock scanner update**

```bash
git add backend/internal/repository/archer_test.go
git commit -m "test(repository): support slot enum and int scanning in mockMultiRows"
```

---

### Task 2: Slot Repository Unit Tests (TDD - Failing Tests)

**Files:**
- Create: `backend/internal/repository/slot_test.go`

**Interfaces:**
- Consumes:
  - `model.SlotCreate`, `model.SlotSet`, `model.SlotFilter`, `model.SlotRead`, `model.SlotLetter`, `model.FaceType`, `model.Bowstyle` from `internal/model`
  - `repository.NewSlotRepo(db DBTX)`
  - `mockDBTX`, `mockMultiRows` from `archer_test.go`
  - `mockSingleRow`, `mockTx` from `base_test.go`
- Produces:
  - Comprehensive unit test suite covering `FindByID`, `FindBySessionID`, `FindAll`, `Create`, `Update`, `Delete`, `CountBySessionID`, and `WithTx`.

- [ ] **Step 1: Write failing unit tests in `backend/internal/repository/slot_test.go`**

```go
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
	if len(executedArgs) != 1 || executedArgs[0] != slotID {
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
	if len(executedArgs) != 1 || executedArgs[0] != sessionID {
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
	if len(executedArgs) != 1 || executedArgs[0] != slotID {
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
	if len(executedArgs) != 1 || executedArgs[0] != sessionID {
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
```

- [ ] **Step 2: Run test suite to verify tests fail**

Run: `cd backend && go test ./internal/repository/... -v`
Expected: FAIL with compilation error (`undefined: repository.NewSlotRepo`)

---

### Task 3: Slot Repository Implementation (`slot.go`)

**Files:**
- Create: `backend/internal/repository/slot.go`

**Interfaces:**
- Consumes:
  - `DBTX`, `StmtBuilder`, `ScanOne`, `ScanRows` from `base.go`
  - `model.SlotCreate`, `model.SlotSet`, `model.SlotFilter`, `model.SlotRead` from `model`
  - `apperror.ErrNotFound` from `apperror`
- Produces:
  - `SlotRepo` struct
  - `NewSlotRepo(db DBTX) *SlotRepo`
  - `(r *SlotRepo) WithTx(tx pgx.Tx) *SlotRepo`
  - `(r *SlotRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error)`
  - `(r *SlotRepo) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error)`
  - `(r *SlotRepo) FindAll(ctx context.Context, filter model.SlotFilter) ([]model.SlotRead, error)`
  - `(r *SlotRepo) Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error)`
  - `(r *SlotRepo) Update(ctx context.Context, data model.SlotSet, filter model.SlotFilter) error`
  - `(r *SlotRepo) Delete(ctx context.Context, id uuid.UUID) error`
  - `(r *SlotRepo) CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int, error)`

- [ ] **Step 1: Write `backend/internal/repository/slot.go`**

```go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var slotColumns = []string{
	"slot_id",
	"target_id",
	"archer_id",
	"session_id",
	"slot_letter",
	"face_type",
	"bowstyle",
	"draw_weight",
	"club_id",
	"is_shooting",
	"shot_per_round",
	"interval_seconds",
	"created_at",
}

// SlotRepo manages database operations for shooting slot assignments within a session.
type SlotRepo struct {
	db DBTX
}

// NewSlotRepo constructs a SlotRepo backed by DBTX.
func NewSlotRepo(db DBTX) *SlotRepo {
	return &SlotRepo{db: db}
}

// WithTx returns a new SlotRepo bound to the given transaction.
func (r *SlotRepo) WithTx(tx pgx.Tx) *SlotRepo {
	return &SlotRepo{db: tx}
}

func scanSlot(scanner interface{ Scan(dest ...any) error }) (model.SlotRead, error) {
	var s model.SlotRead
	err := scanner.Scan(
		&s.SlotID,
		&s.TargetID,
		&s.ArcherID,
		&s.SessionID,
		&s.SlotLetter,
		&s.FaceType,
		&s.Bowstyle,
		&s.DrawWeight,
		&s.ClubID,
		&s.IsShooting,
		&s.ShotPerRound,
		&s.IntervalSeconds,
		&s.CreatedAt,
	)
	if err != nil {
		return model.SlotRead{}, err
	}
	return s, nil
}

// FindByID retrieves a slot assignment by primary key identifier.
// Returns nil, nil if no slot exists with the given ID.
func (r *SlotRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error) {
	sql, args, err := StmtBuilder.Select(slotColumns...).
		From("slot").
		Where(squirrel.Eq{"slot_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find slot by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.SlotRead, error) {
		return scanSlot(r)
	})
}

// FindBySessionID retrieves all slot assignments for a given session, ordered chronologically then by slot letter.
func (r *SlotRepo) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error) {
	sql, args, err := StmtBuilder.Select(slotColumns...).
		From("slot").
		Where(squirrel.Eq{"session_id": sessionID}).
		OrderBy("created_at ASC, slot_letter ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find slots by session id query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying slots by session id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.SlotRead, error) {
		return scanSlot(r)
	})
}

// FindAll queries all shooting slots matching the optional criteria in filter.
func (r *SlotRepo) FindAll(ctx context.Context, filter model.SlotFilter) ([]model.SlotRead, error) {
	q := StmtBuilder.Select(slotColumns...).
		From("slot").
		OrderBy("created_at ASC, slot_letter ASC")

	if filter.SlotID != nil {
		q = q.Where(squirrel.Eq{"slot_id": *filter.SlotID})
	}
	if filter.TargetID != nil {
		q = q.Where(squirrel.Eq{"target_id": *filter.TargetID})
	}
	if filter.ArcherID != nil {
		q = q.Where(squirrel.Eq{"archer_id": *filter.ArcherID})
	}
	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
	}
	if filter.SlotLetter != nil {
		q = q.Where(squirrel.Eq{"slot_letter": *filter.SlotLetter})
	}
	if filter.IsShooting != nil {
		q = q.Where(squirrel.Eq{"is_shooting": *filter.IsShooting})
	}
	if filter.ShotPerRound != nil {
		q = q.Where(squirrel.Eq{"shot_per_round": *filter.ShotPerRound})
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all slots query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying slots: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.SlotRead, error) {
		return scanSlot(r)
	})
}

// Create inserts a new slot assignment record and returns the generated UUID identifier.
func (r *SlotRepo) Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error) {
	sql, args, err := StmtBuilder.Insert("slot").
		Columns(
			"target_id",
			"archer_id",
			"session_id",
			"slot_letter",
			"face_type",
			"bowstyle",
			"draw_weight",
			"club_id",
			"is_shooting",
			"shot_per_round",
			"interval_seconds",
		).
		Values(
			data.TargetID,
			data.ArcherID,
			data.SessionID,
			data.SlotLetter,
			data.FaceType,
			data.Bowstyle,
			data.DrawWeight,
			data.ClubID,
			data.IsShooting,
			data.ShotPerRound,
			data.IntervalSeconds,
		).
		Suffix("RETURNING slot_id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create slot query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting slot: %w", err)
	}

	return newID, nil
}

// Update mutates slot assignment fields specified in data for rows matching filter.
// Requires at least one filter criterion to prevent unrestricted updates.
// Returns apperror.ErrNotFound if no slot assignment matched the filter.
func (r *SlotRepo) Update(ctx context.Context, data model.SlotSet, filter model.SlotFilter) error {
	q := StmtBuilder.Update("slot")
	setCount := 0

	if data.IsShooting != nil {
		q = q.Set("is_shooting", *data.IsShooting)
		setCount++
	}
	if data.FaceType != nil {
		q = q.Set("face_type", *data.FaceType)
		setCount++
	}
	if data.SlotLetter != nil {
		q = q.Set("slot_letter", *data.SlotLetter)
		setCount++
	}
	if data.ShotPerRound != nil {
		q = q.Set("shot_per_round", *data.ShotPerRound)
		setCount++
	}
	if data.IntervalSeconds != nil {
		q = q.Set("interval_seconds", *data.IntervalSeconds)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.SlotID != nil {
		q = q.Where(squirrel.Eq{"slot_id": *filter.SlotID})
		whereCount++
	}
	if filter.TargetID != nil {
		q = q.Where(squirrel.Eq{"target_id": *filter.TargetID})
		whereCount++
	}
	if filter.ArcherID != nil {
		q = q.Where(squirrel.Eq{"archer_id": *filter.ArcherID})
		whereCount++
	}
	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
		whereCount++
	}
	if filter.SlotLetter != nil {
		q = q.Where(squirrel.Eq{"slot_letter": *filter.SlotLetter})
		whereCount++
	}
	if filter.IsShooting != nil {
		q = q.Where(squirrel.Eq{"is_shooting": *filter.IsShooting})
		whereCount++
	}
	if filter.ShotPerRound != nil {
		q = q.Where(squirrel.Eq{"shot_per_round": *filter.ShotPerRound})
		whereCount++
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
		whereCount++
	}

	if whereCount == 0 {
		return errors.New("update requires at least one filter condition to prevent unrestricted updates")
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("building update query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing update: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// Delete removes a slot assignment by primary key identifier.
// Returns apperror.ErrNotFound if no slot assignment existed with the given ID.
func (r *SlotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("slot").
		Where(squirrel.Eq{"slot_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete slot query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete slot: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// CountBySessionID counts the total number of slot assignments for a given session.
func (r *SlotRepo) CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("slot").
		Where(squirrel.Eq{"session_id": sessionID}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count by session id query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting slots by session id: %w", err)
	}

	return count, nil
}
```

- [ ] **Step 2: Run unit tests to verify they pass**

Run: `cd backend && go test ./internal/repository/... -v`
Expected: PASS

- [ ] **Step 3: Commit repository implementation**

```bash
git add backend/internal/repository/slot.go backend/internal/repository/slot_test.go
git commit -m "feat(repository): implement slot repository with session relationship queries"
```

---

### Task 4: Full Verification, Formatting, & Build

**Files:**
- None (verification across existing and new files)

- [ ] **Step 1: Run formatting check with gofumpt**

Run: `cd backend && gofumpt -l -w internal/repository/`
Expected: Clean or auto-formatted

- [ ] **Step 2: Run golangci-lint**

Run: `cd backend && golangci-lint run ./internal/repository/...`
Expected: 0 errors

- [ ] **Step 3: Run race-enabled unit test suite**

Run: `cd backend && go test -race ./internal/repository/... -v`
Expected: PASS with 0 race warnings

- [ ] **Step 4: Run go vet and build**

Run: `cd backend && go vet ./... && go build ./...`
Expected: Clean exit code 0

- [ ] **Step 5: Run full backend test suite**

Run: `cd backend && go test ./... -v`
Expected: All tests pass

- [ ] **Step 6: Commit any formatting or lint fixes if needed**

```bash
git add -u
git commit -m "chore(repository): format and lint slot repository"
```
