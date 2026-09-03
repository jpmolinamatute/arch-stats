package model

import (
	"time"

	"github.com/google/uuid"
)

// SessionCreate represents the payload to create a new shooting session.
type SessionCreate struct {
	OwnerArcherID   uuid.UUID `json:"owner_archer_id" validate:"required"`
	SessionLocation string    `json:"session_location" validate:"required,min=1,max=255"`
	IsIndoor        bool      `json:"is_indoor"`
	IsOpened        bool      `json:"is_opened"`
}

// SessionSet represents mutable fields when updating a session.
type SessionSet struct {
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	SessionLocation *string    `json:"session_location,omitempty" validate:"omitempty,min=1,max=255"`
	IsOpened        *bool      `json:"is_opened,omitempty"`
	IsIndoor        *bool      `json:"is_indoor,omitempty"`
}

// SessionFilter represents criteria to query shooting sessions.
type SessionFilter struct {
	SessionID       *uuid.UUID `json:"session_id,omitempty"`
	OwnerArcherID   *uuid.UUID `json:"owner_archer_id,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	SessionLocation *string    `json:"session_location,omitempty"`
	IsOpened        *bool      `json:"is_opened,omitempty"`
	IsIndoor        *bool      `json:"is_indoor,omitempty"`
}

// SessionUpdate wraps filter criteria and mutation data for sessions.
type SessionUpdate struct {
	Where SessionFilter `json:"where"`
	Data  SessionSet    `json:"data"`
}

// SessionRead represents a persisted shooting session.
type SessionRead struct {
	SessionID       uuid.UUID  `json:"session_id"`
	OwnerArcherID   uuid.UUID  `json:"owner_archer_id"`
	SessionLocation string     `json:"session_location"`
	IsIndoor        bool       `json:"is_indoor"`
	IsOpened        bool       `json:"is_opened"`
	CreatedAt       time.Time  `json:"created_at"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
}

// SessionID represents a standalone session identifier wrapper.
type SessionID struct {
	SessionID *uuid.UUID `json:"session_id,omitempty"`
}
