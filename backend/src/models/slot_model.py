from uuid import UUID

from asyncpg import Pool

from models.parent_model import DBNotFound, ParentModel
from schema import (
    BowStyleType,
    FaceType,
    FullSlotInfo,
    SlotCreate,
    SlotFilter,
    SlotJoinRequest,
    SlotLeaveRequest,
    SlotLetterType,
    SlotRead,
    SlotSet,
    TargetRead,
)


class SlotModel(ParentModel[SlotCreate, SlotSet, SlotRead, SlotFilter]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("slot", db_pool, SlotRead)

    async def refresh_open_participants(self) -> None:
        """Refresh the materialized view that backs open participants lists.

        Tries a concurrent refresh first (requires a suitable unique index on the
        materialized view). Falls back to a regular refresh if concurrent refresh
        is not supported in the current database state.
        """
        # try:
        #     await self.execute("REFRESH MATERIALIZED VIEW CONCURRENTLY open_participants;")
        # except Exception as exc:  # DBException or asyncpg errors
        #     self.logger.debug(
        #         "Concurrent refresh failed for open_participants, falling back. Reason: %s",
        #         exc,
        #     )
        await self.execute("REFRESH MATERIALIZED VIEW open_participants;")

    # pylint: disable=[too-many-arguments]
    async def create_one(
        self,
        *,
        session_id: UUID,
        target_id: UUID,
        archer_id: UUID,
        face_type: FaceType,
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
        slot_id = await self.insert_one(new_slotassignment)
        # await self.listen_for_shots(slot_id)
        # Ensure open_participants reflects the latest mutation
        await self.refresh_open_participants()
        return slot_id

    async def get_available_targets(self, req_data: SlotJoinRequest) -> list[TargetRead]:
        """
        Fetch available target for a session and distance (from get_available_targets function).
        """
        sql = self.sql_builder.build_select_function("get_available_targets", 2)
        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(sql, req_data.session_id, req_data.distance)
        return [TargetRead(**row) for row in rows]

    async def stop_all_in_session(self, session_id: UUID) -> None:
        """Mark all slot assignments in a session as not shooting."""
        data = SlotSet(is_shooting=False)
        where = SlotFilter(session_id=session_id, is_shooting=True)
        await self.update(data, where)
        await self.refresh_open_participants()

    async def leave_session(self, session_req: SlotLeaveRequest) -> None:
        """Mark a specific archer's assignment in a session as not shooting."""

        data = SlotSet(is_shooting=False)
        where = SlotFilter(
            session_id=session_req.session_id,
            is_shooting=True,
            archer_id=session_req.archer_id,
        )
        await self.update(data, where)
        await self.refresh_open_participants()

    async def get_assigned_letters(self, target_id: UUID) -> set[SlotLetterType]:
        """Return the set of slot letters currently in use (is_shooting=TRUE) for a target."""

        where = SlotFilter(target_id=target_id, is_shooting=True)
        rows = await self.get_all(where, [])

        return {SlotLetterType(row.slot_letter) for row in rows}

    async def get_slot_with_lane(self, slot_id: UUID) -> SlotRead:
        """Fetch a single slot with computed slot identifier."""
        select_stm = self.sql_builder.build_select_function("get_slot_with_lane", 1)
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Fetching: %s", select_stm)
            row = await conn.fetchrow(select_stm, slot_id)
            if not row:
                raise DBNotFound(f"{self.name}: No record found")
            return self.read_schema(**row)

    async def get_full_slot_info(
        self, *, slot_id: UUID | None = None, archer_id: UUID | None = None
    ) -> FullSlotInfo:
        """Return one FullSlotInfo row from the open_participants view.

        Exactly one selector must be provided.

        Args:
            slot_id: Fetch by slot identifier when provided.
            archer_id: Fetch the active slot for this archer when provided.

        Returns:
            FullSlotInfo: Complete view row for the selected participant.

        Raises:
            ValueError: If both or neither of slot_id/archer_id are provided.
            DBNotFound: If no matching row exists in the view.
        """
        if (slot_id is None and archer_id is None) or (
            slot_id is not None and archer_id is not None
        ):
            raise ValueError("Provide exactly one of slot_id or archer_id")

        if slot_id is not None:
            filters = SlotFilter(slot_id=slot_id)
        else:
            filters = SlotFilter(archer_id=archer_id)

        # Mirror ParentModel.build_select_sql_stm but target the view
        dump = filters.model_dump(by_alias=True, exclude_unset=True, exclude_none=True)
        keys = list(dump.keys())
        values = tuple(dump.values()) if dump else ()
        conditions = [f"{key} = ${i}" for i, key in enumerate(keys, start=1)]

        sql = self.sql_builder.build_select_view(
            view_name="open_participants",
            order_by_clause="",
            columns=[],
            conditions=conditions,
            limit=1,
        )
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Fetching: %s", sql)
            row = await conn.fetchrow(sql, *values)
            if row is None:
                raise DBNotFound(f"{self.name}: No record found")
            return FullSlotInfo(**row)
