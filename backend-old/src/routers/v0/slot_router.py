from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, Response, status

from core import SlotManager
from routers.deps.auth import require_auth
from routers.deps.models import get_slot_manager
from schema import FullSlotInfo, SlotJoinRequest, SlotJoinResponse

router = APIRouter(prefix="/session", tags=["Slots"])


@router.get(
    "/slot/archer/{archer_id:uuid}",
    response_model=FullSlotInfo,
    status_code=status.HTTP_200_OK,
)
async def get_archer_current_slot(
    archer_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    slot_manager: Annotated[SlotManager, Depends(get_slot_manager)],
) -> FullSlotInfo:
    """
    Get the archer's current active slot assignment (open session and is_shooting = TRUE).

    Responses: 200 OK, 404 Not Found.
    """
    return await slot_manager.get_full_slot_info(
        current_archer_id=current_archer_id, archer_id=archer_id
    )


@router.get("/slot/{slot:uuid}", response_model=FullSlotInfo, status_code=status.HTTP_200_OK)
async def get_slot(
    slot: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    slot_manager: Annotated[SlotManager, Depends(get_slot_manager)],
) -> FullSlotInfo:
    """
    Get active slot assignment details (open session, is_shooting = TRUE).

    Responses: 200 OK, 404 Not Found.
    """
    return await slot_manager.get_full_slot_info(current_archer_id=current_archer_id, slot_id=slot)


@router.post(
    "/slot",
    response_model=SlotJoinResponse,
    status_code=status.HTTP_200_OK,
)
async def join_session(
    payload: SlotJoinRequest,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    slot_manager: Annotated[SlotManager, Depends(get_slot_manager)],
) -> SlotJoinResponse:
    """
    Join a session (assign archer to a slot).

    Responses: 200 OK, 400 Bad Request.
    """
    return await slot_manager.assign_archer_to_slot(payload, current_archer_id)


@router.patch(
    "/slot/re-join/{slot_id:uuid}",
    response_model=SlotJoinResponse,
    status_code=status.HTTP_200_OK,
)
async def re_join_session(
    slot_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    slot_manager: Annotated[SlotManager, Depends(get_slot_manager)],
) -> SlotJoinResponse:
    """
    Re-join a session (reassign archer to a slot).

    Responses: 200 OK, 400 Bad Request, 403 Forbidden, 422 Unprocessable Content.
    """
    return await slot_manager.re_join_session(slot_id, current_archer_id)


@router.patch("/slot/leave/{slot_id:uuid}", status_code=status.HTTP_200_OK)
async def leave_session(
    slot_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    slot_manager: Annotated[SlotManager, Depends(get_slot_manager)],
) -> Response:
    """
    Leave a session (stop archer's slot assignment).

    Responses: 200 OK, 400 Bad Request.
    """
    await slot_manager.leave_session(slot_id, current_archer_id)
    return Response(status_code=status.HTTP_200_OK)
