from uuid import uuid4

import pytest
from asyncpg import Pool

from shared.factories import create_fake_target, create_many_sessions
from shared.models import DBNotFound, TargetsDB


@pytest.mark.asyncio
async def test_create_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload = create_fake_target(session[0].session_id)
    payload.faces[0].human_identifier = "T01"
    target_id = await db.insert_one(payload)
    assert target_id is not None
    target = await db.get_one_by_id(target_id)
    assert target is not None
    assert target.faces[0].human_identifier == "T01"


@pytest.mark.asyncio
async def test_get_all_targets(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload1 = create_fake_target(session[0].session_id)
    payload1.faces[0].human_identifier = "T02"
    payload2 = create_fake_target(session[0].session_id)
    payload2.faces[0].human_identifier = "T03"
    await db.insert_one(payload1)
    await db.insert_one(payload2)
    targets = await db.get_all()
    identifiers = [face.human_identifier for t in targets for face in t.faces]
    assert "T02" in identifiers and "T03" in identifiers


@pytest.mark.asyncio
async def test_get_specific_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload = create_fake_target(session[0].session_id, human_identifier="T04")
    target_id = await db.insert_one(payload)
    fetched = await db.get_one({"id": target_id})
    assert fetched is not None
    assert target_id == fetched.get_id()


@pytest.mark.asyncio
async def test_delete_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload = create_fake_target(session[0].session_id)
    target_id = await db.insert_one(payload)
    await db.delete_one(target_id)
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(target_id)


@pytest.mark.asyncio
async def test_get_one_by_id_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload = create_fake_target(session[0].session_id)
    payload.faces[0].human_identifier = "T99"
    target_id = await db.insert_one(payload)
    target = await db.get_one_by_id(target_id)
    assert target.faces[0].human_identifier == "T99"


@pytest.mark.asyncio
async def test_get_nonexistent_target_raises(db_pool_initialed: Pool) -> None:
    db = TargetsDB(db_pool_initialed)
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(uuid4())


@pytest.mark.asyncio
async def test_delete_nonexistent_target_raises(db_pool_initialed: Pool) -> None:
    db = TargetsDB(db_pool_initialed)
    with pytest.raises(DBNotFound):
        await db.delete_one(uuid4())
