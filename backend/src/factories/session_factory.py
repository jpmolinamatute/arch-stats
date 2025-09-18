"""
Factory for creating Session records in the database for testing or seeding.

Matches migration V004 (sessions table): requires owner_archer_id, session_location,
is_indoor, is_opened. created_at defaults; closed_at optional.
Ensures at most one open session per archer to satisfy the partial unique index.
"""

import random
from collections.abc import Sequence
from uuid import UUID

from asyncpg import Pool
from faker import Faker


fake = Faker()


async def create_sessions(pool: Pool, qty: int) -> Sequence[UUID]:
    """
    Create `qty` Session records.

    Strategy:
    - Fetch all archer IDs.
    - For the first N=min(qty, num_archers), create OPEN sessions (is_opened=True),
      one per distinct archer, to satisfy the unique-open-session-per-archer index.
    - For any remaining sessions (if qty > num_archers), create CLOSED sessions
      (is_opened=False), which can reuse archers without violating the index.

    Args:
        pool: DB pool.
        qty: number of sessions to create.

    Returns:
        Sequence of created session UUIDs.

    Raises:
        ValueError: if there are no archers available to own sessions.
    """
    session_ids: list[UUID] = []
    async with pool.acquire() as conn:
        rows = await conn.fetch("SELECT archer_id FROM archers;")
        archer_ids: list[UUID] = [row["archer_id"] for row in rows]
        if not archer_ids:
            raise ValueError("No archers found. Create archers before creating sessions.")

        # Determine how many open sessions we can safely create (one per archer).
        num_open = min(qty, len(archer_ids))

        # Create the open sessions first, one per distinct archer.
        for i in range(num_open):
            owner = archer_ids[i]
            session_location = fake.city()
            is_indoor = random.choice([True, False])
            result = await conn.fetchrow(
                """
                INSERT INTO session (owner_archer_id, session_location, is_indoor, is_opened)
                VALUES ($1, $2, $3, TRUE)
                RETURNING session_id
                """,
                owner,
                session_location,
                is_indoor,
            )
            if result is not None and "session_id" in result:
                session_ids.append(result["session_id"])  # UUID

        # Create any remaining sessions as closed, reusing archers round-robin.
        for j in range(num_open, qty):
            owner = archer_ids[j % len(archer_ids)]
            session_location = fake.city()
            is_indoor = random.choice([True, False])
            result = await conn.fetchrow(
                """
                INSERT INTO session (owner_archer_id, session_location, is_indoor, is_opened)
                VALUES ($1, $2, $3, FALSE)
                RETURNING session_id
                """,
                owner,
                session_location,
                is_indoor,
            )
            if result is not None and "session_id" in result:
                session_ids.append(result["session_id"])  # UUID

    return session_ids
