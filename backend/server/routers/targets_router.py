from uuid import UUID

from fastapi import APIRouter, Depends, status
from fastapi.responses import JSONResponse

from server.models import DBState, DictValues, TargetsDB
from server.routers.utils import HTTPResponse, db_response
from server.schema import TargetsCreate, TargetsFilters, TargetsUpdate


TargetsRouter = APIRouter(prefix="/target")


async def get_targets_db() -> TargetsDB:
    db_pool = await DBState.get_db_pool()
    return TargetsDB(db_pool)


@TargetsRouter.get("", response_model=HTTPResponse[list[DictValues]])
async def get_targets(
    filters: TargetsFilters = Depends(),
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Retrieve all targets.
    """
    filters_dict = filters.model_dump(exclude_none=True)
    return await db_response(targets_db.get_all, status.HTTP_200_OK, filters_dict)


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
