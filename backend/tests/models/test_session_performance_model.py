from __future__ import annotations

import pytest
from asyncpg import Pool

from shared.factories.arrows_factory import create_many_arrows
from shared.factories.sessions_factory import create_fake_session, insert_sessions_db
from shared.factories.shots_factory import create_many_shots
from shared.factories.targets_factory import create_many_targets
from shared.models import SessionPerformanceModel


async def _seed_open_session_with_data(db_pool: Pool) -> None:
    # One open session
    s = await insert_sessions_db(db_pool, [create_fake_session(is_opened=True)])
    session = s[0]
    # Target(s) and faces for scoring
    await create_many_targets(db_pool, session.get_id(), targets_count=1)
    # Arrows
    arrows = await create_many_arrows(db_pool, arrows_count=3)
    arrow_ids = [a.get_id() for a in arrows]
    # Shots
    await create_many_shots(db_pool, arrow_ids, session.get_id(), shots_count=5)


@pytest.mark.asyncio
async def test_session_performance_view_returns_rows(db_pool_initialed: Pool) -> None:
    await _seed_open_session_with_data(db_pool_initialed)

    sp = SessionPerformanceModel(db_pool_initialed)
    # View is created during app startup normally; ensure present for direct model test
    await sp.create()
    rows = await sp.get_all(None)
    assert isinstance(rows, list)
    assert len(rows) > 0
    first = rows[0]
    assert first.session_id is not None
    assert first.arrow_id is not None
    # Speed/flight time may be None if landing_time missing; should exist for our seeded shots
    assert first.time_of_flight_seconds is not None
