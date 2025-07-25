from uuid import uuid4

import pytest
from asyncpg import Pool

from server.models import ArrowsDB, DBNotFound
from server.schema import ArrowsUpdate
from tests.factories import create_fake_arrow


@pytest.mark.asyncio
async def test_create_arrow(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    payload = create_fake_arrow(human_identifier="B01")
    arrow_id = await db.insert_one(payload)
    assert arrow_id == payload.arrow_id


@pytest.mark.asyncio
async def test_get_all_arrows(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    payload1 = create_fake_arrow(human_identifier="B02")
    payload2 = create_fake_arrow(human_identifier="B03")
    await db.insert_one(payload1)
    await db.insert_one(payload2)
    arrows = await db.get_all()
    identifiers = [a.human_identifier for a in arrows]
    assert "B02" in identifiers and "B03" in identifiers


@pytest.mark.asyncio
async def test_get_specific_arrow(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    payload = create_fake_arrow(human_identifier="B04")
    arrow_id = await db.insert_one(payload)
    fetched = await db.get_one_by_id(arrow_id)
    assert fetched is not None
    assert fetched.human_identifier == "B04"


@pytest.mark.asyncio
async def test_update_arrow(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    payload = create_fake_arrow(human_identifier="B05", length=27.0)
    arrow_id = await db.insert_one(payload)

    update = ArrowsUpdate(length=30.0)
    await db.update_one(arrow_id, update)
    fetched = await db.get_one_by_id(arrow_id)
    assert fetched.length == 30.0


@pytest.mark.asyncio
async def test_update_nonexistent_arrow_raises(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    random_id = uuid4()
    update = ArrowsUpdate(length=31.5)
    with pytest.raises(DBNotFound):
        await db.update_one(random_id, update)


@pytest.mark.asyncio
async def test_delete_nonexistent_arrow_raises(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    random_id = uuid4()
    with pytest.raises(DBNotFound):
        await db.delete_one(random_id)


@pytest.mark.asyncio
async def test_get_nonexistent_arrow_raises(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    random_id = uuid4()
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(random_id)


@pytest.mark.asyncio
async def test_delete_arrow(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    payload = create_fake_arrow(human_identifier="Z99")
    arrow_id = await db.insert_one(payload)
    await db.delete_one(arrow_id)
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(arrow_id)
