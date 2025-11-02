"""
Factory for creating Session records in the database for testing or seeding.

Matches migration V004 (sessions table): requires owner_archer_id, session_location,
is_indoor, is_opened. created_at defaults; closed_at optional.
Ensures at most one open session per archer to satisfy the partial unique index.
"""

import random
from collections.abc import Sequence
from uuid import UUID

from asyncpg import Connection, Pool
from asyncpg.pool import PoolConnectionProxy
from faker import Faker


fake = Faker()


async def _insert_sessions(
    conn: Connection | PoolConnectionProxy,
    archer_ids: list[UUID],
    total_number_session: int,
    open_session: int,
) -> list[UUID]:
    """
    Helper to insert sessions into the database.

    Args:
        conn: asyncpg connection.
        archer_ids: list of archer UUIDs.
        total_number_session: total number of sessions to create.
        open_session: number of open sessions to create (rest will be closed).

    Returns:
        List of created session UUIDs.
    """
    session_ids: list[UUID] = []
    for i in range(total_number_session):
        owner = archer_ids[i] if i < open_session else archer_ids[i % len(archer_ids)]
        session_location = fake.city()
        is_indoor = random.choice([True, False])
        shot_per_round = random.choice([3, 4, 5, 6])
        is_opened = i < open_session
        result = await conn.fetchrow(
            """
            INSERT INTO session (
                owner_archer_id,
                session_location,
                is_indoor,
                shot_per_round,
                is_opened
            )
            VALUES ($1, $2, $3, $4, $5)
            RETURNING session_id
            """,
            owner,
            session_location,
            is_indoor,
            shot_per_round,
            is_opened,
        )
        if result is not None and "session_id" in result:
            session_ids.append(result["session_id"])
    return session_ids


async def create_sessions(pool: Pool, qty: int) -> Sequence[UUID]:
    """
    Create `qty` Session records.

    Strategy:
    - Fetch all archer IDs.
    - For the first N=min(qty, num_archers), create OPEN sessions (is_opened=True),
      one per distinct archer, to satisfy the unique-open-session-per-archer index.
    - For any remaining sessions (if qty > num_archers), create CLOSED sessions
      (is_opened=False), which can reuse archer without violating the index.

    Args:
        pool: DB pool.
        qty: number of sessions to create.

    Returns:
        Sequence of created session UUIDs.

    Raises:
        ValueError: if there are no archer available to own sessions.
    """
    async with pool.acquire() as conn:
        rows = await conn.fetch("SELECT archer_id FROM archer;")
        archer_ids: list[UUID] = [row["archer_id"] for row in rows]
        if not archer_ids:
            raise ValueError("No archer found. Create archer before creating sessions.")

        num_open = min(qty, len(archer_ids))
        session_ids = await _insert_sessions(conn, archer_ids, qty, num_open)
    return session_ids
