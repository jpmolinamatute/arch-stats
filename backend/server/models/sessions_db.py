from asyncpg import Pool

from server.models.base_db import DBBase
from server.schema import SessionsCreate, SessionsRead, SessionsUpdate


# pylint: disable=too-few-public-methods
class SessionsDB(DBBase[SessionsCreate, SessionsUpdate, SessionsRead]):
    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY,
            is_opened BOOLEAN NOT NULL,
            start_time TIMESTAMP WITH TIME ZONE NOT NULL,
            location VARCHAR(255) NOT NULL,
            end_time TIMESTAMP WITH TIME ZONE,
            CONSTRAINT open_session_no_end_time CHECK (
                (is_opened = TRUE AND end_time IS NULL) OR (is_opened = FALSE)
            ),
            CONSTRAINT closed_session_with_end_time CHECK (
                (is_opened = FALSE AND end_time IS NOT NULL) OR (is_opened = TRUE)
            )
        """
        super().__init__("sessions", schema, SessionsRead, db_pool)
