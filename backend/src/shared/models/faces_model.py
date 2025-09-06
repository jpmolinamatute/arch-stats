from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import DBException, ParentModel
from shared.schema.faces_schema import FacesCreate, FacesFilters, FacesRead, FacesUpdate


class FacesModel(ParentModel[FacesCreate, FacesUpdate, FacesRead, FacesFilters]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("faces", db_pool, FacesRead)

    async def create(self) -> None:
        await self.create_validation_function()
        faces_schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            x REAL NOT NULL,
            y REAL NOT NULL,
            human_identifier TEXT NOT NULL,
            radii REAL[] NOT NULL,
            points INTEGER[] NOT NULL,
            target_id UUID NOT NULL,
            FOREIGN KEY (target_id) REFERENCES targets (id) ON DELETE CASCADE,
            UNIQUE (target_id, human_identifier),
            CHECK (
                array_length(radii, 1) > 0 AND
                array_length(points, 1) > 0 AND
                array_length(points, 1) = array_length(radii, 1) AND
                validate_face_row(target_id, x, y, radii)
            )
        """
        await self.execute(f"CREATE TABLE IF NOT EXISTS {self.name} ({faces_schema});")
        await self.execute(
            f"CREATE INDEX IF NOT EXISTS idx_{self.name}_target_id ON {self.name} (target_id);"
        )

    async def drop(self) -> None:
        await self.execute(f"DROP TABLE IF EXISTS {self.name} CASCADE;")

    async def create_validation_function(self) -> None:
        validation_function = """
            CREATE OR REPLACE FUNCTION validate_face_row(
                target_id UUID,
                x REAL,
                y REAL,
                radii REAL[]
            ) RETURNS BOOLEAN AS $$
            DECLARE
                max_radius REAL;
                tgt_max_x REAL;
                tgt_max_y REAL;
            BEGIN
                SELECT max_x, max_y INTO tgt_max_x, tgt_max_y FROM targets WHERE id = target_id;

                -- Compute max radius
                SELECT MAX(r) INTO max_radius FROM unnest(radii) AS r;

                -- Ensure face fully contained within target bounds (allowing edge contact)
                IF (x + max_radius) >= tgt_max_x OR (y + max_radius) >= tgt_max_y THEN
                    RETURN FALSE;
                END IF;
                IF (x - max_radius) <= 0 OR (y - max_radius) <= 0 THEN
                    RETURN FALSE;
                END IF;

                RETURN TRUE;
            END;
            $$ LANGUAGE plpgsql STABLE;
        """
        await self.execute(validation_function)

    async def get_one_by_id(self, _id: UUID) -> FacesRead:
        """
        Retrieve a single record by its UUID.
        """
        if not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        where = FacesFilters(id=_id)
        return await self.get_one(where)
