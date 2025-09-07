from __future__ import annotations

from uuid import UUID

from asyncpg import Pool

from shared.models.parent_model import ParentModel
from shared.schema import PerformanceFilter, PerformanceRead


class PerformanceDB(
    ParentModel[PerformanceRead, PerformanceRead, PerformanceRead, PerformanceFilter]
):
    """Read-only wrapper over the performance SQL VIEW.

    Aligns with reference SQL: joins shots, arrows, and targets (by session)
    and exposes computed columns `time_of_flight_seconds`, `arrow_speed`, and
    `score` (via get_shot_score).
    """

    def __init__(self, db_pool: Pool) -> None:
        # This is a VIEW, so the schema is not a CREATE TABLE schema. We keep a minimal
        # placeholder but do not call create_table; instead expose create_view().
        super().__init__("performance", db_pool, PerformanceRead)

    async def create(self) -> None:
        """Create or replace the performance SQL view and helper function.

        Matches the provided SQL reference: uses get_shot_score(shot_x, shot_y,
        target_id, target_max_x, target_max_y) and selects a minimal column set.
        """
        await self._create_get_shot_score_function()

        view_sql = f"""
            CREATE OR REPLACE VIEW {self.name} AS
            SELECT
                shots.id,
                shots.arrow_engage_time,
                shots.arrow_disengage_time,
                shots.arrow_landing_time,
                shots.x,
                shots.y,
                arrows.human_identifier AS arrow_human_identifier,
                arrows.id AS arrow_id,
                extract(
                    EPOCH FROM (shots.arrow_landing_time - shots.arrow_disengage_time)
                ) AS time_of_flight_seconds,
                targets.distance::FLOAT / nullif(
                    extract(EPOCH FROM (shots.arrow_landing_time - shots.arrow_disengage_time)),
                    0
                ) AS arrow_speed,
                get_shot_score(shots.x, shots.y, targets.id, targets.max_x, targets.max_y) AS score
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
        await self.execute(view_sql)

    async def _create_get_shot_score_function(self) -> None:
        """Create get_shot_score helper used by the performance view (matches SQL)."""
        function_sql = """
        CREATE OR REPLACE FUNCTION get_shot_score(
            shot_x REAL,
            shot_y REAL,
            target_id UUID,
            target_max_x REAL,
            target_max_y REAL
        ) RETURNS INTEGER AS $$
        DECLARE
            face RECORD;
            distance REAL;
            score INTEGER := NULL;
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
            WHERE target_id = get_shot_score.target_id
            LIMIT 3;

            -- If no faces exist, return NULL
            IF face_count = 0 THEN
                RETURN NULL;
            END IF;

            -- Check faces for the target
            FOR face IN (
                SELECT x, y, radii, points
                FROM faces
                WHERE target_id = get_shot_score.target_id
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

            -- Return the highest score, or 0 if no face was hit
            RETURN max_score;
        END;
        $$ LANGUAGE plpgsql STABLE;
        """
        await self.execute(function_sql)

    async def get_by_session_id(self, session_id: UUID) -> list[PerformanceRead]:
        """Fetch performance rows for a given session.

        The view doesn't project session_id; query via join to shots.
        """
        sql = f"""
            SELECT p.*
            FROM {self.name} p
            JOIN shots s ON s.id = p.id
            WHERE s.session_id = $1
        """
        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(sql, session_id)
        return [self.read_schema(**r) for r in rows]

    async def drop(self) -> None:
        """Drop the performance SQL view if it exists."""
        async with self.db_pool.acquire() as conn:
            await conn.execute(f"DROP VIEW IF EXISTS {self.name};")

    async def insert_one(self, data: PerformanceRead) -> UUID:
        raise NotImplementedError(f"{self.name} view is read-only")

    async def update_one(self, _id: UUID, data: PerformanceRead) -> None:
        raise NotImplementedError(f"{self.name} view is read-only")

    async def delete_one(self, _id: UUID) -> None:
        raise NotImplementedError(f"{self.name} view is read-only")
