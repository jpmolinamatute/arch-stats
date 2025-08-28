from asyncpg import Pool

from shared.models.base_db import DBBase
from shared.schema import ArrowsCreate, ArrowsRead, ArrowsUpdate


class ArrowsDB(DBBase[ArrowsCreate, ArrowsUpdate, ArrowsRead]):
    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY,
            length REAL NOT NULL,
            human_identifier VARCHAR(10) NOT NULL,
            is_programmed BOOLEAN NOT NULL DEFAULT FALSE,
            registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            voided_date TIMESTAMP WITH TIME ZONE,
            is_active BOOLEAN NOT NULL DEFAULT TRUE,
            label_position REAL,
            weight REAL,
            diameter REAL,
            spine REAL,
            UNIQUE (human_identifier),
            CONSTRAINT arrows_active_voided_consistency CHECK (
                (is_active = FALSE AND voided_date IS NOT NULL)
                OR (is_active = TRUE AND voided_date IS NULL)
            )
        """
        super().__init__("arrows", schema, db_pool, ArrowsRead)
