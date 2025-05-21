from uuid import UUID, uuid4

from fastapi import APIRouter, status

from database import ArrowsDB, DBException, DBState
from database.schema import ArrowCreate, ArrowRead, ArrowUpdate, HTTPResponse


ArrowRouter = APIRouter(prefix="/arrow")


# GET all arrows
@ArrowRouter.get("/", response_model=HTTPResponse[list[ArrowRead]])
async def get_all_arrows() -> HTTPResponse[list[ArrowRead]]:
    try:
        pool = await DBState.get_db_pool()
        arrows = ArrowsDB(pool)
        resp = HTTPResponse[list[ArrowRead]](code=status.HTTP_200_OK, data=await arrows.get_all())
    except Exception as e:
        resp = HTTPResponse[list[ArrowRead]](
            code=status.HTTP_500_INTERNAL_SERVER_ERROR, data=None, errors=[str(e)]
        )
    return resp


# POST a new arrow
@ArrowRouter.post("/", response_model=HTTPResponse[None])
async def add_arrow(arrow_data: ArrowCreate) -> HTTPResponse[None]:
    pool = await DBState.get_db_pool()
    arrows = ArrowsDB(pool)
    await arrows.insert_one(arrow_data)
    resp = HTTPResponse[None](code=status.HTTP_201_CREATED)
    return resp


# GET a specific arrow by ID
@ArrowRouter.get("/{arrow_id}", response_model=HTTPResponse[ArrowRead | None])
async def get_arrow(arrow_id: UUID) -> HTTPResponse[ArrowRead | None]:
    try:
        pool = await DBState.get_db_pool()
        arrows = ArrowsDB(pool)
        result = await arrows.get_one(arrow_id)
        if result:
            resp = HTTPResponse[ArrowRead | None](code=status.HTTP_200_OK, data=result)
        else:
            resp = HTTPResponse[ArrowRead | None](
                code=status.HTTP_404_NOT_FOUND,
                data=None,
                errors=[f"Arrow with id '{arrow_id}' was not found"],
            )
    except Exception as e:
        resp = HTTPResponse[ArrowRead | None](
            code=status.HTTP_500_INTERNAL_SERVER_ERROR, data=None, errors=[str(e)]
        )
    return resp


# DELETE a specific arrow by ID
@ArrowRouter.delete("/{arrow_id}", response_model=HTTPResponse[None])
async def delete_arrow(arrow_id: UUID) -> HTTPResponse[None]:
    try:
        pool = await DBState.get_db_pool()
        arrows = ArrowsDB(pool)
        await arrows.delete_one(arrow_id)
        return HTTPResponse[None](code=status.HTTP_204_NO_CONTENT, data=None)
    except DBException as e:
        return HTTPResponse[None](code=status.HTTP_404_NOT_FOUND, data=None, errors=[str(e)])
    except Exception as e:
        return HTTPResponse[None](
            code=status.HTTP_500_INTERNAL_SERVER_ERROR, data=None, errors=[str(e)]
        )


# PATCH (partial update) a specific arrow by ID
@ArrowRouter.patch("/{arrow_id}", response_model=HTTPResponse[None])
async def patch_arrow(arrow_id: UUID, update: ArrowUpdate) -> HTTPResponse[None]:
    pool = await DBState.get_db_pool()
    arrows = ArrowsDB(pool)
    await arrows.update_one(arrow_id, update)
    return HTTPResponse[None](code=status.HTTP_202_ACCEPTED, data=None)


# GET a new UUID string
@ArrowRouter.get("/new_id")
async def get_arrow_uuid() -> str:
    return str(uuid4())
