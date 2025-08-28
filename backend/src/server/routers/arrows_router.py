import logging
from uuid import UUID, uuid4

from asyncpg import Pool
from fastapi import APIRouter, Depends, Request, status
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse

from server.routers.utils import HTTPResponse, db_response
from shared.models import ArrowsDB
from shared.schema import ArrowsCreate, ArrowsFilters, ArrowsRead, ArrowsUpdate


ArrowsRouter = APIRouter(prefix="/arrow")


async def get_arrows_db(request: Request) -> ArrowsDB:
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting ArrowsDB")
    db_pool: Pool = request.app.state.db_pool
    return ArrowsDB(db_pool)


@ArrowsRouter.get("/new_arrow_uuid", response_model=HTTPResponse[str])
async def get_arrow_uuid() -> JSONResponse:
    """
    Generate and return a new unique UUID for use with arrow registration.

    Returns:
        str: A new UUID as a string.
    """
    resp = HTTPResponse[UUID](code=status.HTTP_200_OK, data=uuid4(), errors=[])
    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content=jsonable_encoder(resp),
    )


# **************************************************************************************************
# **************************************************************************************************
# Path: /arrow/{arrow_id}
# **************************************************************************************************
# **************************************************************************************************


@ArrowsRouter.get("/{arrow_id}", response_model=HTTPResponse[ArrowsRead])
async def get_arrow(
    arrow_id: UUID,
    arrows_db: ArrowsDB = Depends(get_arrows_db),
) -> JSONResponse:
    """
    Retrieve an arrow from the database by its unique ID.

    Args:
        arrow_id (UUID): The unique identifier of the arrow to delete.

    Returns:
        HTTPResponse: Success/failure message.
    """
    return await db_response(arrows_db.get_one_by_id, status.HTTP_200_OK, arrow_id)


@ArrowsRouter.delete("/{arrow_id}", response_model=HTTPResponse[None])
async def delete_arrow(
    arrow_id: UUID,
    arrows_db: ArrowsDB = Depends(get_arrows_db),
) -> JSONResponse:
    """
    Delete an arrow from the database by its unique ID.

    Args:
        arrow_id (UUID): The unique identifier of the arrow to delete.

    Returns:
        HTTPResponse: Success/failure message.
    """
    return await db_response(arrows_db.delete_one, status.HTTP_204_NO_CONTENT, arrow_id)


@ArrowsRouter.patch("/{arrow_id}", response_model=HTTPResponse[None])
async def patch_arrow(
    arrow_id: UUID,
    update: ArrowsUpdate,
    arrows_db: ArrowsDB = Depends(get_arrows_db),
) -> JSONResponse:
    """
    Partially update an existing arrow's data.

    Args:
        arrow_id (UUID): The unique identifier of the arrow to update.
        update (ArrowsUpdate): The data to update.

    Returns:
        HTTPResponse: Success/failure message.
    """
    return await db_response(arrows_db.update_one, status.HTTP_202_ACCEPTED, arrow_id, update)


# **************************************************************************************************
# **************************************************************************************************
# Path: /arrow
# **************************************************************************************************
# **************************************************************************************************


@ArrowsRouter.get("", response_model=HTTPResponse[list[ArrowsRead]])
async def get_all_arrows(
    filters: ArrowsFilters = Depends(),
    arrows_db: ArrowsDB = Depends(get_arrows_db),
) -> JSONResponse:
    """
    Retrieve all arrows.
    """
    filters_dict = filters.model_dump(exclude_none=True)
    return await db_response(arrows_db.get_all, status.HTTP_200_OK, filters_dict)


@ArrowsRouter.post("", response_model=HTTPResponse[UUID])
async def add_arrow(
    arrow_data: ArrowsCreate,
    arrows_db: ArrowsDB = Depends(get_arrows_db),
) -> JSONResponse:
    """
    Create and register a new arrow.

    Args:
        arrow_data (ArrowsCreate): The data for the arrow to be created.

    Returns:
        HTTPResponse: Success/failure message.
    """
    return await db_response(arrows_db.insert_one, status.HTTP_201_CREATED, arrow_data)
