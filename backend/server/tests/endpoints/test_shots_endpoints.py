import datetime
import math
import urllib.parse
from typing import Any
from uuid import uuid4

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from server.models.base_db import DictValues
from server.tests.endpoints.test_arrows_endpoints import (
    create_arrow_payload,
    create_many_arrows,
)


SHOTS_ENDPOINT = "/api/v0/shot"
ARROWS_ENDPOINT = "/api/v0/arrow"


def create_shots_payload(arrow_id: str, **overrides: Any) -> DictValues:
    now = datetime.datetime.now(datetime.timezone.utc)
    data: DictValues = {
        "id": uuid4(),
        "arrow_id": arrow_id,
        "arrow_engage_time": now,
        "arrow_disengage_time": now + datetime.timedelta(seconds=2),
        "arrow_landing_time": now + datetime.timedelta(seconds=4),
        "x_coordinate": 10.1,
        "y_coordinate": 5.3,
    }
    data.update(overrides)
    return data


async def insert_arrow(async_client: AsyncClient) -> str:
    payload = create_arrow_payload()
    resp = await async_client.post(ARROWS_ENDPOINT, json=payload)
    resp_json = resp.json()
    data: str = resp_json["data"]
    assert resp.status_code == 201
    return data


async def insert_shot_direct(db_pool: Pool, shot_row: DictValues) -> None:
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


async def create_many_shots(
    db_pool: Pool, arrows_id: list[str], count: int = 5
) -> list[DictValues]:
    shots = []
    for i in range(count):
        shot_row = create_shots_payload(arrows_id[i])
        await insert_shot_direct(db_pool, shot_row)
        shots.append(shot_row)
    return shots


@pytest.mark.asyncio
async def test_shot_read_and_delete(async_client: AsyncClient, db_pool: Pool) -> None:
    # Insert arrow via API
    arrow_id = await insert_arrow(async_client)
    # Insert shot directly in the DB
    shot_row = create_shots_payload(arrow_id)
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


@pytest.mark.asyncio
async def test_shots_filtering(async_client: AsyncClient, db_pool: Pool) -> None:
    # Create arrows and shots
    arrows = await create_many_arrows(async_client, 5)
    shots = await create_many_shots(db_pool, arrows, 5)

    # --- Filter by arrow_id ---
    arrow_id = arrows[2]
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?arrow_id={arrow_id}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(s["arrow_id"] == arrow_id for s in data)
    assert len(data) >= 1

    # --- Filter by x_coordinate ---
    x_val: float = shots[3]["x_coordinate"]  # type: ignore[assignment]
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?x_coordinate={x_val}")
    resp_json = resp.json()

    assert resp.status_code == 200
    data = resp_json["data"]
    assert all(math.isclose(float(s["x_coordinate"]), x_val, rel_tol=1e-6) for s in data)
    assert len(data) >= 1

    # --- Filter by y_coordinate ---
    y_val: float = shots[4]["y_coordinate"]  # type: ignore[assignment]
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?y_coordinate={y_val}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(math.isclose(float(s["y_coordinate"]), y_val, rel_tol=1e-6) for s in data)
    assert len(data) >= 1

    # --- Filter by arrow_engage_time ---

    dt_iso: str = shots[1]["arrow_engage_time"].isoformat()  # type: ignore[union-attr]
    dt_encoded = urllib.parse.quote(dt_iso, safe="")
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?arrow_engage_time={dt_encoded}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert len(data) == 1
    assert data[0]["arrow_engage_time"].startswith(dt_iso[:19])  # ignore microseconds

    # --- Filter by multiple fields ---
    multi_arrow_id = arrows[1]
    multi_x: float = shots[1]["x_coordinate"]  # type: ignore[assignment]
    resp = await async_client.get(
        f"{SHOTS_ENDPOINT}?arrow_id={multi_arrow_id}&x_coordinate={multi_x}"
    )
    assert resp.status_code == 200
    resp_json = resp.json()
    data = resp_json["data"]
    assert all(
        s["arrow_id"] == multi_arrow_id
        and math.isclose(float(s["x_coordinate"]), multi_x, rel_tol=1e-6)
        for s in data
    )


@pytest.mark.asyncio
async def test_shots_filter_no_match(async_client: AsyncClient) -> None:
    # Non-existent arrow_id
    fake_id = str(uuid4())
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?arrow_id={fake_id}")
    assert resp.status_code == 200
    assert resp.json()["data"] == []
