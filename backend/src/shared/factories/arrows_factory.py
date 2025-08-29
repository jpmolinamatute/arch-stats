import random
from datetime import datetime, timezone
from typing import Any
from uuid import uuid4

from asyncpg import Pool

from shared.schema import ArrowsCreate
from shared.schema.arrows_schema import ArrowsRead


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


async def insert_arrows_db(db_pool: Pool, arrow_rows: list[ArrowsCreate]) -> list[ArrowsRead]:
    """Insert multiple arrows and return their ArrowsRead rows."""
    if not arrow_rows:
        return []
    insert_sql = """
        INSERT INTO arrows (
            id, length, human_identifier, is_programmed, label_position, weight, diameter, spine
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    """
    results: list[ArrowsRead] = []
    async with db_pool.acquire() as conn:
        async with conn.transaction():
            stmt = await conn.prepare(insert_sql)
            for arrow in arrow_rows:
                row = await stmt.fetchrow(
                    arrow.arrow_id,
                    arrow.length,
                    arrow.human_identifier,
                    arrow.is_programmed,
                    arrow.label_position,
                    arrow.weight,
                    arrow.diameter,
                    arrow.spine,
                )
                assert row is not None
                payload = arrow.model_dump(exclude_none=True, by_alias=True)
                payload["id"] = row["id"]
                results.append(ArrowsRead(**payload))
    return results


async def create_many_arrows(db_pool: Pool, arrows_count: int = 5) -> list[ArrowsRead]:
    arrows_to_insert: list[ArrowsCreate] = []
    for i in range(arrows_count):
        arrows_to_insert.append(
            create_fake_arrow(
                is_programmed=bool(i % 2),
                spine=400 + 10 * i,
                weight=350.0 + 5 * i,
                human_identifier=f"arrow_{i}",
                length=29.0 + i,
                label_position=1.0 + 0.5 * i,
                diameter=5.7 + 0.1 * i,
            )
        )
    return await insert_arrows_db(db_pool, arrows_to_insert)
