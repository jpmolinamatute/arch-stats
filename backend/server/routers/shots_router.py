from datetime import datetime
from uuid import UUID

from fastapi import APIRouter, Depends, Request, status
from fastapi.responses import JSONResponse

from server.models import DBState, DictValues, ShotsDB
from server.routers.utils import HTTPResponse, db_response, get_all


ShotsRouter = APIRouter(prefix="/shot")


async def get_shots_db() -> ShotsDB:
    db_pool = await DBState.get_db_pool()
    return ShotsDB(db_pool)


def fix_shots_filter_types(filters_str: dict[str, str]) -> DictValues:
    date_fields = {"arrow_engage_time", "arrow_disengage_time", "arrow_landing_time"}
    uuid_fields = {"arrow_id"}
    float_fields = {"x_coordinate", "y_coordinate"}
    filters: DictValues = {}
    for key, value in filters_str.items():
        if key in date_fields:
            filters[key] = datetime.fromisoformat(value)
        elif key in float_fields:
            filters[key] = float(value)
        elif key in uuid_fields:
            filters[key] = UUID(value)
        else:
            raise ValueError(f"ERROR: unknown field '{key}'")
    return filters


@ShotsRouter.get("", response_model=HTTPResponse[list[DictValues]])
async def get_all_shots(
    request: Request,
    shots_db: ShotsDB = Depends(get_shots_db),
) -> JSONResponse:
    """
    Retrieve all shots.
    """
    return await get_all(
        request,
        fix_shots_filter_types,
        shots_db.get_all,
    )


@ShotsRouter.get("/{shot_id}", response_model=HTTPResponse[DictValues])
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
