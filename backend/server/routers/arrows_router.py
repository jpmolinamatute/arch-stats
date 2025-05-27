from uuid import UUID, uuid4

from fastapi import APIRouter, Depends, status, Request
from fastapi.responses import JSONResponse

from server.models import ArrowsDB, DBState, DictValues
from server.routers.utils import HTTPResponse, db_response
from server.schema import ArrowsCreate, ArrowsUpdate


ArrowsRouter = APIRouter(prefix="/arrow")


async def get_arrows_db() -> ArrowsDB:
    db_pool = await DBState.get_db_pool()
    return ArrowsDB(db_pool)


@ArrowsRouter.get("/new_arrow_uuid", response_model=str)
async def get_arrow_uuid() -> str:
    """
    Generate and return a new unique UUID for use with arrow registration.

    Returns:
        str: A new UUID as a string.
    """
    return str(uuid4())


@ArrowsRouter.get("", response_model=HTTPResponse[list[DictValues]])
async def get_all_arrows(
    request: Request,
    arrows_db: ArrowsDB = Depends(get_arrows_db),
) -> JSONResponse:
    """
    Retrieve all arrows registered in the system.
    """
    filters = dict(request.query_params.items())
    return await db_response(arrows_db.get_all, status.HTTP_200_OK, filters)


@ArrowsRouter.get("/{arrow_id}", response_model=HTTPResponse[DictValues])
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
        HTTPResponse: Success/failure message. Does not return the created arrow.
    """
    return await db_response(arrows_db.insert_one, status.HTTP_201_CREATED, arrow_data)


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
    return await db_response(arrows_db.update_one, status.HTTP_204_NO_CONTENT, arrow_id, update)
