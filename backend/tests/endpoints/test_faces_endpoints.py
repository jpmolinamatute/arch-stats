from uuid import uuid4

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from shared.factories import create_fake_target, create_many_sessions


FACES_TARGET_ENDPOINT = "/api/v0/target"
FACES_ENDPOINT = "/api/v0/face"


@pytest.mark.asyncio
async def test_create_multiple_faces_and_list(async_client: AsyncClient, db_pool: Pool) -> None:
    # Create session + target first
    sessions = await create_many_sessions(db_pool, 1)
    target_payload = create_fake_target(sessions[0].session_id).model_dump(
        mode="json", by_alias=True
    )
    resp = await async_client.post(f"{FACES_TARGET_ENDPOINT}", json=target_payload)
    assert resp.status_code == 201

    # Fetch target id
    resp = await async_client.get(f"{FACES_TARGET_ENDPOINT}?session_id={sessions[0].session_id}")
    target_id = resp.json()["data"][0]["id"]

    faces_payload = [
        {
            "x": 10.0,
            "y": 10.0,
            "human_identifier": "face-A",
            "radii": [4.0, 8.0],
            "points": [10, 9],
        },
        {
            "x": 25.0,
            "y": 25.0,
            "human_identifier": "face-B",
            "radii": [4.0, 8.0, 12.0],
            "points": [10, 9, 8],
        },
    ]
    resp = await async_client.post(
        f"{FACES_TARGET_ENDPOINT}/{target_id}/face",
        json=faces_payload,
    )
    assert resp.status_code == 201
    body = resp.json()
    assert body["code"] == 201
    created_ids = body["data"]
    assert len(created_ids) == 2

    # List faces for target
    resp = await async_client.get(f"{FACES_TARGET_ENDPOINT}/{target_id}/face")
    assert resp.status_code == 200
    faces = resp.json()["data"]
    assert len(faces) == 2
    assert {f["human_identifier"] for f in faces} == {"face-A", "face-B"}


@pytest.mark.asyncio
async def test_list_all_faces_and_filter(async_client: AsyncClient, db_pool: Pool) -> None:
    sessions = await create_many_sessions(db_pool, 1)
    target_payload = create_fake_target(sessions[0].session_id).model_dump(
        mode="json", by_alias=True
    )
    resp = await async_client.post(f"{FACES_TARGET_ENDPOINT}", json=target_payload)
    assert resp.status_code == 201
    resp = await async_client.get(f"{FACES_TARGET_ENDPOINT}?session_id={sessions[0].session_id}")
    target_id = resp.json()["data"][0]["id"]

    # Insert faces
    for hid in ["H1", "H2"]:
        payload = [
            {
                "x": 15.0,
                "y": 15.0,
                "human_identifier": hid,
                "radii": [4.0],
                "points": [10],
            }
        ]
        await async_client.post(f"{FACES_TARGET_ENDPOINT}/{target_id}/face", json=payload)

    # List all
    resp = await async_client.get(FACES_ENDPOINT)
    assert resp.status_code == 200
    all_faces = resp.json()["data"]
    assert len(all_faces) >= 2

    # Filter (by human_identifier via query param not present, so expect 422 if invalid field)
    # FacesFilters only supports id, target_id, human_identifier
    resp = await async_client.get(f"{FACES_ENDPOINT}?human_identifier=H1")
    assert resp.status_code == 200
    filtered = resp.json()["data"]
    assert len(filtered) == 1 and filtered[0]["human_identifier"] == "H1"


@pytest.mark.asyncio
async def test_delete_face(async_client: AsyncClient, db_pool: Pool) -> None:
    sessions = await create_many_sessions(db_pool, 1)
    target_payload = create_fake_target(sessions[0].session_id).model_dump(
        mode="json", by_alias=True
    )
    resp = await async_client.post(f"{FACES_TARGET_ENDPOINT}", json=target_payload)
    assert resp.status_code == 201
    resp = await async_client.get(f"{FACES_TARGET_ENDPOINT}?session_id={sessions[0].session_id}")
    target_id = resp.json()["data"][0]["id"]

    face_payload = [
        {
            "x": 5.0,
            "y": 5.0,
            "human_identifier": "DEL-ME",
            "radii": [4.0],
            "points": [10],
        }
    ]
    resp = await async_client.post(f"{FACES_TARGET_ENDPOINT}/{target_id}/face", json=face_payload)
    face_id = resp.json()["data"][0]

    # Delete
    resp = await async_client.delete(f"{FACES_ENDPOINT}/{face_id}")
    assert resp.status_code == 204
    j = resp.json()
    assert j["code"] == 204 and j["errors"] == []

    # Confirm gone (list faces for target should not include it)
    resp = await async_client.get(f"{FACES_TARGET_ENDPOINT}/{target_id}/face")
    remaining = resp.json()["data"]
    assert all(f["id"] != face_id for f in remaining)


@pytest.mark.asyncio
async def test_create_faces_exceeds_limit(async_client: AsyncClient, db_pool: Pool) -> None:
    sessions = await create_many_sessions(db_pool, 1)
    target_payload = create_fake_target(sessions[0].session_id).model_dump(
        mode="json", by_alias=True
    )
    await async_client.post(f"{FACES_TARGET_ENDPOINT}", json=target_payload)
    resp = await async_client.get(f"{FACES_TARGET_ENDPOINT}?session_id={sessions[0].session_id}")
    target_id = resp.json()["data"][0]["id"]

    too_many = [
        {"x": 1.0, "y": 1.0, "human_identifier": "A", "radii": [4.0], "points": [10]},
        {"x": 2.0, "y": 2.0, "human_identifier": "B", "radii": [4.0], "points": [9]},
        {"x": 3.0, "y": 3.0, "human_identifier": "C", "radii": [4.0], "points": [8]},
        {"x": 4.0, "y": 4.0, "human_identifier": "D", "radii": [4.0], "points": [7]},
    ]
    resp = await async_client.post(f"{FACES_TARGET_ENDPOINT}/{target_id}/face", json=too_many)
    assert resp.status_code == 422  # validation error due to max_length=3


@pytest.mark.asyncio
async def test_delete_face_invalid_uuid(async_client: AsyncClient) -> None:
    resp = await async_client.delete(f"{FACES_ENDPOINT}/not-a-uuid")
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_list_faces_for_target_not_found(async_client: AsyncClient) -> None:
    # Target doesn't exist; listing should just return empty list (no DB FK traversal needed)
    fake_target = uuid4()
    resp = await async_client.get(f"{FACES_TARGET_ENDPOINT}/{fake_target}/face")
    assert resp.status_code == 200
    assert resp.json()["data"] == []
