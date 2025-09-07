from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import DBException, DBNotFound, ParentModel
from shared.schema import SessionsCreate, SessionsFilters, SessionsRead, SessionsUpdate


# pylint: disable=too-few-public-methods
class SessionsModel(ParentModel[SessionsCreate, SessionsUpdate, SessionsRead, SessionsFilters]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("sessions", db_pool, SessionsRead)

    async def create(self) -> None:
        schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            is_opened BOOLEAN NOT NULL,
            start_time TIMESTAMP WITH TIME ZONE NOT NULL,
            location VARCHAR(255) NOT NULL,
            end_time TIMESTAMP WITH TIME ZONE,
            is_indoor BOOLEAN DEFAULT FALSE,
            CHECK (
                (is_opened AND end_time IS NULL)
                OR (NOT is_opened AND end_time IS NOT NULL)
            )
        """
        await self.execute(f"CREATE TABLE IF NOT EXISTS {self.name} ({schema});")

    async def drop(self) -> None:
        await self.execute(f"DROP TABLE IF EXISTS {self.name} CASCADE;")

    async def get_open_session(self) -> SessionsRead | None:
        result = None
        try:
            where = SessionsFilters(is_opened=True)
            result = await self.get_one(where)
        except DBNotFound:
            pass
        return result

    async def get_one_by_id(self, _id: UUID) -> SessionsRead:
        """
        Retrieve a single record by its UUID.
        """
        if not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        where = SessionsFilters(id=_id)
        return await self.get_one(where)
