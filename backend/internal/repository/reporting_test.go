package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestReportingRepo_GetSessionSummary_Stub(t *testing.T) {
	repo := repository.NewReportingRepo(&mockDBTX{})
	sessionID := uuid.New()

	summary, err := repo.GetSessionSummary(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected summary report, got nil")
	}
	if summary.SessionID != sessionID {
		t.Errorf("expected sessionID %v, got %v", sessionID, summary.SessionID)
	}
}

func TestReportingRepo_GetArcherPerformance_Stub(t *testing.T) {
	repo := repository.NewReportingRepo(&mockDBTX{})
	archerID := uuid.New()
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	trends, err := repo.GetArcherPerformance(context.Background(), archerID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trends == nil {
		t.Fatal("expected trends slice, got nil")
	}
}
