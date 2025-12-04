from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status

from models import DBNotFound, ShotModel, SlotModel
from routers.deps.auth import require_auth
from routers.deps.models import get_shot_model, get_slot_model
from schema import ShotCreate, ShotFilter, ShotId, ShotRead, SlotFilter


router = APIRouter(prefix="/shot", tags=["Shots"])


@router.post("", response_model=ShotId | list[ShotId], status_code=status.HTTP_201_CREATED)
async def create_shot(
    shots: ShotCreate | list[ShotCreate],
    current_archer_id: UUID = Depends(require_auth),
    shot_model: ShotModel = Depends(get_shot_model),
    slot_model: SlotModel = Depends(get_slot_model),
) -> ShotId | list[ShotId]:
    try:
        result: ShotId | list[ShotId]
        if isinstance(shots, ShotCreate):
            # Verify that the slot belongs to the archer
            slot = await slot_model.get_one(SlotFilter(slot_id=shots.slot_id))
            if slot.archer_id != current_archer_id:
                raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")

            shot_id = await shot_model.insert_one(shots)
            result = ShotId(shot_id=shot_id)
        elif isinstance(shots, list) and len(shots) > 0:
            slot_ids = {shot.slot_id for shot in shots}
            if len(slot_ids) != 1:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="All shots must belong to the same slot",
                )

            slot_id = slot_ids.pop()
            slot = await slot_model.get_one(SlotFilter(slot_id=slot_id))
            if slot.archer_id != current_archer_id:
                raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")

            shot_ids = await shot_model.insert_many(shots)
            result = [ShotId(shot_id=shot_id) for shot_id in shot_ids]
        else:
            raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Invalid input")
        return result
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
