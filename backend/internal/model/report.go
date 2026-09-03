package model

import (
	"time"

	"github.com/google/uuid"
)

// SessionSummaryReport represents aggregated performance statistics for a single completed session.
type SessionSummaryReport struct {
	SessionID       uuid.UUID  `json:"session_id"`
	SessionLocation string     `json:"session_location"`
	TotalShots      int        `json:"total_shots"`
	AverageScore    float64    `json:"average_score"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
}

// ScoringTrend represents a discrete data point in a time-series of an archer's scoring progression.
type ScoringTrend struct {
	SessionID    uuid.UUID `json:"session_id"`
	Timestamp    time.Time `json:"timestamp"`
	AverageScore float64   `json:"average_score"`
	TotalShots   int       `json:"total_shots"`
}

// ArcherPerformanceReport represents historical career metrics across multiple sessions.
type ArcherPerformanceReport struct {
	ArcherID      uuid.UUID `json:"archer_id"`
	TotalSessions int       `json:"total_sessions"`
	TotalShots    int       `json:"total_shots"`
	AverageScore  float64   `json:"average_score"`
	BestScore     int       `json:"best_score"`
	TotalXCount   int       `json:"total_x_count"`
}
