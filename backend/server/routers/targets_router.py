from uuid import UUID

from fastapi import APIRouter, Depends, status
from fastapi.responses import JSONResponse

from server.models import DBState, TargetsDB
from server.routers.utils import HTTPResponse, db_response
from server.schema import TargetsCreate, TargetsRead, TargetsUpdate


TargetsRouter = APIRouter(prefix="/target")


async def get_targets_db() -> TargetsDB:
    db_pool = await DBState.get_db_pool()
    return TargetsDB(db_pool)


# GET all targets
@TargetsRouter.get("/", response_model=HTTPResponse[list[TargetsRead]])
async def get_all_targets(
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Retrieve all target configurations.

    Returns:
        HTTPResponse: A list of all targets stored in the system.
    """
    return await db_response(targets_db.get_all, status.HTTP_200_OK)


# POST a new target
@TargetsRouter.post("/", response_model=HTTPResponse[None])
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
    return await db_response(targets_db.insert_one, status.HTTP_200_OK, target_data)


# GET a specific target by ID
@TargetsRouter.get("/{target_id}", response_model=HTTPResponse[TargetsRead])
async def get_target(
    target_id: UUID,
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Retrieve details for a specific target by its unique ID.

    Args:
        target_id (UUID): The unique identifier of the target.

    Returns:
        HTTPResponse: The target object if found, or an error message if not found.
    """
    return await db_response(targets_db.get_one, status.HTTP_200_OK, target_id)


# DELETE a specific target by ID
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


# PATCH (partial update) a specific target by ID
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
    return await db_response(targets_db.delete_one, status.HTTP_202_ACCEPTED, target_id, update)
