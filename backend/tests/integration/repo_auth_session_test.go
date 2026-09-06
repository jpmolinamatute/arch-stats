package integration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestAuthSessionRepo_CreateAndFindByTokenHash(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	repo := repository.NewAuthSessionRepo(testPool)
	hash := sha256.Sum256([]byte("test-session-raw-token-1"))
	tokenHash := hash[:]
	ua := "Mozilla/5.0 (IntegrationTest)"
	ip := "127.0.0.1"
	now := time.Now().UTC().Truncate(time.Microsecond)
	expiresAt := now.Add(24 * time.Hour)

	err = repo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: tokenHash,
		CreatedAt:        now,
		ExpiresAt:        expiresAt,
		UA:               &ua,
		IPInet:           &ip,
	})
	if err != nil {
		t.Fatalf("Create() auth session failed: %v", err)
	}

	found, err := repo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		t.Fatalf("FindByTokenHash() failed: %v", err)
	}
	if found == nil {
		t.Fatalf("expected auth session, got nil")
	}
	if found.ArcherID != archer.ArcherID {
		t.Errorf("ArcherID = %v, want %v", found.ArcherID, archer.ArcherID)
	}
	if !bytes.Equal(found.SessionTokenHash, tokenHash) {
		t.Errorf("SessionTokenHash mismatch")
	}
	if found.UA == nil || *found.UA != ua {
		t.Errorf("UA = %v, want %q", found.UA, ua)
	}
	if found.IPInet == nil || *found.IPInet != ip {
		t.Errorf("IPInet = %v, want %q", found.IPInet, ip)
	}

	// Missing token hash returns nil, nil
	missingHash := sha256.Sum256([]byte("non-existent-token"))
	missing, err := repo.FindByTokenHash(ctx, missingHash[:])
	if err != nil {
		t.Fatalf("FindByTokenHash(missing) error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing token hash, got %+v", missing)
	}
}

func TestAuthSessionRepo_DeleteByArcherID(t *testing.T) {
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

	repo := repository.NewAuthSessionRepo(testPool)

	hash1 := sha256.Sum256([]byte("token-archer1-a"))
	hash2 := sha256.Sum256([]byte("token-archer1-b"))
	hash3 := sha256.Sum256([]byte("token-archer2"))

	expires := time.Now().UTC().Add(time.Hour)
	for _, h := range [][]byte{hash1[:], hash2[:]} {
		if err := repo.Create(ctx, model.AuthSessionCreate{
			ArcherID:         archer1.ArcherID,
			SessionTokenHash: h,
			ExpiresAt:        expires,
		}); err != nil {
			t.Fatalf("creating session for archer 1 failed: %v", err)
		}
	}
	if err := repo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer2.ArcherID,
		SessionTokenHash: hash3[:],
		ExpiresAt:        expires,
	}); err != nil {
		t.Fatalf("creating session for archer 2 failed: %v", err)
	}

	// Delete all sessions for archer 1
	if err := repo.DeleteByArcherID(ctx, archer1.ArcherID); err != nil {
		t.Fatalf("DeleteByArcherID() failed: %v", err)
	}

	// Verify archer 1 sessions are gone
	s1, err := repo.FindByTokenHash(ctx, hash1[:])
	if err != nil || s1 != nil {
		t.Errorf("expected hash1 deleted, got err=%v, session=%+v", err, s1)
	}
	s2, err := repo.FindByTokenHash(ctx, hash2[:])
	if err != nil || s2 != nil {
		t.Errorf("expected hash2 deleted, got err=%v, session=%+v", err, s2)
	}

	// Verify archer 2 session still intact
	s3, err := repo.FindByTokenHash(ctx, hash3[:])
	if err != nil || s3 == nil {
		t.Errorf("expected hash3 present, got err=%v, session=%+v", err, s3)
	}
}

func TestAuthSessionRepo_DeleteExpired(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	repo := repository.NewAuthSessionRepo(testPool)

	expiredHash := sha256.Sum256([]byte("token-expired"))
	activeHash := sha256.Sum256([]byte("token-active"))

	// Create expired session (expires in the past)
	if err := repo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: expiredHash[:],
		ExpiresAt:        time.Now().UTC().Add(-2 * time.Hour),
	}); err != nil {
		t.Fatalf("create expired session failed: %v", err)
	}

	// Create active session (expires in the future)
	if err := repo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: activeHash[:],
		ExpiresAt:        time.Now().UTC().Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("create active session failed: %v", err)
	}

	deletedCount, err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired() failed: %v", err)
	}
	if deletedCount != 1 {
		t.Errorf("deletedCount = %d, want 1", deletedCount)
	}

	// Verify expired is gone and active remains
	exp, err := repo.FindByTokenHash(ctx, expiredHash[:])
	if err != nil || exp != nil {
		t.Errorf("expired session should be deleted, got: %+v", exp)
	}
	act, err := repo.FindByTokenHash(ctx, activeHash[:])
	if err != nil || act == nil {
		t.Errorf("active session should still exist, got: err=%v, act=%+v", err, act)
	}
}

func TestAuthSessionRepo_RevokeAndIndividualDelete(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	repo := repository.NewAuthSessionRepo(testPool)
	hash := sha256.Sum256([]byte("token-to-revoke"))

	if err := repo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: hash[:],
		ExpiresAt:        time.Now().UTC().Add(time.Hour),
	}); err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	// Revoke
	now := time.Now().UTC()
	if err := repo.RevokeByTokenHash(ctx, hash[:], now); err != nil {
		t.Fatalf("RevokeByTokenHash() failed: %v", err)
	}

	session, err := repo.FindByTokenHash(ctx, hash[:])
	if err != nil || session == nil {
		t.Fatalf("FindByTokenHash() failed: %v", err)
	}
	if session.RevokedAt == nil {
		t.Fatalf("expected revoked_at to be set")
	}

	// Revoking again returns ErrNotFound
	if err := repo.RevokeByTokenHash(ctx, hash[:], now); err != apperror.ErrNotFound {
		t.Errorf("expected ErrNotFound revoking already revoked session, got %v", err)
	}

	// Delete by token hash
	if err := repo.DeleteByTokenHash(ctx, hash[:]); err != nil {
		t.Fatalf("DeleteByTokenHash() failed: %v", err)
	}
	if err := repo.DeleteByTokenHash(ctx, hash[:]); err != apperror.ErrNotFound {
		t.Errorf("expected ErrNotFound deleting non-existent session, got %v", err)
	}
}
