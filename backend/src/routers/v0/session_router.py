from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, status

from core import SessionManager
from routers.deps.auth import require_auth
from routers.deps.models import get_session_manager
from schema import SessionCreate, SessionId, SessionRead

router = APIRouter(prefix="/session", tags=["Sessions"])


@router.get(
    "/archer/{archer_id:uuid}/open-session",
    response_model=SessionId,
    status_code=status.HTTP_200_OK,
)
async def get_open_session_for_archer(
    archer_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    session_manager: Annotated[SessionManager, Depends(get_session_manager)],
) -> SessionId:
    """
    Get the open session ID owned by the archer.
    Responses: 200 OK
    """
    return await session_manager.get_open_session_for_archer(current_archer_id, archer_id)


@router.get(
    "/archer/{archer_id:uuid}/close-session",
    response_model=list[SessionRead],
    status_code=status.HTTP_200_OK,
)
async def get_closed_session_for_archer(
    archer_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    session_manager: Annotated[SessionManager, Depends(get_session_manager)],
) -> list[SessionRead]:
    """
    Get the closed session ID owned by the archer.
    Responses: 200 OK
    """
    return await session_manager.get_closed_session_for_archer(current_archer_id, archer_id)


@router.get(
    "/archer/{archer_id:uuid}/participating",
    response_model=SessionId,
    status_code=status.HTTP_200_OK,
)
async def get_participating_session_for_archer(
    archer_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    session_manager: Annotated[SessionManager, Depends(get_session_manager)],
) -> SessionId:
    """
    Get the open session ID of the open session an archer is currently participating in.
    Responses: 200 OK, 422 Unprocessable Entity.
    """
    return await session_manager.get_participating_session_for_archer(current_archer_id, archer_id)


@router.get("/open", response_model=list[SessionRead], status_code=status.HTTP_200_OK)
async def get_all_open_sessions(
    session_manager: Annotated[SessionManager, Depends(get_session_manager)],
) -> list[SessionRead]:
    """
    List all open sessions.

    Responses: 200 OK.
    """
    return await session_manager.get_all_open_sessions()


@router.post("", response_model=SessionId, status_code=status.HTTP_201_CREATED)
async def create_session(
    session_data: SessionCreate,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    session_manager: Annotated[SessionManager, Depends(get_session_manager)],
) -> SessionId:
    """
    Create a session and initial slot assignment.

    Responses: 201 Created, 422 Unprocessable Content.
    """
    return await session_manager.create_session(session_data, current_archer_id)


@router.get("/{session:uuid}", response_model=SessionRead, status_code=status.HTTP_200_OK)
async def get_session(
    session: UUID,
    session_manager: Annotated[SessionManager, Depends(get_session_manager)],
) -> SessionRead:
    """
    Get session details.

    Responses: 200 OK, 404 Not Found.
    """
    return await session_manager.get_session(session)


@router.patch("/re-open", response_model=SessionId, status_code=status.HTTP_200_OK)
async def re_open_session(
    session: SessionId,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    session_manager: Annotated[SessionManager, Depends(get_session_manager)],
) -> SessionId:
    """
    Re-open a closed session.

    Responses: 200 OK, 404 Not Found, 422 Unprocessable Content.
    """
    return await session_manager.re_open_session(session, current_archer_id)


@router.patch("/close", response_model=dict[str, str], status_code=status.HTTP_200_OK)
async def close_session(
    session: SessionId,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    session_manager: Annotated[SessionManager, Depends(get_session_manager)],
) -> dict[str, str]:
    """
    Close a session.

    Responses: 200 OK, 404 Not Found, 422 Unprocessable Content.
    """
    return await session_manager.close_session(session, current_archer_id)
