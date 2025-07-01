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

    found = None
    for t in targets:
        if t["session_id"] == payload_dict["session_id"]:
            for face in t["faces"]:
                for face_payload in payload_dict["faces"]:
                    if face["human_identifier"] == face_payload["human_identifier"]:
                        found = t
                        break
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
        "faces",
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

    # --- Filter by max_x (float, use math.isclose) ---
    max_x: float = targets[4].max_x
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?max_x={str(max_x)}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    print(f"{type(data[0]['max_x'])}    {type(max_x)}")
    print(f"{data[0]['max_x']=}    {max_x=}")
    assert all(math.isclose(float(t["max_x"]), max_x, rel_tol=1e-6) for t in data)

    # --- Filter by human_identifier ---
    hid: str = targets[3].faces[0].human_identifier
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?human_identifier={urllib.parse.quote(hid)}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(any(face["human_identifier"] == hid for face in t["faces"]) for t in data)

    # --- Filter by multiple fields ---

    multi_hid: str = targets[4].faces[0].human_identifier
    url = f"{TARGETS_ENDPOINT}?session_id={session_id}&"
    url += f"human_identifier={urllib.parse.quote(multi_hid)}"
    resp = await async_client.get(url)
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(
        t["session_id"] == session_id
        and any(face["human_identifier"] == multi_hid for face in t["faces"])
        for t in data
    )


@pytest.mark.asyncio
async def test_targets_filter_no_match(async_client: AsyncClient) -> None:
    fake_uuid = str(uuid4())
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={fake_uuid}")
    assert resp.status_code == 200
    assert resp.json()["data"] == []
