import random
from typing import Any
from uuid import UUID

from asyncpg import Pool

from shared.schema import TargetFace, TargetsCreate, TargetsRead


def create_fake_radii_list(no_items: int) -> list[float]:
    return [4.0 * (i + 1) for i in range(no_items)]


def create_fake_points_list(no_items: int) -> list[int]:
    return [10 - i for i in range(no_items)]


def create_fake_axis(biggest_radii: float, max_allowed: float) -> float:
    axis = 0.0
    found = False
    while not found:
        axis = random.random() * max_allowed
        if axis + biggest_radii <= max_allowed and axis - biggest_radii >= 0:
            found = True
    return axis


def create_no_faces(face_count: int) -> int:
    number: int
    if face_count == 0:
        number = random.randint(0, 4)
    elif 0 < face_count <= 4:
        number = face_count
    else:
        raise ValueError("face_count must be >= 0 and <= 4")
    return number


def create_fake_faces(face_count: int, max_x: float, max_y: float) -> list[TargetFace]:
    """
    Generate a list of TargetFace objects following constraints:
    - 0 through 4 faces per target. Unless explicitly specified `face_count`.
    - face[].human_identifier unique per target.
    - radii list has 3..10 items, ascending.
    - points length matches radii.
    - points values start at 10 and descend.
    - The minimum value a radii is 4.
    - Each item in the radii list increments by 4 (4, 8, 12, ...).
    - x + max(radii) <= max_x and y + max(radii) <= max_y.
    - x - max(radii) >= 0 and y - max(radii) >= 0.
    """

    # Decide face count (respect overall 0..4 rule)
    no_faces = create_no_faces(face_count)

    faces: list[TargetFace] = []

    for i in range(no_faces):
        hid = f"A{i+1}"
        radii = create_fake_radii_list(random.randint(3, 10))
        faces.append(
            TargetFace(
                x=create_fake_axis(biggest_radii=max(radii), max_allowed=max_x),
                y=create_fake_axis(biggest_radii=max(radii), max_allowed=max_y),
                radii=radii,
                points=create_fake_points_list(len(radii)),
                human_identifier=hid,
            )
        )

    return faces


def create_fake_target(session_id: UUID, face_count: int = 0, **overrides: Any) -> TargetsCreate:
    """
    Generate a realistic TargetsCreate payload with constraints:

    """

    # Overall target dimensions (choose reasonable, non-tiny bounds)
    max_x = random.uniform(120.0, 240.0)
    max_y = random.uniform(120.0, 240.0)

    faces: list[TargetFace] = create_fake_faces(face_count=face_count, max_x=max_x, max_y=max_y)

    data = TargetsCreate(
        max_x=max_x,
        max_y=max_y,
        session_id=session_id,
        faces=faces,
    )

    return data.model_copy(update=overrides)


async def insert_targets_db(db_pool: Pool, target_rows: list[TargetsCreate]) -> list[TargetsRead]:
    """Insert targets and return their TargetsRead rows."""
    results: list[TargetsRead] = []
    if not target_rows:
        return results

    insert_sql = """
        INSERT INTO targets (max_x, max_y, session_id, faces)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    """
    async with db_pool.acquire() as conn:
        async with conn.transaction():
            stmt = await conn.prepare(insert_sql)
            for target in target_rows:
                row = await stmt.fetchrow(
                    target.max_x,
                    target.max_y,
                    target.session_id,
                    target.faces_as_json(),
                )
                assert row is not None
                payload = target.model_dump(mode="json", by_alias=True)
                payload["id"] = row["id"]
                results.append(TargetsRead(**payload))
    return results


async def create_many_targets(
    db_pool: Pool,
    session_id: UUID,
    targets_count: int = 5,
) -> list[TargetsRead]:
    targets_to_insert: list[TargetsCreate] = []
    for _ in range(targets_count):
        targets_to_insert.append(
            create_fake_target(
                session_id=session_id,
                face_count=1,  # ensure at least one face for predictable tests
            )
        )
    results = await insert_targets_db(db_pool, targets_to_insert)
    return results
