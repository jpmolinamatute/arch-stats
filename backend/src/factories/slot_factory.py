"""
Factory for creating SlotAssignment records in the database for testing or seeding.

Matches migration V006 (slots table), inserting:
- archer_id (FK), target_id (FK), session_id (FK)
- face_type (enum `face_type`), slot_letter (A-D), is_shooting (bool)

Also enforces in-factory:
- (archer_id, session_id) uniqueness to avoid violating uq_archer_per_session
- Choose targets that belong to the selected session when the provided lists overlap
"""

import random
from collections.abc import Sequence
from uuid import UUID

from asyncpg import Pool


_FACE_TYPES = [
    "40cm_full",
    "60cm_full",
    "80cm_full",
    "122cm_full",
    "40cm_6rings",
    "60cm_6rings",
    "80cm_6rings",
    "122cm_6rings",
    "40cm_triple_vertical",
    "60cm_triple_triangular",
    "none",
]

# Valid slot letters for a target (module-level to avoid extra locals in the function)
SLOT_LETTERS = ["A", "B", "C", "D"]


async def create_slot_assignments(
    pool: Pool,
    qty: int,
    archer_ids: Sequence[UUID],
    target_ids: Sequence[UUID],
    session_ids: Sequence[UUID],
) -> Sequence[UUID]:
    """
    Create `qty` SlotAssignment records in the database.

    Args:
        pool: The asyncpg Pool to use for DB access.
        qty: Number of slot assignments to create.
        archer_ids: List of archer IDs to assign.
        target_ids: List of target IDs to assign.
        session_ids: List of session IDs to assign.
    Returns:
        List of inserted slot assignment IDs.
    """
    slot_ids: list[UUID] = []
    if not archer_ids or not target_ids or not session_ids:
        return slot_ids

    async with pool.acquire() as conn:
        rows = await conn.fetch(
            "SELECT target_id, session_id FROM target WHERE target_id = ANY($1::uuid[])",
            list(target_ids),
        )
        target_to_session: dict[UUID, UUID] = {row["target_id"]: row["session_id"] for row in rows}
        filtered_targets: list[UUID] = [
            tid
            for tid in target_ids
            if tid in target_to_session and target_to_session[tid] in session_ids
        ]
        if not filtered_targets:
            filtered_targets = list(target_to_session.keys())
        if not filtered_targets:
            return slot_ids

        used_pairs: set[tuple[UUID, UUID]] = set()

        for _ in range(qty * 10):
            if len(slot_ids) >= qty:
                break
            archer_id = random.choice(archer_ids)
            target_id = random.choice(filtered_targets)
            session_id = target_to_session.get(target_id)
            if session_id is None or (archer_id, session_id) in used_pairs:
                continue

            rec = await conn.fetchrow(
                """
                INSERT INTO slot (
                    archer_id, target_id, session_id, face_type, slot_letter, is_shooting
                )
                VALUES ($1, $2, $3, $4, $5, $6)
                RETURNING slot_id
                """,
                archer_id,
                target_id,
                session_id,
                random.choice(_FACE_TYPES),
                random.choice(SLOT_LETTERS),
                bool(random.getrandbits(1)),
            )
            if rec is None or "slot_id" not in rec:
                continue
            used_pairs.add((archer_id, session_id))
            slot_ids.append(rec["slot_id"])

    return slot_ids
