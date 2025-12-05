from collections.abc import Callable
from typing import Any
from uuid import UUID

from httpx import AsyncClient


async def join_session(
    client: AsyncClient,
    session_id: UUID,
    archer_id: UUID,
    jwt_for: Callable[[UUID], str],
    distance: int = 30,
) -> Any:
    client.cookies.set("arch_stats_auth", jwt_for(archer_id), path="/")
    payload = {
        "session_id": str(session_id),
        "archer_id": str(archer_id),
        "distance": distance,
        "face_type": "wa_60cm_full",
        "is_shooting": True,
        "bowstyle": "recurve",
        "draw_weight": 30.0,
    }
    r = await client.post("/api/v0/session/slot", json=payload)
    assert r.status_code == 200
    join_data = r.json()
    # Augment with target_id by querying current slot info
    cs = await client.get(f"/api/v0/session/slot/archer/{archer_id}")
    assert cs.status_code == 200
    join_data["target_id"] = cs.json()["target_id"]
    return join_data
