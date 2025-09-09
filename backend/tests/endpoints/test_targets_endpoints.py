import math
from uuid import uuid4

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from shared.factories import (
    create_fake_target,
    create_many_sessions,
    insert_targets_db,
)


TARGETS_ENDPOINT = "/api/v0/target"


@pytest.mark.asyncio
async def test_target_crud_workflow(async_client: AsyncClient, db_pool: Pool) -> None:
    # --- Create a session first ---
    session = await create_many_sessions(db_pool, 1)
    # --- Create Target ---
    target = create_fake_target(session_id=session[0].session_id)
    payload_dict = target.model_dump(mode="json", by_alias=True)
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload_dict)
    data = resp.json()
    assert resp.status_code == 201
    assert data["code"] == 201
    assert data["errors"] == []

    # --- Get All Targets ---
    resp = await async_client.get(TARGETS_ENDPOINT)
    assert resp.status_code == 200
    targets = resp.json()["data"]

    found = next(
        (t for t in targets if t["session_id"] == payload_dict["session_id"]),
        None,
    )
    assert found is not None
    target_id = found["id"]

    # --- Get Target by session_id ---
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={session[0].session_id}")
    resp_json = resp.json()
    assert resp.status_code == 200
    target_data = resp_json["data"]
    assert target_data[0]["id"] == target_id

    # --- Delete Target ---
    resp = await async_client.delete(f"{TARGETS_ENDPOINT}/{target_id}")
    assert resp.status_code == 204
    assert resp.json()["errors"] == []

    # --- Confirm Gone ---
    resp = await async_client.get(f"{TARGETS_ENDPOINT}/{target_id}")
    assert resp.status_code == 404


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "missing_field",
    [
        "max_x",
        "max_y",
        "session_id",
    ],
)
async def test_target_missing_required_fields(
    async_client: AsyncClient, missing_field: str
) -> None:
    payload = create_fake_target(uuid4())
    payload_dict = payload.model_dump(mode="json", by_alias=True)
    payload_dict.pop(missing_field, None)
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload_dict)
    assert resp.status_code == 422
    result = resp.json()
    assert "detail" in result
    assert any(missing_field in str(item["loc"]) for item in result["detail"])


@pytest.mark.asyncio
async def test_get_target_by_id(async_client: AsyncClient, db_pool: Pool) -> None:
    sessions = await create_many_sessions(db_pool, 1)
    target = create_fake_target(session_id=sessions[0].session_id)
    payload = target.model_dump(mode="json", by_alias=True)
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload)
    assert resp.status_code == 201

    # fetch created target id via filter
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={sessions[0].session_id}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert len(data) >= 1
    tid = data[0]["id"]

    resp = await async_client.get(f"{TARGETS_ENDPOINT}/{tid}")
    assert resp.status_code == 200
    got = resp.json()["data"]
    assert got["id"] == tid


@pytest.mark.asyncio
async def test_get_target_invalid_uuid(async_client: AsyncClient) -> None:
    resp = await async_client.get(f"{TARGETS_ENDPOINT}/not-a-uuid")
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_delete_nonexistent_target(async_client: AsyncClient) -> None:
    resp = await async_client.delete(f"{TARGETS_ENDPOINT}/{str(uuid4())}")
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_delete_target_invalid_uuid(async_client: AsyncClient) -> None:
    resp = await async_client.delete(f"{TARGETS_ENDPOINT}/not-a-uuid")
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_get_all_targets_empty(async_client: AsyncClient) -> None:
    resp = await async_client.get(TARGETS_ENDPOINT)
    assert resp.status_code == 200
    targets = resp.json()["data"]
    assert targets == [] or targets is None


@pytest.mark.asyncio
async def test_post_target_with_extra_field(async_client: AsyncClient) -> None:
    payload = create_fake_target(uuid4())
    payload_dict = payload.model_dump(mode="json", by_alias=True)
    payload_dict["unexpected_field"] = "forbidden"
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload_dict)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_targets_filtering(async_client: AsyncClient, db_pool: Pool) -> None:
    # Create multiple sessions and one target per session (unique constraint: one target per
    # session)
    sessions = await create_many_sessions(db_pool, 5)
    session_uuid = sessions[0].session_id
    session_id = str(session_uuid)

    # Insert one target per session
    targets = await insert_targets_db(
        db_pool,
        [create_fake_target(s.session_id) for s in sessions],
    )

    # --- Filter by session_id ---

    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={session_id}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(t["session_id"] == session_id for t in data)

    # --- Filter by max_x (float, use math.isclose) ---
    max_x: float = targets[4].max_x
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?max_x={max_x}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(math.isclose(float(t["max_x"]), max_x, rel_tol=1e-6) for t in data)

    # Note: human_identifier filtering now belongs to faces API.
    # Targets API filters only by its own fields.


@pytest.mark.asyncio
async def test_targets_filter_invalid_values(async_client: AsyncClient) -> None:
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id=not-a-uuid")
    assert resp.status_code == 422
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?max_x=not-a-float")
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_patch_nonexistent_target(async_client: AsyncClient, db_pool: Pool) -> None:
    sessions = await create_many_sessions(db_pool, 1)
    # TargetsUpdate requires full shape; reuse a valid create payload
    payload = create_fake_target(session_id=sessions[0].session_id).model_dump(
        mode="json", by_alias=True
    )
    resp = await async_client.patch(f"{TARGETS_ENDPOINT}/{uuid4()}", json=payload)
    # Expected: non-existent id should result in 404 (DBNotFound).
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_patch_target_with_no_fields(async_client: AsyncClient) -> None:
    """TargetsUpdate inherits required fields, so empty body is a 422 validation error."""
    resp = await async_client.patch(f"{TARGETS_ENDPOINT}/{uuid4()}", json={})
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_post_target_wrong_data_types(async_client: AsyncClient) -> None:
    valid = create_fake_target(uuid4()).model_dump(mode="json", by_alias=True)
    # wrong types
    cases = [
        ("max_x", "not-a-float"),
        ("max_y", "not-a-float"),
        ("session_id", "not-a-uuid"),
    ]
    for field, wrong in cases:
        payload = {**valid}
        payload[field] = wrong
        resp = await async_client.post(TARGETS_ENDPOINT, json=payload)
        assert resp.status_code == 422


@pytest.mark.asyncio
async def test_targets_calibrate_endpoint(async_client: AsyncClient) -> None:
    resp = await async_client.get(f"{TARGETS_ENDPOINT}/calibrate")
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 200
    assert body["errors"] == []
    assert body["data"] is not None


@pytest.mark.asyncio
async def test_targets_filter_no_match(async_client: AsyncClient) -> None:
    fake_uuid = str(uuid4())
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={fake_uuid}")
    assert resp.status_code == 200
    assert resp.json()["data"] == []


@pytest.mark.asyncio
async def test_unique_target_per_session_endpoint(async_client: AsyncClient, db_pool: Pool) -> None:
    """Posting a second target for the same session should return 400 (DBException)."""
    sessions = await create_many_sessions(db_pool, 1)
    session_id = sessions[0].session_id
    first = create_fake_target(session_id=session_id).model_dump(mode="json", by_alias=True)
    second = create_fake_target(session_id=session_id).model_dump(mode="json", by_alias=True)
    # First insert succeeds
    resp1 = await async_client.post(TARGETS_ENDPOINT, json=first)
    assert resp1.status_code == 201
    # Second insert should violate UNIQUE(session_id)
    resp2 = await async_client.post(TARGETS_ENDPOINT, json=second)
    assert resp2.status_code == 400
    body = resp2.json()
    assert body["code"] == 400
    assert body["data"] is None
    result = any("unique" in err.lower() or "constraint" in err.lower() for err in body["errors"])
    assert result or body["errors"] != []
