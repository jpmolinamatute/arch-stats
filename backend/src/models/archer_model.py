from asyncpg import Pool

from models.parent_model import ParentModel
from schema import ArcherCreate, ArcherFilter, ArcherRead, ArcherSet


class ArcherModel(ParentModel[ArcherCreate, ArcherSet, ArcherRead, ArcherFilter]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("archer", db_pool, ArcherRead)

    async def get_by_google_subject(self, sub: str) -> ArcherRead | None:
        where = ArcherFilter(google_subject=sub)
        select_stm, params = self.build_select_sql_stm(where, [])
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Fetching: %s", select_stm)
            row = await conn.fetchrow(select_stm, *params)
            return None if row is None else self.read_schema(**row)
