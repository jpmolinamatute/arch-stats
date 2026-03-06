from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, status

from core import ShotManager
from routers.deps.auth import require_auth
from routers.deps.models import get_shot_manager
from schema import ShotCreate, ShotId, ShotRead

router = APIRouter(prefix="/shot", tags=["Shots"])


@router.post("", response_model=ShotId | list[ShotId], status_code=status.HTTP_201_CREATED)
async def create_shot(
    shots: ShotCreate | list[ShotCreate],
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    shot_manager: Annotated[ShotManager, Depends(get_shot_manager)],
) -> ShotId | list[ShotId]:
    return await shot_manager.create(shots, current_archer_id)


@router.get("/by-slot/{slot_id:uuid}", response_model=list[ShotRead])
async def get_shots_by_slot(
    slot_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    shot_manager: Annotated[ShotManager, Depends(get_shot_manager)],
) -> list[ShotRead]:
    return await shot_manager.get_shots_by_slot(slot_id, current_archer_id)


@router.get("/count-by-slot/{slot_id:uuid}", response_model=int)
async def get_shots_count_by_slot(
    slot_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    shot_manager: Annotated[ShotManager, Depends(get_shot_manager)],
) -> int:
    return await shot_manager.get_shots_count_by_slot(slot_id, current_archer_id)
