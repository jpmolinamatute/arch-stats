import datetime
import urllib.parse
from uuid import UUID, uuid4

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from tests.factories import create_fake_session, create_many_sessions


SESSIONS_ENDPOINT = "/api/v0/session"


@pytest.mark.asyncio
async def test_session_crud_workflow(async_client: AsyncClient) -> None:
    # --- Create Session ---
    payload = create_fake_session()
    payload_dict = payload.model_dump(mode="json", by_alias=True)
    resp = await async_client.post(SESSIONS_ENDPOINT, json=payload_dict)
    data = resp.json()
    assert resp.status_code == 201
    assert data["code"] == 201
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
    payload = create_fake_session()
    payload_dict = payload.model_dump(mode="json", by_alias=True)
    payload_dict.pop(missing_field, None)
    resp = await async_client.post(SESSIONS_ENDPOINT, json=payload_dict)
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
    payload = create_fake_session()
    payload_dict = payload.model_dump(mode="json", by_alias=True)
    payload_dict["unexpected_field"] = "forbidden"
    resp = await async_client.post(SESSIONS_ENDPOINT, json=payload_dict)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_sessions_filtering(async_client: AsyncClient, db_pool: Pool) -> None:
    # Create multiple sessions
    sessions = await create_many_sessions(db_pool, 6)

    # --- Filter by is_opened ---
    resp = await async_client.get(f"{SESSIONS_ENDPOINT}?is_opened=true")
    resp_json = resp.json()
    assert resp.status_code == 200
    data = resp_json["data"]
    assert all(s["is_opened"] is True for s in data)

    resp = await async_client.get(f"{SESSIONS_ENDPOINT}?is_opened=false")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(s["is_opened"] is False for s in data)

    # --- Filter by location ---
    location = "Range_1"
    resp = await async_client.get(f"{SESSIONS_ENDPOINT}?location={location}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(s["location"] == location for s in data)

    # --- Filter by specific start_time (should be 1 match) ---
    target_time = sessions[2].start_time.isoformat()
    encoded_time = urllib.parse.quote(target_time, safe="")
    resp = await async_client.get(f"{SESSIONS_ENDPOINT}?start_time={encoded_time}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert len(data) == 1
    assert data[0]["start_time"].startswith(target_time[:19])  # Ignore microseconds

    # --- Filter by multiple fields ---
    resp = await async_client.get(f"{SESSIONS_ENDPOINT}?location=Range_1&is_opened=true")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(s["location"] == "Range_1" and s["is_opened"] is True for s in data)


@pytest.mark.asyncio
async def test_sessions_filter_no_match(async_client: AsyncClient) -> None:
    resp = await async_client.get(f"{SESSIONS_ENDPOINT}?location=DoesNotExist")
    assert resp.status_code == 200
    assert resp.json()["data"] == []
