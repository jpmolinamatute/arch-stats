from uuid import UUID

from asyncpg import Pool

from models.parent_model import DBNotFound, ParentModel
from schema import (
    BowStyleType,
    SlotCreate,
    SlotFilter,
    SlotJoinRequest,
    SlotLeaveRequest,
    SlotLetterType,
    SlotRead,
    SlotSet,
    TargetFaceType,
    TargetRead,
)


class SlotModel(ParentModel[SlotCreate, SlotSet, SlotRead, SlotFilter]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("slot", db_pool, SlotRead)

    # pylint: disable=[too-many-arguments]
    async def create_one(
        self,
        *,
        session_id: UUID,
        target_id: UUID,
        archer_id: UUID,
        face_type: TargetFaceType,
        slot_letter: SlotLetterType,
        bowstyle: BowStyleType,
        draw_weight: float,
        club_id: UUID | None,
    ) -> UUID:

        new_slotassignment = SlotCreate(
            session_id=session_id,
            target_id=target_id,
            archer_id=archer_id,
            face_type=face_type,
            slot_letter=slot_letter,
            is_shooting=True,
            bowstyle=bowstyle,
            draw_weight=draw_weight,
            club_id=club_id,
        )
        return await self.insert_one(new_slotassignment)

    async def get_open_participants(self) -> list[SlotRead]:
        """Fetch participants currently shooting in open session (from open_participants view)."""
        sql = "SELECT * FROM open_participants;"
        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(sql)
        return [SlotRead(**row) for row in rows]

    async def get_available_targets(self, req_data: SlotJoinRequest) -> list[TargetRead]:
        """
        Fetch available target for a session and distance (from get_available_targets function).
        """
        sql = """
            SELECT * FROM get_available_targets($1, $2);
        """
        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(sql, req_data.session_id, req_data.distance)
        return [TargetRead(**row) for row in rows]

    async def get_open_participants_by_session(self, session_id: UUID) -> list[SlotRead]:
        """Fetch open participants for a given session from view open_participants."""
        sql = "SELECT * FROM open_participants WHERE session_id = $1;"
        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(sql, session_id)
        return [SlotRead(**row) for row in rows]

    async def stop_all_in_session(self, session_id: UUID) -> None:
        """Mark all slot assignments in a session as not shooting."""
        data = SlotSet(is_shooting=False)
        where = SlotFilter(session_id=session_id, is_shooting=True)
        await self.update(data, where)

    async def leave_session(self, session_req: SlotLeaveRequest) -> None:
        """Mark a specific archer's assignment in a session as not shooting."""

        data = SlotSet(is_shooting=False)
        where = SlotFilter(
            session_id=session_req.session_id,
            is_shooting=True,
            archer_id=session_req.archer_id,
        )
        await self.update(data, where)

    async def get_assigned_letters(self, target_id: UUID) -> set[SlotLetterType]:
        """Return the set of slot letters currently in use (is_shooting=TRUE) for a target."""

        where = SlotFilter(target_id=target_id, is_shooting=True)
        rows = await self.get_all(where, [])

        return {SlotLetterType(row.slot_letter) for row in rows}

    async def get_one(self, where: SlotFilter) -> SlotRead:
        """Fetch a single slot with computed slot identifier."""
        select_stm = "SELECT * FROM get_slot_with_lane($1, $2);"
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Fetching: %s", select_stm)
            row = await conn.fetchrow(select_stm, where.archer_id, where.session_id)
            if not row:
                raise DBNotFound(f"{self.name}: No record found")
            return self.read_schema(**row)
