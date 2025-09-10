import logging
import random
from typing import Annotated
from uuid import UUID

from asyncpg import Pool
from fastapi import APIRouter, Body, Depends, Request, status
from fastapi.responses import JSONResponse

from server.routers.utils import HTTPResponse, db_response
from shared.models import FacesModel
from shared.schema.faces_schema import Face, FaceCalibration, FacesCreate, FacesFilters, FacesRead


FacesRouter = APIRouter(prefix="")


async def get_faces_db(request: Request) -> FacesModel:
    """Dependency to provide a FacesModel bound to the shared DB pool."""
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting FacesModel")
    db_pool: Pool = request.app.state.db_pool
    return FacesModel(db_pool)


async def async_wrapper() -> FaceCalibration:
    return FaceCalibration(
        x=random.uniform(120.0, 240.0),
        y=random.uniform(0.0, 100.0),
        radii=[10.0, 20.0, 30.0, 40.0, 50.0],
    )


@FacesRouter.get("/face/calibrate")
async def calibrate_target() -> JSONResponse:
    # As of now, this is a dummy endpoint to calibrate a new target. At some point we are going to
    # start using real data from sensors (hardware)
    return await db_response(async_wrapper, status.HTTP_200_OK)


# ************************************************************
# Nested resource under target: /api/v0/target/{target_id}/face
# ************************************************************


@FacesRouter.get("/target/{target_id}/face", response_model=HTTPResponse[list[FacesRead]])
async def list_faces_for_target(
    target_id: UUID, faces_db: FacesModel = Depends(get_faces_db)
) -> JSONResponse:
    """Return all faces for a given target id (empty list if none)."""
    filters = FacesFilters(target_id=target_id)
    return await db_response(faces_db.get_all, status.HTTP_200_OK, filters)


@FacesRouter.post("/target/{target_id}/face", response_model=HTTPResponse[list[UUID]])
async def create_faces_for_target(
    target_id: UUID,
    faces_payload: Annotated[list[Face], Body(..., min_length=1, max_length=3)],
    faces_db: FacesModel = Depends(get_faces_db),
) -> JSONResponse:
    """Create one or more faces (1..3) for a given target.

    Each Face payload omits target_id path param which is injected server-side.
    Returns list of created face ids.
    """
    created_ids: list[UUID] = []
    # Insert sequentially (small batches) – if any fails, raise immediately.
    for face in faces_payload:
        payload = FacesCreate(
            target_id=target_id,
            x=face.x,
            y=face.y,
            human_identifier=face.human_identifier,
            radii=face.radii,
            points=face.points,
        )
        face_id = await faces_db.insert_one(payload)
        created_ids.append(face_id)

    async def _return_ids() -> list[UUID]:
        return created_ids

    return await db_response(_return_ids, status.HTTP_201_CREATED)


# ************************************************************
# Flat face resources
# ************************************************************


@FacesRouter.get("/face", response_model=HTTPResponse[list[FacesRead]])
async def list_all_faces(
    faces_db: FacesModel = Depends(get_faces_db),
    filters: FacesFilters = Depends(),
) -> JSONResponse:
    """Return all faces (optionally filterable by query params)."""
    return await db_response(faces_db.get_all, status.HTTP_200_OK, filters)


@FacesRouter.delete("/face/{face_id}", response_model=HTTPResponse[None])
async def delete_face(face_id: UUID, faces_db: FacesModel = Depends(get_faces_db)) -> JSONResponse:
    """Delete a face by id."""
    return await db_response(faces_db.delete_one, status.HTTP_204_NO_CONTENT, face_id)
