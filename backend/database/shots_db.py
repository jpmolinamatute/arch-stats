from asyncpg import Pool

from database.base import DBBase
from database.schema import ShotsCreate, ShotsRead, ShotsUpdate


# pylint: disable=too-few-public-methods
class ShotsDB(DBBase[ShotsCreate, ShotsUpdate, ShotsRead]):
    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY,
            arrow_id UUID NOT NULL,
            arrow_engage_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrow_disengage_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrow_landing_time TIMESTAMP WITH TIME ZONE,
            x_coordinate REAL,
            y_coordinate REAL,
            FOREIGN KEY (arrow_id) REFERENCES arrows (id) ON DELETE CASCADE
        """
        super().__init__("shots", schema, ShotsRead, db_pool)
