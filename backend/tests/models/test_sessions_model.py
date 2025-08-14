from datetime import datetime, timezone
from uuid import UUID, uuid4

import pytest
from asyncpg import Pool

from server.models import DBNotFound, SessionsDB
from server.schema import SessionsUpdate
from shared.factories import create_fake_session


@pytest.mark.asyncio
async def test_create_session(db_pool_initialed: Pool) -> None:
    db = SessionsDB(db_pool_initialed)
    payload = create_fake_session(location="Test Range 1", is_opened=True)
    session_id = await db.insert_one(payload)
    assert isinstance(session_id, UUID)
    session = await db.get_one_by_id(session_id)
    assert session.location == "Test Range 1"


@pytest.mark.asyncio
async def test_get_all_sessions(db_pool_initialed: Pool) -> None:
    db = SessionsDB(db_pool_initialed)
    payload1 = create_fake_session(location="Range 1")
    payload2 = create_fake_session(location="Range 2")
    await db.insert_one(payload1)
    await db.insert_one(payload2)
    sessions = await db.get_all()
    locations = [s.location for s in sessions]
    assert "Range 1" in locations and "Range 2" in locations


@pytest.mark.asyncio
async def test_get_specific_session(db_pool_initialed: Pool) -> None:
    db = SessionsDB(db_pool_initialed)
    payload = create_fake_session(location="Main Field")
    session_id = await db.insert_one(payload)
    fetched = await db.get_one_by_id(session_id)
    assert fetched is not None
    assert fetched.location == "Main Field"


@pytest.mark.asyncio
async def test_update_session(db_pool_initialed: Pool) -> None:
    db = SessionsDB(db_pool_initialed)
    payload = create_fake_session(location="Before Update", is_opened=True)
    session_id = await db.insert_one(payload)
    update = SessionsUpdate(
        location="After Update", is_opened=False, end_time=datetime.now(timezone.utc)
    )
    await db.update_one(session_id, update)
    fetched = await db.get_one_by_id(session_id)
    assert fetched.location == "After Update"
    assert fetched.is_opened is False


@pytest.mark.asyncio
async def test_delete_session(db_pool_initialed: Pool) -> None:
    db = SessionsDB(db_pool_initialed)
    payload = create_fake_session(location="To Delete")
    session_id = await db.insert_one(payload)
    await db.delete_one(session_id)
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(session_id)


@pytest.mark.asyncio
async def test_update_nonexistent_session_raises(db_pool_initialed: Pool) -> None:
    db = SessionsDB(db_pool_initialed)
    random_id = uuid4()
    update = SessionsUpdate(location="Ghost Session")
    with pytest.raises(DBNotFound):
        await db.update_one(random_id, update)


@pytest.mark.asyncio
async def test_delete_nonexistent_session_raises(db_pool_initialed: Pool) -> None:
    db = SessionsDB(db_pool_initialed)
    random_id = uuid4()
    with pytest.raises(DBNotFound):
        await db.delete_one(random_id)


@pytest.mark.asyncio
async def test_get_nonexistent_session_raises(db_pool_initialed: Pool) -> None:
    db = SessionsDB(db_pool_initialed)
    random_id = uuid4()
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(random_id)
