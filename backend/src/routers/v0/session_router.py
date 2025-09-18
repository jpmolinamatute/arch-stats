import logging
from uuid import UUID

from asyncpg import Pool
from fastapi import APIRouter, Depends, HTTPException, Request, Response, status

from core import SlotManager, SlotManagerError, decode_token
from models import DBException, DBNotFound, SessionModel, SlotModel
from schema import (
    SessionCreate,
    SessionFilter,
    SessionId,
    SessionRead,
    SlotFilter,
    SlotId,
    SlotJoinRequest,
    SlotJoinResponse,
    SlotLeaveRequest,
    SlotRead,
)


router = APIRouter(prefix="/session", tags=["Sessions"])


async def require_auth(request: Request) -> UUID:
    """Validate JWT from auth cookie and return authenticated archer id.

    Raises 401 if the cookie is missing or invalid.
    """
    token = request.cookies.get("arch_stats_auth")
    if not token:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User is not authorized to use this endpoint",
        )
    try:
        sub = decode_token(token, "sub")
        if not isinstance(sub, str):
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="User is not authorized to use this endpoint",
            )
        return UUID(sub)
    except Exception as exc:  # broad to normalize to 401
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User is not authorized to use this endpoint",
        ) from exc


async def get_session_model(request: Request) -> SessionModel:
    """Dependency provider returning a `SessionModel` bound to the pool."""
    await require_auth(request)
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SessionModel")
    db_pool: Pool = request.app.state.db_pool
    return SessionModel(db_pool)


async def get_slot_model(request: Request) -> SlotModel:
    """Dependency provider returning a `SlotModel` bound to the pool."""
    await require_auth(request)
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SlotModel")
    db_pool: Pool = request.app.state.db_pool
    return SlotModel(db_pool)


async def get_slot_manager(request: Request) -> SlotManager:
    """Dependency provider returning a `SlotManager` for multi-step ops."""
    await require_auth(request)
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SlotManager")
    db_pool: Pool = request.app.state.db_pool
    return SlotManager(db_pool)


@router.get(
    "/archer/{archer_id:uuid}/open-session",
    response_model=SessionId,
    status_code=status.HTTP_200_OK,
)
async def get_open_session_for_archer(
    archer_id: UUID,
    current_archer_id: UUID = Depends(require_auth),
    session_model: SessionModel = Depends(get_session_model),
) -> SessionId:
    """
    Get the open session ID owned by the archer.
    Responses: 200 OK
    """
    # Enforce that the caller is the same archer
    if current_archer_id != archer_id:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")
    session_id = await session_model.archer_open_session(archer_id)
    return SessionId(session_id=session_id)


@router.get(
    "/archer/{archer_id:uuid}/close-session",
    response_model=list[SessionRead],
    status_code=status.HTTP_200_OK,
)
async def get_closed_session_for_archer(
    archer_id: UUID,
    current_archer_id: UUID = Depends(require_auth),
    session_model: SessionModel = Depends(get_session_model),
) -> list[SessionRead]:
    """
    Get the closed session ID owned by the archer.
    Responses: 200 OK
    """
    # Enforce that the caller is the same archer
    if current_archer_id != archer_id:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")
    return await session_model.get_all_closed_sessions_owned_by(archer_id)


@router.get(
    "/archer/{archer_id:uuid}/participating",
    response_model=SessionId,
    status_code=status.HTTP_200_OK,
)
async def get_participating_session_for_archer(
    archer_id: UUID,
    current_archer_id: UUID = Depends(require_auth),
    session_model: SessionModel = Depends(get_session_model),
) -> SessionId:
    """
    Get the open session ID of the open session an archer is currently participating in.
    Responses: 200 OK, 422 Unprocessable Entity.
    """
    # Enforce that the caller is the same archer
    if current_archer_id != archer_id:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")
    # Only raise 403 for identity mismatch above; for unexpected errors, surface 500.
    session_id = await session_model.is_archer_participating(archer_id)
    return SessionId(session_id=session_id)


@router.get("/open", response_model=list[SessionRead], status_code=status.HTTP_200_OK)
async def get_all_open_sessions(
    session_model: SessionModel = Depends(get_session_model),
) -> list[SessionRead]:
    """
    List all open sessions.

    Responses: 200 OK.
    """
    try:
        sessions = await session_model.get_all_open_sessions()
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(e)) from e
    return sessions


@router.post("", response_model=SessionId, status_code=status.HTTP_201_CREATED)
async def create_session(
    session_data: SessionCreate,
    current_archer_id: UUID = Depends(require_auth),
    session_model: SessionModel = Depends(get_session_model),
) -> SessionId:
    """
    Create a session and initial slot assignment.

    Responses: 201 Created, 422 Unprocessable Content.
    """
    try:
        # Identity check: only the authenticated archer can open a session for themselves
        if current_archer_id != session_data.owner_archer_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="ERROR: user not allowed to open a session for another archer",
            )

        session_id = await session_model.create_session(session_data)
        return SessionId(session_id=session_id)
    except ValueError as e:
        msg = str(e)
        # Conflict when owner already has an open session
        if msg == "Archer already have an opened session":
            raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=msg) from e
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg) from e


@router.get("/{session:uuid}", response_model=SessionRead, status_code=status.HTTP_200_OK)
async def get_session(
    session: UUID,
    session_model: SessionModel = Depends(get_session_model),
) -> SessionRead:
    """
    Get session details.

    Responses: 200 OK, 404 Not Found.
    """
    try:
        return await session_model.get_one(SessionFilter(session_id=session))
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e


@router.patch("/re-open", response_model=SessionId, status_code=status.HTTP_200_OK)
async def re_open_session(
    session: SessionId,
    current_archer_id: UUID = Depends(require_auth),
    session_model: SessionModel = Depends(get_session_model),
) -> SessionId:
    """
    Re-open a closed session.

    Responses: 200 OK, 404 Not Found, 422 Unprocessable Content.
    """
    try:
        await session_model.re_open_session(session, current_archer_id)
        return session
    except DBNotFound as e:
        # Standardize to 422 for non-existent or already-open sessions
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
            detail="ERROR: Session either doesn't exist or it was already open",
        ) from e
    except DBException as e:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e)) from e
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=str(e)) from e


@router.patch("/close", response_model=None, status_code=status.HTTP_204_NO_CONTENT)
async def close_session(
    session: SessionId,
    session_model: SessionModel = Depends(get_session_model),
) -> Response:
    """
    Close a session.

    Responses: 204 No Content, 404 Not Found, 422 Unprocessable Content.
    """
    try:
        await session_model.close_session(session)
        return Response(status_code=status.HTTP_204_NO_CONTENT)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
    except ValueError as e:
        msg = str(e)
        if msg == "ERROR: session_id wasn't provided":
            raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=msg) from e
        if msg == "ERROR: cannot close session with active participants":
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg
            ) from e
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg) from e


@router.post(
    "/slot",
    response_model=SlotId,
    status_code=status.HTTP_200_OK,
)
async def join_session(
    payload: SlotJoinRequest,
    slot_manager: SlotManager = Depends(get_slot_manager),
) -> SlotId:
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


@router.get(
    "/{session_id:uuid}/archer/{archer_id:uuid}/current-slot",
    response_model=SlotRead,
    status_code=status.HTTP_200_OK,
)
async def get_archer_current_slot(
    session_id: UUID,
    archer_id: UUID,
    current_archer_id: UUID = Depends(require_auth),
    slot_model: SlotModel = Depends(get_slot_model),
) -> SlotRead:
    """
    Get the archer's slot assignment for a given session regardless if is_shooting.

    Responses: 200 OK, 404 Not Found.
    """
    # Enforce that the caller is the same archer
    if current_archer_id != archer_id:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")

    try:
        where = SlotFilter(session_id=session_id, archer_id=archer_id)
        return await slot_model.get_one(where)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e


@router.patch("/slot/re-join", status_code=status.HTTP_200_OK)
async def re_join_session(
    session_req: SlotJoinRequest,
    current_archer_id: UUID = Depends(require_auth),
    slot_manager: SlotManager = Depends(get_slot_manager),
) -> SlotJoinResponse:
    """
    Re-join a session (reassign archer to a slot).

    Responses: 200 OK, 400 Bad Request, 403 Forbidden.
    """
    try:
        exists = await slot_manager.session.does_session_exist(session_req.session_id)
        if not exists:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
                detail="ERROR: the archer is either not allowed to re-join or they are already in",
            )
        return await slot_manager.re_join_session(session_req, current_archer_id)
    except DBNotFound as e:
        # Standardize to 422 for non-existent or archer already rejoined
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
            detail="ERROR: the archer is either not allowed to re-join or they are already in",
        ) from e
    except DBException as e:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e)) from e
    except SlotManagerError as e:
        # If the current user is attempting to re-join on behalf of another archer
        # the SlotManager returns a specific error. Map that to 403.
        if str(e) == "ERROR: user not allowed to re-join":
            raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e)) from e
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e)) from e


@router.patch("/slot/leave", status_code=status.HTTP_200_OK)
async def leave_session(
    session_req: SlotLeaveRequest,
    current_archer_id: UUID = Depends(require_auth),
    slot_model: SlotModel = Depends(get_slot_model),
    session_model: SessionModel = Depends(get_session_model),
) -> Response:
    """
    Leave a session (stop archer's slot assignment).

    Responses: 200 OK, 400 Bad Request.
    """
    try:
        # Identity check: only the authenticated archer can leave for themselves
        if current_archer_id != session_req.archer_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN, detail="ERROR: user not allowed to leave"
            )
        # Validate that session exists and is open
        exists = await session_model.does_session_exist(session_req.session_id)
        if not exists:
            raise DBNotFound("ERROR: Session either doesn't exist or it was already closed")
        try:
            await slot_model.leave_session(session_req)
        except DBNotFound as _:
            # Archer wasn't participating in this session
            raise HTTPException(
                status_code=status.HTTP_409_CONFLICT,
                detail="ERROR: archer is not participating in this session",
            ) from _
        return Response(status_code=status.HTTP_200_OK)
    except DBNotFound as e:
        # Normalize to 422 for non-existent/closed sessions
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=str(e)) from e
    except HTTPException as e:  # rethrow plain to avoid nesting prefix like '400: ...'
        raise e
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e)) from e
