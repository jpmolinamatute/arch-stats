from asyncpg import Pool

from database.base import DBBase

# pylint: disable=too-few-public-methods


class ArrowsDB(DBBase):
    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY,
            weight REAL DEFAULT 0.0,
            diameter REAL DEFAULT 0.0,
            spine REAL DEFAULT 0.0,
            length REAL NOT NULL,
            human_identifier VARCHAR(10),
            label_position REAL NOT NULL,
            is_programmed BOOLEAN NOT NULL DEFAULT FALSE,
            UNIQUE (human_identifier)
        """
        super().__init__("arrows", schema, db_pool)
