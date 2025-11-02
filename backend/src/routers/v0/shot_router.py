from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status

from models import DBNotFound, ShotModel, SlotModel
from routers.deps.auth import require_auth
from routers.deps.models import get_shot_model, get_slot_model
from schema import ShotCreate, ShotFilter, ShotId, ShotRead, SlotFilter


router = APIRouter(prefix="/shot", tags=["Shots"])


@router.post("", response_model=ShotId, status_code=status.HTTP_201_CREATED)
async def create_shot(
    shot: ShotCreate,
    current_archer_id: UUID = Depends(require_auth),
    shot_model: ShotModel = Depends(get_shot_model),
    slot_model: SlotModel = Depends(get_slot_model),
) -> ShotId:
    try:
        # Verify that the slot belongs to the archer
        slot = await slot_model.get_one(SlotFilter(slot_id=shot.slot_id))
        if slot.archer_id != current_archer_id:
            raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")

        shot_id = await shot_model.insert_one(shot)
        return ShotId(shot_id=shot_id)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e


@router.get("/by-slot/{slot_id:uuid}", response_model=list[ShotRead])
async def get_shots_by_slot(
    slot_id: UUID,
    current_archer_id: UUID = Depends(require_auth),
    shot_model: ShotModel = Depends(get_shot_model),
    slot_model: SlotModel = Depends(get_slot_model),
) -> list[ShotRead]:
    try:
        # Verify that the slot belongs to the archer
        slot = await slot_model.get_one(SlotFilter(slot_id=slot_id))
        if slot.archer_id != current_archer_id:
            raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")

        shots = await shot_model.get_all(ShotFilter(slot_id=slot_id), [])
        return shots
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
