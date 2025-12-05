"""Endpoint tests for slot-related endpoints (join/leave/re-join)."""

from collections.abc import Callable
from uuid import UUID

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from factories.archer_factory import create_archers
from tests.utils import join_session


@pytest.mark.asyncio
async def test_join_slot_requires_auth(client: AsyncClient) -> None:
    """POST /session/slot must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """

    # Construct a valid SlotJoinRequest payload shape so the request reaches auth
    payload = {
        "session_id": "00000000-0000-0000-0000-000000000000",
        "archer_id": "00000000-0000-0000-0000-000000000000",
        "distance": 30,
        "face_type": "wa_60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }

    resp = await client.post("/api/v0/session/slot", json=payload)
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_leave_slot_requires_auth(client: AsyncClient) -> None:
    """PATCH /session/slot/leave must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """

    # Any UUID path with no auth should trigger 401 via auth dependency
    resp = await client.patch("/api/v0/session/slot/leave/00000000-0000-0000-0000-000000000001")
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_join_session_assigns_slot_and_marks_participating(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": str(session_id),
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "wa_60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == 200
    data = resp.json()
    assert "slot" in data and data["slot"] == "1A"
    assert "slot_id" in data and isinstance(UUID(data["slot_id"]), UUID)
    # # Participant should now be participating
    resp = await client.get(f"/api/v0/session/archer/{participant_id}/participating")
    assert resp.status_code == 200
    assert UUID(resp.json()["session_id"]) == session_id


@pytest.mark.asyncio
async def test_leave_session_clears_participation(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": str(session_id),
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "wa_60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == 200

    # Participant leaves (use slot_id)
    slot_id = resp.json()["slot_id"]
    resp = await client.patch(f"/api/v0/session/slot/leave/{slot_id}")
    assert resp.status_code == 200
    assert resp.content == b""

    # Participant should no longer be participating
    resp = await client.get(f"/api/v0/session/archer/{participant_id}/participating")
    assert resp.status_code == 200
    assert resp.json()["session_id"] is None


@pytest.mark.asyncio
async def test_fifth_archer_joins_new_target_when_first_is_full(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    # pylint: disable=too-many-locals
    """When 4 archers occupy a target at the same distance, the next archer must be
    allocated to a new target (new target_id and next lane starting at 2).

    Flow:
    - Create 6 archers (owner + 5 participants)
    - Owner opens a session
    - First 4 participants at the same distance join and must share the same target
      (lane 1, slots A-D)
    - The 5th and 6th participants at the same distance must be allocated to a NEW target
      (lane 2, slot A & B)
    """

    owner_id, a1, a2, a3, a4, a5 = await create_archers(db_pool, 6)

    # Owner authenticates and creates a session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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
    j1 = await join_session(client, session_id, owner_id, jwt_for)
    first_target_id = j1["target_id"]
    assert j1["slot"] == "1A"  # first joiner gets

    j2 = await join_session(client, session_id, a1, jwt_for)
    assert j2["target_id"] == first_target_id
    assert j2["slot"] == "1B"

    j3 = await join_session(client, session_id, a2, jwt_for)
    assert j3["target_id"] == first_target_id
    assert j3["slot"] == "1C"

    j4 = await join_session(client, session_id, a3, jwt_for)
    assert j4["target_id"] == first_target_id
    assert j4["slot"] == "1D"

    # The 5th participant must trigger creation of a new target at the same distance
    j5 = await join_session(client, session_id, a4, jwt_for)
    second_target_id = j5["target_id"]
    assert second_target_id != first_target_id
    # Expect lane to increment to 2 and letter reset to A
    assert j5["slot"] == "2A"

    j6 = await join_session(client, session_id, a5, jwt_for)
    assert j6["target_id"] == second_target_id
    # Expect lane to increment to 2 and letter reset to A
    assert j6["slot"] == "2B"


@pytest.mark.asyncio
async def test_new_targets_created_for_different_distances(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """When archers join a session at different distances (20, 30, 40),
    separate targets should be created for each distance. Expect 3 targets
    with lanes 1, 2, 3 respectively, and each assignment should be '1A', '2A', '3A'.
    """

    owner_id, a2, a3 = await create_archers(db_pool, 3)

    # Owner authenticates and creates a session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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
    j20 = await join_session(client, session_id, owner_id, jwt_for, 20)
    t1 = j20["target_id"]
    assert j20["slot"] == "1A"

    # Next archer joins at 30m => new target, lane 2, slot A
    j30 = await join_session(client, session_id, a2, jwt_for, 30)
    t2 = j30["target_id"]
    assert t2 != t1
    assert j30["slot"] == "2A"

    # Next archer joins at 40m => new target, lane 3, slot A
    j40 = await join_session(client, session_id, a3, jwt_for, 40)
    t3 = j40["target_id"]
    assert t3 not in {t1, t2}
    assert j40["slot"] == "3A"


@pytest.mark.asyncio
async def test_join_nonexistent_session_returns_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Joining a non-existent session should return 422 with a clear message."""

    [archer_id] = await create_archers(db_pool, 1)

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    payload = {
        "session_id": "00000000-0000-0000-0000-000000000000",  # non-existent
        "archer_id": str(archer_id),
        "distance": 30,
        "face_type": "wa_60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }

    resp = await client.post("/api/v0/session/slot", json=payload)
    assert resp.status_code == 422
    assert resp.json()["detail"] == "ERROR: Session either doesn't exist or it was already closed"


@pytest.mark.asyncio
async def test_join_session_validation_wrong_types(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /session/slot returns 422 when fields have wrong types.

    We assert that error detail entries include the expected field paths.
    """

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

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
async def test_join_session_validation_missing_fields(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """POST /session/slot returns 422 when required fields are missing.

    We assert that error detail entries include the expected field paths.
    """

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Missing multiple required fields
    bad_payload = {
        # "session_id": omitted
        # "archer_id": omitted
        # "distance": omitted
        "face_type": "wa_60cm_full",
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
async def test_cannot_join_second_open_session(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    An archer already participating in one open session cannot join another.

    Expectation: 400 Bad Request with detail 'ERROR: archer already participating in an open
    session'.
    """

    owner1, owner2, participant = await create_archers(db_pool, 3)

    # Owner1 creates session S1
    client.cookies.set("arch_stats_auth", jwt_for(owner1), path="/")
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
    await join_session(client, s1, participant, jwt_for, 30)

    # Owner2 creates session S2
    client.cookies.set("arch_stats_auth", jwt_for(owner2), path="/")
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
    client.cookies.set("arch_stats_auth", jwt_for(participant), path="/")
    resp = await client.post(
        "/api/v0/session/slot",
        json={
            "session_id": s2,
            "archer_id": str(participant),
            "distance": 30,
            "face_type": "wa_60cm_full",
            "is_shooting": True,
            "bowstyle": "recurve",
            "draw_weight": 30.0,
        },
    )

    assert resp.status_code == 409
    assert resp.json()["detail"] == "ERROR: archer already participating in an open session"


@pytest.mark.asyncio
async def test_cannot_join_same_session_twice(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """An archer attempting to join the same session twice should be blocked.

    Expectation: 400 Bad Request with detail 'ERROR: archer already joined this session'.
    """

    owner, participant = await create_archers(db_pool, 2)

    # Owner creates session S
    client.cookies.set("arch_stats_auth", jwt_for(owner), path="/")
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
    client.cookies.set("arch_stats_auth", jwt_for(participant), path="/")
    first = await client.post(
        "/api/v0/session/slot",
        json={
            "session_id": s,
            "archer_id": str(participant),
            "distance": 30,
            "face_type": "wa_60cm_full",
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
            "face_type": "wa_60cm_full",
            "is_shooting": True,
            "bowstyle": "recurve",
            "draw_weight": 30.0,
        },
    )
    assert second.status_code == 409
    assert second.json()["detail"] == "ERROR: archer already joined this session"


@pytest.mark.asyncio
async def test_leave_nonexistent_slot_returns_409(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Leaving a non-existent slot should return 409 not participating."""

    [archer_id] = await create_archers(db_pool, 1)

    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    resp = await client.patch("/api/v0/session/slot/leave/00000000-0000-0000-0000-000000000001")
    assert resp.status_code == 409
    assert resp.json()["detail"] == "ERROR: archer is not participating in this session"


@pytest.mark.asyncio
async def test_leave_closed_session_returns_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Leaving an existing but already-closed session should return 422 with the same message."""

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
    create_payload = {
        "owner_archer_id": str(owner_id),
        "session_location": "Main Range",
        "is_indoor": False,
        "is_opened": True,
    }
    resp = await client.post("/api/v0/session", json=create_payload)
    assert resp.status_code == 201
    session_id = resp.json()["session_id"]

    # Participant joins and capture slot id
    j = await join_session(client, UUID(session_id), participant_id, jwt_for, 30)

    # Participant leaves first so the session can be closed
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    leave_slot_id = j["slot_id"]
    resp = await client.patch(f"/api/v0/session/slot/leave/{leave_slot_id}")
    assert resp.status_code == 200

    # Close the session as owner
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
    resp = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert resp.status_code == 200
    assert resp.json() == {"status": "closed"}

    # Participant attempts to leave after close -> expect 409 not participating
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    resp = await client.patch(f"/api/v0/session/slot/leave/{leave_slot_id}")
    assert resp.status_code == 409
    assert resp.json()["detail"] == "ERROR: archer is not participating in this session"


@pytest.mark.asyncio
async def test_leave_open_session_not_participating_returns_400(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Leaving an open session when the archer is not participating should return 400."""

    owner_id, stranger_id = await create_archers(db_pool, 2)

    # Owner creates an open session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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

    # Stranger (not participating) attempts to leave (bogus slot id)
    client.cookies.set("arch_stats_auth", jwt_for(stranger_id), path="/")
    resp = await client.patch("/api/v0/session/slot/leave/00000000-0000-0000-0000-000000000001")
    assert resp.status_code == 409
    assert resp.json()["detail"] == "ERROR: archer is not participating in this session"


@pytest.mark.asyncio
async def test_leave_session_validation_bad_uuid(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """PATCH /session/slot/leave returns 422 when slot_id path is malformed."""

    # Authenticate any archer so auth passes
    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    resp = await client.patch("/api/v0/session/slot/leave/not-a-uuid")
    # Starlette's UUID path converter does not match invalid UUIDs -> route 404
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_leave_session_validation_wrong_types(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """PATCH /session/slot/leave returns 422 when slot_id type is invalid."""

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    resp = await client.patch("/api/v0/session/slot/leave/123")
    # Non-UUID path won't match route, expect 404
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_rejoin_session_happy_path(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """An archer can re-join a session by reactivating a previous assignment.

    Flow:
    - Owner creates session
    - Participant joins -> gets target/slot
    - Participant leaves -> clears participation
    - Participant re-joins via PATCH /session/slot/re-join -> 200 with target/slot
    """

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": session_id,
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "wa_60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == 200
    first = resp.json()
    assert "slot" in first

    # Participant leaves using slot id
    joined_slot_id = first["slot_id"]
    resp = await client.patch(f"/api/v0/session/slot/leave/{joined_slot_id}")
    assert resp.status_code == 200

    # Re-join via new endpoint (only slot_id is required now)
    resp = await client.patch(f"/api/v0/session/slot/re-join/{first['slot_id']}")
    assert resp.status_code == 200
    rejoined = resp.json()
    assert rejoined["slot"] == first["slot"]

    # Participating should now show this session again
    resp = await client.get(f"/api/v0/session/archer/{participant_id}/participating")
    assert resp.status_code == 200
    assert resp.json()["session_id"] == session_id


@pytest.mark.asyncio
async def test_rejoin_nonexistent_session_returns_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Re-joining a non-existent session should return 422 with a clear message."""

    [archer_id] = await create_archers(db_pool, 1)
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")

    # Non-existent slot_id should yield 422 with standardized message
    resp = await client.patch(
        "/api/v0/session/slot/re-join/00000000-0000-0000-0000-000000000001",
    )
    assert resp.status_code == 422
    assert (
        resp.json()["detail"]
        == "ERROR: the archer is either not allowed to re-join or they are already in"
    )


@pytest.mark.asyncio
async def test_rejoin_as_another_archer_returns_403(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """A user cannot re-join on behalf of another archer.

    Flow:
    - Owner creates a session
    - Archer A joins and leaves (creating an inactive assignment)
    - Archer B attempts to re-join using Archer A's archer_id in payload
    - Expect 403 Forbidden with a clear error message
    """

    owner_id, archer_a, archer_b = await create_archers(db_pool, 3)

    # Owner creates session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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
    client.cookies.set("arch_stats_auth", jwt_for(archer_a), path="/")
    join_payload = {
        "session_id": session_id,
        "archer_id": str(archer_a),
        "distance": 30,
        "face_type": "wa_60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    resp = await client.post("/api/v0/session/slot", json=join_payload)
    assert resp.status_code == 200
    joined_slot_id = resp.json()["slot_id"]

    # Archer A leaves
    resp = await client.patch(f"/api/v0/session/slot/leave/{joined_slot_id}")
    assert resp.status_code == 200

    # Archer B attempts to re-join for Archer A -> must be forbidden
    client.cookies.set("arch_stats_auth", jwt_for(archer_b), path="/")
    resp = await client.patch(f"/api/v0/session/slot/re-join/{joined_slot_id}")
    assert resp.status_code == 403
    assert resp.json()["detail"] == "ERROR: user not allowed to re-join"


@pytest.mark.asyncio
async def test_rejoin_requires_auth(client: AsyncClient) -> None:
    """PATCH /session/slot/re-join must require authentication.

    Expectation: 401 Unauthorized when no auth cookie/token is provided.
    """

    resp = await client.patch(
        "/api/v0/session/slot/re-join/00000000-0000-0000-0000-000000000001",
    )
    assert resp.status_code == 401
    assert resp.json()["detail"] == "User is not authorized to use this endpoint"


@pytest.mark.asyncio
async def test_rejoin_closed_session_returns_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Re-joining a closed session should return 422 with standardized message."""

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": session_id,
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "wa_60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    rj = await client.post("/api/v0/session/slot", json=join_payload)
    assert rj.status_code == 200
    joined_slot_id = rj.json()["slot_id"]
    rl = await client.patch(f"/api/v0/session/slot/leave/{joined_slot_id}")
    assert rl.status_code == 200

    # Close session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
    rc = await client.patch("/api/v0/session/close", json={"session_id": session_id})
    assert rc.status_code == 200
    assert rc.json() == {"status": "closed"}
    rejoin_payload = {"slot_id": rj.json()["slot_id"]}

    # Attempt to re-join closed session
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    resp = await client.patch(f"/api/v0/session/slot/re-join/{rejoin_payload['slot_id']}")
    assert resp.status_code == 422
    assert resp.json()["detail"] == "ERROR: The session doesn't exist or it was already closed"


@pytest.mark.asyncio
async def test_rejoin_without_leaving_returns_422(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Re-joining while already actively participating should return 422."""

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    join_payload = {
        "session_id": session_id,
        "archer_id": str(participant_id),
        "distance": 30,
        "face_type": "wa_60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    jr = await client.post("/api/v0/session/slot", json=join_payload)
    assert jr.status_code == 200
    rejoin_payload = {"slot_id": jr.json()["slot_id"]}
    # Attempt to re-join without leaving
    resp = await client.patch(f"/api/v0/session/slot/re-join/{rejoin_payload['slot_id']}")
    assert resp.status_code == 422
    assert (
        resp.json()["detail"]
        == "ERROR: the archer is either not allowed to re-join or they are already in"
    )


@pytest.mark.asyncio
async def test_letter_reuse_after_leave(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """Freed slot letters should be reused for the next joiner at the same target."""

    owner_id, a1, a2, a3 = await create_archers(db_pool, 4)

    # Owner creates session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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
    j1 = await join_session(client, session_id, a1, jwt_for, 30)
    assert j1["slot"] == "1A"
    j2 = await join_session(client, session_id, a2, jwt_for, 30)
    assert j2["slot"] == "1B"

    # a1 leaves -> free 'A'
    client.cookies.set("arch_stats_auth", jwt_for(a1), path="/")
    r = await client.patch(f"/api/v0/session/slot/leave/{j1['slot_id']}")
    assert r.status_code == 200

    # a3 joins -> should reuse '1A'
    j3 = await join_session(client, session_id, a3, jwt_for, 30)
    assert j3["slot"] == "1A"


@pytest.mark.asyncio
async def test_leave_same_session_twice_returns_400(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """
    Leaving the same session twice should return 409 for the second attempt
    (not participating conflict).
    """

    owner_id, participant_id = await create_archers(db_pool, 2)

    # Owner creates session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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

    # Participant joins and capture slot id
    j = await join_session(client, UUID(session_id), participant_id, jwt_for, 30)

    # First leave using slot id captured from initial join
    client.cookies.set("arch_stats_auth", jwt_for(participant_id), path="/")
    leave_slot_id = j["slot_id"]
    r1 = await client.patch(f"/api/v0/session/slot/leave/{leave_slot_id}")
    assert r1.status_code == 200

    # Second leave -> 409
    r2 = await client.patch(f"/api/v0/session/slot/leave/{leave_slot_id}")
    assert r2.status_code == 409
    assert r2.json()["detail"] == "ERROR: archer is not participating in this session"


@pytest.mark.asyncio
async def test_leave_session_forbidden_when_not_self(
    client: AsyncClient, db_pool: Pool, jwt_for: Callable[[UUID], str]
) -> None:
    """PATCH /session/slot/leave must return 403 if an authenticated user tries
    to leave on behalf of another archer.

    Expectation: 403 Forbidden with detail
    'ERROR: user not allowed to leave'.

    Flow:
    - Owner creates session
    - Participant joins
    - Stranger attempts to leave using participant's archer_id -> 403
    """

    owner_id, participant_id, stranger_id = await create_archers(db_pool, 3)

    # Owner creates a session
    client.cookies.set("arch_stats_auth", jwt_for(owner_id), path="/")
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

    # Participant joins and capture slot id
    pj = await join_session(client, UUID(session_id), participant_id, jwt_for, 30)

    # Stranger attempts to leave on behalf of participant -> 403 using participant's slot_id
    client.cookies.set("arch_stats_auth", jwt_for(stranger_id), path="/")
    resp = await client.patch(f"/api/v0/session/slot/leave/{pj['slot_id']}")
    assert resp.status_code == 403
    assert resp.json()["detail"] == "ERROR: user not allowed to leave"
