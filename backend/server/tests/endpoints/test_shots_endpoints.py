import math
import urllib.parse
from uuid import uuid4

import pytest
from asyncpg import Pool
from httpx import AsyncClient

from server.tests.factories import (
    create_many_arrows,
    create_many_sessions,
    create_many_shots,
)


SHOTS_ENDPOINT = "/api/v0/shot"


@pytest.mark.asyncio
async def test_shot_read_and_delete(async_client: AsyncClient, db_pool: Pool) -> None:
    # Insert arrow via API
    arrows = await create_many_arrows(db_pool, 5)
    session = await create_many_sessions(db_pool, 1)
    arrows_ids = [r.arrow_id for r in arrows]
    session_id = session[0].session_id
    # Insert shot directly in the DB
    shot_row = await create_many_shots(db_pool, arrows_ids, session_id, 5)
    shot_uuid = shot_row[0].shot_id
    shot_id = str(shot_uuid)
    arrow_uuid = shot_row[0].arrow_id
    arrow_id = str(arrow_uuid)

    # --- Get all shots ---
    resp = await async_client.get(SHOTS_ENDPOINT)
    assert resp.status_code == 200
    shots = resp.json()["data"]
    assert any(s["id"] == shot_id for s in shots)

    # --- Get shot by id ---
    resp = await async_client.get(f"{SHOTS_ENDPOINT}/{shot_id}")
    assert resp.status_code == 200
    shot = resp.json()["data"]
    assert shot["id"] == shot_id
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
    arrows = await create_many_arrows(db_pool, 5)
    session = await create_many_sessions(db_pool, 1)
    arrows_ids = [r.arrow_id for r in arrows]
    shots = await create_many_shots(db_pool, arrows_ids, session[0].session_id, 5)

    # --- Filter by arrow_id ---
    arrow_id = str(arrows[2].arrow_id)
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?arrow_id={arrow_id}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(s["arrow_id"] == arrow_id for s in data)
    assert len(data) >= 1

    # --- Filter by x_coordinate ---
    x_val = shots[3].x_coordinate or 0.0
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?x_coordinate={x_val}")

    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(math.isclose(float(s["x_coordinate"]), x_val, rel_tol=1e-6) for s in data)
    assert len(data) >= 1

    # --- Filter by y_coordinate ---
    y_val = shots[4].y_coordinate or 0.0
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?y_coordinate={y_val}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(math.isclose(float(s["y_coordinate"]), y_val, rel_tol=1e-6) for s in data)
    assert len(data) >= 1

    # --- Filter by arrow_engage_time ---

    dt_iso = shots[1].arrow_engage_time.isoformat()
    dt_encoded = urllib.parse.quote(dt_iso, safe="")
    resp = await async_client.get(f"{SHOTS_ENDPOINT}?arrow_engage_time={dt_encoded}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert len(data) == 1
    assert data[0]["arrow_engage_time"].startswith(dt_iso[:19])  # ignore microseconds

    # --- Filter by multiple fields ---
    multi_arrow_id = str(arrows[1].arrow_id)
    multi_x = shots[1].x_coordinate or 0.0
    resp = await async_client.get(
        f"{SHOTS_ENDPOINT}?arrow_id={multi_arrow_id}&x_coordinate={multi_x}"
    )
    assert resp.status_code == 200
    data = resp.json()["data"]
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
