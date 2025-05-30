from datetime import datetime, timezone, timedelta
from typing import Any

from asyncpg import Pool

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


async def create_many_sessions(db_pool: Pool, count: int = 5) -> list[SessionsRead]:
    sessions = []
    for i in range(count):
        is_opened = bool(i % 2)
        start_time = datetime.now(timezone.utc) + timedelta(days=i)
        session = create_fake_sessions(
            location=f"Range_{i%2}",
            start_time=start_time,
        )
        payload_dict = session.model_dump(exclude_none=True, by_alias=True)
        payload_dict["end_time"] = None
        if not is_opened:
            payload_dict["end_time"] = start_time + timedelta(hours=1)
            payload_dict["is_opened"] = False

        insert_sql = """
            INSERT INTO sessions (
                is_opened, start_time, location, end_time
            ) VALUES (
                $1, $2, $3, $4
            ) RETURNING id
        """
        async with db_pool.acquire() as conn:
            row = await conn.fetchrow(
                insert_sql,
                payload_dict["is_opened"],
                payload_dict["start_time"],
                payload_dict["location"],
                payload_dict["end_time"],
            )
            assert row is not None
            payload_dict["id"] = row["id"]
        sessions.append(SessionsRead(**payload_dict))
    return sessions
