"""
Factory for creating Target records in the database for testing or seeding.

This factory aligns with the `targets` table defined in migrations:
    target_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions (session_id) ON DELETE CASCADE,
    distance INTEGER NOT NULL,
    lane INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
and uses the helper function `get_next_lane(session_id)` to allocate lanes
sequentially per session.
"""

import random
from collections.abc import Sequence
from uuid import UUID

from asyncpg import Pool


async def create_targets(pool: Pool, qty: int, session_id: UUID) -> Sequence[UUID]:
    """
    Create `qty` Target records in the database for a given session.

    The session must exist in the `sessions` table. The `distance` for each
    created target is assigned independently (valid range 1..100), and
    the `lane` is assigned using `COALESCE(get_next_lane(session_id), 1)` to
    satisfy the unique (session_id, lane) constraint.

    Args:
        pool: The asyncpg Pool to use for DB access.
        qty: Number of targets to create.
        session_id: The session UUID to which these targets belong.

    Returns:
        A list of inserted target UUIDs.

    Raises:
        ValueError: If the provided session_id does not exist.
    """
    target_ids: list[UUID] = []
    async with pool.acquire() as conn:
        # Verify the session exists.
        row = await conn.fetchrow("SELECT 1 FROM session WHERE session_id = $1", session_id)
        if row is None:
            raise ValueError(f"Session {session_id} does not exist")

        # Use a transaction so successive calls to get_next_lane() observe our own inserts.
        async with conn.transaction():
            for _ in range(qty):
                # Choose a valid distance (schema allows 1..100). Use typical WA distances.
                distance = random.choice([18, 30, 50, 70])
                result = await conn.fetchrow(
                    """
                    INSERT INTO target (session_id, distance, lane)
                    VALUES ($1, $2, COALESCE(get_next_lane($1), 1))
                    RETURNING target_id
                    """,
                    session_id,
                    distance,
                )
                if result is not None and "target_id" in result:
                    target_ids.append(result["target_id"])  # UUID
    return target_ids
