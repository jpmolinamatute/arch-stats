from uuid import uuid4

import pytest
import pytest_asyncio
from asyncpg import Pool

from server.models import DBException, DBNotFound, ShotsDB
from server.schema import ShotsCreate, ShotsUpdate
from shared.factories import create_fake_shot, create_many_arrows, create_many_sessions


# pylint: disable=redefined-outer-name


@pytest_asyncio.fixture
async def fake_shot_data(db_pool_initialed: Pool) -> ShotsCreate:
    """Creates a valid ShotsCreate schema."""
    arrow = await create_many_arrows(db_pool_initialed, 1)
    session = await create_many_sessions(db_pool_initialed, 1)
    return create_fake_shot(arrow[0].arrow_id, session[0].session_id)


@pytest.mark.asyncio
async def test_insert_and_get_one(db_pool_initialed: Pool, fake_shot_data: ShotsCreate) -> None:
    shots_db = ShotsDB(db_pool_initialed)
    # Insert a new shot
    shot_id = await shots_db.insert_one(fake_shot_data)
    assert shot_id is not None

    # Retrieve by ID
    shot = await shots_db.get_one_by_id(shot_id)
    assert shot.arrow_id == fake_shot_data.arrow_id
    assert shot.session_id == fake_shot_data.session_id
    assert shot.x == pytest.approx(fake_shot_data.x, rel=1e-6)
    assert shot.y == pytest.approx(fake_shot_data.y, rel=1e-6)


@pytest.mark.asyncio
async def test_delete_and_not_found(db_pool_initialed: Pool, fake_shot_data: ShotsCreate) -> None:
    shots_db = ShotsDB(db_pool_initialed)
    shot_id = await shots_db.insert_one(fake_shot_data)
    await shots_db.delete_one(shot_id)
    with pytest.raises(DBNotFound):
        await shots_db.get_one_by_id(shot_id)


@pytest.mark.asyncio
async def test_get_all(db_pool_initialed: Pool, fake_shot_data: ShotsCreate) -> None:
    shots_db = ShotsDB(db_pool_initialed)
    # Insert multiple shots
    ids = []
    for i in range(3):
        new_data = fake_shot_data.model_copy()
        new_data.x = 1.1 + i
        new_data.y = 2.2 + i
        ids.append(await shots_db.insert_one(new_data))

    all_shots = await shots_db.get_all()
    assert len(all_shots) == 3
    xs = [shot.x for shot in all_shots if shot.x]
    ys = [shot.y for shot in all_shots if shot.y]
    assert sorted(xs) == pytest.approx([1.1, 2.1, 3.1], rel=1e-6)
    assert sorted(ys) == pytest.approx([2.2, 3.2, 4.2], rel=1e-6)


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
async def test_update_nonexistent_shot_raises(db_pool_initialed: Pool) -> None:
    shots_db = ShotsDB(db_pool_initialed)
    update = ShotsUpdate()
    with pytest.raises(DBException):
        await shots_db.update_one(uuid4(), update)
