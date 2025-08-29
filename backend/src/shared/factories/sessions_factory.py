from datetime import datetime, timedelta, timezone
from typing import Any

from asyncpg import Pool

from shared.schema import SessionsCreate, SessionsRead


SESSIONS_ENDPOINT = "/api/v0/session"


def create_fake_session(**overrides: Any) -> SessionsCreate:
    """
    Default payload for session creation, with overrides for specific test cases.
    """
    data = SessionsCreate(
        is_opened=True,
        start_time=datetime.now(timezone.utc),
        location="Test Range",
        is_indoor=False,
        distance=18,
    )
    return data.model_copy(update=overrides)


async def insert_sessions_db(
    db_pool: Pool, session_rows: list[SessionsCreate]
) -> list[SessionsRead]:
    """Insert multiple sessions and return their SessionsRead rows."""
    if not session_rows:
        return []
    insert_sql = """
        INSERT INTO sessions (
            is_opened, start_time, location, is_indoor, distance, end_time
        ) VALUES (
            $1, $2, $3, $4, $5, $6
        ) RETURNING id
    """
    results: list[SessionsRead] = []
    async with db_pool.acquire() as conn:
        async with conn.transaction():
            stmt = await conn.prepare(insert_sql)
            for s in session_rows:
                # End time inferred: if is_opened False, set end_time to start_time + 1h
                end_time = None
                if not s.is_opened:
                    end_time = s.start_time + timedelta(hours=1)
                row = await stmt.fetchrow(
                    s.is_opened,
                    s.start_time,
                    s.location,
                    s.is_indoor,
                    s.distance,
                    end_time,
                )
                assert row is not None
                payload = s.model_dump(exclude_none=False, by_alias=True)
                payload["end_time"] = end_time
                payload["id"] = row["id"]
                results.append(SessionsRead(**payload))
    return results


async def create_many_sessions(db_pool: Pool, sessions_count: int = 5) -> list[SessionsRead]:
    sessions_to_insert: list[SessionsCreate] = []
    for i in range(sessions_count):
        start_time = datetime.now(timezone.utc) + timedelta(days=i)
        is_opened = bool(i % 2)
        sessions_to_insert.append(
            create_fake_session(
                location=f"Range_{i%2}",
                start_time=start_time,
                is_opened=is_opened,
            )
        )
    return await insert_sessions_db(db_pool, sessions_to_insert)
