import logging
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Request, Response, status

from models import ArcherModel, DBNotFound
from routers.deps.models import get_archer_model
from schema import ArcherCreate, ArcherFilter, ArcherId, ArcherRead, ArcherUpdate


router = APIRouter(prefix="/archer", tags=["Archers"])


@router.get("/", response_model=list[ArcherRead])
async def list_archers(
    archer_model: ArcherModel = Depends(get_archer_model),
) -> list[ArcherRead]:
    """
    List archers.

    Returns all archers matching default filter.

    Responses: 200 OK.
    """
    archers = await archer_model.get_all(ArcherFilter(), [])
    return archers


@router.get("/{archer_id:uuid}", response_model=ArcherRead, status_code=status.HTTP_200_OK)
async def get_archer(
    archer_id: UUID,
    request: Request,
    archer_model: ArcherModel = Depends(get_archer_model),
) -> ArcherRead:
    """
    Get a single archer by id.

    Responses: 200 OK, 404 Not Found, 422 Unprocessable Content.
    """
    try:
        logger: logging.Logger = request.app.state.logger
        logger.debug(f"Getting archer with id: {archer_id}")
        where = ArcherFilter(archer_id=archer_id)
        return await archer_model.get_one(where)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=str(e)) from e


@router.post(
    "/",
    response_model=ArcherId,
    status_code=status.HTTP_201_CREATED,
)
async def create_archer(
    data: ArcherCreate,
    archer_model: ArcherModel = Depends(get_archer_model),
) -> ArcherId:
    """
    Create a new archer.

    Returns newly created archer id.

    Responses: 201 Created.
    """
    archer_id = await archer_model.insert_one(data)
    return ArcherId(archer_id=archer_id)


@router.patch(
    "/",
    response_class=Response,
    status_code=status.HTTP_200_OK,
)
async def update_archer(
    data: ArcherUpdate, archer_model: ArcherModel = Depends(get_archer_model)
) -> Response:
    """
    Update archer fields matching filter.

    Responses: 200 OK, 404 Not Found, 422 Unprocessable Content.
    """
    try:
        await archer_model.update(data.data, data.where)
        return Response(status_code=status.HTTP_200_OK)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=str(e)) from e


@router.delete("/{archer_id:uuid}", status_code=status.HTTP_204_NO_CONTENT, response_class=Response)
async def delete_archer(
    archer_id: UUID, archer_model: ArcherModel = Depends(get_archer_model)
) -> Response:
    """
    Delete an archer by id.

    Responses: 204 No Content, 404 Not Found, 422 Unprocessable Content.
    """
    try:
        await archer_model.delete_one(archer_id)
        return Response(status_code=status.HTTP_204_NO_CONTENT)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=str(e)) from e
