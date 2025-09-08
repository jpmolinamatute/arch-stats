from datetime import datetime, timezone
from uuid import UUID, uuid4

import pytest
from asyncpg import Pool

from shared.factories import create_fake_session
from shared.models import DBException, DBNotFound, SessionsModel
from shared.schema import SessionsUpdate


@pytest.mark.asyncio
async def test_create_session(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    payload = create_fake_session(location="Test Range 1", is_opened=True)
    session_id = await db.insert_one(payload)
    assert isinstance(session_id, UUID)
    session = await db.get_one_by_id(session_id)
    assert session.location == "Test Range 1"


@pytest.mark.asyncio
async def test_get_all_sessions(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    # Create first session (open) then close it to produce a closed row
    first_id = await db.insert_one(create_fake_session(location="Range 1", is_opened=True))
    await db.update_one(
        first_id,
        SessionsUpdate(is_opened=False, end_time=datetime.now(timezone.utc)),
    )
    # Create a second (current open) session
    second_id = await db.insert_one(create_fake_session(location="Range 2", is_opened=True))
    sessions = await db.get_all()
    locations = [s.location for s in sessions]
    assert "Range 1" in locations and "Range 2" in locations
    # Validate open/closed distribution
    open_sessions = [s for s in sessions if s.is_opened]
    assert len(open_sessions) == 1
    assert open_sessions[0].session_id == second_id


@pytest.mark.asyncio
async def test_get_specific_session(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    payload = create_fake_session(location="Main Field")
    session_id = await db.insert_one(payload)
    fetched = await db.get_one_by_id(session_id)
    assert fetched is not None
    assert fetched.location == "Main Field"


@pytest.mark.asyncio
async def test_update_session(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
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
    db = SessionsModel(db_pool_initialed)
    payload = create_fake_session(location="To Delete")
    session_id = await db.insert_one(payload)
    await db.delete_one(session_id)
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(session_id)


@pytest.mark.asyncio
async def test_update_nonexistent_session_raises(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    random_id = uuid4()
    update = SessionsUpdate(location="Ghost Session")
    with pytest.raises(DBNotFound):
        await db.update_one(random_id, update)


@pytest.mark.asyncio
async def test_delete_nonexistent_session_raises(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    random_id = uuid4()
    with pytest.raises(DBNotFound):
        await db.delete_one(random_id)


@pytest.mark.asyncio
async def test_get_nonexistent_session_raises(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    random_id = uuid4()
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(random_id)


@pytest.mark.asyncio
async def test_get_open_session_none_when_no_open(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    # No sessions yet => no open session
    found = await db.get_open_session()
    assert found is None


@pytest.mark.asyncio
async def test_get_open_session_helper(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    open_session = create_fake_session(is_opened=True)
    await db.insert_one(open_session)
    found = await db.get_open_session()
    assert found is not None and found.is_opened is True


@pytest.mark.asyncio
async def test_insert_second_open_session_rejected(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    await db.insert_one(create_fake_session(is_opened=True, location="First"))
    with pytest.raises(DBException):
        await db.insert_one(create_fake_session(is_opened=True, location="Second"))


@pytest.mark.asyncio
async def test_update_to_open_when_other_open_rejected(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    # Create a session and immediately close it to have a closed record
    closed_id = await db.insert_one(create_fake_session(is_opened=True, location="Closed"))
    await db.update_one(
        closed_id, SessionsUpdate(is_opened=False, end_time=datetime.now(timezone.utc))
    )
    # Create a new open session
    open_id = await db.insert_one(create_fake_session(is_opened=True, location="Open"))
    # Attempt to re-open the previously closed one (should fail while another open exists)
    with pytest.raises(DBException):
        await db.update_one(closed_id, SessionsUpdate(is_opened=True))
    # Close current open properly
    await db.update_one(
        open_id, SessionsUpdate(is_opened=False, end_time=datetime.now(timezone.utc))
    )
    # Now reopen the previously closed session (allowed)
    await db.update_one(closed_id, SessionsUpdate(is_opened=True))


@pytest.mark.asyncio
async def test_session_check_constraints(db_pool_initialed: Pool) -> None:
    db = SessionsModel(db_pool_initialed)
    # Violates open_session_no_end_time (opened True but end_time provided)
    bad = create_fake_session(is_opened=True)
    sid = await db.insert_one(bad)
    with pytest.raises(DBException):
        await db.update_one(sid, SessionsUpdate(end_time=datetime.now(timezone.utc)))
