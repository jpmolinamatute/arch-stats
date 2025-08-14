from uuid import UUID

from asyncpg import Pool

from server.models.base_db import DBBase
from server.schema import TargetsCreate, TargetsRead, TargetsUpdate


# pylint: disable=too-few-public-methods


class TargetsDB(DBBase[TargetsCreate, TargetsUpdate, TargetsRead]):
    def __init__(self, db_pool: Pool) -> None:
        # face column looks like this:
        # [
        #     {
        #         "x": 123.456,
        #         "y": 111.456,
        #         "human_identifier": "a1",
        #         "radii": [50.456, 60.456, 80.456, 90.456, 100.456],
        #         "points": [10, 9, 8, 7, 5]
        #     },
        #     {
        #         "x": 123.456,
        #         "y": 111.49,
        #         "human_identifier": "a1",
        #         "radii": [50.456, 60.456, 80.456, 90.456, 100.456],
        #         "points": [10, 9, 8, 7, 5]
        #     },
        #     {
        #         "x": 123.49,
        #         "y": 111.456,
        #         "human_identifier": "a1",
        #         "radii": [50.456, 60.456, 80.456, 90.456, 100.456],
        #         "points": [10, 9, 8, 7, 5]
        #     }
        # ]
        schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            max_x REAL NOT NULL,
            max_y REAL NOT NULL,
            session_id UUID NOT NULL,
            faces JSONB NOT NULL,
            FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE,
            CHECK (validate_faces_jsonb(faces))
        """
        super().__init__("targets", schema, db_pool, TargetsRead, ["faces"])

    async def create_validation_function(self) -> None:
        """
        Create constraints for the targets table.
        """
        await self.db_pool.execute(
            """
            CREATE OR REPLACE FUNCTION validate_faces_jsonb(faces JSONB)
            RETURNS BOOLEAN AS $$
            DECLARE
                face JSONB;
                radii_len INT;
                identifiers TEXT[];
                ident TEXT;
                i INT;
            BEGIN
                identifiers := ARRAY[]::TEXT[];

                FOR face IN SELECT * FROM jsonb_array_elements(faces)
                LOOP
                    IF (face->>'x') IS NULL OR (face->>'y') IS NULL THEN
                        RETURN FALSE;
                    END IF;
                    IF (face->>'x')::REAL <= 0 OR (face->>'y')::REAL <= 0 THEN
                        RETURN FALSE;
                    END IF;

                    radii_len := jsonb_array_length(face->'radii');
                    IF radii_len IS NULL OR
                    radii_len = 0 THEN
                        RETURN FALSE;
                    END IF;

                    FOR i IN 0..(radii_len - 1) LOOP
                        IF (face->'radii'->>i)::REAL <= 0 THEN
                            RETURN FALSE;
                        END IF;
                    END LOOP;

                    ident := face->>'human_identifier';
                    IF ident IS NULL OR ident = '' THEN
                        RETURN FALSE;
                    END IF;
                    IF ident = ANY(identifiers) THEN
                        RETURN FALSE;
                    END IF;
                    identifiers := array_append(identifiers, ident);
                END LOOP;

                RETURN TRUE;
            END;
            $$ LANGUAGE plpgsql IMMUTABLE;

        """
        )

    async def get_by_session_id(self, session_id: UUID) -> list[TargetsRead]:
        return await self.get_all({"session_id": session_id})
