package model

import (
	"time"

	"github.com/google/uuid"
)

// SlotCreate represents the payload to directly assign an archer to a target slot.
type SlotCreate struct {
	ArcherID        uuid.UUID  `json:"archer_id" validate:"required"`
	SessionID       uuid.UUID  `json:"session_id" validate:"required"`
	TargetID        uuid.UUID  `json:"target_id" validate:"required"`
	SlotLetter      SlotLetter `json:"slot_letter" validate:"required"`
	FaceType        FaceType   `json:"face_type" validate:"required"`
	Bowstyle        Bowstyle   `json:"bowstyle" validate:"required"`
	DrawWeight      float64    `json:"draw_weight" validate:"required,gt=0,lte=200"`
	ClubID          *uuid.UUID `json:"club_id,omitempty"`
	IsShooting      bool       `json:"is_shooting"`
	ShotPerRound    *int       `json:"shot_per_round,omitempty" validate:"omitempty,gte=3,lte=10"`
	IntervalSeconds int        `json:"interval_seconds" validate:"gte=1,lte=100"`
}

// SlotJoinRequest represents a request by an archer to join a session at a preferred distance.
type SlotJoinRequest struct {
	ArcherID        uuid.UUID  `json:"archer_id" validate:"required"`
	SessionID       uuid.UUID  `json:"session_id" validate:"required"`
	Distance        int        `json:"distance" validate:"required,gte=1,lte=100"`
	FaceType        FaceType   `json:"face_type" validate:"required"`
	Bowstyle        Bowstyle   `json:"bowstyle" validate:"required"`
	DrawWeight      float64    `json:"draw_weight" validate:"required,gt=0,lte=200"`
	ClubID          *uuid.UUID `json:"club_id,omitempty"`
	IsShooting      bool       `json:"is_shooting"`
	ShotPerRound    *int       `json:"shot_per_round,omitempty" validate:"omitempty,gte=3,lte=10"`
	IntervalSeconds int        `json:"interval_seconds" validate:"gte=1,lte=100"`
}

// SlotJoinResponse represents the assigned slot code and slot ID returned upon joining.
type SlotJoinResponse struct {
	SlotID uuid.UUID `json:"slot_id"`
	Slot   string    `json:"slot"`
}

// SlotReJoinRequest represents a request to re-enter an existing slot assignment.
type SlotReJoinRequest struct {
	SlotID    uuid.UUID `json:"slot_id" validate:"required"`
	SessionID uuid.UUID `json:"session_id" validate:"required"`
	ArcherID  uuid.UUID `json:"archer_id" validate:"required"`
}

// SlotLeaveRequest represents an archer leaving their slot.
type SlotLeaveRequest struct {
	ArcherID  uuid.UUID `json:"archer_id" validate:"required"`
	SessionID uuid.UUID `json:"session_id" validate:"required"`
}

// SlotID represents a standalone slot identifier wrapper.
type SlotID struct {
	SlotID uuid.UUID `json:"slot_id"`
}

// SlotRead represents a persisted slot assignment record.
type SlotRead struct {
	SlotID          uuid.UUID  `json:"slot_id"`
	TargetID        uuid.UUID  `json:"target_id"`
	ArcherID        uuid.UUID  `json:"archer_id"`
	SessionID       uuid.UUID  `json:"session_id"`
	SlotLetter      SlotLetter `json:"slot_letter"`
	Slot            *string    `json:"slot,omitempty"`
	FaceType        FaceType   `json:"face_type"`
	Bowstyle        Bowstyle   `json:"bowstyle"`
	DrawWeight      float64    `json:"draw_weight"`
	ClubID          *uuid.UUID `json:"club_id,omitempty"`
	IsShooting      bool       `json:"is_shooting"`
	ShotPerRound    *int       `json:"shot_per_round,omitempty"`
	IntervalSeconds int        `json:"interval_seconds"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}

// SlotSet represents mutable fields on an existing slot assignment.
type SlotSet struct {
	IsShooting      *bool       `json:"is_shooting,omitempty"`
	FaceType        *FaceType   `json:"face_type,omitempty"`
	SlotLetter      *SlotLetter `json:"slot_letter,omitempty"`
	ShotPerRound    *int        `json:"shot_per_round,omitempty" validate:"omitempty,gte=3,lte=10"`
	IntervalSeconds *int        `json:"interval_seconds,omitempty" validate:"omitempty,gte=1,lte=100"`
}

// SlotFilter represents criteria to query slot assignments.
type SlotFilter struct {
	SlotID       *uuid.UUID  `json:"slot_id,omitempty"`
	TargetID     *uuid.UUID  `json:"target_id,omitempty"`
	ArcherID     *uuid.UUID  `json:"archer_id,omitempty"`
	SessionID    *uuid.UUID  `json:"session_id,omitempty"`
	SlotLetter   *SlotLetter `json:"slot_letter,omitempty"`
	IsShooting   *bool       `json:"is_shooting,omitempty"`
	ShotPerRound *int        `json:"shot_per_round,omitempty"`
	CreatedAt    *time.Time  `json:"created_at,omitempty"`
}

// SlotUpdate wraps filter criteria and mutation data for slot assignments.
type SlotUpdate struct {
	Where SlotFilter `json:"where"`
	Data  SlotSet    `json:"data"`
}

// FullSlotInfo represents a comprehensive denormalized slot assignment including lane, distance, and composite code.
type FullSlotInfo struct {
	SlotID          uuid.UUID  `json:"slot_id"`
	TargetID        uuid.UUID  `json:"target_id"`
	ArcherID        uuid.UUID  `json:"archer_id"`
	SessionID       uuid.UUID  `json:"session_id"`
	SlotLetter      SlotLetter `json:"slot_letter"`
	Lane            int        `json:"lane"`
	Distance        int        `json:"distance"`
	Slot            string     `json:"slot"`
	FaceType        FaceType   `json:"face_type"`
	Bowstyle        Bowstyle   `json:"bowstyle"`
	DrawWeight      float64    `json:"draw_weight"`
	ClubID          *uuid.UUID `json:"club_id,omitempty"`
	IsShooting      bool       `json:"is_shooting"`
	ShotPerRound    *int       `json:"shot_per_round,omitempty"`
	IntervalSeconds int        `json:"interval_seconds"`
	CreatedAt       time.Time  `json:"created_at"`
}
