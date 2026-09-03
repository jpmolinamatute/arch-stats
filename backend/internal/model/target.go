package model

import (
	"time"

	"github.com/google/uuid"
)

// TargetCreate represents the payload to create a lane target in a session.
type TargetCreate struct {
	SessionID uuid.UUID `json:"session_id" validate:"required"`
	Distance  int       `json:"distance" validate:"required,gte=1,lte=100"`
	Lane      int       `json:"lane" validate:"required,gte=1,lte=100"`
}

// TargetSet represents updates to a target's distance or lane.
type TargetSet struct {
	Distance *int `json:"distance,omitempty" validate:"omitempty,gte=1,lte=100"`
	Lane     *int `json:"lane,omitempty" validate:"omitempty,gte=1,lte=100"`
}

// TargetFilter represents criteria to filter targets.
type TargetFilter struct {
	TargetID  *uuid.UUID `json:"target_id,omitempty"`
	SessionID *uuid.UUID `json:"session_id,omitempty"`
	Distance  *int       `json:"distance,omitempty"`
	Lane      *int       `json:"lane,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// TargetUpdate wraps filter criteria and mutation data for targets.
type TargetUpdate struct {
	Where TargetFilter `json:"where"`
	Data  TargetSet    `json:"data"`
}

// TargetRead represents a persisted target with lane, distance, and optional occupancy count.
type TargetRead struct {
	TargetID  uuid.UUID `json:"target_id"`
	SessionID uuid.UUID `json:"session_id"`
	Distance  int       `json:"distance"`
	Lane      int       `json:"lane"`
	Occupied  *int      `json:"occupied,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
