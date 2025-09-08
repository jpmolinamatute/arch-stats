from __future__ import annotations

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from shared.factories.arrows_factory import create_many_arrows
from shared.factories.sessions_factory import create_fake_session, insert_sessions_db
from shared.factories.shots_factory import create_many_shots
from shared.factories.targets_factory import create_many_targets
from shared.models import SessionPerformanceModel


pytestmark = pytest.mark.asyncio


async def _seed(db_pool: Pool) -> None:
    session = (await insert_sessions_db(db_pool, [create_fake_session(is_opened=True)]))[0]
    await create_many_targets(db_pool, session.get_id(), targets_count=1)
    arrows = await create_many_arrows(db_pool, arrows_count=2)
    arrows_ids = [a.get_id() for a in arrows]
    await create_many_shots(db_pool, arrows_ids, session.get_id(), shots_count=3)
    # Ensure the view exists
    await SessionPerformanceModel(db_pool).create()


async def test_get_session_performance_returns_rows(
    async_client: AsyncClient, db_pool_initialed: Pool
) -> None:
    await _seed(db_pool_initialed)

    resp = await async_client.get("/api/v0/session_performance")
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 200
    assert body["errors"] == []
    data = body["data"]
    assert isinstance(data, list)
    assert len(data) > 0
    first = data[0]
    print(first)
    for key in [
        "id",
        "session_id",
        "arrow_id",
        "arrow_engage_time",
        "arrow_disengage_time",
        "arrow_landing_time",
        "x",
        "y",
        "time_of_flight_seconds",
        "arrow_speed",
        "score",
        "human_identifier",
    ]:
        assert key in first
