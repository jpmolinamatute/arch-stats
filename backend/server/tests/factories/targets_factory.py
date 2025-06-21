from typing import Any
from uuid import UUID

from asyncpg import Pool

from server.schema import TargetFace, TargetsCreate, TargetsRead


TARGETS_ENDPOINT = "/api/v0/target"


def create_fake_target(session_id: UUID, **overrides: Any) -> TargetsCreate:
    """
    Generate a valid payload for TargetsCreate.
    Adjust the fields to match your schema exactly!
    """

    data = TargetsCreate(
        max_x_coordinate=122.0,
        max_y_coordinate=122.0,
        session_id=session_id,
        faces=[
            TargetFace(
                center_x=123.0,
                center_y=123.0,
                radius=[50.0, 60.0, 80.0, 90.0, 100.0],
                points=[10, 9, 8, 7, 5],
                human_identifier="a1",
            ),
            TargetFace(
                center_x=124.0,
                center_y=124.0,
                radius=[55.0, 65.0, 85.0, 95.0, 105.0],
                points=[11, 10, 9, 8, 6],
                human_identifier="a2",
            ),
            TargetFace(
                center_x=125.0,
                center_y=125.0,
                radius=[60.0, 70.0, 90.0, 100.0, 110.0],
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
                max_x_coordinate, max_y_coordinate, session_id, faces
            ) VALUES ($1, $2, $3, $4)
            RETURNING id
        """
        async with db_pool.acquire() as conn:
            row = await conn.fetchrow(
                insert_sql,
                target.max_x_coordinate,
                target.max_y_coordinate,
                target.session_id,
                target.faces_as_json(),
            )
            assert row is not None
        payload_dict = target.model_dump(mode="json", by_alias=True)
        payload_dict["id"] = row["id"]
        targets.append(TargetsRead(**payload_dict))
    return targets
