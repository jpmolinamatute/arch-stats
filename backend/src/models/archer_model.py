from asyncpg import Pool

from models.parent_model import DBNotFound, ParentModel
from schema import ArcherCreate, ArcherFilter, ArcherRead, ArcherSet


class ArcherModel(ParentModel[ArcherCreate, ArcherSet, ArcherRead, ArcherFilter]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("archer", db_pool, ArcherRead)

    async def get_by_google_subject(self, sub: str) -> ArcherRead | None:
        where = ArcherFilter(google_subject=sub)
        query_data = self.build_select_sql_stm(where, [])
        try:
            row = await self.fetchrow(query_data)
        except DBNotFound:
            return None
        return self.read_schema(**row)
