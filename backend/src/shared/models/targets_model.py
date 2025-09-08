from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import DBException, ParentModel
from shared.schema import TargetsCreate, TargetsFilters, TargetsRead, TargetsUpdate


class TargetsModel(ParentModel[TargetsCreate, TargetsUpdate, TargetsRead, TargetsFilters]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("targets", db_pool, TargetsRead)

    async def create(self) -> None:
        """Create the targets table and its session_id index idempotently.

        - Enforces one target per session via UNIQUE(session_id).
        - Adds FK to sessions(id) with ON DELETE CASCADE.
        - Creates idx_targets_session_id to speed up session queries.
        """
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
        """Drop the targets table and its session_id index idempotently."""
        async with self.db_pool.acquire() as conn:
            await conn.execute(f"DROP INDEX IF EXISTS idx_{self.name}_session_id;")
            await conn.execute(f"DROP TABLE IF EXISTS {self.name};")

    async def get_by_session_id(self, session_id: UUID) -> list[TargetsRead]:
        """Fetch all targets for a given session id.

        Args:
            session_id: Session identifier.

        Returns:
            List of TargetsRead entries.
        """
        where = TargetsFilters(session_id=session_id)
        return await self.get_all(where)

    async def get_one_by_id(self, _id: UUID) -> TargetsRead:
        """Fetch a single target by id.

        Args:
            _id: Target identifier.

        Returns:
            Target row validated as TargetsRead.

        Raises:
            DBException: If the provided id is invalid.
        """
        if not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        where = TargetsFilters(id=_id)
        return await self.get_one(where)
