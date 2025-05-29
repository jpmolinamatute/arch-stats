from asyncpg import Pool

from server.models.base_db import DBBase
from server.schema import ArrowsCreate, ArrowsUpdate


class ArrowsDB(DBBase[ArrowsCreate, ArrowsUpdate]):
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
        super().__init__("arrows", schema, db_pool)
