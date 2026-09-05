package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// SlotRepository defines persistence operations required by SlotService.
type SlotRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error)
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error)
	FindAll(ctx context.Context, filter model.SlotFilter) ([]model.SlotRead, error)
	Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.SlotSet, filter model.SlotFilter) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SlotService encapsulates business logic, state rules, and validation for slot assignments.
type SlotService struct {
	slotRepo    SlotRepository
	sessionRepo SessionRepository
}

// NewSlotService constructs a SlotService with repository dependency injection.
func NewSlotService(slotRepo SlotRepository, sessionRepo SessionRepository) *SlotService {
	return &SlotService{
		slotRepo:    slotRepo,
		sessionRepo: sessionRepo,
	}
}

// GetByID retrieves a slot assignment by primary key identifier.
// Returns apperror.ErrNotFound if the slot does not exist.
func (s *SlotService) GetByID(ctx context.Context, id uuid.UUID) (*model.SlotRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	slot, err := s.slotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching slot: %w", err)
	}
	if slot == nil {
		return nil, apperror.ErrNotFound
	}
	return slot, nil
}

// ListBySessionID retrieves all slot assignments belonging to a given session.
func (s *SlotService) ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.SlotRead, error) {
	if sessionID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	slots, err := s.slotRepo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("listing slots by session: %w", err)
	}
	if slots == nil {
		return []model.SlotRead{}, nil
	}
	return slots, nil
}

// Create validates incoming slot assignment data, verifies that the target session exists and is open,
// and persists the slot record.
// Returns apperror.ErrNotFound if the session does not exist, or apperror.ErrValidation if the session is closed.
//
//nolint:gocritic // hugeParam: data value parameter matches domain model parameter specification
func (s *SlotService) Create(ctx context.Context, data model.SlotCreate) (uuid.UUID, error) {
	if err := validateSlotCreate(data); err != nil {
		return uuid.Nil, err
	}

	session, err := s.sessionRepo.FindByID(ctx, data.SessionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("verifying session: %w", err)
	}
	if session == nil {
		return uuid.Nil, apperror.ErrNotFound
	}
	if !session.IsOpened {
		return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "session is not open")
	}

	id, err := s.slotRepo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating slot: %w", err)
	}
	return id, nil
}

// Update validates mutation fields and updates a slot assignment.
// Returns apperror.ErrNotFound if the slot does not exist.
func (s *SlotService) Update(ctx context.Context, id uuid.UUID, data model.SlotSet) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	if err := validateSlotSet(data); err != nil {
		return err
	}

	existing, err := s.slotRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying slot: %w", err)
	}
	if existing == nil {
		return apperror.ErrNotFound
	}

	filter := model.SlotFilter{SlotID: &id}
	if err := s.slotRepo.Update(ctx, data, filter); err != nil {
		return fmt.Errorf("updating slot: %w", err)
	}
	return nil
}

// Delete removes a slot assignment by primary key identifier.
// Returns apperror.ErrNotFound if the slot does not exist.
func (s *SlotService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	if err := s.slotRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("deleting slot: %w", err)
	}
	return nil
}

//nolint:gocritic // hugeParam: data value parameter matches domain model parameter specification
func validateSlotCreate(data model.SlotCreate) error {
	if data.ArcherID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "archer_id is required")
	}
	if data.SessionID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session_id is required")
	}
	if data.TargetID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "target_id is required")
	}
	if !isValidSlotLetter(data.SlotLetter) {
		return apperror.Wrap(apperror.ErrValidation, "invalid slot_letter")
	}
	if !isValidFaceType(data.FaceType) {
		return apperror.Wrap(apperror.ErrValidation, "invalid face_type")
	}
	if !isValidBowstyle(data.Bowstyle) {
		return apperror.Wrap(apperror.ErrValidation, "invalid bowstyle")
	}
	if data.DrawWeight <= 0 || data.DrawWeight > 200 {
		return apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}
	if data.IntervalSeconds < 1 || data.IntervalSeconds > 100 {
		return apperror.Wrap(apperror.ErrValidation, "interval_seconds must be between 1 and 100")
	}
	if data.ShotPerRound != nil && (*data.ShotPerRound < 3 || *data.ShotPerRound > 10) {
		return apperror.Wrap(apperror.ErrValidation, "shot_per_round must be between 3 and 10")
	}
	return nil
}

func validateSlotSet(data model.SlotSet) error {
	if data.FaceType != nil && !isValidFaceType(*data.FaceType) {
		return apperror.Wrap(apperror.ErrValidation, "invalid face_type")
	}
	if data.SlotLetter != nil && !isValidSlotLetter(*data.SlotLetter) {
		return apperror.Wrap(apperror.ErrValidation, "invalid slot_letter")
	}
	if data.ShotPerRound != nil && (*data.ShotPerRound < 3 || *data.ShotPerRound > 10) {
		return apperror.Wrap(apperror.ErrValidation, "shot_per_round must be between 3 and 10")
	}
	if data.IntervalSeconds != nil && (*data.IntervalSeconds < 1 || *data.IntervalSeconds > 100) {
		return apperror.Wrap(apperror.ErrValidation, "interval_seconds must be between 1 and 100")
	}
	return nil
}

func isValidSlotLetter(l model.SlotLetter) bool {
	switch l {
	case model.SlotLetterA, model.SlotLetterB, model.SlotLetterC, model.SlotLetterD:
		return true
	default:
		return false
	}
}

func isValidFaceType(f model.FaceType) bool {
	switch f {
	case model.FaceTypeWA40Full,
		model.FaceTypeWA60Full,
		model.FaceTypeWA80Full,
		model.FaceTypeWA122Full,
		model.FaceTypeWA406Rings,
		model.FaceTypeWA606Rings,
		model.FaceTypeWA806Rings,
		model.FaceTypeWA1226Rings,
		model.FaceTypeWA40TripleVertical,
		model.FaceTypeWA60TripleTriangular,
		model.FaceTypeNone:
		return true
	default:
		return false
	}
}
