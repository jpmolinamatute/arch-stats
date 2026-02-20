from uuid import UUID

from asyncpg import Pool

from models.parent_model import ParentModel
from schema import ShotCreate, ShotFilter, ShotRead, ShotSet


class ShotModel(ParentModel[ShotCreate, ShotSet, ShotRead, ShotFilter]):
    """Model for shot-related DB access and notifications."""

    def __init__(self, db_pool: Pool) -> None:
        super().__init__("shot", db_pool, ShotRead)

    async def count_by_slot(self, slot_id: UUID) -> int:
        """Retrieve all shots (count only) for a given slot."""
        sql = f"SELECT COUNT(*) FROM {self.name} WHERE slot_id = $1"
        row = await self.fetchrow((sql, (slot_id,)))
        return row[0]
