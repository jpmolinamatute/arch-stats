import logging
from uuid import UUID

from asyncpg import Pool
from fastapi import APIRouter, Depends, Request, status
from fastapi.responses import JSONResponse

from server.routers.utils import HTTPResponse, db_response
from shared.schema import ShotsFilters, ShotsRead
from shared.models import ShotsDB


ShotsRouter = APIRouter(prefix="/shot")


async def get_shots_db(request: Request) -> ShotsDB:
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting ShotsDB")
    db_pool: Pool = request.app.state.db_pool
    return ShotsDB(db_pool)


@ShotsRouter.get("/session-id/{session_id}")
async def get_all_shots_sessionid(
    session_id: UUID,
    shots_db: ShotsDB = Depends(get_shots_db),
) -> JSONResponse:

    return await db_response(shots_db.get_by_session_id, status.HTTP_200_OK, session_id)


# **************************************************************************************************
# **************************************************************************************************
# Path: /shot/{shot_id}
# **************************************************************************************************
# **************************************************************************************************


@ShotsRouter.get("/{shot_id}", response_model=HTTPResponse[ShotsRead])
async def get_shot(
    shot_id: UUID,
    shots_db: ShotsDB = Depends(get_shots_db),
) -> JSONResponse:
    return await db_response(shots_db.get_one_by_id, status.HTTP_200_OK, shot_id)


@ShotsRouter.delete("/{shot_id}", response_model=HTTPResponse[None])
async def delete_shot(
    shot_id: UUID,
    shots_db: ShotsDB = Depends(get_shots_db),
) -> JSONResponse:
    """
    Delete a specific shot by its unique ID.

    Args:
        shot_id (UUID): The unique identifier of the shot to delete.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(shots_db.delete_one, status.HTTP_204_NO_CONTENT, shot_id)


# **************************************************************************************************
# **************************************************************************************************
# Path: /shot
# **************************************************************************************************
# **************************************************************************************************


@ShotsRouter.get("", response_model=HTTPResponse[list[ShotsRead]])
async def get_all_shots(
    filters: ShotsFilters = Depends(),
    shots_db: ShotsDB = Depends(get_shots_db),
) -> JSONResponse:
    """
    Retrieve all shots.
    """
    filters_dict = filters.model_dump(exclude_none=True)
    return await db_response(shots_db.get_all, status.HTTP_200_OK, filters_dict)
