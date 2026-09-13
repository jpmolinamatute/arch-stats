package model

import (
	"time"

	"github.com/google/uuid"
)

// BowCreate represents the payload required to register a new bow.
type BowCreate struct {
	ArcherID   uuid.UUID `json:"archer_id"`
	Name       string    `json:"name" validate:"required,min=1,max=255"`
	Bowstyle   Bowstyle  `json:"bowstyle" validate:"required"`
	DrawWeight float64   `json:"draw_weight" validate:"required,gt=0,lte=200"`
}

// BowRead represents the full persisted bow domain model.
type BowRead struct {
	BowID      uuid.UUID `json:"bow_id"`
	ArcherID   uuid.UUID `json:"archer_id"`
	Name       string    `json:"name"`
	Bowstyle   Bowstyle  `json:"bowstyle"`
	DrawWeight float64   `json:"draw_weight"`
	IsDeleted  bool      `json:"is_deleted"`
	CreatedAt  time.Time `json:"created_at"`
}

// BowSet represents mutable fields when updating bow specifications.
type BowSet struct {
	Name       *string   `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Bowstyle   *Bowstyle `json:"bowstyle,omitempty"`
	DrawWeight *float64  `json:"draw_weight,omitempty" validate:"omitempty,gt=0,lte=200"`
}

// BowFilter represents criteria to query or filter bows.
type BowFilter struct {
	BowID     *uuid.UUID `json:"bow_id,omitempty"`
	ArcherID  *uuid.UUID `json:"archer_id,omitempty"`
	Bowstyle  *Bowstyle  `json:"bowstyle,omitempty"`
	IsDeleted *bool      `json:"is_deleted,omitempty"`
}

// BowUpdate wraps target filter criteria and field updates for bows.
type BowUpdate struct {
	Where BowFilter `json:"where"`
	Data  BowSet    `json:"data"`
}
