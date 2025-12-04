from uuid import UUID

from asyncpg import Pool

from models.parent_model import DBNotFound, ParentModel
from schema import TargetCreate, TargetFilter, TargetRead, TargetSet


class TargetModel(ParentModel[TargetCreate, TargetSet, TargetRead, TargetFilter]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("target", db_pool, TargetRead)

    async def create_one(self, session_id: UUID, distance: int, lane: int) -> UUID:
        new_target = TargetCreate(session_id=session_id, distance=distance, lane=lane)
        return await self.insert_one(new_target)

    async def get_next_empty_lane(self, session_id: UUID) -> int:
        """Get the next empty lane for a session by executing get_next_lane SQL function."""
        sql, params = self.build_select_function_sql_stm("get_next_lane", [session_id])
        try:
            row = await self.fetchrow((sql, params))
        except DBNotFound:
            row = None
        # Returns 1 if no lanes exist for the session (function returns NULL)
        return row["get_next_lane"] if row and row["get_next_lane"] is not None else 1

    async def get_lane(self, target_id: UUID) -> int | None:
        where = TargetFilter(target_id=target_id)
        sql_statement, params = self.build_select_sql_stm(where, ["lane"])
        try:
            row = await self.fetchrow((sql_statement, params))
        except DBNotFound:
            row = None
        return row["lane"] if row else None
