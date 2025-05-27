import datetime
from typing import Any
from uuid import uuid4, UUID

import pytest
from httpx import AsyncClient

SESSIONS_ENDPOINT = "/api/v0/session"


def sessions_payload(**overrides: Any) -> dict[str, object]:
    """
    Default payload for session creation, with overrides for specific test cases.
    """
    data = {
        "is_opened": True,
        "start_time": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "location": "Test Range",
    }
    data.update(overrides)
    return data


@pytest.mark.asyncio
async def test_session_crud_workflow(async_client: AsyncClient) -> None:
    # --- Create Session ---
    payload = sessions_payload()
    resp = await async_client.post(SESSIONS_ENDPOINT, json=payload)
    data = resp.json()
    assert resp.status_code == 200
    assert data["code"] == 200
    assert data["errors"] == []

    session_id = data["data"]
    # Should be a valid UUID
    uuid_obj = UUID(session_id)
    assert str(uuid_obj) == session_id

    # --- Get All Sessions ---
    resp = await async_client.get(SESSIONS_ENDPOINT)
    assert resp.status_code == 200
    sessions = resp.json()["data"]
    assert any(s["id"] == session_id for s in sessions)

    # --- Get One Session ---
    resp = await async_client.get(f"{SESSIONS_ENDPOINT}/{session_id}")
    assert resp.status_code == 200
    session = resp.json()["data"]
    assert session["id"] == session_id

    # --- Update Session ---
    patch = {
        "is_opened": False,
        "end_time": (
            datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(minutes=5)
        ).isoformat(),
        # Add more fields as needed for SessionsUpdate
    }
    resp = await async_client.patch(f"{SESSIONS_ENDPOINT}/{session_id}", json=patch)
    resp_json = resp.json()
    print(resp_json)
    assert resp.status_code == 202
    assert resp_json["errors"] == []

    # --- Confirm Update ---
    resp = await async_client.get(f"{SESSIONS_ENDPOINT}/{session_id}")
    session = resp.json()["data"]
    assert session["is_opened"] is False

    # --- Delete Session ---
    resp = await async_client.delete(f"{SESSIONS_ENDPOINT}/{session_id}")
    assert resp.status_code == 204
    assert resp.json()["errors"] == []

    # --- Confirm Gone ---
    resp = await async_client.get(f"{SESSIONS_ENDPOINT}/{session_id}")
    assert resp.status_code == 404


@pytest.mark.asyncio
@pytest.mark.parametrize("missing_field", ["is_opened", "start_time", "location"])
async def test_session_missing_required_fields(
    async_client: AsyncClient, missing_field: str
) -> None:
    payload = sessions_payload()
    payload.pop(missing_field, None)
    resp = await async_client.post(SESSIONS_ENDPOINT, json=payload)
    assert resp.status_code == 422
    result = resp.json()
    assert "detail" in result
    assert any(missing_field in str(item["loc"]) for item in result["detail"])


@pytest.mark.asyncio
async def test_patch_nonexistent_session(async_client: AsyncClient) -> None:
    resp = await async_client.patch(f"{SESSIONS_ENDPOINT}/{uuid4()}", json={"is_opened": False})
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_delete_nonexistent_session(async_client: AsyncClient) -> None:
    resp = await async_client.delete(f"{SESSIONS_ENDPOINT}/{uuid4()}")
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_get_all_sessions_empty(async_client: AsyncClient) -> None:
    resp = await async_client.get(SESSIONS_ENDPOINT)
    assert resp.status_code == 200
    sessions = resp.json()["data"]
    assert sessions == [] or sessions is None


@pytest.mark.asyncio
async def test_post_session_with_extra_field(async_client: AsyncClient) -> None:
    payload = sessions_payload()
    payload["unexpected_field"] = "forbidden"
    resp = await async_client.post(SESSIONS_ENDPOINT, json=payload)
    assert resp.status_code == 422
