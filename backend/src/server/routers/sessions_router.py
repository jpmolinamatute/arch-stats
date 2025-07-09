import logging
from uuid import UUID

from asyncpg import Pool
from fastapi import APIRouter, Depends, Request, status
from fastapi.responses import JSONResponse

from server.models import SessionsDB
from server.routers.utils import HTTPResponse, db_response
from server.schema import SessionsCreate, SessionsFilters, SessionsRead, SessionsUpdate


SessionsRouter = APIRouter()


async def get_sessions_db(request: Request) -> SessionsDB:
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SessionsDB")
    db_pool: Pool = request.app.state.db_pool
    return SessionsDB(db_pool)


@SessionsRouter.get("/session", response_model=HTTPResponse[list[SessionsRead]])
async def get_sessions(
    filters: SessionsFilters = Depends(),
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> JSONResponse:
    """
    Retrieve all sessions.
    """
    filters_dict = filters.model_dump(exclude_none=True)
    return await db_response(sessions_db.get_all, status.HTTP_200_OK, filters_dict)


@SessionsRouter.get("/session/open", response_model=HTTPResponse[SessionsRead])
async def get_open_session(
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> JSONResponse:
    return await db_response(sessions_db.get_open_session, status.HTTP_200_OK)


@SessionsRouter.get("/session/{session_id}", response_model=HTTPResponse[SessionsRead])
async def get_session(
    session_id: UUID,
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> JSONResponse:
    """
    Get a session by its unique ID.

    Args:
        session_id (UUID): The unique identifier of the session to delete.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(sessions_db.get_one_by_id, status.HTTP_200_OK, session_id)


@SessionsRouter.post("/session", response_model=HTTPResponse[None])
async def add_session(
    session_data: SessionsCreate,
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> JSONResponse:
    """
    Create and register a new session.

    Args:
        session_data (SessionsCreate): The data required to start a new session.

    Returns:
        HTTPResponse: Success or error message. Does not return the created session.
    """
    return await db_response(sessions_db.insert_one, status.HTTP_201_CREATED, session_data)


@SessionsRouter.delete("/session/{session_id}", response_model=HTTPResponse[None])
async def delete_session(
    session_id: UUID,
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> JSONResponse:
    """
    Delete a session by its unique ID.

    Args:
        session_id (UUID): The unique identifier of the session to delete.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(sessions_db.delete_one, status.HTTP_204_NO_CONTENT, session_id)


@SessionsRouter.patch("/session/{session_id}", response_model=HTTPResponse[None])
async def patch_session(
    session_id: UUID,
    update: SessionsUpdate,
    sessions_db: SessionsDB = Depends(get_sessions_db),
) -> JSONResponse:
    """
    Partially update an existing session's data.

    Args:
        session_id (UUID): The unique identifier of the session to update.
        update (SessionsUpdate): The fields to update in the session.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(sessions_db.update_one, status.HTTP_202_ACCEPTED, session_id, update)
