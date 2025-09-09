from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import DBException, ParentModel
from shared.schema import ArrowsCreate, ArrowsFilters, ArrowsRead, ArrowsUpdate


class ArrowsModel(ParentModel[ArrowsCreate, ArrowsUpdate, ArrowsRead, ArrowsFilters]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("arrows", db_pool, ArrowsRead)

    async def create(self) -> None:
        """Create the arrows table and related indexes."""
        # - Enforce consistency via CHECK: is_active <-> voided_date presence.
        # - Add a partial UNIQUE index on human_identifier for active rows
        #   (WHERE is_active IS TRUE).
        schema = """
            id UUID PRIMARY KEY,
            length REAL NOT NULL,
            human_identifier VARCHAR(10) NOT NULL,
            is_programmed BOOLEAN NOT NULL DEFAULT FALSE,
            registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
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
        """.strip()
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Creating table %s", self.name)
            await conn.execute(f"CREATE TABLE IF NOT EXISTS {self.name} ({schema});")
            self.logger.debug("Creating index %s", f"idx_{self.name}_uniq_hid")
            await conn.execute(
                f"""
                CREATE UNIQUE INDEX IF NOT EXISTS idx_{self.name}_uniq_hid
                ON {self.name} (human_identifier)
                WHERE is_active IS TRUE;
            """
            )

    async def drop(self) -> None:
        """Drop the arrows table and its partial unique index idempotently."""
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Dropping index %s", f"idx_{self.name}_uniq_hid")
            await conn.execute(f"DROP INDEX IF EXISTS idx_{self.name}_uniq_hid;")

            self.logger.debug("Dropping table %s", self.name)
            await conn.execute(f"DROP TABLE IF EXISTS {self.name};")

    async def get_one_by_id(self, _id: UUID) -> ArrowsRead:
        """Fetch a single arrow by id.

        Args:
            _id: Arrow identifier.

        Returns:
            Arrow row validated as ArrowsRead.

        Raises:
            DBException: If the provided id is invalid.
        """
        if not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        where = ArrowsFilters(id=_id)
        return await self.get_one(where)
