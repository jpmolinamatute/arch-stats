import random
from uuid import UUID, uuid4

from asyncpg import Pool

from shared.schema.faces_schema import FacesCreate, FacesRead


def create_fake_radii_list(no_items: int) -> list[float]:
    return [4.0 * (i + 1) for i in range(no_items)]


def create_fake_points_list(no_items: int) -> list[int]:
    return [10 - i for i in range(no_items)]


def create_fake_axis(biggest_radii: float, max_allowed: float) -> float:
    axis = 0.0
    found = False
    while not found:
        axis = random.random() * max_allowed
        if axis + biggest_radii < max_allowed and axis - biggest_radii > 0:
            found = True
    return axis


def valid_no_faces(face_count: int) -> int:
    number: int
    if face_count == 0:
        number = random.randint(0, 4)
    elif 0 < face_count <= 4:
        number = face_count
    else:
        raise ValueError("face_count must be >= 0 and <= 4")
    return number


def create_fake_face(target_id: UUID, max_x: float, max_y: float) -> FacesCreate:
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

    hid = str(uuid4())
    radii = create_fake_radii_list(random.randint(3, 10))
    data = FacesCreate(
        target_id=target_id,
        x=create_fake_axis(biggest_radii=max(radii), max_allowed=max_x),
        y=create_fake_axis(biggest_radii=max(radii), max_allowed=max_y),
        radii=radii,
        points=create_fake_points_list(len(radii)),
        human_identifier=hid,
    )

    return data


async def insert_faces_db(db_pool: Pool, face_rows: list[FacesCreate]) -> list[FacesRead]:
    """Insert faces and return their FacesRead rows."""
    results: list[FacesRead] = []
    if not face_rows:
        return results

    insert_sql = """
        INSERT INTO faces (x, y, human_identifier, radii, points, target_id)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    """
    async with db_pool.acquire() as conn:
        async with conn.transaction():
            stmt = await conn.prepare(insert_sql)
            for face in face_rows:
                row = await stmt.fetchrow(
                    face.x,
                    face.y,
                    face.human_identifier,
                    face.radii,
                    face.points,
                    face.target_id,
                )
                assert row is not None
                payload = face.model_dump(mode="json", by_alias=True)
                payload["id"] = row["id"]
                results.append(FacesRead(**payload))
    return results


async def create_many_faces(
    db_pool: Pool,
    target_id: UUID,
    max_x: float,
    max_y: float,
    face_count: int = 0,
) -> list[FacesRead]:
    faces: list[FacesCreate] = []
    # Decide face count (respect overall 0..4 rule)
    no_faces = valid_no_faces(face_count)
    for _ in range(no_faces):
        faces.append(create_fake_face(target_id, max_x, max_y))
    results = await insert_faces_db(db_pool, faces)
    return results
