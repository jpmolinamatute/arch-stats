from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import DBException, ParentModel
from shared.schema import ArrowsCreate, ArrowsFilters, ArrowsRead, ArrowsUpdate


class ArrowsModel(ParentModel[ArrowsCreate, ArrowsUpdate, ArrowsRead, ArrowsFilters]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("arrows", db_pool, ArrowsRead)

    async def create(self) -> None:
        schema = """
            id UUID PRIMARY KEY,
            length REAL NOT NULL,
            human_identifier VARCHAR(10) NOT NULL,
            is_programmed BOOLEAN NOT NULL DEFAULT FALSE,
            registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            voided_date TIMESTAMP WITH TIME ZONE,
            is_active BOOLEAN NOT NULL DEFAULT TRUE,
            label_position REAL,
            weight REAL,
            diameter REAL,
            spine REAL,
            CONSTRAINT arrows_active_voided_consistency CHECK (
                (is_active = FALSE AND voided_date IS NOT NULL)
                OR (is_active = TRUE AND voided_date IS NULL)
            )
        """
        await self.execute(f"CREATE TABLE IF NOT EXISTS {self.name} ({schema});")
        await self.execute(
            f"""
            CREATE UNIQUE INDEX IF NOT EXISTS {self.name}_uniq_active_human_identifier
            ON {self.name} (human_identifier)
            WHERE is_active IS TRUE;
            """
        )

    async def drop(self) -> None:
        await self.execute(f"DROP TABLE IF EXISTS {self.name} CASCADE;")

    async def get_one_by_id(self, _id: UUID) -> ArrowsRead:
        """
        Retrieve a single record by its UUID.
        """
        if not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        where = ArrowsFilters(id=_id)
        return await self.get_one(where)
