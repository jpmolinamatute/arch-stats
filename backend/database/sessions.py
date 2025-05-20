from asyncpg import Pool

from database.base import DBBase


# pylint: disable=too-few-public-methods
class SessionsDB(DBBase):
    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY,
            is_opened BOOLEAN NOT NULL,
            start_time TIMESTAMP WITH TIME ZONE NOT NULL,
            end_time TIMESTAMP WITH TIME ZONE,
            location VARCHAR(255) NOT NULL,
            CONSTRAINT open_session_no_end_time CHECK (
                (is_opened = TRUE AND end_time IS NULL) OR (is_opened = FALSE)
            ),
            CONSTRAINT closed_session_with_end_time CHECK (
                (is_opened = FALSE AND end_time IS NOT NULL) OR (is_opened = TRUE)
            )
        """
        super().__init__("sessions", schema, db_pool)
