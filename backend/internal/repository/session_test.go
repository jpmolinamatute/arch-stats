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
	if len(executedArgs) != 1 || (executedArgs[0] != sessionID && executedArgs[0] != sessionID.String()) {
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
	if executedArgs[0] != archerID && executedArgs[0] != archerID.String() {
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
		case string:
			if v == sessionID.String() {
				foundSessionID = true
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
	if len(executedArgs) != 1 || (executedArgs[0] != sessionID && executedArgs[0] != sessionID.String()) {
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

func TestSessionRepo_FindParticipating_Success(t *testing.T) {
	archerID := uuid.New()
	expectedSessionID := uuid.New()

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if !strings.Contains(sql, "FROM slot s") || !strings.Contains(sql, "JOIN session ses") {
				t.Errorf("unexpected query SQL: %s", sql)
			}
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = expectedSessionID
					return nil
				},
			}
		},
	}

	repo := repository.NewSessionRepo(mock)
	got, err := repo.FindParticipating(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != expectedSessionID {
		t.Fatalf("expected session ID %v, got %v", expectedSessionID, got)
	}
}

func TestSessionRepo_FindParticipating_NotFound(t *testing.T) {
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
	got, err := repo.FindParticipating(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil session ID, got %v", got)
	}
}

func TestSessionRepo_FindParticipating_DBError(t *testing.T) {
	dbErr := errors.New("query participating error")
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
	_, err := repo.FindParticipating(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
}
