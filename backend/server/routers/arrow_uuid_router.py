from uuid import uuid4

from fastapi import APIRouter

ArrowUUIDRouter = APIRouter(prefix="/new_arrow")


# GET a new UUID string
@ArrowUUIDRouter.get("/", response_model=str)
async def get_arrow_uuid() -> str:
    """
    Generate and return a new unique UUID for use with arrow registration.

    Returns:
        str: A new UUID as a string.
    """
    return str(uuid4())
