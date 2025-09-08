import random
from typing import Any
from uuid import UUID

from asyncpg import Pool

from shared.factories.faces_factory import create_many_faces
from shared.schema import TargetsCreate, TargetsRead


def create_fake_target(session_id: UUID, **overrides: Any) -> TargetsCreate:
    """
    Generate a realistic TargetsCreate payload with constraints:

    """

    # Overall target dimensions (choose reasonable, non-tiny bounds)
    max_x = random.uniform(120.0, 240.0)
    max_y = random.uniform(120.0, 240.0)

    data = TargetsCreate(
        max_x=max_x,
        max_y=max_y,
        session_id=session_id,
        distance=random.randint(10, 70),
    )

    return data.model_copy(update=overrides)


async def insert_targets_db(db_pool: Pool, target_rows: list[TargetsCreate]) -> list[TargetsRead]:
    """Insert targets and return their TargetsRead rows."""
    results: list[TargetsRead] = []
    if not target_rows:
        return results

    insert_sql = """
        INSERT INTO targets (max_x, max_y, session_id, distance)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    """
    async with db_pool.acquire() as conn:
        async with conn.transaction():
            insert_target_stmt = await conn.prepare(insert_sql)
            for target in target_rows:
                row = await insert_target_stmt.fetchrow(
                    target.max_x,
                    target.max_y,
                    target.session_id,
                    target.distance,
                )
                assert row is not None
                target_id = row["id"]

                payload = target.model_dump(mode="json", by_alias=True)
                payload["id"] = target_id
                results.append(TargetsRead(**payload))
    return results


async def create_many_targets(
    db_pool: Pool,
    session_id: UUID,
    targets_count: int = 5,
) -> list[TargetsRead]:
    targets_to_insert: list[TargetsCreate] = []
    for _ in range(targets_count):
        targets_to_insert.append(create_fake_target(session_id=session_id))
    results = await insert_targets_db(db_pool, targets_to_insert)
    for target in results:
        await create_many_faces(db_pool, target.get_id(), target.max_x, target.max_y)
    return results
