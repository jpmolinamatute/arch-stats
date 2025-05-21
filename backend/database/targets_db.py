from asyncpg import Pool

from database.base import DBBase
from database.schema import TargetsCreate, TargetsRead, TargetsUpdate


# pylint: disable=too-few-public-methods


class TargetsDB(DBBase[TargetsCreate, TargetsUpdate, TargetsRead]):
    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY,
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
        super().__init__("targets", schema, TargetsRead, db_pool)
