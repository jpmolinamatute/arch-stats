"""Endpoint tests for session-related endpoints.

Covers:
- GET   /session/archer/{archer_id}/open-session
- GET   /session/archer/{archer_id}/participating
- GET   /session/open
- GET   /session/archer/{archer_id}/close-session
- POST  /session
- PATCH /session/close
- PATCH /session/re-open
"""

from collections.abc import Callable
from http import HTTPStatus
from typing import Any
from uuid import UUID

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from factories.archer_factory import create_archers


@pytest.mark.asyncio
async def test_close_session_requires_auth(client: AsyncClient) -> None:
    """PATCH /session/close must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """

    # No auth cookie; even with a dummy session_id, the auth dependency should reject
    payload = {"session_id": "00000000-0000-0000-0000-000000000000"}
    resp = await client.patch("/api/v0/session/close", json=payload)
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == HTTPStatus.UNAUTHORIZED


@pytest.mark.asyncio
async def test_create_session_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """POST /session must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """

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
    assert resp.status_code == HTTPStatus.UNAUTHORIZED


@pytest.mark.asyncio
async def test_open_sessions_list_requires_auth(client: AsyncClient) -> None:
    """GET /session/open must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """

    resp = await client.get("/api/v0/session/open")
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == HTTPStatus.UNAUTHORIZED


@pytest.mark.asyncio
async def test_participating_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/archer/{archer_id}/participating must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """

    # Seed one archer and attempt call without any authentication
    [archer_id] = await create_archers(db_pool, 1)

    resp = await client.get(f"/api/v0/session/archer/{archer_id}/participating")

    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == HTTPStatus.UNAUTHORIZED


@pytest.mark.asyncio
async def test_open_session_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/archer/{archer_id}/open-session must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """

    # Seed one archer and attempt call without any authentication
    [owner_id] = await create_archers(db_pool, 1)

    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")

    # When auth is enforced in the router/dependency, this should be 401
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == HTTPStatus.UNAUTHORIZED


@pytest.mark.asyncio
async def test_closed_sessions_list_requires_auth(client: AsyncClient, db_pool: Pool) -> None:
    """GET /session/archer/{archer_id}/close-session must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """

    # Seed one archer and attempt call without any authentication
    [owner_id] = await create_archers(db_pool, 1)

    resp = await client.get(f"/api/v0/session/archer/{owner_id}/close-session")

    # When auth is enforced in the router/dependency, this should be 401
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == HTTPStatus.UNAUTHORIZED


@pytest.mark.asyncio
async def test_open_session_lifecycle(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    # Seed one owner archer
    [owner_id] = await create_archers(db_pool, 1)

    # Set auth cookie for owner to pass the new auth requirement

    token = jwt_for(owner_id)
    client.cookies.set("arch_stats_auth", token, path="/")

    # Initially: no open session for owner (router returns 200 with null)
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == HTTPStatus.OK
    assert resp.json()["session_id"] is None

    # Create a session via POST /session
    payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == HTTPStatus.CREATED
    session_id = UUID(resp.json()["session_id"])

    # Now owner should have an open session
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == HTTPStatus.OK
    assert UUID(resp.json()["session_id"]) == session_id

    # There should be one open session in the list
    resp = await client.get("/api/v0/session/open")
    assert resp.status_code == HTTPStatus.OK
    sessions = resp.json()
    assert len(sessions) == 1
    assert UUID(sessions[0]["session_id"]) == session_id


@pytest.mark.asyncio
async def test_closed_sessions_forbidden_when_not_self(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    GET /session/archer/{archer_id}/close-session should return 403 when authenticated archer
    differs.

    Flow:
    - Create two archers (owner and other)
    - Authenticate as other
    - Request owner's closed sessions -> 403
    """

    owner_id, other_id = await create_archers(db_pool, 2)

    # Authenticate as the other archer
    client.cookies.set("arch_stats_auth", jwt_for(other_id), path="/")

    resp = await client.get(f"/api/v0/session/archer/{owner_id}/close-session")
    assert resp.status_code == HTTPStatus.FORBIDDEN
    assert resp.json()["detail"] == "Forbidden"


@pytest.mark.asyncio
async def test_closed_sessions_empty_initially(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """With no closed sessions, GET /session/archer/{id}/close-session returns empty list."""

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

    resp = await client.get(f"/api/v0/session/archer/{owner_id}/close-session")
    assert resp.status_code == HTTPStatus.OK
    data = resp.json()
    assert isinstance(data, list)
    assert data == []


@pytest.mark.asyncio
async def test_closed_sessions_after_closing_multiple(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    After opening and closing multiple sessions, closed list returns them all sorted by
    created_at desc (implementation specific order not asserted, but membership is)."""

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

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
    assert resp.status_code == HTTPStatus.CREATED
    s1 = UUID(resp.json()["session_id"])
    resp = await client.patch("/api/v0/session/close", json={"session_id": str(s1)})
    assert resp.status_code == HTTPStatus.OK
    assert resp.json() == {"status": "closed"}

    # Create and close second session
    resp = await client.post("/api/v0/session", json=_make_create_payload())
    assert resp.status_code == HTTPStatus.CREATED
    s2 = UUID(resp.json()["session_id"])
    resp = await client.patch("/api/v0/session/close", json={"session_id": str(s2)})
    assert resp.status_code == HTTPStatus.OK
    assert resp.json() == {"status": "closed"}

    # Fetch closed sessions list
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/close-session")
    assert resp.status_code == HTTPStatus.OK
    items = resp.json()
    assert isinstance(items, list)
    ids = {UUID(it["session_id"]) for it in items}
    assert {s1, s2}.issubset(ids)
    # Assert is_opened is False for all
    assert all(it.get("is_opened") is False for it in items)


@pytest.mark.asyncio
async def test_participant_not_participating_initially(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    _, participant_id = await create_archers(db_pool, 2)

    # Authenticate as participant and check participating state
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    resp = await client.get(f"/api/v0/session/archer/{participant_id}/participating")
    assert resp.status_code == HTTPStatus.OK
    assert resp.json()["session_id"] is None


@pytest.mark.asyncio
async def test_close_session_rules(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    owner_id, p1 = await create_archers(db_pool, 2)
    owner_token = jwt_for(owner_id)
    client.cookies.set("arch_stats_auth", owner_token, path="/")
    payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Outdoor Field",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == HTTPStatus.CREATED
    session_id = UUID(resp.json()["session_id"])

    # Join participant so closing should fail until they leave
    join_payload = {
        "session_id": str(session_id),
        "archer_id": str(p1),
        "distance": 70,
        "face_type": "wa_122cm_full",
        "bowstyle": "compound",
        "draw_weight": 52.3,
    }
    # Authenticate as participant for joining
    participant_token = jwt_for(p1)
    client.cookies.set("arch_stats_auth", participant_token, path="/")
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == HTTPStatus.OK
    initial_slot_id = resp.json()["slot_id"]

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
    assert resp.status_code == HTTPStatus.UNPROCESSABLE_ENTITY

    # Participant leaves (switch to participant auth)
    client.cookies.set("arch_stats_auth", participant_token, path="/")
    # leave using the slot id from the initial join
    resp = await client.patch(f"/api/v0/session/slot/leave/{initial_slot_id}")
    assert resp.status_code == HTTPStatus.OK

    # Now owner can close
    # Now owner can close (switch to owner auth)
    client.cookies.set("arch_stats_auth", owner_token, path="/")
    resp = await client.patch(
        "/api/v0/session/close",
        json={
            "session_id": str(session_id),
        },
    )
    assert resp.status_code == HTTPStatus.OK
    assert resp.json() == {"status": "closed"}


@pytest.mark.asyncio
async def test_create_second_session_fails_with_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Owner cannot open a second session while one is already open.

    Flow:
    - authenticate as owner
    - create first session -> 201
    - attempt second create -> 422 with specific message
    """

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

    payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }

    # First creation succeeds
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == HTTPStatus.CREATED

    # Second creation must fail
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == HTTPStatus.CONFLICT
    assert resp.json()["detail"] == "Archer already has an opened session"


@pytest.mark.asyncio
async def test_create_session_validation_missing_fields(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /session returns 422 when required fields are missing.

    Checks that error detail entries include the expected field paths.
    """

    # Authenticate any archer so auth passes
    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Missing owner_archer_id and is_opened
    bad_payload = {
        "session_location": "Main Range",
        "is_indoor": True,
    }

    resp = await client.post("/api/v0/session", json=bad_payload)
    assert resp.status_code == HTTPStatus.UNPROCESSABLE_ENTITY

    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    # Expect both fields to be flagged missing under body
    assert ("body", "owner_archer_id") in locs
    assert ("body", "is_opened") in locs


@pytest.mark.asyncio
async def test_create_session_validation_wrong_types(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /session returns 422 when fields have wrong types.

    Checks that error detail entries include the expected field paths.
    """

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Wrong types for several fields
    bad_payload = {
        "owner_archer_id": 123,  # should be UUID string
        "session_location": 42,  # should be string
        "is_indoor": "not_bool",  # should be bool
        "is_opened": "not_bool",  # should be bool
    }

    resp = await client.post("/api/v0/session", json=bad_payload)
    assert resp.status_code == HTTPStatus.UNPROCESSABLE_ENTITY

    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    # Validate that each problematic field is referenced in errors
    assert ("body", "owner_archer_id") in locs
    assert ("body", "session_location") in locs
    assert ("body", "is_indoor") in locs
    assert ("body", "is_opened") in locs


@pytest.mark.asyncio
async def test_close_already_closed_session_returns_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """PATCH /session/close should return 422 if the session is already closed."""

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

    # Create and then close the session
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == HTTPStatus.CREATED
    session_id = resp.json()["session_id"]

    # First close succeeds
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == HTTPStatus.OK
    assert resp.json() == {"status": "closed"}

    # Second close should fail with 404
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == HTTPStatus.NOT_FOUND
    print(resp.json()["detail"])
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already closed"


@pytest.mark.asyncio
async def test_close_session_missing_session_id_returns_400(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """PATCH /session/close should return 400 when session_id is missing in payload."""

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

    resp = await client.patch("/api/v0/session/close", json={})
    assert resp.status_code == HTTPStatus.BAD_REQUEST
    assert resp.json()["detail"] == "ERROR: session_id wasn't provided"


@pytest.mark.asyncio
async def test_close_session_invalid_session_id_type_returns_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """PATCH /session/close should return 422 when session_id has invalid type."""

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

    # Provide an invalid UUID string to trigger body validation
    resp = await client.patch("/api/v0/session/close", json={"session_id": "not-a-uuid"})
    assert resp.status_code == HTTPStatus.UNPROCESSABLE_ENTITY
    detail = resp.json().get("detail", [])
    locs = [tuple(err.get("loc", ())) for err in detail]
    assert ("body", "session_id") in locs


@pytest.mark.asyncio
async def test_reopen_session_happy_path(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """PATCH /session/re-open should re-open a previously closed session.

    Flow:
    - Owner opens a session
    - Owner closes the session
    - Owner re-opens the session (202 Accepted)
    - The session is visible again as open for the owner and in the open list
    """

    # Seed one owner archer and authenticate
    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

    # Create a session via POST /session
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == HTTPStatus.CREATED
    session_id = resp.json()["session_id"]

    # Verify it's open for the owner
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == HTTPStatus.OK
    assert resp.json()["session_id"] == session_id

    # Close the session
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == HTTPStatus.OK
    assert resp.json() == {"status": "closed"}

    # After closing, the owner should have no open session
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == HTTPStatus.OK
    assert resp.json()["session_id"] is None

    # Re-open the session
    resp = await client.patch("/api/v0/session/re-open", json={"session_id": session_id})
    assert resp.status_code == HTTPStatus.OK
    assert resp.json()["session_id"] == session_id

    # Owner should now report this session as open again
    resp = await client.get(f"/api/v0/session/archer/{owner_id}/open-session")
    assert resp.status_code == HTTPStatus.OK
    assert resp.json()["session_id"] == session_id

    # And it should appear in the list of open sessions
    resp = await client.get("/api/v0/session/open")
    assert resp.status_code == HTTPStatus.OK
    sessions = resp.json()
    assert any(s.get("session_id") == session_id for s in sessions)


@pytest.mark.asyncio
async def test_reopen_already_open_session_returns_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """PATCH /session/re-open should fail with 422 if the session is already open."""

    # Seed one owner archer and authenticate
    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

    # Create a session via POST /session (it starts open)
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == HTTPStatus.CREATED
    session_id = resp.json()["session_id"]

    # Attempt to re-open an already open session
    resp = await client.patch("/api/v0/session/re-open", json={"session_id": session_id})
    assert resp.status_code == HTTPStatus.UNPROCESSABLE_ENTITY
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already open"


@pytest.mark.asyncio
async def test_reopen_session_forbidden_when_not_owner(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """PATCH /session/re-open should return 403 if the requester is not the owner."""

    # Seed two archers: owner and stranger
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

    # Owner closes the session
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == HTTPStatus.OK
    assert resp.json() == {"status": "closed"}

    # Stranger attempts to re-open
    client.cookies.set("arch_stats_auth", jwt_for(stranger_id), path="/")
    resp = await client.patch("/api/v0/session/re-open", json={"session_id": session_id})
    assert resp.status_code == HTTPStatus.FORBIDDEN
    assert resp.json()["detail"] == "Archer is not allow to re-open this session"


@pytest.mark.asyncio
async def test_open_session_for_archer_forbidden_when_not_self(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """GET /session/archer/{archer_id}/open-session must return 403 for other users."""

    archer_a, archer_b = await create_archers(db_pool, 2)
    client.cookies.set("arch_stats_auth", jwt_for(archer_b), path="/")
    resp = await client.get(f"/api/v0/session/archer/{archer_a}/open-session")
    assert resp.status_code == HTTPStatus.FORBIDDEN
    assert resp.json()["detail"] == "Forbidden"


@pytest.mark.asyncio
async def test_participating_forbidden_when_not_self(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """GET /session/archer/{archer_id}/participating must return 403 for other users."""

    archer_a, archer_b = await create_archers(db_pool, 2)
    client.cookies.set("arch_stats_auth", jwt_for(archer_b), path="/")
    resp = await client.get(f"/api/v0/session/archer/{archer_a}/participating")
    assert resp.status_code == HTTPStatus.FORBIDDEN
    assert resp.json()["detail"] == "Forbidden"


@pytest.mark.asyncio
async def test_reopen_nonexistent_session_returns_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Re-opening a non-existent session should return 422 with standardized message."""

    [owner_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")

    resp = await client.patch(
        "/api/v0/session/re-open",
        json={"session_id": "00000000-0000-0000-0000-000000000000"},
    )
    assert resp.status_code == HTTPStatus.UNPROCESSABLE_ENTITY
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already open"


@pytest.mark.asyncio
async def test_open_sessions_list_empty_returns_empty(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """GET /session/open should return an empty list when there are no open sessions."""

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")
    resp = await client.get("/api/v0/session/open")
    assert resp.status_code == HTTPStatus.OK
    assert resp.json() == []


@pytest.mark.asyncio
async def test_create_session_forbidden_when_not_self(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /session must return 403 if an authenticated user tries to open
    a session on behalf of another archer.

    Expectation: 403 Forbidden with detail
    'ERROR: user not allowed to open a session for another archer'.
    """

    archer_a, archer_b = await create_archers(db_pool, 2)

    # Authenticate as archer_b but attempt to create for archer_a
    client.cookies.set("arch_stats_auth", jwt_for(archer_b), path="/")
    payload = {
        "owner_archer_id": str(archer_a),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=payload)
    assert resp.status_code == HTTPStatus.FORBIDDEN
    assert resp.json()["detail"] == "ERROR: user not allowed to open a session for another archer"
