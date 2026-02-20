from collections.abc import Callable
from http import HTTPStatus
from typing import Any
from uuid import UUID

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from factories.archer_factory import create_archers
from factories.session_factory import create_sessions
from factories.slot_factory import create_slot_assignments
from factories.target_factory import create_targets

MIN_SCORE = 0
MAX_SCORE = 10
INVALID_SCORE_LOW = -1
INVALID_SCORE_HIGH = 11
SHOT_BATCH_SIZE = 6
TARGET_X_Y = 5.0
TARGET_SCORE = 5
TEST_SCORE = 9


async def _create_shot_for_test(
    client: AsyncClient,
    jwt_for: Callable[[UUID], str],
    slot_id: UUID,
    archer_id: UUID,
    value: int,
) -> Any:
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
    return r.json()


async def _verify_returned_shots(
    client: AsyncClient,
    jwt_for: Callable[[UUID], str],
    slot_id: UUID,
    archer_id: UUID,
    expected_shots: list[Any],
) -> None:
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")
    resp = await client.get(f"/api/v0/shot/by-slot/{slot_id}")
    assert resp.status_code == HTTPStatus.OK
    data = resp.json()
    assert len(data) == len(expected_shots)
    got_ids = {UUID(s["shot_id"]) for s in data}
    expected_ids = {UUID(s["shot_id"]) for s in expected_shots}
    assert got_ids == expected_ids


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
    (session_id,) = await create_sessions(db_pool, 1)

    # Create 1 target for the session
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)

    # Create 2 slots (one for each archer)
    slot1_id, slot2_id = await create_slot_assignments(
        db_pool,
        2,
        archer_ids=[archer1_id, archer2_id],
        target_id=target_id,
        session_id=session_id,
    )

    # Create 4 shots for each slot
    slot1_shots = [
        await _create_shot_for_test(client, jwt_for, slot1_id, archer1_id, v) for v in range(1, 5)
    ]
    slot2_shots = [
        await _create_shot_for_test(client, jwt_for, slot2_id, archer2_id, v) for v in range(5, 9)
    ]

    # Test: archer1 gets only their shots for slot1
    await _verify_returned_shots(client, jwt_for, slot1_id, archer1_id, slot1_shots)

    # Test: archer2 gets only their shots for slot2
    await _verify_returned_shots(client, jwt_for, slot2_id, archer2_id, slot2_shots)

    # Test: archer1 forbidden from accessing archer2's slot
    client.cookies.set("arch_stats_auth", jwt_for(archer1_id), path="/")
    resp = await client.get(f"/api/v0/shot/by-slot/{slot2_id}")
    assert resp.status_code == HTTPStatus.FORBIDDEN
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
        "score": TEST_SCORE,
        # "is_x": omitted (defaults to False)
        # "arrow_id": omitted (allowed)
    }
    resp = await client.post("/api/v0/shot", json=payload)
    assert resp.status_code == HTTPStatus.CREATED
    shot_id = UUID(resp.json()["shot_id"])  # ensure a valid UUID

    # Verify via GET that row exists with expected values and defaults
    resp_get = await client.get(f"/api/v0/shot/by-slot/{slot_id}")
    assert resp_get.status_code == HTTPStatus.OK
    items = resp_get.json()
    found = next((s for s in items if UUID(s["shot_id"]) == shot_id), None)
    assert found is not None
    assert found["x"] == pytest.approx(1.23)
    assert found["y"] == pytest.approx(-4.56)
    assert found["score"] == TEST_SCORE
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
            "score": INVALID_SCORE_LOW,
        },
    )
    assert resp.status_code == HTTPStatus.UNPROCESSABLE_ENTITY

    # score > 10
    resp = await client.post(
        "/api/v0/shot",
        json={
            "slot_id": str(slot_id),
            "x": 0.0,
            "y": 0.0,
            "score": INVALID_SCORE_HIGH,
        },
    )
    assert resp.status_code == HTTPStatus.UNPROCESSABLE_ENTITY

    # boundaries 0 and 10 allowed
    ok0 = await client.post(
        "/api/v0/shot",
        json={
            "slot_id": str(slot_id),
            "x": 0.0,
            "y": 0.0,
            "score": MIN_SCORE,
        },
    )
    assert ok0.status_code == HTTPStatus.CREATED

    ok10 = await client.post(
        "/api/v0/shot",
        json={
            "slot_id": str(slot_id),
            "x": 0.5,
            "y": 0.5,
            "score": MAX_SCORE,
        },
    )
    assert ok10.status_code == HTTPStatus.CREATED


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
    assert r1.status_code == HTTPStatus.UNPROCESSABLE_ENTITY

    # x and y provided, missing score
    r2 = await client.post("/api/v0/shot", json={"slot_id": str(slot_id), "x": 0.1, "y": 0.2})
    assert r2.status_code == HTTPStatus.UNPROCESSABLE_ENTITY

    # Only score provided
    r3 = await client.post("/api/v0/shot", json={"slot_id": str(slot_id), "score": 5})
    assert r3.status_code == HTTPStatus.UNPROCESSABLE_ENTITY

    # Valid: all None (omit all three)
    r4 = await client.post("/api/v0/shot", json={"slot_id": str(slot_id)})
    assert r4.status_code == HTTPStatus.CREATED

    # Valid: all present
    r5 = await client.post(
        "/api/v0/shot",
        json={"slot_id": str(slot_id), "x": -0.3, "y": 1.1, "score": 7},
    )
    assert r5.status_code == HTTPStatus.CREATED


@pytest.mark.asyncio
async def test_create_multiple_shots_success(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /shot with a list of shots (bulk insert)."""

    # Arrange
    (archer_id,) = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[archer_id], target_id=target_id, session_id=session_id
    )

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Prepare 6 shots
    shots_payload = []
    for i in range(SHOT_BATCH_SIZE):
        shots_payload.append(
            {
                "slot_id": str(slot_id),
                "x": float(i),
                "y": float(i),
                "score": i,  # 0..5
            }
        )

    # Act
    resp = await client.post("/api/v0/shot", json=shots_payload)

    # Assert
    assert resp.status_code == HTTPStatus.CREATED
    data = resp.json()
    assert isinstance(data, list)
    assert len(data) == SHOT_BATCH_SIZE
    # Verify they are all valid UUIDs
    created_ids = [UUID(item["shot_id"]) for item in data]
    assert len(set(created_ids)) == SHOT_BATCH_SIZE

    # Verify persistence
    resp_get = await client.get(f"/api/v0/shot/by-slot/{slot_id}")
    assert resp_get.status_code == HTTPStatus.OK
    fetched_shots = resp_get.json()
    assert len(fetched_shots) == SHOT_BATCH_SIZE

    # Verify content of one of them (e.g. score 5)
    shot_5 = next((s for s in fetched_shots if s["score"] == TARGET_SCORE), None)
    assert shot_5 is not None
    assert shot_5["x"] == TARGET_X_Y
    assert shot_5["y"] == TARGET_X_Y


@pytest.mark.asyncio
async def test_create_multiple_shots_different_slots_fails(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /shot with a list of shots for DIFFERENT slots should fail."""

    # Arrange
    (archer_id,) = await create_archers(db_pool, 1)
    session1_id, session2_id = await create_sessions(db_pool, 2)
    (target1_id,) = await create_targets(db_pool, 1, session_id=session1_id)
    (target2_id,) = await create_targets(db_pool, 1, session_id=session2_id)

    # Create 1 slot in each session for the same archer
    (slot1_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[archer_id], target_id=target1_id, session_id=session1_id
    )
    (slot2_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[archer_id], target_id=target2_id, session_id=session2_id
    )

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Prepare shots for different slots
    shots_payload = [
        {
            "slot_id": str(slot1_id),
            "x": 1.0,
            "y": 1.0,
            "score": 1,
        },
        {
            "slot_id": str(slot2_id),
            "x": 2.0,
            "y": 2.0,
            "score": 2,
        },
        {
            "slot_id": str(slot1_id),
            "x": 3.0,
            "y": 3.0,
            "score": 3,
        },
    ]

    # Act
    resp = await client.post("/api/v0/shot", json=shots_payload)

    # Assert
    assert resp.status_code == HTTPStatus.BAD_REQUEST
    assert resp.json()["detail"] == "All shots must belong to the same slot"


@pytest.mark.asyncio
async def test_create_shots_empty_list_fails(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /shot with an empty list should fail."""

    # Arrange
    (archer_id,) = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Act
    resp = await client.post("/api/v0/shot", json=[])

    # Assert
    assert resp.status_code == HTTPStatus.BAD_REQUEST
    assert resp.json()["detail"] == "Invalid input"


@pytest.mark.asyncio
async def test_create_shots_fewer_than_three_fails(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /shot with a list of fewer than 3 shots should fail."""

    # Arrange
    (archer_id,) = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[archer_id], target_id=target_id, session_id=session_id
    )

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Prepare 2 shots (less than 3)
    shots_payload = [
        {"slot_id": str(slot_id), "x": 1.0, "y": 1.0, "score": 10},
        {"slot_id": str(slot_id), "x": 2.0, "y": 2.0, "score": 9},
    ]

    # Act
    resp = await client.post("/api/v0/shot", json=shots_payload)

    # Assert
    assert resp.status_code == HTTPStatus.BAD_REQUEST
    assert resp.json()["detail"] == "Invalid input"


@pytest.mark.asyncio
async def test_create_shots_more_than_ten_fails(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /shot with a list of more than 10 shots should fail."""

    # Arrange
    (archer_id,) = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[archer_id], target_id=target_id, session_id=session_id
    )

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Prepare 11 shots (more than 10)
    shots_payload = []
    for i in range(11):
        shots_payload.append({"slot_id": str(slot_id), "x": float(i), "y": float(i), "score": 10})

    # Act
    resp = await client.post("/api/v0/shot", json=shots_payload)

    # Assert
    assert resp.status_code == HTTPStatus.BAD_REQUEST
    assert resp.json()["detail"] == "Invalid input"


@pytest.mark.asyncio
async def test_count_by_slot_returns_zero_for_empty_slot(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    # 1. Setup
    (owner_id,) = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[owner_id], target_id=target_id, session_id=session_id
    )

    # 2. Call count
    client.cookies.set("arch_stats_auth", jwt_for(owner_id))
    response = await client.get(f"/api/v0/shot/count-by-slot/{slot_id}")

    assert response.status_code == HTTPStatus.OK
    assert response.json() == 0


@pytest.mark.asyncio
async def test_count_by_slot_returns_correct_count_for_multiple_shots(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    # 1. Setup
    (owner_id,) = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[owner_id], target_id=target_id, session_id=session_id
    )

    # 2. Add shots
    shots = [
        {"slot_id": str(slot_id), "x": 5.0, "y": 5.0, "score": 10},
        {"slot_id": str(slot_id), "x": 5.0, "y": 5.0, "score": 9},
        {"slot_id": str(slot_id), "x": 5.0, "y": 5.0, "score": 8},
    ]
    client.cookies.set("arch_stats_auth", jwt_for(owner_id))
    await client.post("/api/v0/shot", json=shots)

    # 3. Call count
    response = await client.get(f"/api/v0/shot/count-by-slot/{slot_id}")

    assert response.status_code == HTTPStatus.OK
    assert response.json() == len(shots)
