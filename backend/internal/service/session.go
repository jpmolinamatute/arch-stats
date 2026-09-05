package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// SessionRepository defines persistence operations required by SessionService.
type SessionRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error)
	FindOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error)
	FindAll(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error)
	Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.SessionSet, filter model.SessionFilter) error
	Close(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SessionService encapsulates business logic and lifecycle rules for shooting sessions.
type SessionService struct {
	repo SessionRepository
}

// NewSessionService constructs a SessionService with repository dependency injection.
func NewSessionService(repo SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

// GetByID retrieves a shooting session by primary key identifier.
// Returns apperror.ErrNotFound if the session does not exist.
func (s *SessionService) GetByID(ctx context.Context, id uuid.UUID) (*model.SessionRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	session, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching session: %w", err)
	}
	if session == nil {
		return nil, apperror.ErrNotFound
	}
	return session, nil
}

// GetOpen retrieves the active open session owned by the archer.
// Returns apperror.ErrNotFound if no open session exists.
func (s *SessionService) GetOpen(ctx context.Context, archerID uuid.UUID) (*model.SessionRead, error) {
	if archerID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "archer id is required")
	}

	session, err := s.repo.FindOpen(ctx, archerID)
	if err != nil {
		return nil, fmt.Errorf("fetching open session: %w", err)
	}
	if session == nil {
		return nil, apperror.ErrNotFound
	}
	return session, nil
}

// List retrieves all shooting sessions matching the specified filter criteria.
func (s *SessionService) List(ctx context.Context, filter model.SessionFilter) ([]model.SessionRead, error) {
	sessions, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	if sessions == nil {
		return []model.SessionRead{}, nil
	}
	return sessions, nil
}

// Create validates session data and ensures the owner does not already have an open session before creating.
// Returns apperror.ErrConflict if an open session already exists for the archer.
func (s *SessionService) Create(ctx context.Context, data model.SessionCreate) (uuid.UUID, error) {
	if data.OwnerArcherID == uuid.Nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "owner_archer_id is required")
	}
	if strings.TrimSpace(data.SessionLocation) == "" {
		return uuid.Nil, apperror.Wrap(apperror.ErrValidation, "session_location is required")
	}

	openSession, err := s.repo.FindOpen(ctx, data.OwnerArcherID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("checking for active open session: %w", err)
	}
	if openSession != nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrConflict, "archer already has an open session")
	}

	data.IsOpened = true
	id, err := s.repo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating session: %w", err)
	}
	return id, nil
}

// Close closes an active shooting session after validating that it exists and is currently open.
// Returns apperror.ErrNotFound if missing, or apperror.ErrValidation if already closed.
func (s *SessionService) Close(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	session, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying session: %w", err)
	}
	if session == nil {
		return apperror.ErrNotFound
	}
	if !session.IsOpened {
		return apperror.Wrap(apperror.ErrValidation, "session is already closed")
	}

	if err := s.repo.Close(ctx, id); err != nil {
		return fmt.Errorf("closing session: %w", err)
	}
	return nil
}

// ReOpen re-opens a closed shooting session after verifying it exists, is closed, and
// the archer does not currently have another open session.
// Returns apperror.ErrNotFound if missing, apperror.ErrValidation if already open,
// or apperror.ErrConflict if another open session exists for the archer.
func (s *SessionService) ReOpen(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	session, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying session: %w", err)
	}
	if session == nil {
		return apperror.ErrNotFound
	}
	if session.IsOpened {
		return apperror.Wrap(apperror.ErrValidation, "session is already open")
	}

	openSession, err := s.repo.FindOpen(ctx, session.OwnerArcherID)
	if err != nil {
		return fmt.Errorf("checking for conflicting open session: %w", err)
	}
	if openSession != nil {
		return apperror.Wrap(apperror.ErrConflict, "archer already has an open session")
	}

	isOpened := true
	updateData := model.SessionSet{
		IsOpened: &isOpened,
	}
	filter := model.SessionFilter{
		SessionID: &id,
	}
	if err := s.repo.Update(ctx, updateData, filter); err != nil {
		return fmt.Errorf("reopening session: %w", err)
	}
	return nil
}

// Delete removes a session by primary key identifier.
// Returns apperror.ErrNotFound if the session does not exist.
func (s *SessionService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}
