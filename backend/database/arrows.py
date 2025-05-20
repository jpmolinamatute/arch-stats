from asyncpg import Pool

from database.base import DBBase
from database.schema import ArrowBase

# pylint: disable=too-few-public-methods


class ArrowsDB(DBBase[ArrowBase]):
    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY,
            length REAL NOT NULL,
            human_identifier VARCHAR(10) NOT NULL,
            is_programmed BOOLEAN NOT NULL DEFAULT FALSE,
            label_position REAL,
            weight REAL,
            diameter REAL,
            spine REAL,
            UNIQUE (human_identifier)
        """
        super().__init__("arrows", schema, ArrowBase, db_pool)
