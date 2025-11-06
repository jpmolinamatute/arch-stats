from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Response, status

from core import SlotManager, SlotManagerError
from models import DBException, DBNotFound, SlotModel
from routers.deps.auth import require_auth
from routers.deps.models import get_slot_manager, get_slot_model
from schema import FullSlotInfo, SlotFilter, SlotJoinRequest, SlotJoinResponse


router = APIRouter(prefix="/session", tags=["Slots"])


@router.get(
    "/slot/archer/{archer_id:uuid}",
    response_model=FullSlotInfo,
    status_code=status.HTTP_200_OK,
)
async def get_archer_current_slot(
    archer_id: UUID,
    current_archer_id: UUID = Depends(require_auth),
    slot_model: SlotModel = Depends(get_slot_model),
) -> FullSlotInfo:
    """
    Get the archer's current active slot assignment (open session and is_shooting = TRUE).

    Responses: 200 OK, 404 Not Found.
    """
    if current_archer_id != archer_id:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")

    try:
        return await slot_model.get_full_slot_info(archer_id=archer_id)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e


@router.get("/slot/{slot:uuid}", response_model=FullSlotInfo, status_code=status.HTTP_200_OK)
async def get_slot(
    slot: UUID,
    current_archer_id: UUID = Depends(require_auth),
    slot_model: SlotModel = Depends(get_slot_model),
) -> FullSlotInfo:
    """
    Get active slot assignment details (open session, is_shooting = TRUE).

    Responses: 200 OK, 404 Not Found.
    """
    try:
        slot_row = await slot_model.get_full_slot_info(slot_id=slot)
        if current_archer_id != slot_row.archer_id:
            raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")
        return slot_row
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e


@router.post(
    "/slot",
    response_model=SlotJoinResponse,
    status_code=status.HTTP_200_OK,
)
async def join_session(
    payload: SlotJoinRequest,
    slot_manager: SlotManager = Depends(get_slot_manager),
) -> SlotJoinResponse:
    """
    Join a session (assign archer to a slot).

    Responses: 200 OK, 400 Bad Request.
    """
    try:
        return await slot_manager.assign_archer_to_slot(payload)
    except DBNotFound as e:
        # Non-existent or closed session should be treated as unprocessable (422)
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=str(e)) from e
    except SlotManagerError as e:
        msg = str(e)
        if msg in (
            "ERROR: archer already participating in an open session",
            "ERROR: archer already joined this session",
        ):
            raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=msg) from e
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=msg) from e


@router.patch(
    "/slot/re-join/{slot_id:uuid}", response_model=SlotJoinResponse, status_code=status.HTTP_200_OK
)
async def re_join_session(
    slot_id: UUID,
    current_archer_id: UUID = Depends(require_auth),
    slot_manager: SlotManager = Depends(get_slot_manager),
    slot_model: SlotModel = Depends(get_slot_model),
) -> SlotJoinResponse:
    """
    Re-join a session (reassign archer to a slot).

    Responses: 200 OK, 400 Bad Request, 403 Forbidden, 422 Unprocessable Content.
    """
    try:
        # Authorization: ensure the slot belongs to the current user
        slot_row = await slot_model.get_one(SlotFilter(slot_id=slot_id))
        if slot_row.archer_id != current_archer_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="ERROR: user not allowed to re-join",
            )
        # Delegate state checks and reactivation to SlotManager
        return await slot_manager.re_join_session(slot_row)
    except DBNotFound as e:
        # Pass through specific messages from SlotManager ensuring 422
        msg = str(e)
        if msg == "ERROR: The session doesn't exist or it was already closed":
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg
            ) from e
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
            detail="ERROR: the archer is either not allowed to re-join or they are already in",
        ) from e
    except DBException as e:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e)) from e
    except SlotManagerError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e)) from e


@router.patch("/slot/leave/{slot_id:uuid}", status_code=status.HTTP_200_OK)
async def leave_session(
    slot_id: UUID,
    current_archer_id: UUID = Depends(require_auth),
    slot_manager: SlotManager = Depends(get_slot_manager),
    slot_model: SlotModel = Depends(get_slot_model),
) -> Response:
    """
    Leave a session (stop archer's slot assignment).

    Responses: 200 OK, 400 Bad Request.
    """
    try:
        # Fetch slot to check ownership
        slot_row = await slot_model.get_one(SlotFilter(slot_id=slot_id))
        if slot_row.archer_id != current_archer_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN, detail="ERROR: user not allowed to leave"
            )

        # Delegate active-state and session-open checks to SlotManager
        await slot_manager.leave_session(slot_row)
        return Response(status_code=status.HTTP_200_OK)
    except DBNotFound as e:
        msg = str(e)
        if msg == "ERROR: Session either doesn't exist or it was already closed":
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg
            ) from e
        # Archer wasn't participating (slot not active or not found)
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="ERROR: archer is not participating in this session",
        ) from e
