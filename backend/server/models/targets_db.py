from asyncpg import Pool

from server.models.base_db import DBBase
from server.schema import TargetsCreate, TargetsRead, TargetsUpdate


# pylint: disable=too-few-public-methods


class TargetsDB(DBBase[TargetsCreate, TargetsUpdate, TargetsRead]):
    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            max_x_coordinate REAL NOT NULL,
            max_y_coordinate REAL NOT NULL,
            radius REAL [] NOT NULL,
            points INT [] NOT NULL,
            height REAL NOT NULL,
            human_identifier VARCHAR(10) NOT NULL,
            session_id UUID NOT NULL,
            CHECK (array_length(radius, 1) = array_length(points, 1)),
            UNIQUE (session_id, human_identifier),
            FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
        """
        super().__init__("targets", schema, db_pool, TargetsRead)
