# Repository Layer for Auth Identity, Bow, Arrow, and Archer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement database access repositories using Squirrel query building and pgx/v5 for the separated `auth` identity table, equipment tables (`bow` and `arrow`), and update the `archer` repository to require `ArcherID` provided from `auth`.

**Architecture:** Add `BowRepo` in `internal/repository/bow.go`, `ArrowRepo` in `internal/repository/arrow.go`, `AuthIdentityRepo` in `internal/repository/auth_identity.go`, and refactor `ArcherRepo.Create` in `internal/repository/archer.go` to operate over the `DBTX` interface using Squirrel parameterized queries ($1, $2) and transaction support via `WithTx(tx pgx.Tx)`. Write exhaustive query building and execution unit tests using mocked `DBTX` in `bow_test.go`, `arrow_test.go`, `auth_identity_test.go`, and `archer_test.go`.

**Tech Stack:** Go 1.27+, `github.com/Masterminds/squirrel`, `github.com/jackc/pgx/v5`, `github.com/google/uuid`, standard library `context`, `time`, `errors`, `fmt`.

**Spec:** [docs/onboarding_refactor/tasks/003_repository_layer.md](../onboarding_refactor/tasks/003_repository_layer.md)

## Global Constraints

- Repository package is strictly `backend/internal/repository`.
- Database operations must operate against the `DBTX` interface (`Query`, `QueryRow`, `Exec`).
- All repositories must provide transaction propagation via `WithTx(tx pgx.Tx) *Repo`.
- Parameterized SQL placeholder format must strictly use PostgreSQL dollar format (`squirrel.Dollar` / `StmtBuilder`).
- All errors must be wrapped with contextual messages using `%w` or standard `apperror` sentinels (`apperror.ErrNotFound`).
- Verification commands:
  - `cd backend && go test ./internal/repository/... -v` must exit 0.
  - `cd backend && go vet ./internal/repository/...` must report 0 issues.
  - `cd backend && golangci-lint run ./internal/repository/...` must report 0 issues.
- All steps must follow strict Test-Driven Development (TDD): write failing tests first, verify failure, implement minimal code, verify pass, commit.
- Mark all acceptance criteria and step checkboxes in `docs/onboarding_refactor/tasks/003_repository_layer.md` as done (`[x]`) upon completion.
- Maintain a table-only live tracker in `docs/plans/task.md` throughout execution.

---

### Task 1: Setup Task Tracking & Git Branch

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: `main`
- Produces: Branch `feature/003-repository-layer` and live progress tracker in `docs/plans/task.md`

- [x] **Step 1: Create live progress tracker in `docs/plans/task.md`**

Update `docs/plans/task.md` with table format:

```markdown
# Live Task Tracker: 003 Repository Layer

| Task # | Task Description | Status | Verification |
|---|---|---|---|
| Task 1 | Setup Task Tracking & Git Branch | In Progress | git branch checked |
| Task 2 | Test Support Enhancements (`archer_test.go`) | Pending | `go test ./internal/repository/...` passes |
| Task 3 | Bow Repository (`bow.go` & `bow_test.go`) | Pending | `go test ./internal/repository/... -run TestBowRepo` |
| Task 4 | Arrow Repository (`arrow.go` & `arrow_test.go`) | Pending | `go test ./internal/repository/... -run TestArrowRepo` |
| Task 5 | Auth Identity Repository (`auth_identity.go` & `auth_identity_test.go`) | Pending | `go test ./internal/repository/... -run TestAuthIdentityRepo` |
| Task 6 | Archer Repository Refactor (`archer.go` & `archer_test.go`) | Pending | `go test ./internal/repository/... -run TestArcherRepo` |
| Task 7 | Package Verification & Quality Assurance | Pending | `go test`, `go vet`, `golangci-lint` clean |
| Task 8 | Mark Task 003 Documentation & Live Tracker as Completed | Pending | checklist review |
```

- [x] **Step 2: Create and switch to git branch `feature/003-repository-layer`**

Run:
```bash
git switch -c feature/003-repository-layer
```
Verify:
```bash
git branch --show-current
```
Expected output: `feature/003-repository-layer`

- [x] **Step 3: Update `docs/plans/task.md`**

Update status of Task 1 to `Done`.

---

### Task 2: Test Support Enhancements in `archer_test.go`

**Files:**
- Modify: `backend/internal/repository/archer_test.go:70-198`

**Interfaces:**
- Consumes: `mockMultiRows` in `backend/internal/repository/archer_test.go`
- Produces: Support in `mockMultiRows.Scan` for `*int16` and `*model.ArrowStatus` destinations so subsequent arrow tests can scan mock rows without panic or zero values.

- [x] **Step 1: Write test case verifying `mockMultiRows.Scan` handles `*int16` and `*model.ArrowStatus`**

Add a test in `backend/internal/repository/archer_test.go`:

```go
func TestMockMultiRows_ScanArrowTypes(t *testing.T) {
	setVal := int16(2)
	statusVal := model.ArrowStatusInUse
	mr := &mockMultiRows{
		records: [][]any{
			{setVal, statusVal},
		},
	}
	if !mr.Next() {
		t.Fatal("expected next row")
	}

	var scannedSet int16
	var scannedStatus model.ArrowStatus
	if err := mr.Scan(&scannedSet, &scannedStatus); err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}
	if scannedSet != 2 {
		t.Errorf("expected scannedSet 2, got %d", scannedSet)
	}
	if scannedStatus != model.ArrowStatusInUse {
		t.Errorf("expected scannedStatus %q, got %q", model.ArrowStatusInUse, scannedStatus)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestMockMultiRows_ScanArrowTypes
```
Expected output: FAIL (`expected scannedSet 2, got 0` or `expected scannedStatus "in_use", got ""`).

- [x] **Step 3: Update `mockMultiRows.Scan` in `archer_test.go`**

In `backend/internal/repository/archer_test.go`, add `*int16` and `*model.ArrowStatus` handling to `mockMultiRows.Scan`:

```go
		case *int16:
			switch val := v.(type) {
			case int16:
				*d = val
			case int:
				*d = int16(val)
			}
		case *model.ArrowStatus:
			*d = v.(model.ArrowStatus)
```

- [x] **Step 4: Run test to verify it passes**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestMockMultiRows_ScanArrowTypes
```
Expected output: PASS

- [x] **Step 5: Commit**

```bash
git add backend/internal/repository/archer_test.go
git commit -m "test(repo): add int16 and arrow status scan support to mock multi rows"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 2 to `Done`.

---

### Task 3: Bow Repository (`bow.go` & `bow_test.go`)

**Files:**
- Create: `backend/internal/repository/bow_test.go`
- Create: `backend/internal/repository/bow.go`

**Interfaces:**
- Consumes:
  - `model.BowCreate`, `model.BowRead` from `backend/internal/model/bow.go`
  - `DBTX`, `StmtBuilder`, `ScanRows`, `ScanOne`, `findByID`, `createReturningID` from `backend/internal/repository/base.go`
- Produces:
  - `BowRepo` struct
  - `NewBowRepo(db DBTX) *BowRepo`
  - `(r *BowRepo) WithTx(tx pgx.Tx) *BowRepo`
  - `(r *BowRepo) Create(ctx context.Context, data model.BowCreate) (uuid.UUID, error)`
  - `(r *BowRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.BowRead, error)`
  - `(r *BowRepo) FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.BowRead, error)`
  - `(r *BowRepo) CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)`

- [x] **Step 1: Write failing unit tests in `bow_test.go`**

Create `backend/internal/repository/bow_test.go`:

```go
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
					*(dest[0].(*uuid.UUID)) = expectedID
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
	if !strings.Contains(executedSQL, "is_deleted = false") && !strings.Contains(executedSQL, "is_deleted = FALSE") {
		t.Errorf("expected is_deleted filter in query: %s", executedSQL)
	}
	if len(executedArgs) != 1 || executedArgs[0] != archerID {
		t.Errorf("expected archerID arg %v, got %v", archerID, executedArgs)
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
					*(dest[0].(*int)) = 3
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
	if !strings.Contains(executedSQL, "is_deleted = false") && !strings.Contains(executedSQL, "is_deleted = FALSE") {
		t.Errorf("expected is_deleted filter in query: %s", executedSQL)
	}
	if len(executedArgs) != 1 || executedArgs[0] != archerID {
		t.Errorf("expected archerID arg %v, got %v", archerID, executedArgs)
	}
}

func TestBowRepo_WithTx(t *testing.T) {
	repo := repository.NewBowRepo(&mockDBTX{})
	txRepo := repo.WithTx(&mockTx{})
	if txRepo == nil {
		t.Fatal("expected non-nil txRepo")
	}
}
```

- [x] **Step 2: Run test to verify failure**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestBowRepo
```
Expected output: Compilation failure (`undefined: repository.NewBowRepo`).

- [x] **Step 3: Implement `bow.go`**

Create `backend/internal/repository/bow.go`:

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

var bowColumns = []string{
	"bow_id",
	"archer_id",
	"name",
	"bowstyle",
	"draw_weight",
	"is_deleted",
	"created_at",
}

// BowRepo manages database operations for bow equipment inventory.
type BowRepo struct {
	db DBTX
}

// NewBowRepo constructs a BowRepo backed by DBTX.
func NewBowRepo(db DBTX) *BowRepo {
	return &BowRepo{db: db}
}

// WithTx returns a new BowRepo bound to the given transaction.
func (r *BowRepo) WithTx(tx pgx.Tx) *BowRepo {
	return &BowRepo{db: tx}
}

func scanBow(scanner interface{ Scan(dest ...any) error }) (model.BowRead, error) {
	var b model.BowRead
	err := scanner.Scan(
		&b.BowID,
		&b.ArcherID,
		&b.Name,
		&b.Bowstyle,
		&b.DrawWeight,
		&b.IsDeleted,
		&b.CreatedAt,
	)
	if err != nil {
		return model.BowRead{}, err
	}
	return b, nil
}

// Create inserts a new bow equipment record, returning the generated UUID.
func (r *BowRepo) Create(ctx context.Context, data model.BowCreate) (uuid.UUID, error) {
	cols := []string{"archer_id", "name", "bowstyle", "draw_weight"}
	vals := []any{data.ArcherID, data.Name, data.Bowstyle, data.DrawWeight}

	builder := StmtBuilder.Insert("bow").
		Columns(cols...).
		Values(vals...)

	return createReturningID(ctx, r.db, builder, "bow_id")
}

// FindByID retrieves a bow by primary key identifier.
func (r *BowRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.BowRead, error) {
	return findByID(ctx, r.db, "bow", "bow_id", bowColumns, id, func(row pgx.Row) (model.BowRead, error) {
		return scanBow(row)
	})
}

// FindAllByArcherID retrieves all active bows belonging to the specified archer ordered by creation time.
func (r *BowRepo) FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.BowRead, error) {
	sql, args, err := StmtBuilder.Select(bowColumns...).
		From("bow").
		Where(squirrel.Eq{"archer_id": archerID}).
		Where(squirrel.Eq{"is_deleted": false}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all bows query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying bows: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.BowRead, error) {
		return scanBow(r)
	})
}

// CountByArcherID counts active (non-deleted) bows owned by an archer.
func (r *BowRepo) CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("bow").
		Where(squirrel.Eq{"archer_id": archerID}).
		Where(squirrel.Eq{"is_deleted": false}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count bows query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting bows by archer id: %w", err)
	}

	return count, nil
}
```

- [x] **Step 4: Run tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestBowRepo
```
Expected output: PASS (all 6 tests passing).

- [x] **Step 5: Commit**

```bash
git add backend/internal/repository/bow.go backend/internal/repository/bow_test.go
git commit -m "feat(repo): implement bow repository and unit tests"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 3 to `Done`.

---

### Task 4: Arrow Repository (`arrow.go` & `arrow_test.go`)

**Files:**
- Create: `backend/internal/repository/arrow_test.go`
- Create: `backend/internal/repository/arrow.go`

**Interfaces:**
- Consumes:
  - `model.ArrowCreate`, `model.ArrowBatchCreate`, `model.ArrowRead`, `model.ArrowStatusInUse` from `backend/internal/model/arrow.go` & `backend/internal/model/enums.go`
  - `DBTX`, `StmtBuilder`, `ScanRows`, `ScanOne`, `findByID`, `createReturningID` from `backend/internal/repository/base.go`
- Produces:
  - `ArrowRepo` struct
  - `NewArrowRepo(db DBTX) *ArrowRepo`
  - `(r *ArrowRepo) WithTx(tx pgx.Tx) *ArrowRepo`
  - `(r *ArrowRepo) Create(ctx context.Context, data model.ArrowCreate) (uuid.UUID, error)`
  - `(r *ArrowRepo) CreateBatch(ctx context.Context, data model.ArrowBatchCreate) ([]model.ArrowRead, error)`
  - `(r *ArrowRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ArrowRead, error)`
  - `(r *ArrowRepo) FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.ArrowRead, error)`
  - `(r *ArrowRepo) CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)`
  - `(r *ArrowRepo) CountInUseByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)`
  - `(r *ArrowRepo) FindByArcherAndSet(ctx context.Context, archerID uuid.UUID, set int16) ([]model.ArrowRead, error)`

- [x] **Step 1: Write failing unit tests in `arrow_test.go`**

Create `backend/internal/repository/arrow_test.go`:

```go
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
					*(dest[0].(*uuid.UUID)) = expectedID
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
					*(dest[0].(*uuid.UUID)) = expectedID
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
	if !strings.Contains(executedSQL, "is_deleted = false") && !strings.Contains(executedSQL, "is_deleted = FALSE") {
		t.Errorf("expected is_deleted filter in query: %s", executedSQL)
	}
	if len(executedArgs) != 1 || executedArgs[0] != archerID {
		t.Errorf("expected archerID arg %v, got %v", archerID, executedArgs)
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
					*(dest[0].(*int)) = 6
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
	if !strings.Contains(executedSQL, "is_deleted = false") && !strings.Contains(executedSQL, "is_deleted = FALSE") {
		t.Errorf("expected is_deleted filter in query: %s", executedSQL)
	}
	if len(executedArgs) != 1 || executedArgs[0] != archerID {
		t.Errorf("expected archerID arg %v, got %v", archerID, executedArgs)
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
					*(dest[0].(*int)) = 5
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
	if len(executedArgs) != 2 || executedArgs[0] != archerID || executedArgs[1] != model.ArrowStatusInUse {
		t.Errorf("expected args [archerID, in_use], got %v", executedArgs)
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
	if len(executedArgs) != 2 || executedArgs[0] != archerID || executedArgs[1] != int16(2) {
		t.Errorf("expected args [archerID, 2], got %v", executedArgs)
	}
}

func TestArrowRepo_WithTx(t *testing.T) {
	repo := repository.NewArrowRepo(&mockDBTX{})
	txRepo := repo.WithTx(&mockTx{})
	if txRepo == nil {
		t.Fatal("expected non-nil txRepo")
	}
}
```

- [x] **Step 2: Run test to verify failure**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestArrowRepo
```
Expected output: Compilation failure (`undefined: repository.NewArrowRepo`).

- [x] **Step 3: Implement `arrow.go`**

Create `backend/internal/repository/arrow.go`:

```go
package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var arrowColumns = []string{
	"arrow_id",
	"archer_id",
	"arrow_set",
	"arrow_number",
	"status",
	"spine",
	"length",
	"weight",
	"is_deleted",
	"created_at",
}

// ArrowRepo manages database operations for arrow equipment inventory.
type ArrowRepo struct {
	db DBTX
}

// NewArrowRepo constructs an ArrowRepo backed by DBTX.
func NewArrowRepo(db DBTX) *ArrowRepo {
	return &ArrowRepo{db: db}
}

// WithTx returns a new ArrowRepo bound to the given transaction.
func (r *ArrowRepo) WithTx(tx pgx.Tx) *ArrowRepo {
	return &ArrowRepo{db: tx}
}

func scanArrow(scanner interface{ Scan(dest ...any) error }) (model.ArrowRead, error) {
	var a model.ArrowRead
	err := scanner.Scan(
		&a.ArrowID,
		&a.ArcherID,
		&a.ArrowSet,
		&a.ArrowNumber,
		&a.Status,
		&a.Spine,
		&a.Length,
		&a.Weight,
		&a.IsDeleted,
		&a.CreatedAt,
	)
	if err != nil {
		return model.ArrowRead{}, err
	}
	return a, nil
}

// Create inserts a single arrow record, defaulting status to 'in_use' if omitted.
func (r *ArrowRepo) Create(ctx context.Context, data model.ArrowCreate) (uuid.UUID, error) {
	status := model.ArrowStatusInUse
	if data.Status != nil {
		status = *data.Status
	}

	cols := []string{"archer_id", "arrow_set", "arrow_number", "status", "spine", "length", "weight"}
	vals := []any{data.ArcherID, data.ArrowSet, data.ArrowNumber, status, data.Spine, data.Length, data.Weight}

	builder := StmtBuilder.Insert("arrow").
		Columns(cols...).
		Values(vals...)

	return createReturningID(ctx, r.db, builder, "arrow_id")
}

// CreateBatch inserts multiple arrow records in a single query, defaulting status to 'in_use'.
func (r *ArrowRepo) CreateBatch(ctx context.Context, data model.ArrowBatchCreate) ([]model.ArrowRead, error) {
	if data.Count <= 0 {
		return nil, nil
	}

	status := model.ArrowStatusInUse
	if data.Status != nil {
		status = *data.Status
	}

	cols := []string{"archer_id", "arrow_set", "arrow_number", "status", "spine", "length", "weight"}
	builder := StmtBuilder.Insert("arrow").Columns(cols...)

	for i := 1; i <= data.Count; i++ {
		builder = builder.Values(
			data.ArcherID,
			data.ArrowSet,
			int16(i),
			status,
			data.Spine,
			data.Length,
			data.Weight,
		)
	}

	builder = builder.Suffix("RETURNING " + strings.Join(arrowColumns, ", "))
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building create batch arrow query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("inserting arrow batch: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ArrowRead, error) {
		return scanArrow(r)
	})
}

// FindByID retrieves an arrow by primary key identifier.
func (r *ArrowRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ArrowRead, error) {
	return findByID(ctx, r.db, "arrow", "arrow_id", arrowColumns, id, func(row pgx.Row) (model.ArrowRead, error) {
		return scanArrow(row)
	})
}

// FindAllByArcherID retrieves all active arrows for an archer ordered by set and number.
func (r *ArrowRepo) FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.ArrowRead, error) {
	sql, args, err := StmtBuilder.Select(arrowColumns...).
		From("arrow").
		Where(squirrel.Eq{
			"archer_id":  archerID,
			"is_deleted": false,
		}).
		OrderBy("arrow_set ASC, arrow_number ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all arrows query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying arrows by archer id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ArrowRead, error) {
		return scanArrow(r)
	})
}

// CountByArcherID counts all active arrows registered by an archer.
func (r *ArrowRepo) CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("arrow").
		Where(squirrel.Eq{
			"archer_id":  archerID,
			"is_deleted": false,
		}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count arrows query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting arrows by archer id: %w", err)
	}

	return count, nil
}

// CountInUseByArcherID counts active arrows with status 'in_use' registered by an archer.
func (r *ArrowRepo) CountInUseByArcherID(ctx context.Context, archerID uuid.UUID) (int, error) {
	sql, args, err := StmtBuilder.Select("COUNT(*)").
		From("arrow").
		Where(squirrel.Eq{
			"archer_id":  archerID,
			"status":     model.ArrowStatusInUse,
			"is_deleted": false,
		}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count in use arrows query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting in use arrows: %w", err)
	}

	return count, nil
}

// FindByArcherAndSet retrieves all active arrows in a specific set for an archer ordered by arrow number.
func (r *ArrowRepo) FindByArcherAndSet(ctx context.Context, archerID uuid.UUID, set int16) ([]model.ArrowRead, error) {
	sql, args, err := StmtBuilder.Select(arrowColumns...).
		From("arrow").
		Where(squirrel.Eq{
			"archer_id":  archerID,
			"arrow_set":  set,
			"is_deleted": false,
		}).
		OrderBy("arrow_number ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find arrows by set query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying arrows by set: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ArrowRead, error) {
		return scanArrow(r)
	})
}
```

- [x] **Step 4: Run tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestArrowRepo
```
Expected output: PASS (all 11 tests passing).

- [x] **Step 5: Commit**

```bash
git add backend/internal/repository/arrow.go backend/internal/repository/arrow_test.go
git commit -m "feat(repo): implement arrow repository and unit tests"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 4 to `Done`.

---

### Task 5: Auth Identity Repository (`auth_identity.go` & `auth_identity_test.go`)

**Files:**
- Create: `backend/internal/repository/auth_identity_test.go`
- Create: `backend/internal/repository/auth_identity.go`

**Interfaces:**
- Consumes:
  - `model.AuthIdentityRead` from `backend/internal/model/auth.go`
  - `DBTX`, `StmtBuilder`, `ScanOne`, `findByID`, `createReturningID`, `execUpdate` from `backend/internal/repository/base.go`
- Produces:
  - `AuthIdentityRepo` struct
  - `NewAuthIdentityRepo(db DBTX) *AuthIdentityRepo`
  - `(r *AuthIdentityRepo) WithTx(tx pgx.Tx) *AuthIdentityRepo`
  - `(r *AuthIdentityRepo) FindByGoogleSubject(ctx context.Context, sub string) (*model.AuthIdentityRead, error)`
  - `(r *AuthIdentityRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.AuthIdentityRead, error)`
  - `(r *AuthIdentityRepo) Create(ctx context.Context, subject string, picture *string) (uuid.UUID, error)`
  - `(r *AuthIdentityRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, login time.Time, pic *string) error`

- [x] **Step 1: Write failing unit tests in `auth_identity_test.go`**

Create `backend/internal/repository/auth_identity_test.go`:

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

func sampleAuthIdentityRow(id uuid.UUID, subject string, picture *string, lastLogin, createdAt time.Time) []any {
	return []any{
		id,
		subject,
		picture,
		lastLogin,
		createdAt,
	}
}

func TestAuthIdentityRepo_FindByGoogleSubject_Success(t *testing.T) {
	archerID := uuid.New()
	sub := "google-subject-123456"
	pic := "https://lh3.googleusercontent.com/avatar.jpg"
	now := time.Now().Truncate(time.Second).UTC()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleAuthIdentityRow(archerID, sub, &pic, now, now)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewAuthIdentityRepo(mock)
	auth, err := repo.FindByGoogleSubject(context.Background(), sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth == nil {
		t.Fatal("expected auth identity, got nil")
	}
	if auth.ArcherID != archerID {
		t.Errorf("expected archerID %v, got %v", archerID, auth.ArcherID)
	}
	if auth.GoogleSubject != sub {
		t.Errorf("expected subject %s, got %s", sub, auth.GoogleSubject)
	}
	if auth.GooglePictureURL == nil || *auth.GooglePictureURL != pic {
		t.Errorf("expected picture url %s, got %v", pic, auth.GooglePictureURL)
	}
	if !strings.Contains(executedSQL, "FROM auth") {
		t.Errorf("expected FROM auth in query: %s", executedSQL)
	}
	if len(executedArgs) != 1 || executedArgs[0] != sub {
		t.Errorf("expected sub arg %s, got %v", sub, executedArgs)
	}
}

func TestAuthIdentityRepo_FindByGoogleSubject_NotFound(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewAuthIdentityRepo(mock)
	auth, err := repo.FindByGoogleSubject(context.Background(), "unknown-sub")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth != nil {
		t.Errorf("expected nil on ErrNoRows, got %v", auth)
	}
}

func TestAuthIdentityRepo_FindByID_Success(t *testing.T) {
	archerID := uuid.New()
	sub := "google-subject-789"
	now := time.Now().Truncate(time.Second).UTC()

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleAuthIdentityRow(archerID, sub, nil, now, now)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewAuthIdentityRepo(mock)
	auth, err := repo.FindByID(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth == nil {
		t.Fatal("expected auth identity, got nil")
	}
	if auth.ArcherID != archerID {
		t.Errorf("expected archerID %v, got %v", archerID, auth.ArcherID)
	}
	if auth.GoogleSubject != sub {
		t.Errorf("expected subject %s, got %s", sub, auth.GoogleSubject)
	}
	if auth.GooglePictureURL != nil {
		t.Errorf("expected nil picture url, got %v", auth.GooglePictureURL)
	}
}

func TestAuthIdentityRepo_FindByID_NotFound(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewAuthIdentityRepo(mock)
	auth, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth != nil {
		t.Errorf("expected nil on ErrNoRows, got %v", auth)
	}
}

func TestAuthIdentityRepo_Create_WithPicture(t *testing.T) {
	expectedID := uuid.New()
	sub := "google-sub-new"
	pic := "https://example.com/pic.png"
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

	repo := repository.NewAuthIdentityRepo(mock)
	id, err := repo.Create(context.Background(), sub, &pic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != expectedID {
		t.Errorf("expected id %v, got %v", expectedID, id)
	}
	if !strings.Contains(executedSQL, "INSERT INTO auth") {
		t.Errorf("expected INSERT INTO auth in query: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "RETURNING archer_id") {
		t.Errorf("expected RETURNING archer_id in query: %s", executedSQL)
	}
	if len(executedArgs) != 2 || executedArgs[0] != sub || executedArgs[1] != pic {
		t.Errorf("expected args [%s, %s], got %v", sub, pic, executedArgs)
	}
}

func TestAuthIdentityRepo_Create_WithoutPicture(t *testing.T) {
	expectedID := uuid.New()
	sub := "google-sub-nopic"
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*(dest[0].(*uuid.UUID)) = expectedID
					return nil
				},
			}
		},
	}

	repo := repository.NewAuthIdentityRepo(mock)
	id, err := repo.Create(context.Background(), sub, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != expectedID {
		t.Errorf("expected id %v, got %v", expectedID, id)
	}
	if len(executedArgs) != 1 || executedArgs[0] != sub {
		t.Errorf("expected single arg [%s], got %v", sub, executedArgs)
	}
}

func TestAuthIdentityRepo_UpdateLastLogin_Success(t *testing.T) {
	archerID := uuid.New()
	loginTime := time.Now().Truncate(time.Second).UTC()
	pic := "https://example.com/new-pic.png"
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := repository.NewAuthIdentityRepo(mock)
	err := repo.UpdateLastLogin(context.Background(), archerID, loginTime, &pic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(executedSQL, "UPDATE auth") {
		t.Errorf("expected UPDATE auth in query: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "last_login_at") || !strings.Contains(executedSQL, "google_picture_url") {
		t.Errorf("expected columns updated in query: %s", executedSQL)
	}
	if len(executedArgs) != 3 {
		t.Fatalf("expected 3 args, got %d", len(executedArgs))
	}
}

func TestAuthIdentityRepo_UpdateLastLogin_NotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}

	repo := repository.NewAuthIdentityRepo(mock)
	err := repo.UpdateLastLogin(context.Background(), uuid.New(), time.Now(), nil)
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when 0 rows affected, got: %v", err)
	}
}

func TestAuthIdentityRepo_WithTx(t *testing.T) {
	repo := repository.NewAuthIdentityRepo(&mockDBTX{})
	txRepo := repo.WithTx(&mockTx{})
	if txRepo == nil {
		t.Fatal("expected non-nil txRepo")
	}
}
```

- [x] **Step 2: Run test to verify failure**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestAuthIdentityRepo
```
Expected output: Compilation failure (`undefined: repository.NewAuthIdentityRepo`).

- [x] **Step 3: Implement `auth_identity.go`**

Create `backend/internal/repository/auth_identity.go`:

```go
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var authIdentityColumns = []string{
	"archer_id",
	"google_subject",
	"google_picture_url",
	"last_login_at",
	"created_at",
}

// AuthIdentityRepo manages database operations for OAuth authentication credentials.
type AuthIdentityRepo struct {
	db DBTX
}

// NewAuthIdentityRepo constructs an AuthIdentityRepo backed by DBTX.
func NewAuthIdentityRepo(db DBTX) *AuthIdentityRepo {
	return &AuthIdentityRepo{db: db}
}

// WithTx returns a new AuthIdentityRepo bound to the given transaction.
func (r *AuthIdentityRepo) WithTx(tx pgx.Tx) *AuthIdentityRepo {
	return &AuthIdentityRepo{db: tx}
}

func scanAuthIdentity(scanner interface{ Scan(dest ...any) error }) (model.AuthIdentityRead, error) {
	var a model.AuthIdentityRead
	err := scanner.Scan(
		&a.ArcherID,
		&a.GoogleSubject,
		&a.GooglePictureURL,
		&a.LastLoginAt,
		&a.CreatedAt,
	)
	if err != nil {
		return model.AuthIdentityRead{}, err
	}
	return a, nil
}

// FindByGoogleSubject retrieves an auth record by Google account subject claim.
func (r *AuthIdentityRepo) FindByGoogleSubject(ctx context.Context, sub string) (*model.AuthIdentityRead, error) {
	sql, args, err := StmtBuilder.Select(authIdentityColumns...).
		From("auth").
		Where(squirrel.Eq{"google_subject": sub}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find auth by google subject query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.AuthIdentityRead, error) {
		return scanAuthIdentity(r)
	})
}

// FindByID retrieves an auth identity by primary key identifier.
func (r *AuthIdentityRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.AuthIdentityRead, error) {
	return findByID(ctx, r.db, "auth", "archer_id", authIdentityColumns, id, func(row pgx.Row) (model.AuthIdentityRead, error) {
		return scanAuthIdentity(row)
	})
}

// Create inserts a new auth identity credential record returning the generated archer_id UUID.
func (r *AuthIdentityRepo) Create(ctx context.Context, subject string, picture *string) (uuid.UUID, error) {
	cols := []string{"google_subject"}
	vals := []any{subject}
	if picture != nil {
		cols = append(cols, "google_picture_url")
		vals = append(vals, *picture)
	}

	builder := StmtBuilder.Insert("auth").
		Columns(cols...).
		Values(vals...)

	return createReturningID(ctx, r.db, builder, "archer_id")
}

// UpdateLastLogin updates the timestamp of most recent sign-in and optionally refreshes the avatar URL.
func (r *AuthIdentityRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, login time.Time, pic *string) error {
	q := StmtBuilder.Update("auth").
		Set("last_login_at", login).
		Where(squirrel.Eq{"archer_id": id})
	if pic != nil {
		q = q.Set("google_picture_url", *pic)
	}
	return execUpdate(ctx, r.db, q)
}
```

- [x] **Step 4: Run tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestAuthIdentityRepo
```
Expected output: PASS (all 9 tests passing).

- [x] **Step 5: Commit**

```bash
git add backend/internal/repository/auth_identity.go backend/internal/repository/auth_identity_test.go
git commit -m "feat(repo): implement auth identity repository and unit tests"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 5 to `Done`.

---

### Task 6: Archer Repository Refactor & Test Hardening (`archer.go` & `archer_test.go`)

**Files:**
- Modify: `backend/internal/repository/archer.go:157-182`
- Modify: `backend/internal/repository/archer_test.go:367-397`

**Interfaces:**
- Consumes:
  - `model.ArcherCreate` with required `ArcherID` provided from `auth`
- Produces:
  - Refactored `(r *ArcherRepo) Create` strictly requiring `data.ArcherID` without mock auth identity creation fallback.

- [x] **Step 1: Update `archer_test.go` with test asserting `Create` requires `ArcherID` and update success test**

In `backend/internal/repository/archer_test.go`, update `TestArcherRepo_Create_Success` to supply `ArcherID` and add `TestArcherRepo_Create_MissingArcherID`:

```go
func TestArcherRepo_Create_Success(t *testing.T) {
	generatedID := uuid.New()
	authArcherID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					id := dest[0].(*uuid.UUID)
					*id = generatedID
					return nil
				},
			}
		},
	}

	repo := repository.NewArcherRepo(mock)
	createPayload := model.ArcherCreate{
		ArcherID:    &authArcherID,
		FirstName:   "Robin",
		LastName:    "Hood",
		Email:       "robin@sherwood.org",
		DateOfBirth: "1990-01-15",
		Gender:      model.GenderMale,
	}

	id, err := repo.Create(context.Background(), createPayload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != generatedID {
		t.Errorf("expected generated id %v, got %v", generatedID, id)
	}
	if !strings.Contains(executedSQL, "INSERT INTO archer") {
		t.Errorf("expected INSERT INTO archer in query: %s", executedSQL)
	}
	if len(executedArgs) != 6 || executedArgs[0] != authArcherID {
		t.Errorf("expected first arg authArcherID %v, got %v", authArcherID, executedArgs)
	}
}

func TestArcherRepo_Create_MissingArcherID(t *testing.T) {
	repo := repository.NewArcherRepo(&mockDBTX{})
	createPayload := model.ArcherCreate{
		ArcherID:    nil,
		FirstName:   "Robin",
		LastName:    "Hood",
		Email:       "robin@sherwood.org",
		DateOfBirth: "1990-01-15",
		Gender:      model.GenderMale,
	}

	_, err := repo.Create(context.Background(), createPayload)
	if err == nil {
		t.Fatal("expected error when archer_id is nil, got nil")
	}
	if !strings.Contains(err.Error(), "archer_id is required") {
		t.Errorf("expected error message mentioning archer_id is required, got: %v", err)
	}
}
```

- [x] **Step 2: Run test to verify `TestArcherRepo_Create_MissingArcherID` fails**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestArcherRepo_Create_MissingArcherID
```
Expected output: FAIL (`expected error when archer_id is nil, got nil`).

- [x] **Step 3: Update `archer.go` `Create` method**

In `backend/internal/repository/archer.go`, ensure `"errors"` is imported and update `Create`:

```go
// Create inserts a new archer row, returning the generated UUID.
//
//nolint:gocritic // hugeParam: data value parameter matches repository interface specification
func (r *ArcherRepo) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	dob, err := time.Parse("2006-01-02", data.DateOfBirth)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing date_of_birth: %w", err)
	}

	if data.ArcherID == nil || *data.ArcherID == uuid.Nil {
		return uuid.Nil, errors.New("creating archer: archer_id is required from auth identity")
	}
	archerID := *data.ArcherID

	cols := []string{"archer_id", "first_name", "last_name", "email", "date_of_birth", "gender"}
	vals := []any{archerID, data.FirstName, data.LastName, data.Email, dob, data.Gender}

	builder := StmtBuilder.Insert("archer").
		Columns(cols...).
		Values(vals...)

	return createReturningID(ctx, r.db, builder, "archer_id")
}
```

- [x] **Step 4: Run all `ArcherRepo` unit tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/repository/... -v -run TestArcherRepo
```
Expected output: PASS (all tests passing).

- [x] **Step 5: Commit**

```bash
git add backend/internal/repository/archer.go backend/internal/repository/archer_test.go
git commit -m "refactor(repo): enforce archer_id requirement from auth identity in archer repo Create"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 6 to `Done`.

---

### Task 7: Full Repository Package Verification & Quality Assurance

**Files:**
- None (verification across package)

**Interfaces:**
- Consumes: all files in `backend/internal/repository`
- Produces: verified test suite, clean vet, and clean linting reports

- [x] **Step 1: Run complete repository test suite**

Run:
```bash
cd backend && go test -race ./internal/repository/... -v
```
Expected output: PASS with code 0.

- [x] **Step 2: Run `go vet`**

Run:
```bash
cd backend && go vet ./internal/repository/...
```
Expected output: clean (no issues).

- [x] **Step 3: Run `golangci-lint`**

Run:
```bash
cd backend && golangci-lint run ./internal/repository/...
```
Expected output: clean (0 issues).

- [x] **Step 4: Run Go linting script**

Run:
```bash
./scripts/linting.bash --go
```
Expected output: all linters passed.

- [x] **Step 5: Update `docs/plans/task.md`**

Update status of Task 7 to `Done`.

---

### Task 8: Documentation & Checklist Completion

**Files:**
- Modify: `docs/onboarding_refactor/tasks/003_repository_layer.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: completed work from Tasks 1-7
- Produces: updated documentation with all acceptance criteria and steps checked off (`[x]`)

- [x] **Step 1: Update `docs/onboarding_refactor/tasks/003_repository_layer.md`**

Mark each Acceptance Criteria item as done (`[x]`):
- [x] New repository `backend/internal/repository/bow.go` implements:
    - [x] `Create(ctx context.Context, data model.BowCreate) (uuid.UUID, error)`
    - [x] `FindByID(ctx context.Context, id uuid.UUID) (*model.BowRead, error)`
    - [x] `FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.BowRead, error)`
    - [x] `CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)`
- [x] New repository `backend/internal/repository/arrow.go` implements:
    - [x] `Create(ctx context.Context, data model.ArrowCreate) (uuid.UUID, error)` (inserts `status`, default `'in_use'`)
    - [x] `CreateBatch(ctx context.Context, data model.ArrowBatchCreate) ([]model.ArrowRead, error)` (inserts `status`, default `'in_use'`)
    - [x] `FindByID(ctx context.Context, id uuid.UUID) (*model.ArrowRead, error)`
    - [x] `FindAllByArcherID(ctx context.Context, archerID uuid.UUID) ([]model.ArrowRead, error)`
    - [x] `CountByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)`
    - [x] `CountInUseByArcherID(ctx context.Context, archerID uuid.UUID) (int, error)` (for arrow ceiling)
    - [x] `FindByArcherAndSet(ctx context.Context, archerID uuid.UUID, set int16) ([]model.ArrowRead, error)`
- [x] New repository `backend/internal/repository/auth_identity.go` implements:
    - [x] `FindByGoogleSubject(ctx context.Context, sub string) (*model.AuthIdentityRead, error)`
    - [x] `FindByID(ctx context.Context, id uuid.UUID) (*model.AuthIdentityRead, error)`
    - [x] `Create(ctx context.Context, subject string, picture *string) (uuid.UUID, error)`
    - [x] `UpdateLastLogin(ctx context.Context, id uuid.UUID, login time.Time, pic *string) error`
- [x] Modified `backend/internal/repository/archer.go`:
    - [x] Column selection updated: `archer_id`, `email`, `first_name`, `last_name`, `date_of_birth`, `gender`, `is_deleted`.
    - [x] Scanning functions updated to omit legacy equipment/Google columns.
    - [x] `Create` accepts `model.ArcherCreate` with `ArcherID` provided from `auth`.
- [x] Repositories support transactions via `WithTx(tx pgx.Tx)`.
- [x] Unit tests for all repositories using mock `DBTX` verify SQL query and argument construction.
- [x] `cd backend && go test ./internal/repository/... -v` passes.
- [x] `cd backend && golangci-lint run ./internal/repository/...` reports no issues.

Mark each Step item in `docs/onboarding_refactor/tasks/003_repository_layer.md` as done (`[x]`):
- [x] Step 1: Write failing query building tests for Bow and Arrow repos
- [x] Step 2: Run tests to verify failure
- [x] Step 3: Implement `bow.go`
- [x] Step 4: Implement `arrow.go`
- [x] Step 5: Implement `auth_identity.go`
- [x] Step 6: Update `archer.go`
- [x] Step 7: Run repository unit tests
- [x] Step 8: Commit changes

- [x] **Step 2: Update `docs/plans/task.md`**

Mark all tasks as `Done` in `docs/plans/task.md`.

- [x] **Step 3: Commit documentation updates**

```bash
git add docs/onboarding_refactor/tasks/003_repository_layer.md docs/plans/task.md
git commit -m "docs: mark task 003 repository layer as completed"
```
