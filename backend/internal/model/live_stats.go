package model

import (
	"time"

	"github.com/google/uuid"
)

// ShotScore represents score and timing information for an individual shot in live views.
type ShotScore struct {
	ShotID    uuid.UUID `json:"shot_id"`
	Score     int       `json:"score" validate:"gte=0,lte=10"`
	IsX       bool      `json:"is_x"`
	CreatedAt time.Time `json:"created_at"`
}

// Stats represents real-time aggregate scoring metrics for a single slot.
type Stats struct {
	SlotID        uuid.UUID `json:"slot_id"`
	NumberOfShots int       `json:"number_of_shots" validate:"gte=0"`
	TotalScore    int       `json:"total_score" validate:"gte=0"`
	MaxScore      int       `json:"max_score" validate:"gte=0"`
	Mean          float64   `json:"mean"`
}

// LiveStat aggregates recent shots and running statistics for a slot.
type LiveStat struct {
	Scores []ShotScore `json:"scores"`
	Stats  Stats       `json:"stats"`
}

// WebSocketMessage represents a live update broadcast over WebSockets.
type WebSocketMessage struct {
	TS          time.Time     `json:"ts"`
	ContentType WSContentType `json:"content_type"`
	Content     LiveStat      `json:"content"`
}
