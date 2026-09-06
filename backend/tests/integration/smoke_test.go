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
