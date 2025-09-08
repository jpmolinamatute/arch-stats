from __future__ import annotations

from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import ParentModel
from shared.schema import (
    SessionPerformanceCreate,
    SessionPerformanceFilter,
    SessionPerformanceRead,
    SessionPerformanceUpdate,
)


class SessionPerformanceModel(
    ParentModel[
        SessionPerformanceCreate,
        SessionPerformanceUpdate,
        SessionPerformanceRead,
        SessionPerformanceFilter,
    ]
):
    """Read-only wrapper over the session performance SQL view."""

    def __init__(self, db_pool: Pool) -> None:
        super().__init__("session_performance", db_pool, SessionPerformanceRead)
        self.func_name = "get_shot_score"

    async def create(self) -> None:
        """Create or replace the performance SQL view and helper function.

        Uses {self.func_name}(shot_x, shot_y, target_id, target_max_x,
        target_max_y) and selects a minimal column set.
        """
        function_sql = f"""
            CREATE OR REPLACE FUNCTION {self.func_name}(
                shot_x REAL,
                shot_y REAL,
                target_id UUID,
                target_max_x REAL,
                target_max_y REAL
            ) RETURNS INTEGER AS $$
            DECLARE
                face RECORD;
                distance REAL;
                score INTEGER := 0;
                max_score INTEGER := 0;
                face_count INTEGER;
            BEGIN
                -- If shot coordinates are null, it missed the target entirely
                IF shot_x IS NULL OR shot_y IS NULL THEN
                    RETURN NULL;
                END IF;

                -- Check unrealistic shot coordinates
                IF shot_x < 0 OR shot_y < 0 OR shot_x > target_max_x OR shot_y > target_max_y THEN
                    RETURN NULL;
                END IF;

                -- Check if any faces exist for the target
                SELECT COUNT(*) INTO face_count
                FROM faces
                WHERE faces.target_id = {self.func_name}.target_id
                LIMIT 3;

                -- If no faces exist, return NULL
                IF face_count = 0 THEN
                    RETURN NULL;
                END IF;

                -- Check up to 3 faces for the target
                FOR face IN (
                    SELECT x, y, radii, points
                    FROM faces
                    WHERE faces.target_id = {self.func_name}.target_id
                ) LOOP
                    distance := SQRT(POWER(shot_x - face.x, 2) + POWER(shot_y - face.y, 2));
                    FOR i IN 1..array_length(face.radii, 1) LOOP
                        IF distance <= face.radii[i] THEN
                            score := face.points[i];
                            IF max_score IS NULL OR score > max_score THEN
                                max_score := score;
                            END IF;
                            EXIT;
                        END IF;
                    END LOOP;
                END LOOP;

                -- Return the highest score, or NULL if no face was hit
                RETURN max_score;
            END;
            $$ LANGUAGE plpgsql STABLE;
        """

        view_sql = f"""
            CREATE OR REPLACE VIEW {self.name} AS
            SELECT
                shots.id,
                shots.session_id,
                shots.arrow_engage_time,
                shots.arrow_disengage_time,
                shots.arrow_landing_time,
                shots.x,
                shots.y,
                arrows.id AS arrow_id,
                arrows.human_identifier,
                extract(
                    EPOCH FROM (shots.arrow_landing_time - shots.arrow_disengage_time)
                ) AS time_of_flight_seconds,
                targets.distance::FLOAT / nullif(
                    extract(EPOCH FROM (shots.arrow_landing_time - shots.arrow_disengage_time)),
                    0
                ) AS arrow_speed,
                {self.func_name}(
                    shots.x,
                    shots.y,
                    targets.id,
                    targets.max_x,
                    targets.max_y
                ) AS score
            FROM
                shots
            INNER JOIN
                arrows ON shots.arrow_id = arrows.id
            INNER JOIN
                targets ON shots.session_id = targets.session_id
            WHERE shots.session_id = (
                SELECT sessions.id
                FROM sessions
                WHERE sessions.is_opened = TRUE
                ORDER BY sessions.start_time DESC
                LIMIT 1
            );
        """
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Creating function %s", self.func_name)
            await conn.execute(function_sql)
            self.logger.debug("Creating view %s", self.name)
            await conn.execute(view_sql)

    async def drop(self) -> None:
        """Drop the performance SQL view and helper function if they exist."""
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Dropping view %s", self.name)
            await conn.execute(f"DROP VIEW IF EXISTS {self.name};")
            self.logger.debug("Dropping function %s", self.func_name)
            await conn.execute(
                f"DROP FUNCTION IF EXISTS {self.func_name}(real, real, uuid, real, real);"
            )

    async def get_all(
        self, where: SessionPerformanceFilter | None = None
    ) -> list[SessionPerformanceRead]:
        """Fetch all rows from the performance view.

        Args:
            where: Unused placeholder for API parity.

        Returns:
            List of SessionPerformanceRead entries for the current open session.
        """
        sql_statement = f"SELECT * FROM {self.name};"

        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(sql_statement)

        return [self.read_schema(**row) for row in rows]

    async def insert_one(self, data: SessionPerformanceCreate) -> UUID:
        """Disallow inserts because this model wraps a read-only view.

        Raises:
            NotImplementedError: Always.
        """
        raise NotImplementedError(f"{self.name} view is read-only")

    async def update_one(self, _id: UUID, data: SessionPerformanceUpdate) -> None:
        """Disallow updates because this model wraps a read-only view.

        Raises:
            NotImplementedError: Always.
        """
        raise NotImplementedError(f"{self.name} view is read-only")

    async def delete_one(self, _id: UUID) -> None:
        """Disallow deletes because this model wraps a read-only view.

        Raises:
            NotImplementedError: Always.
        """
        raise NotImplementedError(f"{self.name} view is read-only")
