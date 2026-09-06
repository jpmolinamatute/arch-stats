package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
	"github.com/jpmolinamatute/arch-stats/backend/internal/service"
)

type shotServiceAdapter struct {
	svc *service.ShotService
}

func newShotServiceAdapter(svc *service.ShotService) *shotServiceAdapter {
	return &shotServiceAdapter{svc: svc}
}

func (a *shotServiceAdapter) Create(ctx context.Context, shot model.ShotCreate, _ uuid.UUID) (uuid.UUID, error) {
	return a.svc.Create(ctx, shot)
}

func (a *shotServiceAdapter) CreateBatch(ctx context.Context, shots []model.ShotCreate, _ uuid.UUID) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(shots))
	for _, s := range shots {
		id, err := a.svc.Create(ctx, s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (a *shotServiceAdapter) GetBySlot(ctx context.Context, slotID, _ uuid.UUID) ([]model.ShotRead, error) {
	return a.svc.ListBySlotID(ctx, slotID)
}

func (a *shotServiceAdapter) CountBySlot(ctx context.Context, slotID, _ uuid.UUID) (int, error) {
	shots, err := a.svc.ListBySlotID(ctx, slotID)
	if err != nil {
		return 0, err
	}
	return len(shots), nil
}

type slotServiceAdapter struct {
	svc         *service.SlotService
	slotRepo    *repository.SlotRepo
	sessionRepo *repository.SessionRepo
	targetRepo  *repository.TargetRepo
}

func newSlotServiceAdapter(
	svc *service.SlotService,
	slotRepo *repository.SlotRepo,
	sessionRepo *repository.SessionRepo,
	targetRepo *repository.TargetRepo,
) *slotServiceAdapter {
	return &slotServiceAdapter{
		svc:         svc,
		slotRepo:    slotRepo,
		sessionRepo: sessionRepo,
		targetRepo:  targetRepo,
	}
}

func (a *slotServiceAdapter) GetArcherCurrentSlot(ctx context.Context, archerID uuid.UUID) (*model.FullSlotInfo, error) {
	isShooting := true
	slots, err := a.slotRepo.FindAll(ctx, model.SlotFilter{
		ArcherID:   &archerID,
		IsShooting: &isShooting,
	})
	if err != nil {
		return nil, fmt.Errorf("finding current slot: %w", err)
	}
	if len(slots) == 0 {
		return nil, apperror.ErrNotFound
	}

	slot := slots[0]
	target, err := a.targetRepo.FindByID(ctx, slot.TargetID)
	if err != nil {
		return nil, fmt.Errorf("finding target: %w", err)
	}
	if target == nil {
		return nil, apperror.ErrNotFound
	}

	return toFullSlotInfo(&slot, target), nil
}

func (a *slotServiceAdapter) GetSlot(ctx context.Context, slotID uuid.UUID) (*model.FullSlotInfo, error) {
	slot, err := a.slotRepo.FindByID(ctx, slotID)
	if err != nil {
		return nil, fmt.Errorf("finding slot: %w", err)
	}
	if slot == nil {
		return nil, apperror.ErrNotFound
	}

	target, err := a.targetRepo.FindByID(ctx, slot.TargetID)
	if err != nil {
		return nil, fmt.Errorf("finding target: %w", err)
	}
	if target == nil {
		return nil, apperror.ErrNotFound
	}

	return toFullSlotInfo(slot, target), nil
}

//nolint:gocritic // hugeParam: req matches SlotService interface specification
func (a *slotServiceAdapter) JoinSession(ctx context.Context, req model.SlotJoinRequest) (*model.SlotJoinResponse, error) {
	session, err := a.sessionRepo.FindByID(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("finding session: %w", err)
	}
	if session == nil || session.ClosedAt != nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "session either does not exist or is closed")
	}

	targets, err := a.targetRepo.FindBySessionID(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("finding session targets: %w", err)
	}

	var targetID uuid.UUID
	var lane int
	for _, t := range targets {
		if t.Distance == req.Distance {
			targetID = t.TargetID
			lane = t.Lane
			break
		}
	}

	if targetID == uuid.Nil {
		lane = len(targets) + 1
		targetID, err = a.targetRepo.Create(ctx, model.TargetCreate{
			SessionID: req.SessionID,
			Distance:  req.Distance,
			Lane:      lane,
		})
		if err != nil {
			return nil, fmt.Errorf("creating target: %w", err)
		}
	}

	letter := model.SlotLetterA
	slotID, err := a.slotRepo.Create(ctx, model.SlotCreate{
		TargetID:        targetID,
		ArcherID:        req.ArcherID,
		SessionID:       req.SessionID,
		SlotLetter:      letter,
		FaceType:        req.FaceType,
		Bowstyle:        req.Bowstyle,
		DrawWeight:      req.DrawWeight,
		ClubID:          req.ClubID,
		ShotPerRound:    req.ShotPerRound,
		IntervalSeconds: req.IntervalSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("creating slot: %w", err)
	}

	return &model.SlotJoinResponse{
		SlotID: slotID,
		Slot:   fmt.Sprintf("%d%s", lane, letter),
	}, nil
}

func (a *slotServiceAdapter) ReJoinSession(ctx context.Context, slotID, archerID uuid.UUID) (*model.SlotJoinResponse, error) {
	slot, err := a.slotRepo.FindByID(ctx, slotID)
	if err != nil {
		return nil, fmt.Errorf("finding slot: %w", err)
	}
	if slot == nil {
		return nil, apperror.ErrNotFound
	}
	if slot.ArcherID != archerID {
		return nil, apperror.ErrForbidden
	}

	isShooting := true
	if err := a.slotRepo.Update(ctx, model.SlotSet{IsShooting: &isShooting}, model.SlotFilter{SlotID: &slotID}); err != nil {
		return nil, fmt.Errorf("rejoining slot: %w", err)
	}

	target, err := a.targetRepo.FindByID(ctx, slot.TargetID)
	if err != nil || target == nil {
		return nil, apperror.ErrNotFound
	}

	return &model.SlotJoinResponse{
		SlotID: slotID,
		Slot:   fmt.Sprintf("%d%s", target.Lane, slot.SlotLetter),
	}, nil
}

func (a *slotServiceAdapter) LeaveSession(ctx context.Context, slotID, archerID uuid.UUID) error {
	slot, err := a.slotRepo.FindByID(ctx, slotID)
	if err != nil {
		return fmt.Errorf("finding slot: %w", err)
	}
	if slot == nil {
		return apperror.ErrNotFound
	}
	if slot.ArcherID != archerID {
		return apperror.ErrForbidden
	}

	isShooting := false
	if err := a.slotRepo.Update(ctx, model.SlotSet{IsShooting: &isShooting}, model.SlotFilter{SlotID: &slotID}); err != nil {
		return fmt.Errorf("leaving slot: %w", err)
	}
	return nil
}

func toFullSlotInfo(slot *model.SlotRead, target *model.TargetRead) *model.FullSlotInfo {
	var createdAt time.Time
	if slot.CreatedAt != nil {
		createdAt = *slot.CreatedAt
	}
	return &model.FullSlotInfo{
		SlotID:          slot.SlotID,
		TargetID:        slot.TargetID,
		ArcherID:        slot.ArcherID,
		SessionID:       slot.SessionID,
		SlotLetter:      slot.SlotLetter,
		Lane:            target.Lane,
		Distance:        target.Distance,
		Slot:            fmt.Sprintf("%d%s", target.Lane, slot.SlotLetter),
		FaceType:        slot.FaceType,
		Bowstyle:        slot.Bowstyle,
		DrawWeight:      slot.DrawWeight,
		ClubID:          slot.ClubID,
		IsShooting:      slot.IsShooting,
		ShotPerRound:    slot.ShotPerRound,
		IntervalSeconds: slot.IntervalSeconds,
		CreatedAt:       createdAt,
	}
}
