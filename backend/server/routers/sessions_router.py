from uuid import UUID

from fastapi import APIRouter, status, Depends

from database import SessionsDB, DBState
from database.schema import SessionsCreate, SessionsRead, SessionsUpdate, HTTPResponse
from server.routers.utils import db_response

SessionsRouter = APIRouter(prefix="/session")


async def get_sessions_db() -> SessionsDB:
    db_pool = await DBState.get_db_pool()
    return SessionsDB(db_pool)


# GET all sessions
@SessionsRouter.get("/", response_model=HTTPResponse[list[SessionsRead]])
async def get_all_sessions(
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> HTTPResponse[list[SessionsRead]]:
    """
    Retrieve all sessions.

    Returns:
        HTTPResponse: A list of all session records in the database.
    """
    return await db_response(sessions_db.get_all, status.HTTP_200_OK)


# POST a new session
@SessionsRouter.post("/", response_model=HTTPResponse[None])
async def add_session(
    session_data: SessionsCreate,
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> HTTPResponse[None]:
    """
    Create and register a new session.

    Args:
        session_data (SessionsCreate): The data required to start a new session.

    Returns:
        HTTPResponse: Success or error message. Does not return the created session.
    """
    return await db_response(sessions_db.insert_one, status.HTTP_200_OK, session_data)


# GET a specific session by ID
@SessionsRouter.get("/{session_id}", response_model=HTTPResponse[SessionsRead])
async def get_session(
    session_id: UUID,
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> HTTPResponse[SessionsRead]:
    """
    Retrieve details for a specific session by its unique ID.

    Args:
        session_id (UUID): The unique identifier of the session.

    Returns:
        HTTPResponse: The session record if found, or an error message if not found.
    """
    return await db_response(sessions_db.get_one, status.HTTP_200_OK, session_id)


# DELETE a specific session by ID
@SessionsRouter.delete("/{session_id}", response_model=HTTPResponse[None])
async def delete_session(
    session_id: UUID,
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> HTTPResponse[None]:
    """
    Delete a session by its unique ID.

    Args:
        session_id (UUID): The unique identifier of the session to delete.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(sessions_db.delete_one, status.HTTP_204_NO_CONTENT, session_id)


# PATCH (partial update) a specific session by ID
@SessionsRouter.patch("/{session_id}", response_model=HTTPResponse[None])
async def patch_session(
    session_id: UUID,
    update: SessionsUpdate,
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> HTTPResponse[None]:
    """
    Partially update an existing session's data.

    Args:
        session_id (UUID): The unique identifier of the session to update.
        update (SessionsUpdate): The fields to update in the session.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(sessions_db.delete_one, status.HTTP_202_ACCEPTED, session_id, update)
