from typing import Any
from uuid import UUID
from fastapi import APIRouter  # , HTTPException


ShotsRouter = APIRouter(prefix="/shot")


# GET all shots
@ShotsRouter.get("/", response_model=list[dict[str, Any]])
async def get_all_shots() -> list[dict[str, Any]]:
    return []


# DELETE a specific shot by ID
@ShotsRouter.delete("/{session_id}")
async def delete_shot(session_id: UUID) -> None:
    print(f"This is session_id {session_id}")
