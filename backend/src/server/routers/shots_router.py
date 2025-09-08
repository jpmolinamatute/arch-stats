import logging
from uuid import UUID

from asyncpg import Pool
from fastapi import APIRouter, Depends, Request, status
from fastapi.responses import JSONResponse

from server.routers.utils import HTTPResponse, db_response
from shared.models import ShotsModel
from shared.schema import ShotsFilters, ShotsRead


ShotsRouter = APIRouter(prefix="/shot")


async def get_shots_db(request: Request) -> ShotsModel:
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting ShotsModel")
    db_pool: Pool = request.app.state.db_pool
    return ShotsModel(db_pool)


# # ****************************************
# # ****************************************
# Path: /api/v0/shot/{shot_id}
# # ****************************************
# # ****************************************


@ShotsRouter.delete("/{shot_id}", response_model=HTTPResponse[None])
async def delete_shot(
    shot_id: UUID,
    shots_db: ShotsModel = Depends(get_shots_db),
) -> JSONResponse:
    """
    Delete a specific shot by its unique ID.

    Args:
        shot_id (UUID): The unique identifier of the shot to delete.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(shots_db.delete_one, status.HTTP_204_NO_CONTENT, shot_id)


# # ****************************************
# # ****************************************
# Path: /api/v0/shot
# # ****************************************
# # ****************************************


@ShotsRouter.get("", response_model=HTTPResponse[list[ShotsRead]])
async def get_all_shots(
    filters: ShotsFilters = Depends(),
    shots_db: ShotsModel = Depends(get_shots_db),
) -> JSONResponse:
    """
    Retrieve all shots.
    """
    return await db_response(shots_db.get_all, status.HTTP_200_OK, filters)
