from typing import Any
from uuid import UUID

from fastapi import APIRouter  # , HTTPException


TargetsRouter = APIRouter(prefix="/target")


# POST a new target
@TargetsRouter.post("/")
async def add_session() -> None:
    pass


# GET all sessions
@TargetsRouter.get("/", response_model=list[dict[str, Any]])
async def get_all_sessions() -> list[dict[str, Any]]:
    return []


# GET a specific target by ID
@TargetsRouter.get("/{target_id}", response_model=dict[str, Any])
async def get_session(target_id: UUID) -> dict[str, Any]:
    print(f"This is target_id {target_id}")
    return {}


# DELETE a specific target by ID
@TargetsRouter.delete("/{target_id}")
async def delete_session(target_id: UUID) -> None:
    print(f"This is target_id {target_id}")


# PATCH a specific target by ID
@TargetsRouter.patch("/{target_id}")
async def patch_session(target_id: UUID) -> None:
    print(f"This is target_id {target_id}")
