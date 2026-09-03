# Task 010: Build Repository — Shooting Session Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the shooting session repository (`SessionRepo`) for managing archery shooting sessions (creation, retrieval, filtering, active session lookup, dynamic updates, closing, and deletion) using Squirrel SQL query building and `pgx/v5` against the PostgreSQL `session` table.

**Architecture:** The `SessionRepo` operates on the `DBTX` interface (`*pgxpool.Pool` or `pgx.Tx`), building PostgreSQL parameterized queries with Squirrel targeting the `session` table. It exposes CRUD operations plus domain-specific lifecycle methods (`FindOpen` to fetch an archer's active session, and `Close` to transition `is_opened = false` with a `closed_at` timestamp). Unit tests utilize mocked `DBTX` implementations to verify exact SQL generation, parameter binding, and defensive row scanning without requiring a live database.

**Tech Stack:** Go 1.27+, `github.com/Masterminds/squirrel`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`, standard library (`context`, `errors`, `fmt`, `time`).

**Spec:** [docs/go_refactor/tasks/010-repository_shooting_session.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/010-repository_shooting_session.md)

## Global Constraints

- Git branch: `refactor/010-repository-shooting-session`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/repository`
- All SQL queries must use PostgreSQL dollar placeholder format (`$1`, `$2`) via `StmtBuilder` (`squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)`).
- No ORMs; use `pgx/v5` only via `DBTX`.
- Error handling: Wrap errors with `%w` using contextual descriptive messages.
- Schema alignment note: PostgreSQL migration `004_2025-09-26_shooting_sessions_table.sql` and domain model `model.SessionRead` define columns `session_id`, `owner_archer_id`, `session_location`, `is_indoor`, `is_opened`, `created_at`, `closed_at`. The task spec's informal references to `status`, `archer_id`, and `ended_at` map directly to `is_opened` (bool), `owner_archer_id` (UUID), and `closed_at` (`TIMESTAMPTZ` / `*time.Time`).
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
        ├── archer_test.go             # [MODIFY] Add bool and *bool scan support to mockMultiRows
        ├── session.go                 # [NEW] SessionRepo struct, NewSessionRepo, WithTx, FindByID, FindAll, FindOpen, Create, Update, Close, Delete
        └── session_test.go            # [NEW] Unit test suite covering all SessionRepo methods using mockDBTX
```

---

### Task 1: Git Branch Setup & Mock Scanner Update

**Files:**
- Modify: `backend/internal/repository/archer_test.go:110-138`

**Interfaces:**
- Consumes: `mockMultiRows` from `backend/internal/repository/archer_test.go`
- Produces: Enhanced `mockMultiRows.Scan` supporting destination types `*bool` and `**bool`

- [ ] **Step 1: Check out git branch**

```bash
git switch -c refactor/010-repository-shooting-session
```

- [ ] **Step 2: Add `*bool` and `**bool` scanning to `mockMultiRows.Scan` in `archer_test.go`**

In `backend/internal/repository/archer_test.go`, update `mockMultiRows.Scan` around lines 135-138 to include:

```go
		case *bool:
			*d = v.(bool)
		case **bool:
			switch val := v.(type) {
			case nil:
				*d = nil
			case *bool:
				*d = val
			case bool:
				*d = &val
			}
```

- [ ] **Step 3: Run existing repository tests to verify no regression**

Run: `cd backend && go test ./internal/repository/... -v`
Expected: PASS

- [ ] **Step 4: Commit mock scanner update**

```bash
git add backend/internal/repository/archer_test.go
git commit -m "test(repository): support bool scanning in mockMultiRows"
```

---

### Task 2: Session Repository Unit Tests (TDD - Failing Tests)

**Files:**
- Create: `backend/internal/repository/session_test.go`

**Interfaces:**
- Consumes:
  - `model.SessionCreate`, `model.SessionSet`, `model.SessionFilter`, `model.SessionRead` from `internal/model`
  - `repository.NewSessionRepo(db DBTX)`
  - `mockDBTX`, `mockMultiRows` from `archer_test.go`
  - `mockSingleRow`, `mockTx` from `base_test.go`
- Produces:
  - Comprehensive unit test suite covering `FindByID`, `FindOpen`, `FindAll`, `Create`, `Update`, `Close`, `Delete`, and `WithTx`.

- [ ] **Step 1: Write failing unit tests in `backend/internal/repository/session_test.go`**

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

func sampleSessionRow(sessionID, archerID uuid.UUID, location string, isIndoor, isOpened bool, closedAt *time.Time) []any {
	now := time.Now().Truncate(time.Second).UTC()
	return []any{
		sessionID,
		archerID,
		location,
		isIndoor,
		isOpened,
		now,
		closedAt,
	}
}

func TestSessionRepo_FindByID_Success(t *testing.T) {
	sessionID := uuid.New()
	archerID := uuid.New()
	location := "Outdoor Range 1"

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleSessionRow(sessionID, archerID, location, false, true, nil)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewSessionRepo(mock)
	session, err := repo.FindByID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}

	if !strings.HasPrefix(executedSQL, "SELECT session_id, owner_archer_id, session_location, is_indoor, is_opened, created_at, closed_at FROM session") {
		t.Errorf("unexpected query SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE session_id = $1") {
		t.Errorf("expected WHERE session_id = $1, got: %s", executedSQL)
	}
	if len(executedArgs) != 1 || executedArgs[0] != sessionID {
		t.Errorf("expected arg %v, got %v", sessionID, executedArgs)
	}

	if session.SessionID != sessionID {
		t.Errorf("expected SessionID %v, got %v", sessionID, session.SessionID)
	}
	if session.OwnerArcherID != archerID {
		t.Errorf("expected OwnerArcherID %v, got %v", archerID, session.OwnerArcherID)
	}
	if session.SessionLocation != location {
		t.Errorf("expected SessionLocation %s, got %s", location, session.SessionLocation)
	}
	if session.IsIndoor != false {
		t.Errorf("expected IsIndoor false, got %v", session.IsIndoor)
	}
	if session.IsOpened != true {
		t.Errorf("expected IsOpened true, got %v", session.IsOpened)
	}
	if session.ClosedAt != nil {
		t.Errorf("expected ClosedAt nil, got %v", session.ClosedAt)
	}
}

func TestSessionRepo_FindByID_NotFound(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewSessionRepo(mock)
	session, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session != nil {
		t.Fatalf("expected nil session on not found, got %+v", session)
	}
}

func TestSessionRepo_FindByID_DBError(t *testing.T) {
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

	repo := repository.NewSessionRepo(mock)
	session, err := repo.FindByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if session != nil {
		t.Fatalf("expected nil session on error, got %+v", session)
	}
}

func TestSessionRepo_FindOpen_Success(t *testing.T) {
	sessionID := uuid.New()
	archerID := uuid.New()
	location := "Indoor Target 5"

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleSessionRow(sessionID, archerID, location, true, true, nil)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewSessionRepo(mock)
	session, err := repo.FindOpen(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}

	if !strings.HasPrefix(executedSQL, "SELECT session_id, owner_archer_id, session_location, is_indoor, is_opened, created_at, closed_at FROM session") {
		t.Errorf("unexpected query SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE") || !strings.Contains(executedSQL, "owner_archer_id = $1") || !strings.Contains(executedSQL, "is_opened = $2") {
		t.Errorf("expected WHERE owner_archer_id = $1 AND is_opened = $2, got: %s", executedSQL)
	}
	if len(executedArgs) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(executedArgs))
	}
	if executedArgs[0] != archerID {
		t.Errorf("arg[0] expected archerID %v, got %v", archerID, executedArgs[0])
	}
	if executedArgs[1] != true {
		t.Errorf("arg[1] expected true, got %v", executedArgs[1])
	}
	if session.SessionID != sessionID || session.OwnerArcherID != archerID || !session.IsOpened {
		t.Errorf("scanned session fields unexpected: %+v", session)
	}
}

func TestSessionRepo_FindOpen_NotFound(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewSessionRepo(mock)
	session, err := repo.FindOpen(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session != nil {
		t.Fatalf("expected nil session when no open session, got %+v", session)
	}
}

func TestSessionRepo_FindOpen_DBError(t *testing.T) {
	dbErr := errors.New("open session query failure")
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return dbErr
				},
			}
		},
	}

	repo := repository.NewSessionRepo(mock)
	session, err := repo.FindOpen(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if session != nil {
		t.Fatalf("expected nil session on error, got %+v", session)
	}
}

func TestSessionRepo_FindAll_WithFilters(t *testing.T) {
	sessionID := uuid.New()
	archerID := uuid.New()
	now := time.Now().Truncate(time.Second).UTC()
	closedAt := now.Add(2 * time.Hour)
	loc := "Range North"
	isOpened := false
	isIndoor := true

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			executedSQL = sql
			executedArgs = args
			return &mockMultiRows{
				records: [][]any{
					sampleSessionRow(sessionID, archerID, loc, isIndoor, isOpened, &closedAt),
				},
			}, nil
		},
	}

	filter := model.SessionFilter{
		SessionID:       &sessionID,
		OwnerArcherID:   &archerID,
		CreatedAt:       &now,
		ClosedAt:        &closedAt,
		SessionLocation: &loc,
		IsOpened:        &isOpened,
		IsIndoor:        &isIndoor,
	}

	repo := repository.NewSessionRepo(mock)
	sessions, err := repo.FindAll(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	expectedClauses := []string{
		"session_id = $",
		"owner_archer_id = $",
		"created_at = $",
		"closed_at = $",
		"session_location = $",
		"is_opened = $",
		"is_indoor = $",
		"ORDER BY created_at DESC",
	}
	for _, clause := range expectedClauses {
		if !strings.Contains(executedSQL, clause) {
			t.Errorf("expected query to contain %q, got: %s", clause, executedSQL)
		}
	}
	if len(executedArgs) != 7 {
		t.Errorf("expected 7 arguments for full filter, got %d", len(executedArgs))
	}

	s := sessions[0]
	if s.SessionID != sessionID || s.OwnerArcherID != archerID || s.SessionLocation != loc || s.ClosedAt == nil {
		t.Errorf("scanned session fields mismatch: %+v", s)
	}
}

func TestSessionRepo_FindAll_DBError(t *testing.T) {
	dbErr := errors.New("find all error")
	mock := &mockDBTX{
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return nil, dbErr
		},
	}

	repo := repository.NewSessionRepo(mock)
	sessions, err := repo.FindAll(context.Background(), model.SessionFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if sessions != nil {
		t.Fatalf("expected nil slice on error, got %+v", sessions)
	}
}

func TestSessionRepo_Create_Success(t *testing.T) {
	newID := uuid.New()
	archerID := uuid.New()
	location := "Indoor Club Lane 3"
	isIndoor := true
	isOpened := true

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					d := dest[0].(*uuid.UUID)
					*d = newID
					return nil
				},
			}
		},
	}

	repo := repository.NewSessionRepo(mock)
	createdID, err := repo.Create(context.Background(), model.SessionCreate{
		OwnerArcherID:   archerID,
		SessionLocation: location,
		IsIndoor:        isIndoor,
		IsOpened:        isOpened,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdID != newID {
		t.Errorf("expected created UUID %v, got %v", newID, createdID)
	}

	if !strings.HasPrefix(executedSQL, "INSERT INTO session") {
		t.Errorf("expected INSERT INTO session query, got: %s", executedSQL)
	}
	expectedCols := []string{"owner_archer_id", "session_location", "is_indoor", "is_opened"}
	for _, col := range expectedCols {
		if !strings.Contains(executedSQL, col) {
			t.Errorf("expected query to contain column %q, got: %s", col, executedSQL)
		}
	}
	if !strings.Contains(executedSQL, "RETURNING session_id") {
		t.Errorf("expected RETURNING session_id suffix, got: %s", executedSQL)
	}

	if len(executedArgs) != 4 {
		t.Fatalf("expected 4 arguments, got %d", len(executedArgs))
	}
	if executedArgs[0] != archerID {
		t.Errorf("arg[0] expected archerID %v, got %v", archerID, executedArgs[0])
	}
	if executedArgs[1] != location {
		t.Errorf("arg[1] expected location %s, got %v", location, executedArgs[1])
	}
	if executedArgs[2] != isIndoor {
		t.Errorf("arg[2] expected isIndoor %v, got %v", isIndoor, executedArgs[2])
	}
	if executedArgs[3] != isOpened {
		t.Errorf("arg[3] expected isOpened %v, got %v", isOpened, executedArgs[3])
	}
}

func TestSessionRepo_Create_DBError(t *testing.T) {
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

	repo := repository.NewSessionRepo(mock)
	id, err := repo.Create(context.Background(), model.SessionCreate{
		OwnerArcherID:   uuid.New(),
		SessionLocation: "Field",
		IsIndoor:        false,
		IsOpened:        true,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if id != uuid.Nil {
		t.Errorf("expected uuid.Nil on error, got %v", id)
	}
}

func TestSessionRepo_Update_Success(t *testing.T) {
	sessionID := uuid.New()
	newLoc := "Updated Field Range"
	newIndoor := false
	newOpened := false
	now := time.Now().UTC()

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := repository.NewSessionRepo(mock)
	err := repo.Update(
		context.Background(),
		model.SessionSet{
			SessionLocation: &newLoc,
			IsIndoor:        &newIndoor,
			IsOpened:        &newOpened,
			ClosedAt:        &now,
		},
		model.SessionFilter{
			SessionID: &sessionID,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(executedSQL, "UPDATE session") {
		t.Errorf("expected UPDATE session query, got: %s", executedSQL)
	}
	expectedSets := []string{
		"session_location = $",
		"is_indoor = $",
		"is_opened = $",
		"closed_at = $",
	}
	for _, s := range expectedSets {
		if !strings.Contains(executedSQL, s) {
			t.Errorf("expected query to contain set clause %q, got: %s", s, executedSQL)
		}
	}
	if !strings.Contains(executedSQL, "WHERE session_id = $") {
		t.Errorf("expected query to contain WHERE session_id = $, got: %s", executedSQL)
	}
	if len(executedArgs) != 5 {
		t.Errorf("expected 5 args (4 sets + 1 where), got %d", len(executedArgs))
	}
}

func TestSessionRepo_Update_RowsAffectedZeroReturnsNotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}

	repo := repository.NewSessionRepo(mock)
	loc := "Test"
	id := uuid.New()
	err := repo.Update(
		context.Background(),
		model.SessionSet{SessionLocation: &loc},
		model.SessionFilter{SessionID: &id},
	)
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected apperror.ErrNotFound, got: %v", err)
	}
}

func TestSessionRepo_Update_EmptySetReturnsNil(t *testing.T) {
	execCalled := false
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			execCalled = true
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := repository.NewSessionRepo(mock)
	id := uuid.New()
	err := repo.Update(context.Background(), model.SessionSet{}, model.SessionFilter{SessionID: &id})
	if err != nil {
		t.Fatalf("expected nil error on empty set, got: %v", err)
	}
	if execCalled {
		t.Error("expected Exec not to be called for empty set")
	}
}

func TestSessionRepo_Update_MissingFilterReturnsError(t *testing.T) {
	repo := repository.NewSessionRepo(&mockDBTX{})
	loc := "Range"
	err := repo.Update(context.Background(), model.SessionSet{SessionLocation: &loc}, model.SessionFilter{})
	if err == nil {
		t.Fatal("expected error on empty filter, got nil")
	}
	if !strings.Contains(err.Error(), "requires at least one filter condition") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSessionRepo_Close_Success(t *testing.T) {
	sessionID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	before := time.Now().UTC().Add(-time.Second)
	repo := repository.NewSessionRepo(mock)
	err := repo.Close(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := time.Now().UTC().Add(time.Second)

	if !strings.HasPrefix(executedSQL, "UPDATE session") {
		t.Errorf("expected UPDATE session query, got: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "is_opened = $") {
		t.Errorf("expected query to set is_opened, got: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "closed_at = $") {
		t.Errorf("expected query to set closed_at, got: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE") || !strings.Contains(executedSQL, "session_id = $") || !strings.Contains(executedSQL, "is_opened = $") {
		t.Errorf("expected query to filter by session_id and is_opened = true, got: %s", executedSQL)
	}

	if len(executedArgs) != 4 {
		t.Fatalf("expected 4 arguments (2 sets + 2 where clauses), got %d: %v", len(executedArgs), executedArgs)
	}

	// Verify boolean value false was set for is_opened
	foundFalse := false
	foundTimestamp := false
	foundSessionID := false
	foundTrue := false

	for _, arg := range executedArgs {
		switch v := arg.(type) {
		case bool:
			if v {
				foundTrue = true
			} else {
				foundFalse = true
			}
		case time.Time:
			if !v.Before(before) && !v.After(after) {
				foundTimestamp = true
			}
		case uuid.UUID:
			if v == sessionID {
				foundSessionID = true
			}
		}
	}

	if !foundFalse {
		t.Errorf("expected is_opened = false to be set, args: %v", executedArgs)
	}
	if !foundTimestamp {
		t.Errorf("expected closed_at timestamp between %v and %v, args: %v", before, after, executedArgs)
	}
	if !foundSessionID {
		t.Errorf("expected sessionID %v in where clause, args: %v", sessionID, executedArgs)
	}
	if !foundTrue {
		t.Errorf("expected is_opened = true in where condition, args: %v", executedArgs)
	}
}

func TestSessionRepo_Close_NotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}

	repo := repository.NewSessionRepo(mock)
	err := repo.Close(context.Background(), uuid.New())
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected apperror.ErrNotFound when closing nonexistent or already-closed session, got: %v", err)
	}
}

func TestSessionRepo_Close_DBError(t *testing.T) {
	dbErr := errors.New("db failure closing session")
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, dbErr
		},
	}

	repo := repository.NewSessionRepo(mock)
	err := repo.Close(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
}

func TestSessionRepo_Delete_Success(t *testing.T) {
	sessionID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("DELETE 1"), nil
		},
	}

	repo := repository.NewSessionRepo(mock)
	err := repo.Delete(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(executedSQL, "DELETE FROM session") {
		t.Errorf("expected DELETE FROM session query, got: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE session_id = $1") {
		t.Errorf("expected WHERE session_id = $1, got: %s", executedSQL)
	}
	if len(executedArgs) != 1 || executedArgs[0] != sessionID {
		t.Errorf("expected argument %v, got %v", sessionID, executedArgs)
	}
}

func TestSessionRepo_Delete_NotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("DELETE 0"), nil
		},
	}

	repo := repository.NewSessionRepo(mock)
	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected apperror.ErrNotFound on delete absent, got: %v", err)
	}
}

func TestSessionRepo_Delete_DBError(t *testing.T) {
	dbErr := errors.New("delete session query error")
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, dbErr
		},
	}

	repo := repository.NewSessionRepo(mock)
	err := repo.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
}

func TestSessionRepo_WithTx(t *testing.T) {
	mock := &mockDBTX{}
	repo := repository.NewSessionRepo(mock)

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
Expected: FAIL with compilation error (`undefined: repository.NewSessionRepo`)

---

### Task 3: Session Repository Implementation (`session.go`)

**Files:**
- Create: `backend/internal/repository/session.go`

**Interfaces:**
- Consumes:
  - `DBTX`, `StmtBuilder`, `ScanOne`, `ScanRows` from `base.go`
  - `model.SessionCreate`, `model.SessionSet`, `model.SessionFilter`, `model.SessionRead` from `model`
  - `apperror.ErrNotFound` from `apperror`
- Produces:
  - `SessionRepo` struct
  - `NewSessionRepo(db DBTX) *SessionRepo`
  - `(r *SessionRepo) WithTx(tx pgx.Tx) *SessionRepo`
  - `(r *SessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)`
  - `(r *SessionRepo) FindOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)`
  - `(r *SessionRepo) FindAll(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)`
  - `(r *SessionRepo) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)`
  - `(r *SessionRepo) Update(ctx context.Context, data model.SessionSet, filter model.SessionFilter) error`
  - `(r *SessionRepo) Close(ctx context.Context, id uuid.UUID) error`
  - `(r *SessionRepo) Delete(ctx context.Context, id uuid.UUID) error`

- [ ] **Step 1: Write `backend/internal/repository/session.go`**

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

var sessionColumns = []string{
	"session_id",
	"owner_archer_id",
	"session_location",
	"is_indoor",
	"is_opened",
	"created_at",
	"closed_at",
}

// SessionRepo manages database operations for archery shooting sessions.
type SessionRepo struct {
	db DBTX
}

// NewSessionRepo constructs a SessionRepo backed by DBTX.
func NewSessionRepo(db DBTX) *SessionRepo {
	return &SessionRepo{db: db}
}

// WithTx returns a new SessionRepo bound to the given transaction.
func (r *SessionRepo) WithTx(tx pgx.Tx) *SessionRepo {
	return &SessionRepo{db: tx}
}

func scanSession(scanner interface{ Scan(dest ...any) error }) (model.SessionRead, error) {
	var s model.SessionRead
	err := scanner.Scan(
		&s.SessionID,
		&s.OwnerArcherID,
		&s.SessionLocation,
		&s.IsIndoor,
		&s.IsOpened,
		&s.CreatedAt,
		&s.ClosedAt,
	)
	if err != nil {
		return model.SessionRead{}, err
	}
	return s, nil
}

// FindByID retrieves a shooting session by its primary key identifier.
// Returns nil, nil if no session exists with the given ID.
func (r *SessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
	sql, args, err := StmtBuilder.Select(sessionColumns...).
		From("session").
		Where(squirrel.Eq{"session_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by id query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.SessionRead, error) {
		return scanSession(r)
	})
}

// FindOpen finds the currently open session owned by the specified archer.
// Returns nil, nil if the archer has no currently open session.
func (r *SessionRepo) FindOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error) {
	sql, args, err := StmtBuilder.Select(sessionColumns...).
		From("session").
		Where(squirrel.Eq{"owner_archer_id": archerID}).
		Where(squirrel.Eq{"is_opened": true}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find open query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.SessionRead, error) {
		return scanSession(r)
	})
}

// FindAll queries all shooting sessions matching the optional criteria in filter.
//
//nolint:gocritic // hugeParam: filter value parameter matches repository interface specification
func (r *SessionRepo) FindAll(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	q := StmtBuilder.Select(sessionColumns...).
		From("session").
		OrderBy("created_at DESC")

	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
	}
	if filter.OwnerArcherID != nil {
		q = q.Where(squirrel.Eq{"owner_archer_id": *filter.OwnerArcherID})
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
	}
	if filter.ClosedAt != nil {
		q = q.Where(squirrel.Eq{"closed_at": *filter.ClosedAt})
	}
	if filter.SessionLocation != nil {
		q = q.Where(squirrel.Eq{"session_location": *filter.SessionLocation})
	}
	if filter.IsOpened != nil {
		q = q.Where(squirrel.Eq{"is_opened": *filter.IsOpened})
	}
	if filter.IsIndoor != nil {
		q = q.Where(squirrel.Eq{"is_indoor": *filter.IsIndoor})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find all query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("querying sessions: %w", err)
	}

	return ScanRows(rows, func(r pgx.Rows) (model.SessionRead, error) {
		return scanSession(r)
	})
}

// Create inserts a new shooting session record and returns the generated UUID identifier.
//
//nolint:gocritic // hugeParam: data value parameter matches repository interface specification
func (r *SessionRepo) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
	sql, args, err := StmtBuilder.Insert("session").
		Columns(
			"owner_archer_id",
			"session_location",
			"is_indoor",
			"is_opened",
		).
		Values(
			data.OwnerArcherID,
			data.SessionLocation,
			data.IsIndoor,
			data.IsOpened,
		).
		Suffix("RETURNING session_id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("building create session query: %w", err)
	}

	var newID uuid.UUID
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("inserting session: %w", err)
	}

	return newID, nil
}

// Update mutates session fields specified in data for rows matching filter.
// Requires at least one filter criterion to prevent unbounded updates.
// Returns apperror.ErrNotFound if no session matched the filter.
//
//nolint:gocritic // hugeParam: filter value parameter matches repository interface specification
func (r *SessionRepo) Update(ctx context.Context, data model.SessionSet, filter model.SessionFilter) error {
	q := StmtBuilder.Update("session")
	setCount := 0

	if data.SessionLocation != nil {
		q = q.Set("session_location", *data.SessionLocation)
		setCount++
	}
	if data.IsIndoor != nil {
		q = q.Set("is_indoor", *data.IsIndoor)
		setCount++
	}
	if data.IsOpened != nil {
		q = q.Set("is_opened", *data.IsOpened)
		setCount++
	}
	if data.ClosedAt != nil {
		q = q.Set("closed_at", *data.ClosedAt)
		setCount++
	}

	if setCount == 0 {
		return nil
	}

	whereCount := 0
	if filter.SessionID != nil {
		q = q.Where(squirrel.Eq{"session_id": *filter.SessionID})
		whereCount++
	}
	if filter.OwnerArcherID != nil {
		q = q.Where(squirrel.Eq{"owner_archer_id": *filter.OwnerArcherID})
		whereCount++
	}
	if filter.CreatedAt != nil {
		q = q.Where(squirrel.Eq{"created_at": *filter.CreatedAt})
		whereCount++
	}
	if filter.ClosedAt != nil {
		q = q.Where(squirrel.Eq{"closed_at": *filter.ClosedAt})
		whereCount++
	}
	if filter.SessionLocation != nil {
		q = q.Where(squirrel.Eq{"session_location": *filter.SessionLocation})
		whereCount++
	}
	if filter.IsOpened != nil {
		q = q.Where(squirrel.Eq{"is_opened": *filter.IsOpened})
		whereCount++
	}
	if filter.IsIndoor != nil {
		q = q.Where(squirrel.Eq{"is_indoor": *filter.IsIndoor})
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

// Close marks an active shooting session as closed, setting is_opened = false and closed_at = current UTC timestamp.
// If the session does not exist or is already closed, it returns apperror.ErrNotFound.
func (r *SessionRepo) Close(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	sql, args, err := StmtBuilder.Update("session").
		Set("is_opened", false).
		Set("closed_at", now).
		Where(squirrel.Eq{"session_id": id}).
		Where(squirrel.Eq{"is_opened": true}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building close query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing close session: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// Delete removes a shooting session by its primary key identifier.
// Returns apperror.ErrNotFound if no session existed with the given ID.
func (r *SessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("session").
		Where(squirrel.Eq{"session_id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete session query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete session: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
```

- [ ] **Step 2: Run unit tests to verify they pass**

Run: `cd backend && go test ./internal/repository/... -v`
Expected: PASS

- [ ] **Step 3: Commit repository implementation**

```bash
git add backend/internal/repository/session.go backend/internal/repository/session_test.go
git commit -m "feat(repository): implement shooting session repository"
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
git commit -m "chore(repository): format and lint shooting session repository"
```
