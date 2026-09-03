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

func (m *mockMultiRows) Close()                                       { m.closed = true }
func (m *mockMultiRows) Err() error                                   { return m.err }
func (m *mockMultiRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *mockMultiRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockMultiRows) Next() bool {
	m.idx++
	return m.idx <= len(m.records)
}

func (m *mockMultiRows) Scan(dest ...any) error {
	row := m.records[m.idx-1]
	for i, v := range row {
		switch d := dest[i].(type) {
		case *uuid.UUID:
			switch val := v.(type) {
			case uuid.UUID:
				*d = val
			case *uuid.UUID:
				if val != nil {
					*d = *val
				}
			}
		case **uuid.UUID:
			switch val := v.(type) {
			case nil:
				*d = nil
			case *uuid.UUID:
				*d = val
			case uuid.UUID:
				*d = &val
			}
		case *string:
			switch val := v.(type) {
			case string:
				*d = val
			case *string:
				if val != nil {
					*d = *val
				}
			}
		case **string:
			switch val := v.(type) {
			case nil:
				*d = nil
			case *string:
				*d = val
			case string:
				*d = &val
			}
		case *model.Gender:
			*d = v.(model.Gender)
		case *model.Bowstyle:
			*d = v.(model.Bowstyle)
		case *float64:
			*d = v.(float64)
		case *time.Time:
			*d = v.(time.Time)
		case **time.Time:
			switch val := v.(type) {
			case nil:
				*d = nil
			case *time.Time:
				*d = val
			case time.Time:
				*d = &val
			}
		case *[]byte:
			switch val := v.(type) {
			case []byte:
				*d = val
			case *[]byte:
				if val != nil {
					*d = *val
				}
			}
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
	if mock.lastArgs[0] != archerID.String() {
		t.Errorf("expected query arg %s, got %v", archerID.String(), mock.lastArgs[0])
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
