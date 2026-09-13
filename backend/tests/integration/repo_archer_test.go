package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestArcherRepo_CreateAndFindByID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	repo := repository.NewArcherRepo(testPool)
	unique := uuid.New().String()[:8]
	email := "archer-" + unique + "@example.com"

	var authID uuid.UUID
	err := testPool.QueryRow(ctx, "INSERT INTO auth (google_subject) VALUES ($1) RETURNING archer_id", "google-sub-"+unique).Scan(&authID)
	if err != nil {
		t.Fatalf("inserting auth failed: %v", err)
	}

	id, err := repo.Create(ctx, model.ArcherCreate{
		ArcherID:    &authID,
		FirstName:   "Oliver",
		LastName:    "Queen",
		Email:       email,
		DateOfBirth: "1985-05-16",
		Gender:      model.GenderMale,
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	archer, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if archer == nil {
		t.Fatalf("expected archer, got nil")
	}
	if archer.ArcherID != id {
		t.Errorf("ArcherID = %v, want %v", archer.ArcherID, id)
	}
	if archer.FirstName != "Oliver" || archer.LastName != "Queen" {
		t.Errorf("Name = %s %s, want Oliver Queen", archer.FirstName, archer.LastName)
	}
	if archer.Email != email {
		t.Errorf("Email = %q, want %q", archer.Email, email)
	}
	if archer.DateOfBirth != "1985-05-16" {
		t.Errorf("DateOfBirth = %q, want 1985-05-16", archer.DateOfBirth)
	}
	if archer.Gender != model.GenderMale {
		t.Errorf("Gender = %v, want %v", archer.Gender, model.GenderMale)
	}
}

func TestArcherRepo_FindByEmail(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	repo := repository.NewArcherRepo(testPool)
	unique := uuid.New().String()[:8]
	email := "findbyemail-" + unique + "@example.com"

	created, err := createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Email = email
	})
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	found, err := repo.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("FindByEmail() failed: %v", err)
	}
	if found == nil || found.ArcherID != created.ArcherID {
		t.Fatalf("FindByEmail() returned %+v, want archer %v", found, created.ArcherID)
	}

	// Non-existent email returns nil, nil
	missing, err := repo.FindByEmail(ctx, "nonexistent@example.com")
	if err != nil {
		t.Fatalf("FindByEmail(missing) error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing email, got %+v", missing)
	}
}

func TestArcherRepo_FindByGoogleSubject(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	repo := repository.NewArcherRepo(testPool)
	unique := uuid.New().String()[:8]
	sub := "google-sub-" + unique

	created, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}
	_, err = testPool.Exec(ctx, "UPDATE auth SET google_subject = $2 WHERE archer_id = $1", created.ArcherID, sub)
	if err != nil {
		t.Fatalf("updating auth google_subject failed: %v", err)
	}

	found, err := repo.FindByGoogleSubject(ctx, sub)
	if err != nil {
		t.Fatalf("FindByGoogleSubject() failed: %v", err)
	}
	if found == nil || found.ArcherID != created.ArcherID {
		t.Fatalf("FindByGoogleSubject() returned %+v, want archer %v", found, created.ArcherID)
	}

	// Non-existent subject returns nil, nil
	missing, err := repo.FindByGoogleSubject(ctx, "nonexistent-sub")
	if err != nil {
		t.Fatalf("FindByGoogleSubject(missing) error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing subject, got %+v", missing)
	}
}

func TestArcherRepo_DuplicateEmail_Conflict(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	repo := repository.NewArcherRepo(testPool)
	email := "duplicate@example.com"

	var authID1, authID2 uuid.UUID
	err := testPool.QueryRow(ctx, "INSERT INTO auth (google_subject) VALUES ($1) RETURNING archer_id", "google-sub-dup-1").Scan(&authID1)
	if err != nil {
		t.Fatalf("inserting auth 1 failed: %v", err)
	}
	err = testPool.QueryRow(ctx, "INSERT INTO auth (google_subject) VALUES ($1) RETURNING archer_id", "google-sub-dup-2").Scan(&authID2)
	if err != nil {
		t.Fatalf("inserting auth 2 failed: %v", err)
	}

	_, err = repo.Create(ctx, model.ArcherCreate{
		ArcherID:    &authID1,
		FirstName:   "First",
		LastName:    "Archer",
		Email:       email,
		DateOfBirth: "1992-01-01",
		Gender:      model.GenderFemale,
	})
	if err != nil {
		t.Fatalf("initial Create() failed: %v", err)
	}

	// Attempt duplicate email
	_, err = repo.Create(ctx, model.ArcherCreate{
		ArcherID:    &authID2,
		FirstName:   "Second",
		LastName:    "Archer",
		Email:       email,
		DateOfBirth: "1995-02-02",
		Gender:      model.GenderMale,
	})
	if err == nil {
		t.Fatalf("expected unique constraint error on duplicate email, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate key") && !strings.Contains(err.Error(), "unique") {
		t.Errorf("expected duplicate key/unique error, got: %v", err)
	}
}

func TestArcherRepo_Update(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	repo := repository.NewArcherRepo(testPool)
	created, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	newName := "UpdatedName"

	err = repo.Update(ctx, model.ArcherSet{
		FirstName: &newName,
	}, model.ArcherFilter{
		ArcherID: &created.ArcherID,
	})
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	updated, err := repo.FindByID(ctx, created.ArcherID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if updated.FirstName != newName {
		t.Errorf("FirstName = %q, want %q", updated.FirstName, newName)
	}
	if updated.LastName != created.LastName {
		t.Errorf("LastName changed: got %q, want %q", updated.LastName, created.LastName)
	}
}

func TestArcherRepo_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	repo := repository.NewArcherRepo(testPool)
	nonExistentID := uuid.New()

	archer, err := repo.FindByID(ctx, nonExistentID)
	if err != nil {
		t.Fatalf("FindByID() expected nil error for absent entity, got: %v", err)
	}
	if archer != nil {
		t.Fatalf("expected nil archer for non-existent ID, got: %+v", archer)
	}
}

func TestArcherRepo_FindAll_WithFilters(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	repo := repository.NewArcherRepo(testPool)

	// Create 2 male and 1 female archer
	femaleArcher, err := createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Gender = model.GenderFemale
	})
	if err != nil {
		t.Fatalf("create archer 1 failed: %v", err)
	}

	_, err = createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Gender = model.GenderMale
	})
	if err != nil {
		t.Fatalf("create archer 2 failed: %v", err)
	}

	_, err = createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Gender = model.GenderMale
	})
	if err != nil {
		t.Fatalf("create archer 3 failed: %v", err)
	}

	// Filter by GenderFemale
	targetGender := model.GenderFemale
	results, err := repo.FindAll(ctx, model.ArcherFilter{
		Gender: &targetGender,
	})
	if err != nil {
		t.Fatalf("FindAll(GenderFemale) failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 female archer, got %d", len(results))
	}
	if results[0].ArcherID != femaleArcher.ArcherID {
		t.Errorf("ArcherID = %v, want %v", results[0].ArcherID, femaleArcher.ArcherID)
	}

	// Filter by GenderMale
	maleGender := model.GenderMale
	maleResults, err := repo.FindAll(ctx, model.ArcherFilter{
		Gender: &maleGender,
	})
	if err != nil {
		t.Fatalf("FindAll(GenderMale) failed: %v", err)
	}
	if len(maleResults) != 2 {
		t.Fatalf("expected 2 male archers, got %d", len(maleResults))
	}
}
