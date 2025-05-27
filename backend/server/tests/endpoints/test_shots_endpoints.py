import datetime
import random
from typing import Any
from uuid import uuid4

import pytest
from asyncpg import Pool
from httpx import AsyncClient


SHOTS_ENDPOINT = "/api/v0/shot"
ARROWS_ENDPOINT = "/api/v0/arrow"


def create_arrow_payload(**overrides: Any) -> dict[str, object]:
    ran_int = random.randint(1, 100)
    data = {
        "id": str(uuid4()),
        "length": 29.0,
        "human_identifier": f"arrow-{ran_int}",
        "is_programmed": False,
        "label_position": 1.0,
        "weight": 350.0,
        "diameter": 5.7,
        "spine": 400,
    }
    data.update(overrides)
    return data


def create_shot_row(arrow_id: str) -> dict[str, object]:
    now = datetime.datetime.now(datetime.timezone.utc)
    return {
        "id": uuid4(),
        "arrow_id": arrow_id,
        "arrow_engage_time": now,
        "arrow_disengage_time": now + datetime.timedelta(seconds=2),
        "arrow_landing_time": now + datetime.timedelta(seconds=4),
        "x_coordinate": 10.1,
        "y_coordinate": 5.3,
    }


async def insert_arrow(async_client: AsyncClient) -> str:
    payload = create_arrow_payload()
    resp = await async_client.post(ARROWS_ENDPOINT, json=payload)
    resp_json = resp.json()
    print(resp_json)
    data: str = resp_json["data"]
    assert resp.status_code == 201
    return data


async def insert_shot_direct(db_pool: Pool, shot_row: dict[str, object]) -> None:
    # Direct DB insert; adjust as per your pool/library setup
    async with db_pool.acquire() as conn:
        await conn.execute(
            """
            INSERT INTO shots (
                id, arrow_id, arrow_engage_time, arrow_disengage_time, 
                arrow_landing_time, x_coordinate, y_coordinate
            ) VALUES ($1, $2, $3, $4, $5, $6, $7)
            """,
            shot_row["id"],
            shot_row["arrow_id"],
            shot_row["arrow_engage_time"],
            shot_row["arrow_disengage_time"],
            shot_row["arrow_landing_time"],
            shot_row["x_coordinate"],
            shot_row["y_coordinate"],
        )


@pytest.mark.asyncio
async def test_shot_read_and_delete(async_client: AsyncClient, db_pool: Pool) -> None:
    # Insert arrow via API
    arrow_id = await insert_arrow(async_client)
    # Insert shot directly in the DB
    shot_row = create_shot_row(arrow_id)
    await insert_shot_direct(db_pool, shot_row)
    shot_id = shot_row["id"]

    # --- Get all shots ---
    resp = await async_client.get(SHOTS_ENDPOINT)
    assert resp.status_code == 200
    shots = resp.json()["data"]
    assert any(str(s["id"]) == str(shot_id) for s in shots)

    # --- Get shot by id ---
    resp = await async_client.get(f"{SHOTS_ENDPOINT}/{shot_id}")
    assert resp.status_code == 200
    shot = resp.json()["data"]
    assert str(shot["id"]) == str(shot_id)
    assert shot["arrow_id"] == arrow_id

    # --- Filter by arrow_id ---
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?arrow_id={arrow_id}")
    assert resp.status_code == 200
    filtered_shots = resp.json()["data"]
    assert any(str(s["id"]) == str(shot_id) for s in filtered_shots)

    # --- Delete the shot ---
    resp = await async_client.delete(f"{SHOTS_ENDPOINT}/{shot_id}")
    assert resp.status_code == 204

    # --- Confirm gone ---
    resp = await async_client.get(f"{SHOTS_ENDPOINT}/{shot_id}")
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_get_all_shots_empty(async_client: AsyncClient) -> None:
    resp = await async_client.get(SHOTS_ENDPOINT)
    assert resp.status_code == 200
    shots = resp.json()["data"]
    assert shots == [] or shots is None


@pytest.mark.asyncio
async def test_delete_nonexistent_shot(async_client: AsyncClient) -> None:
    resp = await async_client.delete(f"{SHOTS_ENDPOINT}/{str(uuid4())}")
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_get_nonexistent_shot(async_client: AsyncClient) -> None:
    resp = await async_client.get(f"{SHOTS_ENDPOINT}/{str(uuid4())}")
    assert resp.status_code == 404
