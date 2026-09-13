package model

import (
	"time"

	"github.com/google/uuid"
)

// ArrowCreate represents the payload required to register an individual arrow.
type ArrowCreate struct {
	ArcherID    uuid.UUID    `json:"archer_id"`
	ArrowSet    int16        `json:"arrow_set" validate:"required,gt=0"`
	ArrowNumber int16        `json:"arrow_number" validate:"required,gt=0"`
	Status      *ArrowStatus `json:"status,omitempty"`
	Spine       *float64     `json:"spine,omitempty" validate:"omitempty,gt=0"`
	Length      *float64     `json:"length,omitempty" validate:"omitempty,gt=0"`
	Weight      *float64     `json:"weight,omitempty" validate:"omitempty,gt=0"`
}

// ArrowBatchCreate represents the payload required to register an entire batch of arrows.
type ArrowBatchCreate struct {
	ArcherID uuid.UUID    `json:"archer_id"`
	ArrowSet int16        `json:"arrow_set" validate:"required,gt=0"`
	Count    int          `json:"count" validate:"required,gte=3"`
	Status   *ArrowStatus `json:"status,omitempty"`
	Spine    *float64     `json:"spine,omitempty" validate:"omitempty,gt=0"`
	Length   *float64     `json:"length,omitempty" validate:"omitempty,gt=0"`
	Weight   *float64     `json:"weight,omitempty" validate:"omitempty,gt=0"`
}

// ArrowRead represents the full persisted arrow domain model.
type ArrowRead struct {
	ArrowID     uuid.UUID   `json:"arrow_id"`
	ArcherID    uuid.UUID   `json:"archer_id"`
	ArrowSet    int16       `json:"arrow_set"`
	ArrowNumber int16       `json:"arrow_number"`
	Status      ArrowStatus `json:"status"`
	Spine       *float64    `json:"spine,omitempty"`
	Length      *float64    `json:"length,omitempty"`
	Weight      *float64    `json:"weight,omitempty"`
	IsDeleted   bool        `json:"is_deleted"`
	CreatedAt   time.Time   `json:"created_at"`
}

// ArrowSet represents mutable fields when updating arrow specifications or status.
type ArrowSet struct {
	Status *ArrowStatus `json:"status,omitempty"`
	Spine  *float64     `json:"spine,omitempty" validate:"omitempty,gt=0"`
	Length *float64     `json:"length,omitempty" validate:"omitempty,gt=0"`
	Weight *float64     `json:"weight,omitempty" validate:"omitempty,gt=0"`
}

// ArrowFilter represents criteria to query or filter arrows.
type ArrowFilter struct {
	ArrowID   *uuid.UUID   `json:"arrow_id,omitempty"`
	ArcherID  *uuid.UUID   `json:"archer_id,omitempty"`
	ArrowSet  *int16       `json:"arrow_set,omitempty"`
	Status    *ArrowStatus `json:"status,omitempty"`
	IsDeleted *bool        `json:"is_deleted,omitempty"`
}

// ArrowUpdate wraps target filter criteria and field updates for arrows.
type ArrowUpdate struct {
	Where ArrowFilter `json:"where"`
	Data  ArrowSet    `json:"data"`
}
