from datetime import datetime, timezone, timedelta
from typing import Any

from httpx import AsyncClient

from server.schema import SessionsCreate, SessionsRead


SESSIONS_ENDPOINT = "/api/v0/session"


def create_fake_sessions(**overrides: Any) -> SessionsCreate:
    """
    Default payload for session creation, with overrides for specific test cases.
    """
    data = SessionsCreate(
        is_opened=True,
        start_time=datetime.now(timezone.utc),
        location="Test Range",
    )
    return data.model_copy(update=overrides)


async def create_many_sessions(async_client: AsyncClient, count: int = 5) -> list[SessionsRead]:
    sessions = []
    for i in range(count):
        is_opened = bool(i % 2)
        start_time = datetime.now(timezone.utc) + timedelta(days=i)
        session = create_fake_sessions(
            location=f"Range_{i%2}",
            start_time=start_time,
        )
        payload_dict = session.model_dump(mode="json", by_alias=True)
        resp = await async_client.post(SESSIONS_ENDPOINT, json=payload_dict)
        resp_json = resp.json()
        _id = resp_json["data"]
        payload_dict["id"] = _id
        assert resp.status_code == 201
        if not is_opened:
            end_time = start_time + timedelta(hours=1)
            await async_client.patch(
                f"{SESSIONS_ENDPOINT}/{_id}",
                json={
                    "end_time": end_time.isoformat(),
                    "is_opened": is_opened,
                },
            )
            assert resp.status_code == 201
            payload_dict["end_time"] = end_time
            payload_dict["is_opened"] = False
        payload_dict["start_time"] = start_time
        sessions.append(SessionsRead(**payload_dict))
    return sessions
