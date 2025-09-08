from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import DBException, ParentModel
from shared.schema import TargetsCreate, TargetsFilters, TargetsRead, TargetsUpdate


class TargetsModel(ParentModel[TargetsCreate, TargetsUpdate, TargetsRead, TargetsFilters]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("targets", db_pool, TargetsRead)

    async def create(self) -> None:
        schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            max_x REAL NOT NULL,
            max_y REAL NOT NULL,
            distance INTEGER NOT NULL,
            session_id UUID NOT NULL,
            CONSTRAINT targets_one_per_session UNIQUE (session_id),
            FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
        """
        async with self.db_pool.acquire() as conn:
            await conn.execute(f"CREATE TABLE IF NOT EXISTS {self.name} ({schema});")
            await conn.execute(
                f"""
                    CREATE INDEX IF NOT EXISTS idx_{self.name}_session_id
                    ON {self.name} (session_id);
                """
            )

    async def drop(self) -> None:
        async with self.db_pool.acquire() as conn:
            await conn.execute(f"DROP INDEX IF EXISTS idx_{self.name}_session_id;")
            await conn.execute(f"DROP TABLE IF EXISTS {self.name};")

    async def get_by_session_id(self, session_id: UUID) -> list[TargetsRead]:
        where = TargetsFilters(session_id=session_id)
        return await self.get_all(where)

    async def get_one_by_id(self, _id: UUID) -> TargetsRead:
        """
        Retrieve a single record by its UUID.
        """
        if not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        where = TargetsFilters(id=_id)
        return await self.get_one(where)
