"""
Factory for creating Archer records in the database for testing or seeding.

Matches migration V002 (archer table) required fields.
"""

import random
from datetime import date, timedelta
from uuid import UUID, uuid4

from asyncpg import Pool

from schema import BowStyleType, GenderType

_BOWSTYLE = [style.value for style in BowStyleType]
_GENDERS = [gender.value for gender in GenderType]


async def create_archers(pool: Pool, qty: int) -> list[UUID]:
    """
    Create `qty` Archer records in the database.

    Args:
        pool: The asyncpg Pool to use for DB access.
        qty: Number of archers to create.
    Returns:
        List of inserted archer UUIDs.
    """
    archer_ids: list[UUID] = []
    async with pool.acquire() as conn:
        for i in range(qty):
            # Deterministic but unique values
            uid = uuid4()
            first_name = f"Archer{i}"
            last_name = f"Test{i}"
            email = f"archer{i}-{uid.hex[:8]}@example.com"
            dob = date.today() - timedelta(days=10000 + i)  # Ensure <= current_date
            gender = random.choice(_GENDERS)
            bowstyle = random.choice(_BOWSTYLE)
            google_subject = f"gs-{uid}"
            google_picture_url = f"https://example.com/pic/{uid.hex}.png"

            result = await conn.fetchrow(
                """
                INSERT INTO archer (
                    email, first_name, last_name, date_of_birth, gender, bowstyle,
                    draw_weight, google_picture_url, google_subject
                )
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                RETURNING archer_id
                """,
                email,
                first_name,
                last_name,
                dob,
                gender,
                bowstyle,
                40.0,
                google_picture_url,
                google_subject,
            )
            if result is not None and "archer_id" in result:
                archer_ids.append(result["archer_id"])  # UUID
    return archer_ids
