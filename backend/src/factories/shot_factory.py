"""
Factory for creating Shot records in the database for testing or seeding.

Aligns with migration V006 (shot table):
- shot_id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
- slot_id UUID NOT NULL REFERENCES slot (slot_id) ON DELETE CASCADE
- x DOUBLE PRECISION, y DOUBLE PRECISION, score INTEGER (0..10)
- is_x BOOLEAN NOT NULL DEFAULT FALSE
- arrow_id UUID REFERENCES arrow (arrow_id) ON DELETE SET NULL
- created_at TIMESTAMPTZ NOT NULL DEFAULT now()

Constraints to respect:
- CHECK (score BETWEEN 0 AND 10)
- Either all of (x, y, score) are NULL or all are NOT NULL
 - When score = 10, some shots may be marked as inner-10 via is_x = TRUE (data-level convention)

Notes:
- This factory generates fully recorded shots (x, y, score all present).
- If a non-empty list of arrow_ids is provided, shots randomly reference an arrow; otherwise
  arrow_id is NULL.
"""

import random
from collections.abc import Sequence
from typing import Final
from uuid import UUID

from asyncpg import Pool

# Basic scoring distribution: bias towards mid/high scores, allow some lows.
# Indices 0..10 correspond to the actual score value.
_SCORE_WEIGHTS: Final[list[int]] = [
    # 0  1  2  3  4  5
    1,
    1,
    2,
    3,
    5,
    8,
    # 6   7   8   9    10
    13,
    21,
    34,
    34,
    21,
]

PERFECT_SCORE: Final[int] = 10
X_RING_PROBABILITY: Final[float] = 0.35


def _random_score() -> int:
    """Return a plausible score in [0, 10] using a simple weighted distribution."""
    values = list(range(11))
    return random.choices(values, weights=_SCORE_WEIGHTS, k=1)[0]


def _random_xy() -> tuple[float, float]:
    """Generate shot coordinates using a normal distribution around (0, 0).

    The schema imposes no bounds on x/y; these are illustrative values only.
    """
    # Standard deviation in arbitrary units; tweak as needed for tests.
    return (random.gauss(0.0, 15.0), random.gauss(0.0, 15.0))


async def create_shots(
    pool: Pool,
    qty: int,
    *,
    slot_ids: Sequence[UUID],
    arrow_ids: Sequence[UUID] | None = None,
) -> Sequence[UUID]:
    """Create `qty` Shot records.

    Args:
        pool: Asyncpg pool used for DB access.
        qty: Number of shots to create.
        slot_ids: Sequence of slot UUIDs; shots will be attached to these slots
            (chosen uniformly at random).
        arrow_ids: Optional sequence of arrow UUIDs. If provided and non-empty,
            each shot references a random arrow; otherwise arrow_id is NULL.

    Returns:
        Sequence of created shot UUIDs.
    """
    shot_ids: list[UUID] = []
    if qty <= 0 or not slot_ids:
        return shot_ids

    # Validate that provided slots exist (defensive; avoids FK violations).
    valid_slot_ids: set[UUID]
    async with pool.acquire() as conn:
        rows = await conn.fetch(
            "SELECT slot_id FROM slot WHERE slot_id = ANY($1::uuid[])",
            list(slot_ids),
        )
        valid_slot_ids = {row["slot_id"] for row in rows}
        if not valid_slot_ids:
            return shot_ids

        # Insert shots one by one with RETURNING shot_id
        for _ in range(qty):
            slot_id = random.choice(tuple(valid_slot_ids))
            x, y = _random_xy()
            score = _random_score()
            # Mark some 10s as inner-10 (X) for realism. Keep others FALSE.
            is_x = bool(score == PERFECT_SCORE and random.random() < X_RING_PROBABILITY)
            arrow_id: UUID | None = None
            if arrow_ids:
                arrow_id = random.choice(list(arrow_ids))

            rec = await conn.fetchrow(
                """
                INSERT INTO shot (slot_id, x, y, score, is_x, arrow_id)
                VALUES ($1, $2, $3, $4, $5, $6)
                RETURNING shot_id
                """,
                slot_id,
                x,
                y,
                score,
                is_x,
                arrow_id,
            )
            if rec is None or "shot_id" not in rec:
                continue
            shot_ids.append(rec["shot_id"])  # UUID

    return shot_ids
