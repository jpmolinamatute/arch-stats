package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ReportingRepo provides cross-domain analytical reporting queries.
// Methods accept DBTX and will execute raw SQL or aggregation queries as reporting features are developed.
type ReportingRepo struct {
	db DBTX
}

// NewReportingRepo constructs a new ReportingRepo backed by DBTX.
func NewReportingRepo(db DBTX) *ReportingRepo {
	return &ReportingRepo{db: db}
}

// DB returns the underlying DBTX for reporting queries.
func (r *ReportingRepo) DB() DBTX {
	return r.db
}

// GetSessionSummary returns aggregated performance statistics for a session.
// Initial stub returns placeholder data until analytics endpoints are implemented.
func (r *ReportingRepo) GetSessionSummary(ctx context.Context, sessionID uuid.UUID) (*model.SessionSummaryReport, error) {
	_ = ctx
	return &model.SessionSummaryReport{
		SessionID:       sessionID,
		SessionLocation: "Stub Location",
		TotalShots:      0,
		AverageScore:    0.0,
		StartedAt:       time.Now(),
	}, nil
}

// GetArcherPerformance returns historical scoring progression data points for an archer.
// Initial stub returns empty slice until analytics endpoints are implemented.
func (r *ReportingRepo) GetArcherPerformance(ctx context.Context, archerID uuid.UUID, from, to time.Time) ([]model.ScoringTrend, error) {
	_ = ctx
	_ = archerID
	_ = from
	_ = to
	return make([]model.ScoringTrend, 0), nil
}
