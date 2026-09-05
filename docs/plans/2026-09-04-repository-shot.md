# Task 012: Build Repository — Shot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the shot repository (`ShotRepo`) for managing arrow shot records (creation, batch creation, retrieval by ID, slot-based queries, filtering, dynamic updates, deletion, slot shot count, and latest shot timestamp) using Squirrel SQL query building and `pgx/v5` against the PostgreSQL `shot` table.

**Architecture:** The `ShotRepo` operates on the `DBTX` interface (`*pgxpool.Pool` or `pgx.Tx`), assembling parameterized PostgreSQL queries with Squirrel targeting the `shot` table. It exposes CRUD operations (`Create`, `CreateBatch`, `FindByID`, `FindBySlotID`, `FindAll`, `Update`, `Delete`) and aggregate queries (`CountBySlotID`, `GetLatestShotTime`). Unit tests utilize mocked `DBTX` implementations to verify exact SQL generation, placeholder parameter binding, error wrapping with `%w`, and defensive row scanning without requiring a live database.

**Tech Stack:** Go 1.27+, `github.com/Masterminds/squirrel`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`, standard library (`context`, `errors`, `fmt`, `time`).

**Spec:** [docs/go_refactor/tasks/012-repository_shot.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/012-repository_shot.md)

## Global Constraints

- Git branch: `refactor/012-repository-shot`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/repository`
- All SQL queries must use PostgreSQL dollar placeholder format (`$1`, `$2`) via `StmtBuilder` (`squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)`).
- No ORMs; use `pgx/v5` only via `DBTX`.
- Error handling: Wrap errors with `%w` using contextual descriptive messages (`fmt.Errorf("...: %w", err)`). Return sentinel `apperror.ErrNotFound` on 0 rows affected for mutations where an entity was expected.
- Schema alignment note: PostgreSQL migration `006_2025-10-28_shot_table.sql` and domain models in `backend/internal/model/shot.go` define the `shot` table with columns: `shot_id`, `slot_id`, `x`, `y`, `score`, `arrow_id`, `created_at`, `is_x`.
  - In `012-repository_shot.md`, informal references to `arrow_number`, `x_position`, `y_position` correspond to the actual schema columns `arrow_id`, `x`, `y`, `is_x`, and `score`.
  - `FindBySlotID` queries the `shot` table filtered by `slot_id = $1` ordered by `created_at ASC`.
- Formatting must adhere to `gofumpt` and linting must pass `golangci-lint run ./...`.
- `go test -race ./internal/repository/... -v` must pass.
- `go vet ./...` must report no issues.

---

## File Structure

```
backend/
└── internal/
    ├── model/
    │   └── shot.go                    # [MODIFY] Add mutable fields (X, Y, Score, IsX, ArrowID) to ShotSet
    └── repository/
        ├── base.go                    # [EXISTING] DBTX interface, StmtBuilder, ScanOne, ScanRows
        ├── archer.go                  # [EXISTING] Reference repository implementation
        ├── archer_test.go             # [MODIFY] Extend mockMultiRows.Scan to support **float64
        ├── shot.go                    # [NEW] ShotRepo struct, NewShotRepo, WithTx, FindByID, FindBySlotID, FindAll, Create, CreateBatch, Update, Delete, CountBySlotID, GetLatestShotTime
        └── shot_test.go               # [NEW] Unit test suite covering all ShotRepo methods using mockDBTX
```

---

### Task 1: Git Branch Setup & Model / Mock Scanner Adjustments

**Files:**
- Modify: `backend/internal/model/shot.go:21-23`
- Modify: `backend/internal/repository/archer_test.go:114-116`

**Interfaces:**
- Consumes: `model.ShotSet` from `backend/internal/model/shot.go`
- Produces: Enhanced `model.ShotSet` with mutable coordinate and scoring fields (`X`, `Y`, `Score`, `IsX`, `ArrowID`) and enhanced `mockMultiRows.Scan` supporting `**float64` scan destinations.

- [ ] **Step 1: Check out git branch**

```bash
git switch -c refactor/012-repository-shot
```

- [ ] **Step 2: Update `ShotSet` in `backend/internal/model/shot.go`**

In `backend/internal/model/shot.go`, replace lines 21-22:

```go
// ShotSet represents updates to mutable shot fields.
type ShotSet struct {
	X       *float64   `json:"x,omitempty"`
	Y       *float64   `json:"y,omitempty"`
	IsX     *bool      `json:"is_x,omitempty"`
	Score   *int       `json:"score,omitempty" validate:"omitempty,gte=0,lte=10"`
	ArrowID *uuid.UUID `json:"arrow_id,omitempty"`
}
```

- [ ] **Step 3: Add `**float64` scanning support in `backend/internal/repository/archer_test.go`**

In `backend/internal/repository/archer_test.go`, update `mockMultiRows.Scan` around line 114 to add `case **float64:`:

```go
		case *float64:
			*d = v.(float64)
		case **float64:
			switch val := v.(type) {
			case nil:
				*d = nil
			case *float64:
				*d = val
			case float64:
				*d = &val
			}
```

- [ ] **Step 4: Run existing repository tests to verify no regressions**

```bash
cd backend && go test ./internal/model/... ./internal/repository/... -v
```
Expected: PASS with all existing tests green.

- [ ] **Step 5: Commit changes**

```bash
git add backend/internal/model/shot.go backend/internal/repository/archer_test.go
git commit -m "refactor: add mutable fields to ShotSet and support float64 pointer scanning"
```

---

### Task 2: Core Read Methods (`FindByID`, `FindBySlotID`, `FindAll`)

**Files:**
- Create: `backend/internal/repository/shot_test.go`
- Create: `backend/internal/repository/shot.go`

**Interfaces:**
- Consumes: `DBTX`, `StmtBuilder`, `ScanOne`, `ScanRows` from `backend/internal/repository/base.go`; `model.ShotRead`, `model.ShotFilter` from `backend/internal/model/shot.go`
- Produces:
  - `NewShotRepo(db DBTX) *ShotRepo`
  - `(r *ShotRepo) WithTx(tx pgx.Tx) *ShotRepo`
  - `(r *ShotRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error)`
  - `(r *ShotRepo) FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error)`
  - `(r *ShotRepo) FindAll(ctx context.Context, filter model.ShotFilter) ([]model.ShotRead, error)`

- [ ] **Step 1: Write failing unit tests for read methods in `shot_test.go`**

Create `backend/internal/repository/shot_test.go`:

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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/repository/shot_test.go -v
```
Expected: FAIL with compilation error `undefined: repository.NewShotRepo`.

- [ ] **Step 3: Implement read methods in `backend/internal/repository/shot.go`**

Create `backend/internal/repository/shot.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var shotColumns = []string{
	"shot_id",
	"slot_id",
	"x",
	"y",
	"is_x",
	"score",
	"arrow_id",
	"created_at",
}

// ShotRepo manages database operations for arrow shots within a shooting slot.
type ShotRepo struct {
	db DBTX
}

// NewShotRepo constructs a ShotRepo backed by DBTX.
func NewShotRepo(db DBTX) *ShotRepo {
	return &ShotRepo{db: db}
}

// WithTx returns a new ShotRepo bound to the given transaction.
func (r *ShotRepo) WithTx(tx pgx.Tx) *ShotRepo {
	return &ShotRepo{db: tx}
}

func scanShot(scanner interface{ Scan(dest ...any) error }) (model.ShotRead, error) {
	var s model.ShotRead
	err := scanner.Scan(
		&s.ShotID,
		&s.SlotID,
		&s.X,
		&s.Y,
		&s.IsX,
		&s.Score,
		&s.ArrowID,
		&s.CreatedAt,
	)
	if err != nil {
		return model.ShotRead{}, err
	}
	return s, nil
}

// FindByID retrieves a shot record by its primary key identifier.
// Returns nil, nil if no shot exists with the given ID.
func (r *ShotRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error) {
	sql, args, err := StmtBuilder.Select(shotColumns...).
		From("shot").
		Where(squirrel.Eq{"shot_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find shot by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.ShotRead, error) {
		return scanShot(r)
	})
}

// FindBySlotID retrieves all shot records for a given slot, ordered chronologically.
func (r *ShotRepo) FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error) {
	sql, args, err := StmtBuilder.Select(shotColumns...).
		From("shot").
		Where(squirrel.Eq{"slot_id": slotID}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find shots by slot id query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying shots by slot id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ShotRead, error) {
		return scanShot(r)
	})
}

// FindAll queries all shots matching the optional criteria in filter, ordered chronologically.
func (r *ShotRepo) FindAll(ctx context.Context, filter model.ShotFilter) ([]model.ShotRead, error) {
	q := StmtBuilder.Select(shotColumns...).
		From("shot").
		OrderBy("created_at ASC")

	if filter.ShotID != nil {
		q = q.Where(squirrel.Eq{"shot_id": *filter.ShotID})
	}
	if filter.SlotID != nil {
		q = q.Where(squirrel.Eq{"slot_id": *filter.SlotID})
	}
	if filter.X != nil {
		q = q.Where(squirrel.Eq{"x": *filter.X})
	}
	if filter.Y != nil {
		q = q.Where(squirrel.Eq{"y": *filter.Y})
	}
	if filter.Score != nil {
		q = q.Where(squirrel.Eq{"score": *filter.Score})
	}
	if filter.ArrowID != nil {
		q = q.Where(squirrel.Eq{"arrow_id": *filter.ArrowID})
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all shots query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying shots: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ShotRead, error) {
		return scanShot(r)
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/repository/ -run "TestShotRepo_Find" -v
```
Expected: PASS with `TestShotRepo_FindByID_Success`, `TestShotRepo_FindByID_NotFoundReturnsNil`, `TestShotRepo_FindByID_QueryError`, `TestShotRepo_FindBySlotID_Success`, `TestShotRepo_FindAll_WithFilter` all passing.

- [ ] **Step 5: Commit changes**

```bash
git add backend/internal/repository/shot.go backend/internal/repository/shot_test.go
git commit -m "feat: implement shot repository read methods (FindByID, FindBySlotID, FindAll)"
```

---

### Task 3: Mutation Methods (`Create`, `CreateBatch`, `Update`, `Delete`)

**Files:**
- Modify: `backend/internal/repository/shot_test.go`
- Modify: `backend/internal/repository/shot.go`

**Interfaces:**
- Consumes: `model.ShotCreate`, `model.ShotSet`, `model.ShotFilter` from `backend/internal/model/shot.go`; `apperror.ErrNotFound` from `backend/internal/apperror`
- Produces:
  - `(r *ShotRepo) Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error)`
  - `(r *ShotRepo) CreateBatch(ctx context.Context, data []model.ShotCreate) ([]uuid.UUID, error)`
  - `(r *ShotRepo) Update(ctx context.Context, data model.ShotSet, filter model.ShotFilter) error`
  - `(r *ShotRepo) Delete(ctx context.Context, id uuid.UUID) error`

- [ ] **Step 1: Write failing tests for mutation methods in `shot_test.go`**

Append to `backend/internal/repository/shot_test.go`:

```go
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
					*(dest[0].(*uuid.UUID)) = expectedID
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

	if !strings.HasPrefix(executedSQL, "INSERT INTO shot (slot_id, x, y, is_x, score, arrow_id, created_at)") {
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
					*(dest[0].(*uuid.UUID)) = expectedID
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

	if !strings.HasPrefix(executedSQL, "INSERT INTO shot (slot_id, x, y, is_x, score, arrow_id)") {
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

	if !strings.HasPrefix(executedSQL, "INSERT INTO shot (slot_id, x, y, is_x, score, arrow_id, created_at) VALUES") {
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
	if len(executedArgs) != 1 || executedArgs[0] != shotID {
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/repository/ -run "TestShotRepo_Create|TestShotRepo_Update|TestShotRepo_Delete" -v
```
Expected: FAIL with undefined `Create`, `CreateBatch`, `Update`, `Delete`.

- [ ] **Step 3: Implement mutation methods in `backend/internal/repository/shot.go`**

Append to `backend/internal/repository/shot.go`:

```go
// Create inserts a new shot record and returns the generated UUID identifier.
func (r *ShotRepo) Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error) {
	insertBuilder := StmtBuilder.Insert("shot")
	if data.CreatedAt != nil {
		insertBuilder = insertBuilder.
			Columns("slot_id", "x", "y", "is_x", "score", "arrow_id", "created_at").
			Values(data.SlotID, data.X, data.Y, data.IsX, data.Score, data.ArrowID, *data.CreatedAt)
	} else {
		insertBuilder = insertBuilder.
			Columns("slot_id", "x", "y", "is_x", "score", "arrow_id").
			Values(data.SlotID, data.X, data.Y, data.IsX, data.Score, data.ArrowID)
	}

	sql, args, err := insertBuilder.Suffix("RETURNING shot_id").ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create shot query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting shot: %w", err)
	}

	return newID, nil
}

// CreateBatch inserts a batch of shot records in a single query and returns their generated UUIDs.
func (r *ShotRepo) CreateBatch(ctx context.Context, data []model.ShotCreate) ([]uuid.UUID, error) {
	if len(data) == 0 {
		return nil, nil
	}

	b := StmtBuilder.Insert("shot").
		Columns("slot_id", "x", "y", "is_x", "score", "arrow_id", "created_at")

	for _, s := range data {
		var createdAt any
		if s.CreatedAt != nil {
			createdAt = *s.CreatedAt
		} else {
			createdAt = squirrel.Expr("now()")
		}
		b = b.Values(s.SlotID, s.X, s.Y, s.IsX, s.Score, s.ArrowID, createdAt)
	}

	sql, args, err := b.Suffix("RETURNING shot_id").ToSql()
	if err != nil {
		return nil, fmt.Errorf("building batch create shot query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("executing batch create shot: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (uuid.UUID, error) {
		var id uuid.UUID
		err := r.Scan(&id)
		return id, err
	})
}

// Update mutates shot fields specified in data for rows matching filter.
// Requires at least one filter criterion to prevent unrestricted updates.
// Returns apperror.ErrNotFound if no shot matched the filter.
func (r *ShotRepo) Update(ctx context.Context, data model.ShotSet, filter model.ShotFilter) error {
	q := StmtBuilder.Update("shot")
	setCount := 0

	if data.X != nil {
		q = q.Set("x", *data.X)
		setCount++
	}
	if data.Y != nil {
		q = q.Set("y", *data.Y)
		setCount++
	}
	if data.IsX != nil {
		q = q.Set("is_x", *data.IsX)
		setCount++
	}
	if data.Score != nil {
		q = q.Set("score", *data.Score)
		setCount++
	}
	if data.ArrowID != nil {
		q = q.Set("arrow_id", *data.ArrowID)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.ShotID != nil {
		q = q.Where(squirrel.Eq{"shot_id": *filter.ShotID})
		whereCount++
	}
	if filter.SlotID != nil {
		q = q.Where(squirrel.Eq{"slot_id": *filter.SlotID})
		whereCount++
	}
	if filter.X != nil {
		q = q.Where(squirrel.Eq{"x": *filter.X})
		whereCount++
	}
	if filter.Y != nil {
		q = q.Where(squirrel.Eq{"y": *filter.Y})
		whereCount++
	}
	if filter.Score != nil {
		q = q.Where(squirrel.Eq{"score": *filter.Score})
		whereCount++
	}
	if filter.ArrowID != nil {
		q = q.Where(squirrel.Eq{"arrow_id": *filter.ArrowID})
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
		return fmt.Errorf("building update shot query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing update shot: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// Delete removes a shot record by its primary key identifier.
// Returns apperror.ErrNotFound if no shot existed with the given ID.
func (r *ShotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("shot").
		Where(squirrel.Eq{"shot_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete shot query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete shot: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/repository/ -run "TestShotRepo_Create|TestShotRepo_Update|TestShotRepo_Delete" -v
```
Expected: PASS for all creation, update, and deletion test cases.

- [ ] **Step 5: Commit changes**

```bash
git add backend/internal/repository/shot.go backend/internal/repository/shot_test.go
git commit -m "feat: implement shot repository mutation methods (Create, CreateBatch, Update, Delete)"
```

---

### Task 4: Aggregate Queries (`CountBySlotID`, `GetLatestShotTime`)

**Files:**
- Modify: `backend/internal/repository/shot_test.go`
- Modify: `backend/internal/repository/shot.go`

**Interfaces:**
- Consumes: `DBTX`, `StmtBuilder`, `ScanOne` from `backend/internal/repository/base.go`
- Produces:
  - `(r *ShotRepo) CountBySlotID(ctx context.Context, slotID uuid.UUID) (int, error)`
  - `(r *ShotRepo) GetLatestShotTime(ctx context.Context, slotID uuid.UUID) (*time.Time, error)`

- [ ] **Step 1: Write failing tests for aggregate queries in `shot_test.go`**

Append to `backend/internal/repository/shot_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/repository/ -run "TestShotRepo_CountBySlotID|TestShotRepo_GetLatestShotTime" -v
```
Expected: FAIL with undefined `CountBySlotID`, `GetLatestShotTime`.

- [ ] **Step 3: Implement aggregate methods in `backend/internal/repository/shot.go`**

Append to `backend/internal/repository/shot.go`:

```go
// CountBySlotID counts the total number of shots recorded for a given slot.
func (r *ShotRepo) CountBySlotID(ctx context.Context, slotID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("shot").
		Where(squirrel.Eq{"slot_id": slotID}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count shots by slot id query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting shots by slot id: %w", err)
	}

	return count, nil
}

// GetLatestShotTime retrieves the timestamp of the most recently recorded shot for a given slot.
// Returns nil, nil if no shots exist for the slot.
func (r *ShotRepo) GetLatestShotTime(ctx context.Context, slotID uuid.UUID) (*time.Time, error) {
	sql, args, err := StmtBuilder.Select("created_at").
		From("shot").
		Where(squirrel.Eq{"slot_id": slotID}).
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building get latest shot time query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (time.Time, error) {
		var t time.Time
		err := r.Scan(&t)
		return t, err
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./internal/repository/ -run "TestShotRepo_CountBySlotID|TestShotRepo_GetLatestShotTime" -v
```
Expected: PASS for `CountBySlotID` and `GetLatestShotTime` test cases.

- [ ] **Step 5: Commit changes**

```bash
git add backend/internal/repository/shot.go backend/internal/repository/shot_test.go
git commit -m "feat: implement shot repository aggregate methods (CountBySlotID, GetLatestShotTime)"
```

---

### Task 5: End-to-End Verification, Code Quality & Task Checklist Update

**Files:**
- Modify: `docs/go_refactor/tasks/012-repository_shot.md:20-30,45-82`

**Interfaces:**
- Consumes: Complete repository test suite and linting tools
- Produces: Verified codebase, clean linter reports, and updated task status documentation

- [ ] **Step 1: Run complete repository test suite with race detection**

```bash
cd backend && go test -race ./internal/repository/... -v
```
Expected: PASS for all tests across all repository files.

- [ ] **Step 2: Run `go vet` and `go build`**

```bash
cd backend && go vet ./... && go build ./...
```
Expected: Zero warnings or compilation issues.

- [ ] **Step 3: Run full Go linter and formatter**

```bash
cd backend && golangci-lint run ./...
```
Expected: Zero lint violations.

- [ ] **Step 4: Update `docs/go_refactor/tasks/012-repository_shot.md` checklist**

Update the acceptance criteria and steps in `docs/go_refactor/tasks/012-repository_shot.md` to mark all checkboxes `- [x]` as completed.

- [ ] **Step 5: Commit completion**

```bash
git add docs/go_refactor/tasks/012-repository_shot.md
git commit -m "docs: mark task 012 repository shot as complete"
```
