from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status

from core import ShotManager
from models import DBNotFound, ShotModel
from routers.deps.auth import require_auth
from routers.deps.models import get_shot_manager, get_shot_model
from schema import ShotCreate, ShotFilter, ShotId, ShotRead

router = APIRouter(prefix="/shot", tags=["Shots"])


@router.post("", response_model=ShotId | list[ShotId], status_code=status.HTTP_201_CREATED)
async def create_shot(
    shots: ShotCreate | list[ShotCreate],
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    shot_manager: Annotated[ShotManager, Depends(get_shot_manager)],
) -> ShotId | list[ShotId]:
    try:
        return await shot_manager.create(shots, current_archer_id)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e


@router.get("/by-slot/{slot_id:uuid}", response_model=list[ShotRead])
async def get_shots_by_slot(
    slot_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    shot_model: Annotated[ShotModel, Depends(get_shot_model)],
    shot_manager: Annotated[ShotManager, Depends(get_shot_manager)],
) -> list[ShotRead]:
    try:
        await shot_manager.verify_slot_ownership(slot_id, current_archer_id)
        return await shot_model.get_all(ShotFilter(slot_id=slot_id), [])
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e


@router.get("/count-by-slot/{slot_id:uuid}", response_model=int)
async def get_shots_count_by_slot(
    slot_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    shot_model: Annotated[ShotModel, Depends(get_shot_model)],
    shot_manager: Annotated[ShotManager, Depends(get_shot_manager)],
) -> int:
    try:
        await shot_manager.verify_slot_ownership(slot_id, current_archer_id)
        return await shot_model.count_by_slot(slot_id)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
