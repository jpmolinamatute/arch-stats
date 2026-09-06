# Integration Test Infrastructure with testcontainers-go Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the Go integration test infrastructure using `testcontainers-go` to spin up a real PostgreSQL 17 container, execute Goose migrations, expose a shared `pgxpool.Pool` via `TestMain`, provide reusable test helpers and fixtures (`truncateAll`, `createTestArcher`, `createTestSession`, `jwtForArcher`), verify with smoke tests, and mark Task 037 as completed in the tracking documents.

**Architecture:** A dedicated `backend/tests/integration` package utilizes `TestMain` to orchestrate container lifecycle (`postgres.Run`), wait for DB readiness, initialize `pgxpool.Pool`, and apply Goose schema migrations (`../../migrations`). Shared helpers provide FK-safe table truncation and domain fixture factories using existing repository APIs. A smoke test suite validates end-to-end container initialization, table existence, fixture generation, and truncation.

**Tech Stack:** Go 1.27, `testcontainers-go` (v0.35+), `testcontainers-go/modules/postgres`, `pgx/v5` (`pgxpool`), `goose/v3`, `golang-jwt/jwt/v5`, PostgreSQL 17 (Docker container).

**Spec:** [docs/go_refactor/tasks/037-integration_test_infrastructure.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/037-integration_test_infrastructure.md)

## Global Constraints

- Target branch: `refactor/037-integration-test-infrastructure`
- PostgreSQL container image: `postgres:17`
- Integration test package: `package integration_test` located in `backend/tests/integration/`
- Database pool interface: `*pgxpool.Pool` via `repository.NewPool(ctx, dsn, 2, 5)`
- Migrations path: `../../migrations` via `repository.RunMigrations(ctx, testPool, migrationsDir)`
- FK-safe table truncation order: `shot`, `arrow`, `slot`, `target`, `session`, `auth`, `archer`
- Live tracking checklist in [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) and task spec in [docs/go_refactor/tasks/037-integration_test_infrastructure.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/037-integration_test_infrastructure.md) must be marked completed at the end of implementation

---

### Task 1: Switch Git Branch & Install testcontainers-go Dependencies

**Files:**
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Existing Go module dependencies
- Produces: `testcontainers-go` and `testcontainers-go/modules/postgres` installed in `backend/go.mod` and `backend/go.sum`

- [ ] **Step 1: Create and check out feature branch**

```bash
git checkout -b refactor/037-integration-test-infrastructure
```

- [ ] **Step 2: Initialize live task tracker**

Update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md):

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Dependency Installation | IN_PROGRESS | Switch branch to `refactor/037-integration-test-infrastructure` and install testcontainers-go dependencies |
| Task 2: TestMain Lifecycle & PostgreSQL Container Boot | TODO | Create `testmain_test.go` with PostgreSQL 17 container and Goose migration execution |
| Task 3: Shared Integration Test Helpers | TODO | Create `helpers_test.go` with `truncateAll`, `createTestArcher`, `createTestSession`, and `jwtForArcher` |
| Task 4: Smoke Test Suite | TODO | Create `smoke_test.go` and verify container connectivity, table existence, and helper functions |
| Task 5: Verification, Linting, and Marking Tasks Done | TODO | Run tests with race detection, `go vet`, `golangci-lint`, and update tracking checklists |
```

- [ ] **Step 3: Add testcontainers-go dependencies**

```bash
cd backend
go get github.com/testcontainers/testcontainers-go@latest
go get github.com/testcontainers/testcontainers-go/modules/postgres@latest
go mod tidy
```

- [ ] **Step 4: Verify dependencies in go.mod**

Verify that `github.com/testcontainers/testcontainers-go` and `github.com/testcontainers/testcontainers-go/modules/postgres` appear in `backend/go.mod`:

Run: `git diff backend/go.mod`
Expected: Diff showing added `testcontainers-go` module requirements.

- [ ] **Step 5: Mark Task 1 DONE in task.md and commit**

Update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) marking Task 1 as `DONE`.

```bash
git add backend/go.mod backend/go.sum docs/plans/task.md
git commit -m "feat(test): add testcontainers-go dependencies"
```

---

### Task 2: Implement TestMain Lifecycle & PostgreSQL Container Boot

**Files:**
- Create: `backend/tests/integration/testmain_test.go`

**Interfaces:**
- Consumes:
  - `github.com/testcontainers/testcontainers-go/modules/postgres`
  - `repository.NewPool(ctx context.Context, dsn string, minConns, maxConns int32) (*pgxpool.Pool, error)`
  - `repository.RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error`
- Produces:
  - Package-level variable `var testPool *pgxpool.Pool` in `package integration_test`
  - `TestMain(m *testing.M)` setup and teardown lifecycle

- [ ] **Step 1: Write `backend/tests/integration/testmain_test.go`**

Create `backend/tests/integration/testmain_test.go`:

```go
package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:17",
		postgres.WithDatabase("arch-stats-test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	testPool, err = repository.NewPool(ctx, dsn, 2, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create pool: %v\n", err)
		os.Exit(1)
	}

	// Run goose migrations
	migrationsDir := "../../migrations"
	if err := repository.RunMigrations(ctx, testPool, migrationsDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	testPool.Close()
	_ = pgContainer.Terminate(ctx)
	os.Exit(code)
}
```

- [ ] **Step 2: Verify package compiles without errors**

Run: `cd backend && go test -c ./tests/integration/... -o /dev/null`
Expected: Exit code 0 (compilation passes, no syntax or missing import errors).

- [ ] **Step 3: Update task.md and commit**

Update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) marking Task 2 as `DONE`.

```bash
git add backend/tests/integration/testmain_test.go docs/plans/task.md
git commit -m "feat(test): implement TestMain with testcontainers postgres and goose migrations"
```

---

### Task 3: Shared Integration Test Helpers & Fixtures

**Files:**
- Create: `backend/tests/integration/helpers_test.go`

**Interfaces:**
- Consumes:
  - `auth.BuildJWT(archerID uuid.UUID, sid string, issuedAt, expiresAt time.Time, secret, algorithm string) (string, error)`
  - `repository.NewArcherRepo(db DBTX) *ArcherRepo`
  - `repository.NewSessionRepo(db DBTX) *SessionRepo`
  - `model.ArcherCreate`, `model.ArcherRead`, `model.SessionCreate`, `model.SessionRead`
- Produces:
  - `truncateAll(ctx context.Context, pool *pgxpool.Pool) error`
  - `type ArcherOverride func(*model.ArcherCreate)`
  - `createTestArcher(ctx context.Context, pool *pgxpool.Pool, overrides ...ArcherOverride) (*model.ArcherRead, error)`
  - `type SessionOverride func(*model.SessionCreate)`
  - `createTestSession(ctx context.Context, pool *pgxpool.Pool, archerID uuid.UUID, overrides ...SessionOverride) (*model.SessionRead, error)`
  - `jwtForArcher(archerID uuid.UUID, secret string) string`

- [ ] **Step 1: Write `backend/tests/integration/helpers_test.go`**

Create `backend/tests/integration/helpers_test.go`:

```go
package integration_test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

// truncateAll truncates all application tables in foreign-key safe order.
func truncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	tables := []string{
		"shot",
		"arrow",
		"slot",
		"target",
		"session",
		"auth",
		"archer",
	}
	for _, table := range tables {
		if _, err := pool.Exec(ctx, "TRUNCATE "+table+" RESTART IDENTITY CASCADE"); err != nil {
			return fmt.Errorf("truncating table %s: %w", table, err)
		}
	}
	return nil
}

// ArcherOverride is a functional modifier to customize test archer creation.
type ArcherOverride func(*model.ArcherCreate)

// createTestArcher inserts a test archer into the database and returns the created record.
func createTestArcher(ctx context.Context, pool *pgxpool.Pool, overrides ...ArcherOverride) (*model.ArcherRead, error) {
	uniqueID := uuid.New().String()
	payload := model.ArcherCreate{
		FirstName:     "Robin",
		LastName:      "Hood",
		Email:         fmt.Sprintf("archer-%s@example.com", uniqueID[:8]),
		DateOfBirth:   "1990-05-15",
		Gender:        model.GenderMale,
		Bowstyle:      model.BowstyleRecurve,
		DrawWeight:    42.5,
		GoogleSubject: fmt.Sprintf("google-sub-%s", uniqueID),
	}

	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewArcherRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test archer: %w", err)
	}

	return repo.FindByID(ctx, id)
}

// SessionOverride is a functional modifier to customize test shooting session creation.
type SessionOverride func(*model.SessionCreate)

// createTestSession inserts a test shooting session into the database and returns the created record.
func createTestSession(ctx context.Context, pool *pgxpool.Pool, archerID uuid.UUID, overrides ...SessionOverride) (*model.SessionRead, error) {
	payload := model.SessionCreate{
		OwnerArcherID:   archerID,
		SessionLocation: "Outdoor Range",
		IsIndoor:        false,
		IsOpened:        true,
	}

	for _, fn := range overrides {
		fn(&payload)
	}

	repo := repository.NewSessionRepo(pool)
	id, err := repo.Create(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating test session: %w", err)
	}

	return repo.FindByID(ctx, id)
}

// jwtForArcher generates a valid signed HS256 JWT for the given archer UUID and secret.
func jwtForArcher(archerID uuid.UUID, secret string) string {
	now := time.Now().UTC()
	token, err := auth.BuildJWT(archerID, "test-sid", now, now.Add(time.Hour), secret, "HS256")
	if err != nil {
		panic(fmt.Sprintf("jwtForArcher: %v", err))
	}
	return token
}
```

- [ ] **Step 2: Verify helper compilation**

Run: `cd backend && go test -c ./tests/integration/... -o /dev/null`
Expected: Exit code 0 (compilation clean).

- [ ] **Step 3: Update task.md and commit**

Update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) marking Task 3 as `DONE`.

```bash
git add backend/tests/integration/helpers_test.go docs/plans/task.md
git commit -m "feat(test): add integration test helpers and fixtures"
```

---

### Task 4: Smoke Test Suite & Validation

**Files:**
- Create: `backend/tests/integration/smoke_test.go`

**Interfaces:**
- Consumes:
  - `testPool` from `testmain_test.go`
  - `truncateAll`, `createTestArcher`, `createTestSession`, `jwtForArcher` from `helpers_test.go`
- Produces:
  - `TestSmoke_PoolConnected(t *testing.T)`
  - `TestSmoke_TablesExist(t *testing.T)`
  - `TestSmoke_TruncateAll(t *testing.T)`
  - `TestSmoke_HelpersAndLifecycle(t *testing.T)`

- [ ] **Step 1: Write `backend/tests/integration/smoke_test.go`**

Create `backend/tests/integration/smoke_test.go`:

```go
package integration_test

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

func TestSmoke_PoolConnected(t *testing.T) {
	ctx := context.Background()
	if err := testPool.Ping(ctx); err != nil {
		t.Fatalf("pool not connected: %v", err)
	}
}

func TestSmoke_TablesExist(t *testing.T) {
	ctx := context.Background()
	tables := []string{"archer", "session", "target", "slot", "shot", "arrow", "auth"}
	for _, table := range tables {
		var count int
		err := testPool.QueryRow(ctx,
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1",
			table,
		).Scan(&count)
		if err != nil {
			t.Fatalf("query for table %s failed: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("table %s not found in public schema, count = %d", table, count)
		}
	}
}

func TestSmoke_TruncateAll(t *testing.T) {
	ctx := context.Background()
	if err := truncateAll(ctx, testPool); err != nil {
		t.Fatalf("truncateAll failed on initial DB: %v", err)
	}
}

func TestSmoke_HelpersAndLifecycle(t *testing.T) {
	ctx := context.Background()

	// 1. Create archer with default values
	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}
	if archer == nil || archer.ArcherID == [16]byte{} {
		t.Fatalf("expected valid archer, got nil or zero UUID")
	}

	// 2. Create archer with override
	customArcher, err := createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.FirstName = "Marion"
		a.LastName = "Ravenwood"
		a.Bowstyle = model.BowstyleCompound
	})
	if err != nil {
		t.Fatalf("createTestArcher with override failed: %v", err)
	}
	if customArcher.FirstName != "Marion" || customArcher.Bowstyle != model.BowstyleCompound {
		t.Fatalf("custom archer attributes not applied: %+v", customArcher)
	}

	// 3. Create session for archer
	session, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	if session == nil || session.OwnerArcherID != archer.ArcherID {
		t.Fatalf("expected session owned by %s, got %+v", archer.ArcherID, session)
	}

	// 4. Generate and verify JWT
	jwtSecret := "integration-test-super-secret-key-12345"
	tokenStr := jwtForArcher(archer.ArcherID, jwtSecret)
	if tokenStr == "" {
		t.Fatalf("expected non-empty JWT token")
	}

	parsedToken, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !parsedToken.Valid {
		t.Fatalf("failed to parse valid generated token: %v", err)
	}

	// 5. Test truncateAll clears data
	if err := truncateAll(ctx, testPool); err != nil {
		t.Fatalf("truncateAll failed after inserting data: %v", err)
	}

	var archerCount int
	if err := testPool.QueryRow(ctx, "SELECT COUNT(*) FROM archer").Scan(&archerCount); err != nil {
		t.Fatalf("count archer query failed: %v", err)
	}
	if archerCount != 0 {
		t.Fatalf("expected 0 archers after truncateAll, got %d", archerCount)
	}
}
```

- [ ] **Step 2: Run integration tests**

Run: `cd backend && go test ./tests/integration/... -v -count=1`
Expected:
```
=== RUN   TestSmoke_PoolConnected
--- PASS: TestSmoke_PoolConnected (...)
=== RUN   TestSmoke_TablesExist
--- PASS: TestSmoke_TablesExist (...)
=== RUN   TestSmoke_TruncateAll
--- PASS: TestSmoke_TruncateAll (...)
=== RUN   TestSmoke_HelpersAndLifecycle
--- PASS: TestSmoke_HelpersAndLifecycle (...)
PASS
ok      github.com/jpmolinamatute/arch-stats/backend/tests/integration  ...s
```

- [ ] **Step 3: Update task.md and commit**

Update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) marking Task 4 as `DONE`.

```bash
git add backend/tests/integration/smoke_test.go docs/plans/task.md
git commit -m "test(integration): add smoke tests for testcontainers and helpers"
```

---

### Task 5: Full Verification, Linting, and Marking Tasks Done

**Files:**
- Modify: `docs/plans/task.md`
- Modify: `docs/go_refactor/tasks/037-integration_test_infrastructure.md`

**Interfaces:**
- Consumes: Entire backend codebase and test suite
- Produces: Clean test and lint runs, completed tracking checklists

- [ ] **Step 1: Run complete test suite with race detector**

```bash
cd backend
go test -race ./tests/integration/... -v -count=1
go test ./... -v -count=1
```

Expected: All unit tests and integration tests pass with zero race conditions.

- [ ] **Step 2: Run go vet and golangci-lint**

```bash
cd backend
go vet ./...
golangci-lint run ./...
```

Expected: 0 issues reported.

- [ ] **Step 3: Format code with gofumpt if necessary**

```bash
cd backend
gofumpt -l -w tests/integration/
```

- [ ] **Step 4: Update task.md live tracker marking all tasks DONE**

Update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) to reflect all completed tasks:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Dependency Installation | DONE | Switch branch to `refactor/037-integration-test-infrastructure` and install testcontainers-go dependencies |
| Task 2: TestMain Lifecycle & PostgreSQL Container Boot | DONE | Create `testmain_test.go` with PostgreSQL 17 container and Goose migration execution |
| Task 3: Shared Integration Test Helpers | DONE | Create `helpers_test.go` with `truncateAll`, `createTestArcher`, `createTestSession`, and `jwtForArcher` |
| Task 4: Smoke Test Suite | DONE | Create `smoke_test.go` and verify container connectivity, table existence, and helper functions |
| Task 5: Verification, Linting, and Marking Tasks Done | DONE | Run tests with race detection, `go vet`, `golangci-lint`, and update tracking checklists |
```

- [ ] **Step 5: Mark all checklist items as done in Task 037 spec**

In [docs/go_refactor/tasks/037-integration_test_infrastructure.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/037-integration_test_infrastructure.md), mark all Acceptance Criteria and Steps as checked:

```markdown
## Acceptance Criteria

- [x] `backend/tests/integration/testmain_test.go` provides a `TestMain(m)` function that:
    - Starts a PostgreSQL 17 container via `testcontainers-go`
    - Waits for the container to be healthy
    - Creates a `pgxpool.Pool` connected to the container
    - Runs goose migrations against the test database
    - Exposes the pool to all tests via a package-level variable
    - Tears down the container after all tests complete
- [x] `backend/tests/integration/helpers_test.go` provides shared utilities:
    - `truncateAll(ctx, pool)` — truncates all tables in FK-safe order (matching Python's
    `_truncate_all`)
    - `createTestArcher(ctx, pool, overrides)` — inserts a test archer and returns the record
    - `createTestSession(ctx, pool, archerID, overrides)` — inserts a test shooting session
    - `jwtForArcher(archerID, secret)` — generates a valid JWT for test requests
- [x] `backend/tests/integration/smoke_test.go` contains a single smoke test that:
    - Verifies the pool is connected
    - Verifies migrations ran (tables exist)
    - Verifies `truncateAll` works without error
- [x] `go test ./tests/integration/... -v` passes (smoke test connects, migrates, truncates).
- [x] `go vet ./...` reports no issues.

## Steps

- [x] **Step 1: Add testcontainers-go dependency**
- [x] **Step 2: Write `testmain_test.go`**
- [x] **Step 3: Write `helpers_test.go`**
- [x] **Step 4: Write `smoke_test.go`**
- [x] **Step 5: Run tests**
- [x] **Step 6: Run go vet**
- [x] **Step 7: Commit**
```

- [ ] **Step 6: Final commit**

```bash
git add docs/plans/task.md docs/go_refactor/tasks/037-integration_test_infrastructure.md
git commit -m "chore: mark task 037 as complete in tracking docs"
```
