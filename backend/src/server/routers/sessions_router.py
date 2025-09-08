import logging
from uuid import UUID

from asyncpg import Pool
from fastapi import APIRouter, Depends, Request, status
from fastapi.responses import JSONResponse

from server.routers.utils import HTTPResponse, db_response
from shared.models import SessionsModel
from shared.schema import SessionsCreate, SessionsFilters, SessionsRead, SessionsUpdate


SessionsRouter = APIRouter(prefix="/session")


async def get_sessions_db(request: Request) -> SessionsModel:
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SessionsModel")
    db_pool: Pool = request.app.state.db_pool
    return SessionsModel(db_pool)


@SessionsRouter.get("/open", response_model=HTTPResponse[SessionsRead])
async def get_open_session(
    sessions_db: SessionsModel = Depends(get_sessions_db),
) -> JSONResponse:
    return await db_response(sessions_db.get_open_session, status.HTTP_200_OK)


# # ****************************************
# # ****************************************
# Path: /api/v0/session/{session_id}
# # ****************************************
# # ****************************************


@SessionsRouter.get("/{session_id}", response_model=HTTPResponse[SessionsRead])
async def get_session(
    session_id: UUID,
    sessions_db: SessionsModel = Depends(get_sessions_db),
) -> JSONResponse:
    """
    Get a session by its unique ID.

    Args:
        session_id (UUID): The unique identifier of the session to delete.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(sessions_db.get_one_by_id, status.HTTP_200_OK, session_id)


@SessionsRouter.delete("/{session_id}", response_model=HTTPResponse[None])
async def delete_session(
    session_id: UUID,
    sessions_db: SessionsModel = Depends(get_sessions_db),
) -> JSONResponse:
    """
    Delete a session by its unique ID.

    Args:
        session_id (UUID): The unique identifier of the session to delete.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(sessions_db.delete_one, status.HTTP_204_NO_CONTENT, session_id)


@SessionsRouter.patch("/{session_id}", response_model=HTTPResponse[None])
async def patch_session(
    session_id: UUID,
    update: SessionsUpdate,
    sessions_db: SessionsModel = Depends(get_sessions_db),
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


# # ****************************************
# # ****************************************
# Path: /api/v0/session
# # ****************************************
# # ****************************************


@SessionsRouter.get("", response_model=HTTPResponse[list[SessionsRead]])
async def get_all_sessions(
    filters: SessionsFilters = Depends(),
    sessions_db: SessionsModel = Depends(get_sessions_db),
) -> JSONResponse:
    """
    Retrieve all sessions.
    """
    return await db_response(sessions_db.get_all, status.HTTP_200_OK, filters)


@SessionsRouter.post("", response_model=HTTPResponse[None])
async def create_session(
    session_data: SessionsCreate,
    sessions_db: SessionsModel = Depends(get_sessions_db),
) -> JSONResponse:
    """
    Create and register a new session.

    Args:
        session_data (SessionsCreate): The data required to start a new session.

    Returns:
        HTTPResponse: Success or error message. Does not return the created session.
    """
    return await db_response(sessions_db.insert_one, status.HTTP_201_CREATED, session_data)
