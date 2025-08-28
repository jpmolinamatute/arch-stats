from uuid import uuid4

import pytest
from asyncpg import Pool

from shared.factories import create_fake_arrow
from shared.models import ArrowsDB, DBException, DBNotFound
from shared.schema import ArrowsUpdate


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


@pytest.mark.asyncio
async def test_unique_human_identifier_constraint(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    p1 = create_fake_arrow(human_identifier="UNIQ-1")
    p2 = create_fake_arrow(human_identifier="UNIQ-1")
    await db.insert_one(p1)
    with pytest.raises(DBException):
        await db.insert_one(p2)


@pytest.mark.asyncio
async def test_active_voided_consistency_check(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    p = create_fake_arrow(human_identifier="CCHK-1", is_active=True)
    _id = await db.insert_one(p)
    # Setting voided_date while keeping is_active True should violate check
    with pytest.raises(DBException):
        await db.update_one(_id, ArrowsUpdate(voided_date=p.registration_date))


@pytest.mark.asyncio
async def test_get_all_filtered(db_pool_initialed: Pool) -> None:
    db = ArrowsDB(db_pool_initialed)
    await db.insert_one(create_fake_arrow(human_identifier="F-A", is_programmed=True))
    await db.insert_one(create_fake_arrow(human_identifier="F-B", is_programmed=False))
    only_prog = await db.get_all({"is_programmed": True})
    assert all(a.is_programmed for a in only_prog)
