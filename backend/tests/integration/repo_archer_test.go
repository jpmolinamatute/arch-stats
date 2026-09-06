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
	sub := "google-sub-" + unique

	id, err := repo.Create(ctx, model.ArcherCreate{
		FirstName:     "Oliver",
		LastName:      "Queen",
		Email:         email,
		DateOfBirth:   "1985-05-16",
		Gender:        model.GenderMale,
		Bowstyle:      model.BowstyleRecurve,
		DrawWeight:    45.0,
		GoogleSubject: sub,
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
	if archer.Bowstyle != model.BowstyleRecurve {
		t.Errorf("Bowstyle = %v, want %v", archer.Bowstyle, model.BowstyleRecurve)
	}
	if archer.DrawWeight != 45.0 {
		t.Errorf("DrawWeight = %v, want 45.0", archer.DrawWeight)
	}
	if archer.GoogleSubject != sub {
		t.Errorf("GoogleSubject = %q, want %q", archer.GoogleSubject, sub)
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

	created, err := createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.GoogleSubject = sub
	})
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
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

	_, err := repo.Create(ctx, model.ArcherCreate{
		FirstName:     "First",
		LastName:      "Archer",
		Email:         email,
		DateOfBirth:   "1992-01-01",
		Gender:        model.GenderFemale,
		Bowstyle:      model.BowstyleBarebow,
		DrawWeight:    30.0,
		GoogleSubject: "sub-1",
	})
	if err != nil {
		t.Fatalf("initial Create() failed: %v", err)
	}

	// Attempt duplicate email
	_, err = repo.Create(ctx, model.ArcherCreate{
		FirstName:     "Second",
		LastName:      "Archer",
		Email:         email,
		DateOfBirth:   "1995-02-02",
		Gender:        model.GenderMale,
		Bowstyle:      model.BowstyleCompound,
		DrawWeight:    50.0,
		GoogleSubject: "sub-2",
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
	newBowstyle := model.BowstyleBarebow
	newDrawWeight := 36.0

	err = repo.Update(ctx, model.ArcherSet{
		FirstName:  &newName,
		Bowstyle:   &newBowstyle,
		DrawWeight: &newDrawWeight,
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
	if updated.Bowstyle != newBowstyle {
		t.Errorf("Bowstyle = %v, want %v", updated.Bowstyle, newBowstyle)
	}
	if updated.DrawWeight != newDrawWeight {
		t.Errorf("DrawWeight = %v, want %v", updated.DrawWeight, newDrawWeight)
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

	// Create 2 recurve and 1 compound archer
	_, err := createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Bowstyle = model.BowstyleRecurve
		a.Gender = model.GenderFemale
	})
	if err != nil {
		t.Fatalf("create archer 1 failed: %v", err)
	}

	_, err = createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Bowstyle = model.BowstyleRecurve
		a.Gender = model.GenderMale
	})
	if err != nil {
		t.Fatalf("create archer 2 failed: %v", err)
	}

	compoundArcher, err := createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Bowstyle = model.BowstyleCompound
		a.Gender = model.GenderMale
	})
	if err != nil {
		t.Fatalf("create archer 3 failed: %v", err)
	}

	// Filter by BowstyleCompound
	targetBowstyle := model.BowstyleCompound
	results, err := repo.FindAll(ctx, model.ArcherFilter{
		Bowstyle: &targetBowstyle,
	})
	if err != nil {
		t.Fatalf("FindAll(BowstyleCompound) failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 compound archer, got %d", len(results))
	}
	if results[0].ArcherID != compoundArcher.ArcherID {
		t.Errorf("ArcherID = %v, want %v", results[0].ArcherID, compoundArcher.ArcherID)
	}

	// Filter by BowstyleRecurve
	recurveBowstyle := model.BowstyleRecurve
	recurveResults, err := repo.FindAll(ctx, model.ArcherFilter{
		Bowstyle: &recurveBowstyle,
	})
	if err != nil {
		t.Fatalf("FindAll(BowstyleRecurve) failed: %v", err)
	}
	if len(recurveResults) != 2 {
		t.Fatalf("expected 2 recurve archers, got %d", len(recurveResults))
	}
}
