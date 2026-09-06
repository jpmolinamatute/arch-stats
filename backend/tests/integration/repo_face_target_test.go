package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestFaceRepo_FindAll(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewFaceRepo(testPool)

	faces, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() faces error: %v", err)
	}
	if len(faces) == 0 {
		t.Fatalf("expected non-empty face catalog")
	}

	foundWA40 := false
	for _, f := range faces {
		if f.FaceType == model.FaceTypeWA40Full {
			foundWA40 = true
			if f.FaceName == "" || f.ViewBox == 0 {
				t.Errorf("face WA40Full missing attributes: %+v", f)
			}
		}
	}
	if !foundWA40 {
		t.Errorf("FaceTypeWA40Full not found in catalog")
	}
}

func TestFaceRepo_FindByType(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewFaceRepo(testPool)

	faces, err := repo.FindByType(ctx, model.FaceTypeWA80Full)
	if err != nil {
		t.Fatalf("FindByType() error: %v", err)
	}
	if len(faces) == 0 {
		t.Fatalf("expected at least 1 face definition for WA80Full")
	}
	for _, f := range faces {
		if f.FaceType != model.FaceTypeWA80Full {
			t.Errorf("FaceType = %v, want %v", f.FaceType, model.FaceTypeWA80Full)
		}
	}

	// Non-existent face type returns empty slice
	empty, err := repo.FindByType(ctx, model.FaceType("non_existent_face_type"))
	if err != nil {
		t.Fatalf("FindByType(non-existent) error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 results, got %d", len(empty))
	}
}

func TestFaceRepo_FindByID(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewFaceRepo(testPool)

	face, err := repo.FindByID(ctx, string(model.FaceTypeWA60Full))
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if face == nil {
		t.Fatalf("expected to find face for WA60Full, got nil")
	}
	if face.FaceType != model.FaceTypeWA60Full {
		t.Errorf("FaceType = %v, want %v", face.FaceType, model.FaceTypeWA60Full)
	}

	// Non-existent ID returns nil, nil
	missing, err := repo.FindByID(ctx, "unknown-face-id")
	if err != nil {
		t.Fatalf("FindByID(unknown) error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing face ID, got %+v", missing)
	}
}

func TestTargetRepo_CreateAndFindBySlotID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create archer failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	repo := repository.NewTargetRepo(testPool)
	targetID, err := repo.Create(ctx, model.TargetCreate{
		SessionID: session.SessionID,
		Distance:  70,
		Lane:      3,
	})
	if err != nil {
		t.Fatalf("Create() target failed: %v", err)
	}

	target, err := repo.FindByID(ctx, targetID)
	if err != nil {
		t.Fatalf("FindByID() target failed: %v", err)
	}
	if target == nil {
		t.Fatalf("expected target, got nil")
	}
	if target.TargetID != targetID {
		t.Errorf("TargetID = %v, want %v", target.TargetID, targetID)
	}
	if target.Distance != 70 {
		t.Errorf("Distance = %d, want 70", target.Distance)
	}
	if target.Lane != 3 {
		t.Errorf("Lane = %d, want 3", target.Lane)
	}

	// Create slot associated with this target
	slot, err := createTestSlot(ctx, testPool, targetID, archer.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	// Find target by slot ID
	targetsBySlot, err := repo.FindBySlotID(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("FindBySlotID() failed: %v", err)
	}
	if len(targetsBySlot) != 1 || targetsBySlot[0].TargetID != targetID {
		t.Fatalf("FindBySlotID() = %+v, want target %v", targetsBySlot, targetID)
	}
}

func TestTargetRepo_FindBySessionID_OrderedByLane(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create archer failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	repo := repository.NewTargetRepo(testPool)

	// Create targets on lane 5, lane 1, and lane 3
	tLane5, err := repo.Create(ctx, model.TargetCreate{
		SessionID: session.SessionID, Distance: 18, Lane: 5,
	})
	if err != nil {
		t.Fatalf("create lane 5 failed: %v", err)
	}
	tLane1, err := repo.Create(ctx, model.TargetCreate{
		SessionID: session.SessionID, Distance: 18, Lane: 1,
	})
	if err != nil {
		t.Fatalf("create lane 1 failed: %v", err)
	}
	tLane3, err := repo.Create(ctx, model.TargetCreate{
		SessionID: session.SessionID, Distance: 18, Lane: 3,
	})
	if err != nil {
		t.Fatalf("create lane 3 failed: %v", err)
	}

	targets, err := repo.FindBySessionID(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("FindBySessionID() failed: %v", err)
	}
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}

	// Verify ordered by lane ASC (1, 3, 5)
	if targets[0].TargetID != tLane1 || targets[1].TargetID != tLane3 || targets[2].TargetID != tLane5 {
		t.Errorf("targets ordering mismatch: got lanes [%d, %d, %d], want [1, 3, 5]",
			targets[0].Lane, targets[1].Lane, targets[2].Lane)
	}
}

func TestTargetRepo_Update(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create archer failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	repo := repository.NewTargetRepo(testPool)
	id, err := repo.Create(ctx, model.TargetCreate{
		SessionID: session.SessionID, Distance: 18, Lane: 1,
	})
	if err != nil {
		t.Fatalf("Create() target failed: %v", err)
	}

	newDist := 50
	newLane := 4
	err = repo.Update(ctx, model.TargetSet{
		Distance: &newDist,
		Lane:     &newLane,
	}, model.TargetFilter{
		TargetID: &id,
	})
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	updated, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if updated.Distance != 50 {
		t.Errorf("Distance = %d, want 50", updated.Distance)
	}
	if updated.Lane != 4 {
		t.Errorf("Lane = %d, want 4", updated.Lane)
	}
}

func TestTargetRepo_Delete(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create archer failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	repo := repository.NewTargetRepo(testPool)
	id, err := repo.Create(ctx, model.TargetCreate{
		SessionID: session.SessionID, Distance: 18, Lane: 1,
	})
	if err != nil {
		t.Fatalf("Create() target failed: %v", err)
	}

	if err := repo.Delete(ctx, id); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	found, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if found != nil {
		t.Fatalf("expected target to be deleted, got %+v", found)
	}

	// Deleting again returns ErrNotFound
	if err := repo.Delete(ctx, id); err != apperror.ErrNotFound {
		t.Errorf("expected ErrNotFound deleting absent target, got %v", err)
	}
}

func TestTargetRepo_UniqueLanePerSession_Constraint(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("create archer failed: %v", err)
	}
	session, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	repo := repository.NewTargetRepo(testPool)
	_, err = repo.Create(ctx, model.TargetCreate{
		SessionID: session.SessionID, Distance: 18, Lane: 1,
	})
	if err != nil {
		t.Fatalf("Create() initial target failed: %v", err)
	}

	// Duplicate lane on same session violates targets_one_per_session
	_, err = repo.Create(ctx, model.TargetCreate{
		SessionID: session.SessionID, Distance: 30, Lane: 1,
	})
	if err == nil {
		t.Fatalf("expected unique constraint violation for duplicate lane in session, got nil")
	}
	if !strings.Contains(err.Error(), "targets_one_per_session") && !strings.Contains(err.Error(), "duplicate key") {
		t.Errorf("expected targets_one_per_session error, got %v", err)
	}
}
