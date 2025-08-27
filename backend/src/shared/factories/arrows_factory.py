import random
from datetime import datetime, timezone
from typing import Any
from uuid import uuid4

from asyncpg import Pool

from server.schema import ArrowsCreate


ARROWS_ENDPOINT = "/api/v0/arrow"


def create_fake_arrow(**overrides: Any) -> ArrowsCreate:
    ran_int = random.randint(1, 100)
    data = ArrowsCreate(
        id=uuid4(),
        length=29.0,
        human_identifier=f"arrow-{ran_int}",
        registration_date=datetime.now(timezone.utc),
        is_programmed=False,
        is_active=True,
        voided_date=None,
        label_position=1.0,
        weight=350.0,
        diameter=5.7,
        spine=400,
    )
    return data.model_copy(update=overrides)


async def create_many_arrows(db_pool: Pool, count: int = 5) -> list[ArrowsCreate]:
    arrows = []
    for i in range(count):
        arrow_data = create_fake_arrow(
            is_programmed=bool(i % 2),
            spine=400 + 10 * i,
            weight=350.0 + 5 * i,
            human_identifier=f"arrow_{i}",
            length=29.0 + i,
            label_position=1.0 + 0.5 * i,
            diameter=5.7 + 0.1 * i,
        )
        # Prepare SQL for insert (with all columns)
        insert_sql = """
            INSERT INTO arrows (
                id, length, human_identifier, is_programmed, label_position, weight, diameter, spine
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING id
        """
        async with db_pool.acquire() as conn:
            row = await conn.fetchrow(
                insert_sql,
                arrow_data.arrow_id,
                arrow_data.length,
                arrow_data.human_identifier,
                arrow_data.is_programmed,
                arrow_data.label_position,
                arrow_data.weight,
                arrow_data.diameter,
                arrow_data.spine,
            )
            assert row is not None
            arrows.append(arrow_data)
    return arrows
