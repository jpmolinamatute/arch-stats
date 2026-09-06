package integration_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestShotRepo_CreateAndFindBySlotID(t *testing.T) {
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
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}

	repo := repository.NewShotRepo(testPool)
	x := 1.25
	y := 2.50
	score := 9
	id, err := repo.Create(ctx, model.ShotCreate{
		SlotID: slot.SlotID,
		X:      &x,
		Y:      &y,
		IsX:    false,
		Score:  &score,
	})
	if err != nil {
		t.Fatalf("Create() shot failed: %v", err)
	}

	shot, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if shot == nil {
		t.Fatalf("expected shot, got nil")
	}
	if shot.ShotID != id {
		t.Errorf("ShotID = %v, want %v", shot.ShotID, id)
	}
	if shot.SlotID != slot.SlotID {
		t.Errorf("SlotID = %v, want %v", shot.SlotID, slot.SlotID)
	}
	if shot.Score == nil || *shot.Score != 9 {
		t.Errorf("Score = %v, want 9", shot.Score)
	}
	if shot.IsX {
		t.Errorf("IsX = true, want false")
	}
	if shot.X == nil || *shot.X != 1.25 || shot.Y == nil || *shot.Y != 2.50 {
		t.Errorf("Coordinates = (%v, %v), want (1.25, 2.50)", shot.X, shot.Y)
	}

	// Verify FindBySlotID
	shots, err := repo.FindBySlotID(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("FindBySlotID() failed: %v", err)
	}
	if len(shots) != 1 || shots[0].ShotID != id {
		t.Fatalf("FindBySlotID() = %+v, want shot %v", shots, id)
	}
}

func TestShotRepo_MultipleShots_ChronologicalOrdering(t *testing.T) {
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
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}

	repo := repository.NewShotRepo(testPool)

	baseTime := time.Now().UTC().Add(-10 * time.Minute)
	t1 := baseTime
	t2 := baseTime.Add(1 * time.Minute)
	t3 := baseTime.Add(2 * time.Minute)

	x := 0.0
	y := 0.0
	s1Score := 7
	s2Score := 8
	s3Score := 10

	// Insert in non-chronological order: t2, t1, t3
	s2ID, err := repo.Create(ctx, model.ShotCreate{
		SlotID:    slot.SlotID,
		X:         &x,
		Y:         &y,
		Score:     &s2Score,
		CreatedAt: &t2,
	})
	if err != nil {
		t.Fatalf("create shot 2 failed: %v", err)
	}

	s1ID, err := repo.Create(ctx, model.ShotCreate{
		SlotID:    slot.SlotID,
		X:         &x,
		Y:         &y,
		Score:     &s1Score,
		CreatedAt: &t1,
	})
	if err != nil {
		t.Fatalf("create shot 1 failed: %v", err)
	}

	s3ID, err := repo.Create(ctx, model.ShotCreate{
		SlotID:    slot.SlotID,
		X:         &x,
		Y:         &y,
		Score:     &s3Score,
		CreatedAt: &t3,
	})
	if err != nil {
		t.Fatalf("create shot 3 failed: %v", err)
	}

	shots, err := repo.FindBySlotID(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("FindBySlotID() failed: %v", err)
	}
	if len(shots) != 3 {
		t.Fatalf("expected 3 shots, got %d", len(shots))
	}

	// Verify chronological ordering
	if shots[0].ShotID != s1ID || shots[1].ShotID != s2ID || shots[2].ShotID != s3ID {
		t.Errorf("ordering mismatch: got [%v, %v, %v], want [%v, %v, %v]",
			shots[0].ShotID, shots[1].ShotID, shots[2].ShotID, s1ID, s2ID, s3ID)
	}
}

func TestShotRepo_UpdateScore(t *testing.T) {
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
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}

	repo := repository.NewShotRepo(testPool)
	x := 2.0
	y := 2.0
	score := 8
	shotID, err := repo.Create(ctx, model.ShotCreate{
		SlotID: slot.SlotID,
		X:      &x,
		Y:      &y,
		Score:  &score,
	})
	if err != nil {
		t.Fatalf("Create() shot failed: %v", err)
	}

	// Update score to 10 with is_x = true and new coordinates
	newScore := 10
	isX := true
	newX := 0.1
	newY := 0.1
	err = repo.Update(ctx, model.ShotSet{
		Score: &newScore,
		IsX:   &isX,
		X:     &newX,
		Y:     &newY,
	}, model.ShotFilter{
		ShotID: &shotID,
	})
	if err != nil {
		t.Fatalf("Update() shot failed: %v", err)
	}

	updated, err := repo.FindByID(ctx, shotID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if updated.Score == nil || *updated.Score != 10 {
		t.Errorf("Score = %v, want 10", updated.Score)
	}
	if !updated.IsX {
		t.Errorf("IsX = false, want true")
	}
	if updated.X == nil || *updated.X != 0.1 || updated.Y == nil || *updated.Y != 0.1 {
		t.Errorf("Coordinates = (%v, %v), want (0.1, 0.1)", updated.X, updated.Y)
	}
}

func TestShotRepo_Delete(t *testing.T) {
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
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}
	shot, err := createTestShot(ctx, testPool, slot.SlotID)
	if err != nil {
		t.Fatalf("create shot failed: %v", err)
	}

	repo := repository.NewShotRepo(testPool)
	if err := repo.Delete(ctx, shot.ShotID); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	found, err := repo.FindByID(ctx, shot.ShotID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if found != nil {
		t.Fatalf("expected shot to be deleted, got %+v", found)
	}

	// Deleting non-existent shot returns ErrNotFound
	if err := repo.Delete(ctx, shot.ShotID); err != apperror.ErrNotFound {
		t.Errorf("expected ErrNotFound deleting absent shot, got %v", err)
	}
}

func TestShotRepo_CountBySlotID_And_LatestShotTime(t *testing.T) {
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
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}

	repo := repository.NewShotRepo(testPool)

	count, err := repo.CountBySlotID(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("CountBySlotID() initial failed: %v", err)
	}
	if count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}

	latestTime, err := repo.GetLatestShotTime(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("GetLatestShotTime() initial error: %v", err)
	}
	if latestTime != nil {
		t.Fatalf("expected nil for empty slot, got %v", latestTime)
	}

	// Create 2 shots
	t1 := time.Now().UTC().Add(-5 * time.Minute)
	t2 := time.Now().UTC().Add(-1 * time.Minute)
	x := 0.0
	y := 0.0
	score := 9

	_, _ = repo.Create(ctx, model.ShotCreate{
		SlotID: slot.SlotID, X: &x, Y: &y, Score: &score, CreatedAt: &t1,
	})
	_, _ = repo.Create(ctx, model.ShotCreate{
		SlotID: slot.SlotID, X: &x, Y: &y, Score: &score, CreatedAt: &t2,
	})

	count, err = repo.CountBySlotID(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("CountBySlotID() failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	latestTime, err = repo.GetLatestShotTime(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("GetLatestShotTime() failed: %v", err)
	}
	if latestTime == nil {
		t.Fatalf("expected non-nil latestTime")
	}
	if latestTime.Sub(t2) > time.Second || t2.Sub(*latestTime) > time.Second {
		t.Errorf("latestTime = %v, want ~%v", latestTime, t2)
	}
}

func TestShotRepo_CreateBatch(t *testing.T) {
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
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}

	repo := repository.NewShotRepo(testPool)
	x := 1.0
	y := 1.0
	score8 := 8
	score9 := 9
	score10 := 10

	batch := []model.ShotCreate{
		{SlotID: slot.SlotID, X: &x, Y: &y, Score: &score8},
		{SlotID: slot.SlotID, X: &x, Y: &y, Score: &score9},
		{SlotID: slot.SlotID, X: &x, Y: &y, Score: &score10, IsX: true},
	}

	ids, err := repo.CreateBatch(ctx, batch)
	if err != nil {
		t.Fatalf("CreateBatch() failed: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("expected 3 generated UUIDs, got %d", len(ids))
	}

	shots, err := repo.FindBySlotID(ctx, slot.SlotID)
	if err != nil {
		t.Fatalf("FindBySlotID() failed: %v", err)
	}
	if len(shots) != 3 {
		t.Fatalf("expected 3 shots in slot, got %d", len(shots))
	}
}

func TestShotRepo_CoordinateAndScore_Constraints(t *testing.T) {
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
	target, err := createTestTarget(ctx, testPool, session.SessionID)
	if err != nil {
		t.Fatalf("create target failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, session.SessionID)
	if err != nil {
		t.Fatalf("create slot failed: %v", err)
	}

	repo := repository.NewShotRepo(testPool)

	// Constraint 1: coords and score must be all null or all non-null
	x := 1.0
	_, err = repo.Create(ctx, model.ShotCreate{
		SlotID: slot.SlotID,
		X:      &x,
		// Y and Score missing
	})
	if err == nil {
		t.Fatalf("expected error violating shot_coords_score_all_or_none, got nil")
	}
	if !strings.Contains(err.Error(), "shot_coords_score_all_or_none") {
		t.Errorf("expected shot_coords_score_all_or_none error, got: %v", err)
	}

	// Constraint 2: is_x requires score = 10
	y := 1.0
	score9 := 9
	_, err = repo.Create(ctx, model.ShotCreate{
		SlotID: slot.SlotID,
		X:      &x,
		Y:      &y,
		Score:  &score9,
		IsX:    true,
	})
	if err == nil {
		t.Fatalf("expected error violating shot_is_x_requires_ten, got nil")
	}
	if !strings.Contains(err.Error(), "shot_is_x_requires_ten") {
		t.Errorf("expected shot_is_x_requires_ten error, got: %v", err)
	}
}
