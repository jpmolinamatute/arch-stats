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


@pytest.mark.asyncio
async def test_close_session_forbidden_for_non_owner(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    Test that a user cannot close another user's session.
    Current behavior (vulnerable): Returns 200 OK.
    Expected behavior (fixed): Returns 403 Forbidden.
    """
    owner_id, stranger_id = await create_archers(db_pool, 2)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == HTTPStatus.CREATED
    session_id = resp.json()["session_id"]

    # Stranger attempts to close it
    client.cookies.set("arch_stats_auth", jwt_for(stranger_id), path="/")
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})

    assert resp.status_code == HTTPStatus.FORBIDDEN
    assert resp.json()["detail"] == "Forbidden"


@pytest.mark.asyncio
async def test_reopen_session_blocked_if_already_open(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    Test that a user cannot re-open a session if they already have an open one.
    Current behavior (vulnerable): Returns 200 OK (user has 2 open sessions).
    Expected behavior (fixed): Returns 422 Unprocessable Entity.
    """
    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

    # 1. Create Session A
    resp = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Range A",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    session_a_id = resp.json()["session_id"]

    # 2. Close Session A
    await client.patch("/api/v0/session/close", json={"session_id": session_a_id})

    # 3. Create Session B (now open)
    resp = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Range B",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    # session_b_id = resp.json()["session_id"]

    # 4. Attempt to Re-open Session A
    resp = await client.patch("/api/v0/session/re-open", json={"session_id": session_a_id})

    assert resp.status_code == HTTPStatus.CONFLICT  # Or 422
    assert "already has an opened session" in resp.json()["detail"]


@pytest.mark.asyncio
async def test_create_shot_blocked_on_closed_session(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    Test that shots cannot be added to a closed session ("Zombie Shots").
    Current behavior (vulnerable): Returns 201 Created.
    Expected behavior (fixed): Returns 422 Unprocessable Entity.
    """
    [archer_id] = await create_archers(db_pool, 1)
    (session_id,) = await create_sessions(db_pool, 1)  # Starts open? verify factory
    (target_id,) = await create_targets(db_pool, 1, session_id=session_id)
    (slot_id,) = await create_slot_assignments(
        db_pool, 1, archer_ids=[archer_id], target_id=target_id, session_id=session_id
    )

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Close the session
    # Note: Using DB directly or endpoint? Endpoint is safer to mimic real flow
    # But wait, create_sessions factory makes it... let's check. likely open.
    # Let's ensure it's closed.

    # ACTUALLY, checking factory: create_sessions inserts directly. is_opened defaults to True?
    # Let's use endpoint to close it to be sure/easy, or update DB.
    # We are authorized as owner.
    # Wait, create_sessions might not set us as owner if we didn't pass owner_id?
    # Valid point. safer to just update DB directly to close it.

    await db_pool.execute("UPDATE session SET is_opened = FALSE WHERE session_id = $1", session_id)

    # Attempt to add shot
    payload = {"slot_id": str(slot_id), "x": 0.0, "y": 0.0, "score": 10}
    resp = await client.post("/api/v0/shot", json=payload)

    assert resp.status_code == HTTPStatus.UNPROCESSABLE_ENTITY
    assert "Cannot add shots to a closed session" in resp.json()["detail"]
