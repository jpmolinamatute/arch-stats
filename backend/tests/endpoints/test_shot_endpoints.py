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


@pytest.mark.asyncio
async def test_create_shot_happy_path_from_schema(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /shot happy path following ShotBase: provide slot_id + x,y,score.

    Omit optional is_x and arrow_id and verify defaults via GET.
    """

    # Arrange: archer -> session -> target -> slot
    (archer_id,) = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool,
        1,
        archer_ids=[archer_id],
        target_id=target_id,
        session_id=session_id,
    )

    # Authenticate as owning archer
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    payload = {
        "slot_id": str(slot_id),
        "x": 1.23,
        "y": -4.56,
        "score": 9,
        # "is_x": omitted (defaults to False)
        # "arrow_id": omitted (allowed)
    }
    resp = await client.post("/api/v0/shot", json=payload)
    assert resp.status_code == 201
    shot_id = UUID(resp.json()["shot_id"])  # ensure a valid UUID

    # Verify via GET that row exists with expected values and defaults
    resp_get = await client.get(f"/api/v0/shot/by-slot/{slot_id}")
    assert resp_get.status_code == 200
    items = resp_get.json()
    found = next((s for s in items if UUID(s["shot_id"]) == shot_id), None)
    assert found is not None
    assert found["x"] == pytest.approx(1.23)
    assert found["y"] == pytest.approx(-4.56)
    assert found["score"] == 9
    assert found["is_x"] is False
    assert found["arrow_id"] is None


@pytest.mark.asyncio
async def test_create_shot_rejects_score_out_of_range(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """score must be within 0..10; out-of-range payloads return 422 (validation)."""

    (archer_id,) = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[archer_id], target_id=target_id, session_id=session_id
    )

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # score < 0
    resp = await client.post(
        "/api/v0/shot",
        json={
            "slot_id": str(slot_id),
            "x": 0.0,
            "y": 0.0,
            "score": -1,
        },
    )
    assert resp.status_code == 422

    # score > 10
    resp = await client.post(
        "/api/v0/shot",
        json={
            "slot_id": str(slot_id),
            "x": 0.0,
            "y": 0.0,
            "score": 11,
        },
    )
    assert resp.status_code == 422

    # boundaries 0 and 10 allowed
    ok0 = await client.post(
        "/api/v0/shot",
        json={
            "slot_id": str(slot_id),
            "x": 0.0,
            "y": 0.0,
            "score": 0,
        },
    )
    assert ok0.status_code == 201

    ok10 = await client.post(
        "/api/v0/shot",
        json={
            "slot_id": str(slot_id),
            "x": 0.5,
            "y": 0.5,
            "score": 10,
        },
    )
    assert ok10.status_code == 201


@pytest.mark.asyncio
async def test_create_shot_rejects_incomplete_coordinates_set(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """All-or-none validation: (x,y,score) must be all present or all None."""

    (archer_id,) = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[archer_id], target_id=target_id, session_id=session_id
    )

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Only x provided
    r1 = await client.post("/api/v0/shot", json={"slot_id": str(slot_id), "x": 0.1})
    assert r1.status_code == 422

    # x and y provided, missing score
    r2 = await client.post("/api/v0/shot", json={"slot_id": str(slot_id), "x": 0.1, "y": 0.2})
    assert r2.status_code == 422

    # Only score provided
    r3 = await client.post("/api/v0/shot", json={"slot_id": str(slot_id), "score": 5})
    assert r3.status_code == 422

    # Valid: all None (omit all three)
    r4 = await client.post("/api/v0/shot", json={"slot_id": str(slot_id)})
    assert r4.status_code == 201

    # Valid: all present
    r5 = await client.post(
        "/api/v0/shot",
        json={"slot_id": str(slot_id), "x": -0.3, "y": 1.1, "score": 7},
    )
    assert r5.status_code == 201
