from uuid import UUID

from fastapi import APIRouter, Depends, Request, status
from fastapi.responses import JSONResponse

from server.models import DBState, DictValues, TargetsDB
from server.routers.utils import HTTPResponse, db_response, get_all
from server.schema import TargetsCreate, TargetsUpdate


TargetsRouter = APIRouter(prefix="/target")


async def get_targets_db() -> TargetsDB:
    db_pool = await DBState.get_db_pool()
    return TargetsDB(db_pool)


def fix_targets_filter_types(filters_str: dict[str, str]) -> DictValues:
    # Cast each filter value to the proper type for the column

    float_fields = {"max_x_coordinate", "max_y_coordinate", "height"}
    uuid_fields = {"session_id"}
    str_fields = {"human_identifier"}
    filters: DictValues = {}
    for key, value in filters_str.items():
        if key in uuid_fields:
            filters[key] = UUID(value)
        elif key in float_fields:
            filters[key] = float(value)
        elif key in str_fields:
            filters[key] = value
        else:
            raise ValueError(f"ERROR: unknown field '{key}'")

    return filters


@TargetsRouter.get("", response_model=HTTPResponse[list[DictValues]])
async def get_targets(
    request: Request,
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Retrieve all targets.
    """
    return await get_all(
        request,
        fix_targets_filter_types,
        targets_db.get_all,
    )


@TargetsRouter.get("/{target_id}", response_model=HTTPResponse[DictValues])
async def get_target(
    target_id: UUID,
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    return await db_response(targets_db.get_one_by_id, status.HTTP_200_OK, target_id)


@TargetsRouter.post("", response_model=HTTPResponse[None])
async def add_target(
    target_data: TargetsCreate,
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Create and register a new target.

    Args:
        target_data (TargetsCreate): The data required to define a new target.

    Returns:
        HTTPResponse: Success or error message. Does not return the created target.
    """

    return await db_response(targets_db.insert_one, status.HTTP_201_CREATED, target_data)


@TargetsRouter.delete("/{target_id}", response_model=HTTPResponse[None])
async def delete_target(
    target_id: UUID,
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Delete a target configuration by its unique ID.

    Args:
        target_id (UUID): The unique identifier of the target to delete.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(targets_db.delete_one, status.HTTP_204_NO_CONTENT, target_id)


@TargetsRouter.patch("/{target_id}", response_model=HTTPResponse[None])
async def patch_target(
    target_id: UUID,
    update: TargetsUpdate,
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Partially update an existing target's configuration.

    Args:
        target_id (UUID): The unique identifier of the target to update.
        update (TargetsUpdate): The fields to update in the target.

    Returns:
        HTTPResponse: Success or error message.
    """
    return await db_response(targets_db.update_one, status.HTTP_202_ACCEPTED, target_id, update)
