package model

import (
	"time"

	"github.com/google/uuid"
)

// ShotCreate represents the payload to register a single shot.
// Coordinates x, y, and score must either all be present or all be nil.
type ShotCreate struct {
	SlotID    uuid.UUID  `json:"slot_id" validate:"required"`
	X         *float64   `json:"x,omitempty"`
	Y         *float64   `json:"y,omitempty"`
	IsX       bool       `json:"is_x"`
	Score     *int       `json:"score,omitempty" validate:"omitempty,gte=0,lte=10"`
	ArrowID   *uuid.UUID `json:"arrow_id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// ShotSet represents updates to mutable shot fields.
type ShotSet struct {
	X       *float64   `json:"x,omitempty"`
	Y       *float64   `json:"y,omitempty"`
	IsX     *bool      `json:"is_x,omitempty"`
	Score   *int       `json:"score,omitempty" validate:"omitempty,gte=0,lte=10"`
	ArrowID *uuid.UUID `json:"arrow_id,omitempty"`
}

// ShotFilter represents criteria to query shots.
type ShotFilter struct {
	ShotID    *uuid.UUID `json:"shot_id,omitempty"`
	SlotID    *uuid.UUID `json:"slot_id,omitempty"`
	X         *float64   `json:"x,omitempty"`
	Y         *float64   `json:"y,omitempty"`
	Score     *int       `json:"score,omitempty" validate:"omitempty,gte=0,lte=10"`
	ArrowID   *uuid.UUID `json:"arrow_id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// ShotUpdate wraps filter and update data for shots.
type ShotUpdate struct {
	Where ShotFilter `json:"where"`
	Data  ShotSet    `json:"data"`
}

// ShotRead represents a persisted shot record.
type ShotRead struct {
	ShotID    uuid.UUID  `json:"shot_id"`
	SlotID    uuid.UUID  `json:"slot_id"`
	X         *float64   `json:"x,omitempty"`
	Y         *float64   `json:"y,omitempty"`
	IsX       bool       `json:"is_x"`
	Score     *int       `json:"score,omitempty"`
	ArrowID   *uuid.UUID `json:"arrow_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ShotID represents a standalone shot identifier wrapper.
type ShotID struct {
	ShotID uuid.UUID `json:"shot_id"`
}
