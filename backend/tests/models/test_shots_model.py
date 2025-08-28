from uuid import uuid4

import pytest
import pytest_asyncio
from asyncpg import Pool

from shared.schema import ShotsCreate
from shared.factories import create_fake_shot, create_many_arrows, create_many_sessions
from shared.models import DBNotFound, ShotsDB


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
