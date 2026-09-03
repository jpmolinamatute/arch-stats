package repository_test

import (
	"bytes"
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
	if !bytes.Equal(executedArgs[1].([]byte), tokenHash) {
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
	if len(executedArgs) != 1 || !bytes.Equal(executedArgs[0].([]byte), tokenHash) {
		t.Errorf("expected argument tokenHash, got %v", executedArgs)
	}

	if session.AuthID != authID {
		t.Errorf("expected AuthID %v, got %v", authID, session.AuthID)
	}
	if session.ArcherID != archerID {
		t.Errorf("expected ArcherID %v, got %v", archerID, session.ArcherID)
	}
	if !bytes.Equal(session.SessionTokenHash, tokenHash) {
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
	if len(executedArgs) != 1 || executedArgs[0] != archerID.String() {
		t.Errorf("expected archerID argument %s, got %v", archerID.String(), executedArgs[0])
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
