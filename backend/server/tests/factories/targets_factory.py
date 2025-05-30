from typing import Any
from uuid import UUID

from asyncpg import Pool


from server.schema import TargetsCreate, TargetsRead

TARGETS_ENDPOINT = "/api/v0/target"


def create_fake_target(session_id: UUID, **overrides: Any) -> TargetsCreate:
    """
    Generate a valid payload for TargetsCreate.
    Adjust the fields to match your schema exactly!
    """

    data = TargetsCreate(
        max_x_coordinate=122.0,
        max_y_coordinate=122.0,
        radius=[3.0, 6.0, 9.0, 12.0, 15.0],
        points=[10, 8, 6, 4, 2],
        height=140.0,
        session_id=session_id,
        human_identifier="T1",
    )
    return data.model_copy(update=overrides)


async def create_many_targets(
    db_pool: Pool,
    session_id: UUID,
    count: int = 5,
) -> list[TargetsRead]:
    targets = []
    for i in range(count):
        target = create_fake_target(
            session_id=session_id,
            max_x_coordinate=120.0 + i,
            max_y_coordinate=121.0 + i,
            height=140.0 + i,
            human_identifier=f"target_{i}",
        )
        insert_sql = """
            INSERT INTO targets (
                max_x_coordinate, max_y_coordinate, radius, points, height, human_identifier, session_id
            ) VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING id
        """
        async with db_pool.acquire() as conn:
            row = await conn.fetchrow(
                insert_sql,
                target.max_x_coordinate,
                target.max_y_coordinate,
                target.radius,
                target.points,
                target.height,
                target.human_identifier,
                target.session_id,
            )
            assert row is not None
        payload_dict = target.model_dump(mode="json", by_alias=True)
        payload_dict["id"] = row["id"]
        targets.append(TargetsRead(**payload_dict))
    return targets
