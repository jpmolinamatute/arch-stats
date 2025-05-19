from uuid import UUID, uuid4

from fastapi import APIRouter  # , HTTPException

from server.schema.arrow import ArrowCreate, ArrowRead, ArrowUpdate
from database.arrows_model import Arrows

router = APIRouter(prefix="/arrow")


# GET all arrows
@router.get("/", response_model=list[ArrowRead])
async def get_all_arrows() -> list[ArrowRead]:
    return []


# POST a new arrow
@router.post("/", response_model=ArrowCreate)
async def add_arrow() -> ArrowCreate:
    arrow_in: ArrowCreate = request.body
    arrow = Arrows(**arrow_in.model_dump())
    session.add(arrow)
    session.commit()
    return arrow_in


# GET a specific arrow by ID
@router.get("/{arrow_id}", response_model=ArrowRead)
async def get_arrow(arrow_id: UUID) -> ArrowRead:
    return {}


# DELETE a specific arrow by ID
@router.delete("/{arrow_id}")
async def delete_arrow(arrow_id: UUID) -> dict:
    return {"status": "deleted", "id": str(arrow_id)}


# PATCH (partial update) a specific arrow by ID
@router.patch("/{arrow_id}", response_model=ArrowRead)
async def patch_arrow(arrow_id: UUID, update: ArrowUpdate) -> ArrowRead:
    pass


# GET a new UUID string
@router.get("/new_id")
async def get_arrow_uuid() -> str:
    return str(uuid4())
