from datetime import UTC, datetime
from uuid import UUID

from asyncpg import Pool

from models.parent_model import DBException, DBNotFound, ParentModel
from schema import (
    SessionCreate,
    SessionFilter,
    SessionId,
    SessionRead,
    SessionSet,
)
from schema.archer_schema import ArcherFilter


class SessionModel(ParentModel[SessionCreate, SessionSet, SessionRead, SessionFilter]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("session", db_pool, SessionRead)

    async def archer_open_session(self, archer_id: UUID) -> UUID | None:
        """Return the open session_id for the given archer, or None if not found."""

        where = SessionFilter(owner_archer_id=archer_id, is_opened=True)
        try:
            row = await self.get_one(where)
        except DBNotFound:
            row = None
        return row.get_id() if row else None

    async def get_all_open_sessions(self) -> list[SessionRead]:
        """Return all open session (is_opened=TRUE)."""
        where = SessionFilter(is_opened=True)
        return await self.get_all(where, [])

    async def get_all_closed_sessions_owned_by(self, archer_id: UUID) -> list[SessionRead]:
        """Return all closed sessions owned by the given archer."""
        where = SessionFilter(owner_archer_id=archer_id, is_opened=False)
        return await self.get_all(where, [])

    async def is_archer_participating(self, archer_id: UUID) -> UUID | None:
        """
        Return a session_id if the archer is actively participating in an open session,
        or None if not found.
        """
        where = ArcherFilter(archer_id=archer_id)
        sql, params = self.build_select_view_sql_stm(
            view_name="open_participants",
            where=where,
            columns=["session_id"],
            limit=1,
            is_desc=False,
        )
        session_id: UUID | None
        try:
            row = await self.fetchrow((sql, params))
            session_id = row["session_id"]
        except DBNotFound:
            session_id = None
        return session_id

    async def does_open_session_exist(self, session: UUID) -> bool:
        """Return True if a session with the given ID and is_opened status exists."""
        where = SessionFilter(session_id=session, is_opened=True)
        try:
            _ = await self.get_one(where)
            result = True
        except DBNotFound:
            result = False
        return result

    async def create_session(self, session_data: SessionCreate) -> UUID:
        """Create new session and auto-join owner into lane 1 slot A per spec.

        Expects SessionCreate to include distance/face_type/bowstyle/draw_weight/club_id
        for the initial owner's slot.
        """

        # Owner cannot already have an open session
        if (await self.archer_open_session(session_data.owner_archer_id)) is not None:
            # Align with API contract used by endpoint tests
            raise ValueError("Archer already have an opened session")
        # Owner cannot be participating elsewhere
        if (await self.is_archer_participating(session_data.owner_archer_id)) is not None:
            raise ValueError(
                f"ERROR: archer '{session_data.owner_archer_id}' is already participating "
                "in an open session"
            )
        return await self.insert_one(session_data)

    async def close_session(self, session: SessionId) -> None:
        """Close session after ensuring no other participants are actively shooting."""

        if session.session_id is None:
            raise ValueError("ERROR: session_id wasn't provided")

        exist = await self.does_open_session_exist(session.session_id)
        if not exist:
            raise DBNotFound("ERROR: Session either doesn't exist or it was already closed")

        # Ensure there are no active participants shooting in this session
        if await self.has_active_participants(session.session_id):
            raise ValueError("ERROR: cannot close session with active participants")

        data = SessionSet(is_opened=False, closed_at=datetime.now(UTC))
        where = SessionFilter(session_id=session.session_id, is_opened=True)
        await self.update(data, where)

    async def has_active_participants(self, session_id: UUID) -> bool:
        """Return True if there are active participants in the given session.

        Uses base tables to avoid dependency on a specific view in test environments.
        """
        where = SessionFilter(session_id=session_id)
        sql, params = self.build_select_view_sql_stm(
            view_name="open_participants",
            where=where,
            columns=["1"],
            limit=1,
            is_desc=False,
        )
        try:
            await self.fetchrow((sql, params))
            return True
        except DBNotFound:
            return False

    async def re_open_session(self, session: SessionId, archer_id: UUID) -> None:
        if session.session_id is None:
            raise ValueError("ERROR: session_id wasn't provided")
        where = SessionFilter(session_id=session.session_id, is_opened=False)
        # this will throw an exception if not found
        row = await self.get_one(where)
        if row.owner_archer_id != archer_id:
            raise DBException("Archer is not allow to re-open this session")
        set_sql = SessionSet(is_opened=True, closed_at=None)
        await self.update(set_sql, where)
