from uuid import uuid4

import pytest
from asyncpg import Pool

from shared.factories import create_fake_target, create_many_sessions
from shared.models import DBNotFound, TargetsModel
from shared.schema.targets_schema import TargetsFilters


@pytest.mark.asyncio
async def test_create_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsModel(db_pool_initialed)
    payload = create_fake_target(session[0].session_id)
    target_id = await db.insert_one(payload)
    assert target_id is not None
    target = await db.get_one_by_id(target_id)
    assert target is not None
    assert target.session_id == session[0].session_id


@pytest.mark.asyncio
async def test_get_all_targets(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsModel(db_pool_initialed)
    payload1 = create_fake_target(session[0].session_id)
    payload2 = create_fake_target(session[0].session_id)
    await db.insert_one(payload1)
    await db.insert_one(payload2)
    targets = await db.get_all()
    assert len(targets) >= 2
    assert all(t.session_id == session[0].session_id for t in targets)


@pytest.mark.asyncio
async def test_get_specific_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsModel(db_pool_initialed)
    payload = create_fake_target(session[0].session_id)
    target_id = await db.insert_one(payload)
    where = TargetsFilters(id=target_id)
    fetched = await db.get_one(where)
    assert fetched is not None
    assert target_id == fetched.get_id()


@pytest.mark.asyncio
async def test_delete_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsModel(db_pool_initialed)
    payload = create_fake_target(session[0].session_id)
    target_id = await db.insert_one(payload)
    await db.delete_one(target_id)
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(target_id)


@pytest.mark.asyncio
async def test_get_one_by_id_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsModel(db_pool_initialed)
    payload = create_fake_target(session[0].session_id)
    target_id = await db.insert_one(payload)
    target = await db.get_one_by_id(target_id)
    assert target.get_id() == target_id


@pytest.mark.asyncio
async def test_get_nonexistent_target_raises(db_pool_initialed: Pool) -> None:
    db = TargetsModel(db_pool_initialed)
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(uuid4())


@pytest.mark.asyncio
async def test_delete_nonexistent_target_raises(db_pool_initialed: Pool) -> None:
    db = TargetsModel(db_pool_initialed)
    with pytest.raises(DBNotFound):
        await db.delete_one(uuid4())


@pytest.mark.asyncio
async def test_get_by_session_id(db_pool_initialed: Pool) -> None:
    sessions = await create_many_sessions(db_pool_initialed, 2)
    db = TargetsModel(db_pool_initialed)
    t1 = create_fake_target(sessions[0].session_id)
    t2 = create_fake_target(sessions[0].session_id)
    t3 = create_fake_target(sessions[1].session_id)
    await db.insert_one(t1)
    await db.insert_one(t2)
    await db.insert_one(t3)

    s1_targets = await db.get_by_session_id(sessions[0].session_id)
    assert len(s1_targets) == 2
    assert all(t.session_id == sessions[0].session_id for t in s1_targets)
