from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import DBException, DBNotFound, ParentModel
from shared.schema import SessionsCreate, SessionsFilters, SessionsRead, SessionsUpdate


class SessionsModel(ParentModel[SessionsCreate, SessionsUpdate, SessionsRead, SessionsFilters]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("sessions", db_pool, SessionsRead)

    async def create(self) -> None:
        """Create the sessions table if absent.

        - Columns: UUID PK, is_opened, start_time, location, optional end_time, is_indoor.
        - CHECK: end_time is NULL when opened and NOT NULL when closed.
        """
        schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            is_opened BOOLEAN NOT NULL,
            start_time TIMESTAMP WITH TIME ZONE NOT NULL,
            location VARCHAR(255) NOT NULL,
            end_time TIMESTAMP WITH TIME ZONE,
            is_indoor BOOLEAN DEFAULT FALSE,
            CHECK (
                (is_opened AND end_time IS NULL)
                OR (NOT is_opened AND end_time IS NOT NULL)
            )
        """.strip()
        index = f"""
            CREATE UNIQUE INDEX IF NOT EXISTS {self.name}_one_open_idx
            ON {self.name} (is_opened)
            WHERE is_opened is TRUE
        """.strip()
        async with self.db_pool.acquire() as conn:
            await conn.execute(f"CREATE TABLE IF NOT EXISTS {self.name} ({schema});")
            await conn.execute(index)

    async def drop(self) -> None:
        """Drop the sessions table if it exists."""
        async with self.db_pool.acquire() as conn:
            await conn.execute(f"DROP TABLE IF EXISTS {self.name};")
            await conn.execute(f"DROP INDEX IF EXISTS {self.name}_one_open_idx;")

    async def get_open_session(self) -> SessionsRead | None:
        """Fetch the currently open session, if any.

        Returns:
            SessionsRead when found, otherwise None.
        """
        result = None
        try:
            where = SessionsFilters(is_opened=True)
            result = await self.get_one(where)
        except DBNotFound:
            pass
        return result

    async def get_one_by_id(self, _id: UUID) -> SessionsRead:
        """Fetch a single session by id.

        Args:
            _id: Session identifier.

        Returns:
            Session row validated as SessionsRead.

        Raises:
            DBException: If the provided id is invalid.
        """
        if not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        where = SessionsFilters(id=_id)
        return await self.get_one(where)

    async def insert_one(self, data: SessionsCreate) -> UUID:
        """Insert a session ensuring only one open session exists.

        If attempting to insert an open session while another is open, raises
        DBException with a clear message. Database partial unique index acts as
        final guard against race conditions.
        """
        if data.is_opened:
            existing = await self.get_open_session()
            if existing is not None:
                raise DBException("Only one session can be opened at a time.")
        return await super().insert_one(data)

    async def update_one(self, _id: UUID, data: SessionsUpdate) -> None:
        """Update a session enforcing single-open-session invariant.

        - Disallow adding end_time to an open session (DB CHECK also enforces).
        - If setting is_opened True, ensure no *other* session is currently open.
        """
        # Fetch current row first so 404 (DBNotFound) takes precedence
        current = await self.get_one_by_id(_id)

        wants_open = data.is_opened is True
        wants_close = data.is_opened is False

        if wants_open:
            open_session = await self.get_open_session()
            if open_session is not None and open_session.session_id != _id:
                raise DBException("Only one session can be opened at a time.")
            if data.end_time is not None:
                raise DBException("Open session cannot have an end_time.")
            # Re-opening a previously closed session: ensure end_time cleared so CHECK passes
            if current.is_opened is False:
                data.end_time = None

        elif wants_close and current.is_opened and data.end_time is None:
            raise DBException("Closed session must include an end_time.")
        await super().update_one(_id, data)
