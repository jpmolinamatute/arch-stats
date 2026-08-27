from collections.abc import Callable
from http import HTTPStatus
from uuid import UUID

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from factories.archer_factory import create_archers
from factories.session_factory import create_sessions
from factories.slot_factory import create_slot_assignments
from factories.target_factory import create_targets

TEST_SCORE = 9


async def _create_shot_for_test(
    client: AsyncClient,
    jwt_for: Callable[[UUID], str],
    slot_id: UUID,
    archer_id: UUID,
    value: int,
) -> None:
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")
    payload = {
        "slot_id": str(slot_id),
        "x": 10.0 + value,
        "y": 20.0 + value,
        "score": value,
        "arrow_id": None,
    }
    r = await client.post("/api/v0/shot", json=payload)
    assert r.status_code == HTTPStatus.CREATED


@pytest.mark.asyncio
async def test_get_stats_returns_correct_data(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    Verify GET /stats/{slot_id} returns correct compiled stats and shots list.
    """

    # 1. Arrange
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

    # 2. Add some shots
    score1 = 5
    score2 = 9
    # Shot 1: score 5
    await _create_shot_for_test(client, jwt_for, slot_id, archer_id, score1)
    # Shot 2: score 9
    await _create_shot_for_test(client, jwt_for, slot_id, archer_id, score2)

    # 3. Act
    # Verify authentication is likely not strictly required for stats *if* it's public,
    # but the router likely requires it?
    # Let's check stats_router logic...
    # The `get_stats` endpoint currently does NOT explicitly depend on `require_auth` or similar,
    # unless `get_shot_model` implicitly does?
    # `get_shot_model` usually just returns the model instance.
    # However, usually endpoints are protected.
    # checking `stats_router.py`:
    # @router.get("/{slot_id:uuid}", ...)
    # async def get_stats(slot_id: UUID, shot_model: Annotated[ShotModel, Depends(get_shot_model)])
    # It doesn't seem to enforce auth.

    resp = await client.get(f"/api/v0/stats/{slot_id}")

    # 4. Assert
    assert resp.status_code == HTTPStatus.OK
    data = resp.json()

    # Check 'scores' list
    scores = data["scores"]
    expected_count = 2
    assert len(scores) == expected_count
    # Sort by score or specific ID to verify? The order might be creation time DB default.
    sorted_score_values = sorted([s["score"] for s in scores])
    assert sorted_score_values == [score1, score2]

    # Verify structure of first shot
    first_shot = scores[0]
    assert "shot_id" in first_shot
    assert "score" in first_shot
    assert "is_x" in first_shot
    assert "created_at" in first_shot

    # Check 'stats' object
    stats = data["stats"]
    expected_total_score = score1 + score2
    expected_max_score = expected_count * 10
    expected_mean = float(expected_total_score) / expected_count

    assert stats["slot_id"] == str(slot_id)
    assert stats["number_of_shots"] == expected_count
    assert stats["total_score"] == expected_total_score
    assert stats["max_score"] == expected_max_score
    assert stats["mean"] == expected_mean


@pytest.mark.asyncio
async def test_get_stats_empty_slot(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    Verify GET /stats/{slot_id} returns zeros/empty list for a slot with no shots.
    """
    (archer_id,) = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[archer_id], target_id=target_id, session_id=session_id
    )

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")
    resp = await client.get(f"/api/v0/stats/{slot_id}")
    assert resp.status_code == HTTPStatus.OK
    data = resp.json()

    assert data["scores"] == []
    stats = data["stats"]
    assert stats["number_of_shots"] == 0
    assert stats["total_score"] == 0
    assert stats["max_score"] == 0
    assert stats["mean"] == 0.0
