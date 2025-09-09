from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import DBException, ParentModel
from shared.schema.faces_schema import FacesCreate, FacesFilters, FacesRead, FacesUpdate


class FacesModel(ParentModel[FacesCreate, FacesUpdate, FacesRead, FacesFilters]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("faces", db_pool, FacesRead)
        self.func_name = "validate_face_row"

    async def create(self) -> None:
        """Create the faces table, validation function, and indexes.

        - Defines/updates validate_face_row() used by a CHECK to keep a face
          fully within its target's bounds based on center (x, y) and
          max(radii).
        - Creates the faces table with FK to targets(id), UNIQUE(target_id,
          human_identifier), and radii/points non-empty with equal length.
        - Adds idx_faces_target_id to speed up lookups by target.
        """
        faces_schema = f"""
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            x REAL NOT NULL,
            y REAL NOT NULL,
            human_identifier TEXT NOT NULL,
            radii REAL [] NOT NULL,
            points INTEGER [] NOT NULL,
            target_id UUID NOT NULL,
            FOREIGN KEY (target_id) REFERENCES targets (id) ON DELETE CASCADE,
            UNIQUE (target_id, human_identifier),
            CHECK (
                array_length(radii, 1) > 0
                AND array_length(points, 1) > 0
                AND array_length(points, 1) = array_length(radii, 1)
                AND {self.func_name}(target_id, x, y, radii)
            )
        """.strip()
        function_sql = f"""
            CREATE OR REPLACE FUNCTION {self.func_name}(
                target_id UUID,
                x REAL,
                y REAL,
                radii REAL []
            ) RETURNS BOOLEAN AS $$
            DECLARE
                max_radius REAL;
                tgt_max_x REAL;
                tgt_max_y REAL;
            BEGIN
                SELECT max_x, max_y INTO tgt_max_x, tgt_max_y FROM targets WHERE id = target_id;

                -- Compute max radius
                SELECT MAX(r) INTO max_radius FROM unnest(radii) AS r;

                -- Ensure face fully contained within target bounds (strictly inside)
                IF (x + max_radius) > tgt_max_x OR (y + max_radius) > tgt_max_y THEN
                    RETURN FALSE;
                END IF;
                IF (x - max_radius) < 0 OR (y - max_radius) < 0 THEN
                    RETURN FALSE;
                END IF;

                RETURN TRUE;
            END;
            $$ LANGUAGE plpgsql STABLE;
        """.strip()
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Creating function %s", self.func_name)
            await conn.execute(function_sql)

            self.logger.debug("Creating table %s", self.name)
            await conn.execute(f"CREATE TABLE IF NOT EXISTS {self.name} ({faces_schema});")

            self.logger.debug("Creating index %s", f"idx_{self.name}_target_id")
            await conn.execute(
                f"CREATE INDEX IF NOT EXISTS idx_{self.name}_target_id ON {self.name} (target_id);"
            )

    async def drop(self) -> None:
        """Drop the faces table, index, and validation function idempotently."""
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Dropping index %s", f"idx_{self.name}_target_id")
            await conn.execute(f"DROP INDEX IF EXISTS idx_{self.name}_target_id;")

            self.logger.debug("Dropping table %s", self.name)
            await conn.execute(f"DROP TABLE IF EXISTS {self.name};")

            self.logger.debug("Dropping function %s", self.func_name)
            await conn.execute(f"DROP FUNCTION IF EXISTS {self.func_name};")

    async def get_one_by_id(self, _id: UUID) -> FacesRead:
        """Fetch a single face by id.

        Args:
            _id: Face identifier.

        Returns:
            Face row validated as FacesRead.

        Raises:
            DBException: If the provided id is invalid.
        """
        if not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        where = FacesFilters(id=_id)
        return await self.get_one(where)
