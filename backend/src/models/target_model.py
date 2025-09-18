from uuid import UUID

from asyncpg import Pool

from models.parent_model import ParentModel
from schema import TargetCreate, TargetFilter, TargetRead, TargetSet


class TargetModel(ParentModel[TargetCreate, TargetSet, TargetRead, TargetFilter]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("target", db_pool, TargetRead)

    async def create_one(self, session_id: UUID, distance: int, lane: int) -> UUID:
        new_target = TargetCreate(session_id=session_id, distance=distance, lane=lane)
        return await self.insert_one(new_target)

    async def get_next_empty_lane(self, session_id: UUID) -> int:
        """Get the next empty lane for a session by executing get_next_lane SQL function."""
        sql = "SELECT get_next_lane($1) AS next_lane;"
        async with self.db_pool.acquire() as conn:
            row = await conn.fetchrow(sql, session_id)
        # Returns 1 if no lanes exist for the session (function returns NULL)
        return row["next_lane"] if row and row["next_lane"] is not None else 1

    async def get_lane(self, target_id: UUID) -> int | None:
        where = TargetFilter(target_id=target_id)
        sql_statement, params = self.build_select_sql_stm(where, ["lane"], 1)
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Fetching: %s", sql_statement)
            row = await conn.fetchrow(sql_statement, *params)
        return row["lane"] if row else None
