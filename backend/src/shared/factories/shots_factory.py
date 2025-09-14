import random
from datetime import datetime, timedelta, timezone
from typing import Any
from uuid import UUID

from asyncpg import Pool

from shared.schema import ShotsCreate, ShotsRead


async def insert_shots_db(db_pool: Pool, shot_rows: list[ShotsCreate]) -> list[ShotsRead]:
    """Insert multiple shots and return their ShotsRead rows."""
    if not shot_rows:
        return []
    insert_sql = """
        INSERT INTO shots (
            arrow_id, session_id, arrow_engage_time, arrow_disengage_time, 
            arrow_landing_time, x, y
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
        """
    results: list[ShotsRead] = []
    async with db_pool.acquire() as conn:
        async with conn.transaction():
            stmt = await conn.prepare(insert_sql)
            for shot in shot_rows:
                row = await stmt.fetchrow(
                    shot.arrow_id,
                    shot.session_id,
                    shot.arrow_engage_time,
                    shot.arrow_disengage_time,
                    shot.arrow_landing_time,
                    shot.x,
                    shot.y,
                )
                assert row is not None
                payload = shot.model_dump(exclude_none=True, by_alias=True)
                payload["id"] = row["id"]
                results.append(ShotsRead(**payload))
    return results


def create_fake_shot(arrow_id: UUID, session_id: UUID, **overrides: Any) -> ShotsCreate:
    now = datetime.now(timezone.utc)
    data = ShotsCreate(
        arrow_id=arrow_id,
        session_id=session_id,
        arrow_engage_time=now,
        arrow_disengage_time=now + timedelta(seconds=2),
        arrow_landing_time=now + timedelta(seconds=4),
        x=random.uniform(0.0, 100.0),
        y=random.uniform(0.0, 100.0),
    )

    return data.model_copy(update=overrides)


async def create_many_shots(
    db_pool: Pool, arrows_id: list[UUID], session_id: UUID, shots_count: int = 5
) -> list[ShotsRead]:
    shot_rows: list[ShotsCreate] = []
    if not arrows_id:
        return []
    for i in range(shots_count):
        # cycle through provided arrows if fewer than shots_count
        shot_rows.append(create_fake_shot(arrows_id[i % len(arrows_id)], session_id))
    return await insert_shots_db(db_pool, shot_rows)
