import math
import urllib.parse
from typing import Any
from uuid import uuid4

import pytest
from httpx import AsyncClient

from server.models.base_db import DictValues
from server.tests.endpoints.test_sessions_endpoints import (
    create_many_sessions,
    create_sessions_payload,
)


TARGETS_ENDPOINT = "/api/v0/target"


def create_targets_payload(session_id: str, **overrides: Any) -> DictValues:
    """
    Generate a valid payload for TargetsCreate.
    Adjust the fields to match your schema exactly!
    """

    data: DictValues = {
        "max_x_coordinate": 122.0,
        "max_y_coordinate": 122.0,
        "radius": [3.0, 6.0, 9.0, 12.0, 15.0],
        "points": [10, 8, 6, 4, 2],
        "height": 140.0,
        "human_identifier": "T1",
        "session_id": session_id,
    }
    data.update(overrides)
    return data


async def insert_session(async_client: AsyncClient) -> str:
    payload = create_sessions_payload()
    resp = await async_client.post("/api/v0/session", json=payload)
    resp_json = resp.json()
    data: str = resp_json["data"]
    assert resp.status_code == 201
    return data


async def create_many_targets(
    async_client: AsyncClient,
    sessions: list[DictValues],
    count: int = 5,
) -> list[DictValues]:
    targets = []
    for i in range(count):
        payload = create_targets_payload(
            session_id=sessions[i]["id"],  # type: ignore[arg-type]
            max_x_coordinate=120.0 + i,
            max_y_coordinate=121.0 + i,
            height=140.0 + i,
            human_identifier=f"target_{i}",
        )
        resp = await async_client.post(TARGETS_ENDPOINT, json=payload)
        assert resp.status_code == 201
        targets.append(payload)
    return targets


@pytest.mark.asyncio
async def test_target_crud_workflow(async_client: AsyncClient) -> None:
    # --- Create a session first ---
    session_id = await insert_session(async_client)

    # --- Create Target ---
    payload = create_targets_payload(session_id=session_id)
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload)
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
            t["human_identifier"] == payload["human_identifier"]
            and t["session_id"] == payload["session_id"]
        ):
            found = t
            break
    assert found is not None
    target_id = found["id"]

    # --- Get Target by session_id ---
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={session_id}")
    resp_json = resp.json()
    assert resp.status_code == 200
    target = resp_json["data"]
    assert target[0]["id"] == target_id

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
    payload = create_targets_payload(str(uuid4()))
    payload.pop(missing_field, None)
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload)
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
    payload = create_targets_payload(str(uuid4()))
    payload["unexpected_field"] = "forbidden"
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_targets_filtering(async_client: AsyncClient) -> None:
    sessions = await create_many_sessions(async_client, 5)
    targets = await create_many_targets(async_client, sessions, 5)

    # --- Filter by session_id ---
    session_id = sessions[2]["id"]
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={session_id}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(t["session_id"] == session_id for t in data)

    # --- Filter by max_x_coordinate (float, use math.isclose) ---
    mx_val: float = targets[4]["max_x_coordinate"]  # type: ignore[assignment]
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?max_x_coordinate={mx_val}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(math.isclose(float(t["max_x_coordinate"]), mx_val, rel_tol=1e-6) for t in data)

    # --- Filter by human_identifier ---
    hid: str = targets[3]["human_identifier"]  # type: ignore[assignment]
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?human_identifier={urllib.parse.quote(hid)}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(t["human_identifier"] == hid for t in data)

    # --- Filter by multiple fields ---
    multi_sid: str = sessions[1]["id"]  # type: ignore[assignment]
    multi_hid: str = targets[4]["human_identifier"]  # type: ignore[assignment]
    resp = await async_client.get(
        f"{TARGETS_ENDPOINT}?session_id={multi_sid}&human_identifier={urllib.parse.quote(multi_hid)}"
    )
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(t["session_id"] == multi_sid and t["human_identifier"] == multi_hid for t in data)


@pytest.mark.asyncio
async def test_targets_filter_no_match(async_client: AsyncClient) -> None:
    fake_uuid = str(uuid4())
    resp = await async_client.get(f"{TARGETS_ENDPOINT}?session_id={fake_uuid}")
    assert resp.status_code == 200
    assert resp.json()["data"] == []
