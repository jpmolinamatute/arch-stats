package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestSlotRepo_CreateAndFindBySessionID(t *testing.T) {
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

	repo := repository.NewSlotRepo(testPool)
	shotPerRound := 6
	id, err := repo.Create(ctx, model.SlotCreate{
		TargetID:        target.TargetID,
		ArcherID:        archer.ArcherID,
		SessionID:       session.SessionID,
		SlotLetter:      model.SlotLetterA,
		FaceType:        model.FaceTypeWA40Full,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      42.0,
		IsShooting:      true,
		ShotPerRound:    &shotPerRound,
		IntervalSeconds: 30,
	})
	if err != nil {
		t.Fatalf("Create() slot failed: %v", err)
	}

	slot, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if slot == nil {
		t.Fatalf("expected slot, got nil")
	}
	if slot.SlotID != id {
		t.Errorf("SlotID = %v, want %v", slot.SlotID, id)
	}
	if slot.SlotLetter != model.SlotLetterA {
		t.Errorf("SlotLetter = %v, want A", slot.SlotLetter)
	}
	if slot.FaceType != model.FaceTypeWA40Full {
		t.Errorf("FaceType = %v, want %v", slot.FaceType, model.FaceTypeWA40Full)
	}
	if slot.IntervalSeconds != 30 {
		t.Errorf("IntervalSeconds = %d, want 30", slot.IntervalSeconds)
	}

	// Verify FindBySessionID returns the slot
	slots, err := repo.FindBySessionID(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("FindBySessionID() failed: %v", err)
	}
	if len(slots) != 1 || slots[0].SlotID != id {
		t.Fatalf("FindBySessionID() = %+v, want slot %v", slots, id)
	}
}

func TestSlotRepo_MultipleSlots_OrderingBySlotLetter(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	owner, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create owner failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, owner.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}

	archerB, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create archer B failed: %v", err)
	}
	archerC, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create archer C failed: %v", err)
	}

	// Create slot C then slot A then slot B
	slotC, err := createTestSlot(ctx, testPool, target.TargetID, archerC.ArcherID, session.SessionID, func(s *model.SlotCreate) {
		s.SlotLetter = model.SlotLetterC
	})
	if err != nil {
		t.Fatalf("create slot C failed: %v", err)
	}

	slotA, err := createTestSlot(ctx, testPool, target.TargetID, owner.ArcherID, session.SessionID, func(s *model.SlotCreate) {
		s.SlotLetter = model.SlotLetterA
	})
	if err != nil {
		t.Fatalf("create slot A failed: %v", err)
	}

	slotB, err := createTestSlot(ctx, testPool, target.TargetID, archerB.ArcherID, session.SessionID, func(s *model.SlotCreate) {
		s.SlotLetter = model.SlotLetterB
	})
	if err != nil {
		t.Fatalf("create slot B failed: %v", err)
	}

	repo := repository.NewSlotRepo(testPool)
	slots, err := repo.FindBySessionID(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("FindBySessionID() failed: %v", err)
	}
	if len(slots) != 3 {
		t.Fatalf("expected 3 slots, got %d", len(slots))
	}

	// Verify all slots exist in result
	foundA, foundB, foundC := false, false, false
	for _, s := range slots {
		switch s.SlotID {
		case slotA.SlotID:
			foundA = true
		case slotB.SlotID:
			foundB = true
		case slotC.SlotID:
			foundC = true
		}
	}
	if !foundA || !foundB || !foundC {
		t.Errorf("missing slots from FindBySessionID: foundA=%v, foundB=%v, foundC=%v", foundA, foundB, foundC)
	}
}

func TestSlotRepo_CountBySessionID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	owner, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create owner failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, owner.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}

	repo := repository.NewSlotRepo(testPool)

	count, err := repo.CountBySessionID(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("CountBySessionID() initial failed: %v", err)
	}
	if count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}

	_, err = createTestSlot(ctx, testPool, target.TargetID, owner.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}

	count, err = repo.CountBySessionID(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("CountBySessionID() after insert failed: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestSlotRepo_Update(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	owner, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create owner failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, owner.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, owner.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}

	repo := repository.NewSlotRepo(testPool)
	isShooting := false
	newInterval := 45
	newFace := model.FaceTypeWA80Full

	err = repo.Update(ctx, model.SlotSet{
		IsShooting:      &isShooting,
		IntervalSeconds: &newInterval,
		FaceType:        &newFace,
	}, model.SlotFilter{
		SlotID: &slot.SlotID,
	})
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	updated, err := repo.FindByID(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if updated.IsShooting {
		t.Errorf("expected is_shooting = false, got true")
	}
	if updated.IntervalSeconds != 45 {
		t.Errorf("IntervalSeconds = %d, want 45", updated.IntervalSeconds)
	}
	if updated.FaceType != model.FaceTypeWA80Full {
		t.Errorf("FaceType = %v, want %v", updated.FaceType, model.FaceTypeWA80Full)
	}
}

func TestSlotRepo_Delete(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	owner, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create owner failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, owner.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, owner.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}

	repo := repository.NewSlotRepo(testPool)
	if err := repo.Delete(ctx, slot.SlotID); err != nil {
		t.Fatalf("Delete() slot failed: %v", err)
	}

	deleted, err := repo.FindByID(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if deleted != nil {
		t.Fatalf("expected slot to be deleted, got: %+v", deleted)
	}

	// Deleting again returns ErrNotFound
	if err := repo.Delete(ctx, slot.SlotID); err != apperror.ErrNotFound {
		t.Errorf("expected ErrNotFound on deleting absent slot, got %v", err)
	}
}

func TestSlotRepo_UniqueArcherPerSession_Constraint(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	owner, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create owner failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, owner.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}

	_, err = createTestSlot(ctx, testPool, target.TargetID, owner.ArcherID, session.SessionID, func(s *model.SlotCreate) {
		s.SlotLetter = model.SlotLetterA
	})
	if err != nil {
		t.Fatalf("create initial slot failed: %v", err)
	}

	// Attempting to assign same archer to another slot in the same session violates uq_archer_per_session
	_, err = createTestSlot(ctx, testPool, target.TargetID, owner.ArcherID, session.SessionID, func(s *model.SlotCreate) {
		s.SlotLetter = model.SlotLetterB
	})
	if err == nil {
		t.Fatalf("expected unique constraint error for duplicate archer in session, got nil")
	}
	if !strings.Contains(err.Error(), "uq_archer_per_session") && !strings.Contains(err.Error(), "duplicate key") {
		t.Errorf("expected uq_archer_per_session violation, got %v", err)
	}
}
