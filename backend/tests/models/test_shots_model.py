from uuid import uuid4

import pytest
import pytest_asyncio
from asyncpg import Pool

from shared.factories import create_many_arrows, create_many_sessions
from shared.models import DBNotFound, ShotsModel


# pylint: disable=redefined-outer-name


@pytest_asyncio.fixture
async def _setup_arrows_sessions(db_pool_initialed: Pool) -> None:
    await create_many_arrows(db_pool_initialed, 1)
    await create_many_sessions(db_pool_initialed, 1)


@pytest.mark.asyncio
async def test_delete_nonexistent_raises_once(db_pool_initialed: Pool) -> None:
    shots_db = ShotsModel(db_pool_initialed)
    non_existing_id = uuid4()
    with pytest.raises(DBNotFound):
        await shots_db.delete_one(non_existing_id)


# second duplicate removed; behavior covered above
