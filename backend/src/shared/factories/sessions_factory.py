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
            is_opened, start_time, location, is_indoor, end_time
        ) VALUES (
            $1, $2, $3, $4, $5
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
                    end_time,
                )
                assert row is not None
                payload = s.model_dump(exclude_none=False, by_alias=True)
                payload["end_time"] = end_time
                payload["id"] = row["id"]
                results.append(SessionsRead(**payload))
    return results


async def create_many_sessions(db_pool: Pool, sessions_count: int = 5) -> list[SessionsRead]:
    """Create multiple sessions respecting single-open-session rule.

    Generates (sessions_count - 1) closed sessions (each with synthetic end_time)
    plus one open session (the most recent chronologically). If sessions_count is
    0, returns empty list; if 1, returns just one open session.
    """
    if sessions_count <= 0:
        return []
    sessions_to_insert: list[SessionsCreate] = []
    base_time = datetime.now(timezone.utc)
    # Closed sessions first
    for i in range(max(0, sessions_count - 1)):
        start_time = base_time - timedelta(days=sessions_count - 1 - i)
        sessions_to_insert.append(
            create_fake_session(
                location=f"Range_closed_{i}",
                start_time=start_time,
                is_opened=False,
                end_time=start_time + timedelta(hours=3),
            )
        )
    # Single open session (latest start_time)
    sessions_to_insert.append(
        create_fake_session(location="Range_open", start_time=base_time, is_opened=True)
    )
    return await insert_sessions_db(db_pool, sessions_to_insert)
