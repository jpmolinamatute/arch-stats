from asyncpg import Pool

from server.models.base_db import DBBase
from server.schema import ShotsCreate, ShotsUpdate


# pylint: disable=too-few-public-methods
class ShotsDB(DBBase[ShotsCreate, ShotsUpdate]):

    def __init__(self, db_pool: Pool) -> None:
        schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            arrow_id UUID NOT NULL,
            session_id UUID NOT NULL,
            arrow_engage_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrow_disengage_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrow_landing_time TIMESTAMP WITH TIME ZONE,
            x_coordinate REAL,
            y_coordinate REAL,
            FOREIGN KEY (arrow_id) REFERENCES arrows (id) ON DELETE CASCADE,
            FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE,
            CHECK (
                (
                    arrow_landing_time IS NOT NULL AND
                    x_coordinate IS NOT NULL AND
                    y_coordinate IS NOT NULL
                )
                OR
                (
                    arrow_landing_time IS NULL AND
                    x_coordinate IS NULL AND
                    y_coordinate IS NULL
                )
            )
        """
        super().__init__("shots", schema, db_pool)

    async def create_notification(self, channel: str) -> None:
        """Ensure the notification function and trigger exist for new shots inserts."""

        function_sql = f"""
        CREATE OR REPLACE FUNCTION notify_new_shot_{channel}() RETURNS TRIGGER AS $$
        BEGIN
            PERFORM pg_notify('{channel}', row_to_json(NEW)::text);
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
        """
        trigger_sql = f"""
        DO $$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM pg_trigger WHERE tgname = 'shots_notify_trigger_{channel}'
            ) THEN
                EXECUTE format(
                    'CREATE TRIGGER shots_notify_trigger_%s AFTER INSERT ON shots
                     FOR EACH ROW EXECUTE FUNCTION notify_new_shot_%s();',
                    '{channel}', '{channel}');
            END IF;
        END$$;
        """
        async with self.db_pool.acquire() as conn:
            await conn.execute(function_sql)
            await conn.execute(trigger_sql)
