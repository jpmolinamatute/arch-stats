import random
from typing import Any
from uuid import UUID

from asyncpg import Pool

from shared.schema import TargetFace, TargetsCreate, TargetsRead


TARGETS_ENDPOINT = "/api/v0/target"


def create_fake_target(session_id: UUID, **overrides: Any) -> TargetsCreate:
    """
    Generate a valid payload for TargetsCreate.
    Adjust the fields to match your schema exactly!
    """
    max_x = random.uniform(0, 100)
    max_y = random.uniform(0, 100)
    data = TargetsCreate(
        max_x=max_x,
        max_y=max_y,
        session_id=session_id,
        faces=[
            TargetFace(
                x=random.uniform(0, max_x),
                y=random.uniform(0, max_y),
                radii=[50.0, 60.0, 80.0, 90.0, 100.0],
                points=[10, 9, 8, 7, 5],
                human_identifier="a1",
            ),
            TargetFace(
                x=random.uniform(0, max_x),
                y=random.uniform(0, max_y),
                radii=[55.0, 65.0, 85.0, 95.0, 105.0],
                points=[11, 10, 9, 8, 6],
                human_identifier="a2",
            ),
            TargetFace(
                x=random.uniform(0, max_x),
                y=random.uniform(0, max_y),
                radii=[60.0, 70.0, 90.0, 100.0, 110.0],
                points=[12, 11, 10, 9, 7],
                human_identifier="a3",
            ),
        ],
    )
    return data.model_copy(update=overrides)


async def create_many_targets(
    db_pool: Pool,
    session_id: UUID,
    count: int = 5,
) -> list[TargetsRead]:
    targets = []
    for _ in range(count):
        target = create_fake_target(
            session_id=session_id,
        )
        insert_sql = """
            INSERT INTO targets (
                max_x, max_y, session_id, faces
            ) VALUES ($1, $2, $3, $4)
            RETURNING id
        """
        async with db_pool.acquire() as conn:
            row = await conn.fetchrow(
                insert_sql,
                target.max_x,
                target.max_y,
                target.session_id,
                target.faces_as_json(),
            )
            assert row is not None
        payload_dict = target.model_dump(mode="json", by_alias=True)
        payload_dict["id"] = row["id"]
        targets.append(TargetsRead(**payload_dict))
    return targets
