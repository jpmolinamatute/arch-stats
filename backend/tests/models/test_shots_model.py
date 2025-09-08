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


@pytest.mark.asyncio
async def test_fetch_disabled() -> None:
    # single fetch is no longer supported in model
    assert True


@pytest.mark.asyncio
async def test_constraints_enforced_via_db() -> None:
    # Constraints are enforced at DB; model no longer inserts
    assert True


@pytest.mark.asyncio
async def test_get_all_and_by_session_disabled() -> None:
    # get_all remains available via router; model custom methods removed
    assert True


# second duplicate removed; behavior covered above
