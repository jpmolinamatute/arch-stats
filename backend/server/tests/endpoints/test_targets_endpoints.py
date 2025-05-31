import math
import urllib.parse
from uuid import uuid4

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from server.tests.factories import create_fake_target, create_many_sessions, create_many_targets


TARGETS_ENDPOINT = "/api/v0/target"


@pytest.mark.asyncio
async def test_target_crud_workflow(async_client: AsyncClient, db_pool: Pool) -> None:
    # --- Create a session first ---
    session = await create_many_sessions(db_pool, 1)
    session_id = session[0].session_id
    # --- Create Target ---
    target = create_fake_target(session_id=session_id)
    payload_dict = target.model_dump(mode="json", by_alias=True)
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload_dict)
    data = resp.json()
    assert resp.status_code == 201
    assert data["code"] == 201
    assert data["errors"] == []

    # If you return the created target's id, extract it here
    # target_id = data["data"]
    # Otherwise, fetch from the list

    # --- Get All Targets ---
    resp = await async_client.get(TARGETS_ENDPOINT)
    assert resp.status_code == 200
    targets = resp.json()["data"]
    # Find our just-created target (by session_id or human_identifier)
    found = None
    for t in targets:
        if (
            t["human_identifier"] == payload_dict["human_identifier"]
            and t["session_id"] == payload_dict["session_id"]
        ):
            found = t
            break
    assert found is not None
    target_id = found["id"]

    # --- Get Target by session_id ---
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={session_id}")
    resp_json = resp.json()
    assert resp.status_code == 200
    target_data = resp_json["data"]
    assert target_data[0]["id"] == target_id

    # --- Update (Patch) Target ---
    patch = {
        "height": 145.0,
        "human_identifier": "T2",
    }
    resp = await async_client.patch(f"{TARGETS_ENDPOINT}/{target_id}", json=patch)
    assert resp.status_code == 202
    assert resp.json()["errors"] == []

    # --- Confirm Update ---
    resp = await async_client.get(f"{TARGETS_ENDPOINT}/{target_id}")
    updated = resp.json()["data"]
    assert updated["height"] == 145.0
    assert updated["human_identifier"] == "T2"

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
        "max_x_coordinate",
        "max_y_coordinate",
        "radius",
        "points",
        "height",
        "human_identifier",
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
async def test_patch_nonexistent_target(async_client: AsyncClient) -> None:
    resp = await async_client.patch(f"{TARGETS_ENDPOINT}/{str(uuid4())}", json={"height": 123.0})
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_delete_nonexistent_target(async_client: AsyncClient) -> None:
    resp = await async_client.delete(f"{TARGETS_ENDPOINT}/{str(uuid4())}")
    assert resp.status_code == 404


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
    sessions = await create_many_sessions(db_pool, 1)
    session_uuid = sessions[0].session_id
    session_id = str(session_uuid)
    targets = await create_many_targets(db_pool, session_uuid, 5)

    # --- Filter by session_id ---

    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={session_id}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(t["session_id"] == session_id for t in data)

    # --- Filter by max_x_coordinate (float, use math.isclose) ---
    mx_val: float = targets[4].max_x_coordinate
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?max_x_coordinate={mx_val}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(math.isclose(float(t["max_x_coordinate"]), mx_val, rel_tol=1e-6) for t in data)

    # --- Filter by human_identifier ---
    hid: str = targets[3].human_identifier
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?human_identifier={urllib.parse.quote(hid)}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(t["human_identifier"] == hid for t in data)

    # --- Filter by multiple fields ---

    multi_hid: str = targets[4].human_identifier
    url = f"{TARGETS_ENDPOINT}?session_id={session_id}&"
    url += f"human_identifier={urllib.parse.quote(multi_hid)}"
    resp = await async_client.get(url)
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(t["session_id"] == session_id and t["human_identifier"] == multi_hid for t in data)


@pytest.mark.asyncio
async def test_targets_filter_no_match(async_client: AsyncClient) -> None:
    fake_uuid = str(uuid4())
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={fake_uuid}")
    assert resp.status_code == 200
    assert resp.json()["data"] == []
