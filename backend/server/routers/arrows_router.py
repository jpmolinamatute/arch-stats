from uuid import UUID

from fastapi import APIRouter, Depends, status
from fastapi.responses import JSONResponse

from server.models import ArrowsDB, DBState
from server.routers.utils import HTTPResponse, db_response
from server.schema import ArrowsCreate, ArrowsRead, ArrowsUpdate


ArrowsRouter = APIRouter(prefix="/arrow")


async def get_arrows_db() -> ArrowsDB:
    db_pool = await DBState.get_db_pool()
    return ArrowsDB(db_pool)


# GET all arrows
@ArrowsRouter.get("/", response_model=HTTPResponse[list[ArrowsRead]])
async def get_all_arrows(
    arrows_db: ArrowsDB = Depends(get_arrows_db),
) -> JSONResponse:
    """
    Retrieve all arrows registered in the system.

    Returns:
        HTTPResponse: A list of all arrow objects currently stored in the database.
    """
    return await db_response(arrows_db.get_all, status.HTTP_200_OK)


# POST a new arrow
@ArrowsRouter.post("/", response_model=HTTPResponse[None])
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


# GET a specific arrow by ID
@ArrowsRouter.get("/{arrow_id}", response_model=HTTPResponse[ArrowsRead])
async def get_arrow(
    arrow_id: UUID,
    arrows_db: ArrowsDB = Depends(get_arrows_db),
) -> JSONResponse:
    """
    Retrieve details for a single arrow by its unique ID.

    Args:
        arrow_id (UUID): The unique identifier of the arrow.

    Returns:
        HTTPResponse: The arrow object if found, or error message if not found.
    """
    return await db_response(arrows_db.get_one, status.HTTP_200_OK, arrow_id)


# DELETE a specific arrow by ID
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


# PATCH (partial update) a specific arrow by ID
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
