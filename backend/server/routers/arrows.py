from uuid import UUID, uuid4

from fastapi import APIRouter  # , HTTPException

from server.schema.arrow import ArrowCreate, ArrowRead, ArrowUpdate

ArrowRouter = APIRouter(prefix="/arrow")


# GET all arrows
@ArrowRouter.get("/", response_model=list[ArrowRead])
async def get_all_arrows() -> list[ArrowRead]:
    return []


# POST a new arrow
@ArrowRouter.post("/", response_model=ArrowCreate)
async def add_arrow() -> ArrowCreate:
    return {}


# GET a specific arrow by ID
@ArrowRouter.get("/{arrow_id}", response_model=ArrowRead)
async def get_arrow(arrow_id: UUID) -> ArrowRead:
    print(f"This is arrow_id {arrow_id}")
    return {}


# DELETE a specific arrow by ID
@ArrowRouter.delete("/{arrow_id}")
async def delete_arrow(arrow_id: UUID) -> dict[str, str]:
    return {"status": "deleted", "id": str(arrow_id)}


# PATCH (partial update) a specific arrow by ID
@ArrowRouter.patch("/{arrow_id}", response_model=ArrowRead)
async def patch_arrow(arrow_id: UUID, update: ArrowUpdate) -> ArrowRead:
    print(f"This is arrow_id {arrow_id} update {update}")
    return {}


# GET a new UUID string
@ArrowRouter.get("/new_id")
async def get_arrow_uuid() -> str:
    return str(uuid4())
