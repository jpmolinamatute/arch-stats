# Task 009: Build Repository — Auth Session Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the auth session repository (`AuthSessionRepo`) for managing authentication session records (creation, lookup by token hash, archer session invalidation, and expired session cleanup) using Squirrel query building and `pgx/v5` against the PostgreSQL `auth` table.

**Architecture:** The `AuthSessionRepo` operates on the `DBTX` interface (`*pgxpool.Pool` or `pgx.Tx`), utilizing parameterized Squirrel SQL queries targeting the `auth` table. Sessions store SHA-256 hashed tokens, client metadata (UA, IP), and expiration dates. Methods return domain models (`model.AuthSessionRead` / `model.AuthRead`) and use row scanning with defensive handling for inet/string conversion. Unit tests mock the `DBTX` interface to verify exact query generation, bound parameters, and row scanning without requiring a running database.

**Tech Stack:** Go 1.27+, `github.com/Masterminds/squirrel`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`, standard library (`context`, `crypto/sha256`, `errors`, `fmt`, `net/netip`, `time`).

**Spec:** [docs/go_refactor/tasks/009-repository_auth_session.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/009-repository_auth_session.md)

## Global Constraints

- Git branch: `refactor/009-repository-auth-session`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/repository`
- All SQL queries must use PostgreSQL dollar placeholder format (`$1`, `$2`) via `StmtBuilder` (`squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)`).
- No ORMs; use `pgx/v5` only via `DBTX`.
- Error handling: Wrap errors with `%w` using contextual descriptive messages.
- Formatting must adhere to `gofumpt` and linting must pass `golangci-lint run ./...`.
- `go test -race ./internal/repository/... -v` must pass.
- `go vet ./...` must report no issues.

---

## File Structure

```
backend/
├── internal/
│   ├── model/
│   │   ├── auth.go                    # [MODIFY] Add AuthSessionCreate and AuthSessionRead aliases
│   │   └── model_test.go              # [MODIFY] Verify alias compatibility with existing tests
│   └── repository/
│       ├── base.go                    # [EXISTING] DBTX interface, StmtBuilder, ScanOne, ScanRows
│       ├── archer.go                  # [EXISTING] Reference repository implementation
│       ├── auth_session.go            # [NEW] AuthSessionRepo struct, constructor, WithTx, Create, FindByTokenHash, DeleteByArcherID, DeleteExpired, RevokeByTokenHash, DeleteByTokenHash
│       └── auth_session_test.go       # [NEW] Unit tests covering all AuthSessionRepo methods with mockDBTX
```

---

### Task 1: Git Branch Setup & Model Aliases

**Files:**
- Modify: `backend/internal/model/auth.go`
- Modify: `backend/internal/model/model_test.go`

**Interfaces:**
- Consumes: `model.AuthCreate`, `model.AuthRead`
- Produces:
  - `model.AuthSessionCreate` (type alias for `AuthCreate`)
  - `model.AuthSessionRead` (type alias for `AuthRead`)

- [x] **Step 1: Check out git branch**

```bash
git switch -c refactor/009-repository-auth-session
```

- [x] **Step 2: Add type aliases to `backend/internal/model/auth.go`**

Add `AuthSessionCreate` and `AuthSessionRead` to `backend/internal/model/auth.go`:

```go
// AuthSessionCreate is an alias for AuthCreate to support session repository naming conventions.
type AuthSessionCreate = AuthCreate

// AuthSessionRead is an alias for AuthRead to support session repository naming conventions.
type AuthSessionRead = AuthRead
```

- [x] **Step 3: Add test in `backend/internal/model/model_test.go` verifying alias compatibility**

In `backend/internal/model/model_test.go`, append test function:

```go
func TestAuthSessionAliases(t *testing.T) {
	now := time.Now().UTC()
	ua := "Mozilla/5.0"
	ip := "192.168.1.1"

	create := AuthSessionCreate{
		ArcherID:         uuid.New(),
		SessionTokenHash: []byte("hash-token-123"),
		CreatedAt:        now,
		ExpiresAt:        now.Add(time.Hour),
		UA:               &ua,
		IPInet:           &ip,
	}

	var baseCreate AuthCreate = create
	if baseCreate.ArcherID != create.ArcherID {
		t.Fatalf("expected ArcherID %v, got %v", create.ArcherID, baseCreate.ArcherID)
	}

	read := AuthSessionRead{
		AuthID:           uuid.New(),
		ArcherID:         create.ArcherID,
		SessionTokenHash: create.SessionTokenHash,
		CreatedAt:        now,
		ExpiresAt:        now.Add(time.Hour),
		UA:               &ua,
		IPInet:           &ip,
	}

	var baseRead AuthRead = read
	if baseRead.AuthID != read.AuthID {
		t.Fatalf("expected AuthID %v, got %v", read.AuthID, baseRead.AuthID)
	}
}
```

- [x] **Step 4: Run model tests to verify passes**

Run: `cd backend && go test ./internal/model/... -v`
Expected: PASS

- [x] **Step 5: Commit model alias updates**

```bash
git add backend/internal/model/auth.go backend/internal/model/model_test.go
git commit -m "feat(model): add AuthSessionCreate and AuthSessionRead type aliases"
```

---

### Task 2: Auth Session Repository Unit Tests (TDD - Failing Tests)

**Files:**
- Create: `backend/internal/repository/auth_session_test.go`

**Interfaces:**
- Consumes:
  - `model.AuthSessionCreate`, `model.AuthSessionRead`
  - `repository.NewAuthSessionRepo(db DBTX)`
  - `mockDBTX` from `archer_test.go`
  - `mockSingleRow` from `base_test.go`
- Produces:
  - Unit test suite testing `Create`, `FindByTokenHash`, `DeleteByArcherID`, `DeleteExpired`, `RevokeByTokenHash`, `DeleteByTokenHash`, and `WithTx`.

- [x] **Step 1: Write failing unit tests in `backend/internal/repository/auth_session_test.go`**

```go
package repository_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/netip"
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

func sampleAuthSessionRow(authID, archerID uuid.UUID, tokenHash []byte, ip any) []any {
	now := time.Now().Truncate(time.Second).UTC()
	ua := "Mozilla/5.0 (X11; Linux x86_64)"
	return []any{
		authID,
		archerID,
		tokenHash,
		now,
		now.Add(24 * time.Hour),
		nil, // revoked_at
		&ua,
		ip,
	}
}

func TestAuthSessionRepo_Create_Success(t *testing.T) {
	archerID := uuid.New()
	rawHash := sha256.Sum256([]byte("session-secret-token"))
	tokenHash := rawHash[:]
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)
	ua := "Mozilla/5.0 (Test Browser)"
	ip := "192.168.1.100"

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("INSERT 0 1"), nil
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	err := repo.Create(context.Background(), model.AuthSessionCreate{
		ArcherID:         archerID,
		SessionTokenHash: tokenHash,
		CreatedAt:        now,
		ExpiresAt:        expiresAt,
		UA:               &ua,
		IPInet:           &ip,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(executedSQL, "INSERT INTO auth") {
		t.Errorf("expected INSERT INTO auth query, got: %s", executedSQL)
	}
	expectedCols := []string{"archer_id", "session_token_hash", "created_at", "expires_at", "ua", "ip_inet"}
	for _, col := range expectedCols {
		if !strings.Contains(executedSQL, col) {
			t.Errorf("expected query to contain column %q, got: %s", col, executedSQL)
		}
	}

	if len(executedArgs) != 6 {
		t.Fatalf("expected 6 arguments, got %d", len(executedArgs))
	}
	if executedArgs[0] != archerID {
		t.Errorf("arg[0] (archer_id): expected %v, got %v", archerID, executedArgs[0])
	}
	if string(executedArgs[1].([]byte)) != string(tokenHash) {
		t.Errorf("arg[1] (session_token_hash): expected %x, got %x", tokenHash, executedArgs[1])
	}
	if executedArgs[2] != now {
		t.Errorf("arg[2] (created_at): expected %v, got %v", now, executedArgs[2])
	}
	if executedArgs[3] != expiresAt {
		t.Errorf("arg[3] (expires_at): expected %v, got %v", expiresAt, executedArgs[3])
	}
	if executedArgs[4] != &ua {
		t.Errorf("arg[4] (ua): expected %v, got %v", &ua, executedArgs[4])
	}
	if executedArgs[5] != &ip {
		t.Errorf("arg[5] (ip_inet): expected %v, got %v", &ip, executedArgs[5])
	}
}

func TestAuthSessionRepo_Create_DefaultCreatedAt(t *testing.T) {
	archerID := uuid.New()
	rawHash := sha256.Sum256([]byte("token-without-created-at"))
	tokenHash := rawHash[:]

	var executedArgs []any
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedArgs = args
			return pgconn.NewCommandTag("INSERT 0 1"), nil
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	before := time.Now().UTC().Add(-time.Second)
	err := repo.Create(context.Background(), model.AuthSessionCreate{
		ArcherID:         archerID,
		SessionTokenHash: tokenHash,
		ExpiresAt:        time.Now().UTC().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	after := time.Now().UTC().Add(time.Second)
	createdAt, ok := executedArgs[2].(time.Time)
	if !ok {
		t.Fatalf("arg[2] (created_at) should be time.Time, got %T", executedArgs[2])
	}
	if createdAt.Before(before) || createdAt.After(after) {
		t.Errorf("created_at %v not between %v and %v", createdAt, before, after)
	}
}

func TestAuthSessionRepo_Create_DBError(t *testing.T) {
	dbErr := errors.New("db insert failed")
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, dbErr
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	err := repo.Create(context.Background(), model.AuthSessionCreate{
		ArcherID:         uuid.New(),
		SessionTokenHash: []byte("hash"),
		ExpiresAt:        time.Now().UTC().Add(time.Hour),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
}

func TestAuthSessionRepo_FindByTokenHash_Success(t *testing.T) {
	authID := uuid.New()
	archerID := uuid.New()
	rawHash := sha256.Sum256([]byte("test-find-hash"))
	tokenHash := rawHash[:]
	ipStr := "10.0.0.1"

	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			executedSQL = sql
			executedArgs = args
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleAuthSessionRow(authID, archerID, tokenHash, &ipStr)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	session, err := repo.FindByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}

	if !strings.HasPrefix(executedSQL, "SELECT auth_id, archer_id, session_token_hash, created_at, expires_at, revoked_at, ua, ip_inet FROM auth") {
		t.Errorf("unexpected query SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE session_token_hash = $1") {
		t.Errorf("expected query to contain WHERE session_token_hash = $1, got: %s", executedSQL)
	}
	if len(executedArgs) != 1 || string(executedArgs[0].([]byte)) != string(tokenHash) {
		t.Errorf("expected argument tokenHash, got %v", executedArgs)
	}

	if session.AuthID != authID {
		t.Errorf("expected AuthID %v, got %v", authID, session.AuthID)
	}
	if session.ArcherID != archerID {
		t.Errorf("expected ArcherID %v, got %v", archerID, session.ArcherID)
	}
	if string(session.SessionTokenHash) != string(tokenHash) {
		t.Errorf("expected SessionTokenHash %x, got %x", tokenHash, session.SessionTokenHash)
	}
	if session.IPInet == nil || *session.IPInet != ipStr {
		t.Errorf("expected IPInet %q, got %v", ipStr, session.IPInet)
	}
}

func TestAuthSessionRepo_FindByTokenHash_NetipPrefix(t *testing.T) {
	authID := uuid.New()
	archerID := uuid.New()
	tokenHash := []byte("prefix-token-hash")
	prefix := netip.MustParsePrefix("172.16.0.1/32")

	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					row := sampleAuthSessionRow(authID, archerID, tokenHash, prefix)
					mr := &mockMultiRows{records: [][]any{row}}
					mr.Next()
					return mr.Scan(dest...)
				},
			}
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	session, err := repo.FindByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.IPInet == nil || *session.IPInet != "172.16.0.1" {
		t.Errorf("expected IPInet '172.16.0.1', got %v", session.IPInet)
	}
}

func TestAuthSessionRepo_FindByTokenHash_NotFound(t *testing.T) {
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	session, err := repo.FindByTokenHash(context.Background(), []byte("unknown-hash"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session != nil {
		t.Fatalf("expected nil session for not found, got %+v", session)
	}
}

func TestAuthSessionRepo_FindByTokenHash_DBError(t *testing.T) {
	dbErr := errors.New("query failure")
	mock := &mockDBTX{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockSingleRow{
				scanFn: func(dest ...any) error {
					return dbErr
				},
			}
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	session, err := repo.FindByTokenHash(context.Background(), []byte("some-hash"))
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

func TestAuthSessionRepo_DeleteByArcherID_Success(t *testing.T) {
	archerID := uuid.New()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("DELETE 3"), nil
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	err := repo.DeleteByArcherID(context.Background(), archerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(executedSQL, "DELETE FROM auth") {
		t.Errorf("expected DELETE FROM auth query, got: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE archer_id = $1") {
		t.Errorf("expected WHERE archer_id = $1, got: %s", executedSQL)
	}
	if len(executedArgs) != 1 || executedArgs[0] != archerID {
		t.Errorf("expected archerID argument, got %v", executedArgs)
	}
}

func TestAuthSessionRepo_DeleteByArcherID_Error(t *testing.T) {
	dbErr := errors.New("delete failed")
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, dbErr
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	err := repo.DeleteByArcherID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
}

func TestAuthSessionRepo_DeleteExpired_Success(t *testing.T) {
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("DELETE 5"), nil
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	before := time.Now().UTC().Add(-time.Second)
	count, err := repo.DeleteExpired(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := time.Now().UTC().Add(time.Second)

	if count != 5 {
		t.Errorf("expected deleted count 5, got %d", count)
	}
	if !strings.HasPrefix(executedSQL, "DELETE FROM auth") {
		t.Errorf("expected DELETE FROM auth query, got: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE expires_at < $1") {
		t.Errorf("expected WHERE expires_at < $1, got: %s", executedSQL)
	}

	if len(executedArgs) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(executedArgs))
	}
	ts, ok := executedArgs[0].(time.Time)
	if !ok {
		t.Fatalf("expected time.Time argument, got %T", executedArgs[0])
	}
	if ts.Before(before) || ts.After(after) {
		t.Errorf("expires_at param %v not between %v and %v", ts, before, after)
	}
}

func TestAuthSessionRepo_DeleteExpired_Error(t *testing.T) {
	dbErr := errors.New("cleanup delete failure")
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, dbErr
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	count, err := repo.DeleteExpired(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr wrapped, got: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0 on error, got %d", count)
	}
}

func TestAuthSessionRepo_RevokeByTokenHash_Success(t *testing.T) {
	tokenHash := []byte("revoke-hash")
	revokedAt := time.Now().UTC()
	var executedSQL string
	var executedArgs []any

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			executedArgs = args
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	err := repo.RevokeByTokenHash(context.Background(), tokenHash, revokedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(executedSQL, "UPDATE auth SET revoked_at = $1") {
		t.Errorf("unexpected query SQL: %s", executedSQL)
	}
	if !strings.Contains(executedSQL, "WHERE session_token_hash = $2 AND revoked_at IS NULL") {
		t.Errorf("expected query to contain WHERE session_token_hash = $2 AND revoked_at IS NULL, got: %s", executedSQL)
	}
	if len(executedArgs) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(executedArgs))
	}
}

func TestAuthSessionRepo_RevokeByTokenHash_NotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	err := repo.RevokeByTokenHash(context.Background(), []byte("not-found-hash"), time.Now().UTC())
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected apperror.ErrNotFound, got: %v", err)
	}
}

func TestAuthSessionRepo_DeleteByTokenHash_Success(t *testing.T) {
	tokenHash := []byte("delete-token-hash")
	var executedSQL string

	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedSQL = sql
			return pgconn.NewCommandTag("DELETE 1"), nil
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	err := repo.DeleteByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(executedSQL, "DELETE FROM auth WHERE session_token_hash = $1") {
		t.Errorf("unexpected query SQL: %s", executedSQL)
	}
}

func TestAuthSessionRepo_DeleteByTokenHash_NotFound(t *testing.T) {
	mock := &mockDBTX{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("DELETE 0"), nil
		},
	}

	repo := repository.NewAuthSessionRepo(mock)
	err := repo.DeleteByTokenHash(context.Background(), []byte("absent-hash"))
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected apperror.ErrNotFound, got: %v", err)
	}
}

func TestAuthSessionRepo_WithTx(t *testing.T) {
	mock := &mockDBTX{}
	repo := repository.NewAuthSessionRepo(mock)

	tx := &mockTx{}
	repoWithTx := repo.WithTx(tx)
	if repoWithTx == nil {
		t.Fatal("expected non-nil repoWithTx")
	}
	if repoWithTx == repo {
		t.Errorf("expected different repo instance from WithTx")
	}
}
```

- [x] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/repository/... -v`
Expected: FAIL (compilation error: undefined `repository.NewAuthSessionRepo`)

---

### Task 3: Auth Session Repository Implementation (`auth_session.go`)

**Files:**
- Create: `backend/internal/repository/auth_session.go`

**Interfaces:**
- Consumes:
  - `DBTX`, `StmtBuilder`, `ScanOne` from `base.go`
  - `model.AuthSessionCreate`, `model.AuthSessionRead` from `model`
  - `apperror.ErrNotFound` from `apperror`
- Produces:
  - `AuthSessionRepo` struct
  - `NewAuthSessionRepo(db DBTX) *AuthSessionRepo`
  - `(r *AuthSessionRepo) WithTx(tx pgx.Tx) *AuthSessionRepo`
  - `(r *AuthSessionRepo) Create(ctx context.Context, data model.AuthSessionCreate) error`
  - `(r *AuthSessionRepo) FindByTokenHash(ctx context.Context, hash []byte) (*model.AuthSessionRead, error)`
  - `(r *AuthSessionRepo) DeleteByArcherID(ctx context.Context, archerID uuid.UUID) error`
  - `(r *AuthSessionRepo) DeleteExpired(ctx context.Context) (int64, error)`
  - `(r *AuthSessionRepo) RevokeByTokenHash(ctx context.Context, hash []byte, revokedAt time.Time) error`
  - `(r *AuthSessionRepo) DeleteByTokenHash(ctx context.Context, hash []byte) error`

- [x] **Step 1: Write `backend/internal/repository/auth_session.go`**

```go
package repository

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var authSessionColumns = []string{
	"auth_id",
	"archer_id",
	"session_token_hash",
	"created_at",
	"expires_at",
	"revoked_at",
	"ua",
	"ip_inet",
}

// AuthSessionRepo manages database operations for authentication sessions.
type AuthSessionRepo struct {
	db DBTX
}

// NewAuthSessionRepo constructs an AuthSessionRepo backed by DBTX.
func NewAuthSessionRepo(db DBTX) *AuthSessionRepo {
	return &AuthSessionRepo{db: db}
}

// WithTx returns a new AuthSessionRepo bound to the given transaction.
func (r *AuthSessionRepo) WithTx(tx pgx.Tx) *AuthSessionRepo {
	return &AuthSessionRepo{db: tx}
}

func scanAuthSession(scanner interface{ Scan(dest ...any) error }) (model.AuthSessionRead, error) {
	var (
		s     model.AuthSessionRead
		ipRaw any
	)

	err := scanner.Scan(
		&s.AuthID,
		&s.ArcherID,
		&s.SessionTokenHash,
		&s.CreatedAt,
		&s.ExpiresAt,
		&s.RevokedAt,
		&s.UA,
		&ipRaw,
	)
	if err != nil {
		return model.AuthSessionRead{}, err
	}

	if ipRaw != nil {
		switch v := ipRaw.(type) {
		case string:
			s.IPInet = &v
		case *string:
			s.IPInet = v
		case netip.Prefix:
			str := v.Addr().String()
			s.IPInet = &str
		case *netip.Prefix:
			if v != nil {
				str := v.Addr().String()
				s.IPInet = &str
			}
		case netip.Addr:
			str := v.String()
			s.IPInet = &str
		case *netip.Addr:
			if v != nil {
				str := v.String()
				s.IPInet = &str
			}
		case fmt.Stringer:
			str := v.String()
			s.IPInet = &str
		default:
			str := fmt.Sprintf("%v", ipRaw)
			s.IPInet = &str
		}
	}

	return s, nil
}

// Create inserts a new auth session record.
//
//nolint:gocritic // hugeParam: data value parameter matches repository interface specification
func (r *AuthSessionRepo) Create(ctx context.Context, data model.AuthSessionCreate) error {
	createdAt := data.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	sql, args, err := StmtBuilder.Insert("auth").
		Columns(
			"archer_id",
			"session_token_hash",
			"created_at",
			"expires_at",
			"ua",
			"ip_inet",
		).
		Values(
			data.ArcherID,
			data.SessionTokenHash,
			createdAt,
			data.ExpiresAt,
			data.UA,
			data.IPInet,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("building create auth session query: %w", err)
	}

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("inserting auth session: %w", err)
	}

	return nil
}

// FindByTokenHash retrieves an auth session by its token hash.
// Returns nil, nil if no session exists for the given hash.
func (r *AuthSessionRepo) FindByTokenHash(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
	sql, args, err := StmtBuilder.Select(authSessionColumns...).
		From("auth").
		Where(squirrel.Eq{"session_token_hash": hash}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building find by token hash query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return ScanOne(row, func(r pgx.Row) (model.AuthSessionRead, error) {
		return scanAuthSession(r)
	})
}

// DeleteByArcherID removes all auth sessions belonging to the specified archer (e.g. logout all sessions).
// Operation is idempotent and returns nil if no sessions exist.
func (r *AuthSessionRepo) DeleteByArcherID(ctx context.Context, archerID uuid.UUID) error {
	sql, args, err := StmtBuilder.Delete("auth").
		Where(squirrel.Eq{"archer_id": archerID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete by archer id query: %w", err)
	}

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("deleting auth sessions by archer id: %w", err)
	}

	return nil
}

// DeleteExpired removes all sessions whose expires_at is prior to the current time.
// Returns the count of deleted session records.
func (r *AuthSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	sql, args, err := StmtBuilder.Delete("auth").
		Where(squirrel.Lt{"expires_at": time.Now().UTC()}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("building delete expired query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("executing delete expired auth sessions: %w", err)
	}

	return tag.RowsAffected(), nil
}

// RevokeByTokenHash marks an active session as revoked. Returns apperror.ErrNotFound
// if no matching unrevoked session exists.
func (r *AuthSessionRepo) RevokeByTokenHash(ctx context.Context, hash []byte, revokedAt time.Time) error {
	sql, args, err := StmtBuilder.Update("auth").
		Set("revoked_at", revokedAt).
		Where(squirrel.Eq{"session_token_hash": hash}).
		Where(squirrel.Eq{"revoked_at": nil}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building revoke by token hash query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing revoke by token hash: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

// DeleteByTokenHash removes an auth session identified by its token hash.
// If no session was deleted, returns apperror.ErrNotFound.
func (r *AuthSessionRepo) DeleteByTokenHash(ctx context.Context, hash []byte) error {
	sql, args, err := StmtBuilder.Delete("auth").
		Where(squirrel.Eq{"session_token_hash": hash}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete by token hash query: %w", err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("executing delete by token hash: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
```

- [x] **Step 2: Run repository unit tests to verify they pass**

Run: `cd backend && go test -race ./internal/repository/... -v`
Expected: PASS (all tests pass)

- [x] **Step 3: Commit `auth_session.go` and `auth_session_test.go`**

```bash
git add backend/internal/repository/auth_session.go backend/internal/repository/auth_session_test.go
git commit -m "feat(repository): implement AuthSessionRepo with squirrel query building"
```

---

### Task 4: Formatting, Linting & Full Verification

**Files:**
- Modify: `backend/internal/repository/auth_session.go` (if formatting needed)
- Modify: `backend/internal/repository/auth_session_test.go` (if formatting needed)

**Interfaces:**
- Consumes: `gofumpt`, `golangci-lint`, `go vet`, `go test`
- Produces: Clean, passing build and linter report

- [x] **Step 1: Format files with gofumpt**

Run:
```bash
cd backend
gofumpt -l -w internal/model/auth.go internal/model/model_test.go internal/repository/auth_session.go internal/repository/auth_session_test.go
```

- [x] **Step 2: Run golangci-lint**

Run:
```bash
cd backend
golangci-lint run ./...
```
Expected: PASS (no lint issues reported)

- [x] **Step 3: Run go vet**

Run:
```bash
cd backend
go vet ./...
```
Expected: PASS (clean)

- [x] **Step 4: Run full test suite with race detector**

Run:
```bash
cd backend
go test -race ./... -v
```
Expected: PASS (all packages pass)

- [x] **Step 5: Verify build compiles cleanly**

Run:
```bash
cd backend
go build ./...
```
Expected: PASS (clean binary compile)

- [x] **Step 6: Commit any formatting or lint fixes**

```bash
git add -u
git commit -m "chore: format and lint auth session repository" || echo "No formatting changes needed"
```

---

### Task 5: Documentation & Acceptance Criteria Completion

**Files:**
- Modify: `docs/go_refactor/tasks/009-repository_auth_session.md`

**Interfaces:**
- Consumes: Completed implementation and test results
- Produces: Updated checklist in task specification

- [x] **Step 1: Update acceptance criteria and steps in task 009 doc**

Update checkboxes in `docs/go_refactor/tasks/009-repository_auth_session.md`:
```markdown
## Acceptance Criteria

- [x] `backend/internal/repository/auth_session.go` implements `AuthSessionRepo` with methods:
    - `Create(ctx, data model.AuthSessionCreate) error`
    - `FindByTokenHash(ctx, hash []byte) (*model.AuthSessionRead, error)`
    - `DeleteByArcherID(ctx, archerID uuid.UUID) error` (logout all sessions)
    - `DeleteExpired(ctx) (int64, error)` (cleanup expired sessions)
- [x] All queries use squirrel.
- [x] Unit tests verify query building and scan logic.
- [x] `go test ./internal/repository/...` passes.
- [x] `go vet ./...` reports no issues.
```

- [x] **Step 2: Commit documentation update**

```bash
git add docs/go_refactor/tasks/009-repository_auth_session.md
git commit -m "docs: mark task 009 acceptance criteria as completed"
```
