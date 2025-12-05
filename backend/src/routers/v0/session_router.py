from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status

from models import DBException, DBNotFound, SessionModel
from routers.deps.auth import require_auth
from routers.deps.models import get_session_model
from schema import SessionCreate, SessionFilter, SessionId, SessionRead


router = APIRouter(prefix="/session", tags=["Sessions"])


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


@router.patch("/close", response_model=dict[str, str], status_code=status.HTTP_200_OK)
async def close_session(
    session: SessionId,
    session_model: SessionModel = Depends(get_session_model),
) -> dict[str, str]:
    """
    Close a session.

    Responses: 200 OK, 404 Not Found, 422 Unprocessable Content.
    """
    try:
        await session_model.close_session(session)
        return {"status": "closed"}
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
