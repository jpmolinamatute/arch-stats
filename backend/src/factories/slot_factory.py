"""
Factory for creating SlotAssignment records in the database for testing or seeding.

Matches migration V006 (slots table), inserting:
- archer_id (FK), target_id (FK), session_id (FK)
- face_type (enum `face_type`), slot_letter (A-D), is_shooting (bool)

Also enforces in-factory:
- (archer_id, session_id) uniqueness to avoid violating uq_archer_per_session
- Choose targets that belong to the selected session when the provided lists overlap

Refactor note:
- If `archer_ids` is None or empty, auto-create `qty` archers using archer_factory.
- If `session_ids` is None or empty, auto-create ONE session using session_factory.
- If `target_ids` is None or empty, auto-create ONE target (on the ensured session)
    using target_factory.
This reduces duplication and makes the factory easier to use in tests/seeders.
"""

import random
from typing import Final
from uuid import UUID, uuid4

from asyncpg import Pool

# Reuse other factories to ensure required related records exist when not provided
from factories.archer_factory import create_archers
from factories.session_factory import create_sessions
from factories.target_factory import create_targets
from schema import BowStyleType, FaceType, SlotLetterType

_BOWSTYLE = [style.value for style in BowStyleType]
_SLOT_LETTERS = [letter.value for letter in SlotLetterType]
_FACE_TYPES = [face.value for face in FaceType]

CLUB_ID_ASSIGNMENT_PROBABILITY: Final[float] = 0.7


async def get_archers(pool: Pool, qty: int, archer_ids: list[UUID] | None) -> list[UUID]:
    """
    Ensure we have exactly `qty` archer IDs.

    If `archer_ids` is None or empty, create `qty` archers using the archer factory.
    If provided, validate its length matches `qty`.

    Args:
        pool: The asyncpg Pool to use for DB access.
        qty: Number of archers required.
        archer_ids: Optional list of archer IDs to use as-is.

    Returns:
        A sequence of archer UUIDs of length `qty`.

    Raises:
        ValueError: If provided `archer_ids` length doesn't match `qty`.
    """
    if not archer_ids:
        archer_ids = await create_archers(pool, qty)
    if len(archer_ids) != qty:
        raise ValueError("qty must match len(archer_ids)")
    return archer_ids


async def create_slot_assignments(
    pool: Pool,
    qty: int,
    archer_ids: list[UUID] | None = None,
    target_id: UUID | None = None,
    session_id: UUID | None = None,
) -> list[UUID]:
    """
    Create `qty` SlotAssignment records in the database.

    Args:
        pool: The asyncpg Pool to use for DB access.
        qty: Number of slot assignments to create.
        archer_ids: Archer IDs to assign; if None, `qty` archers will be created.
        target_id: Target ID to use; if None, one target will be created.
        session_id: Session ID to use; if None, one session will be created.
    Returns:
        List of inserted slot assignment IDs.
    """
    slot_ids: list[UUID] = []
    if qty <= 0:
        return slot_ids

    # Ensure required related records exist, reusing other factories.
    # 1) Ensure archers: need exactly `qty` archers to match requested slots.
    archer_ids = await get_archers(pool, qty, archer_ids)

    # 2) Ensure a session: only one needed.
    if not session_id:
        created_sessions = await create_sessions(pool, 1)
        session_id = created_sessions[0]

    # 3) Ensure a target for that session: only one needed.
    if not target_id:
        created_targets = await create_targets(pool, 1, session_id)
        target_id = created_targets[0]

    async with pool.acquire() as conn:
        # Ensure caller did not pass duplicate archers for the same session
        # which would violate the (archer_id, session_id) uniqueness.
        if len(set(archer_ids)) != len(archer_ids):
            raise ValueError("archer_ids must be unique for a given session")

        for archer_id in archer_ids:
            # Generate plausible test values for new columns
            bowstyle = random.choice(_BOWSTYLE)
            draw_weight = round(random.uniform(10.0, 60.0), 1)  # 10.0–60.0 kg
            # club_id is optional, randomly assign None or a fake UUID
            club_id = None
            if random.random() < CLUB_ID_ASSIGNMENT_PROBABILITY:
                # 70% chance to assign a club_id (simulate real data)
                club_id = uuid4()

            rec = await conn.fetchrow(
                """
                INSERT INTO slot (
                    archer_id, target_id, session_id, face_type, slot_letter, is_shooting,
                    bowstyle, draw_weight, club_id
                )
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                RETURNING slot_id
                """,
                archer_id,
                target_id,
                session_id,
                random.choice(_FACE_TYPES),
                random.choice(_SLOT_LETTERS),
                bool(random.getrandbits(1)),
                bowstyle,
                draw_weight,
                club_id,
            )
            if rec is None or "slot_id" not in rec:
                raise RuntimeError(
                    f"Failed to create slot for archer_id={archer_id} in session_id={session_id}"
                )
            slot_ids.append(rec["slot_id"])

    return slot_ids
