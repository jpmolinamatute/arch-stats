# Task 008: Build Repository Layer — Base Patterns + Archer Repository Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the foundation of the repository layer by defining `DBTX`, transaction management (`WithTx`), row scanning utilities, and implementing the `ArcherRepo`, `MaintenanceRepo`, and `ReportingRepo` with Squirrel query building and pgx/v5.

**Architecture:** The repository layer abstracts PostgreSQL persistence behind the `DBTX` interface (compatible with `*pgxpool.Pool` and `pgx.Tx`), decoupling query execution from concrete pool instances. All queries use parameterized Squirrel SQL builders with PostgreSQL dollar formatting (`$1`, `$2`), returning domain models from `internal/model/` and sentinel errors from `internal/apperror/`. Unit tests use a mock `DBTX` to verify query construction, bound arguments, and row mapping without requiring an external database.

**Tech Stack:** Go 1.27+, `github.com/Masterminds/squirrel`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`, standard library (`context`, `errors`, `fmt`, `time`).

**Spec:** [docs/go_refactor/tasks/008-repository_base_and_archer.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/008-repository_base_and_archer.md)

## Global Constraints

- Git branch: `refactor/008-repository-base-and-archer`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/repository`
- All SQL queries must use PostgreSQL dollar placeholder format (`$1`, `$2`) via `squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)`, except where raw SQL is explicitly allowed (`MaintenanceRepo`, `ReportingRepo`).
- No ORMs; use `pgx/v5` only.
- Strict error handling: wrap errors with `%w` and return sentinel errors (`apperror.ErrNotFound`) when operations find or affect 0 rows.
- Formatting must adhere to `gofumpt` and linting must pass `golangci-lint run ./...`.
- `go test -race ./internal/repository/... -v` must pass.
- `go vet ./...` must report no issues.

---

## File Structure

```
backend/
├── go.mod                                          # [MODIFY] Add Masterminds/squirrel, promote google/uuid to direct
├── go.sum                                          # [MODIFY] Dependency checksums
└── internal/
    └── repository/
        ├── pool.go                                 # [EXISTING] Connection pool setup
        ├── pool_test.go                            # [EXISTING] Pool unit tests
        ├── migrate.go                              # [EXISTING] Goose migration runner
        ├── migrate_test.go                         # [EXISTING] Migration tests
        ├── base.go                                 # [NEW] DBTX interface, Transactor, WithTx, ScanRows, ScanOne, statement builder
        ├── base_test.go                            # [NEW] Unit tests for base utilities and WithTx
        ├── archer.go                               # [NEW] ArcherRepo CRUD with Squirrel
        ├── archer_test.go                          # [NEW] Unit tests for ArcherRepo query building, scanning, and execution
        ├── maintenance.go                          # [NEW] MaintenanceRepo (materialized view refresh, schema version)
        ├── maintenance_test.go                     # [NEW] Unit tests for MaintenanceRepo
        ├── reporting.go                            # [NEW] ReportingRepo (session summary, archer performance stubs)
        └── reporting_test.go                       # [NEW] Unit tests for ReportingRepo
```

---

### Task 1: Module Dependencies & Git Branch Setup

**Files:**
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`

**Interfaces:**
- Consumes: Go toolchain (`go get`)
- Produces:
  - `github.com/Masterminds/squirrel v1.5.4` (direct)
  - `github.com/google/uuid v1.6.0` (promoted to direct)

- [ ] **Step 1: Check out git branch**

```bash
git checkout -b refactor/008-repository-base-and-archer
```

- [ ] **Step 2: Add squirrel and uuid dependencies**

```bash
cd backend
go get github.com/Masterminds/squirrel@v1.5.4
go get github.com/google/uuid@v1.6.0
go mod tidy
```

- [ ] **Step 3: Verify go.mod and go.sum are valid**

Run: `cd backend && go vet ./...`
Expected: PASS with no errors

- [ ] **Step 4: Commit dependency updates**

```bash
git add backend/go.mod backend/go.sum
git commit -m "deps: add squirrel query builder and promote uuid to direct dependency"
```

---

### Task 2: Repository Base Patterns (`base.go`) & Unit Tests (`base_test.go`)

**Files:**
- Create: `backend/internal/repository/base.go`
- Create: `backend/internal/repository/base_test.go`

**Interfaces:**
- Consumes:
  - `github.com/jackc/pgx/v5`
  - `github.com/jackc/pgx/v5/pgconn`
  - `github.com/jackc/pgx/v5/pgxpool`
  - `github.com/Masterminds/squirrel`
- Produces:
  - `type DBTX interface` with `Query`, `QueryRow`, `Exec`
  - `type Transactor interface` with `Begin(ctx context.Context) (pgx.Tx, error)`
  - `var stmtBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)`
  - `func WithTx(ctx context.Context, db Transactor, fn func(tx pgx.Tx) error) error`
  - `func ScanRows[T any](rows pgx.Rows, scanFn func(pgx.Rows) (T, error)) ([]T, error)`
  - `func ScanOne[T any](row pgx.Row, scanFn func(pgx.Row) (T, error)) (*T, error)`

- [ ] **Step 1: Write failing unit tests for base repository utilities**

Create `backend/internal/repository/base_test.go`:

```go
package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

// mockTx implements pgx.Tx for testing WithTx.
type mockTx struct {
	committed  bool
	rolledBack bool
	commitErr  error
	rbErr      error
}

func (m *mockTx) Begin(ctx context.Context) (pgx.Tx, error) { return m, nil }
func (m *mockTx) Commit(ctx context.Context) error {
	m.committed = true
	return m.commitErr
}
func (m *mockTx) Rollback(ctx context.Context) error {
	m.rolledBack = true
	return m.rbErr
}
func (m *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (m *mockTx) LargeObjects() pgx.LargeObjects                             { return pgx.LargeObjects{} }
func (m *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (m *mockTx) Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error) {
	return pgconn.CommandTag{}, nil
}
func (m *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return nil
}
func (m *mockTx) Conn() *pgx.Conn { return nil }

type mockTransactor struct {
	tx       *mockTx
	beginErr error
}

func (m *mockTransactor) Begin(ctx context.Context) (pgx.Tx, error) {
	if m.beginErr != nil {
		return nil, m.beginErr
	}
	return m.tx, nil
}

// mockSingleRow implements pgx.Row.
type mockSingleRow struct {
	scanFn func(dest ...any) error
}

func (r *mockSingleRow) Scan(dest ...any) error {
	return r.scanFn(dest...)
}

func TestWithTx_SuccessCommits(t *testing.T) {
	tx := &mockTx{}
	transactor := &mockTransactor{tx: tx}

	err := repository.WithTx(context.Background(), transactor, func(tx pgx.Tx) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !tx.committed {
		t.Error("expected transaction to be committed")
	}
}

func TestWithTx_CallbackErrorRollsBack(t *testing.T) {
	tx := &mockTx{}
	transactor := &mockTransactor{tx: tx}
	callbackErr := errors.New("something went wrong")

	err := repository.WithTx(context.Background(), transactor, func(tx pgx.Tx) error {
		return callbackErr
	})
	if !errors.Is(err, callbackErr) {
		t.Fatalf("expected callbackErr, got: %v", err)
	}
	if tx.committed {
		t.Error("expected transaction not to be committed")
	}
	if !tx.rolledBack {
		t.Error("expected transaction to be rolled back")
	}
}

func TestWithTx_BeginError(t *testing.T) {
	beginErr := errors.New("begin failed")
	transactor := &mockTransactor{beginErr: beginErr}

	err := repository.WithTx(context.Background(), transactor, func(tx pgx.Tx) error {
		return nil
	})
	if !errors.Is(err, beginErr) {
		t.Fatalf("expected beginErr, got: %v", err)
	}
}

func TestScanOne_Success(t *testing.T) {
	row := &mockSingleRow{
		scanFn: func(dest ...any) error {
			val := dest[0].(*string)
			*val = "archer-1"
			return nil
		},
	}

	result, err := repository.ScanOne(row, func(r pgx.Row) (string, error) {
		var s string
		err := r.Scan(&s)
		return s, err
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if result == nil || *result != "archer-1" {
		t.Fatalf("expected 'archer-1', got: %v", result)
	}
}

func TestScanOne_NoRowsReturnsNil(t *testing.T) {
	row := &mockSingleRow{
		scanFn: func(dest ...any) error {
			return pgx.ErrNoRows
		},
	}

	result, err := repository.ScanOne(row, func(r pgx.Row) (string, error) {
		var s string
		err := r.Scan(&s)
		return s, err
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got: %v", result)
	}
}

func TestScanOne_ScanError(t *testing.T) {
	scanErr := errors.New("scan failed")
	row := &mockSingleRow{
		scanFn: func(dest ...any) error {
			return scanErr
		},
	}

	_, err := repository.ScanOne(row, func(r pgx.Row) (string, error) {
		var s string
		err := r.Scan(&s)
		return s, err
	})
	if !errors.Is(err, scanErr) {
		t.Fatalf("expected scanErr, got: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/repository -run "TestWithTx|TestScanOne" -v`
Expected: FAIL with compilation error (undefined `repository.WithTx`, `repository.ScanOne`)

- [ ] **Step 3: Implement `backend/internal/repository/base.go`**

Create `backend/internal/repository/base.go`:

```go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX abstracts database operations across connection pools and transactions.
// Both *pgxpool.Pool and pgx.Tx satisfy this interface.
type DBTX interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Transactor abstracts beginning a database transaction.
// *pgxpool.Pool satisfies this interface.
type Transactor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// stmtBuilder is the configured Squirrel statement builder using PostgreSQL dollar placeholders ($1, $2).
var stmtBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// WithTx wraps a function in a database transaction. The callback receives a pgx.Tx
// which satisfies the DBTX interface, so any repository method can participate in the transaction.
func WithTx(ctx context.Context, db Transactor, fn func(tx pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) // no-op if already committed

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ScanRows iterates over pgx.Rows, applying scanFn for each row, and returns a slice of results.
// It ensures rows are closed and checks rows.Err().
func ScanRows[T any](rows pgx.Rows, scanFn func(pgx.Rows) (T, error)) ([]T, error) {
	defer rows.Close()

	items := make([]T, 0)
	for rows.Next() {
		item, err := scanFn(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return items, nil
}

// ScanOne scans a single row using scanFn. If no row is returned (pgx.ErrNoRows),
// it returns nil, nil to signify an absent entity.
func ScanOne[T any](row pgx.Row, scanFn func(pgx.Row) (T, error)) (*T, error) {
	item, err := scanFn(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning row: %w", err)
	}
	return &item, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/repository -run "TestWithTx|TestScanOne" -v`
Expected: PASS

- [ ] **Step 5: Commit base repository patterns**

```bash
git add backend/internal/repository/base.go backend/internal/repository/base_test.go
git commit -m "feat: add repository DBTX interface, WithTx, statement builder, and scanning helpers"
```

---

### Task 3: Archer Repository Query Building & CRUD Operations (`archer.go` & `archer_test.go`)

**Files:**
- Create: `backend/internal/repository/archer.go`
- Create: `backend/internal/repository/archer_test.go`

**Interfaces:**
- Consumes:
  - `DBTX`, `stmtBuilder`, `ScanRows`, `ScanOne`
  - `model.ArcherRead`, `model.ArcherCreate`, `model.ArcherSet`, `model.ArcherFilter`
  - `apperror.ErrNotFound`
- Produces:
  - `type ArcherRepo struct { db DBTX }`
  - `func NewArcherRepo(db DBTX) *ArcherRepo`
  - `func (r *ArcherRepo) WithTx(tx pgx.Tx) *ArcherRepo`
  - `func (r *ArcherRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)`
  - `func (r *ArcherRepo) FindByEmail(ctx context.Context, email string) (*model.ArcherRead, error)`
  - `func (r *ArcherRepo) FindByGoogleSubject(ctx context.Context, sub string) (*model.ArcherRead, error)`
  - `func (r *ArcherRepo) FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)`
  - `func (r *ArcherRepo) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)`
  - `func (r *ArcherRepo) Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error`
  - `func (r *ArcherRepo) Delete(ctx context.Context, id uuid.UUID) error`

- [ ] **Step 1: Write failing tests for ArcherRepo**

Create `backend/internal/repository/archer_test.go`:

```go
package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

// mockDBTX records queries and executes configured function mocks.
type mockDBTX struct {
	lastSQL    string
	lastArgs   []any
	queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
	queryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	execFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *mockDBTX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	m.lastSQL = sql
	m.lastArgs = args
	if m.queryFn != nil {
		return m.queryFn(ctx, sql, args...)
	}
	return nil, nil
}

func (m *mockDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	m.lastSQL = sql
	m.lastArgs = args
	if m.queryRowFn != nil {
		return m.queryRowFn(ctx, sql, args...)
	}
	return &mockSingleRow{scanFn: func(dest ...any) error { return pgx.ErrNoRows }}
}

func (m *mockDBTX) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	m.lastSQL = sql
	m.lastArgs = args
	if m.execFn != nil {
		return m.execFn(ctx, sql, args...)
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

// mockMultiRows implements pgx.Rows for testing FindAll.
type mockMultiRows struct {
	records [][]any
	idx     int
	err     error
	closed  bool
}

func (m *mockMultiRows) Close()                                           { m.closed = true }
func (m *mockMultiRows) Err() error                                       { return m.err }
func (m *mockMultiRows) CommandTag() pgconn.CommandTag                    { return pgconn.CommandTag{} }
func (m *mockMultiRows) FieldDescriptions() []pgconn.FieldDescription     { return nil }
func (m *mockMultiRows) Next() bool {
	m.idx++
	return m.idx <= len(m.records)
}
func (m *mockMultiRows) Scan(dest ...any) error {
	row := m.records[m.idx-1]
	for i, v := range row {
		switch d := dest[i].(type) {
		case *uuid.UUID:
			*d = v.(uuid.UUID)
		case **uuid.UUID:
			if v == nil {
				*d = nil
			} else {
				u := v.(uuid.UUID)
				*d = &u
			}
		case *string:
			*d = v.(string)
		case **string:
			if v == nil {
				*d = nil
			} else {
				s := v.(string)
				*d = &s
			}
		case *model.Gender:
			*d = v.(model.Gender)
		case *model.Bowstyle:
			*d = v.(model.Bowstyle)
		case *float64:
			*d = v.(float64)
		case *time.Time:
			*d = v.(time.Time)
		case *any:
			*d = v
		}
	}
	return nil
}
func (m *mockMultiRows) Values() ([]any, error) { return m.records[m.idx-1], nil }
func (m *mockMultiRows) RawValues() [][]byte    { return nil }
func (m *mockMultiRows) Conn() *pgx.Conn        { return nil }

func sampleArcherRow(id uuid.UUID, email, googleSub string) []any {
	now := time.Now().Truncate(time.Second)
	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	pic := "https://example.com/photo.jpg"
	return []any{
		id,
		"Robin",
		"Hood",
		email,
		dob,
		model.GenderMale,
		model.BowstyleBarebow,
		42.5,
		nil, // club_id
		&pic,
		googleSub,
		now,
		now,
	}
}

func TestArcherRepo_FindByID_Success(t *testing.T) {
	archerID := uuid.New()
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleArcherRow(archerID, "robin@sherwood.org", "sub-123")
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewArcherRepo(mock)
	archer, err := repo.FindByID(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if archer == nil {
		t.Fatal("expected archer, got nil")
	}
	if archer.ArcherID != archerID {
		t.Errorf("expected id %v, got %v", archerID, archer.ArcherID)
	}
	if archer.Email != "robin@sherwood.org" {
		t.Errorf("expected email robin@sherwood.org, got %s", archer.Email)
	}
	if archer.DateOfBirth != "1990-01-15" {
		t.Errorf("expected dob 1990-01-15, got %s", archer.DateOfBirth)
	}
	if mock.lastArgs[0] != archerID {
		t.Errorf("expected query arg %v, got %v", archerID, mock.lastArgs[0])
	}
}

func TestArcherRepo_FindByID_NotFound(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewArcherRepo(mock)
	archer, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if archer != nil {
		t.Errorf("expected nil archer on ErrNoRows, got %v", archer)
	}
}

func TestArcherRepo_FindByEmail_Success(t *testing.T) {
	archerID := uuid.New()
	email := "target@example.com"
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleArcherRow(archerID, email, "sub-456")
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewArcherRepo(mock)
	archer, err := repo.FindByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if archer == nil || archer.Email != email {
		t.Fatalf("expected archer with email %s, got %v", email, archer)
	}
	if mock.lastArgs[0] != email {
		t.Errorf("expected query arg %s, got %v", email, mock.lastArgs[0])
	}
}

func TestArcherRepo_FindByGoogleSubject_Success(t *testing.T) {
	archerID := uuid.New()
	sub := "google-sub-789"
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleArcherRow(archerID, "user@gmail.com", sub)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewArcherRepo(mock)
	archer, err := repo.FindByGoogleSubject(context.Background(), sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if archer == nil || archer.GoogleSubject != sub {
		t.Fatalf("expected archer with google subject %s, got %v", sub, archer)
	}
	if mock.lastArgs[0] != sub {
		t.Errorf("expected query arg %s, got %v", sub, mock.lastArgs[0])
	}
}

func TestArcherRepo_FindAll_WithFilters(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return &mockMultiRows{
				records: [][]any{
					sampleArcherRow(id1, "a1@test.com", "sub1"),
					sampleArcherRow(id2, "a2@test.com", "sub2"),
				},
			}, nil
		},
	}

	repo := repository.NewArcherRepo(mock)
	gender := model.GenderMale
	bowstyle := model.BowstyleBarebow
	filter := model.ArcherFilter{
		Gender:   &gender,
		Bowstyle: &bowstyle,
	}

	archers, err := repo.FindAll(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(archers) != 2 {
		t.Fatalf("expected 2 archers, got %d", len(archers))
	}
	if len(mock.lastArgs) != 2 {
		t.Errorf("expected 2 filter args, got %d", len(mock.lastArgs))
	}
}

func TestArcherRepo_Create_Success(t *testing.T) {
	generatedID := uuid.New()
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
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
		FirstName:     "Robin",
		LastName:      "Hood",
		Email:         "robin@sherwood.org",
		DateOfBirth:   "1990-01-15",
		Gender:        model.GenderMale,
		Bowstyle:      model.BowstyleBarebow,
		DrawWeight:    42.5,
		GoogleSubject: "google-sub-robin",
	}

	id, err := repo.Create(context.Background(), createPayload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != generatedID {
		t.Errorf("expected generated id %v, got %v", generatedID, id)
	}
}

func TestArcherRepo_Create_InvalidDateFormat(t *testing.T) {
	repo := repository.NewArcherRepo(&mockDBTX{})
	createPayload := model.ArcherCreate{
		FirstName:   "Robin",
		LastName:    "Hood",
		Email:       "robin@sherwood.org",
		DateOfBirth: "invalid-date",
	}

	_, err := repo.Create(context.Background(), createPayload)
	if err == nil {
		t.Fatal("expected error on invalid date_of_birth format, got nil")
	}
}

func TestArcherRepo_Update_Success(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := repository.NewArcherRepo(mock)
	newFirst := "Robbie"
	targetID := uuid.New()
	setPayload := model.ArcherSet{
		FirstName: &newFirst,
	}
	filter := model.ArcherFilter{
		ArcherID: &targetID,
	}

	err := repo.Update(context.Background(), setPayload, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestArcherRepo_Update_RowsAffectedZeroReturnsNotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}

	repo := repository.NewArcherRepo(mock)
	newFirst := "Robbie"
	targetID := uuid.New()
	setPayload := model.ArcherSet{FirstName: &newFirst}
	filter := model.ArcherFilter{ArcherID: &targetID}

	err := repo.Update(context.Background(), setPayload, filter)
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when 0 rows updated, got: %v", err)
	}
}

func TestArcherRepo_Update_EmptySetReturnsNil(t *testing.T) {
	repo := repository.NewArcherRepo(&mockDBTX{})
	targetID := uuid.New()
	err := repo.Update(context.Background(), model.ArcherSet{}, model.ArcherFilter{ArcherID: &targetID})
	if err != nil {
		t.Fatalf("expected nil on empty set, got: %v", err)
	}
}

func TestArcherRepo_Update_MissingFilterReturnsError(t *testing.T) {
	repo := repository.NewArcherRepo(&mockDBTX{})
	newFirst := "Robbie"
	err := repo.Update(context.Background(), model.ArcherSet{FirstName: &newFirst}, model.ArcherFilter{})
	if err == nil {
		t.Fatal("expected error on empty filter, got nil")
	}
}

func TestArcherRepo_Delete_Success(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("DELETE 1"), nil
		},
	}

	repo := repository.NewArcherRepo(mock)
	err := repo.Delete(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestArcherRepo_Delete_RowsAffectedZeroReturnsNotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("DELETE 0"), nil
		},
	}

	repo := repository.NewArcherRepo(mock)
	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when 0 rows deleted, got: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/repository -run "TestArcherRepo" -v`
Expected: FAIL with compilation error (undefined `repository.NewArcherRepo`)

- [ ] **Step 3: Implement `backend/internal/repository/archer.go`**

Create `backend/internal/repository/archer.go`:

```go
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var archerColumns = []string{
	"archer_id",
	"first_name",
	"last_name",
	"email",
	"date_of_birth",
	"gender",
	"bowstyle",
	"draw_weight",
	"club_id",
	"google_picture_url",
	"google_subject",
	"last_login_at",
	"created_at",
}

// ArcherRepo manages database operations for archer profiles.
type ArcherRepo struct {
	db DBTX
}

// NewArcherRepo constructs an ArcherRepo backed by DBTX.
func NewArcherRepo(db DBTX) *ArcherRepo {
	return &ArcherRepo{db: db}
}

// WithTx returns a new ArcherRepo bound to the given transaction.
func (r *ArcherRepo) WithTx(tx pgx.Tx) *ArcherRepo {
	return &ArcherRepo{db: tx}
}

func scanArcher(scanner interface{ Scan(dest ...any) error }) (model.ArcherRead, error) {
	var (
		a      model.ArcherRead
		dobRaw any
	)

	err := scanner.Scan(
		&a.ArcherID,
		&a.FirstName,
		&a.LastName,
		&a.Email,
		&dobRaw,
		&a.Gender,
		&a.Bowstyle,
		&a.DrawWeight,
		&a.ClubID,
		&a.GooglePictureURL,
		&a.GoogleSubject,
		&a.LastLoginAt,
		&a.CreatedAt,
	)
	if err != nil {
		return model.ArcherRead{}, err
	}

	switch v := dobRaw.(type) {
	case time.Time:
		a.DateOfBirth = v.Format("2006-01-02")
	case string:
		a.DateOfBirth = v
	default:
		return model.ArcherRead{}, fmt.Errorf("unexpected type for date_of_birth: %T", dobRaw)
	}

	return a, nil
}

// FindByID retrieves an archer by primary key identifier.
func (r *ArcherRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	sql, args, err := stmtBuilder.Select(archerColumns...).
		From("archer").
		Where(squirrel.Eq{"archer_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.ArcherRead, error) {
		return scanArcher(r)
	})
}

// FindByEmail retrieves an archer by email address.
func (r *ArcherRepo) FindByEmail(ctx context.Context, email string) (*model.ArcherRead, error) {
	sql, args, err := stmtBuilder.Select(archerColumns...).
		From("archer").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by email query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.ArcherRead, error) {
		return scanArcher(r)
	})
}

// FindByGoogleSubject retrieves an archer by OAuth Google Subject identifier.
func (r *ArcherRepo) FindByGoogleSubject(ctx context.Context, sub string) (*model.ArcherRead, error) {
	sql, args, err := stmtBuilder.Select(archerColumns...).
		From("archer").
		Where(squirrel.Eq{"google_subject": sub}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by google subject query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.ArcherRead, error) {
		return scanArcher(r)
	})
}

// FindAll queries all archers matching the optional criteria in filter.
func (r *ArcherRepo) FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
	q := stmtBuilder.Select(archerColumns...).
		From("archer").
		OrderBy("created_at DESC")

	if filter.ArcherID != nil {
		q = q.Where(squirrel.Eq{"archer_id": *filter.ArcherID})
	}
	if filter.FirstName != nil {
		q = q.Where(squirrel.Eq{"first_name": *filter.FirstName})
	}
	if filter.LastName != nil {
		q = q.Where(squirrel.Eq{"last_name": *filter.LastName})
	}
	if filter.Gender != nil {
		q = q.Where(squirrel.Eq{"gender": *filter.Gender})
	}
	if filter.Bowstyle != nil {
		q = q.Where(squirrel.Eq{"bowstyle": *filter.Bowstyle})
	}
	if filter.DrawWeight != nil {
		q = q.Where(squirrel.Eq{"draw_weight": *filter.DrawWeight})
	}
	if filter.ClubID != nil {
		q = q.Where(squirrel.Eq{"club_id": *filter.ClubID})
	}
	if filter.GoogleSubject != nil {
		q = q.Where(squirrel.Eq{"google_subject": *filter.GoogleSubject})
	}
	if filter.LastLoginAt != nil {
		q = q.Where(squirrel.Eq{"last_login_at": *filter.LastLoginAt})
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying archers: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.ArcherRead, error) {
		return scanArcher(r)
	})
}

// Create inserts a new archer row, returning the generated UUID.
func (r *ArcherRepo) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	dob, err := time.Parse("2006-01-02", data.DateOfBirth)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing date_of_birth: %w", err)
	}

	sql, args, err := stmtBuilder.Insert("archer").
		Columns(
			"first_name",
			"last_name",
			"email",
			"date_of_birth",
			"gender",
			"bowstyle",
			"draw_weight",
			"club_id",
			"google_picture_url",
			"google_subject",
		).
		Values(
			data.FirstName,
			data.LastName,
			data.Email,
			dob,
			data.Gender,
			data.Bowstyle,
			data.DrawWeight,
			data.ClubID,
			data.GooglePictureURL,
			data.GoogleSubject,
		).
		Suffix("RETURNING archer_id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting archer: %w", err)
	}

	return newID, nil
}

// Update updates fields of an archer specified by data for rows matching filter.
func (r *ArcherRepo) Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
	q := stmtBuilder.Update("archer")
	setCount := 0

	if data.FirstName != nil {
		q = q.Set("first_name", *data.FirstName)
		setCount++
	}
	if data.LastName != nil {
		q = q.Set("last_name", *data.LastName)
		setCount++
	}
	if data.Gender != nil {
		q = q.Set("gender", *data.Gender)
		setCount++
	}
	if data.Bowstyle != nil {
		q = q.Set("bowstyle", *data.Bowstyle)
		setCount++
	}
	if data.DrawWeight != nil {
		q = q.Set("draw_weight", *data.DrawWeight)
		setCount++
	}
	if data.ClubID != nil {
		q = q.Set("club_id", *data.ClubID)
		setCount++
	}
	if data.GooglePictureURL != nil {
		q = q.Set("google_picture_url", *data.GooglePictureURL)
		setCount++
	}
	if data.LastLoginAt != nil {
		q = q.Set("last_login_at", *data.LastLoginAt)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.ArcherID != nil {
		q = q.Where(squirrel.Eq{"archer_id": *filter.ArcherID})
		whereCount++
	}
	if filter.FirstName != nil {
		q = q.Where(squirrel.Eq{"first_name": *filter.FirstName})
		whereCount++
	}
	if filter.LastName != nil {
		q = q.Where(squirrel.Eq{"last_name": *filter.LastName})
		whereCount++
	}
	if filter.Gender != nil {
		q = q.Where(squirrel.Eq{"gender": *filter.Gender})
		whereCount++
	}
	if filter.Bowstyle != nil {
		q = q.Where(squirrel.Eq{"bowstyle": *filter.Bowstyle})
		whereCount++
	}
	if filter.DrawWeight != nil {
		q = q.Where(squirrel.Eq{"draw_weight": *filter.DrawWeight})
		whereCount++
	}
	if filter.ClubID != nil {
		q = q.Where(squirrel.Eq{"club_id": *filter.ClubID})
		whereCount++
	}
	if filter.GoogleSubject != nil {
		q = q.Where(squirrel.Eq{"google_subject": *filter.GoogleSubject})
		whereCount++
	}
	if filter.LastLoginAt != nil {
		q = q.Where(squirrel.Eq{"last_login_at": *filter.LastLoginAt})
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

// Delete removes an archer by primary key identifier.
func (r *ArcherRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := stmtBuilder.Delete("archer").
		Where(squirrel.Eq{"archer_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/repository -run "TestArcherRepo" -v`
Expected: PASS

- [ ] **Step 5: Commit Archer repository**

```bash
git add backend/internal/repository/archer.go backend/internal/repository/archer_test.go
git commit -m "feat: add ArcherRepo with Squirrel CRUD queries and tests"
```

---

### Task 4: Maintenance Repository (`maintenance.go` & `maintenance_test.go`)

**Files:**
- Create: `backend/internal/repository/maintenance.go`
- Create: `backend/internal/repository/maintenance_test.go`

**Interfaces:**
- Consumes: `DBTX`
- Produces:
  - `type MaintenanceRepo struct { db DBTX }`
  - `func NewMaintenanceRepo(db DBTX) *MaintenanceRepo`
  - `func (r *MaintenanceRepo) RefreshOpenParticipants(ctx context.Context) error`
  - `func (r *MaintenanceRepo) GetSchemaVersion(ctx context.Context) (int64, error)`

- [ ] **Step 1: Write failing tests for MaintenanceRepo**

Create `backend/internal/repository/maintenance_test.go`:

```go
package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestMaintenanceRepo_RefreshOpenParticipants_Success(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			expectedSQL := "REFRESH MATERIALIZED VIEW CONCURRENTLY open_participants"
			if sql != expectedSQL {
				t.Errorf("expected SQL %q, got %q", expectedSQL, sql)
			}
			return pgconn.NewCommandTag("REFRESH MATERIALIZED VIEW"), nil
		},
	}

	repo := repository.NewMaintenanceRepo(mock)
	err := repo.RefreshOpenParticipants(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMaintenanceRepo_RefreshOpenParticipants_Error(t *testing.T) {
	dbErr := errors.New("database failure")
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, dbErr
		},
	}

	repo := repository.NewMaintenanceRepo(mock)
	err := repo.RefreshOpenParticipants(context.Background())
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

func TestMaintenanceRepo_GetSchemaVersion_Success(t *testing.T) {
	expectedVersion := int64(6)
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					v := dest[0].(*int64)
					*v = expectedVersion
					return nil
				},
			}
		},
	}

	repo := repository.NewMaintenanceRepo(mock)
	version, err := repo.GetSchemaVersion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != expectedVersion {
		t.Errorf("expected version %d, got %d", expectedVersion, version)
	}
}

func TestMaintenanceRepo_GetSchemaVersion_Error(t *testing.T) {
	dbErr := errors.New("relation does not exist")
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return dbErr
				},
			}
		},
	}

	repo := repository.NewMaintenanceRepo(mock)
	_, err := repo.GetSchemaVersion(context.Background())
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/repository -run "TestMaintenanceRepo" -v`
Expected: FAIL with compilation error (undefined `repository.NewMaintenanceRepo`)

- [ ] **Step 3: Implement `backend/internal/repository/maintenance.go`**

Create `backend/internal/repository/maintenance.go`:

```go
package repository

import (
	"context"
	"fmt"
)

// MaintenanceRepo manages administrative and maintenance database operations.
type MaintenanceRepo struct {
	db DBTX
}

// NewMaintenanceRepo constructs a new MaintenanceRepo backed by DBTX.
func NewMaintenanceRepo(db DBTX) *MaintenanceRepo {
	return &MaintenanceRepo{db: db}
}

// RefreshOpenParticipants concurrently refreshes the open_participants materialized view.
func (r *MaintenanceRepo) RefreshOpenParticipants(ctx context.Context) error {
	if _, err := r.db.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY open_participants"); err != nil {
		return fmt.Errorf("refreshing open_participants materialized view: %w", err)
	}
	return nil
}

// GetSchemaVersion retrieves the current applied goose schema version from goose_db_version.
func (r *MaintenanceRepo) GetSchemaVersion(ctx context.Context) (int64, error) {
	var version int64
	row := r.db.QueryRow(ctx, "SELECT COALESCE(MAX(version_id), 0) FROM goose_db_version WHERE is_applied = true")
	if err := row.Scan(&version); err != nil {
		return 0, fmt.Errorf("getting schema version: %w", err)
	}
	return version, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/repository -run "TestMaintenanceRepo" -v`
Expected: PASS

- [ ] **Step 5: Commit Maintenance repository**

```bash
git add backend/internal/repository/maintenance.go backend/internal/repository/maintenance_test.go
git commit -m "feat: add MaintenanceRepo for materialized view refresh and schema version check"
```

---

### Task 5: Reporting Repository Stubs (`reporting.go` & `reporting_test.go`)

**Files:**
- Create: `backend/internal/repository/reporting.go`
- Create: `backend/internal/repository/reporting_test.go`

**Interfaces:**
- Consumes:
  - `DBTX`
  - `model.SessionSummaryReport`, `model.ScoringTrend`
- Produces:
  - `type ReportingRepo struct { db DBTX }`
  - `func NewReportingRepo(db DBTX) *ReportingRepo`
  - `func (r *ReportingRepo) GetSessionSummary(ctx context.Context, sessionID uuid.UUID) (*model.SessionSummaryReport, error)`
  - `func (r *ReportingRepo) GetArcherPerformance(ctx context.Context, archerID uuid.UUID, from, to time.Time) ([]model.ScoringTrend, error)`

- [ ] **Step 1: Write failing tests for ReportingRepo**

Create `backend/internal/repository/reporting_test.go`:

```go
package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestReportingRepo_GetSessionSummary_Stub(t *testing.T) {
	repo := repository.NewReportingRepo(&mockDBTX{})
	sessionID := uuid.New()

	summary, err := repo.GetSessionSummary(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected summary report, got nil")
	}
	if summary.SessionID != sessionID {
		t.Errorf("expected sessionID %v, got %v", sessionID, summary.SessionID)
	}
}

func TestReportingRepo_GetArcherPerformance_Stub(t *testing.T) {
	repo := repository.NewReportingRepo(&mockDBTX{})
	archerID := uuid.New()
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	trends, err := repo.GetArcherPerformance(context.Background(), archerID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trends == nil {
		t.Fatal("expected trends slice, got nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/repository -run "TestReportingRepo" -v`
Expected: FAIL with compilation error (undefined `repository.NewReportingRepo`)

- [ ] **Step 3: Implement `backend/internal/repository/reporting.go`**

Create `backend/internal/repository/reporting.go`:

```go
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ReportingRepo provides cross-domain analytical reporting queries.
// Methods accept DBTX and will execute raw SQL or aggregation queries as reporting features are developed.
type ReportingRepo struct {
	db DBTX
}

// NewReportingRepo constructs a new ReportingRepo backed by DBTX.
func NewReportingRepo(db DBTX) *ReportingRepo {
	return &ReportingRepo{db: db}
}

// GetSessionSummary returns aggregated performance statistics for a session.
// Initial stub returns placeholder data until analytics endpoints are implemented.
func (r *ReportingRepo) GetSessionSummary(ctx context.Context, sessionID uuid.UUID) (*model.SessionSummaryReport, error) {
	return &model.SessionSummaryReport{
		SessionID:       sessionID,
		SessionLocation: "Stub Location",
		TotalShots:      0,
		AverageScore:    0.0,
		StartedAt:       time.Now(),
	}, nil
}

// GetArcherPerformance returns historical scoring progression data points for an archer.
// Initial stub returns empty slice until analytics endpoints are implemented.
func (r *ReportingRepo) GetArcherPerformance(ctx context.Context, archerID uuid.UUID, from, to time.Time) ([]model.ScoringTrend, error) {
	return make([]model.ScoringTrend, 0), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/repository -run "TestReportingRepo" -v`
Expected: PASS

- [ ] **Step 5: Commit Reporting repository**

```bash
git add backend/internal/repository/reporting.go backend/internal/repository/reporting_test.go
git commit -m "feat: add ReportingRepo stubs for session summary and archer performance"
```

---

### Task 6: Formatting, Linting, Vetting, and Full Suite Verification

**Files:**
- Repository package files under `backend/internal/repository/`

**Interfaces:**
- Consumes: Go toolchain, `gofumpt`, `golangci-lint`
- Produces: 100% clean compilation, linting, formatting, and unit tests

- [ ] **Step 1: Run code formatting**

```bash
cd backend
gofumpt -l -w internal/repository/
```

- [ ] **Step 2: Run golangci-lint**

```bash
cd backend
golangci-lint run ./...
```
Expected: PASS with 0 lint errors

- [ ] **Step 3: Run go vet**

```bash
cd backend
go vet ./...
```
Expected: PASS with 0 vet warnings

- [ ] **Step 4: Run all unit tests with race detector**

```bash
cd backend
go test -race ./... -v
```
Expected: PASS for all packages (`internal/apperror`, `internal/config`, `internal/model`, `internal/repository`)

- [ ] **Step 5: Verify build compiles**

```bash
cd backend
go build ./...
```
Expected: PASS

- [ ] **Step 6: Commit any final formatting adjustments if produced**

```bash
git status --porcelain
# If files were modified by gofumpt:
# git add backend/internal/repository/
# git commit -m "style: format repository package with gofumpt"
```

---

## Self-Review Checklist

- **Spec Coverage:**
  - `base.go` defines `DBTX` interface (`Query`, `QueryRow`, `Exec`) -> Task 2
  - `base.go` defines common helper functions for scanning rows -> Task 2 (`ScanRows`, `ScanOne`)
  - `base.go` defines `WithTx(ctx, pool, fn)` transaction wrapper -> Task 2
  - `archer.go` implements `ArcherRepo` (`FindByID`, `FindByEmail`, `FindByGoogleSubject`, `FindAll`, `Create`, `Update`, `Delete`) -> Task 3
  - `maintenance.go` implements `MaintenanceRepo` (`RefreshOpenParticipants`, `GetSchemaVersion`) -> Task 4
  - `reporting.go` implements `ReportingRepo` (`GetSessionSummary`, `GetArcherPerformance`) -> Task 5
  - All queries use `squirrel` with PostgreSQL dollar placeholder format -> Task 2 & 3
  - Unit tests with mock DBTX verify query building logic -> Tasks 2, 3, 4, 5
  - `go test ./internal/repository/...` passes -> Verified in Tasks 2, 3, 4, 5, 6
  - `go vet ./...` reports no issues -> Task 6
- **Placeholder Scan:**
  - No "TBD", "TODO", "implement later", or omitted code blocks exist in any step.
- **Type Consistency:**
  - `DBTX` matches exactly across `base.go`, `archer.go`, `maintenance.go`, `reporting.go`.
  - `model.ArcherRead`, `model.ArcherCreate`, `model.ArcherSet`, `model.ArcherFilter` types match `internal/model/archer.go`.
  - Date parsing formats `2006-01-02` match date-only convention in `model.ArcherRead.DateOfBirth`.
