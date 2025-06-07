from uuid import UUID

from fastapi import APIRouter, Depends, status
from fastapi.responses import JSONResponse

from server.models import DBState, ShotsDB
from server.routers.utils import HTTPResponse, db_response
from server.schema import ShotsFilters, ShotsRead


ShotsRouter = APIRouter()


async def get_shots_db() -> ShotsDB:
    db_pool = await DBState.get_db_pool()
    return ShotsDB(db_pool)


@ShotsRouter.get("/shot", response_model=HTTPResponse[list[ShotsRead]])
async def get_all_shots(
    filters: ShotsFilters = Depends(),
    shots_db: ShotsDB = Depends(get_shots_db),
) -> JSONResponse:
    """
    Retrieve all shots.
    """
    filters_dict = filters.model_dump(exclude_none=True)
    return await db_response(shots_db.get_all, status.HTTP_200_OK, filters_dict)


@ShotsRouter.get("/shot/{shot_id}", response_model=HTTPResponse[ShotsRead])
async def get_shot(
    shot_id: UUID,
    shots_db: ShotsDB = Depends(get_shots_db),
) -> JSONResponse:
    return await db_response(shots_db.get_one_by_id, status.HTTP_200_OK, shot_id)


@ShotsRouter.delete("/shot/{shot_id}", response_model=HTTPResponse[None])
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
