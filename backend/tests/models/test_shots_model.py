from uuid import uuid4

import pytest
import pytest_asyncio
from asyncpg import Pool

from shared.factories import create_fake_shot, create_many_arrows, create_many_sessions
from shared.models import DBException, DBNotFound, ShotsDB
from shared.schema import ShotsCreate


# pylint: disable=redefined-outer-name


@pytest_asyncio.fixture
async def fake_shot_data(db_pool_initialed: Pool) -> ShotsCreate:
    """Creates a valid ShotsCreate schema."""
    arrow = await create_many_arrows(db_pool_initialed, 1)
    session = await create_many_sessions(db_pool_initialed, 1)
    return create_fake_shot(arrow[0].arrow_id, session[0].session_id)


@pytest.mark.asyncio
async def test_delete_nonexistent_raises(db_pool_initialed: Pool) -> None:
    shots_db = ShotsDB(db_pool_initialed)
    non_existing_id = uuid4()
    with pytest.raises(DBNotFound):
        await shots_db.delete_one(non_existing_id)


@pytest.mark.asyncio
async def test_get_nonexistent_shot_raises(db_pool_initialed: Pool) -> None:
    shots_db = ShotsDB(db_pool_initialed)
    with pytest.raises(DBNotFound):
        await shots_db.get_one_by_id(uuid4())


@pytest.mark.asyncio
async def test_insert_and_fetch_hit_shot(db_pool_initialed: Pool) -> None:
    # Arrange
    arrows = await create_many_arrows(db_pool_initialed, 1)
    sessions = await create_many_sessions(db_pool_initialed, 1)
    shots_db = ShotsDB(db_pool_initialed)
    payload = create_fake_shot(arrows[0].arrow_id, sessions[0].session_id)

    # Act
    shot_id = await shots_db.insert_one(payload)

    # Assert
    got = await shots_db.get_one_by_id(shot_id)
    assert got.x is not None and got.y is not None and got.arrow_landing_time is not None


@pytest.mark.asyncio
async def test_insert_and_fetch_miss_shot(db_pool_initialed: Pool) -> None:
    # Arrange
    arrows = await create_many_arrows(db_pool_initialed, 1)
    sessions = await create_many_sessions(db_pool_initialed, 1)
    shots_db = ShotsDB(db_pool_initialed)
    payload = create_fake_shot(
        arrows[0].arrow_id,
        sessions[0].session_id,
        arrow_landing_time=None,
        x=None,
        y=None,
    )

    # Act
    shot_id = await shots_db.insert_one(payload)

    # Assert
    got = await shots_db.get_one_by_id(shot_id)
    assert got.arrow_landing_time is None and got.x is None and got.y is None


@pytest.mark.asyncio
async def test_constraint_violation_partial_coordinates_fail(db_pool_initialed: Pool) -> None:
    # landing time present but x/y missing must fail per CHECK constraint
    arrows = await create_many_arrows(db_pool_initialed, 1)
    sessions = await create_many_sessions(db_pool_initialed, 1)
    shots_db = ShotsDB(db_pool_initialed)
    bad = create_fake_shot(
        arrows[0].arrow_id,
        sessions[0].session_id,
        x=None,
        y=None,
        # keep default non-null landing time -> violates constraint
    )
    with pytest.raises(DBException):
        await shots_db.insert_one(bad)


@pytest.mark.asyncio
async def test_get_all_and_by_session(db_pool_initialed: Pool) -> None:
    arrows = await create_many_arrows(db_pool_initialed, 3)
    sessions = await create_many_sessions(db_pool_initialed, 2)
    # Insert 2 shots for session A and 1 for session B
    a1 = create_fake_shot(arrows[0].arrow_id, sessions[0].session_id)
    a2 = create_fake_shot(arrows[1].arrow_id, sessions[0].session_id)
    b1 = create_fake_shot(arrows[2].arrow_id, sessions[1].session_id)
    shots_db = ShotsDB(db_pool_initialed)
    await shots_db.insert_one(a1)
    await shots_db.insert_one(a2)
    await shots_db.insert_one(b1)

    all_shots = await shots_db.get_all()
    assert len(all_shots) >= 3

    only_session_a = await shots_db.get_by_session_id(sessions[0].session_id)
    assert len(only_session_a) == 2
    assert all(s.session_id == sessions[0].session_id for s in only_session_a)


@pytest.mark.asyncio
async def test_cascade_delete_on_arrow_removes_shots(db_pool_initialed: Pool) -> None:
    arrows = await create_many_arrows(db_pool_initialed, 1)
    sessions = await create_many_sessions(db_pool_initialed, 1)
    # Insert shot
    shots_db = ShotsDB(db_pool_initialed)
    shot_id = await shots_db.insert_one(
        create_fake_shot(arrows[0].arrow_id, sessions[0].session_id)
    )
    # Delete arrow; FK ON DELETE CASCADE should remove shot
    async with db_pool_initialed.acquire() as conn:
        await conn.execute("DELETE FROM arrows WHERE id = $1", arrows[0].arrow_id)

    with pytest.raises(DBNotFound):
        await shots_db.get_one_by_id(shot_id)
