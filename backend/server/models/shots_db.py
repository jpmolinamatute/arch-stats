from asyncpg import Pool

from server.models.base_db import DBBase
from server.schema import ShotsCreate, ShotsUpdate


# pylint: disable=too-few-public-methods
class ShotsDB(DBBase[ShotsCreate, ShotsUpdate]):

    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            arrow_id UUID NOT NULL,
            arrow_engage_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrow_disengage_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrow_landing_time TIMESTAMP WITH TIME ZONE,
            x_coordinate REAL,
            y_coordinate REAL,
            FOREIGN KEY (arrow_id) REFERENCES arrows (id) ON DELETE CASCADE
        """
        super().__init__("shots", schema, db_pool)
