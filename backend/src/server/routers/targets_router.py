import logging
from uuid import UUID, uuid4

from asyncpg import Pool
from fastapi import APIRouter, Depends, Request, status
from fastapi.responses import JSONResponse

from server.routers.utils import HTTPResponse, db_response
from server.schema import TargetsCreate, TargetsFilters, TargetsRead, TargetsUpdate
from shared.factories import create_fake_target
from shared.models import TargetsDB


TargetsRouter = APIRouter(prefix="/target")


async def get_targets_db(request: Request) -> TargetsDB:
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SessionsDB")
    db_pool: Pool = request.app.state.db_pool
    return TargetsDB(db_pool)


async def async_wrapper(session_id: UUID) -> TargetsCreate:
    return create_fake_target(session_id)


@TargetsRouter.get("/calibrate")
async def calibrate_target() -> JSONResponse:
    return await db_response(async_wrapper, status.HTTP_200_OK, uuid4())


# **************************************************************************************************
# **************************************************************************************************
# Path: /target/{target_id}
# **************************************************************************************************
# **************************************************************************************************


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
    target_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Partially update an existing target's data.

    Args:
        target_id (UUID): The unique identifier of the target to update.
        update (TargetsUpdate): The data to update.

    Returns:
        HTTPResponse: Success/failure message.
    """
    return await db_response(target_db.update_one, status.HTTP_202_ACCEPTED, target_id, update)


@TargetsRouter.get("/{target_id}", response_model=HTTPResponse[TargetsRead])
async def get_target(
    target_id: UUID,
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Retrieve an target from the database by its unique ID.

    Args:
        target_id (UUID): The unique identifier of the target to delete.

    Returns:
        HTTPResponse: Success/failure message.
    """
    return await db_response(targets_db.get_one_by_id, status.HTTP_200_OK, target_id)


# **************************************************************************************************
# **************************************************************************************************
# Path: /target
# **************************************************************************************************
# **************************************************************************************************


@TargetsRouter.post("", response_model=HTTPResponse[None])
async def add_target(
    target_data: TargetsCreate,
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Create new target.

    Args:
        target_data (TargetsCreate): The data required to define a new target.

    Returns:
        HTTPResponse: Success or error message. Does not return the created target.
    """

    return await db_response(targets_db.insert_one, status.HTTP_201_CREATED, target_data)


@TargetsRouter.get("", response_model=HTTPResponse[list[TargetsRead]])
async def get_all_targets(
    filters: TargetsFilters = Depends(),
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:
    """
    Retrieve all targets.
    """
    filters_dict = filters.model_dump(exclude_none=True)
    return await db_response(targets_db.get_all, status.HTTP_200_OK, filters_dict)


@TargetsRouter.get("/session-id/{session_id}")
async def get_all_targets_sessionid(
    session_id: UUID,
    targets_db: TargetsDB = Depends(get_targets_db),
) -> JSONResponse:

    return await db_response(targets_db.get_by_session_id, status.HTTP_200_OK, session_id)
