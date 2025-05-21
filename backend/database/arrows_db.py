from uuid import UUID

from asyncpg import Pool

from database.base_db import DBBase
from database.schema import ArrowsCreate, ArrowsRead, ArrowsUpdate


class ArrowsDB(DBBase[ArrowsCreate, ArrowsUpdate, ArrowsRead]):
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
        super().__init__("arrows", schema, ArrowsRead, db_pool)

    async def insert_one(self, data: ArrowsCreate) -> None:
        await super().insert_one(data)

    async def update_one(self, _id: UUID, data: ArrowsUpdate) -> None:
        await super().update_one(_id, data)
