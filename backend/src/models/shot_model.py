from asyncpg import Pool

from models.parent_model import ParentModel
from schema import ShotCreate, ShotFilter, ShotRead, ShotSet


class ShotModel(ParentModel[ShotCreate, ShotSet, ShotRead, ShotFilter]):
    """Model for shot-related DB access and notifications."""

    def __init__(self, db_pool: Pool) -> None:
        super().__init__("shot", db_pool, ShotRead)
