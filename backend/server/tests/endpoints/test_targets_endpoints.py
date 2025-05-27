import datetime
from typing import Any
from uuid import uuid4

import pytest
from httpx import AsyncClient


TARGETS_ENDPOINT = "/api/v0/target"


def targets_payload(**overrides: Any) -> dict[str, object]:
    """
    Generate a valid payload for TargetsCreate.
    Adjust the fields to match your schema exactly!
    """
    # Example fields (adjust as needed):
    data = {
        "max_x_coordinate": 122.0,
        "max_y_coordinate": 122.0,
        "radius": [3.0, 6.0, 9.0, 12.0, 15.0],
        "points": [10, 8, 6, 4, 2],
        "height": 140.0,
        "human_identifier": "T1",
        "session_id": str(uuid4()),
    }
    data.update(overrides)
    return data


async def create_session(async_client: AsyncClient) -> str:
    payload = {
        "is_opened": True,
        "start_time": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "location": "Test Range",
    }
    resp = await async_client.post("/api/v0/session", json=payload)
    assert resp.status_code == 200
    session_id: str = resp.json()["data"]
    return session_id


@pytest.mark.asyncio
async def test_target_crud_workflow(async_client: AsyncClient) -> None:
    # --- Create a session first ---
    session_id = await create_session(async_client)

    # --- Create Target ---
    payload = targets_payload(session_id=session_id)
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload)
    data = resp.json()
    assert resp.status_code == 200
    assert data["code"] == 200
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
    payload = targets_payload()
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
    payload = targets_payload()
    payload["unexpected_field"] = "forbidden"
    resp = await async_client.post(TARGETS_ENDPOINT, json=payload)
    assert resp.status_code == 422
