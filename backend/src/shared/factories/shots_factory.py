import random
from datetime import datetime, timedelta, timezone
from typing import Any
from uuid import UUID

from asyncpg import Pool

from shared.schema import ShotsCreate, ShotsRead


async def insert_shot_db(db_pool: Pool, shot_row: ShotsCreate) -> UUID:
    async with db_pool.acquire() as conn:
        row = await conn.fetchrow(
            """
            INSERT INTO shots (
                arrow_id, session_id, arrow_engage_time, arrow_disengage_time, 
                arrow_landing_time, x, y
            ) VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING id
            """,
            shot_row.arrow_id,
            shot_row.session_id,
            shot_row.arrow_engage_time,
            shot_row.arrow_disengage_time,
            shot_row.arrow_landing_time,
            shot_row.x,
            shot_row.y,
        )
        if row is None:
            raise RuntimeError("Insert failed; no id returned")
        _id: UUID = row["id"]
        return _id


def create_fake_shot(arrow_id: UUID, session_id: UUID, **overrides: Any) -> ShotsCreate:
    now = datetime.now(timezone.utc)
    data = ShotsCreate(
        arrow_id=arrow_id,
        session_id=session_id,
        arrow_engage_time=now,
        arrow_disengage_time=now + timedelta(seconds=2),
        arrow_landing_time=now + timedelta(seconds=4),
        x=random.uniform(0, 100),
        y=random.uniform(0, 100),
    )

    return data.model_copy(update=overrides)


async def create_many_shots(
    db_pool: Pool, arrows_id: list[UUID], session_id: UUID, count: int = 5
) -> list[ShotsRead]:
    shots = []
    for i in range(count):
        payload = create_fake_shot(arrows_id[i], session_id)
        shot_id = await insert_shot_db(db_pool, payload)
        payload_dict = payload.model_dump(exclude_none=True, by_alias=True)
        payload_dict["id"] = shot_id
        shots.append(ShotsRead(**payload_dict))
    return shots
