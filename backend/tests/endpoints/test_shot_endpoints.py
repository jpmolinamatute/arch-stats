from collections.abc import Callable
from typing import Any
from uuid import UUID

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from factories.archer_factory import create_archers
from factories.session_factory import create_sessions
from factories.slot_factory import create_slot_assignments
from factories.target_factory import create_targets


@pytest.mark.asyncio
async def test_get_shots_by_slot_returns_only_expected(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    Create 1 session, 1 target, 2 archers, 2 slots, 8 shots (4 per slot),
    and verify GET /shot/by-slot/{slot_id} returns only the expected shots.
    """

    # Create 2 archers
    archer1_id, archer2_id = await create_archers(db_pool, 2)

    # Create 1 session (owned by any archer)
    session_ids = await create_sessions(db_pool, 1)
    session_id = session_ids[0]

    # Create 1 target for the session
    target_ids = await create_targets(db_pool, 1, session_id=session_id)
    target_id = target_ids[0]

    # Create 2 slots (one for each archer)
    slot1_id, slot2_id = await create_slot_assignments(
        db_pool,
        2,
        archer_ids=[archer1_id, archer2_id],
        target_id=target_id,
        session_id=session_id,
    )

    # Helper to create a shot
    async def create_shot(slot_id: UUID, archer_id: UUID, value: int) -> Any:
        client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")
        payload = {
            "slot_id": str(slot_id),
            "x": 10.0 + value,
            "y": 20.0 + value,
            "score": value,
            "arrow_id": None,
        }
        r = await client.post("/api/v0/shot", json=payload)

        assert r.status_code == 201
        return r.json()

    # Create 4 shots for each slot
    slot1_shots = [await create_shot(slot1_id, archer1_id, v) for v in range(1, 5)]
    slot2_shots = [await create_shot(slot2_id, archer2_id, v) for v in range(5, 9)]

    # Test: archer1 gets only their shots for slot1
    client.cookies.set("arch_stats_auth", jwt_for(archer1_id), path="/")
    resp = await client.get(f"/api/v0/shot/by-slot/{slot1_id}")
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) == 4
    got_ids = {UUID(s["shot_id"]) for s in data}
    expected_ids = {UUID(s["shot_id"]) for s in slot1_shots}
    assert got_ids == expected_ids

    # Test: archer2 gets only their shots for slot2
    client.cookies.set("arch_stats_auth", jwt_for(archer2_id), path="/")
    resp = await client.get(f"/api/v0/shot/by-slot/{slot2_id}")
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) == 4
    got_ids = {UUID(s["shot_id"]) for s in data}
    expected_ids = {UUID(s["shot_id"]) for s in slot2_shots}
    assert got_ids == expected_ids

    # Test: archer1 forbidden from accessing archer2's slot
    client.cookies.set("arch_stats_auth", jwt_for(archer1_id), path="/")
    resp = await client.get(f"/api/v0/shot/by-slot/{slot2_id}")
    assert resp.status_code == 403
    assert resp.json()["detail"] == "Forbidden"
