package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestSessionRepo_CreateAndFindByID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	repo := repository.NewSessionRepo(testPool)
	id, err := repo.Create(ctx, model.SessionCreate{
		OwnerArcherID:   archer.ArcherID,
		SessionLocation: "Sherwood Range",
		IsIndoor:        false,
		IsOpened:        true,
	})
	if err != nil {
		t.Fatalf("Create() session failed: %v", err)
	}

	session, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if session == nil {
		t.Fatalf("expected session, got nil")
	}
	if session.SessionID != id {
		t.Errorf("SessionID = %v, want %v", session.SessionID, id)
	}
	if session.OwnerArcherID != archer.ArcherID {
		t.Errorf("OwnerArcherID = %v, want %v", session.OwnerArcherID, archer.ArcherID)
	}
	if session.SessionLocation != "Sherwood Range" {
		t.Errorf("SessionLocation = %q, want 'Sherwood Range'", session.SessionLocation)
	}
	if !session.IsOpened {
		t.Errorf("IsOpened = false, want true")
	}
	if session.IsIndoor {
		t.Errorf("IsIndoor = true, want false")
	}
	if session.ClosedAt != nil {
		t.Errorf("ClosedAt = %v, want nil", session.ClosedAt)
	}

	// Missing ID returns nil, nil
	missing, err := repo.FindByID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("FindByID(missing) error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing session ID, got %+v", missing)
	}
}

func TestSessionRepo_FindOpen(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	repo := repository.NewSessionRepo(testPool)

	// Archer has no open session initially
	openSess, err := repo.FindOpen(ctx, archer.ArcherID)
	if err != nil {
		t.Fatalf("FindOpen() error: %v", err)
	}
	if openSess != nil {
		t.Fatalf("expected no open session, got %+v", openSess)
	}

	// Create open session
	created, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	openSess, err = repo.FindOpen(ctx, archer.ArcherID)
	if err != nil {
		t.Fatalf("FindOpen() after create failed: %v", err)
	}
	if openSess == nil || openSess.SessionID != created.SessionID {
		t.Fatalf("FindOpen() = %+v, want session %v", openSess, created.SessionID)
	}
}

func TestSessionRepo_Close(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	repo := repository.NewSessionRepo(testPool)
	session, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond) // ensure closed_at > created_at

	if err := repo.Close(ctx, session.SessionID); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	closed, err := repo.FindByID(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if closed.IsOpened {
		t.Errorf("expected is_opened = false, got true")
	}
	if closed.ClosedAt == nil {
		t.Fatalf("expected closed_at to be set, got nil")
	}

	// Closing already closed session returns ErrNotFound
	if err := repo.Close(ctx, session.SessionID); err != apperror.ErrNotFound {
		t.Errorf("expected ErrNotFound closing already closed session, got %v", err)
	}
}

func TestSessionRepo_FindAll_WithFilters(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer1, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create archer 1 failed: %v", err)
	}
	archer2, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create archer 2 failed: %v", err)
	}

	repo := repository.NewSessionRepo(testPool)

	// Create 1 indoor session for archer 1
	s1, err := repo.Create(ctx, model.SessionCreate{
		OwnerArcherID:   archer1.ArcherID,
		SessionLocation: "Indoor Range A",
		IsIndoor:        true,
		IsOpened:        true,
	})
	if err != nil {
		t.Fatalf("create session 1 failed: %v", err)
	}

	// Create 1 outdoor session for archer 2
	s2, err := repo.Create(ctx, model.SessionCreate{
		OwnerArcherID:   archer2.ArcherID,
		SessionLocation: "Outdoor Field B",
		IsIndoor:        false,
		IsOpened:        true,
	})
	if err != nil {
		t.Fatalf("create session 2 failed: %v", err)
	}

	// FindAll without filter returns both
	all, err := repo.FindAll(ctx, model.SessionFilter{})
	if err != nil {
		t.Fatalf("FindAll() failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(all))
	}

	// Filter by OwnerArcherID = archer1
	archer1Filter, err := repo.FindAll(ctx, model.SessionFilter{
		OwnerArcherID: &archer1.ArcherID,
	})
	if err != nil {
		t.Fatalf("FindAll(archer1) failed: %v", err)
	}
	if len(archer1Filter) != 1 || archer1Filter[0].SessionID != s1 {
		t.Fatalf("expected session %v for archer 1, got %+v", s1, archer1Filter)
	}

	// Filter by IsIndoor = false
	isIndoorFalse := false
	outdoorFilter, err := repo.FindAll(ctx, model.SessionFilter{
		IsIndoor: &isIndoorFalse,
	})
	if err != nil {
		t.Fatalf("FindAll(outdoor) failed: %v", err)
	}
	if len(outdoorFilter) != 1 || outdoorFilter[0].SessionID != s2 {
		t.Fatalf("expected session %v for outdoor, got %+v", s2, outdoorFilter)
	}
}

func TestSessionRepo_FindParticipating(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}

	repo := repository.NewSessionRepo(testPool)

	// Archer is not yet in a slot
	partSessionID, err := repo.FindParticipating(ctx, archer.ArcherID)
	if err != nil {
		t.Fatalf("FindParticipating() error: %v", err)
	}
	if partSessionID != nil {
		t.Fatalf("expected nil participating session, got %v", *partSessionID)
	}

	// Assign archer to slot in open session
	_, err = createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	partSessionID, err = repo.FindParticipating(ctx, archer.ArcherID)
	if err != nil {
		t.Fatalf("FindParticipating() after slot creation failed: %v", err)
	}
	if partSessionID == nil || *partSessionID != session.SessionID {
		t.Fatalf("FindParticipating() = %v, want %v", partSessionID, session.SessionID)
	}
}
