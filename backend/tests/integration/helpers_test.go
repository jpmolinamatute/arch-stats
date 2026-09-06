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
