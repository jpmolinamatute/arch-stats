from uuid import UUID

from fastapi import APIRouter, Depends, status, Request
from fastapi.responses import JSONResponse

from server.models import DBState, ShotsDB, DictValues
from server.routers.utils import HTTPResponse, db_response


ShotsRouter = APIRouter(prefix="/shot")


async def get_shots_db() -> ShotsDB:
    db_pool = await DBState.get_db_pool()
    return ShotsDB(db_pool)


@ShotsRouter.get("/", response_model=HTTPResponse[list[DictValues] | DictValues])
async def get_all_shots(
    request: Request,
    shots_db: ShotsDB = Depends(get_shots_db),
) -> JSONResponse:
    """
    Retrieve all recorded shots.
    """
    filters = dict(request.query_params.items())

    return await db_response(shots_db.get_all, status.HTTP_200_OK, filters)


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
