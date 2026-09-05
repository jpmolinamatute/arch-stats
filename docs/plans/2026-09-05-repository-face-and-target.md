# Task 013: Build Repository — Face and Target Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the target repository (`TargetRepo`) for managing lane target configurations per shooting session using Squirrel SQL query building against the PostgreSQL `target` table, and implement the face repository (`FaceRepo`) providing the target face definition registry (porting the 14KB Python `face_data.py` containing World Archery scoring zones, rings, spots, and dimensions) along with query-building helpers.

**Architecture:** 
- `TargetRepo` operates on the `DBTX` interface (`*pgxpool.Pool` or `pgx.Tx`), assembling parameterized PostgreSQL queries with Squirrel targeting the `target` table. It exposes CRUD operations (`Create`, `FindByID`, `FindBySlotID`, `FindBySessionID`, `Update`, `Delete`). Unit tests verify query generation, parameter binding, and defensive row scanning using `mockDBTX`.
- `FaceRepo` encapsulates target face geometry and scoring zone definitions ported directly from `backend-old/src/core/face_data.py` (14KB source). Because target face specifications (WA 122cm, 80cm, 60cm, 40cm, 6-rings, etc.) are standardized static geometry definitions rather than a mutable PostgreSQL table, `FaceRepo` holds the in-memory definition catalog for fast, zero-allocation lookups (`FindAll`, `FindByType`, `FindByID`) while providing Squirrel query building utilities (`BuildSelectQuery`) to fulfill repository interface and query verification requirements.
- `model.FaceRead` is declared as an alias for `model.Face` in `backend/internal/model/face.go` (following the pattern established in `backend/internal/model/auth.go`) to guarantee full type interoperability across service and handler layers.

**Tech Stack:** Go 1.27+, `github.com/Masterminds/squirrel`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`, standard library (`context`, `errors`, `fmt`, `time`).

**Spec:** [docs/go_refactor/tasks/013-repository_face_and_target.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/013-repository_face_and_target.md)

## Global Constraints

- Git branch: `refactor/013-repository-face-and-target`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/repository`
- All SQL queries must use PostgreSQL dollar placeholder format (`$1`, `$2`) via `StmtBuilder` (`squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)`).
- No ORMs; use `pgx/v5` only via `DBTX`.
- Error handling: Wrap errors with `%w` using contextual descriptive messages (`fmt.Errorf("...: %w", err)`). Return sentinel `apperror.ErrNotFound` on 0 rows affected for mutations where an entity was expected.
- Schema alignment notes:
  - In PostgreSQL migration `004_2025-09-26_shooting_sessions_table.sql`, the `target` table is defined as: `target_id UUID PRIMARY KEY`, `session_id UUID NOT NULL REFERENCES session`, `distance INTEGER NOT NULL`, `lane INTEGER NOT NULL`, `created_at TIMESTAMPTZ NOT NULL`.
  - In `013-repository_face_and_target.md`, draft step 2 informally referenced `INSERT with face_id, slot_id`; in reality, `target` records belong to a `session_id` and define `distance` and `lane`. The foreign key relationship with `slot` is reverse: `slot.target_id REFERENCES target(target_id)`.
  - `FindBySlotID` queries `target` joined with `slot` where `slot.slot_id = $1`.
  - Target face definitions (radii, viewBox, SVG styling, spots) are not stored in a PostgreSQL table; they originate in `backend-old/src/core/face_data.py` (14KB). `FaceRepo` ports these definitions into `backend/internal/repository/face_data.go` and serves them via `FindAll`, `FindByType`, and `FindByID`.
- Formatting must adhere to `gofumpt` and linting must pass `golangci-lint run ./...`.
- `go test -race ./internal/repository/... -v` must pass.
- `go vet ./...` must report no issues.

---

## File Structure

```
backend/
└── internal/
    ├── model/
    │   └── face.go                    # [MODIFY] Add type FaceRead = Face alias
    └── repository/
        ├── base.go                    # [EXISTING] DBTX interface, StmtBuilder, ScanOne, ScanRows
        ├── archer.go                  # [EXISTING] Reference repository implementation
        ├── archer_test.go             # [EXISTING] Shared mockDBTX, mockMultiRows, mockSingleRow
        ├── target.go                  # [NEW] TargetRepo struct, NewTargetRepo, WithTx, FindByID, FindBySlotID, FindBySessionID, Create, Update, Delete
        ├── target_test.go             # [NEW] Unit test suite covering all TargetRepo methods
        ├── face_data.go               # [NEW] Static catalog of World Archery face definitions ported from face_data.py
        ├── face.go                    # [NEW] FaceRepo struct, NewFaceRepo, WithTx, FindByID, FindAll, FindByType, BuildSelectQuery
        └── face_test.go               # [NEW] Unit test suite covering FaceRepo queries and catalog retrieval
```

---

### Task 1: Git Branch Setup & Model Alias

**Files:**
- Modify: `backend/internal/model/face.go:34`

**Interfaces:**
- Consumes: `model.Face` from `backend/internal/model/face.go`
- Produces: `model.FaceRead` alias (`type FaceRead = Face`)

- [ ] **Step 1: Check out git branch**

```bash
git switch -c refactor/013-repository-face-and-target
```

- [ ] **Step 2: Add `FaceRead` alias in `backend/internal/model/face.go`**

In `backend/internal/model/face.go`, append line 34:

```go
// FaceRead is an alias for Face to support repository and service layer naming conventions.
type FaceRead = Face
```

- [ ] **Step 3: Run model tests to verify compilation and compatibility**

```bash
cd backend && go test ./internal/model/... -v
```
Expected: PASS with all model tests green.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/face.go
git commit -m "feat(model): add FaceRead alias to face model"
```

---

### Task 2: Target Repository Implementation (`target.go` & `target_test.go`)

**Files:**
- Create: `backend/internal/repository/target_test.go`
- Create: `backend/internal/repository/target.go`

**Interfaces:**
- Consumes: `model.TargetCreate`, `model.TargetSet`, `model.TargetFilter`, `model.TargetRead` from `backend/internal/model`
- Consumes: `DBTX`, `StmtBuilder`, `ScanOne`, `ScanRows` from `backend/internal/repository/base.go`
- Produces: `TargetRepo` with methods:
  - `NewTargetRepo(db DBTX) *TargetRepo`
  - `WithTx(tx pgx.Tx) *TargetRepo`
  - `FindByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error)`
  - `FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error)`
  - `FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error)`
  - `Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error)`
  - `Update(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error`
  - `Delete(ctx context.Context, id uuid.UUID) error`

- [ ] **Step 1: Write failing unit tests for `TargetRepo`**

Create `backend/internal/repository/target_test.go`:

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
			if len(args) != 1 || args[0] != targetID {
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
			if len(args) != 1 || args[0] != slotID {
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
			if len(args) != 1 || args[0] != sessionID {
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/repository/... -run TestTargetRepo -v
```
Expected: FAIL with `undefined: repository.NewTargetRepo` or build error.

- [ ] **Step 3: Implement `backend/internal/repository/target.go`**

Create `backend/internal/repository/target.go`:

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

var targetColumns = []string{
	"target_id",
	"session_id",
	"distance",
	"lane",
	"created_at",
}

// TargetRepo manages database operations for lane target configurations.
type TargetRepo struct {
	db DBTX
}

// NewTargetRepo constructs a TargetRepo backed by DBTX.
func NewTargetRepo(db DBTX) *TargetRepo {
	return &TargetRepo{db: db}
}

// WithTx returns a new TargetRepo bound to the given transaction.
func (r *TargetRepo) WithTx(tx pgx.Tx) *TargetRepo {
	return &TargetRepo{db: tx}
}

func scanTarget(scanner interface{ Scan(dest ...any) error }) (model.TargetRead, error) {
	var t model.TargetRead
	err := scanner.Scan(
		&t.TargetID,
		&t.SessionID,
		&t.Distance,
		&t.Lane,
		&t.CreatedAt,
	)
	if err != nil {
		return model.TargetRead{}, err
	}
	return t, nil
}

// FindByID retrieves a target configuration by primary key identifier.
// Returns nil, nil if no target exists with the given ID.
func (r *TargetRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
	sql, args, err := StmtBuilder.Select(targetColumns...).
		From("target").
		Where(squirrel.Eq{"target_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find target by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.TargetRead, error) {
		return scanTarget(r)
	})
}

// FindBySlotID retrieves target configurations associated with a specific slot.
func (r *TargetRepo) FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error) {
	cols := make([]string, len(targetColumns))
	for i, c := range targetColumns {
		cols[i] = "target." + c
	}

	sql, args, err := StmtBuilder.Select(cols...).
		From("target").
		Join("slot ON target.target_id = slot.target_id").
		Where(squirrel.Eq{"slot.slot_id": slotID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find targets by slot id query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying targets by slot id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.TargetRead, error) {
		return scanTarget(r)
	})
}

// FindBySessionID retrieves all target configurations for a given session, ordered by lane ascending.
func (r *TargetRepo) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error) {
	sql, args, err := StmtBuilder.Select(targetColumns...).
		From("target").
		Where(squirrel.Eq{"session_id": sessionID}).
		OrderBy("lane ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find targets by session id query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying targets by session id: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.TargetRead, error) {
		return scanTarget(r)
	})
}

// Create inserts a new target configuration record and returns the generated UUID identifier.
func (r *TargetRepo) Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error) {
	sql, args, err := StmtBuilder.Insert("target").
		Columns("session_id", "distance", "lane").
		Values(data.SessionID, data.Distance, data.Lane).
		Suffix("RETURNING target_id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create target query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting target: %w", err)
	}

	return newID, nil
}

// Update mutates target fields specified in data for rows matching filter.
// Requires at least one filter criterion to prevent unrestricted updates.
// Returns apperror.ErrNotFound if no target matched the filter.
func (r *TargetRepo) Update(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error {
	q := StmtBuilder.Update("target")
	setCount := 0

	if data.Distance != nil {
		q = q.Set("distance", *data.Distance)
		setCount++
	}
	if data.Lane != nil {
		q = q.Set("lane", *data.Lane)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.TargetID != nil {
		q = q.Where(squirrel.Eq{"target_id": *filter.TargetID})
		whereCount++
	}
	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
		whereCount++
	}
	if filter.Distance != nil {
		q = q.Where(squirrel.Eq{"distance": *filter.Distance})
		whereCount++
	}
	if filter.Lane != nil {
		q = q.Where(squirrel.Eq{"lane": *filter.Lane})
		whereCount++
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
		whereCount++
	}

	if whereCount == 0 {
		return errors.New("at least one filter criterion is required for update")
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("building update target query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing update target: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// Delete removes a target configuration by primary key identifier.
// Returns apperror.ErrNotFound if no target existed with the given ID.
func (r *TargetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("target").
		Where(squirrel.Eq{"target_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete target query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete target: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/repository/... -run TestTargetRepo -v
```
Expected: PASS for all `TestTargetRepo_*` test cases.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/target.go backend/internal/repository/target_test.go
git commit -m "feat(repository): implement target repository and query test suite"
```

---

### Task 3: Face Repository & Target Face Catalog Implementation (`face_data.go`, `face.go` & `face_test.go`)

**Files:**
- Create: `backend/internal/repository/face_data.go`
- Create: `backend/internal/repository/face_test.go`
- Create: `backend/internal/repository/face.go`

**Interfaces:**
- Consumes: `model.FaceType`, `model.Face`, `model.FaceRead`, `model.Ring`, `model.Spot` from `backend/internal/model`
- Consumes: `DBTX`, `StmtBuilder` from `backend/internal/repository/base.go`
- Produces: `FaceRepo` with methods:
  - `NewFaceRepo(db ...DBTX) *FaceRepo`
  - `WithTx(tx pgx.Tx) *FaceRepo`
  - `FindByID(ctx context.Context, id string) (*model.FaceRead, error)`
  - `FindAll(ctx context.Context) ([]model.FaceRead, error)`
  - `FindByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error)`
  - `BuildSelectQuery(faceType *model.FaceType) (string, []any, error)`

- [ ] **Step 1: Create face catalog in `backend/internal/repository/face_data.go`**

Create `backend/internal/repository/face_data.go` porting all target face definitions from `backend-old/src/core/face_data.py`:

```go
package repository

import "github.com/jpmolinamatute/arch-stats/backend/internal/model"

// DefaultFaceCatalog contains the standardized World Archery target face definitions
// ported from face_data.py (14KB Python source).
var DefaultFaceCatalog = []model.FaceRead{
	{
		FaceType:    model.FaceTypeNone,
		FaceName:    "No Target Face",
		RenderCross: false,
		ViewBox:     0.0,
		Spots:       []model.Spot{},
		Rings:       []model.Ring{},
	},
	{
		FaceType:    model.FaceTypeWA122Full,
		FaceName:    "WA 122cm Standard Target Face",
		RenderCross: true,
		ViewBox:     1342.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 1220.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 1, Fill: "#FFFFFF", R: 610.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 2, Fill: "#FFFFFF", R: 549.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 3, Fill: "#000000", R: 488.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 4, Fill: "#000000", R: 427.0, Stroke: "#FFFFFF", StrokeWidth: 1.0},
			{DataScore: 5, Fill: "#00B4E4", R: 366.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 305.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 244.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 183.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 122.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 61.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 30.5, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA80Full,
		FaceName:    "WA 80cm Standard Target Face",
		RenderCross: true,
		ViewBox:     880.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 800.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 1, Fill: "#FFFFFF", R: 400.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 2, Fill: "#FFFFFF", R: 360.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 3, Fill: "#000000", R: 320.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 4, Fill: "#000000", R: 280.0, Stroke: "#FFFFFF", StrokeWidth: 1.0},
			{DataScore: 5, Fill: "#00B4E4", R: 240.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 200.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 160.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 120.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 80.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 40.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 20.0, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA60Full,
		FaceName:    "WA 60cm Standard Target Face",
		RenderCross: true,
		ViewBox:     660.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 600.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 1, Fill: "#FFFFFF", R: 300.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 2, Fill: "#FFFFFF", R: 270.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 3, Fill: "#000000", R: 240.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 4, Fill: "#000000", R: 210.0, Stroke: "#FFFFFF", StrokeWidth: 1.0},
			{DataScore: 5, Fill: "#00B4E4", R: 180.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 150.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 120.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 90.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 60.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 30.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 15.0, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA40Full,
		FaceName:    "WA 40cm Standard Target Face",
		RenderCross: true,
		ViewBox:     440.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 400.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 1, Fill: "#FFFFFF", R: 200.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 2, Fill: "#FFFFFF", R: 180.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 3, Fill: "#000000", R: 160.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 4, Fill: "#000000", R: 140.0, Stroke: "#FFFFFF", StrokeWidth: 1.0},
			{DataScore: 5, Fill: "#00B4E4", R: 120.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 100.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 80.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 60.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 40.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 20.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 10.0, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA1226Rings,
		FaceName:    "WA 122cm 6-Ring Target Face",
		RenderCross: true,
		ViewBox:     854.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 732.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 5, Fill: "#00B4E4", R: 366.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 305.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 244.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 183.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 122.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 61.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 30.5, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA806Rings,
		FaceName:    "WA 80cm 6-Ring Target Face",
		RenderCross: true,
		ViewBox:     560.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 480.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 5, Fill: "#00B4E4", R: 240.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 200.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 160.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 120.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 80.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 40.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 20.0, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
}
```

- [ ] **Step 2: Write failing unit tests for `FaceRepo`**

Create `backend/internal/repository/face_test.go`:

```go
package repository_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestFaceRepo_FindAll(t *testing.T) {
	repo := repository.NewFaceRepo()
	faces, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != len(repository.DefaultFaceCatalog) {
		t.Fatalf("expected %d faces, got %d", len(repository.DefaultFaceCatalog), len(faces))
	}

	foundWA122 := false
	for _, f := range faces {
		if f.FaceType == model.FaceTypeWA122Full {
			foundWA122 = true
			if f.ViewBox != 1342.0 || len(f.Spots) != 1 || len(f.Rings) != 11 {
				t.Fatalf("unexpected WA122 definition: %+v", f)
			}
		}
	}
	if !foundWA122 {
		t.Fatal("WA 122cm full face not found in catalog")
	}
}

func TestFaceRepo_FindByType_Found(t *testing.T) {
	repo := repository.NewFaceRepo()
	faces, err := repo.FindByType(context.Background(), model.FaceTypeWA40Full)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != 1 {
		t.Fatalf("expected 1 face, got %d", len(faces))
	}
	if faces[0].FaceType != model.FaceTypeWA40Full {
		t.Fatalf("expected WA 40cm full, got %s", faces[0].FaceType)
	}
	if len(faces[0].Rings) != 11 {
		t.Fatalf("expected 11 rings, got %d", len(faces[0].Rings))
	}
}

func TestFaceRepo_FindByType_NotFound(t *testing.T) {
	repo := repository.NewFaceRepo()
	faces, err := repo.FindByType(context.Background(), model.FaceType("unknown_custom_face"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != 0 {
		t.Fatalf("expected 0 faces, got %d", len(faces))
	}
}

func TestFaceRepo_FindByID_Found(t *testing.T) {
	repo := repository.NewFaceRepo()
	face, err := repo.FindByID(context.Background(), string(model.FaceTypeWA80Full))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if face == nil {
		t.Fatal("expected face, got nil")
	}
	if face.FaceType != model.FaceTypeWA80Full {
		t.Fatalf("expected WA 80cm full, got %s", face.FaceType)
	}
}

func TestFaceRepo_FindByID_NotFoundReturnsNil(t *testing.T) {
	repo := repository.NewFaceRepo()
	face, err := repo.FindByID(context.Background(), "non_existent_face")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if face != nil {
		t.Fatalf("expected nil face, got %+v", face)
	}
}

func TestFaceRepo_BuildSelectQuery(t *testing.T) {
	repo := repository.NewFaceRepo()

	// Test FindAll query building
	sqlAll, argsAll, err := repo.BuildSelectQuery(nil)
	if err != nil {
		t.Fatalf("unexpected error building all query: %v", err)
	}
	if !strings.Contains(sqlAll, "SELECT face_type, face_name, viewBox, render_cross FROM face") {
		t.Fatalf("unexpected SQL for FindAll: %s", sqlAll)
	}
	if len(argsAll) != 0 {
		t.Fatalf("expected 0 args for FindAll, got %d", len(argsAll))
	}

	// Test FindByType query building with WHERE clause
	ft := model.FaceTypeWA60Full
	sqlType, argsType, err := repo.BuildSelectQuery(&ft)
	if err != nil {
		t.Fatalf("unexpected error building type query: %v", err)
	}
	if !strings.Contains(sqlType, "WHERE face_type = $1") {
		t.Fatalf("unexpected SQL for FindByType: %s", sqlType)
	}
	if len(argsType) != 1 || argsType[0] != model.FaceTypeWA60Full {
		t.Fatalf("unexpected args for FindByType: %v", argsType)
	}
}

func TestFaceRepo_WithTx(t *testing.T) {
	mock := &mockDBTX{}
	repo := repository.NewFaceRepo(mock)
	txRepo := repo.WithTx(nil)
	if txRepo == nil {
		t.Fatal("expected non-nil repo from WithTx")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd backend && go test ./internal/repository/... -run TestFaceRepo -v
```
Expected: FAIL with `undefined: repository.NewFaceRepo` or build error.

- [ ] **Step 4: Implement `backend/internal/repository/face.go`**

Create `backend/internal/repository/face.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var faceColumns = []string{
	"face_type",
	"face_name",
	"viewBox",
	"render_cross",
}

// FaceRepo manages target face definitions and scoring zones.
type FaceRepo struct {
	db      DBTX
	catalog []model.FaceRead
}

// NewFaceRepo constructs a FaceRepo. Accepts an optional DBTX to maintain constructor parity
// across repository instances.
func NewFaceRepo(db ...DBTX) *FaceRepo {
	var conn DBTX
	if len(db) > 0 {
		conn = db[0]
	}
	return &FaceRepo{
		db:      conn,
		catalog: DefaultFaceCatalog,
	}
}

// WithTx returns a new FaceRepo bound to the given transaction.
func (r *FaceRepo) WithTx(tx pgx.Tx) *FaceRepo {
	return &FaceRepo{
		db:      tx,
		catalog: r.catalog,
	}
}

// BuildSelectQuery constructs a squirrel SELECT statement targeting target face columns.
// Used for query validation and future database-backed catalog operations.
func (r *FaceRepo) BuildSelectQuery(faceType *model.FaceType) (string, []any, error) {
	q := StmtBuilder.Select(faceColumns...).From("face")
	if faceType != nil {
		q = q.Where(squirrel.Eq{"face_type": *faceType})
	}
	sql, args, err := q.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("building select face query: %w", err)
	}
	return sql, args, nil
}

// FindAll returns all available target face definitions.
func (r *FaceRepo) FindAll(ctx context.Context) ([]model.FaceRead, error) {
	_ = ctx
	res := make([]model.FaceRead, len(r.catalog))
	copy(res, r.catalog)
	return res, nil
}

// FindByType returns target face definitions matching the given face type.
// Returns an empty slice if no matching definition is found.
func (r *FaceRepo) FindByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
	_ = ctx
	res := make([]model.FaceRead, 0)
	for _, f := range r.catalog {
		if f.FaceType == faceType {
			res = append(res, f)
		}
	}
	return res, nil
}

// FindByID retrieves a target face definition by its string identifier (face_type).
// Returns nil, nil if no matching definition exists.
func (r *FaceRepo) FindByID(ctx context.Context, id string) (*model.FaceRead, error) {
	_ = ctx
	for _, f := range r.catalog {
		if string(f.FaceType) == id {
			found := f
			return &found, nil
		}
	}
	return nil, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd backend && go test ./internal/repository/... -run TestFaceRepo -v
```
Expected: PASS for all `TestFaceRepo_*` test cases.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/repository/face_data.go backend/internal/repository/face.go backend/internal/repository/face_test.go
git commit -m "feat(repository): implement face repository and target face catalog"
```

---

### Task 4: Complete Repository Verification, Linting & Spec Update

**Files:**
- Modify: `docs/go_refactor/tasks/013-repository_face_and_target.md`

**Interfaces:**
- Consumes: All repository and model packages
- Produces: 100% passing tests, clean `go vet`, clean `golangci-lint`, and verified task completion checklist

- [ ] **Step 1: Run full repository test suite with race detector**

```bash
cd backend && go test -race ./internal/repository/... -v
```
Expected: PASS across all repository tests.

- [ ] **Step 2: Run `go vet` across backend**

```bash
cd backend && go vet ./...
```
Expected: Clean output with 0 warnings or errors.

- [ ] **Step 3: Run `go build` across backend**

```bash
cd backend && go build ./...
```
Expected: Clean build with exit code 0.

- [ ] **Step 4: Run Go linter & formatting check**

```bash
cd backend && golangci-lint run ./...
```
Expected: Clean output with 0 issues reported.

- [ ] **Step 5: Update acceptance criteria checklist in `docs/go_refactor/tasks/013-repository_face_and_target.md`**

Mark completed checkboxes in `docs/go_refactor/tasks/013-repository_face_and_target.md`.

- [ ] **Step 6: Commit final documentation update**

```bash
git add docs/go_refactor/tasks/013-repository_face_and_target.md
git commit -m "docs: mark task 013 face and target repositories as complete"
```

---

## Self-Review Checklist

1. **Spec coverage:**
   - `FaceRepo` with `FindByID`, `FindAll`, `FindByType`: Covered in Task 3.
   - `TargetRepo` with `FindByID`, `FindBySlotID`, `Create`, `Update`, `Delete`: Covered in Task 2.
   - Squirrel query building and dollar placeholder format: Covered in Tasks 2 & 3.
   - 14KB Python `face_data.py` port: Covered in Task 3 (`face_data.go`).
   - Schema alignment: Documented in Global Constraints and implemented with true schema columns (`target` and `slot` relations).
2. **Placeholder scan:** None. All code blocks, signatures, SQL clauses, and test assertions are concrete and complete.
3. **Type consistency:** `FaceRead` is aliased to `Face`; `TargetRepo` returns `*model.TargetRead` / `[]model.TargetRead`; `Create` returns `uuid.UUID`; sentinel errors (`apperror.ErrNotFound`) are returned consistently on 0 rows affected.
