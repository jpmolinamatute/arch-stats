from uuid import UUID
from typing import Any
from fastapi import APIRouter  # , HTTPException


SessionsRouter = APIRouter(prefix="/session")


# POST a new session
@SessionsRouter.post("/")
async def add_session() -> None:
    pass


# GET all sessions
@SessionsRouter.get("/", response_model=list[dict[str, Any]])
async def get_all_sessions() -> list[dict[str, Any]]:
    return []


# GET a specific session by ID
@SessionsRouter.get("/{session_id}", response_model=dict[str, Any])
async def get_session(session_id: UUID) -> dict[str, Any]:
    print(f"This is session_id {session_id}")
    return {}


# DELETE a specific session by ID
@SessionsRouter.delete("/{session_id}")
async def delete_session(session_id: UUID) -> None:
    print(f"This is session_id {session_id}")


# PATCH a specific session by ID
@SessionsRouter.patch("/{session_id}")
async def patch_session(session_id: UUID) -> None:
    print(f"This is session_id {session_id}")
