"""Endpoint tests for session-related endpoints.

Covers:
- GET   /session/archer/{archer_id}/open-session
- GET   /session/archer/{archer_id}/participating
- GET   /session/open
- GET   /session/archer/{archer_id}/close-session
- POST  /session
- PATCH /session/close
- POST  /session/slot
- PATCH /session/slot/leave
- PATCH /session/re-open
"""

import time
from typing import Any
from uuid import UUID

import jwt
import pytest
from asyncpg import Pool
from httpx import AsyncClient

from core import settings
from factories.archer_factory import create_archers


async def _truncate_all(pool: Pool) -> None:
    async with pool.acquire() as conn:
        # Order matters due to FK
        await conn.execute("TRUNCATE slot RESTART IDENTITY CASCADE")
        await conn.execute("TRUNCATE target RESTART IDENTITY CASCADE")
        await conn.execute("TRUNCATE session RESTART IDENTITY CASCADE")
        await conn.execute("TRUNCATE archer RESTART IDENTITY CASCADE")


def _jwt_for(archer_id: UUID) -> str:
    now = int(time.time())
    return jwt.encode(
        {
            "sub": str(archer_id),
            "iat": now,
            "exp": now + 3600,
            "iss": "arch-stats",
            "typ": "access",
        },
        settings.jwt_secret,
        algorithm=settings.jwt_algorithm,
    )


async def join_session(
    client: AsyncClient, session_id: UUID, archer_id: UUID, distance: int = 30
) -> Any:
    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")
    payload = {
        "session_id": str(session_id),
        "archer_id": str(archer_id),
        "distance": distance,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    r = await client.post("/api/v0/session/slot", json=payload)
    assert r.status_code == 200
    return r.json()


@pytest.mark.asyncio
async def test_close_session_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/close must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """
    await _truncate_all(db_pool)

    # No auth cookie; even with a dummy session_id, the auth dependency should reject
    payload = {"session_id": "00000000-0000-0000-0000-000000000000"}
    resp = await client.patch("/api/v0/session/close", json=payload)
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_create_session_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """POST /session must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """
    await _truncate_all(db_pool)

    # Seed one archer for a valid payload
    [owner_id] = await create_archers(db_pool, 1)

    payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_join_slot_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """POST /session/slot must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """
    await _truncate_all(db_pool)

    # Construct a valid SlotJoinRequest payload shape so the request reaches auth
    payload = {
        "session_id": "00000000-0000-0000-0000-000000000000",
        "archer_id": "00000000-0000-0000-0000-000000000000",
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }

    resp = await client.post("/api/v0/session/slot", json=payload)
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_leave_slot_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/slot/leave must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """
    await _truncate_all(db_pool)

    # Construct a valid SlotLeaveRequest payload so the request reaches auth
    payload = {
        "session_id": "00000000-0000-0000-0000-000000000000",
        "archer_id": "00000000-0000-0000-0000-000000000000",
    }

    resp = await client.patch("/api/v0/session/slot/leave", json=payload)
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_open_sessions_list_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/open must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """
    await _truncate_all(db_pool)

    resp = await client.get("/api/v0/session/open")
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_participating_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/archer/{archer_id}/participating must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """
    await _truncate_all(db_pool)

    # Seed one archer and attempt call without any authentication
    [archer_id] = await create_archers(db_pool, 1)

    resp = await client.get(f"/api/v0/session/archer/{archer_id}/participating")

    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_open_session_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/archer/{archer_id}/open-session must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """
    await _truncate_all(db_pool)

    # Seed one archer and attempt call without any authentication
    [owner_id] = await create_archers(db_pool, 1)

    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")

    # When auth is enforced in the router/dependency, this should be 401
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_closed_sessions_list_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/archer/{archer_id}/close-session must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """
    await _truncate_all(db_pool)

    # Seed one archer and attempt call without any authentication
    [owner_id] = await create_archers(db_pool, 1)

    resp = await client.get(f"/api/v0/session/archer/{owner_id}/close-session")

    # When auth is enforced in the router/dependency, this should be 401
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_open_session_lifecycle(client: AsyncClient, db_pool: Pool) -> None:
    await _truncate_all(db_pool)

    # Seed one owner archer
    [owner_id] = await create_archers(db_pool, 1)

    # Set auth cookie for owner to pass the new auth requirement
    now = int(time.time())
    token = jwt.encode(
        {
            "sub": str(owner_id),
            "iat": now,
            "exp": now + 3600,
            "iss": "arch-stats",
            "typ": "access",
        },
        settings.jwt_secret,
        algorithm=settings.jwt_algorithm,
    )
    client.cookies.set("arch_stats_auth", token, path="/")

    # Initially: no open session for owner (router returns 200 with null)
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == 200
    assert resp.json()["session_id"] is None

    # Create a session via POST /session
    payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == 201
    session_id = UUID(resp.json()["session_id"])

    # Now owner should have an open session
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == 200
    assert UUID(resp.json()["session_id"]) == session_id

    # There should be one open session in the list
    resp = await client.get("/api/v0/session/open")
    assert resp.status_code == 200
    sessions = resp.json()
    assert len(sessions) == 1
    assert UUID(sessions[0]["session_id"]) == session_id


@pytest.mark.asyncio
async def test_closed_sessions_forbidden_when_not_self(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/archer/{archer_id}/close-session should return 403 when authenticated archer differs.

    Flow:
    - Create two archers (owner and other)
    - Authenticate as other
    - Request owner's closed sessions -> 403
    """
    await _truncate_all(db_pool)

    owner_id, other_id = await create_archers(db_pool, 2)

    # Authenticate as the other archer
    client.cookies.set("arch_stats_auth", _jwt_for(other_id), path="/")

    resp = await client.get(f"/api/v0/session/archer/{owner_id}/close-session")
    assert resp.status_code == 403
    assert resp.json()["detail"] == "Forbidden"


@pytest.mark.asyncio
async def test_closed_sessions_empty_initially(client: AsyncClient, db_pool: Pool) -> None:
    """With no closed sessions, GET /session/archer/{id}/close-session returns empty list."""
    await _truncate_all(db_pool)

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")

    resp = await client.get(f"/api/v0/session/archer/{owner_id}/close-session")
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    assert data == []


@pytest.mark.asyncio
async def test_closed_sessions_after_closing_multiple(client: AsyncClient, db_pool: Pool) -> None:
    """After opening and closing multiple sessions, closed list returns them all sorted by created_at desc (implementation specific order not asserted, but membership is)."""
    await _truncate_all(db_pool)

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")

    # Helper to create and then close a session
    def _make_create_payload() -> dict[str, Any]:
        return {
            "owner_archer_id": str(owner_id),
            "session_location": "Main Range",
            "is_indoor": False,
            "is_opened": True,
        }

    # Create and close first session
    resp = await client.post("/api/v0/session", json=_make_create_payload())
    assert resp.status_code == 201
    s1 = UUID(resp.json()["session_id"])
    resp = await client.patch("/api/v0/session/close", json={"session_id": str(s1)})
    assert resp.status_code == 204

    # Create and close second session
    resp = await client.post("/api/v0/session", json=_make_create_payload())
    assert resp.status_code == 201
    s2 = UUID(resp.json()["session_id"])
    resp = await client.patch("/api/v0/session/close", json={"session_id": str(s2)})
    assert resp.status_code == 204

    # Fetch closed sessions list
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/close-session")
    assert resp.status_code == 200
    items = resp.json()
    assert isinstance(items, list)
    ids = {UUID(it["session_id"]) for it in items}
    assert {s1, s2}.issubset(ids)
    # Assert is_opened is False for all
    assert all(it.get("is_opened") is False for it in items)


@pytest.mark.asyncio
async def test_participant_not_participating_initially(client: AsyncClient, db_pool: Pool) -> None:
    await _truncate_all(db_pool)

    _, participant_id = await create_archers(db_pool, 2)

    # Authenticate as participant and check participating state
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    resp = await client.get(f"/api/v0/session/archer/{participant_id}/participating")
    assert resp.status_code == 200
    assert resp.json()["session_id"] is None


@pytest.mark.asyncio
async def test_join_session_assigns_slot_and_marks_participating(
    client: AsyncClient, db_pool: Pool
) -> None:
    await _truncate_all(db_pool)

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": True,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == 201
    session_id = UUID(resp.json()["session_id"])

    # Participant joins via POST /session/slot
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": str(session_id),
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == 200

    data = resp.json()
    assert UUID(data["target_id"])
    assert "slot" in data and isinstance(data["slot"], str)
    lane = data["slot"][:-1]
    letter = data["slot"][-1]
    try:
        result = int(lane)
        assert isinstance(result, int) and int(lane) > 0
    except ValueError:
        assert (
            False
        ), f"The value '{lane}' could not be converted to an integer, but it should have."
    assert isinstance(letter, str) and letter in ("A", "B", "C", "D")

    # Participant should now be participating
    resp = await client.get(f"/api/v0/session/archer/{participant_id}/participating")
    assert resp.status_code == 200
    assert UUID(resp.json()["session_id"]) == session_id


@pytest.mark.asyncio
async def test_leave_session_clears_participation(client: AsyncClient, db_pool: Pool) -> None:
    await _truncate_all(db_pool)

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": True,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == 201
    session_id = UUID(resp.json()["session_id"])

    # Participant joins
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": str(session_id),
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == 200

    # Participant leaves
    resp = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": str(session_id), "archer_id": str(participant_id)},
    )
    assert resp.status_code == 200
    assert resp.content == b""

    # Participant should no longer be participating
    resp = await client.get(f"/api/v0/session/archer/{participant_id}/participating")
    assert resp.status_code == 200
    assert resp.json()["session_id"] is None


@pytest.mark.asyncio
async def test_close_session_rules(client: AsyncClient, db_pool: Pool) -> None:
    await _truncate_all(db_pool)

    owner_id, p1 = await create_archers(db_pool, 2)

    # Open session (authenticate as owner)
    now = int(time.time())
    owner_token = jwt.encode(
        {
            "sub": str(owner_id),
            "iat": now,
            "exp": now + 3600,
            "iss": "arch-stats",
            "typ": "access",
        },
        settings.jwt_secret,
        algorithm=settings.jwt_algorithm,
    )
    client.cookies.set("arch_stats_auth", owner_token, path="/")
    payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Outdoor Field",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == 201
    session_id = UUID(resp.json()["session_id"])

    # Join participant so closing should fail until they leave
    join_payload = {
        "session_id": str(session_id),
        "archer_id": str(p1),
        "distance": 70,
        "face_type": "122cm_full",
        "bowstyle": "compound",
        "draw_weight": 52.3,
    }
    # Authenticate as participant for joining
    participant_now = int(time.time())
    participant_token = jwt.encode(
        {
            "sub": str(p1),
            "iat": participant_now,
            "exp": participant_now + 3600,
            "iss": "arch-stats",
            "typ": "access",
        },
        settings.jwt_secret,
        algorithm=settings.jwt_algorithm,
    )
    client.cookies.set("arch_stats_auth", participant_token, path="/")
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == 200

    # Attempt to close should be blocked by SessionModel (participants still shooting)
    # Close via PATCH /session/close using SessionUpdate with where.session_id
    # Attempt close as owner (switch back to owner auth)
    client.cookies.set("arch_stats_auth", owner_token, path="/")
    resp = await client.patch(
        "/api/v0/session/close",
        json={"session_id": str(session_id)},
    )
    # Either a 422 with specific error
    assert resp.json()["detail"] == "ERROR: cannot close session with active participants"
    assert resp.status_code == 422

    # Participant leaves
    # Participant leaves (switch to participant auth)
    client.cookies.set("arch_stats_auth", participant_token, path="/")
    resp = await client.patch(
        "/api/v0/session/slot/leave", json={"session_id": str(session_id), "archer_id": str(p1)}
    )

    assert resp.status_code == 200

    # Now owner can close
    # Now owner can close (switch to owner auth)
    client.cookies.set("arch_stats_auth", owner_token, path="/")
    resp = await client.patch(
        "/api/v0/session/close",
        json={
            "session_id": str(session_id),
        },
    )
    assert resp.status_code == 204
    assert resp.content == b""


@pytest.mark.asyncio
async def test_create_second_session_fails_with_422(client: AsyncClient, db_pool: Pool) -> None:
    """Owner cannot open a second session while one is already open.

    Flow:
    - authenticate as owner
    - create first session -> 201
    - attempt second create -> 422 with specific message
    """
    await _truncate_all(db_pool)

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")

    payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }

    # First creation succeeds
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == 201

    # Second creation must fail
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == 409
    assert resp.json()["detail"] == "Archer already have an opened session"


@pytest.mark.asyncio
async def test_create_session_validation_missing_fields(client: AsyncClient, db_pool: Pool) -> None:
    """POST /session returns 422 when required fields are missing.

    Checks that error detail entries include the expected field paths.
    """
    await _truncate_all(db_pool)

    # Authenticate any archer so auth passes
    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")

    # Missing owner_archer_id and is_opened
    bad_payload = {
        "session_location": "Main Range",
        "is_indoor": True,
    }

    resp = await client.post("/api/v0/session", json=bad_payload)
    assert resp.status_code == 422

    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    # Expect both fields to be flagged missing under body
    assert ("body", "owner_archer_id") in locs
    assert ("body", "is_opened") in locs


@pytest.mark.asyncio
async def test_create_session_validation_wrong_types(client: AsyncClient, db_pool: Pool) -> None:
    """POST /session returns 422 when fields have wrong types.

    Checks that error detail entries include the expected field paths.
    """
    await _truncate_all(db_pool)

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")

    # Wrong types for several fields
    bad_payload = {
        "owner_archer_id": 123,  # should be UUID string
        "session_location": 42,  # should be string
        "is_indoor": "notabool",  # should be bool
        "is_opened": "notabool",  # should be bool
    }

    resp = await client.post("/api/v0/session", json=bad_payload)
    assert resp.status_code == 422

    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    # Validate that each problematic field is referenced in errors
    assert ("body", "owner_archer_id") in locs
    assert ("body", "session_location") in locs
    assert ("body", "is_indoor") in locs
    assert ("body", "is_opened") in locs


@pytest.mark.asyncio
async def test_close_already_closed_session_returns_422(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/close should return 422 if the session is already closed."""
    await _truncate_all(db_pool)

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")

    # Create and then close the session
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # First close succeeds
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == 204

    # Second close should fail with 404
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == 404
    print(resp.json()["detail"])
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already closed"


@pytest.mark.asyncio
async def test_close_session_missing_session_id_returns_400(
    client: AsyncClient, db_pool: Pool
) -> None:
    """PATCH /session/close should return 400 when session_id is missing in payload."""
    await _truncate_all(db_pool)

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")

    resp = await client.patch("/api/v0/session/close", json={})
    assert resp.status_code == 400
    assert resp.json()["detail"] == "ERROR: session_id wasn't provided"


@pytest.mark.asyncio
async def test_close_session_invalid_session_id_type_returns_422(
    client: AsyncClient, db_pool: Pool
) -> None:
    """PATCH /session/close should return 422 when session_id has invalid type."""
    await _truncate_all(db_pool)

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")

    # Provide an invalid UUID string to trigger body validation
    resp = await client.patch("/api/v0/session/close", json={"session_id": "not-a-uuid"})
    assert resp.status_code == 422
    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    assert ("body", "session_id") in locs


@pytest.mark.asyncio
async def test_reopen_session_happy_path(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/re-open should re-open a previously closed session.

    Flow:
    - Owner opens a session
    - Owner closes the session
    - Owner re-opens the session (202 Accepted)
    - The session is visible again as open for the owner and in the open list
    """
    await _truncate_all(db_pool)

    # Seed one owner archer and authenticate
    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")

    # Create a session via POST /session
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Verify it's open for the owner
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == 200
    assert resp.json()["session_id"] == session_id

    # Close the session
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == 204

    # After closing, the owner should have no open session
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == 200
    assert resp.json()["session_id"] is None

    # Re-open the session
    resp = await client.patch("/api/v0/session/re-open", json={"session_id": session_id})
    assert resp.status_code == 200
    assert resp.json()["session_id"] == session_id

    # Owner should now report this session as open again
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == 200
    assert resp.json()["session_id"] == session_id

    # And it should appear in the list of open sessions
    resp = await client.get("/api/v0/session/open")
    assert resp.status_code == 200
    sessions = resp.json()
    assert any(s.get("session_id") == session_id for s in sessions)


@pytest.mark.asyncio
async def test_reopen_already_open_session_returns_422(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/re-open should fail with 422 if the session is already open."""
    await _truncate_all(db_pool)

    # Seed one owner archer and authenticate
    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")

    # Create a session via POST /session (it starts open)
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Attempt to re-open an already open session
    resp = await client.patch("/api/v0/session/re-open", json={"session_id": session_id})
    assert resp.status_code == 422
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already open"


@pytest.mark.asyncio
async def test_reopen_session_forbidden_when_not_owner(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/re-open should return 403 if the requester is not the owner."""
    await _truncate_all(db_pool)

    # Seed two archers: owner and stranger
    owner_id, stranger_id = await create_archers(db_pool, 2)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Owner closes the session
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == 204

    # Stranger attempts to re-open
    client.cookies.set("arch_stats_auth", _jwt_for(stranger_id), path="/")
    resp = await client.patch("/api/v0/session/re-open", json={"session_id": session_id})
    assert resp.status_code == 403
    assert resp.json()["detail"] == "Archer is not allow to re-open this session"


@pytest.mark.asyncio
async def test_fifth_archer_joins_new_target_when_first_is_full(
    client: AsyncClient, db_pool: Pool
) -> None:
    """When 4 archers occupy a target at the same distance, the next archer must be
    allocated to a new target (new target_id and next lane starting at 2).

    Flow:
    - Create 6 archers (owner + 5 participants)
    - Owner opens a session
    - First 4 participants at the same distance join and must share the same target (lane 1, slots A-D)
    - The 4th  and 5th participants at the same distance must be allocated to a NEW target (lane 2, slot A & B)
    """

    await _truncate_all(db_pool)

    owner_id, a1, a2, a3, a4, a5 = await create_archers(db_pool, 6)

    # Owner authenticates and creates a session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # First 4 joiners should share the same target_id and occupy letters A-D
    j1 = await join_session(client, session_id, owner_id)
    first_target_id = j1["target_id"]
    assert j1["slot"] == "1A"  # first joiner gets

    j2 = await join_session(client, session_id, a1)
    assert j2["target_id"] == first_target_id
    assert j2["slot"] == "1B"

    j3 = await join_session(client, session_id, a2)
    assert j3["target_id"] == first_target_id
    assert j3["slot"] == "1C"

    j4 = await join_session(client, session_id, a3)
    assert j4["target_id"] == first_target_id
    assert j4["slot"] == "1D"

    # The 5th participant must trigger creation of a new target at the same distance
    j5 = await join_session(client, session_id, a4)
    second_target_id = j5["target_id"]
    assert second_target_id != first_target_id
    # Expect lane to increment to 2 and letter reset to A
    assert j5["slot"] == "2A"

    j6 = await join_session(client, session_id, a5)
    assert j6["target_id"] == second_target_id
    # Expect lane to increment to 2 and letter reset to A
    assert j6["slot"] == "2B"


@pytest.mark.asyncio
async def test_new_targets_created_for_different_distances(
    client: AsyncClient, db_pool: Pool
) -> None:
    """When archers join a session at different distances (20, 30, 40),
    separate targets should be created for each distance. Expect 3 targets
    with lanes 1, 2, 3 respectively, and each assignment should be '1A', '2A', '3A'.
    """

    await _truncate_all(db_pool)

    owner_id, a2, a3 = await create_archers(db_pool, 3)

    # Owner authenticates and creates a session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Owner joins at 20m => lane 1, slot A
    j20 = await join_session(client, session_id, owner_id, 20)
    t1 = j20["target_id"]
    assert j20["slot"] == "1A"

    # Next archer joins at 30m => new target, lane 2, slot A
    j30 = await join_session(client, session_id, a2, 30)
    t2 = j30["target_id"]
    assert t2 != t1
    assert j30["slot"] == "2A"

    # Next archer joins at 40m => new target, lane 3, slot A
    j40 = await join_session(client, session_id, a3, 40)
    t3 = j40["target_id"]
    assert t3 not in {t1, t2}
    assert j40["slot"] == "3A"


@pytest.mark.asyncio
async def test_join_nonexistent_session_returns_422(client: AsyncClient, db_pool: Pool) -> None:
    """Joining a non-existent session should return 422 with a clear message."""
    await _truncate_all(db_pool)

    [archer_id] = await create_archers(db_pool, 1)

    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")

    payload = {
        "session_id": "00000000-0000-0000-0000-000000000000",  # non-existent
        "archer_id": str(archer_id),
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }

    resp = await client.post("/api/v0/session/slot", json=payload)
    assert resp.status_code == 422
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already closed"


@pytest.mark.asyncio
async def test_join_session_validation_wrong_types(client: AsyncClient, db_pool: Pool) -> None:
    """POST /session/slot returns 422 when fields have wrong types.

    We assert that error detail entries include the expected field paths.
    """
    await _truncate_all(db_pool)

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")

    # Wrong types: ensure they are non-coercible to trigger validation errors
    bad_payload = {
        "session_id": 123,  # should be UUID string
        "archer_id": 456,  # should be UUID string
        "distance": "thirty",  # should be int
        "face_type": 789,  # should be enum string
        "is_shooting": "notabool",  # should be bool
        "bowstyle": 101112,  # should be enum string
        "draw_weight": "heavy",  # should be float
    }

    resp = await client.post("/api/v0/session/slot", json=bad_payload)
    assert resp.status_code == 422
    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    assert ("body", "session_id") in locs
    assert ("body", "archer_id") in locs
    assert ("body", "distance") in locs
    assert ("body", "face_type") in locs
    assert ("body", "is_shooting") in locs
    assert ("body", "bowstyle") in locs
    assert ("body", "draw_weight") in locs


@pytest.mark.asyncio
async def test_join_session_validation_missing_fields(client: AsyncClient, db_pool: Pool) -> None:
    """POST /session/slot returns 422 when required fields are missing.

    We assert that error detail entries include the expected field paths.
    """
    await _truncate_all(db_pool)

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")

    # Missing multiple required fields
    bad_payload = {
        # "session_id": omitted
        # "archer_id": omitted
        # "distance": omitted
        "face_type": "60cm_full",
        # "is_shooting": omitted (required in request model)
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }

    resp = await client.post("/api/v0/session/slot", json=bad_payload)
    assert resp.status_code == 422
    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    assert ("body", "session_id") in locs
    assert ("body", "archer_id") in locs
    assert ("body", "distance") in locs


@pytest.mark.asyncio
async def test_cannot_join_second_open_session(client: AsyncClient, db_pool: Pool) -> None:
    """
    An archer already participating in one open session cannot join another.

    Expectation: 400 Bad Request with detail 'ERROR: archer already participating in an open
    session'.
    """
    await _truncate_all(db_pool)

    owner1, owner2, participant = await create_archers(db_pool, 3)

    # Owner1 creates session S1
    client.cookies.set("arch_stats_auth", _jwt_for(owner1), path="/")
    r1 = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner1),
            "session_location": "Range A",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert r1.status_code == 201
    s1 = UUID(r1.json()["session_id"])

    # Participant joins S1
    await join_session(client, s1, participant, 30)

    # Owner2 creates session S2
    client.cookies.set("arch_stats_auth", _jwt_for(owner2), path="/")
    r2 = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner2),
            "session_location": "Range B",
            "is_indoor": True,
            "is_opened": True,
        },
    )
    assert r2.status_code == 201
    s2 = r2.json()["session_id"]

    # Participant attempts to join S2 -> should fail with 409
    client.cookies.set("arch_stats_auth", _jwt_for(participant), path="/")
    resp = await client.post(
        "/api/v0/session/slot",
        json={
            "session_id": s2,
            "archer_id": str(participant),
            "distance": 30,
            "face_type": "60cm_full",
            "is_shooting": True,
            "bowstyle": "recurve",
            "draw_weight": 30.0,
        },
    )

    assert resp.status_code == 409
    assert resp.json()["detail"] == "ERROR: archer already participating in an open session"


@pytest.mark.asyncio
async def test_cannot_join_same_session_twice(client: AsyncClient, db_pool: Pool) -> None:
    """An archer attempting to join the same session twice should be blocked.

    Expectation: 400 Bad Request with detail 'ERROR: archer already joined this session'.
    """
    await _truncate_all(db_pool)

    owner, participant = await create_archers(db_pool, 2)

    # Owner creates session S
    client.cookies.set("arch_stats_auth", _jwt_for(owner), path="/")
    r = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner),
            "session_location": "Range Z",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert r.status_code == 201
    s = r.json()["session_id"]

    # Participant joins S successfully
    client.cookies.set("arch_stats_auth", _jwt_for(participant), path="/")
    first = await client.post(
        "/api/v0/session/slot",
        json={
            "session_id": s,
            "archer_id": str(participant),
            "distance": 30,
            "face_type": "60cm_full",
            "is_shooting": True,
            "bowstyle": "recurve",
            "draw_weight": 30.0,
        },
    )
    assert first.status_code == 200

    # Participant attempts to join S again -> should fail with 409 and specific message
    second = await client.post(
        "/api/v0/session/slot",
        json={
            "session_id": s,
            "archer_id": str(participant),
            "distance": 30,
            "face_type": "60cm_full",
            "is_shooting": True,
            "bowstyle": "recurve",
            "draw_weight": 30.0,
        },
    )
    assert second.status_code == 409
    assert second.json()["detail"] == "ERROR: archer already joined this session"


@pytest.mark.asyncio
async def test_leave_nonexistent_session_returns_422(client: AsyncClient, db_pool: Pool) -> None:
    """Leaving a non-existent session should return 422 with a clear message."""
    await _truncate_all(db_pool)

    [archer_id] = await create_archers(db_pool, 1)

    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")

    payload = {
        "session_id": "00000000-0000-0000-0000-000000000000",  # non-existent
        "archer_id": str(archer_id),
    }

    resp = await client.patch("/api/v0/session/slot/leave", json=payload)
    assert resp.status_code == 422
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already closed"


@pytest.mark.asyncio
async def test_leave_closed_session_returns_422(client: AsyncClient, db_pool: Pool) -> None:
    """Leaving an existing but already-closed session should return 422 with the same message."""
    await _truncate_all(db_pool)

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Participant joins
    await join_session(client, UUID(session_id), participant_id, 30)

    # Participant leaves first so the session can be closed
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    resp = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": session_id, "archer_id": str(participant_id)},
    )
    assert resp.status_code == 200

    # Close the session as owner
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == 204

    # Participant attempts to leave after close -> expect 422 with consolidated message
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    resp = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": session_id, "archer_id": str(participant_id)},
    )
    assert resp.status_code == 422
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already closed"


@pytest.mark.asyncio
async def test_leave_open_session_not_participating_returns_400(
    client: AsyncClient, db_pool: Pool
) -> None:
    """Leaving an open session when the archer is not participating should return 400."""
    await _truncate_all(db_pool)

    owner_id, stranger_id = await create_archers(db_pool, 2)

    # Owner creates an open session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    resp = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Range X",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Stranger (not participating) attempts to leave
    client.cookies.set("arch_stats_auth", _jwt_for(stranger_id), path="/")
    resp = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": session_id, "archer_id": str(stranger_id)},
    )
    assert resp.status_code == 409
    assert resp.json()["detail"] == "ERROR: archer is not participating in this session"


@pytest.mark.asyncio
async def test_leave_session_validation_missing_fields(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/slot/leave returns 422 when required fields are missing.

    We assert that error detail entries include the expected field paths.
    """
    await _truncate_all(db_pool)

    # Authenticate any archer so auth passes
    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")

    # Missing both required fields
    bad_payload: dict[str, str] = {
        # "session_id": omitted
        # "archer_id": omitted
    }

    resp = await client.patch("/api/v0/session/slot/leave", json=bad_payload)
    assert resp.status_code == 422
    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    assert ("body", "session_id") in locs
    assert ("body", "archer_id") in locs


@pytest.mark.asyncio
async def test_leave_session_validation_wrong_types(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/slot/leave returns 422 when fields have wrong types.

    We assert that error detail entries include the expected field paths.
    """
    await _truncate_all(db_pool)

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")

    bad_payload = {
        "session_id": 123,  # should be UUID string
        "archer_id": 456,  # should be UUID string
    }

    resp = await client.patch("/api/v0/session/slot/leave", json=bad_payload)
    assert resp.status_code == 422
    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    assert ("body", "session_id") in locs
    assert ("body", "archer_id") in locs


@pytest.mark.asyncio
async def test_rejoin_session_happy_path(client: AsyncClient, db_pool: Pool) -> None:
    """An archer can re-join a session by reactivating a previous assignment.

    Flow:
    - Owner creates session
    - Participant joins -> gets target/slot
    - Participant leaves -> clears participation
    - Participant re-joins via PATCH /session/slot/re-join -> 200 with target/slot
    """
    await _truncate_all(db_pool)

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    resp = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Main Range",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Participant joins
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": session_id,
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == 200
    first = resp.json()
    assert "target_id" in first and "slot" in first

    # Participant leaves
    resp = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": session_id, "archer_id": str(participant_id)},
    )
    assert resp.status_code == 200

    # Re-join via new endpoint
    resp = await client.patch("/api/v0/session/slot/re-join", json=join_payload)
    assert resp.status_code == 200
    rejoined = resp.json()
    assert rejoined["target_id"] == first["target_id"]
    assert rejoined["slot"] == first["slot"]

    # Participating should now show this session again
    resp = await client.get(f"/api/v0/session/archer/{participant_id}/participating")
    assert resp.status_code == 200
    assert resp.json()["session_id"] == session_id


@pytest.mark.asyncio
async def test_rejoin_nonexistent_session_returns_422(client: AsyncClient, db_pool: Pool) -> None:
    """Re-joining a non-existent session should return 422 with a clear message."""
    await _truncate_all(db_pool)

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")

    payload = {
        "session_id": "00000000-0000-0000-0000-000000000000",  # non-existent
        "archer_id": str(archer_id),
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }

    resp = await client.patch("/api/v0/session/slot/re-join", json=payload)
    assert resp.status_code == 422
    assert (
        resp.json()["detail"]
        == "ERROR: the archer is either not allowed to re-join or they are already in"
    )


@pytest.mark.asyncio
async def test_rejoin_as_another_archer_returns_403(client: AsyncClient, db_pool: Pool) -> None:
    """A user cannot re-join on behalf of another archer.

    Flow:
    - Owner creates a session
    - Archer A joins and leaves (creating an inactive assignment)
    - Archer B attempts to re-join using Archer A's archer_id in payload
    - Expect 403 Forbidden with a clear error message
    """
    await _truncate_all(db_pool)

    owner_id, archer_a, archer_b = await create_archers(db_pool, 3)

    # Owner creates session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    resp = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Main Range",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Archer A joins
    client.cookies.set("arch_stats_auth", _jwt_for(archer_a), path="/")
    join_payload = {
        "session_id": session_id,
        "archer_id": str(archer_a),
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == 200

    # Archer A leaves
    resp = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": session_id, "archer_id": str(archer_a)},
    )
    assert resp.status_code == 200

    # Archer B attempts to re-join for Archer A -> must be forbidden
    client.cookies.set("arch_stats_auth", _jwt_for(archer_b), path="/")
    resp = await client.patch("/api/v0/session/slot/re-join", json=join_payload)
    assert resp.status_code == 403
    assert resp.json()["detail"] == "ERROR: user not allowed to re-join"


@pytest.mark.asyncio
async def test_rejoin_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/slot/re-join must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """
    await _truncate_all(db_pool)

    payload = {
        "session_id": "00000000-0000-0000-0000-000000000000",
        "archer_id": "00000000-0000-0000-0000-000000000000",
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }

    resp = await client.patch("/api/v0/session/slot/re-join", json=payload)
    assert resp.status_code == 401
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"


@pytest.mark.asyncio
async def test_rejoin_closed_session_returns_422(client: AsyncClient, db_pool: Pool) -> None:
    """Re-joining a closed session should return 422 with standardized message."""
    await _truncate_all(db_pool)

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    resp = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Main Range",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Participant joins then leaves
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": session_id,
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    rj = await client.post("/api/v0/session/slot", json=join_payload)
    assert rj.status_code == 200
    rl = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": session_id, "archer_id": str(participant_id)},
    )
    assert rl.status_code == 200

    # Close session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    rc = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert rc.status_code == 204

    # Attempt to re-join closed session
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    resp = await client.patch("/api/v0/session/slot/re-join", json=join_payload)
    assert resp.status_code == 422
    assert (
        resp.json()["detail"]
        == "ERROR: the archer is either not allowed to re-join or they are already in"
    )


@pytest.mark.asyncio
async def test_rejoin_without_leaving_returns_422(client: AsyncClient, db_pool: Pool) -> None:
    """Re-joining while already actively participating should return 422."""
    await _truncate_all(db_pool)

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    resp = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Main Range",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Participant joins
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": session_id,
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    jr = await client.post("/api/v0/session/slot", json=join_payload)
    assert jr.status_code == 200

    # Attempt to re-join without leaving
    resp = await client.patch("/api/v0/session/slot/re-join", json=join_payload)
    assert resp.status_code == 422
    assert (
        resp.json()["detail"]
        == "ERROR: the archer is either not allowed to re-join or they are already in"
    )


@pytest.mark.asyncio
async def test_open_session_for_archer_forbidden_when_not_self(
    client: AsyncClient, db_pool: Pool
) -> None:
    """GET /session/archer/{archer_id}/open-session must return 403 for other users."""
    await _truncate_all(db_pool)

    archer_a, archer_b = await create_archers(db_pool, 2)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_b), path="/")
    resp = await client.get(f"/api/v0/session/archer/{archer_a}/open-session")
    assert resp.status_code == 403
    assert resp.json()["detail"] == "Forbidden"


@pytest.mark.asyncio
async def test_participating_forbidden_when_not_self(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/archer/{archer_id}/participating must return 403 for other users."""
    await _truncate_all(db_pool)

    archer_a, archer_b = await create_archers(db_pool, 2)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_b), path="/")
    resp = await client.get(f"/api/v0/session/archer/{archer_a}/participating")
    assert resp.status_code == 403
    assert resp.json()["detail"] == "Forbidden"


@pytest.mark.asyncio
async def test_reopen_nonexistent_session_returns_422(client: AsyncClient, db_pool: Pool) -> None:
    """Re-opening a non-existent session should return 422 with standardized message."""
    await _truncate_all(db_pool)

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")

    resp = await client.patch(
        "/api/v0/session/re-open",
        json={"session_id": "00000000-0000-0000-0000-000000000000"},
    )
    assert resp.status_code == 422
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already open"


@pytest.mark.asyncio
async def test_open_sessions_list_empty_returns_empty(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/open should return an empty list when there are no open sessions."""
    await _truncate_all(db_pool)

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", _jwt_for(archer_id), path="/")
    resp = await client.get("/api/v0/session/open")
    assert resp.status_code == 200
    assert resp.json() == []


@pytest.mark.asyncio
async def test_letter_reuse_after_leave(client: AsyncClient, db_pool: Pool) -> None:
    """Freed slot letters should be reused for the next joiner at the same target."""
    await _truncate_all(db_pool)

    owner_id, a1, a2, a3 = await create_archers(db_pool, 4)

    # Owner creates session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    resp = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Main Range",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert resp.status_code == 201
    session_id = UUID(resp.json()["session_id"])

    # a1 joins -> 1A, a2 joins -> 1B
    j1 = await join_session(client, session_id, a1, 30)
    assert j1["slot"] == "1A"
    j2 = await join_session(client, session_id, a2, 30)
    assert j2["slot"] == "1B"

    # a1 leaves -> free 'A'
    client.cookies.set("arch_stats_auth", _jwt_for(a1), path="/")
    r = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": str(session_id), "archer_id": str(a1)},
    )
    assert r.status_code == 200

    # a3 joins -> should reuse '1A'
    j3 = await join_session(client, session_id, a3, 30)
    assert j3["slot"] == "1A"


@pytest.mark.asyncio
async def test_leave_same_session_twice_returns_400(client: AsyncClient, db_pool: Pool) -> None:
    """Leaving the same session twice should return 409 for the second attempt (not participating conflict)."""
    await _truncate_all(db_pool)

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    resp = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Main Range",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Participant joins
    await join_session(client, UUID(session_id), participant_id, 30)

    # First leave
    client.cookies.set("arch_stats_auth", _jwt_for(participant_id), path="/")
    r1 = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": session_id, "archer_id": str(participant_id)},
    )
    assert r1.status_code == 200

    # Second leave -> 409
    r2 = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": session_id, "archer_id": str(participant_id)},
    )
    assert r2.status_code == 409
    assert r2.json()["detail"] == "ERROR: archer is not participating in this session"


@pytest.mark.asyncio
async def test_create_session_forbidden_when_not_self(client: AsyncClient, db_pool: Pool) -> None:
    """POST /session must return 403 if an authenticated user tries to open
    a session on behalf of another archer.

    Expectation: 403 Forbidden with detail
    'ERROR: user not allowed to open a session for another archer'.
    """
    await _truncate_all(db_pool)

    archer_a, archer_b = await create_archers(db_pool, 2)

    # Authenticate as archer_b but attempt to create for archer_a
    client.cookies.set("arch_stats_auth", _jwt_for(archer_b), path="/")
    payload = {
        "owner_archer_id": str(archer_a),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == 403
    assert resp.json()["detail"] == "ERROR: user not allowed to open a session for another archer"


@pytest.mark.asyncio
async def test_leave_session_forbidden_when_not_self(client: AsyncClient, db_pool: Pool) -> None:
    """PATCH /session/slot/leave must return 403 if an authenticated user tries
    to leave on behalf of another archer.

    Flow:
    - Owner creates session
    - Participant joins
    - Stranger attempts to leave using participant's archer_id -> 403
    """
    await _truncate_all(db_pool)

    owner_id, participant_id, stranger_id = await create_archers(db_pool, 3)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", _jwt_for(owner_id), path="/")
    c = await client.post(
        "/api/v0/session",
        json={
            "owner_archer_id": str(owner_id),
            "session_location": "Main Range",
            "is_indoor": False,
            "is_opened": True,
        },
    )
    assert c.status_code == 201
    session_id = c.json()["session_id"]

    # Participant joins
    await join_session(client, UUID(session_id), participant_id, 30)

    # Stranger attempts to leave on behalf of participant -> 403
    client.cookies.set("arch_stats_auth", _jwt_for(stranger_id), path="/")
    resp = await client.patch(
        "/api/v0/session/slot/leave",
        json={"session_id": session_id, "archer_id": str(participant_id)},
    )
    assert resp.status_code == 403
    assert resp.json()["detail"] == "ERROR: user not allowed to leave"
