"""
Factory for creating Arrow records for testing or seeding.

Aligns with migration V005 (arrow table):
- arrow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
- archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE CASCADE
- spine FLOAT
- length FLOAT
- weight FLOAT
- arrow_number INTEGER NOT NULL
- archer_id UUID REFERENCES archer(archer_id)
- created_at TIMESTAMPTZ NOT NULL DEFAULT now()

Notes:
- This factory generates plausible arrows with typical dimensions and spine.
- `arrow_number` will be 1..qty for each archer.
"""

import random
from typing import Final
from uuid import UUID

from asyncpg import Pool

from factories.archer_factory import create_archers

# Typical ranges
_LENGTH_RANGE_IN: Final[tuple[float, float]] = (26.0, 32.0)  # inches
_WEIGHT_RANGE_GRAINS: Final[tuple[float, float]] = (250.0, 700.0)  # grains
_SPINE_VALUES: Final[list[float]] = [
    250.0,
    300.0,
    340.0,
    400.0,
    500.0,
    600.0,
    700.0,
    800.0,
    900.0,
]


def _rand_length() -> float:
    return round(random.uniform(*_LENGTH_RANGE_IN), 1)


def _rand_weight() -> float:
    return round(random.uniform(*_WEIGHT_RANGE_GRAINS), 1)


def _rand_spine() -> float:
    return random.choice(_SPINE_VALUES)


async def create_arrows(
    pool: Pool,
    qty: int,
    archer_ids: list[UUID] | None = None,
) -> list[UUID]:
    """Create `qty` Arrow records per archer.

    Args:
        pool: Asyncpg pool for DB access.
        qty: Number of arrows to create for each archer.
        archer_ids: Optional list of archers. If None/empty, a single archer
            will be created and used.

    Returns:
        List of created arrow UUIDs.

    Notes:
        - Total arrows created equals `qty * number_of_archers`.
    """
    arrow_ids: list[UUID] = []
    if qty <= 0:
        return arrow_ids

    # Ensure we have at least one archer
    archer_ids = list(archer_ids or [])
    if not archer_ids:
        created = await create_archers(pool, 1)
        archer_ids = list(created)

    async with pool.acquire() as conn:
        for archer_id in archer_ids:
            for i in range(qty):
                length_in = _rand_length()
                weight_gr = _rand_weight()
                spine = _rand_spine()
                # Number arrows 1..qty for each archer
                arrow_number = i + 1

                rec = await conn.fetchrow(
                    """
                    INSERT INTO arrow (archer_id, spine, length, weight, arrow_number)
                    VALUES ($1, $2, $3, $4, $5)
                    RETURNING arrow_id
                    """,
                    archer_id,
                    spine,
                    length_in,
                    weight_gr,
                    arrow_number,
                )
                if rec is None or "arrow_id" not in rec:
                    continue
                arrow_ids.append(rec["arrow_id"])  # UUID

    return arrow_ids
