from asyncpg import Pool

from shared.schema import SessionsCreate, SessionsRead, SessionsUpdate
from shared.models.base_db import DBBase, DBNotFound


# pylint: disable=too-few-public-methods
class SessionsDB(DBBase[SessionsCreate, SessionsUpdate, SessionsRead]):
    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            is_opened BOOLEAN NOT NULL,
            start_time TIMESTAMP WITH TIME ZONE NOT NULL,
            location VARCHAR(255) NOT NULL,
            end_time TIMESTAMP WITH TIME ZONE,
            distance INTEGER DEFAULT 0,
            is_indoor BOOLEAN DEFAULT FALSE,
            CONSTRAINT open_session_no_end_time CHECK (
                (is_opened = TRUE AND end_time IS NULL) OR (is_opened = FALSE)
            ),
            CONSTRAINT closed_session_with_end_time CHECK (
                (is_opened = FALSE AND end_time IS NOT NULL) OR (is_opened = TRUE)
            )
        """
        super().__init__("sessions", schema, db_pool, SessionsRead)

    async def get_open_session(self) -> SessionsRead | None:
        result = None
        try:
            result = await self.get_one({"is_opened": True})
        except DBNotFound:
            pass
        return result
