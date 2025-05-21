from uuid import UUID

from asyncpg import Pool

from database.base import DBBase
from database.schema import ArrowCreate, ArrowRead, ArrowUpdate


# pylint: disable=too-few-public-methods
class ArrowsDB(DBBase[ArrowCreate, ArrowUpdate, ArrowRead]):
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
        super().__init__("arrows", schema, ArrowRead, db_pool)

    async def insert_one(self, data: ArrowCreate) -> None:
        await super().insert_one(data)

    async def update_one(self, _id: UUID, data: ArrowUpdate) -> None:
        await super().update_one(_id, data)
