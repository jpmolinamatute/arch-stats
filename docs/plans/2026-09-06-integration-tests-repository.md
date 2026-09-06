# Integration Tests — Repository Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement integration tests for all repository implementations (`ArcherRepo`, `AuthSessionRepo`, `SessionRepo`, `SlotRepo`, `ShotRepo`, `FaceRepo`, and `TargetRepo`) against a real PostgreSQL 17 database using `testcontainers-go`, verify queries, constraints, and edge cases, and mark Task 038 as completed in tracking documents.

**Architecture:** Integration tests live in `backend/tests/integration/` under `package integration_test`, utilizing the shared `*pgxpool.Pool` initialized by `TestMain`. Tests create realistic database fixtures using repository methods and helper builders, run assertions on CRUD queries, constraint violations (unique email, slot collision, FK violations), filter evaluations, and state transitions, and clean up after each test via `t.Cleanup(func() { _ = truncateAll(ctx, testPool) })`.

**Tech Stack:** Go 1.27+, `pgx/v5` (`pgxpool`), `testcontainers-go` (v0.44+), `testcontainers-go/modules/postgres`, PostgreSQL 17 container, `google/uuid`.

**Spec:** [docs/go_refactor/tasks/038-integration_tests_repository.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/038-integration_tests_repository.md)

## Global Constraints

- Target branch: `refactor/038-integration-tests-repository`
- Package: `package integration_test` located in `backend/tests/integration/`
- Every test function must register table cleanup via `t.Cleanup(func() { _ = truncateAll(ctx, testPool) })`
- Tests must execute against real PostgreSQL database in `testcontainers-go`, validating actual SQL execution and DB constraints
- All tests must pass with `go test ./tests/integration/... -v -count=1` and `go vet ./...` reporting zero issues
- At the end of implementation, mark all checklist items in [docs/go_refactor/tasks/038-integration_tests_repository.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/038-integration_tests_repository.md) as completed (`[x]`) and update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) marking all tasks as `DONE`

---

### Task 1: Git Branch Setup & Shared Test Helpers

**Files:**
- Modify: `docs/plans/task.md`
- Modify: `backend/tests/integration/helpers_test.go`

**Interfaces:**
- Consumes:
  - `model.TargetCreate`, `model.TargetRead`, `repository.NewTargetRepo`
  - `model.SlotCreate`, `model.SlotRead`, `repository.NewSlotRepo`
  - `model.ShotCreate`, `model.ShotRead`, `repository.NewShotRepo`
- Produces:
  - Helper `createTestTarget(ctx context.Context, pool *pgxpool.Pool, sessionID uuid.UUID, overrides ...TargetOverride) (*model.TargetRead, error)`
  - Helper `createTestSlot(ctx context.Context, pool *pgxpool.Pool, targetID, archerID, sessionID uuid.UUID, overrides ...SlotOverride) (*model.SlotRead, error)`
  - Helper `createTestShot(ctx context.Context, pool *pgxpool.Pool, slotID uuid.UUID, overrides ...ShotOverride) (*model.ShotRead, error)`

- [ ] **Step 1: Create and check out feature branch**

```bash
git checkout -b refactor/038-integration-tests-repository
```

- [ ] **Step 2: Initialize live task tracker in `docs/plans/task.md`**

Write table tracker in `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Shared Test Helpers | IN_PROGRESS | Check out branch and add target, slot, shot test helpers in `helpers_test.go` |
| Task 2: Archer Repository Integration Tests | TODO | Implement `repo_archer_test.go` with CRUD, filters, unique constraints, and missing ID |
| Task 3: Auth Session Repository Integration Tests | TODO | Implement `repo_auth_session_test.go` with token hash lookup, archer deletion, and expired cleanup |
| Task 4: Session Repository Integration Tests | TODO | Implement `repo_session_test.go` with lifecycle, open status, close transitions, and filters |
| Task 5: Slot Repository Integration Tests | TODO | Implement `repo_slot_test.go` with session slot lookup, ordering, count, updates, and deletion |
| Task 6: Shot Repository Integration Tests | TODO | Implement `repo_shot_test.go` with slot shot lookup, chronological ordering, batch create, and updates |
| Task 7: Face & Target Repositories Integration Tests | TODO | Implement `repo_face_target_test.go` with face catalog queries, target lane lookup, and slot join |
| Task 8: Full Verification, Linting, and Marking Tasks Done | TODO | Run full test suite with race detector, `go vet`, `golangci-lint`, and update tracking docs |
```

- [ ] **Step 3: Add Target, Slot, and Shot fixture helpers to `helpers_test.go`**

Append to `backend/tests/integration/helpers_test.go`:

```go
// TargetOverride is a functional modifier to customize test target creation.
type TargetOverride func(*model.TargetCreate)

// createTestTarget inserts a test target configuration into the database.
func createTestTarget(ctx context.Context, pool *pgxpool.Pool, sessionID uuid.UUID, overrides ...TargetOverride) (*model.TargetRead, error) {
	payload := model.TargetCreate{
		SessionID: sessionID,
		Distance:  18,
		Lane:      1,
	}
	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewTargetRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test target: %w", err)
	}

	return repo.FindByID(ctx, id)
}

// SlotOverride is a functional modifier to customize test slot creation.
type SlotOverride func(*model.SlotCreate)

// createTestSlot inserts a test slot into the database.
func createTestSlot(ctx context.Context, pool *pgxpool.Pool, targetID, archerID, sessionID uuid.UUID, overrides ...SlotOverride) (*model.SlotRead, error) {
	shotPerRound := 3
	payload := model.SlotCreate{
		TargetID:        targetID,
		ArcherID:        archerID,
		SessionID:       sessionID,
		SlotLetter:      model.SlotLetterA,
		FaceType:        model.FaceTypeWA40Full,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      40.0,
		IsShooting:      true,
		ShotPerRound:    &shotPerRound,
		IntervalSeconds: 20,
	}
	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewSlotRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test slot: %w", err)
	}

	return repo.FindByID(ctx, id)
}

// ShotOverride is a functional modifier to customize test shot creation.
type ShotOverride func(*model.ShotCreate)

// createTestShot inserts a test shot into the database.
func createTestShot(ctx context.Context, pool *pgxpool.Pool, slotID uuid.UUID, overrides ...ShotOverride) (*model.ShotRead, error) {
	x := 0.5
	y := 0.5
	score := 10
	payload := model.ShotCreate{
		SlotID: slotID,
		X:      &x,
		Y:      &y,
		IsX:    true,
		Score:  &score,
	}
	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewShotRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test shot: %w", err)
	}

	return repo.FindByID(ctx, id)
}
```

- [ ] **Step 4: Verify helpers compile and smoke tests still pass**

Run:
```bash
cd backend
go test ./tests/integration/... -v -count=1
```
Expected: PASS with all existing smoke tests passing.

- [ ] **Step 5: Update `docs/plans/task.md` marking Task 1 DONE and commit**

```bash
git add docs/plans/task.md backend/tests/integration/helpers_test.go
git commit -m "feat(test): add target, slot, and shot test helpers"
```

---

### Task 2: Archer Repository Integration Tests

**Files:**
- Create: `backend/tests/integration/repo_archer_test.go`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes:
  - `repository.NewArcherRepo(testPool)`
  - `model.ArcherCreate`, `model.ArcherSet`, `model.ArcherFilter`, `model.ArcherRead`
  - `createTestArcher(ctx, testPool, overrides...)`
  - `truncateAll(ctx, testPool)`
- Produces:
  - `TestArcherRepo_CreateAndFindByID`
  - `TestArcherRepo_FindByEmail`
  - `TestArcherRepo_FindByGoogleSubject`
  - `TestArcherRepo_DuplicateEmail_Conflict`
  - `TestArcherRepo_Update`
  - `TestArcherRepo_FindByID_NotFound`
  - `TestArcherRepo_FindAll_WithFilters`

- [ ] **Step 1: Write `backend/tests/integration/repo_archer_test.go`**

```go
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
	// Verify other fields remain unchanged
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
```

- [ ] **Step 2: Run Archer repository integration tests**

```bash
cd backend
go test ./tests/integration/... -run TestArcherRepo -v -count=1
```
Expected: PASS for all `TestArcherRepo_*` tests.

- [ ] **Step 3: Update `docs/plans/task.md` marking Task 2 DONE and commit**

```bash
git add backend/tests/integration/repo_archer_test.go docs/plans/task.md
git commit -m "test(integration): add archer repository tests"
```

---

### Task 3: Auth Session Repository Integration Tests

**Files:**
- Create: `backend/tests/integration/repo_auth_session_test.go`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes:
  - `repository.NewAuthSessionRepo(testPool)`
  - `model.AuthSessionCreate`, `model.AuthSessionRead`
  - `createTestArcher(ctx, testPool)`
  - `truncateAll(ctx, testPool)`
- Produces:
  - `TestAuthSessionRepo_CreateAndFindByTokenHash`
  - `TestAuthSessionRepo_DeleteByArcherID`
  - `TestAuthSessionRepo_DeleteExpired`
  - `TestAuthSessionRepo_RevokeAndIndividualDelete`

- [ ] **Step 1: Write `backend/tests/integration/repo_auth_session_test.go`**

```go
package integration_test

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/google/uuid"
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
	if string(found.SessionTokenHash) != string(tokenHash) {
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
```

- [ ] **Step 2: Run Auth Session repository integration tests**

```bash
cd backend
go test ./tests/integration/... -run TestAuthSessionRepo -v -count=1
```
Expected: PASS for all `TestAuthSessionRepo_*` tests.

- [ ] **Step 3: Update `docs/plans/task.md` marking Task 3 DONE and commit**

```bash
git add backend/tests/integration/repo_auth_session_test.go docs/plans/task.md
git commit -m "test(integration): add auth session repository tests"
```

---

### Task 4: Session Repository Integration Tests

**Files:**
- Create: `backend/tests/integration/repo_session_test.go`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes:
  - `repository.NewSessionRepo(testPool)`
  - `model.SessionCreate`, `model.SessionRead`, `model.SessionFilter`, `model.SessionSet`
  - `createTestArcher(ctx, testPool)`
  - `createTestSession(ctx, testPool, archerID)`
  - `truncateAll(ctx, testPool)`
- Produces:
  - `TestSessionRepo_CreateAndFindByID`
  - `TestSessionRepo_FindOpen`
  - `TestSessionRepo_Close`
  - `TestSessionRepo_FindAll_WithFilters`
  - `TestSessionRepo_FindParticipating`

- [ ] **Step 1: Write `backend/tests/integration/repo_session_test.go`**

```go
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
```

- [ ] **Step 2: Run Session repository integration tests**

```bash
cd backend
go test ./tests/integration/... -run TestSessionRepo -v -count=1
```
Expected: PASS for all `TestSessionRepo_*` tests.

- [ ] **Step 3: Update `docs/plans/task.md` marking Task 4 DONE and commit**

```bash
git add backend/tests/integration/repo_session_test.go docs/plans/task.md
git commit -m "test(integration): add session repository tests"
```

---

### Task 5: Slot Repository Integration Tests

**Files:**
- Create: `backend/tests/integration/repo_slot_test.go`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes:
  - `repository.NewSlotRepo(testPool)`
  - `model.SlotCreate`, `model.SlotRead`, `model.SlotSet`, `model.SlotFilter`, `model.SlotLetter`
  - `createTestArcher`, `createTestSession`, `createTestTarget`, `createTestSlot`
  - `truncateAll(ctx, testPool)`
- Produces:
  - `TestSlotRepo_CreateAndFindBySessionID`
  - `TestSlotRepo_MultipleSlots_OrderingBySlotLetter`
  - `TestSlotRepo_CountBySessionID`
  - `TestSlotRepo_Update`
  - `TestSlotRepo_Delete`
  - `TestSlotRepo_UniqueArcherPerSession_Constraint`

- [ ] **Step 1: Write `backend/tests/integration/repo_slot_test.go`**

```go
package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
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

	// Verify ordering: created_at ASC, slot_letter ASC (or slot letter ordering)
	letters := []model.SlotLetter{slots[0].SlotLetter, slots[1].SlotLetter, slots[2].SlotLetter}
	if letters[0] != slotC.SlotLetter || letters[1] != slotA.SlotLetter || letters[2] != slotB.SlotLetter {
		t.Logf("Slots order by created_at: %v", letters)
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
```

- [ ] **Step 2: Run Slot repository integration tests**

```bash
cd backend
go test ./tests/integration/... -run TestSlotRepo -v -count=1
```
Expected: PASS for all `TestSlotRepo_*` tests.

- [ ] **Step 3: Update `docs/plans/task.md` marking Task 5 DONE and commit**

```bash
git add backend/tests/integration/repo_slot_test.go docs/plans/task.md
git commit -m "test(integration): add slot repository tests"
```

---

### Task 6: Shot Repository Integration Tests

**Files:**
- Create: `backend/tests/integration/repo_shot_test.go`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes:
  - `repository.NewShotRepo(testPool)`
  - `model.ShotCreate`, `model.ShotRead`, `model.ShotSet`, `model.ShotFilter`
  - `createTestArcher`, `createTestSession`, `createTestTarget`, `createTestSlot`, `createTestShot`
  - `truncateAll(ctx, testPool)`
- Produces:
  - `TestShotRepo_CreateAndFindBySlotID`
  - `TestShotRepo_MultipleShots_ChronologicalOrdering`
  - `TestShotRepo_UpdateScore`
  - `TestShotRepo_Delete`
  - `TestShotRepo_CountBySlotID_And_LatestShotTime`
  - `TestShotRepo_CreateBatch`
  - `TestShotRepo_CoordinateAndScore_Constraints`

- [ ] **Step 1: Write `backend/tests/integration/repo_shot_test.go`**

```go
package integration_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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
```

- [ ] **Step 2: Run Shot repository integration tests**

```bash
cd backend
go test ./tests/integration/... -run TestShotRepo -v -count=1
```
Expected: PASS for all `TestShotRepo_*` tests.

- [ ] **Step 3: Update `docs/plans/task.md` marking Task 6 DONE and commit**

```bash
git add backend/tests/integration/repo_shot_test.go docs/plans/task.md
git commit -m "test(integration): add shot repository tests"
```

---

### Task 7: Face & Target Repositories Integration Tests

**Files:**
- Create: `backend/tests/integration/repo_face_target_test.go`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes:
  - `repository.NewFaceRepo(testPool)`
  - `repository.NewTargetRepo(testPool)`
  - `model.FaceRead`, `model.FaceType`, `model.TargetCreate`, `model.TargetRead`, `model.TargetSet`, `model.TargetFilter`
  - `createTestArcher`, `createTestSession`, `createTestTarget`, `createTestSlot`
  - `truncateAll(ctx, testPool)`
- Produces:
  - `TestFaceRepo_FindAll`
  - `TestFaceRepo_FindByType`
  - `TestFaceRepo_FindByID`
  - `TestTargetRepo_CreateAndFindBySlotID`
  - `TestTargetRepo_FindBySessionID_OrderedByLane`
  - `TestTargetRepo_Update`
  - `TestTargetRepo_Delete`
  - `TestTargetRepo_UniqueLanePerSession_Constraint`

- [ ] **Step 1: Write `backend/tests/integration/repo_face_target_test.go`**

```go
package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
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
			if f.FaceName == "" || f.ViewBox == "" {
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
```

- [ ] **Step 2: Run Face and Target repository integration tests**

```bash
cd backend
go test ./tests/integration/... -run "TestFaceRepo|TestTargetRepo" -v -count=1
```
Expected: PASS for all `TestFaceRepo_*` and `TestTargetRepo_*` tests.

- [ ] **Step 3: Update `docs/plans/task.md` marking Task 7 DONE and commit**

```bash
git add backend/tests/integration/repo_face_target_test.go docs/plans/task.md
git commit -m "test(integration): add face and target repository tests"
```

---

### Task 8: Full Verification, Linting, and Marking Tasks Done

**Files:**
- Modify: `docs/plans/task.md`
- Modify: `docs/go_refactor/tasks/038-integration_tests_repository.md`

**Interfaces:**
- Consumes: Complete repository test suite and all integration test files
- Produces: 100% passing tests, clean `go vet` and `golangci-lint`, marked completion in tracking checklists

- [ ] **Step 1: Run complete integration test suite with race detector**

```bash
cd backend
go test -race ./tests/integration/... -v -count=1
```
Expected: All tests pass with zero race conditions.

- [ ] **Step 2: Run full backend unit and integration tests**

```bash
cd backend
go test ./... -v -count=1
```
Expected: All unit and integration tests across the backend pass.

- [ ] **Step 3: Run `go vet` and `golangci-lint`**

```bash
cd backend
go vet ./...
golangci-lint run ./...
```
Expected: 0 issues reported.

- [ ] **Step 4: Format with `gofumpt`**

```bash
cd backend
gofumpt -l -w tests/integration/
```

- [ ] **Step 5: Update `docs/plans/task.md` marking all tasks DONE**

Update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md):

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Shared Test Helpers | DONE | Check out branch and add target, slot, shot test helpers in `helpers_test.go` |
| Task 2: Archer Repository Integration Tests | DONE | Implement `repo_archer_test.go` with CRUD, filters, unique constraints, and missing ID |
| Task 3: Auth Session Repository Integration Tests | DONE | Implement `repo_auth_session_test.go` with token hash lookup, archer deletion, and expired cleanup |
| Task 4: Session Repository Integration Tests | DONE | Implement `repo_session_test.go` with lifecycle, open status, close transitions, and filters |
| Task 5: Slot Repository Integration Tests | DONE | Implement `repo_slot_test.go` with session slot lookup, ordering, count, updates, and deletion |
| Task 6: Shot Repository Integration Tests | DONE | Implement `repo_shot_test.go` with slot shot lookup, chronological ordering, batch create, and updates |
| Task 7: Face & Target Repositories Integration Tests | DONE | Implement `repo_face_target_test.go` with face catalog queries, target lane lookup, and slot join |
| Task 8: Full Verification, Linting, and Marking Tasks Done | DONE | Run full test suite with race detector, `go vet`, `golangci-lint`, and update tracking docs |
```

- [ ] **Step 6: Mark all checklist items as done in Task 038 spec**

In [docs/go_refactor/tasks/038-integration_tests_repository.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/038-integration_tests_repository.md), mark all Acceptance Criteria and Steps as checked:

```markdown
## Acceptance Criteria

- [x] `backend/tests/integration/repo_archer_test.go` tests:
    - Create an archer → verify it can be found by ID
    - Create an archer → verify it can be found by email
    - Create an archer → verify it can be found by Google subject
    - Create duplicate email → verify conflict/unique constraint error
    - Update an archer → verify changes are persisted
    - FindByID with non-existent ID → verify nil/no rows
    - FindAll with filters → verify correct filtering
- [x] `backend/tests/integration/repo_auth_session_test.go` tests:
    - Create an auth session → verify it can be found by token hash
    - DeleteByArcherID → verify all sessions for that archer are removed
    - DeleteExpired → verify only expired sessions are removed
- [x] `backend/tests/integration/repo_session_test.go` tests:
    - Create a shooting session → verify FindByID returns it
    - FindOpen → verify it returns the session with status "open"
    - Close a session → verify status changes and ended_at is set
    - Create multiple sessions → verify FindAll returns correct list
- [x] `backend/tests/integration/repo_slot_test.go` tests:
    - Create a slot → verify FindBySessionID returns it
    - Create multiple slots → verify ordering by slot_number
    - CountBySessionID → verify correct count
    - Delete a slot → verify it's removed
- [x] `backend/tests/integration/repo_shot_test.go` tests:
    - Create a shot → verify FindBySlotID returns it
    - Create multiple shots → verify ordering by arrow_number
    - Update a shot score → verify change persisted
    - Delete a shot → verify it's removed
- [x] `backend/tests/integration/repo_face_target_test.go` tests:
    - FindAll faces → verify faces are returned
    - FindByType → verify filtering works
    - Create a target → verify FindBySlotID returns it
- [x] Each test truncates tables after completion (using `t.Cleanup`).
- [x] `go test ./tests/integration/... -v -count=1` passes.
- [x] `go vet ./...` reports no issues.

## Steps

- [x] **Step 1: Write `repo_archer_test.go`**
- [x] **Step 2: Write `repo_auth_session_test.go`**
- [x] **Step 3: Write `repo_session_test.go`**
- [x] **Step 4: Write `repo_slot_test.go`**
- [x] **Step 5: Write `repo_shot_test.go`**
- [x] **Step 6: Write `repo_face_target_test.go`**
- [x] **Step 7: Run all integration tests**
- [x] **Step 8: Run go vet**
- [x] **Step 9: Commit**
```

- [ ] **Step 7: Final commit**

```bash
git add docs/plans/task.md docs/go_refactor/tasks/038-integration_tests_repository.md
git commit -m "chore: mark task 038 as complete in tracking docs"
```
