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
					*dest[0].(*uuid.UUID) = expectedID
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
					*dest[0].(*uuid.UUID) = expectedID
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
	if executedArgs[2] != archerID.String() {
		t.Errorf("expected target archerID %s, got %v", archerID.String(), executedArgs[2])
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
